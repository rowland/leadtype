package ltml

import "testing"

func TestStdParagraph_AddTextWithFont_NormalizesXMLWhitespace(t *testing.T) {
	p := &StdParagraph{}
	font := &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 12}

	p.AddTextWithFont("\n        Four score and seven years ago\n        our fathers brought forth.\n", font)

	if len(p.textPieces) != 1 {
		t.Fatalf("expected 1 text piece, got %d", len(p.textPieces))
	}
	got := p.textPieces[0].ResolvedText(nil)
	want := "Four score and seven years ago our fathers brought forth. "
	if got != want {
		t.Fatalf("normalized text = %q, want %q", got, want)
	}
}

func TestStdParagraph_AddTextWithFont_PreservesSpanBoundarySpaces(t *testing.T) {
	p := &StdParagraph{}
	body := &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 12}
	emph := &FontStyle{id: "emph", entries: []fontEntry{{name: "Helvetica"}}, size: 12, weight: "Bold"}

	p.AddTextWithFont("Hello ", body)
	p.AddTextWithFont("big", emph)
	p.AddTextWithFont(" world", body)

	if len(p.textPieces) != 3 {
		t.Fatalf("expected 3 text pieces, got %d", len(p.textPieces))
	}
	if got := p.textPieces[0].ResolvedText(nil); got != "Hello " {
		t.Fatalf("first piece = %q, want %q", got, "Hello ")
	}
	if got := p.textPieces[1].ResolvedText(nil); got != "big" {
		t.Fatalf("second piece = %q, want %q", got, "big")
	}
	if got := p.textPieces[2].ResolvedText(nil); got != " world" {
		t.Fatalf("third piece = %q, want %q", got, " world")
	}
}

func TestStdParagraph_RichText_ReappliesFontsWhenUsingCachedRichText(t *testing.T) {
	p := &StdParagraph{}
	p.font = &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 12}
	p.AddText("Hello")

	probe := &mockWriter{t: t}
	if got := p.RichText(probe); got == nil {
		t.Fatal("expected cached rich text to be built")
	}

	render := &mockWriter{t: t}
	if got := p.RichText(render); got == nil {
		t.Fatal("expected cached rich text to be returned")
	}
	if len(render.setFontCalls) == 0 {
		t.Fatal("expected cached rich text path to apply fonts to the render writer")
	}
}

func TestStdParagraph_DrawContent_TextFillClipsParagraphLinesAcrossWidgetRect(t *testing.T) {
	p := &StdParagraph{}
	p.font = &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 12}
	p.paragraphStyle = &ParagraphStyle{}
	p.SetLeft(10)
	p.SetTop(20)
	p.SetWidth(90)
	p.SetHeight(80)
	p.SetAttrs(map[string]string{"text-fill": "Gold"})
	p.AddText("Lorem ipsum dolor sit amet, consectetur adipiscing elit.")

	w := &labelTestWriter{t: t, fonts: defaultTestFonts(t), lineSpacing: 1.0}
	lines := p.Lines(w, p.lineWidth())
	if len(lines) < 2 {
		t.Fatalf("wrapped line count = %d, want at least 2", len(lines))
	}
	if err := p.DrawContent(w); err != nil {
		t.Fatal(err)
	}
	if len(w.printed) != 0 {
		t.Fatalf("PrintRichText count = %d, want 0", len(w.printed))
	}
	if len(w.clipped) != len(lines) {
		t.Fatalf("ClipRichText count = %d, want %d", len(w.clipped), len(lines))
	}
	if len(w.fillRectPages) != len(lines) {
		t.Fatalf("fill rect count = %d, want %d", len(w.fillRectPages), len(lines))
	}
	if len(w.moves) < len(lines) {
		t.Fatalf("move count = %d, want at least %d", len(w.moves), len(lines))
	}
	lineMoves := w.moves[len(w.moves)-len(lines):]
	for i := 1; i < len(lines); i++ {
		if lineMoves[i][1] <= lineMoves[i-1][1] {
			t.Fatalf("line %d baseline y = %v, want greater than previous line y %v", i, lineMoves[i][1], lineMoves[i-1][1])
		}
	}
}

func TestStdParagraph_DrawContent_TextFillUsesParagraphAlignmentWithPlainBullet(t *testing.T) {
	p := &StdParagraph{}
	p.font = &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 12}
	p.paragraphStyle = &ParagraphStyle{TextStyle: TextStyle{textAlign: HAlignCenter, textAlignSet: true}}
	p.bullet = &BulletStyle{text: "*", width: 18, font: p.font}
	p.SetLeft(10)
	p.SetTop(20)
	p.SetWidth(120)
	p.SetHeight(40)
	p.SetAttrs(map[string]string{
		"text-fill.kind":  "linear-gradient",
		"text-fill.x0":    "0",
		"text-fill.y0":    "0",
		"text-fill.x1":    "120",
		"text-fill.y1":    "0",
		"text-fill.stops": "0:Blue,1:Gold",
	})
	p.AddText("Hello")

	w := &labelTestWriter{t: t, fonts: defaultTestFonts(t), lineSpacing: 1.0}
	lines := p.Lines(w, p.lineWidth())
	if err := p.DrawContent(w); err != nil {
		t.Fatal(err)
	}
	if len(w.plainPrinted) != 1 || w.plainPrinted[0] != "*" {
		t.Fatalf("plain bullet output = %#v, want [*]", w.plainPrinted)
	}
	if len(w.clippedText) != 0 {
		t.Fatalf("bullet should not be clipped, got %#v", w.clippedText)
	}
	if len(w.clipped) != len(lines) {
		t.Fatalf("ClipRichText count = %d, want %d", len(w.clipped), len(lines))
	}
	if len(w.linearPaints) == 0 {
		t.Fatal("expected gradient paints")
	}
	opts := paragraphTextFillOptions(p)
	if got := opts.StringDefault("text-align", ""); got != "center" {
		t.Fatalf("text-align = %q, want center", got)
	}
	if got := opts.FloatDefault("width", 0); got != ContentWidth(p)-p.textIndent() {
		t.Fatalf("width = %v, want %v", got, ContentWidth(p)-p.textIndent())
	}
}

func TestStdParagraph_SplitForHeight_RespectsDefaultsAndSuppressesBullet(t *testing.T) {
	page := &StdPage{pageStyle: &PageStyle{width: 200, height: 200}}
	page.layout = defaultLayouts["vbox"].Clone()

	p := &StdParagraph{}
	_ = p.SetContainer(page)
	p.font = &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 12}
	p.bullet = &BulletStyle{text: "*", width: 18, font: p.font}
	p.splitEnabled = true
	p.orphans = 2
	p.widows = 2
	p.SetWidth(90)
	p.AddText("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur.")

	w := &labelTestWriter{t: t, fonts: defaultTestFonts(t), lineSpacing: 1.0}
	lines := p.Lines(w, p.lineWidth())
	if len(lines) < 5 {
		t.Fatalf("wrapped line count = %d, want at least 5", len(lines))
	}

	avail := p.heightForLines(lines[:2], w)
	result, err := p.SplitForHeight(avail, w)
	if err != nil {
		t.Fatal(err)
	}
	if result == nil {
		t.Fatal("expected paragraph to split")
	}

	head := result.Head.(*StdParagraph)
	tail := result.Tail.(*StdParagraph)
	if len(head.splitLines) != 2 {
		t.Fatalf("head line count = %d, want 2", len(head.splitLines))
	}
	if len(tail.splitLines) != len(lines)-2 {
		t.Fatalf("tail line count = %d, want %d", len(tail.splitLines), len(lines)-2)
	}
	if head.suppressBullet {
		t.Fatal("head should keep bullet")
	}
	if !tail.suppressBullet {
		t.Fatal("tail should suppress bullet")
	}
	if tail.continuationIndent != p.bulletWidth() {
		t.Fatalf("tail continuation indent = %v, want %v", tail.continuationIndent, p.bulletWidth())
	}
}
