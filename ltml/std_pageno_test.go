package ltml

import (
	"testing"

	"github.com/rowland/leadtype/ltml/ltpdf"
	"github.com/rowland/leadtype/rich_text"
)

func TestStdPageNo_SetContainer_RejectsUnsupportedParents(t *testing.T) {
	var pageno StdPageNo
	if err := pageno.SetContainer(&StdContainer{}); err == nil {
		t.Fatal("expected unsupported parent error")
	}
}

func TestStdPageNo_ParseAndInlineParents(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml>
  <page>
    <p>Page <pageno start="10" reset="1" hidden="true" font.weight="Bold" /></p>
    <label>Label <pageno /></label>
    <p><span>Nested <pageno /></span></p>
  </page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}

	page := doc.ltmls[0].Page(0)
	if len(page.children) != 3 {
		t.Fatalf("child count = %d, want 3", len(page.children))
	}
	p := page.children[0].(*StdParagraph)
	if len(p.textPieces) != 2 {
		t.Fatalf("paragraph pieces = %d, want 2", len(p.textPieces))
	}
	pageNoText, ok := p.textPieces[1].content.(*StdPageNo)
	if !ok {
		t.Fatalf("second paragraph piece = %T, want *StdPageNo", p.textPieces[1].content)
	}
	if !pageNoText.hidden || pageNoText.start != 10 || pageNoText.reset != 1 {
		t.Fatalf("pageno attrs = %#v", pageNoText)
	}
	if pageNoText.Font().weight != "Bold" {
		t.Fatalf("pageno font weight = %q, want Bold", pageNoText.Font().weight)
	}

	label := page.children[1].(*StdLabel)
	if len(label.textPieces) != 2 {
		t.Fatalf("label pieces = %d, want 2", len(label.textPieces))
	}
	if _, ok := label.textPieces[1].content.(*StdPageNo); !ok {
		t.Fatalf("second label piece = %T, want *StdPageNo", label.textPieces[1].content)
	}

	nested := page.children[2].(*StdParagraph)
	if len(nested.textPieces) != 2 {
		t.Fatalf("nested paragraph pieces = %d, want 2", len(nested.textPieces))
	}
	if _, ok := nested.textPieces[1].content.(*StdPageNo); !ok {
		t.Fatalf("nested pageno piece = %T, want *StdPageNo", nested.textPieces[1].content)
	}
}

func TestStdPageNo_HiddenContributesNoVisibleText(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml>
  <page>
    <p>Hello <pageno hidden="true" /> world</p>
  </page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}
	doc.ltmls[0].documentPageNo = 7
	w := &labelTestWriter{t: t, fonts: defaultTestFonts(t), lineSpacing: 1.0}
	p := doc.ltmls[0].Page(0).children[0].(*StdParagraph)

	if got := p.RichText(w).String(); got != "Hello world" {
		t.Fatalf("paragraph text = %q, want %q", got, "Hello world")
	}
}

func TestStdPageNo_FontOverrideOnlyAppliesToPageNumberText(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml>
  <page>
    <p>Hello <pageno font.weight="Bold" /> world</p>
  </page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}
	doc.ltmls[0].documentPageNo = 3
	w := ltpdf.NewDocWriter()
	p := doc.ltmls[0].Page(0).children[0].(*StdParagraph)
	rt := p.RichText(w)
	var gotText string
	var regularFont string
	var pageNoFont string
	rt.VisitAll(func(piece *rich_text.RichText) {
		gotText += piece.Text
		if piece.Font == nil {
			return
		}
		switch piece.Text {
		case "3":
			pageNoFont = piece.Font.PostScriptName()
		default:
			if regularFont == "" {
				regularFont = piece.Font.PostScriptName()
			}
		}
	})
	if gotText != "Hello 3 world" {
		t.Fatalf("text = %q, want %q", gotText, "Hello 3 world")
	}
	if regularFont == "" || pageNoFont == "" {
		t.Fatalf("fonts = regular:%q pageno:%q", regularFont, pageNoFont)
	}
	if regularFont == pageNoFont {
		t.Fatalf("page number font = %q, want different from surrounding %q", pageNoFont, regularFont)
	}
}

func TestStdPageNo_PageNumberSequence(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml>
  <page><p>Page <pageno /></p></page>
  <page><p>Page <pageno /></p></page>
  <page><p>Reset next<pageno hidden="true" reset="1" /></p></page>
  <page><p>Page <pageno /></p></page>
  <page><p>Page <pageno start="10" /></p></page>
  <page><p>Page <pageno /></p></page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}

	w := &labelTestWriter{t: t, fonts: defaultTestFonts(t), lineSpacing: 1.0}
	want := []string{"Page 1", "Page 2", "Reset next", "Page 1", "Page 10", "Page 11"}
	for i, expected := range want {
		page := doc.ltmls[0].Page(i)
		if err := page.BeforePrint(w); err != nil {
			t.Fatal(err)
		}
		p := page.children[0].(*StdParagraph)
		if got := p.RichText(w).String(); got != expected {
			t.Fatalf("page %d text = %q, want %q", i+1, got, expected)
		}
	}
}
