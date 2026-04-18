package ltml

import "testing"

func TestFontRem_ParagraphUsesDocumentFontWhenPageHasNoFont(t *testing.T) {
	doc := parseDoc(t, `
		<ltml font.name="Helvetica" font.size="18">
			<page><p font.size="1rem">hello</p></page>
		</ltml>`)

	p := firstParagraph(t, doc)
	w := &labelTestWriter{t: t, lineSpacing: 1.0}
	assertAllLeafFontSizesEqual(t, p.RichText(w), 18)
}

func TestFontRem_ParagraphUsesPageFontWhenPresent(t *testing.T) {
	doc := parseDoc(t, `
		<ltml font.name="Helvetica" font.size="18">
			<page font.name="Helvetica" font.size="15"><p font.size="1rem">hello</p></page>
		</ltml>`)

	p := firstParagraph(t, doc)
	w := &labelTestWriter{t: t, lineSpacing: 1.0}
	assertAllLeafFontSizesEqual(t, p.RichText(w), 15)
}

func TestFontRem_NamedFontStyleResolvesAgainstPageRoot(t *testing.T) {
	doc := parseDoc(t, `
		<ltml font.name="Helvetica" font.size="16">
			<font id="body" name="Helvetica" size="1.25rem" />
			<page font.name="Helvetica" font.size="20"><p font="body">hello</p></page>
		</ltml>`)

	p := firstParagraph(t, doc)
	w := &labelTestWriter{t: t, lineSpacing: 1.0}
	assertAllLeafFontSizesEqual(t, p.RichText(w), 25)
}

func TestFontRem_RuleUsesPageRoot(t *testing.T) {
	doc := parseDoc(t, `
		<ltml font.name="Helvetica" font.size="16">
			<style>p { font.size: 1rem; }</style>
			<page font.name="Helvetica" font.size="20"><p>hello</p></page>
		</ltml>`)

	p := firstParagraph(t, doc)
	w := &labelTestWriter{t: t, lineSpacing: 1.0}
	assertAllLeafFontSizesEqual(t, p.RichText(w), 20)
}

func TestFontRem_FallsBackToBuiltInDefaultWithoutDocumentOrPageFont(t *testing.T) {
	doc := parseDoc(t, `
		<ltml>
			<page><p font.size="1rem">hello</p></page>
		</ltml>`)

	p := firstParagraph(t, doc)
	w := &labelTestWriter{t: t, lineSpacing: 1.0}
	assertAllLeafFontSizesEqual(t, p.RichText(w), 12)
}

func TestFontRem_PageFontInheritedWithoutCompounding(t *testing.T) {
	doc := parseDoc(t, `
		<ltml font.name="Helvetica" font.size="10">
			<page font.name="Helvetica" font.size="1.5rem"><p>hello</p></page>
		</ltml>`)

	p := firstParagraph(t, doc)
	w := &labelTestWriter{t: t, lineSpacing: 1.0}
	assertAllLeafFontSizesEqual(t, p.RichText(w), 15)
}

func TestFontRem_PreUsesFixedFontResolution(t *testing.T) {
	doc := parseDoc(t, `
		<ltml font.name="Helvetica" font.size="16">
			<font id="fixed" name="Courier New" size="1rem" />
			<page><pre>hello</pre></page>
		</ltml>`)

	page := firstPage(t, doc)
	pre, ok := page.children[0].(*StdPre)
	if !ok {
		t.Fatalf("first child is %T, want *StdPre", page.children[0])
	}
	w := &labelTestWriter{t: t, lineSpacing: 1.0}
	rt, err := pre.richTextForLine("hello", w)
	if err != nil {
		t.Fatal(err)
	}
	assertAllLeafFontSizesEqual(t, rt, 16)
}

func TestFontRem_BuiltinHeadingAliasUsesPageRoot(t *testing.T) {
	doc := parseDoc(t, `
		<ltml font.name="Helvetica" font.size="16">
			<page font.name="Helvetica" font.size="12"><h2>hello</h2></page>
		</ltml>`)

	page := firstPage(t, doc)
	label, ok := page.children[0].(*StdLabel)
	if !ok {
		t.Fatalf("first child is %T, want *StdLabel", page.children[0])
	}
	if got := label.AccessibilityRole(); got != "H2" {
		t.Fatalf("AccessibilityRole() = %q, want %q", got, "H2")
	}
	w := &labelTestWriter{t: t, lineSpacing: 1.0}
	assertAllLeafFontSizesEqual(t, label.RichText(w), 21)
}
