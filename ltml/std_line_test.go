package ltml

import "testing"

type lineTestWriter struct {
	labelTestWriter
	lines []lineCall
	caps  []string
}

type lineCall struct {
	x      float64
	y      float64
	angle  float64
	length float64
}

func (w *lineTestWriter) Line(x, y, angle, length float64) {
	w.lines = append(w.lines, lineCall{x: x, y: y, angle: angle, length: length})
}

func (w *lineTestWriter) SetLineCapStyle(style string) (prev string) {
	w.caps = append(w.caps, style)
	return ""
}

func TestStdLine_PreferredDimensions_FromExplicitLength(t *testing.T) {
	line := &StdLine{length: 100, angle: 0}

	if got := line.PreferredWidth(nil); got != 100 {
		t.Fatalf("PreferredWidth() = %v, want 100", got)
	}
	if got := line.PreferredHeight(nil); got != 0 {
		t.Fatalf("PreferredHeight() = %v, want 0", got)
	}
}

func TestStdLine_CalcLength_FitsContentBox(t *testing.T) {
	line := &StdLine{angle: 45}
	line.SetWidth(100)
	line.SetHeight(50)

	got := line.Length()
	want := 70.71067811865476
	if got < want-0.0001 || got > want+0.0001 {
		t.Fatalf("Length() = %v, want %v", got, want)
	}
}

func TestStdLine_CalcLength_UsesWidthOnlyForHorizontalLine(t *testing.T) {
	line := &StdLine{angle: 0}
	line.SetWidth(100)

	if got := line.Length(); got != 100 {
		t.Fatalf("Length() = %v, want 100", got)
	}
}

func TestStdLine_CalcLength_UsesHeightOnlyForVerticalLine(t *testing.T) {
	line := &StdLine{angle: 90}
	line.SetHeight(72)

	if got := line.Length(); got < 71.9999 || got > 72.0001 {
		t.Fatalf("Length() = %v, want 72", got)
	}
}

func TestStdLine_DrawContent_UsesStyleAndCenteredOrigin(t *testing.T) {
	line := &StdLine{angle: 0}
	line.SetLeft(10)
	line.SetTop(20)
	line.SetWidth(100)
	line.SetHeight(20)
	line.SetScope(&defaultScope)
	line.style = &PenStyle{id: "test", color: NamedColor("red"), width: 2, pattern: "dashed"}
	w := &lineTestWriter{}

	if err := line.DrawContent(w); err != nil {
		t.Fatal(err)
	}
	if len(w.lines) != 1 {
		t.Fatalf("Line call count = %d, want 1", len(w.lines))
	}
	call := w.lines[0]
	if call.x != 10 || call.y != 30 {
		t.Fatalf("origin = (%v,%v), want (10,30)", call.x, call.y)
	}
	if call.angle != 0 {
		t.Fatalf("angle = %v, want 0", call.angle)
	}
	if call.length != 100 {
		t.Fatalf("length = %v, want 100", call.length)
	}
}

func TestStdLine_DrawContent_WidthOnlyLineUsesLaidOutTop(t *testing.T) {
	line := &StdLine{angle: 0}
	line.SetLeft(10)
	line.SetTop(50)
	line.SetWidth(100)
	line.SetScope(&defaultScope)
	w := &lineTestWriter{}

	if err := line.DrawContent(w); err != nil {
		t.Fatal(err)
	}
	if len(w.lines) != 1 {
		t.Fatalf("Line call count = %d, want 1", len(w.lines))
	}
	call := w.lines[0]
	if call.y != 50 {
		t.Fatalf("origin y = %v, want 50", call.y)
	}
}

func TestStdLine_Style_DoesNotReuseWidgetBorder(t *testing.T) {
	line := &StdLine{}
	line.SetScope(&defaultScope)
	line.border = &PenStyle{id: "border_only", color: NamedColor("red"), width: 2, pattern: "dashed"}

	style := line.Style()
	if style == nil {
		t.Fatal("Style() returned nil")
	}
	if style.id != "solid" {
		t.Fatalf("Style().id = %q, want %q", style.id, "solid")
	}
	if style.Cap() != "butt_cap" {
		t.Fatalf("Style().Cap() = %q, want %q", style.Cap(), "butt_cap")
	}
}

func TestPenStyle_Apply_UsesCap(t *testing.T) {
	style := &PenStyle{id: "rounded", color: NamedColor("black"), width: 5, pattern: "dashed", cap: "round_cap"}
	w := &lineTestWriter{}

	style.Apply(w)

	if len(w.caps) != 1 || w.caps[0] != "round_cap" {
		t.Fatalf("caps = %v, want [round_cap]", w.caps)
	}
}

func TestParse_LineTag(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml>
  <page>
    <line angle="90" height="1in" style="dotted" />
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
	line, ok := page.children[0].(*StdLine)
	if !ok {
		t.Fatalf("child type = %T, want *StdLine", page.children[0])
	}
	if line.angle != 90 {
		t.Fatalf("angle = %v, want 90", line.angle)
	}
	if line.style == nil || line.style.id != "dotted" {
		t.Fatalf("style = %#v, want dotted pen", line.style)
	}
}
