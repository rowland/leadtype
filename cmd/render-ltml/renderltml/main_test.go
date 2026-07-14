// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package renderltml

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/rowland/leadtype/ltml"
)

type renderLocalComponent struct {
	ltml.StdComponent
}

var (
	registerRenderLocalComponentOnce sync.Once
	renderLocalComponentMu           sync.Mutex
	renderLocalComponentBody         string
)

func (c *renderLocalComponent) DrawContent(w ltml.Writer) error {
	renderLocalComponentMu.Lock()
	defer renderLocalComponentMu.Unlock()
	renderLocalComponentBody = c.Body()
	if body := c.Body(); body == "" {
		return fmt.Errorf("component body is empty")
	}
	return nil
}

func registerRenderLocalComponent(t *testing.T) {
	t.Helper()
	registerRenderLocalComponentOnce.Do(func() {
		if err := ltml.RegisterTagExt("rendercomponenttest", "card", func() any { return &renderLocalComponent{} }); err != nil {
			t.Fatalf("register component tag: %v", err)
		}
	})
	renderLocalComponentMu.Lock()
	renderLocalComponentBody = ""
	renderLocalComponentMu.Unlock()
}

func mustExtraAssets(t *testing.T, values ...string) []extraAsset {
	t.Helper()
	extras, err := parseExtraAssets(values)
	if err != nil {
		t.Fatalf("parse extra assets: %v", err)
	}
	return extras
}

func multipartFilenameParam(t *testing.T, part *multipart.Part) string {
	t.Helper()
	_, params, err := mime.ParseMediaType(part.Header.Get("Content-Disposition"))
	if err != nil {
		t.Fatalf("parse content disposition: %v", err)
	}
	return params["filename"]
}

func TestBuildOptionalAssetFS_ExtraOverridesAssetsDir(t *testing.T) {
	assetsDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(assetsDir, "logo.txt"), []byte("lower"), 0o600); err != nil {
		t.Fatal(err)
	}

	extraDir := t.TempDir()
	extraFile := filepath.Join(extraDir, "logo.txt")
	if err := os.WriteFile(extraFile, []byte("upper"), 0o600); err != nil {
		t.Fatal(err)
	}

	assetFS, cleanup, err := buildOptionalAssetFS(assetsDir, mustExtraAssets(t, extraFile))
	if err != nil {
		t.Fatal(err)
	}
	if cleanup != nil {
		defer cleanup()
	}

	data, err := fs.ReadFile(assetFS, "logo.txt")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "upper" {
		t.Fatalf("expected upper override, got %q", data)
	}
}

func TestBuildOptionalAssetFS_ExtraMapsToVirtualPath(t *testing.T) {
	extraDir := t.TempDir()
	extraFile := filepath.Join(extraDir, "logo.txt")
	if err := os.WriteFile(extraFile, []byte("mapped"), 0o600); err != nil {
		t.Fatal(err)
	}

	assetFS, cleanup, err := buildOptionalAssetFS("", mustExtraAssets(t, extraFile+":assets/logo.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if cleanup != nil {
		defer cleanup()
	}

	data, err := fs.ReadFile(assetFS, "assets/logo.txt")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "mapped" {
		t.Fatalf("expected mapped extra, got %q", data)
	}
}

func TestBuildOptionalAssetFS_PreservesNestedAssetPathsFromAssetsDir(t *testing.T) {
	assetsDir := t.TempDir()
	nestedDir := filepath.Join(assetsDir, "assets")
	if err := os.MkdirAll(nestedDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nestedDir, "logo.txt"), []byte("nested"), 0o600); err != nil {
		t.Fatal(err)
	}

	assetFS, cleanup, err := buildOptionalAssetFS(assetsDir, nil)
	if err != nil {
		t.Fatal(err)
	}
	if cleanup != nil {
		defer cleanup()
	}

	data, err := fs.ReadFile(assetFS, "assets/logo.txt")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "nested" {
		t.Fatalf("expected nested asset, got %q", data)
	}
}

func TestParseExtraAssets(t *testing.T) {
	extras := mustExtraAssets(t, "/tmp/logo.txt", "/tmp/generated.txt:assets/logo.txt")
	if got, want := extras[0], (extraAsset{sourcePath: "/tmp/logo.txt", virtualPath: "logo.txt"}); got != want {
		t.Fatalf("unmapped extra = %#v, want %#v", got, want)
	}
	if got, want := extras[1], (extraAsset{sourcePath: "/tmp/generated.txt", virtualPath: "assets/logo.txt"}); got != want {
		t.Fatalf("mapped extra = %#v, want %#v", got, want)
	}
}

func TestParseExtraAsset_RejectsInvalidVirtualPaths(t *testing.T) {
	tests := []string{
		"/tmp/logo.txt:",
		"/tmp/logo.txt:.",
		"/tmp/logo.txt:./logo.png",
		"/tmp/logo.txt:a/../logo.png",
		"/tmp/logo.txt:/assets/logo.png",
	}
	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			if _, err := parseExtraAsset(tt); err == nil {
				t.Fatal("expected invalid virtual path error")
			}
		})
	}
}

func TestRenderLocal_SetsParserAssetFSForComponentSrc(t *testing.T) {
	registerRenderLocalComponent(t)

	inputDir := t.TempDir()
	inputFile := filepath.Join(inputDir, "report.ltml")
	if err := os.WriteFile(inputFile, []byte(`
<ltml xmlns:xt="rendercomponenttest">
  <page>
    <xt:card src="snippet.xml" />
  </page>
</ltml>`), 0o600); err != nil {
		t.Fatal(err)
	}

	assetsDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(assetsDir, "snippet.xml"), []byte("<p>from render local</p>"), 0o600); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := renderLocal(inputFile, assetsDir, nil, false, false, nil, &out, nil, nil); err != nil {
		t.Fatal(err)
	}
	if out.Len() == 0 {
		t.Fatal("expected PDF output")
	}

	renderLocalComponentMu.Lock()
	defer renderLocalComponentMu.Unlock()
	if got, want := renderLocalComponentBody, "<p>from render local</p>"; got != want {
		t.Fatalf("component body = %q, want %q", got, want)
	}
}

func TestRenderLocal_UADefaultEnablesTaggedPDF(t *testing.T) {
	inputFile := filepath.Join(t.TempDir(), "report.ltml")
	if err := os.WriteFile(inputFile, []byte(`
<ltml>
  <page>
    <p>Hello world</p>
  </page>
</ltml>`), 0o600); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := renderLocal(inputFile, "", nil, true, false, nil, &out, nil, nil); err != nil {
		t.Fatal(err)
	}

	pdfText := out.String()
	for _, fragment := range []string{"/StructTreeRoot", "/S /P", "/ToUnicode"} {
		if !strings.Contains(pdfText, fragment) {
			t.Fatalf("expected tagged PDF fragment %q in output:\n%s", fragment, pdfText)
		}
	}
	if strings.Contains(pdfText, "/ActualText (Hello world)") {
		t.Fatalf("rendered text should not be replaced at structure-element granularity:\n%s", pdfText)
	}
}

func TestRenderLocal_TraceFonts(t *testing.T) {
	inputFile := filepath.Join(t.TempDir(), "report.ltml")
	if err := os.WriteFile(inputFile, []byte(`
<ltml>
  <font id="body" name="Minimal" size="12" />
  <page font="body">
    <p>Hello</p>
    <p>Again</p>
  </page>
</ltml>`), 0o600); err != nil {
		t.Fatal(err)
	}

	fontDir := t.TempDir()
	fontBytes, err := os.ReadFile(filepath.Join("..", "..", "..", "ttf", "testdata", "minimal.ttf"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fontDir, "minimal.ttf"), fontBytes, 0o600); err != nil {
		t.Fatal(err)
	}

	var out, trace bytes.Buffer
	if err := renderLocal(inputFile, "", []string{fontDir}, false, true, nil, &out, &trace, nil); err != nil {
		t.Fatal(err)
	}
	got := trace.String()
	for _, want := range []string{`family="Minimal"`, `postscript="Minimal"`, `minimal.ttf`} {
		if !strings.Contains(got, want) {
			t.Fatalf("trace missing %q in:\n%s", want, got)
		}
	}
	if strings.Contains(got, "font cached") {
		t.Fatalf("trace should not include cache hits:\n%s", got)
	}
	if count := strings.Count(got, `font selected match=exact family="Minimal"`); count != 1 {
		t.Fatalf("selected trace count = %d, want 1:\n%s", count, got)
	}
}

func TestMain_ListFontsUsesOnlyConfiguredDirectories(t *testing.T) {
	first := t.TempDir()
	second := t.TempDir()
	fontBytes, err := os.ReadFile(filepath.Join("..", "..", "..", "ttf", "testdata", "minimal.ttf"))
	if err != nil {
		t.Fatal(err)
	}
	firstFont := filepath.Join(first, "first.ttf")
	secondFont := filepath.Join(second, "second.ttf")
	for _, target := range []string{firstFont, secondFont} {
		if err := os.WriteFile(target, fontBytes, 0o600); err != nil {
			t.Fatal(err)
		}
	}

	var output bytes.Buffer
	code := Main(context.Background(), []string{"-font-dir", first + ", " + second, "-list-fonts"}, &output, nil)
	if code != 0 {
		t.Fatalf("Main exit code = %d; output:\n%s", code, output.String())
	}
	got := output.String()
	if !strings.Contains(got, firstFont) || !strings.Contains(got, secondFont) {
		t.Fatalf("font catalog does not contain both configured directories:\n%s", got)
	}
	if strings.Contains(got, "/System/Library/Fonts") || strings.Contains(got, "/usr/share/fonts") {
		t.Fatalf("custom-only font catalog contains system fonts:\n%s", got)
	}
}

func TestRenderLocal_FontDirectoryOrderControlsSelection(t *testing.T) {
	inputFile := filepath.Join(t.TempDir(), "report.ltml")
	if err := os.WriteFile(inputFile, []byte(`<ltml><font id="body" name="Minimal" size="12"/><page font="body"><p>Hello</p></page></ltml>`), 0o600); err != nil {
		t.Fatal(err)
	}
	first := t.TempDir()
	second := t.TempDir()
	fontBytes, err := os.ReadFile(filepath.Join("..", "..", "..", "ttf", "testdata", "minimal.ttf"))
	if err != nil {
		t.Fatal(err)
	}
	firstFont := filepath.Join(first, "z-first.ttf")
	secondFont := filepath.Join(second, "a-second.ttf")
	for _, target := range []string{firstFont, secondFont} {
		if err := os.WriteFile(target, fontBytes, 0o600); err != nil {
			t.Fatal(err)
		}
	}

	var output, trace bytes.Buffer
	if err := renderLocal(inputFile, "", []string{first, second}, false, true, nil, &output, &trace, nil); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(trace.String(), firstFont) || strings.Contains(trace.String(), secondFont) {
		t.Fatalf("font trace did not select from the first directory:\n%s", trace.String())
	}
}

func TestMain_ListFontsDefaultMatchesAuto(t *testing.T) {
	var omitted, explicit bytes.Buffer
	if code := Main(context.Background(), []string{"-list-fonts"}, &omitted, nil); code != 0 {
		t.Fatalf("omitted -font-dir exit code = %d; output:\n%s", code, omitted.String())
	}
	if code := Main(context.Background(), []string{"-font-dir", "auto", "-list-fonts"}, &explicit, nil); code != 0 {
		t.Fatalf("explicit auto exit code = %d; output:\n%s", code, explicit.String())
	}
	if omitted.String() != explicit.String() {
		t.Fatal("omitted -font-dir catalog differs from explicit auto catalog")
	}
}

func TestMain_RejectsEmptyFontDirectoryEntry(t *testing.T) {
	dir := t.TempDir()
	var output bytes.Buffer
	code := Main(context.Background(), []string{"-font-dir", dir + ",,auto", "-list-fonts"}, &output, nil)
	if code != 2 || !strings.Contains(output.String(), "empty entry") {
		t.Fatalf("Main exit code = %d, output = %q; want usage error for empty entry", code, output.String())
	}
}

func TestMain_UAFlagEnablesTaggedPDF(t *testing.T) {
	inputFile := filepath.Join(t.TempDir(), "report.ltml")
	outputFile := filepath.Join(t.TempDir(), "report.pdf")
	if err := os.WriteFile(inputFile, []byte(`
<ltml>
  <page>
    <p>Hello world</p>
  </page>
</ltml>`), 0o600); err != nil {
		t.Fatal(err)
	}

	var stderr bytes.Buffer
	code := Main(context.Background(), []string{"-ua", "-o", outputFile, inputFile}, &stderr, nil)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %s", code, stderr.String())
	}

	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "/StructTreeRoot") {
		t.Fatalf("expected tagged PDF output, stderr = %s\n%s", stderr.String(), data)
	}
}

func TestLTMLUADefaultFromEnv(t *testing.T) {
	t.Setenv("LTML_UA", "true")
	got, err := ltmlUADefaultFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if !got {
		t.Fatal("LTML_UA=true parsed as false")
	}

	t.Setenv("LTML_UA", "0")
	got, err = ltmlUADefaultFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if got {
		t.Fatal("LTML_UA=0 parsed as true")
	}
}

func TestLTMLUADefaultFromEnvRejectsInvalidValue(t *testing.T) {
	t.Setenv("LTML_UA", "sure")
	if _, err := ltmlUADefaultFromEnv(); err == nil {
		t.Fatal("expected invalid LTML_UA value to fail")
	}
}

func TestBuildRemoteRequestBody_IncludesLTMLAndExtraFiles(t *testing.T) {
	inputDir := t.TempDir()
	inputFile := filepath.Join(inputDir, "report.ltml")
	ltmlBytes := []byte(`<ltml></ltml>`)
	if err := os.WriteFile(inputFile, ltmlBytes, 0o600); err != nil {
		t.Fatal(err)
	}

	extraDir := t.TempDir()
	logoFile := filepath.Join(extraDir, "logo.txt")
	iconFile := filepath.Join(extraDir, "icon.dat")
	if err := os.WriteFile(logoFile, []byte("logo-data"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(iconFile, []byte("icon-data"), 0o600); err != nil {
		t.Fatal(err)
	}

	body, contentType, err := buildRemoteRequestBody(inputFile, ltmlBytes, mustExtraAssets(t, logoFile, iconFile+":assets/icon.dat"))
	if err != nil {
		t.Fatal(err)
	}

	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		t.Fatal(err)
	}
	if mediaType != "multipart/form-data" {
		t.Fatalf("content type = %q, want multipart/form-data", mediaType)
	}

	mr := multipart.NewReader(bytes.NewReader(body), params["boundary"])

	part, err := mr.NextPart()
	if err != nil {
		t.Fatal(err)
	}
	if got := part.FormName(); got != "ltml" {
		t.Fatalf("first part form name = %q, want ltml", got)
	}
	if got := part.FileName(); got != "" {
		t.Fatalf("ltml part filename = %q, want empty", got)
	}
	if got := part.Header.Get("Content-Type"); got != "application/vnd.rowland.leadtype.ltml+xml" {
		t.Fatalf("ltml content type = %q", got)
	}
	data, err := io.ReadAll(part)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(ltmlBytes) {
		t.Fatalf("ltml bytes = %q, want %q", data, ltmlBytes)
	}

	part, err = mr.NextPart()
	if err != nil {
		t.Fatal(err)
	}
	if got := part.FormName(); got != "file" {
		t.Fatalf("second part form name = %q, want file", got)
	}
	if got := part.FileName(); got != "logo.txt" {
		t.Fatalf("second part filename = %q, want logo.txt", got)
	}
	data, err = io.ReadAll(part)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "logo-data" {
		t.Fatalf("second part body = %q", data)
	}

	part, err = mr.NextPart()
	if err != nil {
		t.Fatal(err)
	}
	if got := part.FormName(); got != "file" {
		t.Fatalf("third part form name = %q, want file", got)
	}
	if got := multipartFilenameParam(t, part); got != "assets/icon.dat" {
		t.Fatalf("third part filename = %q, want assets/icon.dat", got)
	}
	data, err = io.ReadAll(part)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "icon-data" {
		t.Fatalf("third part body = %q", data)
	}
}

func TestBuildRemoteRequestBody_RejectsDuplicateExtraVirtualPaths(t *testing.T) {
	extraRoot := t.TempDir()
	firstDir := filepath.Join(extraRoot, "one")
	secondDir := filepath.Join(extraRoot, "two")
	if err := os.MkdirAll(firstDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(secondDir, 0o700); err != nil {
		t.Fatal(err)
	}

	first := filepath.Join(firstDir, "logo.png")
	second := filepath.Join(secondDir, "logo.png")
	if err := os.WriteFile(first, []byte("a"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(second, []byte("b"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, _, err := buildRemoteRequestBody("report.ltml", []byte(`<ltml></ltml>`), mustExtraAssets(t, first+":assets/logo.png", second+":assets/logo.png"))
	if err == nil {
		t.Fatal("expected duplicate virtual path error")
	}
	if !strings.Contains(err.Error(), `duplicate -extra virtual path "assets/logo.png"`) {
		t.Fatalf("error = %v", err)
	}
}

func TestSubmitRemote_WritesResponseBody(t *testing.T) {
	inputFile := filepath.Join(t.TempDir(), "report.ltml")
	if err := os.WriteFile(inputFile, []byte(`<ltml></ltml>`), 0o600); err != nil {
		t.Fatal(err)
	}

	extraFile := filepath.Join(t.TempDir(), "logo.txt")
	if err := os.WriteFile(extraFile, []byte("logo-data"), 0o600); err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if got := r.URL.Path; got != "/render" {
			t.Fatalf("path = %s, want /render", got)
		}

		mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil {
			t.Fatal(err)
		}
		if mediaType != "multipart/form-data" {
			t.Fatalf("content type = %q", mediaType)
		}

		mr := multipart.NewReader(r.Body, params["boundary"])
		part, err := mr.NextPart()
		if err != nil {
			t.Fatal(err)
		}
		if got := part.FormName(); got != "ltml" {
			t.Fatalf("first part form name = %q", got)
		}
		data, err := io.ReadAll(part)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != `<ltml></ltml>` {
			t.Fatalf("ltml body = %q", data)
		}

		part, err = mr.NextPart()
		if err != nil {
			t.Fatal(err)
		}
		if got := part.FormName(); got != "file" {
			t.Fatalf("second part form name = %q", got)
		}
		if got := part.FileName(); got != "logo.txt" {
			t.Fatalf("second part filename = %q", got)
		}
		data, err = io.ReadAll(part)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != "logo-data" {
			t.Fatalf("file body = %q", data)
		}

		w.Header().Set("Content-Type", "application/pdf")
		_, _ = w.Write([]byte("%PDF-remote"))
	}))
	defer srv.Close()

	var out bytes.Buffer
	if err := submitRemote(inputFile, "", srv.URL+"/render", mustExtraAssets(t, extraFile), &out); err != nil {
		t.Fatal(err)
	}
	if got := out.String(); got != "%PDF-remote" {
		t.Fatalf("output = %q, want %%PDF-remote", got)
	}
}

func TestSubmitRemote_RejectsAssetsDir(t *testing.T) {
	inputFile := filepath.Join(t.TempDir(), "report.ltml")
	if err := os.WriteFile(inputFile, []byte(`<ltml></ltml>`), 0o600); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	err := submitRemote(inputFile, t.TempDir(), "http://example.com/render", nil, &out)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "does not support -assets") {
		t.Fatalf("error = %v", err)
	}
}

func TestSubmitRemote_SurfacesNon2xxResponse(t *testing.T) {
	inputFile := filepath.Join(t.TempDir(), "report.ltml")
	if err := os.WriteFile(inputFile, []byte(`<ltml></ltml>`), 0o600); err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "invalid LTML", http.StatusBadRequest)
	}))
	defer srv.Close()

	var out bytes.Buffer
	err := submitRemote(inputFile, "", srv.URL, nil, &out)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "400 Bad Request") || !strings.Contains(err.Error(), "invalid LTML") || !strings.Contains(err.Error(), inputFile) || !strings.Contains(err.Error(), srv.URL) {
		t.Fatalf("error = %v", err)
	}
}

func TestBuildRenderJobs_DefaultOutputPath(t *testing.T) {
	inputFile := filepath.Join(t.TempDir(), "report.ltml")
	jobs, err := buildRenderJobs([]string{inputFile}, "", false)
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 1 {
		t.Fatalf("jobs = %d, want 1", len(jobs))
	}
	if got, want := jobs[0].outputPath, strings.TrimSuffix(inputFile, ".ltml")+".pdf"; got != want {
		t.Fatalf("outputPath = %q, want %q", got, want)
	}
}

func TestDisplayPath_MakesWorkspacePathsRelative(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	inside := filepath.Join(cwd, "ltml", "samples", "test_007_flow_layout.ltml")
	if got, want := displayPath(inside), filepath.Join("ltml", "samples", "test_007_flow_layout.ltml"); got != want {
		t.Fatalf("displayPath(%q) = %q, want %q", inside, got, want)
	}
}

func TestNormalizeInterspersedArgs_AllowsFlagsAfterInputs(t *testing.T) {
	fs := flag.NewFlagSet("render-ltml", flag.ContinueOnError)
	var (
		batch     bool
		submitURL string
		output    string
	)
	fs.BoolVar(&batch, "b", false, "")
	fs.StringVar(&submitURL, "submit", "", "")
	fs.StringVar(&output, "o", "", "")

	args, err := normalizeInterspersedArgs(fs, []string{
		"-b",
		"ltml/samples/one.ltml",
		"ltml/samples/two.ltml",
		"-submit", "http://localhost:1969/render",
		"-o", "/tmp/out",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := fs.Parse(args); err != nil {
		t.Fatal(err)
	}

	if !batch {
		t.Fatal("batch = false, want true")
	}
	if submitURL != "http://localhost:1969/render" {
		t.Fatalf("submitURL = %q, want http://localhost:1969/render", submitURL)
	}
	if output != "/tmp/out" {
		t.Fatalf("output = %q, want /tmp/out", output)
	}
	if got, want := fs.Args(), []string{"ltml/samples/one.ltml", "ltml/samples/two.ltml"}; !equalStrings(got, want) {
		t.Fatalf("args = %#v, want %#v", got, want)
	}
}

func TestValidateArgs_ListFontsAcceptsNoInputs(t *testing.T) {
	if err := validateArgs(runConfig{listFonts: true}, nil); err != nil {
		t.Fatal(err)
	}
}

func TestValidateArgs_ListFontsRejectsInputs(t *testing.T) {
	err := validateArgs(runConfig{listFonts: true}, []string{"report.ltml"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "does not accept input files") {
		t.Fatalf("error = %v", err)
	}
}

func TestBuildRenderJobs_DefaultOutputRequiresExtension(t *testing.T) {
	inputFile := filepath.Join(t.TempDir(), "report")
	_, err := buildRenderJobs([]string{inputFile}, "", false)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "to have an extension") {
		t.Fatalf("error = %v", err)
	}
}

func TestBuildRenderJobs_SingleFileExplicitOutput(t *testing.T) {
	inputFile := filepath.Join(t.TempDir(), "report.ltml")
	outputFile := filepath.Join(t.TempDir(), "out.pdf")
	jobs, err := buildRenderJobs([]string{inputFile}, outputFile, false)
	if err != nil {
		t.Fatal(err)
	}
	if got := jobs[0].outputPath; got != outputFile {
		t.Fatalf("outputPath = %q, want %q", got, outputFile)
	}
}

func TestBuildRenderJobs_BatchOutputDirectory(t *testing.T) {
	root := t.TempDir()
	outputDir := filepath.Join(root, "out")
	if err := os.Mkdir(outputDir, 0o700); err != nil {
		t.Fatal(err)
	}
	first := filepath.Join(root, "one.ltml")
	second := filepath.Join(root, "two.ltml")

	jobs, err := buildRenderJobs([]string{first, second}, outputDir, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 2 {
		t.Fatalf("jobs = %d, want 2", len(jobs))
	}
	if got, want := jobs[0].outputPath, filepath.Join(outputDir, "one.pdf"); got != want {
		t.Fatalf("first outputPath = %q, want %q", got, want)
	}
	if got, want := jobs[1].outputPath, filepath.Join(outputDir, "two.pdf"); got != want {
		t.Fatalf("second outputPath = %q, want %q", got, want)
	}
}

func TestBuildRenderJobs_BatchOutputMustBeDirectory(t *testing.T) {
	root := t.TempDir()
	outputFile := filepath.Join(root, "out.pdf")
	if err := os.WriteFile(outputFile, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	inputFile := filepath.Join(root, "report.ltml")

	_, err := buildRenderJobs([]string{inputFile}, outputFile, true)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "existing directory") {
		t.Fatalf("error = %v", err)
	}
}

func TestRun_LocalModeDefaultOutput(t *testing.T) {
	inputFile := filepath.Join(t.TempDir(), "report.ltml")
	if err := os.WriteFile(inputFile, []byte(`<ltml></ltml>`), 0o600); err != nil {
		t.Fatal(err)
	}

	var log bytes.Buffer
	cfg := runConfig{pollInterval: time.Hour, stderr: &log}
	if err := run(context.Background(), cfg, []string{inputFile}); err != nil {
		t.Fatal(err)
	}

	outputFile := strings.TrimSuffix(inputFile, ".ltml") + ".pdf"
	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Fatal("expected rendered PDF bytes")
	}
	if got := log.String(); !strings.Contains(got, "rendering ") || !strings.Contains(got, "wrote ") {
		t.Fatalf("log output = %q, want render start and completion messages", got)
	}
}

func TestRun_LocalModeProfileWritesSummary(t *testing.T) {
	root := t.TempDir()
	inputFile := filepath.Join(root, "report.ltml")
	outputFile := filepath.Join(root, "report.pdf")
	if err := os.WriteFile(inputFile, []byte(`<ltml><page><p>Hello</p></page></ltml>`), 0o600); err != nil {
		t.Fatal(err)
	}

	var log bytes.Buffer
	code := Main(context.Background(), []string{"-profile", "-o", outputFile, inputFile}, &log, nil)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %s", code, log.String())
	}
	got := log.String()
	for _, want := range []string{"profile for", "leadtype profile:", "ltml.parse", "ltml.render_pass.final", "pdf.write"} {
		if !strings.Contains(got, want) {
			t.Fatalf("profile output = %q, missing %q", got, want)
		}
	}
}

func TestValidateArgs_ProfileSubmitRejected(t *testing.T) {
	err := validateArgs(runConfig{profile: true, submitURL: "http://example.test/render"}, []string{"input.ltml"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "local renders") {
		t.Fatalf("error = %v", err)
	}
}

func TestRun_BatchModeLocalRendersMultipleFiles(t *testing.T) {
	root := t.TempDir()
	first := filepath.Join(root, "one.ltml")
	second := filepath.Join(root, "two.ltml")
	for _, path := range []string{first, second} {
		if err := os.WriteFile(path, []byte(`<ltml></ltml>`), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	cfg := runConfig{batch: true, pollInterval: time.Hour}
	if err := run(context.Background(), cfg, []string{first, second}); err != nil {
		t.Fatal(err)
	}

	for _, outputFile := range []string{
		filepath.Join(root, "one.pdf"),
		filepath.Join(root, "two.pdf"),
	} {
		data, err := os.ReadFile(outputFile)
		if err != nil {
			t.Fatal(err)
		}
		if len(data) == 0 {
			t.Fatalf("expected rendered PDF bytes for %s", outputFile)
		}
	}
}

func TestRun_BatchModeContinuesAfterRenderError(t *testing.T) {
	root := t.TempDir()
	bad := filepath.Join(root, "bad.ltml")
	good := filepath.Join(root, "good.ltml")
	if err := os.WriteFile(bad, []byte(`<ltml`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(good, []byte(`<ltml></ltml>`), 0o600); err != nil {
		t.Fatal(err)
	}

	var log bytes.Buffer
	cfg := runConfig{batch: true, pollInterval: time.Hour, stderr: &log}
	err := run(context.Background(), cfg, []string{bad, good})
	if err == nil {
		t.Fatal("expected batch error")
	}
	if !strings.Contains(err.Error(), "batch completed with 1 render error") {
		t.Fatalf("error = %v", err)
	}

	outputFile := filepath.Join(root, "good.pdf")
	data, readErr := os.ReadFile(outputFile)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if len(data) == 0 {
		t.Fatal("expected rendered PDF bytes for good.ltml")
	}
	if got := log.String(); !strings.Contains(got, "good.pdf") || !strings.Contains(got, "XML syntax error") {
		t.Fatalf("log output = %q, want both good render and bad render messages", got)
	}
}

func TestRun_BatchModeSubmitWritesOneOutputPerInput(t *testing.T) {
	root := t.TempDir()
	outputDir := filepath.Join(root, "out")
	if err := os.Mkdir(outputDir, 0o700); err != nil {
		t.Fatal(err)
	}

	first := filepath.Join(root, "one.ltml")
	second := filepath.Join(root, "two.ltml")
	if err := os.WriteFile(first, []byte(`<ltml id="one"></ltml>`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(second, []byte(`<ltml id="two"></ltml>`), 0o600); err != nil {
		t.Fatal(err)
	}

	var mu sync.Mutex
	var seen []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil {
			t.Fatal(err)
		}
		if mediaType != "multipart/form-data" {
			t.Fatalf("content type = %q", mediaType)
		}
		mr := multipart.NewReader(r.Body, params["boundary"])
		part, err := mr.NextPart()
		if err != nil {
			t.Fatal(err)
		}
		body, err := io.ReadAll(part)
		if err != nil {
			t.Fatal(err)
		}
		mu.Lock()
		seen = append(seen, string(body))
		mu.Unlock()
		_, _ = w.Write([]byte("%PDF-remote"))
	}))
	defer srv.Close()

	cfg := runConfig{
		batch:        true,
		outputPath:   outputDir,
		submitURL:    srv.URL,
		pollInterval: time.Hour,
	}
	if err := run(context.Background(), cfg, []string{first, second}); err != nil {
		t.Fatal(err)
	}

	for _, outputFile := range []string{
		filepath.Join(outputDir, "one.pdf"),
		filepath.Join(outputDir, "two.pdf"),
	} {
		data, err := os.ReadFile(outputFile)
		if err != nil {
			t.Fatal(err)
		}
		if got := string(data); got != "%PDF-remote" {
			t.Fatalf("%s = %q, want %%PDF-remote", outputFile, got)
		}
	}

	mu.Lock()
	defer mu.Unlock()
	if len(seen) != 2 {
		t.Fatalf("requests = %d, want 2", len(seen))
	}
	if !slicesContain(seen, `<ltml id="one"></ltml>`) || !slicesContain(seen, `<ltml id="two"></ltml>`) {
		t.Fatalf("seen bodies = %#v", seen)
	}
}

func TestRun_RemoteModeStillRejectsAssetsDirInBatchMode(t *testing.T) {
	root := t.TempDir()
	first := filepath.Join(root, "one.ltml")
	second := filepath.Join(root, "two.ltml")
	for _, path := range []string{first, second} {
		if err := os.WriteFile(path, []byte(`<ltml></ltml>`), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	cfg := runConfig{
		assetsDir:    t.TempDir(),
		submitURL:    "http://example.com/render",
		batch:        true,
		pollInterval: time.Hour,
	}
	err := run(context.Background(), cfg, []string{first, second})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "does not support -assets") {
		t.Fatalf("error = %v", err)
	}
}

func TestWatchModeRerendersAfterInputChange(t *testing.T) {
	inputFile := filepath.Join(t.TempDir(), "report.ltml")
	if err := os.WriteFile(inputFile, []byte(`<ltml></ltml>`), 0o600); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	events := make(chan renderEvent, 4)
	errCh := make(chan error, 1)
	var log bytes.Buffer
	go func() {
		errCh <- run(ctx, runConfig{
			watch:        true,
			pollInterval: 20 * time.Millisecond,
			stderr:       &log,
			onRender: func(job renderJob, err error) {
				events <- renderEvent{job: job, err: err}
			},
		}, []string{inputFile})
	}()

	first := waitForRenderEvent(t, events)
	if first.err != nil {
		t.Fatalf("initial render error = %v", first.err)
	}

	touchFile(t, inputFile, []byte(`<ltml><page></page></ltml>`))
	second := waitForRenderEvent(t, events)
	if second.err != nil {
		t.Fatalf("rerender error = %v", second.err)
	}
	if second.job.inputPath != first.job.inputPath {
		t.Fatalf("rerender input = %q, want %q", second.job.inputPath, first.job.inputPath)
	}

	cancel()
	waitForWatchExit(t, errCh)
	if got := log.String(); !strings.Contains(got, "watching 1 input(s)") || !strings.Contains(got, "change detected in "+inputFile+"; rerendering") {
		t.Fatalf("log output = %q, want watch startup and input change messages", got)
	}
}

func TestWatchModeRerendersAfterExtraFileChange(t *testing.T) {
	root := t.TempDir()
	inputFile := filepath.Join(root, "report.ltml")
	extraFile := filepath.Join(root, "logo.txt")
	if err := os.WriteFile(inputFile, []byte(`<ltml></ltml>`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(extraFile, []byte("a"), 0o600); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	events := make(chan renderEvent, 4)
	errCh := make(chan error, 1)
	var log bytes.Buffer
	go func() {
		errCh <- run(ctx, runConfig{
			watch:        true,
			extraFiles:   mustExtraAssets(t, extraFile),
			pollInterval: 20 * time.Millisecond,
			stderr:       &log,
			onRender: func(job renderJob, err error) {
				events <- renderEvent{job: job, err: err}
			},
		}, []string{inputFile})
	}()

	if ev := waitForRenderEvent(t, events); ev.err != nil {
		t.Fatalf("initial render error = %v", ev.err)
	}

	touchFile(t, extraFile, []byte("b"))
	if ev := waitForRenderEvent(t, events); ev.err != nil {
		t.Fatalf("rerender error = %v", ev.err)
	}

	cancel()
	waitForWatchExit(t, errCh)
	if got := log.String(); !strings.Contains(got, "change detected in shared assets; rerendering all inputs") {
		t.Fatalf("log output = %q, want shared asset change message", got)
	}
}

func TestWatchModeRerendersAfterAssetsDirChange(t *testing.T) {
	root := t.TempDir()
	inputFile := filepath.Join(root, "report.ltml")
	assetsDir := filepath.Join(root, "assets")
	if err := os.Mkdir(assetsDir, 0o700); err != nil {
		t.Fatal(err)
	}
	assetFile := filepath.Join(assetsDir, "logo.txt")
	if err := os.WriteFile(inputFile, []byte(`<ltml></ltml>`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(assetFile, []byte("a"), 0o600); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	events := make(chan renderEvent, 4)
	errCh := make(chan error, 1)
	go func() {
		errCh <- run(ctx, runConfig{
			watch:        true,
			assetsDir:    assetsDir,
			pollInterval: 20 * time.Millisecond,
			onRender: func(job renderJob, err error) {
				events <- renderEvent{job: job, err: err}
			},
		}, []string{inputFile})
	}()

	if ev := waitForRenderEvent(t, events); ev.err != nil {
		t.Fatalf("initial render error = %v", ev.err)
	}

	touchFile(t, assetFile, []byte("b"))
	if ev := waitForRenderEvent(t, events); ev.err != nil {
		t.Fatalf("rerender error = %v", ev.err)
	}

	cancel()
	waitForWatchExit(t, errCh)
}

func TestWatchModeContinuesAfterFailedRender(t *testing.T) {
	inputFile := filepath.Join(t.TempDir(), "report.ltml")
	if err := os.WriteFile(inputFile, []byte(`<ltml`), 0o600); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	events := make(chan renderEvent, 4)
	errCh := make(chan error, 1)
	go func() {
		errCh <- run(ctx, runConfig{
			watch:        true,
			pollInterval: 20 * time.Millisecond,
			onRender: func(job renderJob, err error) {
				events <- renderEvent{job: job, err: err}
			},
		}, []string{inputFile})
	}()

	first := waitForRenderEvent(t, events)
	if first.err == nil {
		t.Fatal("expected initial render error")
	}

	touchFile(t, inputFile, []byte(`<ltml></ltml>`))
	second := waitForRenderEvent(t, events)
	if second.err != nil {
		t.Fatalf("rerender error = %v", second.err)
	}

	cancel()
	waitForWatchExit(t, errCh)
}

func TestWatchModeSubmitWorksForSingleAndBatchModes(t *testing.T) {
	root := t.TempDir()
	single := filepath.Join(root, "single.ltml")
	first := filepath.Join(root, "one.ltml")
	second := filepath.Join(root, "two.ltml")
	for _, path := range []string{single, first, second} {
		if err := os.WriteFile(path, []byte(`<ltml></ltml>`), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	var mu sync.Mutex
	counts := make(map[string]int)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil {
			t.Fatal(err)
		}
		if mediaType != "multipart/form-data" {
			t.Fatalf("content type = %q", mediaType)
		}

		mr := multipart.NewReader(r.Body, params["boundary"])
		part, err := mr.NextPart()
		if err != nil {
			t.Fatal(err)
		}
		body, err := io.ReadAll(part)
		if err != nil {
			t.Fatal(err)
		}
		mu.Lock()
		counts[string(body)]++
		mu.Unlock()
		_, _ = w.Write([]byte("%PDF-remote"))
	}))
	defer srv.Close()

	runWatch := func(batch bool, inputs []string, mutate func()) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		events := make(chan renderEvent, 8)
		errCh := make(chan error, 1)
		go func() {
			errCh <- run(ctx, runConfig{
				watch:        true,
				batch:        batch,
				submitURL:    srv.URL,
				pollInterval: 20 * time.Millisecond,
				onRender: func(job renderJob, err error) {
					events <- renderEvent{job: job, err: err}
				},
			}, inputs)
		}()

		for range inputs {
			if ev := waitForRenderEvent(t, events); ev.err != nil {
				t.Fatalf("initial render error = %v", ev.err)
			}
		}

		mutate()
		if ev := waitForRenderEvent(t, events); ev.err != nil {
			t.Fatalf("rerender error = %v", ev.err)
		}

		cancel()
		waitForWatchExit(t, errCh)
	}

	runWatch(false, []string{single}, func() {
		touchFile(t, single, []byte(`<ltml><page></page></ltml>`))
	})
	runWatch(true, []string{first, second}, func() {
		touchFile(t, second, []byte(`<ltml><page></page></ltml>`))
	})

	mu.Lock()
	defer mu.Unlock()
	if counts[`<ltml></ltml>`] != 3 {
		t.Fatalf("initial submit count = %d, want 3", counts[`<ltml></ltml>`])
	}
	if counts[`<ltml><page></page></ltml>`] != 2 {
		t.Fatalf("rerender submit count = %d, want 2", counts[`<ltml><page></page></ltml>`])
	}
}

type renderEvent struct {
	job renderJob
	err error
}

func waitForRenderEvent(t *testing.T, events <-chan renderEvent) renderEvent {
	t.Helper()

	select {
	case ev := <-events:
		return ev
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for render event")
		return renderEvent{}
	}
}

func waitForWatchExit(t *testing.T, errCh <-chan error) {
	t.Helper()

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, context.Canceled) {
			t.Fatalf("watch exited with error: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for watch exit")
	}
}

func touchFile(t *testing.T, path string, data []byte) {
	t.Helper()
	time.Sleep(30 * time.Millisecond)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
}

func slicesContain(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func equalStrings(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
