package ltml

import (
	"testing"

	"github.com/rowland/leadtype/afm_fonts"
	"github.com/rowland/leadtype/font"
	"github.com/rowland/leadtype/options"
)

func newPreTestWriter(t *testing.T) *labelTestWriter {
	t.Helper()
	fontSource, err := afm_fonts.Default()
	if err != nil {
		t.Fatal(err)
	}
	face, err := font.New("Courier", options.Options{"size": 12.0}, font.FontSources{fontSource})
	if err != nil {
		t.Fatal(err)
	}
	return &labelTestWriter{
		fonts:       []*font.Font{face},
		fontSize:    12,
		lineSpacing: 1.0,
	}
}

func TestStdPre_Lines_PreservesFormattingAndDedents(t *testing.T) {
	p := &StdPre{}
	p.AddText("\n    first line\n      second line\n\n    third line\n")

	got := p.Lines()
	want := []string{"first line", "  second line", "", "third line"}
	if len(got) != len(want) {
		t.Fatalf("line count = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("line %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestStdPre_Lines_ExpandsTabs(t *testing.T) {
	p := &StdPre{}
	p.AddText("one\ttwo")

	got := p.Lines()
	if len(got) != 1 {
		t.Fatalf("line count = %d, want 1", len(got))
	}
	if got[0] != "one    two" {
		t.Fatalf("line 0 = %q, want %q", got[0], "one    two")
	}
}

func TestStdPre_DrawContent_PrintsEachLineWithoutWrapping(t *testing.T) {
	p := &StdPre{}
	p.SetLeft(10)
	p.SetTop(20)
	p.AddText("\n    alpha\n    beta\n")

	w := newPreTestWriter(t)

	if err := p.DrawContent(w); err != nil {
		t.Fatal(err)
	}
	if len(w.printed) != 2 {
		t.Fatalf("PrintRichText count = %d, want 2", len(w.printed))
	}
	if got := w.printed[0].String(); got != "alpha" {
		t.Fatalf("line 0 = %q, want %q", got, "alpha")
	}
	if got := w.printed[1].String(); got != "beta" {
		t.Fatalf("line 1 = %q, want %q", got, "beta")
	}
	if len(w.moves) != 2 {
		t.Fatalf("MoveTo count = %d, want 2", len(w.moves))
	}
	if w.moves[1][1] <= w.moves[0][1] {
		t.Fatalf("second baseline y = %v, want > first y %v", w.moves[1][1], w.moves[0][1])
	}
}

func TestStdPre_DefaultsToFixedFont(t *testing.T) {
	p := &StdPre{}
	p.SetScope(&defaultScope)

	if got := p.Font().ID(); got != "fixed" {
		t.Fatalf("Font().ID() = %q, want %q", got, "fixed")
	}
}

func TestParse_PreTag(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml>
  <page>
    <pre>
      if x &lt; 1 {
        return
      }
    </pre>
  </page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}

	page := doc.ltmls[0].Page(0)
	if page == nil {
		t.Fatal("page is nil")
	}
	if len(page.children) != 1 {
		t.Fatalf("child count = %d, want 1", len(page.children))
	}
	pre, ok := page.children[0].(*StdPre)
	if !ok {
		t.Fatalf("child type = %T, want *StdPre", page.children[0])
	}
	want := []string{"if x < 1 {", "  return", "}"}
	got := pre.Lines()
	if len(got) != len(want) {
		t.Fatalf("line count = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("line %d = %q, want %q", i, got[i], want[i])
		}
	}
}
