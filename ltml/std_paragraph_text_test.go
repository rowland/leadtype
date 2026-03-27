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
