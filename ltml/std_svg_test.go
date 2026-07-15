package ltml

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

type svgTestWriter struct {
	labelTestWriter
	fileDimensions       map[string][2]int
	inlineDimensions     [2]int
	inlineDimensionCalls []string
	inlineCalls          []svgInlinePrintCall
	fileCalls            []svgFilePrintCall
}

type svgInlinePrintCall struct {
	body   string
	x      float64
	y      float64
	width  *float64
	height *float64
}

type svgFilePrintCall struct {
	filename string
	x        float64
	y        float64
	width    *float64
	height   *float64
}

func (w *svgTestWriter) SVGDimensions(data []byte) (width, height int, err error) {
	w.inlineDimensionCalls = append(w.inlineDimensionCalls, string(data))
	if w.inlineDimensions != [2]int{} {
		return w.inlineDimensions[0], w.inlineDimensions[1], nil
	}
	return 0, 0, nil
}

func (w *svgTestWriter) SVGDimensionsFromFile(filename string) (width, height int, err error) {
	if dims, ok := w.fileDimensions[filename]; ok {
		return dims[0], dims[1], nil
	}
	return 0, 0, nil
}

func (w *svgTestWriter) PrintSVG(data []byte, x, y float64, width, height *float64) (actualWidth, actualHeight float64, err error) {
	w.inlineCalls = append(w.inlineCalls, svgInlinePrintCall{
		body:   string(data),
		x:      x,
		y:      y,
		width:  width,
		height: height,
	})
	return 0, 0, nil
}

func (w *svgTestWriter) PrintSVGFile(filename string, x, y float64, width, height *float64) (actualWidth, actualHeight float64, err error) {
	w.fileCalls = append(w.fileCalls, svgFilePrintCall{
		filename: filename,
		x:        x,
		y:        y,
		width:    width,
		height:   height,
	})
	return 0, 0, nil
}

func TestStdSVG_PreferredWidthAndHeight_UseIntrinsicSize(t *testing.T) {
	svg := &StdSVG{}
	svg.body = `<svg xmlns="http://www.w3.org/2000/svg" width="160" height="100"></svg>`
	w := &svgTestWriter{inlineDimensions: [2]int{160, 100}}

	if got := mustPreferredWidth(t, svg, w); got != 160 {
		t.Fatalf("PreferredWidth() = %v, want 160", got)
	}
	if got := mustPreferredHeight(t, svg, w); got != 100 {
		t.Fatalf("PreferredHeight() = %v, want 100", got)
	}
}

func TestStdSVG_PreferredHeight_InfersAspectRatioFromWidth(t *testing.T) {
	svg := &StdSVG{}
	svg.body = `<svg xmlns="http://www.w3.org/2000/svg" width="160" height="100"></svg>`
	svg.SetWidth(80)
	w := &svgTestWriter{inlineDimensions: [2]int{160, 100}}

	if got := mustPreferredHeight(t, svg, w); got != 50 {
		t.Fatalf("PreferredHeight() = %v, want 50", got)
	}
}

func TestStdSVG_PreferredWidth_InfersAspectRatioFromHeight(t *testing.T) {
	svg := &StdSVG{}
	svg.body = `<svg xmlns="http://www.w3.org/2000/svg" width="80" height="40"></svg>`
	svg.SetHeight(30)
	w := &svgTestWriter{inlineDimensions: [2]int{80, 40}}

	if got := mustPreferredWidth(t, svg, w); got != 60 {
		t.Fatalf("PreferredWidth() = %v, want 60", got)
	}
}

func TestStdSVG_MaxHeightFitsAspectRatio(t *testing.T) {
	svg := &StdSVG{}
	svg.body = `<svg xmlns="http://www.w3.org/2000/svg" width="38" height="1080"></svg>`
	svg.SetMaxHeight(100)
	w := &svgTestWriter{inlineDimensions: [2]int{38, 1080}}

	wantWidth := 100.0 * 38.0 / 1080.0
	if got := mustPreferredHeight(t, svg, w); got != 100 {
		t.Fatalf("PreferredHeight() = %v, want 100", got)
	}
	if got := mustPreferredWidth(t, svg, w); got < wantWidth-0.001 || got > wantWidth+0.001 {
		t.Fatalf("PreferredWidth() = %v, want approx %v", got, wantWidth)
	}
}

func TestStdSVG_DrawContent_FitsResolvedCellWithoutStretching(t *testing.T) {
	svg := &StdSVG{}
	svg.body = `<svg xmlns="http://www.w3.org/2000/svg" width="38" height="1080"></svg>`
	svg.ResolveWidth(200)
	svg.ResolveHeight(100)
	w := &svgTestWriter{inlineDimensions: [2]int{38, 1080}}

	if err := svg.DrawContent(w); err != nil {
		t.Fatal(err)
	}
	if len(w.inlineCalls) != 1 {
		t.Fatalf("inline call count = %d, want 1", len(w.inlineCalls))
	}
	call := w.inlineCalls[0]
	wantWidth := 100.0 * 38.0 / 1080.0
	if call.width == nil || *call.width < wantWidth-0.001 || *call.width > wantWidth+0.001 {
		t.Fatalf("width = %v, want approx %v", call.width, wantWidth)
	}
	if call.height == nil || *call.height != 100 {
		t.Fatalf("height = %v, want 100", call.height)
	}
}

func TestStdSVG_DrawContent_UsesContentBoxDimensionsForInlineSVG(t *testing.T) {
	svg := &StdSVG{}
	svg.body = `<svg xmlns="http://www.w3.org/2000/svg" width="80" height="40"></svg>`
	svg.SetLeft(10)
	svg.SetTop(20)
	svg.SetWidth(120)
	svg.SetHeight(60)
	svg.padding.SetAll("6", "pt")
	w := &svgTestWriter{}

	if err := svg.DrawContent(w); err != nil {
		t.Fatal(err)
	}
	if len(w.inlineCalls) != 1 {
		t.Fatalf("inline call count = %d, want 1", len(w.inlineCalls))
	}
	if len(w.fileCalls) != 0 {
		t.Fatalf("file call count = %d, want 0", len(w.fileCalls))
	}
	call := w.inlineCalls[0]
	if call.x != 16 || call.y != 26 {
		t.Fatalf("x,y = %v,%v, want 16,26", call.x, call.y)
	}
	if call.width == nil || *call.width != 108 {
		t.Fatalf("width = %v, want 108", call.width)
	}
	if call.height == nil || *call.height != 48 {
		t.Fatalf("height = %v, want 48", call.height)
	}
	if call.body != svg.body {
		t.Fatalf("body = %q, want %q", call.body, svg.body)
	}
}

func TestStdSVG_DrawContent_InjectsStyleIntoInlineSVG(t *testing.T) {
	svg := &StdSVG{}
	svg.SetAttrs(map[string]string{
		"style": ".accent { fill: #f6d44e; stroke: #222222; }",
	})
	svg.body = `<svg xmlns="http://www.w3.org/2000/svg" width="80" height="40"><path class="accent" d="M0 0 L10 0 L10 10 Z"/></svg>`
	w := &svgTestWriter{inlineDimensions: [2]int{80, 40}}

	if err := svg.DrawContent(w); err != nil {
		t.Fatal(err)
	}
	if len(w.inlineCalls) != 1 {
		t.Fatalf("inline call count = %d, want 1", len(w.inlineCalls))
	}
	body := w.inlineCalls[0].body
	if !strings.Contains(body, `<style><![CDATA[.accent { fill: #f6d44e; stroke: #222222; }]]></style>`) {
		t.Fatalf("styled body = %q, want injected style element", body)
	}
	if !strings.Contains(body, `><style`) {
		t.Fatalf("styled body = %q, want style inserted after svg start tag", body)
	}
	if len(w.inlineDimensionCalls) == 0 || !strings.Contains(w.inlineDimensionCalls[0], ".accent") {
		t.Fatalf("dimension body = %#v, want injected style", w.inlineDimensionCalls)
	}
}

func TestStdSVG_DrawContent_UsesFilePathWhenSrcIsSet(t *testing.T) {
	svg := &StdSVG{}
	svg.SetDoc(newDocWithOptions(WithAssetFS(testingMapFS("fixture.svg", `<svg xmlns="http://www.w3.org/2000/svg"></svg>`))))
	svg.source.src = "fixture.svg"
	svg.source.srcExplicit = true
	svg.SetLeft(10)
	svg.SetTop(20)
	w := &svgTestWriter{}

	if err := svg.DrawContent(w); err != nil {
		t.Fatal(err)
	}
	if len(w.fileCalls) != 1 {
		t.Fatalf("file call count = %d, want 1", len(w.fileCalls))
	}
	if len(w.inlineCalls) != 0 {
		t.Fatalf("inline call count = %d, want 0", len(w.inlineCalls))
	}
	if got := w.fileCalls[0].filename; got != "fixture.svg" {
		t.Fatalf("filename = %q, want fixture.svg", got)
	}
}

func TestStdSVG_DrawContent_InjectsStyleIntoSrcSVG(t *testing.T) {
	svg := &StdSVG{}
	svg.SetDoc(newDocWithOptions(WithAssetFS(testingMapFS("fixture.svg", `<svg xmlns="http://www.w3.org/2000/svg" width="80" height="40"><path class="accent" d="M0 0 L10 0 L10 10 Z"/></svg>`))))
	svg.SetAttrs(map[string]string{
		"src":   "fixture.svg",
		"style": ".accent { fill: #f6d44e; stroke: #222222; }",
	})
	w := &svgTestWriter{inlineDimensions: [2]int{80, 40}}

	if err := svg.DrawContent(w); err != nil {
		t.Fatal(err)
	}
	if len(w.fileCalls) != 0 {
		t.Fatalf("file call count = %d, want 0", len(w.fileCalls))
	}
	if len(w.inlineCalls) != 1 {
		t.Fatalf("inline call count = %d, want 1", len(w.inlineCalls))
	}
	body := w.inlineCalls[0].body
	if !strings.Contains(body, ".accent { fill: #f6d44e; stroke: #222222; }") {
		t.Fatalf("styled body = %q, want injected style", body)
	}
	if !strings.Contains(body, `class="accent"`) {
		t.Fatalf("styled body = %q, want original SVG content preserved", body)
	}
}

func TestInjectSVGStyle_DifferentStylesProduceDifferentSVGBytes(t *testing.T) {
	data := []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="10" height="10"><rect class="accent" width="10" height="10"/></svg>`)

	blue := string(injectSVGStyle(data, ".accent { fill: blue; }"))
	gold := string(injectSVGStyle(data, ".accent { fill: gold; }"))

	if blue == gold {
		t.Fatal("styled SVG bytes should differ for different injected styles")
	}
	if !strings.Contains(blue, ".accent { fill: blue; }") {
		t.Fatalf("blue styled body = %q, want blue style", blue)
	}
	if !strings.Contains(gold, ".accent { fill: gold; }") {
		t.Fatalf("gold styled body = %q, want gold style", gold)
	}
}

func TestInjectSVGStyle_ExpandsSelfClosingSVGRoot(t *testing.T) {
	styled := string(injectSVGStyle(
		[]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="10" height="10"/>`),
		".accent { fill: gold; }",
	))

	if strings.Contains(styled, "/><style") {
		t.Fatalf("styled body = %q, want expanded svg root", styled)
	}
	if !strings.Contains(styled, `><style><![CDATA[.accent { fill: gold; }]]></style></svg>`) {
		t.Fatalf("styled body = %q, want style inside svg root", styled)
	}
}

func TestParse_SVGTag_CapturesInnerXMLAndPreservesSiblingParsing(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml>
  <page>
    <svg width="2in">
      <svg xmlns="http://www.w3.org/2000/svg" width="80" height="40">
        <rect width="80" height="40" fill="#abcdef"/>
      </svg>
    </svg>
    <div id="after" />
  </page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}

	page := doc.Root().Page(0)
	if page == nil {
		t.Fatal("page is nil")
	}
	if len(page.children) != 2 {
		t.Fatalf("child count = %d, want 2", len(page.children))
	}
	svg, ok := page.children[0].(*StdSVG)
	if !ok {
		t.Fatalf("child type = %T, want *StdSVG", page.children[0])
	}
	if svg.Width() != 144 {
		t.Fatalf("width = %v, want 144", svg.Width())
	}
	if got := svg.Body(); got == "" {
		t.Fatal("expected inline svg body to be captured")
	}
	if after, ok := page.children[1].(*StdContainer); !ok || after.ID != "after" {
		t.Fatalf("second child = %#v, want StdContainer with id=after", page.children[1])
	}
}

func TestParse_SVGTag_SrcOverridesInlineBody(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml>
  <page>
    <svg src="logo.svg">
      <svg xmlns="http://www.w3.org/2000/svg" width="10" height="10"></svg>
    </svg>
  </page>
</ltml>`), WithAssetFS(fstest.MapFS{
		"logo.svg": &fstest.MapFile{Data: []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="10" height="10"></svg>`)},
	}))
	if err != nil {
		t.Fatal(err)
	}

	page := doc.Root().Page(0)
	svg, ok := page.children[0].(*StdSVG)
	if !ok {
		t.Fatalf("child type = %T, want *StdSVG", page.children[0])
	}
	if svg.source.src != "logo.svg" {
		t.Fatalf("src = %q, want logo.svg", svg.source.src)
	}
	if got := svg.Body(); got == "" {
		t.Fatal("expected src-loaded body to be retained")
	}
}

func TestParse_SVGTag_NetworkSrcLoadsBodyWhenEnabled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="10" height="10"></svg>`))
	}))
	defer srv.Close()

	doc, err := Parse([]byte(`
<ltml network-assets="true">
  <page>
    <svg src="` + srv.URL + `" />
  </page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}

	page := doc.Root().Page(0)
	svg, ok := page.children[0].(*StdSVG)
	if !ok {
		t.Fatalf("child type = %T, want *StdSVG", page.children[0])
	}
	ref1, err := svg.assetSource()
	if err != nil {
		t.Fatal(err)
	}
	ref2, err := svg.assetSource()
	if err != nil {
		t.Fatal(err)
	}
	if ref1.kind != assetSourceNetworkTempFile || ref1.identifier == "" {
		t.Fatalf("ref1 = %#v, want network temp file", ref1)
	}
	if ref1.identifier != ref2.identifier {
		t.Fatalf("temp file should be reused, got %q then %q", ref1.identifier, ref2.identifier)
	}
	if got := svg.Body(); got == "" {
		t.Fatal("expected network-loaded body to be retained")
	}
}
