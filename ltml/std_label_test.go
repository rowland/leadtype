package ltml

import (
	"testing"

	"github.com/rowland/leadtype/afm_fonts"
	"github.com/rowland/leadtype/colors"
	"github.com/rowland/leadtype/font"
	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/rich_text"
)

type labelTestWriter struct {
	fonts       []*font.Font
	fontSize    float64
	lineSpacing float64
	moves       [][2]float64
	printed     []*rich_text.RichText
}

func (w *labelTestWriter) FontColor() colors.Color { return 0 }
func (w *labelTestWriter) Fonts() []*font.Font     { return w.fonts }
func (w *labelTestWriter) FontSize() float64       { return w.fontSize }
func (w *labelTestWriter) ImageDimensionsFromFile(filename string) (width, height int, err error) {
	return 0, 0, nil
}
func (w *labelTestWriter) LineSpacing() float64 {
	if w.lineSpacing == 0 {
		return 1.0
	}
	return w.lineSpacing
}
func (w *labelTestWriter) SetLineCapStyle(style string) (prev string) { return "" }
func (w *labelTestWriter) Line(x, y, angle, length float64) {}
func (w *labelTestWriter) LineTo(x, y float64) {}
func (w *labelTestWriter) Loc() (x, y float64) { return 0, 0 }
func (w *labelTestWriter) MoveTo(x, y float64) { w.moves = append(w.moves, [2]float64{x, y}) }
func (w *labelTestWriter) NewPage()            {}
func (w *labelTestWriter) Print(text string) error {
	return nil
}
func (w *labelTestWriter) PrintImageFile(filename string, x, y float64, width, height *float64) (actualWidth, actualHeight float64, err error) {
	return 0, 0, nil
}
func (w *labelTestWriter) PrintParagraph(para []*rich_text.RichText, opts options.Options) {}
func (w *labelTestWriter) PrintRichText(text *rich_text.RichText) {
	w.printed = append(w.printed, text)
}
func (w *labelTestWriter) Rectangle(x, y, width, height float64, border bool, fill bool) {}
func (w *labelTestWriter) Rectangle2(x, y, width, height float64, border bool, fill bool, corners []float64, path, reverse bool) {
}
func (w *labelTestWriter) AddFont(family string, opts options.Options) ([]*font.Font, error) {
	return w.fonts, nil
}
func (w *labelTestWriter) SetFont(name string, size float64, opts options.Options) ([]*font.Font, error) {
	w.fontSize = size
	return w.fonts, nil
}
func (w *labelTestWriter) SetFillColor(value interface{}) (prev colors.Color)  { return 0 }
func (w *labelTestWriter) SetLineColor(value colors.Color) (prev colors.Color) { return 0 }
func (w *labelTestWriter) SetLineDashPattern(pattern string) (prev string)     { return "" }
func (w *labelTestWriter) SetLineSpacing(lineSpacing float64) (prev float64)   { return w.lineSpacing }
func (w *labelTestWriter) SetLineWidth(width float64)                           {}
func (w *labelTestWriter) SetStrikeout(strikeout bool) (prev bool)             { return false }
func (w *labelTestWriter) SetUnderline(underline bool) (prev bool)             { return false }
func (w *labelTestWriter) Strikeout() bool                                      { return false }
func (w *labelTestWriter) Underline() bool                                      { return false }

func TestStdLabel_AddTextWithFont_NormalizesXMLWhitespace(t *testing.T) {
	l := &StdLabel{}
	font := &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 12}

	l.AddTextWithFont("\n        Four score and seven years ago\n        our fathers brought forth.\n", font)

	if len(l.textPieces) != 1 {
		t.Fatalf("expected 1 text piece, got %d", len(l.textPieces))
	}
	got := l.textPieces[0].text
	want := "Four score and seven years ago our fathers brought forth."
	if got != want {
		t.Fatalf("normalized text = %q, want %q", got, want)
	}
}

func TestStdLabel_PreferredHeight_EmptyLabelUsesFontLineHeight(t *testing.T) {
	l := &StdLabel{}
	l.font = &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 12}
	w := &labelTestWriter{lineSpacing: 1.25}

	got := l.PreferredHeight(w)
	want := 15.0
	if got != want {
		t.Fatalf("PreferredHeight() = %v, want %v", got, want)
	}
}

func TestStdLabel_DrawContent_PrintsRichText(t *testing.T) {
	l := &StdLabel{}
	l.font = &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 12}
	l.SetLeft(10)
	l.SetTop(20)
	l.AddText("Hello")

	fontSource, err := afm_fonts.Default()
	if err != nil {
		t.Fatal(err)
	}
	face, err := font.New("Helvetica", options.Options{"size": 12.0}, font.FontSources{fontSource})
	if err != nil {
		t.Fatal(err)
	}
	w := &labelTestWriter{fonts: []*font.Font{face}, lineSpacing: 1.0}

	if err := l.DrawContent(w); err != nil {
		t.Fatal(err)
	}
	if len(w.moves) != 1 {
		t.Fatalf("MoveTo count = %d, want 1", len(w.moves))
	}
	if len(w.printed) != 1 {
		t.Fatalf("PrintRichText count = %d, want 1", len(w.printed))
	}
	if got := w.printed[0].String(); got != "Hello" {
		t.Fatalf("printed text = %q, want %q", got, "Hello")
	}
}

func TestParse_LabelAndBrAlias(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml>
  <page>
    <label>Hello <span font.weight="Bold">world</span></label>
    <br/>
  </page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}

	page := doc.ltmls[0].Page(0)
	if page == nil {
		t.Fatal("page is nil")
	}
	if len(page.children) != 2 {
		t.Fatalf("child count = %d, want 2", len(page.children))
	}
	if _, ok := page.children[0].(*StdLabel); !ok {
		t.Fatalf("first child type = %T, want *StdLabel", page.children[0])
	}
	if _, ok := page.children[1].(*StdLabel); !ok {
		t.Fatalf("second child type = %T, want *StdLabel", page.children[1])
	}

	label := page.children[0].(*StdLabel)
	if len(label.textPieces) != 2 {
		t.Fatalf("text piece count = %d, want 2", len(label.textPieces))
	}
	if got := label.textPieces[0].text; got != "Hello " {
		t.Fatalf("piece 0 = %q, want %q", got, "Hello ")
	}
	if got := label.textPieces[1].text; got != "world" {
		t.Fatalf("piece 1 = %q, want %q", got, "world")
	}
}
