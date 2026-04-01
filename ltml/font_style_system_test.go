package ltml

import (
	"testing"

	"github.com/rowland/leadtype/ltml/ltpdf"
	"github.com/rowland/leadtype/rich_text"
)

func TestFontStyleApply_SystemBoldFallbackChainDropsUnavailableFaces(t *testing.T) {
	pw := ltpdf.NewDocWriter()

	fs := &FontStyle{
		id: "subheading",
		entries: []fontEntry{
			{name: "Amiri"},
			{name: "Arial Unicode MS"},
			{name: "Geeza Pro"},
			{name: "Noto Naskh Arabic"},
			{name: "Noto Sans Arabic"},
		},
		size:   14,
		weight: "Bold",
	}

	fs.Apply(pw)
	fonts := pw.Fonts()
	if len(fonts) == 0 {
		t.Fatal("expected at least one loaded font")
	}
	if len(fonts) < 2 {
		t.Fatalf("expected fallback chain to keep additional coverage fonts, got %d font(s)", len(fonts))
	}
	if fonts[0].PostScriptName() != "GeezaPro-Bold" {
		t.Fatalf("expected bold Geeza primary, got %s", fonts[0].PostScriptName())
	}
	foundHyphenCoverage := false
	for _, f := range fonts[1:] {
		if f.HasRune('-') {
			foundHyphenCoverage = true
			break
		}
	}
	if !foundHyphenCoverage {
		t.Fatalf("expected at least one fallback font with ASCII hyphen coverage")
	}
}

func TestFontStyleApply_SystemBoldFallbackChainProducesSingleFontForArabicHeading(t *testing.T) {
	pw := ltpdf.NewDocWriter()

	fs := &FontStyle{
		id: "subheading",
		entries: []fontEntry{
			{name: "Amiri"},
			{name: "Arial Unicode MS"},
			{name: "Geeza Pro"},
			{name: "Noto Naskh Arabic"},
			{name: "Noto Sans Arabic"},
		},
		size:   14,
		weight: "Bold",
	}

	fs.Apply(pw)
	rt, err := richTextForWriter(pw, "مكتبة المدينة - ربيع عام ألفين وستة وعشرين")
	if err != nil {
		t.Fatalf("richTextForWriter: %v", err)
	}
	merged := rt.Merge()

	leafCount := 0
	merged.VisitAll(func(p *rich_text.RichText) {
		if p.IsLeaf() && p.Text != "" {
			leafCount++
			if p.Font == nil {
				t.Fatalf("leaf %q has nil font", p.Text)
			}
			if p.Text == "-" && p.Font.PostScriptName() == "GeezaPro-Bold" {
				t.Fatalf("expected hyphen to fall back from GeezaPro-Bold")
			}
		}
	})
	if leafCount < 3 {
		t.Fatalf("expected Arabic heading with unsupported hyphen to split into multiple leaves, got %d", leafCount)
	}
}

func richTextForWriter(w Writer, text string) (*rich_text.RichText, error) {
	return rich_text.New(text, w.Fonts(), w.FontSize(), nil)
}
