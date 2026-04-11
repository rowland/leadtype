// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"testing/fstest"
)

type testComponent struct {
	StdComponent
}

var registerTestComponentOnce sync.Once

func registerTestComponentTag(t *testing.T) {
	t.Helper()
	registerTestComponentOnce.Do(func() {
		if err := RegisterTagExt("componenttest", "card", func() any { return &testComponent{} }); err != nil {
			t.Fatalf("register component tag: %v", err)
		}
	})
}

func parsedTestComponent(t *testing.T, input string, assetFS fstest.MapFS) (*Doc, *testComponent) {
	t.Helper()
	registerTestComponentTag(t)

	var opts []ParseOption
	if assetFS != nil {
		opts = append(opts, WithAssetFS(assetFS))
	}
	doc, err := Parse([]byte(input), opts...)
	if err != nil {
		t.Fatal(err)
	}

	page := doc.Page(0)
	if page == nil {
		t.Fatal("document has no page")
	}
	component, ok := page.Widgets()[0].(*testComponent)
	if !ok {
		t.Fatalf("first child = %T, want *testComponent", page.Widgets()[0])
	}
	return doc, component
}

func TestComponent_CapturesInnerXMLAndPreservesSiblingParsing(t *testing.T) {
	doc, component := parsedTestComponent(t, `
<ltml xmlns:xt="componenttest">
  <page>
    <xt:card width="144">
      <p>Hello</p><!-- comment -->
    </xt:card>
    <div id="after" />
  </page>
</ltml>`, nil)

	if got, want := component.Body(), `
      <p>Hello</p><!-- comment -->
    `; got != want {
		t.Fatalf("body = %q, want %q", got, want)
	}
	if !component.WidthIsSet() || component.Width() != 144 {
		t.Fatalf("width = %v set=%v, want 144 and set", component.Width(), component.WidthIsSet())
	}
	if component.Container() == nil {
		t.Fatal("component container was not set")
	}
	if component.Scope() == nil {
		t.Fatal("component scope was not set")
	}
	page := doc.Page(0)
	if len(page.Widgets()) != 2 {
		t.Fatalf("page child count = %d, want 2", len(page.Widgets()))
	}
	if after, ok := page.Widgets()[1].(*StdContainer); !ok || after.ID != "after" {
		t.Fatalf("second child = %#v, want StdContainer with id=after", page.Widgets()[1])
	}
}

func TestComponent_SrcLoadsFromDocAssetFSAndIgnoresInlineBody(t *testing.T) {
	_, component := parsedTestComponent(t, `
<ltml xmlns:xt="componenttest">
  <page>
    <xt:card src="snippet.xml">
      <p>ignored</p>
      <!-- ignored -->
    </xt:card>
  </page>
</ltml>`, fstest.MapFS{
		"snippet.xml": &fstest.MapFile{Data: []byte("<p>from asset fs</p>")},
	})

	if got, want := component.Body(), "<p>from asset fs</p>"; got != want {
		t.Fatalf("body = %q, want %q", got, want)
	}
}

func TestComponent_SrcRejectsInvalidAssetPath(t *testing.T) {
	registerTestComponentTag(t)

	_, err := Parse([]byte(`
<ltml xmlns:xt="componenttest">
  <page>
    <xt:card src="../snippet.xml" />
  </page>
</ltml>`), WithAssetFS(fstest.MapFS{
		"snippet.xml": &fstest.MapFile{Data: []byte("unused")},
	}))
	if err == nil {
		t.Fatal("expected invalid asset path error")
	}
}

func TestComponent_NetworkAssetsRequireDocumentOptIn(t *testing.T) {
	registerTestComponentTag(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("<p>remote body</p>"))
	}))
	defer srv.Close()

	_, err := Parse([]byte(`
<ltml xmlns:xt="componenttest">
  <page>
    <xt:card src="` + srv.URL + `" />
  </page>
</ltml>`))
	if err == nil {
		t.Fatal("expected network-assets disabled error")
	}

	_, component := parsedTestComponent(t, `
<ltml xmlns:xt="componenttest" network-assets="true">
  <page>
    <xt:card src="`+srv.URL+`" />
  </page>
</ltml>`, nil)
	if got, want := component.Body(), "<p>remote body</p>"; got != want {
		t.Fatalf("body = %q, want %q", got, want)
	}
}

func TestComponent_DisableNetworkAssetsOverridesDocumentSetting(t *testing.T) {
	registerTestComponentTag(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("<p>remote body</p>"))
	}))
	defer srv.Close()

	networkAssetsDisabled.Store(false)
	t.Cleanup(func() {
		networkAssetsDisabled.Store(false)
	})
	DisableNetworkAssets()

	_, err := Parse([]byte(`
<ltml xmlns:xt="componenttest" network-assets="true">
  <page>
    <xt:card src="` + srv.URL + `" />
  </page>
</ltml>`))
	if err == nil {
		t.Fatal("expected global network disable error")
	}
}

func TestComponent_SrcLoadsRelativeToParsedFile(t *testing.T) {
	registerTestComponentTag(t)

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "snippet.xml"), []byte("<p>from relative file</p>"), 0o644); err != nil {
		t.Fatal(err)
	}
	inputPath := filepath.Join(dir, "doc.ltml")
	if err := os.WriteFile(inputPath, []byte(`
<ltml xmlns:xt="componenttest">
  <page>
    <xt:card src="snippet.xml" />
  </page>
</ltml>`), 0o644); err != nil {
		t.Fatal(err)
	}

	doc, err := ParseFile(inputPath)
	if err != nil {
		t.Fatal(err)
	}

	page := doc.Page(0)
	component, ok := page.Widgets()[0].(*testComponent)
	if !ok {
		t.Fatalf("first child = %T, want *testComponent", page.Widgets()[0])
	}
	if got, want := component.Body(), "<p>from relative file</p>"; got != want {
		t.Fatalf("body = %q, want %q", got, want)
	}
}

func TestParse_RejectsMultipleTopLevelRoots(t *testing.T) {
	_, err := Parse([]byte(`<ltml></ltml><ltml></ltml>`))
	if err == nil {
		t.Fatal("expected multiple root parse error")
	}
}
