// Copyright 2017 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import "testing"

func TestStdContainer_Path(t *testing.T) {
	i1 := map[string]string{"tag": "foo", "id": "bar", "class": "boom baz"}
	e1 := "foo#bar.baz.boom"

	var c1 StdContainer
	c1.SetIentifiers(i1)
	if c1.Path() != e1 {
		t.Errorf("Expected <%s>, got <%s>", e1, c1.Path())
	}

	i2 := map[string]string{"tag": "abc", "id": "def", "class": "jkl ghi"}
	e2 := "foo#bar.baz.boom/abc#def.ghi.jkl"

	var c2 StdContainer
	c2.SetIentifiers(i2)
	c2.container = &c1
	if c2.Path() != e2 {
		t.Errorf("Expected <%s>, got <%s>", e2, c2.Path())
	}
}

func TestStdContainer_SetAttrs_ClonesLayoutForLayoutPrefixOverrides(t *testing.T) {
	scope := &Scope{}
	scope.SetParentScope(&defaultScope)

	container := &StdContainer{}
	container.SetScope(scope)
	container.SetAttrs(map[string]string{
		"layout":          "vbox",
		"layout.vpadding": "9pt",
	})

	if container.LayoutStyle() == defaultLayouts["vbox"] {
		t.Fatal("layout style reused shared vbox layout, want clone")
	}
	if got := container.LayoutStyle().VPadding(); got != 9 {
		t.Fatalf("layout vpadding = %v, want 9", got)
	}
	if got := defaultLayouts["vbox"].VPadding(); got != 0 {
		t.Fatalf("shared vbox vpadding = %v, want 0", got)
	}
}

func TestStdContainer_SetAttrs_ParagraphStyleOverridesWithoutParent(t *testing.T) {
	container := &StdContainer{}
	container.SetAttrs(map[string]string{
		"paragraph-style.text-align": "center",
	})

	if container.paragraphStyle == nil {
		t.Fatal("paragraphStyle is nil, want cloned default style")
	}
	if got := container.ParagraphStyle().ResolvedTextAlign(container); got != HAlignCenter {
		t.Fatalf("paragraph text-align = %s, want center", got)
	}
}

func TestStdContainer_SetAttrs_PreservesRadialSweepAcrossAttributeLayers(t *testing.T) {
	container := &StdContainer{}
	container.SetAttrs(map[string]string{"sweep": "cw"})
	container.SetAttrs(map[string]string{"rows": "2"})

	if got := container.RadialSweep(); got != radialSweepCW {
		t.Fatalf("radial sweep = %v, want clockwise value from previous attribute layer", got)
	}

	container.SetAttrs(map[string]string{"sweep": "sideways"})
	if got := container.RadialSweep(); got != radialSweepCW {
		t.Fatalf("radial sweep = %v after invalid value, want previous clockwise value", got)
	}

	container.SetAttrs(map[string]string{"sweep": "ccw"})
	if got := container.RadialSweep(); got != radialSweepCCW {
		t.Fatalf("radial sweep = %v, want explicit counterclockwise override", got)
	}
}

func TestStdParagraph_SetAttrs_StyleOverridesWithoutParent(t *testing.T) {
	paragraph := &StdParagraph{}
	paragraph.SetAttrs(map[string]string{
		"style.text-align": "right",
	})

	if paragraph.paragraphStyle == nil {
		t.Fatal("paragraphStyle is nil, want cloned default style")
	}
	if got := paragraph.ParagraphStyle().ResolvedTextAlign(paragraph); got != HAlignRight {
		t.Fatalf("paragraph text-align = %s, want right", got)
	}
}

func TestStdContainer_PrepareListBullets_ULAssignsDirectChildParagraphsOnly(t *testing.T) {
	doc := parseDoc(t, `
		<ltml>
			<bullet id="custom" text="*" width="12pt" />
			<page>
				<ul>
					<p>First</p>
					<label>Not a list item</label>
					<p bullet="custom">Second</p>
					<p>Third</p>
				</ul>
			</page>
		</ltml>`)

	w := &labelTestWriter{t: t, fonts: defaultTestFonts(t)}
	if err := doc.Print(w); err != nil {
		t.Fatal(err)
	}

	list := firstContainer(t, doc)
	first := childParagraph(t, list, 0)
	second := childParagraph(t, list, 2)
	third := childParagraph(t, list, 3)

	if first.Bullet() == nil || first.Bullet().Shape() != "circle" {
		t.Fatalf("first bullet = %#v, want default unordered circle bullet", first.Bullet())
	}
	if first.Bullet().AlignY() != "middle" {
		t.Fatalf("first bullet align-y = %q, want middle", first.Bullet().AlignY())
	}
	if second.Bullet() == nil || second.Bullet().Text() != "*" {
		t.Fatalf("second bullet = %#v, want preserved explicit bullet", second.Bullet())
	}
	if third.Bullet() == nil || third.Bullet().Shape() != "circle" {
		t.Fatalf("third bullet = %#v, want default unordered circle bullet", third.Bullet())
	}
	if third.Bullet().AlignY() != "middle" {
		t.Fatalf("third bullet align-y = %q, want middle", third.Bullet().AlignY())
	}
}

func TestStdContainer_PrepareListBullets_OLNumbersParagraphsAndAutoSizesMarkers(t *testing.T) {
	doc := parseDoc(t, `
		<ltml>
			<page>
				<ol>
					<p>One</p>
					<p>Two</p>
					<p>Three</p>
					<p>Four</p>
					<p>Five</p>
					<p>Six</p>
					<p>Seven</p>
					<p>Eight</p>
					<p>Nine</p>
					<p>Ten</p>
					<p>Eleven</p>
					<p>Twelve</p>
				</ol>
			</page>
		</ltml>`)

	w := &labelTestWriter{t: t, fonts: defaultTestFonts(t)}
	if err := doc.Print(w); err != nil {
		t.Fatal(err)
	}

	list := firstContainer(t, doc)
	first := childParagraph(t, list, 0)
	tenth := childParagraph(t, list, 9)
	twelfth := childParagraph(t, list, 11)

	if first.Bullet() == nil || first.Bullet().Text() != "1." {
		t.Fatalf("first bullet = %#v, want 1.", first.Bullet())
	}
	if tenth.Bullet() == nil || tenth.Bullet().Text() != "10." {
		t.Fatalf("tenth bullet = %#v, want 10.", tenth.Bullet())
	}
	if twelfth.Bullet() == nil || twelfth.Bullet().Text() != "12." {
		t.Fatalf("twelfth bullet = %#v, want 12.", twelfth.Bullet())
	}
	if first.Bullet().Width() != twelfth.Bullet().Width() || tenth.Bullet().Width() != twelfth.Bullet().Width() {
		t.Fatalf("ordered bullet widths = %v/%v/%v, want equal auto-sized widths", first.Bullet().Width(), tenth.Bullet().Width(), twelfth.Bullet().Width())
	}
	if first.Bullet().Width() <= first.bulletTextWidth(w, first.Bullet()) {
		t.Fatalf("ordered marker slot width = %v, want greater than rendered marker width %v", first.Bullet().Width(), first.bulletTextWidth(w, first.Bullet()))
	}
}

func TestStdContainer_PrepareListBullets_OLPreservesExplicitBulletAndCountsItsOrdinal(t *testing.T) {
	doc := parseDoc(t, `
		<ltml>
			<bullet id="custom" text="*" width="12pt" />
			<page>
				<ol>
					<p>First</p>
					<p bullet="custom">Second</p>
					<p>Third</p>
				</ol>
			</page>
		</ltml>`)

	w := &labelTestWriter{t: t, fonts: defaultTestFonts(t)}
	if err := doc.Print(w); err != nil {
		t.Fatal(err)
	}

	list := firstContainer(t, doc)
	first := childParagraph(t, list, 0)
	second := childParagraph(t, list, 1)
	third := childParagraph(t, list, 2)

	if first.Bullet() == nil || first.Bullet().Text() != "1." {
		t.Fatalf("first bullet = %#v, want 1.", first.Bullet())
	}
	if second.Bullet() == nil || second.Bullet().Text() != "*" {
		t.Fatalf("second bullet = %#v, want preserved explicit bullet", second.Bullet())
	}
	if third.Bullet() == nil || third.Bullet().Text() != "3." {
		t.Fatalf("third bullet = %#v, want 3.", third.Bullet())
	}
}

func TestStdContainer_PrepareListBullets_AssignsMultipleListBulletTemplates(t *testing.T) {
	doc, err := Parse([]byte(`
		<ltml>
			<bullet id="logo" src="logo.svg" width="18pt" height="12pt" />
			<page>
				<ol bullets="logo ordered">
					<p>One</p>
					<p>Two</p>
				</ol>
			</page>
		</ltml>`), WithAssetFS(testingMapFS("logo.svg", "<svg/>")))
	if err != nil {
		t.Fatal(err)
	}

	w := &labelTestWriter{t: t, fonts: defaultTestFonts(t)}
	if err := doc.Print(w); err != nil {
		t.Fatal(err)
	}

	list := firstContainer(t, doc)
	first := childParagraph(t, list, 0)
	second := childParagraph(t, list, 1)

	if got := first.Bullets(); len(got) != 2 || got[0].Source() != "logo.svg" || got[1].Text() != "1." {
		t.Fatalf("first bullets = %#v, want logo and 1.", got)
	}
	if got := second.Bullets(); len(got) != 2 || got[0].Source() != "logo.svg" || got[1].Text() != "2." {
		t.Fatalf("second bullets = %#v, want logo and 2.", got)
	}
}

func TestStdContainer_PrepareListBullets_OLKeepsFormattedTemplateWidth(t *testing.T) {
	doc := parseDoc(t, `
		<ltml>
			<bullet id="num" format="%d." width="40pt" />
			<page>
				<ol bullets="num">
					<p>One</p>
					<p>Two</p>
					<p>Three</p>
					<p>Four</p>
					<p>Five</p>
					<p>Six</p>
					<p>Seven</p>
					<p>Eight</p>
					<p>Nine</p>
					<p>Ten</p>
					<p>Eleven</p>
					<p>Twelve</p>
				</ol>
			</page>
		</ltml>`)

	w := &labelTestWriter{t: t, fonts: defaultTestFonts(t)}
	if err := doc.Print(w); err != nil {
		t.Fatal(err)
	}

	list := firstContainer(t, doc)
	first := childParagraph(t, list, 0)
	twelfth := childParagraph(t, list, 11)
	if first.Bullet() == nil || first.Bullet().Width() != 40 || twelfth.Bullet() == nil || twelfth.Bullet().Width() != 40 {
		t.Fatalf("ordered template widths = %#v / %#v, want 40pt for both", first.Bullet(), twelfth.Bullet())
	}
	if first.Bullet().Text() != "1." || twelfth.Bullet().Text() != "12." {
		t.Fatalf("ordered template texts = %q / %q, want 1. / 12.", first.Bullet().Text(), twelfth.Bullet().Text())
	}
}

func TestStdContainer_PrepareListBullets_FormattedTemplateAutoWidthIgnoresDefaultBulletWidth(t *testing.T) {
	doc := parseDoc(t, `
		<ltml>
			<bullet id="num" format="%d." />
			<page>
				<ol bullets="num">
					<p>One</p>
					<p>Two</p>
				</ol>
			</page>
		</ltml>`)

	w := &labelTestWriter{t: t, fonts: defaultTestFonts(t)}
	if err := doc.Print(w); err != nil {
		t.Fatal(err)
	}

	list := firstContainer(t, doc)
	first := childParagraph(t, list, 0)
	if first.Bullet() == nil {
		t.Fatal("first bullet is nil")
	}
	renderedWidth := first.bulletTextWidth(w, first.Bullet())
	if first.Bullet().Width() >= 36 {
		t.Fatalf("auto marker width = %v, want less than default 36pt slot", first.Bullet().Width())
	}
	if first.Bullet().Width() <= renderedWidth {
		t.Fatalf("auto marker width = %v, want greater than rendered marker width %v", first.Bullet().Width(), renderedWidth)
	}
}

func TestStdContainer_PrepareListBullets_DivCanUseBulletsAttribute(t *testing.T) {
	doc := parseDoc(t, `
		<ltml>
			<page>
				<div layout="vbox" bullets="ordered">
					<p>One</p>
				</div>
				<div layout="vbox" bullets="unordered">
					<p>One</p>
				</div>
			</page>
		</ltml>`)

	w := &labelTestWriter{t: t, fonts: defaultTestFonts(t)}
	if err := doc.Print(w); err != nil {
		t.Fatal(err)
	}

	page := firstPage(t, doc)
	ordered := page.children[0].(*StdContainer)
	unordered := page.children[1].(*StdContainer)

	orderedParagraph := childParagraph(t, ordered, 0)
	if orderedParagraph.Bullet() == nil || orderedParagraph.Bullet().Text() != "1." {
		t.Fatalf("bullets=ordered bullet = %#v, want 1.", orderedParagraph.Bullet())
	}
	unorderedParagraph := childParagraph(t, unordered, 0)
	if unorderedParagraph.Bullet() == nil || unorderedParagraph.Bullet().Shape() != "circle" {
		t.Fatalf("bullets=unordered bullet = %#v, want circle", unorderedParagraph.Bullet())
	}
}
