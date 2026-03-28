// Copyright 2026 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"

	"github.com/rowland/leadtype/internal/overlayfs"
	"github.com/rowland/leadtype/ltml"
	"github.com/rowland/leadtype/ltml/ltpdf"
)

type multiFlag []string

func (m *multiFlag) String() string {
	if m == nil || len(*m) == 0 {
		return ""
	}
	return fmt.Sprintf("%v", []string(*m))
}

func (m *multiFlag) Set(v string) error {
	*m = append(*m, v)
	return nil
}

func main() {
	var assetsDir string
	var outputPath string
	var submitURL string
	var extraFiles multiFlag

	flag.StringVar(&assetsDir, "assets", "", "path to asset `directory`")
	flag.StringVar(&assetsDir, "a", "", "path to asset `directory` (shorthand)")
	flag.StringVar(&outputPath, "output", "", "output `file` (default: stdout)")
	flag.StringVar(&outputPath, "o", "", "output `file` (shorthand)")
	flag.StringVar(&submitURL, "submit", "", "submit to remote render `url` instead of rendering locally")
	flag.Var(&extraFiles, "extra", "additional asset `file` (may be repeated)")
	flag.Var(&extraFiles, "e", "additional asset `file` (shorthand)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: render-ltml [flags] <file.ltml>\n\nFlags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}

	if err := run(flag.Arg(0), assetsDir, outputPath, submitURL, []string(extraFiles)); err != nil {
		fmt.Fprintf(os.Stderr, "render-ltml: %v\n", err)
		os.Exit(1)
	}
}

func run(inputFile, assetsDir, outputPath, submitURL string, extraFiles []string) error {
	absInput, err := filepath.Abs(inputFile)
	if err != nil {
		return fmt.Errorf("resolving input: %w", err)
	}

	out, cleanup, err := createOutputWriter(outputPath)
	if err != nil {
		return err
	}
	if cleanup != nil {
		defer cleanup()
	}

	if submitURL != "" {
		return submitRemote(absInput, assetsDir, submitURL, extraFiles, out)
	}
	return renderLocal(absInput, assetsDir, extraFiles, out)
}

func createOutputWriter(outputPath string) (io.Writer, func() error, error) {
	if outputPath == "" {
		return os.Stdout, nil, nil
	}

	absOutput, err := filepath.Abs(outputPath)
	if err != nil {
		return nil, nil, fmt.Errorf("resolving output: %w", err)
	}

	f, err := os.Create(absOutput)
	if err != nil {
		return nil, nil, fmt.Errorf("creating output: %w", err)
	}
	return f, f.Close, nil
}

func renderLocal(absInput, assetsDir string, extraFiles []string, out io.Writer) error {
	doc, err := ltml.ParseFile(absInput)
	if err != nil {
		return fmt.Errorf("parsing %s: %w", absInput, err)
	}

	assetFS, cleanup, err := buildOptionalAssetFS(assetsDir, extraFiles)
	if err != nil {
		return err
	}
	if cleanup != nil {
		defer cleanup()
	}

	w := ltpdf.NewDocWriter()
	if assetFS != nil {
		w.SetAssetFS(assetFS)
	}
	if err := doc.Print(w); err != nil {
		return fmt.Errorf("rendering: %w", err)
	}

	if _, err := w.WriteTo(out); err != nil {
		return fmt.Errorf("writing output: %w", err)
	}
	return nil
}

func submitRemote(absInput, assetsDir, submitURL string, extraFiles []string, out io.Writer) error {
	if assetsDir != "" {
		return fmt.Errorf("remote submission does not support -assets; upload explicit files with -extra instead")
	}

	ltmlBytes, err := os.ReadFile(absInput)
	if err != nil {
		return fmt.Errorf("reading input: %w", err)
	}

	body, contentType, err := buildRemoteRequestBody(absInput, ltmlBytes, extraFiles)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, submitURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", contentType)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("submitting request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg, readErr := readTrimmedResponse(resp.Body)
		if readErr != nil {
			return fmt.Errorf("remote render failed: %s (reading response: %v)", resp.Status, readErr)
		}
		if msg == "" {
			return fmt.Errorf("remote render failed: %s", resp.Status)
		}
		return fmt.Errorf("remote render failed: %s: %s", resp.Status, msg)
	}

	if resp.Body == nil {
		return fmt.Errorf("remote render returned an empty response body")
	}

	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("writing output: %w", err)
	}
	return nil
}

func buildRemoteRequestBody(absInput string, ltmlBytes []byte, extraFiles []string) ([]byte, string, error) {
	if err := validateUniqueExtraBaseNames(extraFiles); err != nil {
		return nil, "", err
	}

	var body bytes.Buffer
	mw := multipart.NewWriter(&body)

	ltmlHeader := textproto.MIMEHeader{}
	ltmlHeader.Set("Content-Disposition", `form-data; name="ltml"`)
	ltmlHeader.Set("Content-Type", "application/vnd.rowland.leadtype.ltml+xml")

	ltmlPart, err := mw.CreatePart(ltmlHeader)
	if err != nil {
		return nil, "", fmt.Errorf("creating LTML part: %w", err)
	}
	if _, err := ltmlPart.Write(ltmlBytes); err != nil {
		return nil, "", fmt.Errorf("writing LTML part: %w", err)
	}

	for _, extraFile := range extraFiles {
		if err := addExtraFilePart(mw, extraFile); err != nil {
			return nil, "", err
		}
	}

	if err := mw.Close(); err != nil {
		return nil, "", fmt.Errorf("finalizing multipart body for %s: %w", absInput, err)
	}
	return body.Bytes(), mw.FormDataContentType(), nil
}

func addExtraFilePart(mw *multipart.Writer, extraFile string) error {
	absExtra, err := filepath.Abs(extraFile)
	if err != nil {
		return fmt.Errorf("resolving extra file %s: %w", extraFile, err)
	}

	f, err := os.Open(absExtra)
	if err != nil {
		return fmt.Errorf("opening extra file %s: %w", extraFile, err)
	}
	defer f.Close()

	fileHeader := textproto.MIMEHeader{}
	fileHeader.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, escapeMultipartFilename(filepath.Base(extraFile))))
	fileHeader.Set("Content-Type", "application/octet-stream")

	part, err := mw.CreatePart(fileHeader)
	if err != nil {
		return fmt.Errorf("creating file part for %s: %w", extraFile, err)
	}
	if _, err := io.Copy(part, f); err != nil {
		return fmt.Errorf("writing file part for %s: %w", extraFile, err)
	}
	return nil
}

func validateUniqueExtraBaseNames(extraFiles []string) error {
	seen := make(map[string]string, len(extraFiles))
	for _, extraFile := range extraFiles {
		base := filepath.Base(extraFile)
		if prev, ok := seen[base]; ok {
			return fmt.Errorf("duplicate -extra base name %q from %s and %s", base, prev, extraFile)
		}
		seen[base] = extraFile
	}
	return nil
}

func escapeMultipartFilename(name string) string {
	replacer := strings.NewReplacer("\\", "\\\\", `"`, "\\\"")
	return replacer.Replace(name)
}

func readTrimmedResponse(r io.Reader) (string, error) {
	data, err := io.ReadAll(io.LimitReader(r, 4096))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

// buildOptionalAssetFS constructs an optional fs.FS that covers assetsDir
// (lower layer) and the named extraFiles (upper layer, each stored under its
// base name). Extra files shadow same-named entries in assetsDir rather than
// erroring on conflict.
//
// Returns nil, nil, nil when neither assetsDir nor extraFiles are provided.
// When a non-nil cleanup function is returned, the caller must invoke it after
// rendering is complete.
func buildOptionalAssetFS(assetsDir string, extraFiles []string) (fs.FS, func(), error) {
	hasAssets := assetsDir != ""
	hasExtras := len(extraFiles) > 0

	if !hasAssets && !hasExtras {
		return nil, nil, nil
	}

	// Place extra files in a temp directory so they can be addressed as an
	// os.DirFS. Symlinks keep memory use low for large files.
	var extraDir string
	var cleanup func()
	if hasExtras {
		tmpDir, err := os.MkdirTemp("", "render-ltml-*")
		if err != nil {
			return nil, nil, fmt.Errorf("creating work dir: %w", err)
		}
		cleanup = func() { os.RemoveAll(tmpDir) }
		extraDir = tmpDir

		for _, f := range extraFiles {
			abs, err := filepath.Abs(f)
			if err != nil {
				cleanup()
				return nil, nil, fmt.Errorf("resolving extra file %s: %w", f, err)
			}
			dst := filepath.Join(extraDir, filepath.Base(f))
			// If two extra files share a base name, the last one wins.
			os.Remove(dst)
			if err := os.Symlink(abs, dst); err != nil {
				cleanup()
				return nil, nil, fmt.Errorf("linking extra file %s: %w", filepath.Base(f), err)
			}
		}
	}

	switch {
	case hasExtras && hasAssets:
		absAssets, err := filepath.Abs(assetsDir)
		if err != nil {
			cleanup()
			return nil, nil, fmt.Errorf("resolving assets dir: %w", err)
		}
		return overlayfs.New(os.DirFS(extraDir), os.DirFS(absAssets)), cleanup, nil

	case hasExtras:
		return os.DirFS(extraDir), cleanup, nil

	default: // hasAssets only
		absAssets, err := filepath.Abs(assetsDir)
		if err != nil {
			return nil, nil, fmt.Errorf("resolving assets dir: %w", err)
		}
		return os.DirFS(absAssets), nil, nil
	}
}
