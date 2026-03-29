// Copyright 2026 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package main

import (
	"bytes"
	"context"
	"errors"
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
	"time"

	"github.com/rowland/leadtype/internal/overlayfs"
	"github.com/rowland/leadtype/ltml"
	"github.com/rowland/leadtype/ltml/ltpdf"
)

const defaultPollInterval = 500 * time.Millisecond

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

type runConfig struct {
	assetsDir    string
	outputPath   string
	submitURL    string
	extraFiles   []string
	watch        bool
	batch        bool
	pollInterval time.Duration
	stderr       io.Writer
	onRender     func(renderJob, error)
}

type renderJob struct {
	inputPath  string
	outputPath string
}

type watchPaths struct {
	inputs    []string
	extras    []string
	assetsDir string
}

type watchState struct {
	inputs    map[string]string
	extras    map[string]string
	assetsDir string
}

func (cfg runConfig) logf(format string, args ...any) {
	w := cfg.stderr
	if w == nil {
		w = io.Discard
	}
	fmt.Fprintf(w, "render-ltml: "+format+"\n", args...)
}

func displayPath(path string) string {
	if path == "" {
		return path
	}
	cwd, err := os.Getwd()
	if err != nil {
		return path
	}
	rel, err := filepath.Rel(cwd, path)
	if err != nil {
		return path
	}
	if rel == "." {
		return rel
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return path
	}
	return rel
}

func main() {
	var cfg runConfig
	var extraFiles multiFlag

	cfg.stderr = os.Stderr
	cfg.pollInterval = defaultPollInterval

	flag.StringVar(&cfg.assetsDir, "assets", "", "path to asset `directory`")
	flag.StringVar(&cfg.assetsDir, "a", "", "path to asset `directory` (shorthand)")
	flag.StringVar(&cfg.outputPath, "output", "", "output `file` or batch output `directory`")
	flag.StringVar(&cfg.outputPath, "o", "", "output `file` or batch output `directory` (shorthand)")
	flag.StringVar(&cfg.submitURL, "submit", "", "submit to remote render `url` instead of rendering locally")
	flag.BoolVar(&cfg.watch, "watch", false, "watch inputs and assets for changes and rerender continuously")
	flag.BoolVar(&cfg.watch, "w", false, "watch inputs and assets for changes and rerender continuously (shorthand)")
	flag.BoolVar(&cfg.batch, "batch", false, "render multiple input files")
	flag.BoolVar(&cfg.batch, "b", false, "render multiple input files (shorthand)")
	flag.Var(&extraFiles, "extra", "additional asset `file` (may be repeated)")
	flag.Var(&extraFiles, "e", "additional asset `file` (shorthand)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: render-ltml [flags] <file>\n")
		fmt.Fprintf(os.Stderr, "   or: render-ltml -b [flags] <file1> <file2> ...\n\nFlags:\n")
		flag.PrintDefaults()
	}
	args, err := normalizeInterspersedArgs(flag.CommandLine, os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "render-ltml: %v\n", err)
		os.Exit(2)
	}
	if err := flag.CommandLine.Parse(args); err != nil {
		os.Exit(2)
	}

	cfg.extraFiles = []string(extraFiles)

	if err := validateArgs(cfg, flag.Args()); err != nil {
		flag.Usage()
		fmt.Fprintf(os.Stderr, "\nrender-ltml: %v\n", err)
		os.Exit(2)
	}

	if err := run(context.Background(), cfg, flag.Args()); err != nil {
		fmt.Fprintf(os.Stderr, "render-ltml: %v\n", err)
		os.Exit(1)
	}
}

func validateArgs(cfg runConfig, inputFiles []string) error {
	if cfg.batch {
		if len(inputFiles) == 0 {
			return fmt.Errorf("batch mode requires at least one input file")
		}
		return nil
	}
	if len(inputFiles) != 1 {
		return fmt.Errorf("expected exactly one input file")
	}
	return nil
}

type boolFlag interface {
	IsBoolFlag() bool
}

func normalizeInterspersedArgs(fs *flag.FlagSet, args []string) ([]string, error) {
	flagArgs := make([]string, 0, len(args))
	positionalArgs := make([]string, 0, len(args))

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			positionalArgs = append(positionalArgs, args[i+1:]...)
			break
		}
		if !strings.HasPrefix(arg, "-") || arg == "-" {
			positionalArgs = append(positionalArgs, arg)
			continue
		}

		name, hasValue := splitFlagArg(arg)
		if name == "" {
			positionalArgs = append(positionalArgs, arg)
			continue
		}

		f := fs.Lookup(name)
		if f == nil {
			flagArgs = append(flagArgs, arg)
			continue
		}

		flagArgs = append(flagArgs, arg)
		if hasValue || isBoolFlagValue(f.Value) {
			continue
		}
		if i+1 >= len(args) {
			return nil, fmt.Errorf("flag needs an argument: -%s", name)
		}
		i++
		flagArgs = append(flagArgs, args[i])
	}

	return append(flagArgs, positionalArgs...), nil
}

func splitFlagArg(arg string) (name string, hasValue bool) {
	if strings.HasPrefix(arg, "--") {
		rest := strings.TrimPrefix(arg, "--")
		if rest == "" {
			return "", false
		}
		if idx := strings.IndexByte(rest, '='); idx >= 0 {
			return rest[:idx], true
		}
		return rest, false
	}
	rest := strings.TrimPrefix(arg, "-")
	if rest == "" {
		return "", false
	}
	if idx := strings.IndexByte(rest, '='); idx >= 0 {
		return rest[:idx], true
	}
	return rest, false
}

func isBoolFlagValue(v flag.Value) bool {
	bf, ok := v.(boolFlag)
	return ok && bf.IsBoolFlag()
}

func run(ctx context.Context, cfg runConfig, inputFiles []string) error {
	jobs, prepErr := buildRenderJobs(inputFiles, cfg.outputPath, cfg.batch)
	if prepErr != nil && (!cfg.batch || len(jobs) == 0) {
		return prepErr
	}

	var renderErr error
	if cfg.watch {
		renderErr = watchAndRender(ctx, cfg, jobs)
	} else {
		renderErr = renderJobs(cfg, jobs)
	}
	if prepErr != nil && renderErr != nil {
		return errors.Join(prepErr, renderErr)
	}
	if prepErr != nil {
		return prepErr
	}
	if renderErr != nil {
		return renderErr
	}
	return nil
}

func buildRenderJobs(inputFiles []string, outputPath string, batch bool) ([]renderJob, error) {
	if len(inputFiles) == 0 {
		return nil, fmt.Errorf("no input files provided")
	}

	var absOutputDir string
	if batch && outputPath != "" {
		var err error
		absOutputDir, err = filepath.Abs(outputPath)
		if err != nil {
			return nil, fmt.Errorf("resolving output directory: %w", err)
		}
		info, err := os.Stat(absOutputDir)
		if err != nil {
			return nil, fmt.Errorf("stat output directory: %w", err)
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("batch mode requires -output to name an existing directory")
		}
	}

	jobs := make([]renderJob, 0, len(inputFiles))
	var errs []error
	for _, inputFile := range inputFiles {
		job, err := buildRenderJob(inputFile, outputPath, absOutputDir, batch)
		if err != nil {
			if !batch {
				return nil, err
			}
			errs = append(errs, err)
			continue
		}
		jobs = append(jobs, job)
	}
	if len(errs) > 0 {
		return jobs, fmt.Errorf("batch completed with %d preparation error(s): %w", len(errs), errors.Join(errs...))
	}
	return jobs, nil
}

func buildRenderJob(inputFile, outputPath, absOutputDir string, batch bool) (renderJob, error) {
	absInput, err := filepath.Abs(inputFile)
	if err != nil {
		return renderJob{}, fmt.Errorf("resolving input: %w", err)
	}

	var jobOutput string
	switch {
	case batch && absOutputDir != "":
		baseOutput, err := defaultOutputBase(absInput)
		if err != nil {
			return renderJob{}, err
		}
		jobOutput = filepath.Join(absOutputDir, baseOutput)
	case !batch && outputPath != "":
		jobOutput, err = filepath.Abs(outputPath)
		if err != nil {
			return renderJob{}, fmt.Errorf("resolving output: %w", err)
		}
	default:
		jobOutput, err = defaultOutputPath(absInput)
		if err != nil {
			return renderJob{}, err
		}
	}

	return renderJob{
		inputPath:  absInput,
		outputPath: jobOutput,
	}, nil
}

func defaultOutputBase(absInput string) (string, error) {
	base := filepath.Base(absInput)
	ext := filepath.Ext(base)
	if ext == "" {
		return "", fmt.Errorf("default output requires input file %s to have an extension", displayPath(absInput))
	}
	return strings.TrimSuffix(base, ext) + ".pdf", nil
}

func defaultOutputPath(absInput string) (string, error) {
	base, err := defaultOutputBase(absInput)
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(absInput), base), nil
}

func renderJobs(cfg runConfig, jobs []renderJob) error {
	var errs []error
	for _, job := range jobs {
		if err := renderJobToFile(cfg, job); err != nil {
			if !cfg.batch {
				return err
			}
			cfg.logf("%v", err)
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("batch completed with %d render error(s): %w", len(errs), errors.Join(errs...))
	}
	return nil
}

func renderJobToFile(cfg runConfig, job renderJob) (err error) {
	if cfg.onRender != nil {
		defer func() {
			cfg.onRender(job, err)
		}()
	}

	cfg.logf("rendering %s -> %s (%s)", displayPath(job.inputPath), displayPath(job.outputPath), renderMode(cfg))

	out, closeOut, err := createOutputFile(job.outputPath)
	if err != nil {
		return err
	}
	defer closeOut()

	if cfg.submitURL != "" {
		err = submitRemote(job.inputPath, cfg.assetsDir, cfg.submitURL, cfg.extraFiles, out)
	} else {
		err = renderLocal(job.inputPath, cfg.assetsDir, cfg.extraFiles, out)
	}
	if err != nil {
		return err
	}
	cfg.logf("wrote %s", displayPath(job.outputPath))
	return nil
}

func createOutputFile(outputPath string) (io.Writer, func() error, error) {
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

func watchAndRender(ctx context.Context, cfg runConfig, jobs []renderJob) error {
	if cfg.pollInterval <= 0 {
		cfg.pollInterval = defaultPollInterval
	}

	state, err := snapshotWatchState(buildWatchPaths(cfg, jobs))
	if err != nil {
		return err
	}

	cfg.logf("watching %d input(s)", len(jobs))
	if err := renderJobs(cfg, jobs); err != nil {
		cfg.logf("%v", err)
	}

	ticker := time.NewTicker(cfg.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			nextState, changedInputs, sharedChanged, err := detectWatchChanges(buildWatchPaths(cfg, jobs), state)
			if err != nil {
				cfg.logf("%v", err)
				continue
			}
			state = nextState

			if sharedChanged {
				cfg.logf("change detected in shared assets; rerendering all inputs")
				if err := renderJobs(cfg, jobs); err != nil {
					cfg.logf("%v", err)
				}
				continue
			}

			for _, changedInput := range changedInputs {
				job, ok := findJobByInput(jobs, changedInput)
				if !ok {
					continue
				}
				cfg.logf("change detected in %s; rerendering", changedInput)
				if err := renderJobToFile(cfg, job); err != nil {
					cfg.logf("%v", err)
				}
			}
		}
	}
}

func renderMode(cfg runConfig) string {
	if cfg.submitURL != "" {
		return "submit"
	}
	return "local"
}

func buildWatchPaths(cfg runConfig, jobs []renderJob) watchPaths {
	inputs := make([]string, 0, len(jobs))
	for _, job := range jobs {
		inputs = append(inputs, job.inputPath)
	}

	extras := make([]string, 0, len(cfg.extraFiles))
	for _, extraFile := range cfg.extraFiles {
		absExtra, err := filepath.Abs(extraFile)
		if err != nil {
			extras = append(extras, extraFile)
			continue
		}
		extras = append(extras, absExtra)
	}

	assetsDir := ""
	if cfg.assetsDir != "" {
		if absAssets, err := filepath.Abs(cfg.assetsDir); err == nil {
			assetsDir = absAssets
		} else {
			assetsDir = cfg.assetsDir
		}
	}

	return watchPaths{
		inputs:    inputs,
		extras:    extras,
		assetsDir: assetsDir,
	}
}

func snapshotWatchState(paths watchPaths) (watchState, error) {
	state := watchState{
		inputs: make(map[string]string, len(paths.inputs)),
		extras: make(map[string]string, len(paths.extras)),
	}

	for _, inputPath := range paths.inputs {
		token, err := pathToken(inputPath)
		if err != nil {
			return watchState{}, err
		}
		state.inputs[inputPath] = token
	}

	for _, extraPath := range paths.extras {
		token, err := pathToken(extraPath)
		if err != nil {
			return watchState{}, err
		}
		state.extras[extraPath] = token
	}

	if paths.assetsDir != "" {
		token, err := dirToken(paths.assetsDir)
		if err != nil {
			return watchState{}, err
		}
		state.assetsDir = token
	}

	return state, nil
}

func detectWatchChanges(paths watchPaths, prev watchState) (watchState, []string, bool, error) {
	next, err := snapshotWatchState(paths)
	if err != nil {
		return watchState{}, nil, false, err
	}

	var changedInputs []string
	for _, inputPath := range paths.inputs {
		if prev.inputs[inputPath] != next.inputs[inputPath] {
			changedInputs = append(changedInputs, inputPath)
		}
	}

	sharedChanged := false
	for _, extraPath := range paths.extras {
		if prev.extras[extraPath] != next.extras[extraPath] {
			sharedChanged = true
			break
		}
	}
	if !sharedChanged && paths.assetsDir != "" && prev.assetsDir != next.assetsDir {
		sharedChanged = true
	}

	return next, changedInputs, sharedChanged, nil
}

func findJobByInput(jobs []renderJob, inputPath string) (renderJob, bool) {
	for _, job := range jobs {
		if job.inputPath == inputPath {
			return job, true
		}
	}
	return renderJob{}, false
}

func pathToken(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "missing", nil
		}
		return "", fmt.Errorf("stat %s: %w", path, err)
	}
	return fmt.Sprintf("%s|%d|%d", info.Mode().String(), info.Size(), info.ModTime().UnixNano()), nil
}

func dirToken(root string) (string, error) {
	info, err := os.Stat(root)
	if err != nil {
		if os.IsNotExist(err) {
			return "missing", nil
		}
		return "", fmt.Errorf("stat %s: %w", root, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%s is not a directory", root)
	}

	var b strings.Builder
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		entryInfo, err := d.Info()
		if err != nil {
			return err
		}

		b.WriteString(rel)
		b.WriteByte('|')
		b.WriteString(entryInfo.Mode().String())
		b.WriteByte('|')
		b.WriteString(fmt.Sprintf("%d|%d\n", entryInfo.Size(), entryInfo.ModTime().UnixNano()))
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("walking %s: %w", root, err)
	}

	return b.String(), nil
}

func renderLocal(absInput, assetsDir string, extraFiles []string, out io.Writer) error {
	doc, err := ltml.ParseFile(absInput)
	if err != nil {
		return fmt.Errorf("parsing %s: %w", displayPath(absInput), err)
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
		return fmt.Errorf("creating request for %s to %s: %w", displayPath(absInput), submitURL, err)
	}
	req.Header.Set("Content-Type", contentType)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("submitting %s to %s: %w", displayPath(absInput), submitURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg, readErr := readTrimmedResponse(resp.Body)
		if readErr != nil {
			return fmt.Errorf("remote render for %s via %s failed: %s (reading response: %v)", displayPath(absInput), submitURL, resp.Status, readErr)
		}
		if msg == "" {
			return fmt.Errorf("remote render for %s via %s failed: %s", displayPath(absInput), submitURL, resp.Status)
		}
		return fmt.Errorf("remote render for %s via %s failed: %s: %s", displayPath(absInput), submitURL, resp.Status, msg)
	}

	if resp.Body == nil {
		return fmt.Errorf("remote render for %s via %s returned an empty response body", displayPath(absInput), submitURL)
	}

	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("writing remote output for %s from %s: %w", displayPath(absInput), submitURL, err)
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
