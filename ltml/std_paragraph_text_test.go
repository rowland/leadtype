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
