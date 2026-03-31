package ltml

import "testing"

func TestParseAndLayout_DirInheritanceAffectsNestedFlowGeometry(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml>
  <page layout="vbox" dir="rtl">
    <div layout="flow" width="200" height="80">
      <rect width="50" height="20" />
      <rect width="50" height="20" />
    </div>
  </page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}

	parsed := doc.ltmls[0]
	page := parsed.Page(0)
	flow, ok := page.Widgets()[0].(*StdContainer)
	if !ok {
		t.Fatalf("page child type = %T, want *StdContainer", page.Widgets()[0])
	}

	if got := page.Dir(); got != DirRTL {
		t.Fatalf("page Dir() = %v, want %v", got, DirRTL)
	}
	if got := flow.Dir(); got != DirRTL {
		t.Fatalf("flow Dir() = %v, want inherited %v", got, DirRTL)
	}

	LayoutContainer(page, &labelTestWriter{t: t})

	wantFlowLeft := ContentRight(page) - flow.Width()
	if got := flow.Left(); got != wantFlowLeft {
		t.Fatalf("flow left = %v, want %v", got, wantFlowLeft)
	}
	if got := flow.Top(); got != 0 {
		t.Fatalf("flow top = %v, want 0", got)
	}

	children := flow.Widgets()
	if len(children) != 2 {
		t.Fatalf("flow child count = %d, want 2", len(children))
	}
	wantFirstLeft := ContentRight(flow) - children[0].Width()
	if got := children[0].Left(); got != wantFirstLeft {
		t.Fatalf("first child left = %v, want %v", got, wantFirstLeft)
	}
	wantSecondLeft := ContentRight(flow) - children[0].Width() - children[1].Width()
	if got := children[1].Left(); got != wantSecondLeft {
		t.Fatalf("second child left = %v, want %v", got, wantSecondLeft)
	}
}

func TestParseAndPrint_RTLFlowOverflowMakesProgressAcrossPages(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml>
  <page layout="flow" dir="rtl" overflow="true">
    <rect width="250" height="200" fill="LightSkyBlue" border="thin" />
    <rect width="250" height="200" fill="PaleGreen" border="thin" />
    <rect width="250" height="200" fill="LightSalmon" border="thin" />
    <rect width="250" height="200" fill="Khaki" border="thin" />
    <rect width="250" height="200" fill="Plum" border="thin" />
    <rect width="250" height="200" fill="LightCyan" border="thin" />
    <rect width="250" height="200" fill="MistyRose" border="thin" />
    <rect width="250" height="200" fill="Lavender" border="thin" />
    <rect width="250" height="200" fill="AliceBlue" border="thin" />
    <rect width="250" height="200" fill="HoneyDew" border="thin" />
  </page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}

	w := &labelTestWriter{t: t}
	if err := doc.Print(w); err != nil {
		t.Fatal(err)
	}

	if got := w.pageCount; got != 2 {
		t.Fatalf("page count = %d, want 2", got)
	}
	if got := doc.ltmls[0].CurrentPhysicalPageNo(); got != 2 {
		t.Fatalf("physical page count = %d, want 2", got)
	}
}
