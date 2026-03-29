// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
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
)

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

	assetFS, cleanup, err := buildOptionalAssetFS(assetsDir, []string{extraFile})
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

	body, contentType, err := buildRemoteRequestBody(inputFile, ltmlBytes, []string{logoFile, iconFile})
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
	if got := part.FileName(); got != "icon.dat" {
		t.Fatalf("third part filename = %q, want icon.dat", got)
	}
	data, err = io.ReadAll(part)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "icon-data" {
		t.Fatalf("third part body = %q", data)
	}
}

func TestBuildRemoteRequestBody_RejectsDuplicateExtraBaseNames(t *testing.T) {
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

	_, _, err := buildRemoteRequestBody("report.ltml", []byte(`<ltml></ltml>`), []string{first, second})
	if err == nil {
		t.Fatal("expected duplicate base name error")
	}
	if !strings.Contains(err.Error(), `duplicate -extra base name "logo.png"`) {
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
	if err := submitRemote(inputFile, "", srv.URL+"/render", []string{extraFile}, &out); err != nil {
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
			extraFiles:   []string{extraFile},
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
