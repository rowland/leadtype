package ltml

import (
	"testing"

	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/pdf"
)

type compressionTestWriter struct {
	labelTestWriter
	compressPages              bool
	compressToUnicode          bool
	compressEmbeddedFonts      bool
	svgGradientStopOpacityMode pdf.SVGGradientStopOpacityMode
	svgBlendMode               pdf.SVGBlendMode
}

type pageOptionTestWriter struct {
	compressionTestWriter
	pageOptions []options.Options
}

func (w *pageOptionTestWriter) NewPageWithOptions(opts options.Options) {
	w.pageOptions = append(w.pageOptions, opts)
	w.NewPage()
}

func (w *compressionTestWriter) CompressPages(value bool) *pdf.DocWriter {
	w.compressPages = value
	return nil
}

func (w *compressionTestWriter) CompressToUnicode(value bool) *pdf.DocWriter {
	w.compressToUnicode = value
	return nil
}

func (w *compressionTestWriter) CompressEmbeddedFonts(value bool) *pdf.DocWriter {
	w.compressEmbeddedFonts = value
	return nil
}

func (w *compressionTestWriter) SetSVGGradientStopOpacityMode(value pdf.SVGGradientStopOpacityMode) pdf.SVGGradientStopOpacityMode {
	prev := w.svgGradientStopOpacityMode
	w.svgGradientStopOpacityMode = value
	return prev
}

func (w *compressionTestWriter) SetSVGBlendMode(value pdf.SVGBlendMode) pdf.SVGBlendMode {
	prev := w.svgBlendMode
	w.svgBlendMode = value
	return prev
}

func TestStdDocument_Print_AppliesCompressionAttrs(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml compress-pages="true" compress-to-unicode="true" compress-embedded-fonts="true">
  <page><label>Hello</label></page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}

	w := &compressionTestWriter{labelTestWriter: labelTestWriter{t: t}}
	if err := doc.Print(w); err != nil {
		t.Fatal(err)
	}
	if !w.compressPages || !w.compressToUnicode || !w.compressEmbeddedFonts {
		t.Fatalf("compression flags = pages:%t toUnicode:%t embedded:%t, want all true",
			w.compressPages, w.compressToUnicode, w.compressEmbeddedFonts)
	}
}

func TestStdDocument_Print_DefaultCompressionAttrsFalse(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml>
  <page><label>Hello</label></page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}

	w := &compressionTestWriter{labelTestWriter: labelTestWriter{t: t}}
	if err := doc.Print(w); err != nil {
		t.Fatal(err)
	}
	if w.compressPages || w.compressToUnicode || w.compressEmbeddedFonts {
		t.Fatalf("compression flags = pages:%t toUnicode:%t embedded:%t, want all false",
			w.compressPages, w.compressToUnicode, w.compressEmbeddedFonts)
	}
}

func TestStdDocument_Print_AppliesSVGGradientStopOpacityMode(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml svg-gradient-stop-opacity-mode="compatibility">
  <page><label>Hello</label></page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}

	w := &compressionTestWriter{labelTestWriter: labelTestWriter{t: t}}
	if err := doc.Print(w); err != nil {
		t.Fatal(err)
	}
	if w.svgGradientStopOpacityMode != pdf.SVGGradientStopOpacityModeCompatibility {
		t.Fatalf("svgGradientStopOpacityMode = %q, want compatibility", w.svgGradientStopOpacityMode)
	}
}

func TestStdDocument_Print_AppliesSVGBlendMode(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml svg-blend-mode="ignore">
  <page><label>Hello</label></page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}

	w := &compressionTestWriter{labelTestWriter: labelTestWriter{t: t}}
	if err := doc.Print(w); err != nil {
		t.Fatal(err)
	}
	if w.svgBlendMode != pdf.SVGBlendModeIgnore {
		t.Fatalf("svgBlendMode = %q, want ignore", w.svgBlendMode)
	}
}

func TestStdPage_Print_AppliesSVGRenderOptionsToPhysicalPage(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml>
  <page svg-gradient-stop-opacity-mode="compatibility" svg-blend-mode="ignore">
    <label>Hello</label>
  </page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}

	w := &pageOptionTestWriter{
		compressionTestWriter: compressionTestWriter{
			labelTestWriter: labelTestWriter{t: t},
		},
	}
	if err := doc.Print(w); err != nil {
		t.Fatal(err)
	}
	if len(w.pageOptions) != 1 {
		t.Fatalf("pageOptions count = %d, want 1", len(w.pageOptions))
	}
	opts := w.pageOptions[0]
	if got := opts[pdf.SVGGradientStopOpacityModeOption]; got != pdf.SVGGradientStopOpacityModeCompatibility {
		t.Fatalf("page SVG gradient stop opacity mode = %v, want compatibility", got)
	}
	if got := opts[pdf.SVGBlendModeOption]; got != pdf.SVGBlendModeIgnore {
		t.Fatalf("page SVG blend mode = %v, want ignore", got)
	}
}
