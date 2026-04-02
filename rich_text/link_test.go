package rich_text

import "testing"

func TestRichText_SplitPreservesLinkMetadata(t *testing.T) {
	piece := &RichText{Text: "Hello world", LinkTarget: "intro"}

	left, right := piece.Split(5)
	if left.LinkTarget != "intro" {
		t.Fatalf("left link target = %q, want intro", left.LinkTarget)
	}
	if right.LinkTarget != "intro" {
		t.Fatalf("right link target = %q, want intro", right.LinkTarget)
	}
}

func TestRichText_TrimSpacePreservesLinkMetadata(t *testing.T) {
	piece := &RichText{Text: "  Hello  ", LinkURI: "https://example.com"}

	trimmed := piece.TrimSpace()
	if trimmed.Text != "Hello" {
		t.Fatalf("trimmed text = %q, want Hello", trimmed.Text)
	}
	if trimmed.LinkURI != "https://example.com" {
		t.Fatalf("trimmed link URI = %q, want https://example.com", trimmed.LinkURI)
	}
}

func TestRichText_ScalePreservesLinkMetadata(t *testing.T) {
	piece := &RichText{Text: "Hello", FontSize: 12, LinkURI: "https://example.com"}

	scaled := piece.Scale(0.5, 6)
	if scaled.LinkURI != "https://example.com" {
		t.Fatalf("scaled link URI = %q, want https://example.com", scaled.LinkURI)
	}
}

func TestRichText_MergeDoesNotCombineDifferentLinks(t *testing.T) {
	root := &RichText{pieces: []*RichText{
		{Text: "Hello", LinkTarget: "one"},
		{Text: " world", LinkTarget: "two"},
	}}

	merged := root.Merge()
	if len(merged.pieces) != 2 {
		t.Fatalf("merged piece count = %d, want 2", len(merged.pieces))
	}
	if merged.pieces[0].LinkTarget != "one" || merged.pieces[1].LinkTarget != "two" {
		t.Fatalf("merged links = %q, %q", merged.pieces[0].LinkTarget, merged.pieces[1].LinkTarget)
	}
}
