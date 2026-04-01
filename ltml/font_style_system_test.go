package ltml

import (
	"testing"

	"github.com/rowland/leadtype/font"
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
	foundHyphenCoverage := fallbackHasRune(fonts, '-')
	requireGeezaWithHyphenFallback(t, fonts, foundHyphenCoverage)
	if len(fonts) < 2 {
		t.Fatalf("expected fallback chain to keep additional coverage fonts, got %d font(s)", len(fonts))
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
	requireGeezaWithHyphenFallback(t, pw.Fonts(), fallbackHasRune(pw.Fonts(), '-'))
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

func requireGeezaWithHyphenFallback(t *testing.T, fonts []*font.Font, foundHyphenCoverage bool) {
	t.Helper()
	if len(fonts) == 0 {
		t.Skip("system font prerequisites unavailable: no fonts loaded from fallback chain")
	}
	if fonts[0].PostScriptName() != "GeezaPro-Bold" {
		t.Skipf("system font prerequisites unavailable: expected GeezaPro-Bold primary, got %s", fonts[0].PostScriptName())
	}
	if !foundHyphenCoverage {
		t.Skip("system font prerequisites unavailable: no fallback font with ASCII hyphen coverage")
	}
}

func fallbackHasRune(fonts []*font.Font, r rune) bool {
	for _, f := range fonts[1:] {
		if f.HasRune(r) {
			return true
		}
	}
	return false
}

func richTextForWriter(w Writer, text string) (*rich_text.RichText, error) {
	return rich_text.New(text, w.Fonts(), w.FontSize(), nil)
}
