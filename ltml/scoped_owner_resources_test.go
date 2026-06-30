package ltml

import "testing"

func TestScopedOwnerResources_PageResolvesLocalFontWithCascadeOverrides(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml>
  <style>
    page { font: body; font.size: 14; }
  </style>
  <page font.weight="Bold">
    <font id="body" name="Courier New" size="10" />
    <p>Hello</p>
  </page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}

	page := doc.Page(0)
	if got := fontStyleName(page.Font()); got != "Courier New" {
		t.Fatalf("page font name = %q, want Courier New", got)
	}
	if got := page.Font().ResolveAgainstBase(defaultFontSize); got != 14 {
		t.Fatalf("page font size = %v, want 14", got)
	}
	if got := page.Font().weight; got != "Bold" {
		t.Fatalf("page font weight = %q, want Bold", got)
	}

	named := FontStyleFor("body", page)
	if named == nil {
		t.Fatal("page-local font body was not registered")
	}
	if got := named.ResolveAgainstBase(defaultFontSize); got != 10 {
		t.Fatalf("named font size = %v, want unchanged 10", got)
	}
	if got := named.weight; got != "" {
		t.Fatalf("named font weight = %q, want unchanged empty weight", got)
	}
}

func TestScopedOwnerResources_PreservesUnitsAtEachCascadeLayer(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml>
  <style>
    page { border: edge; border.width: 2; }
  </style>
  <page units="in">
    <pen id="edge" color="red" width="1" />
  </page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}

	page := doc.Page(0)
	if page.border == nil || page.border.ID() != "edge" {
		t.Fatalf("page border = %#v, want edge", page.border)
	}
	if got := page.border.width; got != 2 {
		t.Fatalf("page border width = %v, want 2pt from the rule's unit context", got)
	}
}

func TestScopedOwnerResources_PageScopesMayReuseAndShadowIDs(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml>
  <font id="body" name="Helvetica" size="9" />
  <page font="body">
    <font id="body" name="Courier New" size="11" />
    <p>One</p>
  </page>
  <page font="body">
    <font id="body" name="Times New Roman" size="12" />
    <p>Two</p>
  </page>
  <page font="body">
    <p>Three</p>
  </page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}

	if got := fontStyleName(doc.Page(0).Font()); got != "Courier New" {
		t.Fatalf("page 1 font = %q, want Courier New", got)
	}
	if got := fontStyleName(doc.Page(1).Font()); got != "Times New Roman" {
		t.Fatalf("page 2 font = %q, want Times New Roman", got)
	}
	if got := fontStyleName(doc.Page(2).Font()); got != "Helvetica" {
		t.Fatalf("page 3 font = %q, want inherited Helvetica", got)
	}
	if got := fontStyleName(FontStyleFor("body", doc.Root())); got != "Helvetica" {
		t.Fatalf("document body font = %q, want unchanged Helvetica", got)
	}
}

func TestScopedOwnerResources_DocumentResolvesLocalFontAndPageStyle(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml font="body" style="book">
  <font id="body" name="Courier New" size="13" />
  <pagestyle id="book" width="432" height="648" />
  <page><p>Hello</p></page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}

	if got := fontStyleName(doc.Root().Font()); got != "Courier New" {
		t.Fatalf("document font = %q, want Courier New", got)
	}
	if got := fontStyleName(doc.Page(0).Font()); got != "Courier New" {
		t.Fatalf("inherited page font = %q, want Courier New", got)
	}
	if got := doc.Page(0).Width(); got != 432 {
		t.Fatalf("page width = %v, want 432", got)
	}
	if got := doc.Page(0).Height(); got != 648 {
		t.Fatalf("page height = %v, want 648", got)
	}
}

func TestScopedOwnerResources_PageResolvesAllResourceBackedAttrs(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml>
  <page border="edge" border-top="edge" fill="paper" layout="stack"
        paragraph-style="copy" style="sheet">
    <pen id="edge" color="red" width="2" />
    <brush id="paper" color="yellow" />
    <para id="copy" text-align="right" />
    <layout id="stack" manager="vbox" vpadding="7" />
    <pagestyle id="sheet" width="300" height="400" />
    <p>Hello</p>
  </page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}

	page := doc.Page(0)
	if page.border == nil || page.border.ID() != "edge" {
		t.Fatalf("page border = %#v, want edge", page.border)
	}
	if page.borders[topSide] == nil || page.borders[topSide].ID() != "edge" {
		t.Fatalf("page top border = %#v, want edge", page.borders[topSide])
	}
	if page.fill == nil || page.fill.ID() != "paper" {
		t.Fatalf("page fill = %#v, want paper", page.fill)
	}
	if page.LayoutStyle() == nil || page.LayoutStyle().ID() != "stack" {
		t.Fatalf("page layout = %#v, want stack", page.LayoutStyle())
	}
	if got := page.LayoutStyle().VPadding(); got != 7 {
		t.Fatalf("page layout vpadding = %v, want 7", got)
	}
	if page.ParagraphStyle() == nil || page.ParagraphStyle().ID() != "copy" {
		t.Fatalf("page paragraph style = %#v, want copy", page.ParagraphStyle())
	}
	if got := page.ParagraphStyle().ResolvedTextAlign(page); got != HAlignRight {
		t.Fatalf("page text align = %v, want right", got)
	}
	if got := page.Width(); got != 300 {
		t.Fatalf("page width = %v, want 300", got)
	}
	if got := page.Height(); got != 400 {
		t.Fatalf("page height = %v, want 400", got)
	}
}

func TestScopedOwnerResources_CanvasResolvesItsLocalResources(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml>
  <canvas key="badge" width="120" height="60" font="body" layout="stack">
    <font id="body" name="Courier New" size="9" />
    <layout id="stack" manager="absolute" />
    <label>Hello</label>
  </canvas>
  <page />
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}

	canvas := doc.Root().Canvas("badge")
	if canvas == nil {
		t.Fatal("canvas badge was not registered")
	}
	if got := fontStyleName(canvas.Font()); got != "Courier New" {
		t.Fatalf("canvas font = %q, want Courier New", got)
	}
	if canvas.LayoutStyle() == nil || canvas.LayoutStyle().ID() != "stack" {
		t.Fatalf("canvas layout = %#v, want stack", canvas.LayoutStyle())
	}
}

func TestScopedOwnerResources_OrdinaryChildrenRemainDeclarationOrdered(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml>
  <page font.name="Helvetica">
    <p font="late" font.size="20">Before</p>
    <font id="late" name="Courier New" size="10" />
    <p font="late">After</p>
  </page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}

	children := doc.Page(0).Widgets()
	if len(children) != 2 {
		t.Fatalf("page child count = %d, want 2", len(children))
	}
	before := children[0].(*StdParagraph)
	after := children[1].(*StdParagraph)
	if got := fontStyleName(before.Font()); got != "Helvetica" {
		t.Fatalf("child before declaration font = %q, want inherited Helvetica", got)
	}
	if got := before.Font().ResolveAgainstBase(defaultFontSize); got != 20 {
		t.Fatalf("child before declaration size = %v, want 20", got)
	}
	if got := fontStyleName(after.Font()); got != "Courier New" {
		t.Fatalf("child after declaration font = %q, want Courier New", got)
	}
}

func fontStyleName(style *FontStyle) string {
	if style == nil || len(style.entries) == 0 {
		return ""
	}
	return style.entries[0].name
}
