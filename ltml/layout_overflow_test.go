package ltml

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/rowland/leadtype/rich_text"
)

type flowTestWidget struct {
	StdWidget
	name             string
	preferredWidth   float64
	preferredHeight  float64
	printedOn        *[]int
	printedHeights   *[]float64
	layoutCalls      int
	preferredWidths  int
	preferredHeights int
}

func (w *flowTestWidget) PreferredWidth(Writer) float64 {
	w.preferredWidths++
	if w.preferredWidth != 0 {
		return w.preferredWidth
	}
	return 100
}

func (w *flowTestWidget) PreferredHeight(Writer) float64 {
	w.preferredHeights++
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
	if w.printedHeights != nil {
		*w.printedHeights = append(*w.printedHeights, w.Height())
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
	table.SetWidthPct(100)
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

func TestLayoutVBox_SkipsBodyChildrenAfterFirstOverflow(t *testing.T) {
	page := &StdPage{pageStyle: &PageStyle{width: 200, height: 100}}
	page.layout = &LayoutStyle{manager: "vbox"}

	first := &flowTestWidget{name: "first", preferredHeight: 40}
	_ = first.SetContainer(page)
	page.AddChild(first)

	overflow := &flowTestWidget{name: "overflow", preferredHeight: 70}
	_ = overflow.SetContainer(page)
	page.AddChild(overflow)

	tail := &flowTestWidget{name: "tail", preferredHeight: 10}
	_ = tail.SetContainer(page)
	page.AddChild(tail)

	LayoutVBox(page, page.layout, &labelTestWriter{t: t})

	if !first.Visible() || first.layoutCalls != 1 {
		t.Fatalf("first visible/layoutCalls = %v/%d, want visible with one layout", first.Visible(), first.layoutCalls)
	}
	if overflow.Visible() || overflow.layoutCalls != 0 {
		t.Fatalf("overflow visible/layoutCalls = %v/%d, want hidden with no layout", overflow.Visible(), overflow.layoutCalls)
	}
	if tail.Visible() || tail.layoutCalls != 0 {
		t.Fatalf("tail visible/layoutCalls = %v/%d, want hidden with no layout", tail.Visible(), tail.layoutCalls)
	}
	if tail.preferredWidths != 0 || tail.preferredHeights != 0 {
		t.Fatalf("tail preferred calls = width:%d height:%d, want none", tail.preferredWidths, tail.preferredHeights)
	}
}

func TestLayoutVBox_AutoHeightDistributionUsesVisibleFragmentOnly(t *testing.T) {
	page := &StdPage{pageStyle: &PageStyle{width: 200, height: 100}}
	page.layout = &LayoutStyle{manager: "vbox"}

	fixed := &flowTestWidget{name: "fixed", preferredHeight: 40}
	_ = fixed.SetContainer(page)
	page.AddChild(fixed)

	visibleAuto := &flowTestWidget{name: "visibleAuto", preferredHeight: 10}
	visibleAuto.SetHeightAuto()
	_ = visibleAuto.SetContainer(page)
	page.AddChild(visibleAuto)

	overflow := &flowTestWidget{name: "overflow", preferredHeight: 60}
	overflow.SetHeightAuto()
	_ = overflow.SetContainer(page)
	page.AddChild(overflow)

	tailAuto := &flowTestWidget{name: "tailAuto", preferredHeight: 10}
	tailAuto.SetHeightAuto()
	_ = tailAuto.SetContainer(page)
	page.AddChild(tailAuto)

	LayoutVBox(page, page.layout, &labelTestWriter{t: t})

	if !visibleAuto.Visible() || visibleAuto.Height() != 60 {
		t.Fatalf("visible auto visible/height = %v/%v, want visible height 60", visibleAuto.Visible(), visibleAuto.Height())
	}
	if overflow.Visible() || overflow.Height() != 60 {
		t.Fatalf("overflow visible/height = %v/%v, want hidden at preferred height 60", overflow.Visible(), overflow.Height())
	}
	if tailAuto.Visible() || tailAuto.preferredHeights != 0 || tailAuto.Height() != 0 {
		t.Fatalf("tail auto visible/preferredHeights/height = %v/%d/%v, want hidden, unmeasured, zero height", tailAuto.Visible(), tailAuto.preferredHeights, tailAuto.Height())
	}
}

func TestStdPage_OverflowNoProgressReturnsLayoutOverflowError(t *testing.T) {
	page := &StdPage{pageStyle: &PageStyle{width: 200, height: 100}}
	page.layout = defaultLayouts["vbox"].Clone()
	page.overflow = true
	doc := newFlowPageDoc(page)

	first := &flowTestWidget{name: "first", preferredHeight: 20}
	_ = first.SetContainer(page)
	page.AddChild(first)

	second := &flowTestWidget{name: "second", preferredHeight: 120}
	_ = second.SetContainer(page)
	page.AddChild(second)

	w := &labelTestWriter{t: t}
	err := doc.Print(w)
	var overflowErr *LayoutOverflowError
	if !errors.As(err, &overflowErr) {
		t.Fatalf("Print error = %v, want LayoutOverflowError", err)
	}
	if overflowErr.AvailableHeight != 100 || overflowErr.RequiredHeight != 120 {
		t.Fatalf("overflow sizes = available %v required %v, want 100 and 120", overflowErr.AvailableHeight, overflowErr.RequiredHeight)
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

func TestStdPage_OverflowRepeatsNestedAlwaysChrome(t *testing.T) {
	page := &StdPage{pageStyle: &PageStyle{width: 200, height: 100}}
	page.layout = defaultLayouts["vbox"].Clone()
	page.overflow = true
	doc := newFlowPageDoc(page)

	var headerPages, footerPages, body1Pages, body2Pages []int

	header := &StdContainer{}
	_ = header.SetContainer(page)
	header.SetAttrs(map[string]string{"align": "top", "display": "always"})
	page.AddChild(header)
	headerInner := &StdContainer{}
	_ = headerInner.SetContainer(header)
	header.AddChild(headerInner)
	headerLeaf := &flowTestWidget{name: "header", preferredHeight: 10, printedOn: &headerPages}
	_ = headerLeaf.SetContainer(headerInner)
	headerInner.AddChild(headerLeaf)

	body1 := &flowTestWidget{name: "body1", preferredHeight: 55, printedOn: &body1Pages}
	_ = body1.SetContainer(page)
	page.AddChild(body1)

	body2 := &flowTestWidget{name: "body2", preferredHeight: 55, printedOn: &body2Pages}
	_ = body2.SetContainer(page)
	page.AddChild(body2)

	footer := &StdContainer{}
	_ = footer.SetContainer(page)
	footer.SetAttrs(map[string]string{"align": "bottom", "display": "always"})
	page.AddChild(footer)
	footerInner := &StdContainer{}
	_ = footerInner.SetContainer(footer)
	footer.AddChild(footerInner)
	footerLeaf := &flowTestWidget{name: "footer", preferredHeight: 10, printedOn: &footerPages}
	_ = footerLeaf.SetContainer(footerInner)
	footerInner.AddChild(footerLeaf)

	w := &labelTestWriter{t: t}
	if err := doc.Print(w); err != nil {
		t.Fatal(err)
	}
	if w.pageCount != 2 {
		t.Fatalf("page count = %d, want 2", w.pageCount)
	}
	if !slices.Equal(headerPages, []int{1, 2}) {
		t.Fatalf("header pages = %v, want [1 2]", headerPages)
	}
	if !slices.Equal(footerPages, []int{1, 2}) {
		t.Fatalf("footer pages = %v, want [1 2]", footerPages)
	}
	if len(body1Pages) != 1 || body1Pages[0] != 1 {
		t.Fatalf("body1 pages = %v, want [1]", body1Pages)
	}
	if len(body2Pages) != 1 || body2Pages[0] != 2 {
		t.Fatalf("body2 pages = %v, want [2]", body2Pages)
	}
}

func TestStdPage_RepeatedChromeDoesNotReserveHiddenNestedFirstChild(t *testing.T) {
	page := &StdPage{pageStyle: &PageStyle{width: 200, height: 100}}
	page.layout = defaultLayouts["vbox"].Clone()
	page.overflow = true
	doc := newFlowPageDoc(page)

	var titlePages, introPages, body1Pages, body2Pages []int

	header := &StdContainer{}
	_ = header.SetContainer(page)
	header.SetAttrs(map[string]string{"align": "top", "display": "always"})
	page.AddChild(header)

	title := &flowTestWidget{name: "title", preferredHeight: 10, printedOn: &titlePages}
	_ = title.SetContainer(header)
	header.AddChild(title)

	intro := &flowTestWidget{name: "intro", preferredHeight: 30, printedOn: &introPages}
	_ = intro.SetContainer(header)
	intro.SetAttrs(map[string]string{"display": "first"})
	header.AddChild(intro)

	body1 := &flowTestWidget{name: "body1", preferredHeight: 45, printedOn: &body1Pages}
	_ = body1.SetContainer(page)
	page.AddChild(body1)

	body2 := &flowTestWidget{name: "body2", preferredHeight: 45, printedOn: &body2Pages}
	_ = body2.SetContainer(page)
	page.AddChild(body2)

	w := &labelTestWriter{t: t}
	if err := doc.Print(w); err != nil {
		t.Fatal(err)
	}
	if w.pageCount != 2 {
		t.Fatalf("page count = %d, want 2", w.pageCount)
	}
	if !slices.Equal(titlePages, []int{1, 2}) {
		t.Fatalf("title pages = %v, want [1 2]", titlePages)
	}
	if !slices.Equal(introPages, []int{1}) {
		t.Fatalf("intro pages = %v, want [1]", introPages)
	}
	if !slices.Equal(body1Pages, []int{1}) || !slices.Equal(body2Pages, []int{2}) {
		t.Fatalf("body pages = %v/%v, want [1]/[2]", body1Pages, body2Pages)
	}
	if body2 := page.flowItems[1].Source; body2.Top() != 10 {
		t.Fatalf("body2 top = %v, want 10 after title-only repeated header", body2.Top())
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

func TestStdPage_FixedHeightVBoxSplitDistributesAutoHeightPerPage(t *testing.T) {
	page := &StdPage{pageStyle: &PageStyle{width: 200, height: 100}}
	page.layout = defaultLayouts["vbox"].Clone()
	page.overflow = true
	doc := newFlowPageDoc(page)

	box := &StdContainer{}
	_ = box.SetContainer(page)
	box.layout = defaultLayouts["vbox"].Clone()
	box.SetWidth(180)
	box.SetHeight(140)
	page.AddChild(box)

	row1 := &flowTestWidget{name: "row1", preferredHeight: 30}
	row1.SetHeight(30)
	_ = row1.SetContainer(box)
	box.AddChild(row1)

	var auto1Pages, auto2Pages, auto3Pages []int
	var auto1Heights, auto2Heights, auto3Heights []float64

	auto1 := &flowTestWidget{name: "auto1", preferredHeight: 10, printedOn: &auto1Pages, printedHeights: &auto1Heights}
	auto1.SetHeightAuto()
	_ = auto1.SetContainer(box)
	box.AddChild(auto1)

	row3 := &flowTestWidget{name: "row3", preferredHeight: 20}
	_ = row3.SetContainer(box)
	box.AddChild(row3)

	auto2 := &flowTestWidget{name: "auto2", preferredHeight: 10, printedOn: &auto2Pages, printedHeights: &auto2Heights}
	auto2.SetHeightAuto()
	_ = auto2.SetContainer(box)
	box.AddChild(auto2)

	row5 := &flowTestWidget{name: "row5", preferredHeight: 20}
	_ = row5.SetContainer(box)
	box.AddChild(row5)

	row6 := &flowTestWidget{name: "row6", preferredHeight: 20}
	_ = row6.SetContainer(box)
	box.AddChild(row6)

	auto3 := &flowTestWidget{name: "auto3", preferredHeight: 10, printedOn: &auto3Pages, printedHeights: &auto3Heights}
	auto3.SetHeightAuto()
	_ = auto3.SetContainer(box)
	box.AddChild(auto3)

	w := &labelTestWriter{t: t}
	if err := doc.Print(w); err != nil {
		t.Fatal(err)
	}
	if w.pageCount != 2 {
		t.Fatalf("page count = %d, want 2", w.pageCount)
	}
	if !slices.Equal(auto1Pages, []int{1}) || len(auto1Heights) != 1 || auto1Heights[0] != 15 {
		t.Fatalf("auto1 = pages:%v heights:%v, want page 1 height 15", auto1Pages, auto1Heights)
	}
	if !slices.Equal(auto2Pages, []int{1}) || len(auto2Heights) != 1 || auto2Heights[0] != 15 {
		t.Fatalf("auto2 = pages:%v heights:%v, want page 1 height 15", auto2Pages, auto2Heights)
	}
	if !slices.Equal(auto3Pages, []int{2}) || len(auto3Heights) != 1 || auto3Heights[0] != 80 {
		t.Fatalf("auto3 = pages:%v heights:%v, want page 2 height 80", auto3Pages, auto3Heights)
	}
}

func TestStdPage_FixedHeightVBoxSplitOnlySharesOnPagesWithAutoChildren(t *testing.T) {
	page := &StdPage{pageStyle: &PageStyle{width: 200, height: 100}}
	page.layout = defaultLayouts["vbox"].Clone()
	page.overflow = true
	doc := newFlowPageDoc(page)

	box := &StdContainer{}
	_ = box.SetContainer(page)
	box.layout = defaultLayouts["vbox"].Clone()
	box.SetWidth(180)
	box.SetHeight(140)
	page.AddChild(box)

	var row1Heights, row2Heights, autoHeights []float64
	var row1Pages, row2Pages, autoPages []int

	row1 := &flowTestWidget{name: "row1", preferredHeight: 50, printedOn: &row1Pages, printedHeights: &row1Heights}
	_ = row1.SetContainer(box)
	box.AddChild(row1)

	row2 := &flowTestWidget{name: "row2", preferredHeight: 50, printedOn: &row2Pages, printedHeights: &row2Heights}
	_ = row2.SetContainer(box)
	box.AddChild(row2)

	auto := &flowTestWidget{name: "auto", preferredHeight: 10, printedOn: &autoPages, printedHeights: &autoHeights}
	auto.SetHeightAuto()
	_ = auto.SetContainer(box)
	box.AddChild(auto)

	row4 := &flowTestWidget{name: "row4", preferredHeight: 20}
	_ = row4.SetContainer(box)
	box.AddChild(row4)

	w := &labelTestWriter{t: t}
	if err := doc.Print(w); err != nil {
		t.Fatal(err)
	}
	if w.pageCount != 2 {
		t.Fatalf("page count = %d, want 2", w.pageCount)
	}
	if !slices.Equal(row1Pages, []int{1}) || len(row1Heights) != 1 || row1Heights[0] != 50 {
		t.Fatalf("row1 = pages:%v heights:%v, want page 1 height 50", row1Pages, row1Heights)
	}
	if !slices.Equal(row2Pages, []int{1}) || len(row2Heights) != 1 || row2Heights[0] != 50 {
		t.Fatalf("row2 = pages:%v heights:%v, want page 1 height 50", row2Pages, row2Heights)
	}
	if !slices.Equal(autoPages, []int{2}) || len(autoHeights) != 1 || autoHeights[0] != 80 {
		t.Fatalf("auto = pages:%v heights:%v, want page 2 height 80", autoPages, autoHeights)
	}
}

func TestStdPage_NaturalVBoxSplitLeavesAutoHeightAtPreferredSize(t *testing.T) {
	page := &StdPage{pageStyle: &PageStyle{width: 200, height: 100}}
	page.layout = defaultLayouts["vbox"].Clone()
	page.overflow = true
	doc := newFlowPageDoc(page)

	box := &StdContainer{}
	_ = box.SetContainer(page)
	box.layout = defaultLayouts["vbox"].Clone()
	box.SetWidth(180)
	page.AddChild(box)

	row1 := &flowTestWidget{name: "row1", preferredHeight: 60}
	_ = row1.SetContainer(box)
	box.AddChild(row1)

	var autoPages []int
	var autoHeights []float64
	auto := &flowTestWidget{name: "auto", preferredHeight: 10, printedOn: &autoPages, printedHeights: &autoHeights}
	auto.SetHeightAuto()
	_ = auto.SetContainer(box)
	box.AddChild(auto)

	row3 := &flowTestWidget{name: "row3", preferredHeight: 60}
	_ = row3.SetContainer(box)
	box.AddChild(row3)

	w := &labelTestWriter{t: t}
	if err := doc.Print(w); err != nil {
		t.Fatal(err)
	}
	if w.pageCount != 2 {
		t.Fatalf("page count = %d, want 2", w.pageCount)
	}
	if !slices.Equal(autoPages, []int{1}) || len(autoHeights) != 1 || autoHeights[0] != 10 {
		t.Fatalf("auto = pages:%v heights:%v, want page 1 height 10", autoPages, autoHeights)
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
	if w.pageCount != 4 {
		t.Fatalf("page count = %d, want 4", w.pageCount)
	}
	pageTexts := map[int][]string{}
	for i, rt := range w.printed {
		pageTexts[w.printedPages[i]] = append(pageTexts[w.printedPages[i]], rt.String())
	}
	page1Text := strings.Join(pageTexts[1], "\n")
	page2Text := strings.Join(pageTexts[2], "\n")
	page3Text := strings.Join(pageTexts[3], "\n")
	page4Text := strings.Join(pageTexts[4], "\n")
	if !strings.Contains(page1Text, "Paragraph split") || !strings.Contains(page2Text, "Paragraph split") || !strings.Contains(page3Text, "Paragraph split") || !strings.Contains(page4Text, "Paragraph split") {
		t.Fatalf("expected repeating paragraph header on all pages, got page1=%q page2=%q page3=%q page4=%q", page1Text, page2Text, page3Text, page4Text)
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
	if !strings.Contains(page4Text, "Even paragraph footer") || strings.Contains(page4Text, "Odd paragraph footer") {
		t.Fatalf("expected only even paragraph footer on page 4, got %q", page4Text)
	}
	pagePlain := map[int][]string{}
	for i, text := range w.plainPrinted {
		pagePlain[w.plainPages[i]] = append(pagePlain[w.plainPages[i]], text)
	}
	if !strings.Contains(strings.Join(pagePlain[1], "\n"), "*") {
		t.Fatalf("expected bullet on page 1, got %q", strings.Join(pagePlain[1], "\n"))
	}
	if strings.Contains(strings.Join(pagePlain[2], "\n"), "*") || strings.Contains(strings.Join(pagePlain[3], "\n"), "*") || strings.Contains(strings.Join(pagePlain[4], "\n"), "*") {
		t.Fatalf("did not expect bullets on continuation pages, got page2=%q page3=%q page4=%q", strings.Join(pagePlain[2], "\n"), strings.Join(pagePlain[3], "\n"), strings.Join(pagePlain[4], "\n"))
	}
}

func TestStdContainer_SplitForHeight_VBoxRepeatsHeadersAndFooters(t *testing.T) {
	page := &StdPage{pageStyle: &PageStyle{width: 200, height: 200}}
	page.layout = defaultLayouts["vbox"].Clone()

	box := &StdContainer{}
	_ = box.SetContainer(page)
	box.layout = defaultLayouts["vbox"].Clone()
	box.SetWidth(180)
	page.AddChild(box)

	add := func(name string, height float64, align string) {
		child := &flowTestWidget{name: name, preferredHeight: height}
		_ = child.SetContainer(box)
		if align != "" {
			child.SetAttrs(map[string]string{"align": align})
		}
		box.AddChild(child)
	}

	add("header", 10, "top")
	add("body-1", 20, "")
	add("body-2", 20, "")
	add("footer", 10, "bottom")

	result, err := box.SplitForHeight(40, &labelTestWriter{t: t})
	if err != nil {
		t.Fatal(err)
	}
	if result == nil {
		t.Fatal("expected split result, got nil")
	}
	head := result.Head.(*StdContainer)
	tail := result.Tail.(*StdContainer)

	var headNames []string
	for _, child := range head.Widgets() {
		headNames = append(headNames, child.(*flowTestWidget).name)
	}
	if got, want := strings.Join(headNames, ","), "header,body-1,footer"; got != want {
		t.Fatalf("head widgets = %q, want %q", got, want)
	}

	var tailNames []string
	for _, child := range tail.Widgets() {
		tailNames = append(tailNames, child.(*flowTestWidget).name)
	}
	if got, want := strings.Join(tailNames, ","), "header,body-2,footer"; got != want {
		t.Fatalf("tail widgets = %q, want %q", got, want)
	}
}

func TestStdContainer_SplitForHeight_VBoxSkipsTailMeasurements(t *testing.T) {
	page := &StdPage{pageStyle: &PageStyle{width: 200, height: 200}}
	page.layout = &LayoutStyle{manager: "vbox"}

	box := &StdContainer{}
	_ = box.SetContainer(page)
	box.layout = &LayoutStyle{manager: "vbox"}
	box.SetWidth(180)
	page.AddChild(box)

	first := &flowTestWidget{name: "first", preferredHeight: 40}
	_ = first.SetContainer(box)
	box.AddChild(first)

	overflow := &flowTestWidget{name: "overflow", preferredHeight: 70}
	_ = overflow.SetContainer(box)
	box.AddChild(overflow)

	tail := &flowTestWidget{name: "tail", preferredHeight: 10}
	_ = tail.SetContainer(box)
	box.AddChild(tail)

	result, err := box.SplitForHeight(100, &labelTestWriter{t: t})
	if err != nil {
		t.Fatal(err)
	}
	if result == nil || result.Head == nil || result.Tail == nil {
		t.Fatalf("split result = %#v, want head and tail", result)
	}
	if tail.preferredWidths != 0 || tail.preferredHeights != 0 {
		t.Fatalf("tail preferred calls = width:%d height:%d, want none", tail.preferredWidths, tail.preferredHeights)
	}
	head := result.Head.(*StdContainer)
	tailFragment := result.Tail.(*StdContainer)
	if got, want := len(head.Widgets()), 1; got != want {
		t.Fatalf("head child count = %d, want %d", got, want)
	}
	if got, want := len(tailFragment.Widgets()), 2; got != want {
		t.Fatalf("tail child count = %d, want %d", got, want)
	}
}

func TestStdContainer_SplitForHeight_VBoxSplitsBodyParagraph(t *testing.T) {
	page := &StdPage{pageStyle: &PageStyle{width: 200, height: 200}}
	page.layout = defaultLayouts["vbox"].Clone()

	box := &StdContainer{}
	_ = box.SetContainer(page)
	box.layout = defaultLayouts["vbox"].Clone()
	box.SetWidth(120)
	page.AddChild(box)

	header := &flowTestWidget{name: "header", preferredHeight: 10}
	_ = header.SetContainer(box)
	header.SetAttrs(map[string]string{"align": "top"})
	box.AddChild(header)

	para := &StdParagraph{}
	_ = para.SetContainer(box)
	para.font = &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 12}
	para.splitDisabled = false
	para.bullet = &BulletStyle{text: "*", width: 18, font: para.font}
	para.AddText("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.")
	box.AddChild(para)

	footer := &flowTestWidget{name: "footer", preferredHeight: 10}
	_ = footer.SetContainer(box)
	footer.SetAttrs(map[string]string{"align": "bottom"})
	box.AddChild(footer)

	w := &labelTestWriter{t: t, fonts: defaultTestFonts(t), lineSpacing: 1.0}
	result, err := box.SplitForHeight(80, w)
	if err != nil {
		t.Fatal(err)
	}
	if result == nil {
		t.Fatal("expected split result, got nil")
	}

	head := result.Head.(*StdContainer)
	tail := result.Tail.(*StdContainer)
	headPara, ok := head.Widgets()[1].(*StdParagraph)
	if !ok {
		t.Fatalf("head body type = %T, want *StdParagraph", head.Widgets()[1])
	}
	tailPara, ok := tail.Widgets()[1].(*StdParagraph)
	if !ok {
		t.Fatalf("tail body type = %T, want *StdParagraph", tail.Widgets()[1])
	}
	if len(headPara.splitLines) == 0 || len(tailPara.splitLines) == 0 {
		t.Fatalf("expected split paragraph fragments, got head=%d tail=%d", len(headPara.splitLines), len(tailPara.splitLines))
	}
	if headPara.suppressBullet {
		t.Fatal("expected head paragraph to keep its bullet")
	}
	if !tailPara.suppressBullet {
		t.Fatal("expected tail paragraph to suppress its continuation bullet")
	}
}

func TestStdContainer_SplitForHeight_VBoxReturnsNilWhenRepeatedChromeConsumesPage(t *testing.T) {
	page := &StdPage{pageStyle: &PageStyle{width: 200, height: 200}}
	page.layout = defaultLayouts["vbox"].Clone()

	box := &StdContainer{}
	_ = box.SetContainer(page)
	box.layout = defaultLayouts["vbox"].Clone()
	box.SetWidth(180)
	page.AddChild(box)

	header := &flowTestWidget{name: "header", preferredHeight: 20}
	_ = header.SetContainer(box)
	header.SetAttrs(map[string]string{"align": "top"})
	box.AddChild(header)

	body := &flowTestWidget{name: "body", preferredHeight: 20}
	_ = body.SetContainer(box)
	box.AddChild(body)

	footer := &flowTestWidget{name: "footer", preferredHeight: 20}
	_ = footer.SetContainer(box)
	footer.SetAttrs(map[string]string{"align": "bottom"})
	box.AddChild(footer)

	result, err := box.SplitForHeight(35, &labelTestWriter{t: t})
	if err != nil {
		t.Fatal(err)
	}
	if result != nil {
		t.Fatalf("expected nil split result, got %#v", result)
	}
}

func TestSample_ListSplit_RepeatsPageChromeAndContinuesOrderedMarkers(t *testing.T) {
	doc := parseDoc(t, `
		<ltml>
			<pagestyle id="tiny" width="200pt" height="120pt" />
			<page style="tiny" margin="0" font.name="Helvetica" font.size="12" layout="vbox">
				<div display="always" align="top" padding="4pt">
					<label display="always" font.weight="Bold">List splitting</label>
				</div>
				<ol padding="4pt">
					<p>
						First ordered item starts the list and leaves room for later items to
						overflow onto following pages.
					</p>
					<p>Second ordered item continues the numbering after the split.</p>
					<p>Third ordered item confirms numbering remains stable on later fragments.</p>
					<p>Fourth ordered item keeps the decimal sequence going.</p>
					<p>Fifth ordered item leaves enough content to require another page.</p>
					<p>Sixth ordered item should not restart at one.</p>
					<p>Seventh ordered item remains aligned with the rest of the list.</p>
					<p>Eighth ordered item gives the next page more than one ordinal.</p>
					<p>Ninth ordered item keeps the list moving.</p>
					<p>Tenth ordered item exercises double-digit width alignment.</p>
				</ol>
				<div display="always" align="bottom" padding="4pt">
					<label display="odd">Odd list footer</label>
					<label display="even">Even list footer</label>
				</div>
			</page>
		</ltml>`)
	w := &labelTestWriter{t: t, fonts: defaultTestFonts(t)}
	if err := doc.Print(w); err != nil {
		t.Fatal(err)
	}
	if w.pageCount < 2 {
		t.Fatalf("page count = %d, want at least 2", w.pageCount)
	}

	pageTexts := map[int][]string{}
	for i, rt := range w.printed {
		pageTexts[w.printedPages[i]] = append(pageTexts[w.printedPages[i]], rt.String())
	}
	page1Text := strings.Join(pageTexts[1], "\n")
	page2Text := strings.Join(pageTexts[2], "\n")
	if !strings.Contains(page1Text, "List splitting") || !strings.Contains(page2Text, "List splitting") {
		t.Fatalf("expected repeating list header on both pages, got page1=%q page2=%q", page1Text, page2Text)
	}
	if !strings.Contains(page1Text, "Odd list footer") || strings.Contains(page1Text, "Even list footer") {
		t.Fatalf("expected only odd footer on page 1, got %q", page1Text)
	}
	if !strings.Contains(page2Text, "Even list footer") || strings.Contains(page2Text, "Odd list footer") {
		t.Fatalf("expected only even footer on page 2, got %q", page2Text)
	}

	pagePlain := map[int][]string{}
	for i, text := range w.plainPrinted {
		pagePlain[w.plainPages[i]] = append(pagePlain[w.plainPages[i]], text)
	}
	if !slices.Contains(pagePlain[1], "1.") {
		t.Fatalf("expected ordered marker 1. on page 1, got %v", pagePlain[1])
	}
	foundSecond := false
	foundTenth := false
	for pageNo := 2; pageNo <= w.pageCount; pageNo++ {
		if slices.Contains(pagePlain[pageNo], "1.") {
			t.Fatalf("did not expect ordered marker 1. after page 1, got page %d markers %v", pageNo, pagePlain[pageNo])
		}
		if slices.Contains(pagePlain[pageNo], "2.") {
			foundSecond = true
		}
		if slices.Contains(pagePlain[pageNo], "10.") {
			foundTenth = true
		}
	}
	if !foundSecond {
		t.Fatalf("expected ordered marker 2. after the first page, got %v", pagePlain)
	}
	if !foundTenth {
		t.Fatalf("expected double-digit ordered marker 10. on a later page, got %v", pagePlain)
	}
}

func TestSample_ListSplit_FirstFragmentBottomTracksRepeatedFooter(t *testing.T) {
	doc := parseDoc(t, `
		<ltml>
			<pagestyle id="tiny" width="200pt" height="120pt" />
			<page style="tiny" margin="0" font.name="Helvetica" font.size="12" layout="vbox">
				<div display="always" align="top" padding="4pt">
					<label display="always" font.weight="Bold">List splitting</label>
					<p>Ordered lists should continue numbering across pages.</p>
				</div>
				<ol padding="4pt" border="thin">
					<p>
						This first ordered item is intentionally long so the list container must split it
						across pages without reprinting the marker on the continuation fragment. Lorem ipsum
						dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut
						labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation
						ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in
						reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur.
					</p>
					<p>Second ordered item continues the numbering after the split.</p>
				</ol>
				<div display="always" align="bottom" padding="4pt">
					<label display="odd">Odd list footer</label>
					<label display="even">Even list footer</label>
				</div>
			</page>
		</ltml>`)
	page := doc.Root().Page(0)
	var list *StdContainer
	for _, child := range page.children {
		candidate, ok := child.(*StdContainer)
		if ok && candidate.listKind == listKindOrdered {
			list = candidate
			break
		}
	}
	if list == nil {
		t.Fatal("expected direct child ordered list")
	}

	w := &labelTestWriter{t: t, fonts: defaultTestFonts(t)}
	page.initFlowItems()
	if err := page.preparePhysicalPage(w, true); err != nil {
		t.Fatal(err)
	}
	result, err := list.SplitForHeight(page.availableHeightForChild(list), w)
	if err != nil {
		t.Fatal(err)
	}
	if result == nil || result.Head == nil {
		t.Fatal("expected ordered list to split")
	}

	head := result.Head.(*StdContainer)
	page.copySplitGeometry(head, list)
	head.LayoutWidget(w)

	lastVisibleBottom := 0.0
	for _, child := range head.Widgets() {
		if child.Visible() && !child.Disabled() {
			lastVisibleBottom = max(lastVisibleBottom, child.Bottom())
		}
	}
	if lastVisibleBottom == 0 {
		t.Fatal("expected visible content on first fragment")
	}

	if gap := head.Bottom() - lastVisibleBottom; gap > 8 {
		t.Fatalf("first fragment bottom gap below visible content = %vpt, want <= 8pt", gap)
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
	table.SetWidth(180)
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
	page := doc.Root().Page(0)
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
	page := doc.Root().Page(0)
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
	page := doc.Root().Page(0)
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
	page := doc.Root().Page(0)
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
	page := doc.Root().Page(0)
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
	page := doc.Root().Page(0)
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

func TestSample_TableAutoSplit_RendersMultiplePages(t *testing.T) {
	doc, err := ParseFile(sampleFile("test_061_table_auto_split.ltml"))
	if err != nil {
		t.Fatal(err)
	}
	w := &labelTestWriter{t: t}
	if err := doc.Print(w); err != nil {
		t.Fatal(err)
	}
	if w.pageCount < 2 {
		t.Fatalf("page count = %d, want at least 2 (printed=%q plain=%q)", w.pageCount, joinRichTexts(w.printed), strings.Join(w.plainPrinted, "\n"))
	}
	pageTexts := map[int][]string{}
	for i, rt := range w.printed {
		pageTexts[w.printedPages[i]] = append(pageTexts[w.printedPages[i]], rt.String())
	}
	page1Text := strings.Join(pageTexts[1], "\n")
	page2Text := strings.Join(pageTexts[2], "\n")
	if !strings.Contains(page1Text, "Repeating table header") || !strings.Contains(page2Text, "Repeating table header") {
		t.Fatalf("expected repeating table header on first two pages, got page1=%q page2=%q", page1Text, page2Text)
	}
	allText := strings.Join(w.plainPrinted, "\n") + "\n" + joinRichTexts(w.printed)
	if !strings.Contains(allText, "row 10 A") || !strings.Contains(allText, "row 10 B") {
		t.Fatalf("expected final table row to render, got %q", allText)
	}
}

func joinRichTexts(texts []*rich_text.RichText) string {
	parts := make([]string, 0, len(texts))
	for _, text := range texts {
		parts = append(parts, text.String())
	}
	return strings.Join(parts, "\n")
}
