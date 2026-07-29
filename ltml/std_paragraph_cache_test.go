package ltml

import (
	"testing"

	"github.com/rowland/leadtype/rich_text"
)

type paragraphCacheWriter struct {
	labelTestWriter
}

func (w *paragraphCacheWriter) SetLineSpacing(value float64) float64 {
	previous := w.lineSpacing
	w.lineSpacing = value
	return previous
}

type mutableDynamicInlineText struct {
	text string
}

func (t *mutableDynamicInlineText) Resolve(*StdDocument) string { return t.text }
func (t *mutableDynamicInlineText) Dynamic() bool               { return true }

func newParagraphCacheFixture(t *testing.T) (*StdDocument, *StdParagraph, *paragraphCacheWriter) {
	t.Helper()
	doc := &StdDocument{}
	doc.renderContext = newDocumentRenderContext(nil, false)
	paragraph := &StdParagraph{}
	paragraph.font = &FontStyle{
		entries:    []fontEntry{{name: "Helvetica"}},
		size:       12,
		lineHeight: 1.25,
	}
	if err := paragraph.SetContainer(doc); err != nil {
		t.Fatal(err)
	}
	writer := &paragraphCacheWriter{
		labelTestWriter: labelTestWriter{
			t:           t,
			fonts:       defaultTestFonts(t),
			lineSpacing: 1,
		},
	}
	return doc, paragraph, writer
}

func TestStdParagraph_LinesCachesStaticTextByParagraphAndWidth(t *testing.T) {
	doc, paragraph, writer := newParagraphCacheFixture(t)
	paragraph.AddText("A sufficiently long paragraph that wraps onto several lines.")

	first := paragraph.Lines(writer, 70)
	second := paragraph.Lines(writer, 70)
	if len(first) == 0 || len(second) == 0 || first[0] != second[0] {
		t.Fatal("same paragraph and width did not reuse cached wrapped lines")
	}
	if got := len(doc.renderContext.paragraphLines); got != 1 {
		t.Fatalf("cache entries = %d, want 1", got)
	}

	narrower := paragraph.Lines(writer, 45)
	if len(narrower) == 0 || narrower[0] == first[0] {
		t.Fatal("different width reused the first wrapped result")
	}
	if got := len(doc.renderContext.paragraphLines); got != 2 {
		t.Fatalf("cache entries = %d, want 2", got)
	}
}

func TestStdParagraph_LinesCacheIsScopedToRenderContext(t *testing.T) {
	doc, paragraph, writer := newParagraphCacheFixture(t)
	paragraph.AddText("A sufficiently long paragraph that wraps onto several lines.")

	first := paragraph.Lines(writer, 55)
	firstContext := doc.renderContext
	doc.renderContext = newDocumentRenderContext(nil, true)
	second := paragraph.Lines(writer, 55)

	if len(first) == 0 || len(second) == 0 || first[0] == second[0] {
		t.Fatal("new render context reused wrapped lines from the previous pass")
	}
	if got := len(firstContext.paragraphLines); got != 1 {
		t.Fatalf("first-pass cache entries = %d, want 1", got)
	}
	if got := len(doc.renderContext.paragraphLines); got != 1 {
		t.Fatalf("second-pass cache entries = %d, want 1", got)
	}
}

func TestStdParagraph_LinesCacheHitReappliesFontState(t *testing.T) {
	_, paragraph, writer := newParagraphCacheFixture(t)
	paragraph.AddText("Cached paragraph text")
	paragraph.Lines(writer, 200)

	writer.fontSize = 7
	writer.lineSpacing = 0.75
	paragraph.Lines(writer, 200)

	if got := writer.FontSize(); got != 12 {
		t.Fatalf("font size after cache hit = %v, want 12", got)
	}
	if got := writer.LineSpacing(); got != 1.25 {
		t.Fatalf("line spacing after cache hit = %v, want 1.25", got)
	}
}

func TestStdParagraph_LinesDoesNotCacheDynamicText(t *testing.T) {
	doc, paragraph, writer := newParagraphCacheFixture(t)
	dynamic := &mutableDynamicInlineText{text: "first"}
	paragraph.AddInlineWithFont(dynamic, paragraph.Font())

	first := paragraph.Lines(writer, 200)
	dynamic.text = "second"
	second := paragraph.Lines(writer, 200)

	if got := len(doc.renderContext.paragraphLines); got != 0 {
		t.Fatalf("dynamic paragraph cache entries = %d, want 0", got)
	}
	if first[0].String() != "first" || second[0].String() != "second" {
		t.Fatalf("dynamic lines = %q, %q; want first, second", first[0].String(), second[0].String())
	}
}

func TestStdParagraph_LinesSpecialLayoutsBypassOrdinaryCache(t *testing.T) {
	t.Run("split", func(t *testing.T) {
		doc, paragraph, writer := newParagraphCacheFixture(t)
		sentinel := []*rich_text.RichText{{}}
		paragraph.splitLines = sentinel

		if got := paragraph.Lines(writer, 100); len(got) != 1 || got[0] != sentinel[0] {
			t.Fatal("split paragraph did not return its precomputed lines")
		}
		if got := len(doc.renderContext.paragraphLines); got != 0 {
			t.Fatalf("split paragraph cache entries = %d, want 0", got)
		}
	})

	t.Run("sector", func(t *testing.T) {
		doc, paragraph, writer := newParagraphCacheFixture(t)
		sector := &StdSector{}
		if err := sector.StdWidget.SetContainer(doc); err != nil {
			t.Fatal(err)
		}
		if err := paragraph.SetContainer(sector); err != nil {
			t.Fatal(err)
		}
		sentinel := []*rich_text.RichText{{}}
		sector.paragraphLayouts = map[*StdParagraph]*sectorParagraphLayout{
			paragraph: {lines: sentinel},
		}

		if got := paragraph.Lines(writer, 100); len(got) != 1 || got[0] != sentinel[0] {
			t.Fatal("sector paragraph did not use its sector-owned layout")
		}
		if got := len(doc.renderContext.paragraphLines); got != 0 {
			t.Fatalf("sector paragraph cache entries = %d, want 0", got)
		}
	})

	t.Run("leader", func(t *testing.T) {
		doc, paragraph, writer := newParagraphCacheFixture(t)
		paragraph.textPieces = []textPiece{
			newStaticTextPiece("Chapter", paragraph.Font()),
			{content: &StdLeader{}, font: paragraph.Font()},
			newStaticTextPiece("12", paragraph.Font()),
		}

		if got := paragraph.Lines(writer, 200); len(got) == 0 {
			t.Fatal("leader paragraph returned no lines")
		}
		if got := len(doc.renderContext.paragraphLines); got != 0 {
			t.Fatalf("leader paragraph cache entries = %d, want 0", got)
		}
	})
}
