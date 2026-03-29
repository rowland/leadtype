package ltml

import (
	"strings"
	"testing"

	"github.com/rowland/leadtype/rich_text"
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

func TestLayoutVBox_DirectChildSplitTableUsesOuterOverflowInsteadOfSelfClipping(t *testing.T) {
	page := &StdPage{pageStyle: &PageStyle{width: 8.5 * 72, height: 11 * 72}}
	page.layout = defaultLayouts["vbox"].Clone()
	page.SetAttrs(map[string]string{"units": "in", "margin": "0.5"})

	header := &flowTestWidget{name: "header", preferredHeight: 48}
	_ = header.SetContainer(page)
	header.SetAttrs(map[string]string{"align": "top", "display": "always"})
	page.AddChild(header)

	footer := &flowTestWidget{name: "footer", preferredHeight: 30}
	_ = footer.SetContainer(page)
	footer.SetAttrs(map[string]string{"align": "bottom", "display": "always"})
	page.AddChild(footer)

	table := &StdContainer{}
	_ = table.SetContainer(page)
	table.layout = defaultLayouts["table"].Clone()
	table.order = TableOrderRows
	table.cols = 2
	table.widthPct = 100
	table.splitEnabled = true
	table.splitExplicit = true
	table.headerRows = 1
	page.AddChild(table)

	add := func(height float64) {
		cell := &flowTestWidget{preferredHeight: height}
		_ = cell.SetContainer(table)
		table.AddChild(cell)
	}
	add(24)
	add(24)
	add(280)
	add(280)
	add(280)
	add(280)
	add(280)
	add(280)

	LayoutVBox(page, defaultLayouts["vbox"], &labelTestWriter{t: t})

	if table.Visible() {
		t.Fatalf("expected direct child split table to be hidden for outer-page splitting, got visible with top=%v height=%v bottom=%v", table.Top(), table.Height(), table.Bottom())
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
		t.Fatalf("page count = %d, want 2 (fillRectPages=%v printed=%q plain=%q)", w.pageCount, w.fillRectPages, joinRichTexts(w.printed), strings.Join(w.plainPrinted, "\n"))
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
		t.Fatalf("page count = %d, want 2 (fillRectPages=%v printed=%q plain=%q)", w.pageCount, w.fillRectPages, joinRichTexts(w.printed), strings.Join(w.plainPrinted, "\n"))
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

func TestStdPage_VBoxOverflowDefaultsToTrue(t *testing.T) {
	page := &StdPage{pageStyle: &PageStyle{width: 200, height: 100}}
	page.layout = defaultLayouts["vbox"].Clone()
	doc := newFlowPageDoc(page)

	var body1Pages, body2Pages []int

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
}

func TestStdPage_FlowOverflowDefaultsToTrue(t *testing.T) {
	page := &StdPage{pageStyle: &PageStyle{width: 100, height: 45}}
	page.layout = defaultLayouts["flow"].Clone()
	doc := newFlowPageDoc(page)

	var firstPages, secondPages, thirdPages []int

	first := &flowTestWidget{name: "first", preferredWidth: 60, preferredHeight: 20, printedOn: &firstPages}
	_ = first.SetContainer(page)
	page.AddChild(first)

	second := &flowTestWidget{name: "second", preferredWidth: 60, preferredHeight: 20, printedOn: &secondPages}
	_ = second.SetContainer(page)
	page.AddChild(second)

	third := &flowTestWidget{name: "third", preferredWidth: 60, preferredHeight: 20, printedOn: &thirdPages}
	_ = third.SetContainer(page)
	page.AddChild(third)

	w := &labelTestWriter{t: t}
	if err := doc.Print(w); err != nil {
		t.Fatal(err)
	}
	if w.pageCount != 2 {
		t.Fatalf("page count = %d, want 2", w.pageCount)
	}
	if len(firstPages) != 1 || firstPages[0] != 1 {
		t.Fatalf("first pages = %v, want [1]", firstPages)
	}
	if len(secondPages) != 1 || secondPages[0] != 1 {
		t.Fatalf("second pages = %v, want [1]", secondPages)
	}
	if len(thirdPages) != 1 || thirdPages[0] != 2 {
		t.Fatalf("third pages = %v, want [2]", thirdPages)
	}
}

func TestStdPage_TableOverflowDefaultsToTrue(t *testing.T) {
	page := &StdPage{pageStyle: &PageStyle{width: 200, height: 100}}
	page.layout = defaultLayouts["table"].Clone()
	page.order = TableOrderRows
	page.cols = 2
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

func TestStdPage_ExplicitOverflowFalseDisablesDefaultRetry(t *testing.T) {
	page := &StdPage{pageStyle: &PageStyle{width: 200, height: 100}}
	page.layout = defaultLayouts["vbox"].Clone()
	page.SetAttrs(map[string]string{"overflow": "false"})
	doc := newFlowPageDoc(page)

	var body1Pages, body2Pages []int

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
	if w.pageCount != 1 {
		t.Fatalf("page count = %d, want 1", w.pageCount)
	}
	if len(body1Pages) != 1 || body1Pages[0] != 1 {
		t.Fatalf("body1 pages = %v, want [1]", body1Pages)
	}
	if len(body2Pages) != 0 {
		t.Fatalf("body2 pages = %v, want []", body2Pages)
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
		t.Fatalf("page count = %d, want 2 (fillRectPages=%v printed=%q plain=%q)", w.pageCount, w.fillRectPages, joinRichTexts(w.printed), strings.Join(w.plainPrinted, "\n"))
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
		t.Fatalf("page count = %d, want 2 (fillRectPages=%v printed=%q plain=%q)", w.pageCount, w.fillRectPages, joinRichTexts(w.printed), strings.Join(w.plainPrinted, "\n"))
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
		t.Fatalf("page count = %d, want 2 (fillRectPages=%v printed=%q plain=%q)", w.pageCount, w.fillRectPages, joinRichTexts(w.printed), strings.Join(w.plainPrinted, "\n"))
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

func TestSample_ParagraphSplit_RepeatsHeaderAndSplitsBody(t *testing.T) {
	doc, err := ParseFile(sampleFile("test_027_paragraph_split.ltml"))
	if err != nil {
		t.Fatal(err)
	}
	w := &labelTestWriter{t: t}
	if err := doc.Print(w); err != nil {
		t.Fatal(err)
	}
	if w.pageCount != 3 {
		t.Fatalf("page count = %d, want 3", w.pageCount)
	}
	pageTexts := map[int][]string{}
	for i, rt := range w.printed {
		pageTexts[w.printedPages[i]] = append(pageTexts[w.printedPages[i]], rt.String())
	}
	page1Text := strings.Join(pageTexts[1], "\n")
	page2Text := strings.Join(pageTexts[2], "\n")
	page3Text := strings.Join(pageTexts[3], "\n")
	if !strings.Contains(page1Text, "Paragraph split") || !strings.Contains(page2Text, "Paragraph split") || !strings.Contains(page3Text, "Paragraph split") {
		t.Fatalf("expected repeating paragraph header on all pages, got page1=%q page2=%q page3=%q", page1Text, page2Text, page3Text)
	}
	if !strings.Contains(page1Text, "Odd paragraph footer") || strings.Contains(page1Text, "Even paragraph footer") {
		t.Fatalf("expected only odd paragraph footer on page 1, got %q", page1Text)
	}
	if !strings.Contains(page2Text, "Even paragraph footer") || strings.Contains(page2Text, "Odd paragraph footer") {
		t.Fatalf("expected only even paragraph footer on page 2, got %q", page2Text)
	}
	if !strings.Contains(page3Text, "Odd paragraph footer") || strings.Contains(page3Text, "Even paragraph footer") {
		t.Fatalf("expected only odd paragraph footer on page 3, got %q", page3Text)
	}
	pagePlain := map[int][]string{}
	for i, text := range w.plainPrinted {
		pagePlain[w.plainPages[i]] = append(pagePlain[w.plainPages[i]], text)
	}
	if !strings.Contains(strings.Join(pagePlain[1], "\n"), "*") {
		t.Fatalf("expected bullet on page 1, got %q", strings.Join(pagePlain[1], "\n"))
	}
	if strings.Contains(strings.Join(pagePlain[2], "\n"), "*") || strings.Contains(strings.Join(pagePlain[3], "\n"), "*") {
		t.Fatalf("did not expect bullets on continuation pages, got page2=%q page3=%q", strings.Join(pagePlain[2], "\n"), strings.Join(pagePlain[3], "\n"))
	}
}

func TestStdContainer_SplitForHeight_TableRepeatsHeaderAndFooterRows(t *testing.T) {
	page := &StdPage{pageStyle: &PageStyle{width: 200, height: 200}}
	page.layout = defaultLayouts["vbox"].Clone()

	table := &StdContainer{}
	_ = table.SetContainer(page)
	table.layout = defaultLayouts["table"].Clone()
	table.order = TableOrderRows
	table.cols = 2
	table.width = 180
	table.widthSet = true
	table.splitEnabled = true
	table.splitExplicit = true
	table.headerRows = 1
	table.footerRows = 1
	page.AddChild(table)

	add := func(name string, height float64) {
		cell := &flowTestWidget{name: name, preferredHeight: height}
		_ = cell.SetContainer(table)
		table.AddChild(cell)
	}

	add("header-a", 10)
	add("header-b", 10)
	add("body-1a", 20)
	add("body-1b", 20)
	add("body-2a", 20)
	add("body-2b", 20)
	add("footer-a", 10)
	add("footer-b", 10)

	result, err := table.SplitForHeight(55, &labelTestWriter{t: t})
	if err != nil {
		t.Fatal(err)
	}
	if result == nil {
		t.Fatal("expected split result, got nil")
	}
	head, ok := result.Head.(*StdContainer)
	if !ok {
		t.Fatalf("head type = %T, want *StdContainer", result.Head)
	}
	tail, ok := result.Tail.(*StdContainer)
	if !ok {
		t.Fatalf("tail type = %T, want *StdContainer", result.Tail)
	}

	headNames := []string{}
	for _, child := range head.Widgets() {
		headNames = append(headNames, child.(*flowTestWidget).name)
	}
	wantHead := []string{"header-a", "header-b", "body-1a", "body-1b", "footer-a", "footer-b"}
	if strings.Join(headNames, ",") != strings.Join(wantHead, ",") {
		t.Fatalf("head widgets = %v, want %v", headNames, wantHead)
	}

	tailNames := []string{}
	for _, child := range tail.Widgets() {
		tailNames = append(tailNames, child.(*flowTestWidget).name)
	}
	wantTail := []string{"header-a", "header-b", "body-2a", "body-2b", "footer-a", "footer-b"}
	if strings.Join(tailNames, ",") != strings.Join(wantTail, ",") {
		t.Fatalf("tail widgets = %v, want %v", tailNames, wantTail)
	}
}

func TestSample_TableSplitHeaders_SplitForHeightUsesBodyBand(t *testing.T) {
	doc, err := ParseFile(sampleFile("test_028_table_split_headers.ltml"))
	if err != nil {
		t.Fatal(err)
	}
	page := doc.ltmls[0].Page(0)
	var table *StdContainer
	for _, child := range page.children {
		if candidate, ok := child.(*StdContainer); ok && candidate.LayoutStyle().manager == "table" {
			table = candidate
			break
		}
	}
	if table == nil {
		t.Fatal("expected direct child table")
	}
	w := &labelTestWriter{t: t}
	page.initFlowItems()
	if err := page.preparePhysicalPage(w, true); err != nil {
		t.Fatal(err)
	}
	if table.Visible() {
		t.Fatalf("expected source table to overflow and be hidden before split, got visible with top=%v height=%v", table.Top(), table.Height())
	}
	result, err := table.SplitForHeight(page.availableHeightForChild(table), w)
	if err != nil {
		t.Fatal(err)
	}
	if result == nil {
		t.Fatal("expected table split result, got nil")
	}
	head := result.Head.(*StdContainer)
	tail := result.Tail.(*StdContainer)
	if len(head.Widgets()) >= len(table.Widgets()) {
		t.Fatalf("expected head fragment to contain fewer widgets than source, got head=%d source=%d", len(head.Widgets()), len(table.Widgets()))
	}
	if len(tail.Widgets()) == 0 {
		t.Fatal("expected non-empty tail fragment")
	}
}

func TestSample_TableSplitHeaders_FirstPageLeavesPendingTail(t *testing.T) {
	doc, err := ParseFile(sampleFile("test_028_table_split_headers.ltml"))
	if err != nil {
		t.Fatal(err)
	}
	page := doc.ltmls[0].Page(0)
	w := &labelTestWriter{t: t}
	page.initFlowItems()
	if err := page.preparePhysicalPage(w, true); err != nil {
		t.Fatal(err)
	}
	printedOnce, err := page.drawVisibleChildren(w)
	if err != nil {
		t.Fatal(err)
	}
	if printedOnce == 0 {
		t.Fatal("expected first physical page to make progress")
	}
	if !page.hasPendingOnceChildren() {
		t.Fatal("expected pending once child after first split fragment")
	}
}

func TestSample_TableSplitHeaders_SecondPagePreviewAcceptsTail(t *testing.T) {
	doc, err := ParseFile(sampleFile("test_028_table_split_headers.ltml"))
	if err != nil {
		t.Fatal(err)
	}
	page := doc.ltmls[0].Page(0)
	w := &labelTestWriter{t: t}
	page.initFlowItems()
	if err := page.preparePhysicalPage(w, true); err != nil {
		t.Fatal(err)
	}
	if _, err := page.drawVisibleChildren(w); err != nil {
		t.Fatal(err)
	}
	page.flowPageIndex++
	if err := page.preparePhysicalPage(w, false); err != nil {
		t.Fatalf("expected second physical page preparation to succeed, got %v", err)
	}
}

func TestSample_TableSplitHeadersFooters_FirstFragmentIncludesTableFooterRow(t *testing.T) {
	doc, err := ParseFile(sampleFile("test_029_table_split_headers_footers.ltml"))
	if err != nil {
		t.Fatal(err)
	}
	page := doc.ltmls[0].Page(0)
	var table *StdContainer
	for _, child := range page.children {
		if candidate, ok := child.(*StdContainer); ok && candidate.LayoutStyle().manager == "table" {
			table = candidate
			break
		}
	}
	if table == nil {
		t.Fatal("expected direct child table")
	}
	w := &labelTestWriter{t: t}
	page.initFlowItems()
	if err := page.preparePhysicalPage(w, true); err != nil {
		t.Fatal(err)
	}
	result, err := table.SplitForHeight(page.availableHeightForChild(table), w)
	if err != nil {
		t.Fatal(err)
	}
	if result == nil {
		t.Fatal("expected split result, got nil")
	}
	head := result.Head.(*StdContainer)
	foundFooter := false
	for _, child := range head.Widgets() {
		if para, ok := child.(*StdParagraph); ok && strings.Contains(para.RichText(w).String(), "Carry forward subtotal") {
			foundFooter = true
			break
		}
	}
	if !foundFooter {
		t.Fatal("expected first table fragment to include repeated footer row")
	}
}

func TestSample_TableSplitHeadersFooters_FirstFragmentFooterFitsAbovePageFooter(t *testing.T) {
	doc, err := ParseFile(sampleFile("test_029_table_split_headers_footers.ltml"))
	if err != nil {
		t.Fatal(err)
	}
	page := doc.ltmls[0].Page(0)
	var table *StdContainer
	for _, child := range page.children {
		if candidate, ok := child.(*StdContainer); ok && candidate.LayoutStyle().manager == "table" {
			table = candidate
			break
		}
	}
	if table == nil {
		t.Fatal("expected direct child table")
	}
	w := &labelTestWriter{t: t}
	page.initFlowItems()
	if err := page.preparePhysicalPage(w, true); err != nil {
		t.Fatal(err)
	}
	result, err := table.SplitForHeight(page.availableHeightForChild(table), w)
	if err != nil {
		t.Fatal(err)
	}
	head := result.Head.(*StdContainer)
	page.copySplitGeometry(head, table)
	head.LayoutWidget(w)

	footerTop := 1e9
	for _, child := range page.Widgets() {
		if child.Align() == AlignBottom && child.Visible() && !child.Disabled() {
			if child.Top() < footerTop {
				footerTop = child.Top()
			}
		}
	}
	foundFooter := false
	for _, child := range head.Widgets() {
		para, ok := child.(*StdParagraph)
		if !ok || !strings.Contains(para.RichText(w).String(), "Carry forward subtotal") {
			continue
		}
		foundFooter = true
		if para.Bottom() > footerTop {
			t.Fatalf("first fragment table footer bottom=%v should be above page footer top=%v", para.Bottom(), footerTop)
		}
	}
	if !foundFooter {
		t.Fatal("expected first fragment footer row")
	}
}

func TestSample_TableSplitHeadersFooters_FirstFragmentFooterHasGapAbovePageFooter(t *testing.T) {
	doc, err := ParseFile(sampleFile("test_029_table_split_headers_footers.ltml"))
	if err != nil {
		t.Fatal(err)
	}
	page := doc.ltmls[0].Page(0)
	var table *StdContainer
	for _, child := range page.children {
		if candidate, ok := child.(*StdContainer); ok && candidate.LayoutStyle().manager == "table" {
			table = candidate
			break
		}
	}
	w := &labelTestWriter{t: t}
	page.initFlowItems()
	if err := page.preparePhysicalPage(w, true); err != nil {
		t.Fatal(err)
	}
	result, err := table.SplitForHeight(page.availableHeightForChild(table), w)
	if err != nil {
		t.Fatal(err)
	}
	head := result.Head.(*StdContainer)
	page.copySplitGeometry(head, table)
	head.LayoutWidget(w)

	pageFooterTop := 1e9
	for _, child := range page.Widgets() {
		if child.Align() == AlignBottom && child.Visible() && !child.Disabled() {
			if child.Top() < pageFooterTop {
				pageFooterTop = child.Top()
			}
		}
	}
	tableFooterBottom := 0.0
	for _, child := range head.Widgets() {
		para, ok := child.(*StdParagraph)
		if !ok || !strings.Contains(para.RichText(w).String(), "Carry forward subtotal") {
			continue
		}
		tableFooterBottom = para.Bottom()
	}
	if gap := pageFooterTop - tableFooterBottom; gap < 6 {
		t.Fatalf("gap between table footer and page footer = %vpt, want at least 6pt", gap)
	}
}

func TestSample_TableSplitHeaders_RepeatsPageChromeAndTableHeader(t *testing.T) {
	doc, err := ParseFile(sampleFile("test_028_table_split_headers.ltml"))
	if err != nil {
		t.Fatal(err)
	}
	w := &labelTestWriter{t: t}
	if err := doc.Print(w); err != nil {
		t.Fatal(err)
	}
	if w.pageCount != 2 {
		t.Fatalf("page count = %d, want 2 (fillRectPages=%v printed=%q plain=%q)", w.pageCount, w.fillRectPages, joinRichTexts(w.printed), strings.Join(w.plainPrinted, "\n"))
	}
	if len(w.fillRectPages) != 8 {
		t.Fatalf("filled rect draw count = %d, want 8", len(w.fillRectPages))
	}
	page1Fills, page2Fills := 0, 0
	for _, pageNo := range w.fillRectPages {
		switch pageNo {
		case 1:
			page1Fills++
		case 2:
			page2Fills++
		}
	}
	if page1Fills != 5 || page2Fills != 3 {
		t.Fatalf("filled rect pages = %v, want 5 fills on page 1 and 3 on page 2", w.fillRectPages)
	}
	pageTexts := map[int][]string{}
	for i, rt := range w.printed {
		pageTexts[w.printedPages[i]] = append(pageTexts[w.printedPages[i]], rt.String())
	}
	page1Text := strings.Join(pageTexts[1], "\n")
	page2Text := strings.Join(pageTexts[2], "\n")
	if !strings.Contains(page1Text, "Table split with headers") || !strings.Contains(page2Text, "Table split with headers") {
		t.Fatalf("expected repeating page header on both pages, got page1=%q page2=%q", page1Text, page2Text)
	}
	if !strings.Contains(page1Text, "Line items") || !strings.Contains(page2Text, "Line items") {
		t.Fatalf("expected table header row on both pages, got page1=%q page2=%q", page1Text, page2Text)
	}
	if !strings.Contains(page1Text, "Odd table footer") || strings.Contains(page1Text, "Even table footer") {
		t.Fatalf("expected only odd page footer on page 1, got %q", page1Text)
	}
	if !strings.Contains(page2Text, "Even table footer") || strings.Contains(page2Text, "Odd table footer") {
		t.Fatalf("expected only even page footer on page 2, got %q", page2Text)
	}
}

func TestSample_TableSplitHeadersFooters_RepeatsTableFooterRows(t *testing.T) {
	doc, err := ParseFile(sampleFile("test_029_table_split_headers_footers.ltml"))
	if err != nil {
		t.Fatal(err)
	}
	w := &labelTestWriter{t: t}
	if err := doc.Print(w); err != nil {
		t.Fatal(err)
	}
	if w.pageCount != 2 {
		t.Fatalf("page count = %d, want 2 (fillRectPages=%v printed=%q plain=%q)", w.pageCount, w.fillRectPages, joinRichTexts(w.printed), strings.Join(w.plainPrinted, "\n"))
	}
	if len(w.fillRectPages) != 12 {
		t.Fatalf("filled rect draw count = %d, want 12", len(w.fillRectPages))
	}
	page1Fills, page2Fills := 0, 0
	for _, pageNo := range w.fillRectPages {
		switch pageNo {
		case 1:
			page1Fills++
		case 2:
			page2Fills++
		}
	}
	if page1Fills != 8 || page2Fills != 4 {
		t.Fatalf("filled rect pages = %v, want 8 fills on page 1 and 4 on page 2", w.fillRectPages)
	}
	pageTexts := map[int][]string{}
	for i, rt := range w.printed {
		pageTexts[w.printedPages[i]] = append(pageTexts[w.printedPages[i]], rt.String())
	}
	page1Text := strings.Join(pageTexts[1], "\n")
	page2Text := strings.Join(pageTexts[2], "\n")
	if !strings.Contains(page1Text, "Invoice table split") || !strings.Contains(page2Text, "Invoice table split") {
		t.Fatalf("expected repeating page header on both pages, got page1=%q page2=%q", page1Text, page2Text)
	}
	if !strings.Contains(page1Text, "Description / Amount") || !strings.Contains(page2Text, "Description / Amount") {
		t.Fatalf("expected table header rows on both pages, got page1=%q page2=%q", page1Text, page2Text)
	}
	if !strings.Contains(page1Text, "Carry forward subtotal") || !strings.Contains(page2Text, "Carry forward subtotal") {
		t.Fatalf("expected repeated table footer rows on both pages, got page1=%q page2=%q", page1Text, page2Text)
	}
	if !strings.Contains(page1Text, "Odd invoice footer") || strings.Contains(page1Text, "Even invoice footer") {
		t.Fatalf("expected only odd page footer on page 1, got %q", page1Text)
	}
	if !strings.Contains(page2Text, "Even invoice footer") || strings.Contains(page2Text, "Odd invoice footer") {
		t.Fatalf("expected only even page footer on page 2, got %q", page2Text)
	}
}

func joinRichTexts(texts []*rich_text.RichText) string {
	parts := make([]string, 0, len(texts))
	for _, text := range texts {
		parts = append(parts, text.String())
	}
	return strings.Join(parts, "\n")
}
