package ltpdf_test

import (
	"io/fs"
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
