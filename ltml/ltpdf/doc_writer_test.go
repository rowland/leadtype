package ltpdf_test

import (
	"bytes"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/rowland/leadtype/ltml"
	"github.com/rowland/leadtype/ltml/ltpdf"
	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/pdf"
	"github.com/rowland/leadtype/ttf_fonts"
)

var _ ltml.Writer = (*ltpdf.DocWriter)(nil)

func TestDocWriter_LayoutProbeWriterPreservesFontSourcesAndAssetFS(t *testing.T) {
	customFonts, err := ttf_fonts.New("../../ttf/testdata/minimal.ttf")
	if err != nil {
		t.Fatalf("loading custom font source: %v", err)
	}

	assetFS := fstest.MapFS{
		"logo.txt": &fstest.MapFile{Data: []byte("asset")},
	}

	base := &ltpdf.DocWriter{DocWriter: pdf.NewDocWriter()}
	base.AddFontSource(customFonts)
	base.SetAssetFS(assetFS)
	base.EnableTaggedPDF(true)

	probe, ok := base.LayoutProbeWriter().(*ltpdf.DocWriter)
	if !ok {
		t.Fatalf("LayoutProbeWriter type = %T, want *ltpdf.DocWriter", base.LayoutProbeWriter())
	}

	if _, err := probe.SetFont("Minimal", 12, options.Options{}); err != nil {
		t.Fatalf("probe lost custom font source: %v", err)
	}
	if !probe.TaggedPDFEnabled() {
		t.Fatal("probe should preserve tagged PDF state")
	}
	data, err := fs.ReadFile(probe.AssetFS(), "logo.txt")
	if err != nil {
		t.Fatalf("probe asset fs read failed: %v", err)
	}
	if string(data) != "asset" {
		t.Fatalf("probe asset fs content = %q, want %q", data, "asset")
	}
}

func TestNewDocWriterWithFontDirsUsesOnlyExplicitDirectories(t *testing.T) {
	w, err := ltpdf.NewDocWriterWithFontDirs(nil)
	if err != nil {
		t.Fatal(err)
	}
	sources := w.FontSources()
	if len(sources) == 0 {
		t.Fatal("writer has no TrueType font source")
	}
	ttFonts, ok := sources[0].(*ttf_fonts.TtfFonts)
	if !ok {
		t.Fatalf("first font source = %T, want *ttf_fonts.TtfFonts", sources[0])
	}
	if ttFonts.Len() != 0 {
		t.Fatalf("explicit empty directory list loaded %d fonts", ttFonts.Len())
	}
}

func TestNewDocWriterUsesSystemFontDirectories(t *testing.T) {
	w := ltpdf.NewDocWriter()
	sources := w.FontSources()
	if len(sources) == 0 {
		t.Fatal("writer has no TrueType font source")
	}
	ttFonts, ok := sources[0].(*ttf_fonts.TtfFonts)
	if !ok {
		t.Fatalf("first font source = %T, want *ttf_fonts.TtfFonts", sources[0])
	}
	if ttFonts.Len() == 0 {
		t.Skip("no fonts found in this host's system font directories")
	}
}

func TestDocWriter_Print_UsesLTMLPageStyleForPhysicalPageSize(t *testing.T) {
	doc, err := ltml.Parse([]byte(`
<ltml>
  <pagestyle id="tiny" width="200" height="300" />
  <page style="tiny" />
</ltml>`))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	w := &ltpdf.DocWriter{DocWriter: pdf.NewDocWriter()}
	if err := doc.Print(w); err != nil {
		t.Fatalf("Print: %v", err)
	}

	var buf bytes.Buffer
	if _, err := w.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}

	pdfText := buf.String()
	if !strings.Contains(pdfText, "/MediaBox [0 0 200 300 ]") {
		t.Fatalf("PDF missing custom MediaBox, got:\n%s", pdfText)
	}
	if !strings.Contains(pdfText, "/CropBox [0 0 200 300 ]") {
		t.Fatalf("PDF missing custom CropBox, got:\n%s", pdfText)
	}
}
