package ltml

import (
	"strings"
	"testing"
)

type flowTestWidget struct {
	StdWidget
	name            string
	preferredWidth  float64
	preferredHeight float64
	printedOn       *[]int
	layoutCalls     int
}

func (w *flowTestWidget) PreferredWidth(Writer) float64 {
	if w.preferredWidth != 0 {
		return w.preferredWidth
	}
	return 100
}

func (w *flowTestWidget) PreferredHeight(Writer) float64 {
	return w.preferredHeight
}

func (w *flowTestWidget) LayoutWidget(Writer) {
	w.layoutCalls++
}

func (w *flowTestWidget) DrawContent(Writer) error {
	if w.printedOn != nil {
		doc := documentForContainer(w.container)
		if doc != nil {
			*w.printedOn = append(*w.printedOn, doc.CurrentPhysicalPageNo())
		}
	}
	return nil
}

func newFlowPageDoc(page *StdPage) *StdDocument {
	doc := &StdDocument{}
	_ = page.SetContainer(doc)
	doc.AddChild(page)
	return doc
}

func TestLayoutVBox_TopAndBottomChildrenPreserveSourceOrder(t *testing.T) {
	page := &StdPage{pageStyle: &PageStyle{width: 200, height: 200}}
	page.layout = defaultLayouts["vbox"].Clone()

	top1 := &flowTestWidget{name: "top1", preferredHeight: 20}
	_ = top1.SetContainer(page)
	top1.SetAttrs(map[string]string{"align": "top"})
	page.AddChild(top1)

	top2 := &flowTestWidget{name: "top2", preferredHeight: 30}
	_ = top2.SetContainer(page)
	top2.SetAttrs(map[string]string{"align": "top"})
	page.AddChild(top2)

	body := &flowTestWidget{name: "body", preferredHeight: 40}
	_ = body.SetContainer(page)
	page.AddChild(body)

	bottom1 := &flowTestWidget{name: "bottom1", preferredHeight: 15}
	_ = bottom1.SetContainer(page)
	bottom1.SetAttrs(map[string]string{"align": "bottom"})
	page.AddChild(bottom1)

	bottom2 := &flowTestWidget{name: "bottom2", preferredHeight: 25}
	_ = bottom2.SetContainer(page)
	bottom2.SetAttrs(map[string]string{"align": "bottom"})
	page.AddChild(bottom2)

	LayoutVBox(page, &LayoutStyle{manager: "vbox"}, &labelTestWriter{t: t})

	if top1.Top() != 0 {
		t.Fatalf("top1 top = %v, want 0", top1.Top())
	}
	if top2.Top() != top1.Bottom() {
		t.Fatalf("top2 top = %v, want %v", top2.Top(), top1.Bottom())
	}
	if bottom2.Bottom() != 200 {
		t.Fatalf("bottom2 bottom = %v, want 200", bottom2.Bottom())
	}
	if bottom1.Bottom() != bottom2.Top() {
		t.Fatalf("bottom1 bottom = %v, want %v", bottom1.Bottom(), bottom2.Top())
	}
}

func TestLayoutFlow_WrapsAndHidesOverflowingWidgets(t *testing.T) {
	page := &StdPage{pageStyle: &PageStyle{width: 100, height: 45}}
	page.layout = defaultLayouts["flow"].Clone()

	first := &flowTestWidget{name: "first", preferredWidth: 60, preferredHeight: 20}
	_ = first.SetContainer(page)
	page.AddChild(first)

	second := &flowTestWidget{name: "second", preferredWidth: 60, preferredHeight: 20}
	_ = second.SetContainer(page)
	page.AddChild(second)

	third := &flowTestWidget{name: "third", preferredWidth: 60, preferredHeight: 20}
	_ = third.SetContainer(page)
	page.AddChild(third)

	LayoutFlow(page, defaultLayouts["flow"], &labelTestWriter{t: t})

	if !first.Visible() || first.Left() != 0 || first.Top() != 0 {
		t.Fatalf("first = visible:%v left:%v top:%v, want visible at 0,0", first.Visible(), first.Left(), first.Top())
	}
	if !second.Visible() || second.Left() != 0 || second.Top() != 20 {
		t.Fatalf("second = visible:%v left:%v top:%v, want visible at 0,20", second.Visible(), second.Left(), second.Top())
	}
	if third.Visible() {
		t.Fatalf("third should be hidden after overflowing flow layout")
	}
}

func TestStdPage_OverflowStopsWithoutPrintingAnyOnceChild(t *testing.T) {
	page := &StdPage{pageStyle: &PageStyle{width: 200, height: 100}}
	page.layout = defaultLayouts["vbox"].Clone()
	page.overflow = true
	doc := newFlowPageDoc(page)

	var headerPages, bodyPages []int

	header := &flowTestWidget{name: "header", preferredHeight: 80, printedOn: &headerPages}
	_ = header.SetContainer(page)
	header.SetAttrs(map[string]string{"align": "top", "display": "always"})
	page.AddChild(header)

	body := &flowTestWidget{name: "body", preferredHeight: 40, printedOn: &bodyPages}
	_ = body.SetContainer(page)
	page.AddChild(body)

	w := &labelTestWriter{t: t}
	if err := doc.Print(w); err != nil {
		t.Fatal(err)
	}
	if w.pageCount != 1 {
		t.Fatalf("page count = %d, want 1", w.pageCount)
	}
	if len(headerPages) != 1 {
		t.Fatalf("header printed %d times, want 1", len(headerPages))
	}
	if len(bodyPages) != 0 {
		t.Fatalf("body printed %d times, want 0", len(bodyPages))
	}
}

func TestStdPage_OverflowRepeatsAlwaysAndAlternatesOddEven(t *testing.T) {
	page := &StdPage{pageStyle: &PageStyle{width: 200, height: 100}}
	page.layout = defaultLayouts["vbox"].Clone()
	page.overflow = true
	doc := newFlowPageDoc(page)

	var oddPages, evenPages, body1Pages, body2Pages []int

	oddFooter := &flowTestWidget{name: "odd", preferredHeight: 10, printedOn: &oddPages}
	_ = oddFooter.SetContainer(page)
	oddFooter.SetAttrs(map[string]string{"align": "bottom", "display": "odd"})
	page.AddChild(oddFooter)

	evenFooter := &flowTestWidget{name: "even", preferredHeight: 10, printedOn: &evenPages}
	_ = evenFooter.SetContainer(page)
	evenFooter.SetAttrs(map[string]string{"align": "bottom", "display": "even"})
	page.AddChild(evenFooter)

	body1 := &flowTestWidget{name: "body1", preferredHeight: 55, printedOn: &body1Pages}
	_ = body1.SetContainer(page)
	page.AddChild(body1)

	body2 := &flowTestWidget{name: "body2", preferredHeight: 55, printedOn: &body2Pages}
	_ = body2.SetContainer(page)
	page.AddChild(body2)

	w := &labelTestWriter{t: t}
	if err := doc.Print(w); err != nil {
		t.Fatal(err)
	}
	if w.pageCount != 2 {
		t.Fatalf("page count = %d, want 2", w.pageCount)
	}
	if len(body1Pages) != 1 || body1Pages[0] != 1 {
		t.Fatalf("body1 pages = %v, want [1]", body1Pages)
	}
	if len(body2Pages) != 1 || body2Pages[0] != 2 {
		t.Fatalf("body2 pages = %v, want [2]", body2Pages)
	}
	if len(oddPages) != 1 || oddPages[0] != 1 {
		t.Fatalf("odd footer pages = %v, want [1]", oddPages)
	}
	if len(evenPages) != 1 || evenPages[0] != 2 {
		t.Fatalf("even footer pages = %v, want [2]", evenPages)
	}
}

func TestStdPage_TableOverflowDefersWholeRow(t *testing.T) {
	page := &StdPage{pageStyle: &PageStyle{width: 200, height: 100}}
	page.layout = defaultLayouts["table"].Clone()
	page.order = TableOrderRows
	page.cols = 2
	page.overflow = true
	doc := newFlowPageDoc(page)

	pages := make([][]int, 6)
	for i := 0; i < 6; i++ {
		cell := &flowTestWidget{
			name:            "cell",
			preferredHeight: 45,
			printedOn:       &pages[i],
		}
		_ = cell.SetContainer(page)
		page.AddChild(cell)
	}

	w := &labelTestWriter{t: t}
	if err := doc.Print(w); err != nil {
		t.Fatal(err)
	}
	if w.pageCount != 2 {
		t.Fatalf("page count = %d, want 2", w.pageCount)
	}
	for i := 0; i < 4; i++ {
		if len(pages[i]) != 1 || pages[i][0] != 1 {
			t.Fatalf("cell %d pages = %v, want [1]", i, pages[i])
		}
	}
	for i := 4; i < 6; i++ {
		if len(pages[i]) != 1 || pages[i][0] != 2 {
			t.Fatalf("cell %d pages = %v, want [2]", i, pages[i])
		}
	}
}

func TestSample_VBoxOverflow_PrintsHeaderOnFirstPageAndBodyAcrossPages(t *testing.T) {
	doc, err := ParseFile(sampleFile("test_024_vbox_overflow.ltml"))
	if err != nil {
		t.Fatal(err)
	}
	w := &labelTestWriter{t: t}
	if err := doc.Print(w); err != nil {
		t.Fatal(err)
	}
	if w.pageCount != 2 {
		t.Fatalf("page count = %d, want 2", w.pageCount)
	}
	if len(w.fillRectPages) != 3 {
		t.Fatalf("filled rect draw count = %d, want 3", len(w.fillRectPages))
	}
	if w.fillRectPages[0] != 1 || w.fillRectPages[1] != 1 || w.fillRectPages[2] != 2 {
		t.Fatalf("filled rect pages = %v, want [1 1 2]", w.fillRectPages)
	}
	var texts []string
	pageTexts := map[int][]string{}
	for _, rt := range w.printed {
		texts = append(texts, rt.String())
	}
	for i, rt := range w.printed {
		pageTexts[w.printedPages[i]] = append(pageTexts[w.printedPages[i]], rt.String())
	}
	allText := strings.Join(texts, "\n")
	if !strings.Contains(allText, "Repeating header") {
		t.Fatalf("expected header text on first page, got %q", allText)
	}
	if !strings.Contains(allText, "Odd page footer") {
		t.Fatalf("expected odd footer text to print, got %q", allText)
	}
	page1Text := strings.Join(pageTexts[1], "\n")
	page2Text := strings.Join(pageTexts[2], "\n")
	if !strings.Contains(page1Text, "This boilerplate appears only on the first page.") {
		t.Fatalf("expected boilerplate on first page, got %q", page1Text)
	}
	if strings.Contains(page2Text, "This boilerplate appears only on the first page.") {
		t.Fatalf("did not expect boilerplate on second page, got %q", page2Text)
	}
	if !strings.Contains(page2Text, "Repeating header") {
		t.Fatalf("expected repeating header on second page, got %q", page2Text)
	}
	if !strings.Contains(page1Text, "Odd page footer") || strings.Contains(page1Text, "Even page footer") {
		t.Fatalf("expected only odd footer on page 1, got %q", page1Text)
	}
	if !strings.Contains(page2Text, "Even page footer") || strings.Contains(page2Text, "Odd page footer") {
		t.Fatalf("expected only even footer on page 2, got %q", page2Text)
	}
}

func TestSample_FlowOverflow_RepeatsBannerAndCarriesRemainingWidgets(t *testing.T) {
	doc, err := ParseFile(sampleFile("test_025_flow_overflow.ltml"))
	if err != nil {
		t.Fatal(err)
	}
	w := &labelTestWriter{t: t}
	if err := doc.Print(w); err != nil {
		t.Fatal(err)
	}
	if w.pageCount != 2 {
		t.Fatalf("page count = %d, want 2", w.pageCount)
	}
	if len(w.fillRectPages) != 7 {
		t.Fatalf("filled rect draw count = %d, want 7", len(w.fillRectPages))
	}
	if w.fillRectPages[0] != 1 || w.fillRectPages[1] != 1 || w.fillRectPages[2] != 1 || w.fillRectPages[3] != 1 || w.fillRectPages[4] != 1 || w.fillRectPages[5] != 1 || w.fillRectPages[6] != 2 {
		t.Fatalf("filled rect pages = %v, want [1 1 1 1 1 1 2]", w.fillRectPages)
	}
	pageTexts := map[int][]string{}
	for i, rt := range w.printed {
		pageTexts[w.printedPages[i]] = append(pageTexts[w.printedPages[i]], rt.String())
	}
	page1Text := strings.Join(pageTexts[1], "\n")
	page2Text := strings.Join(pageTexts[2], "\n")
	if !strings.Contains(page1Text, "Flow banner") || !strings.Contains(page2Text, "Flow banner") {
		t.Fatalf("expected repeating flow banner on both pages, got page1=%q page2=%q", page1Text, page2Text)
	}
	if !strings.Contains(page1Text, "Intro text appears only once.") {
		t.Fatalf("expected intro text on first page, got %q", page1Text)
	}
	if strings.Contains(page2Text, "Intro text appears only once.") {
		t.Fatalf("did not expect intro text on second page, got %q", page2Text)
	}
}

func TestSample_TableOverflow_DefersWholeRows(t *testing.T) {
	doc, err := ParseFile(sampleFile("test_026_table_overflow.ltml"))
	if err != nil {
		t.Fatal(err)
	}
	w := &labelTestWriter{t: t}
	if err := doc.Print(w); err != nil {
		t.Fatal(err)
	}
	if w.pageCount != 2 {
		t.Fatalf("page count = %d, want 2", w.pageCount)
	}
	if len(w.fillRectPages) != 6 {
		t.Fatalf("filled rect draw count = %d, want 6", len(w.fillRectPages))
	}
	if w.fillRectPages[0] != 1 || w.fillRectPages[1] != 1 || w.fillRectPages[2] != 1 || w.fillRectPages[3] != 1 || w.fillRectPages[4] != 2 || w.fillRectPages[5] != 2 {
		t.Fatalf("filled rect pages = %v, want [1 1 1 1 2 2]", w.fillRectPages)
	}
	pageTexts := map[int][]string{}
	for i, rt := range w.printed {
		pageTexts[w.printedPages[i]] = append(pageTexts[w.printedPages[i]], rt.String())
	}
	page1Text := strings.Join(pageTexts[1], "\n")
	page2Text := strings.Join(pageTexts[2], "\n")
	if !strings.Contains(page1Text, "Repeating table header") || !strings.Contains(page2Text, "Repeating table header") {
		t.Fatalf("expected repeating table header on both pages, got page1=%q page2=%q", page1Text, page2Text)
	}
	if !strings.Contains(page1Text, "Odd page table footer") || strings.Contains(page1Text, "Even page table footer") {
		t.Fatalf("expected only odd table footer on page 1, got %q", page1Text)
	}
	if !strings.Contains(page2Text, "Even page table footer") || strings.Contains(page2Text, "Odd page table footer") {
		t.Fatalf("expected only even table footer on page 2, got %q", page2Text)
	}
}
