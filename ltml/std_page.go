// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"errors"
	"fmt"
	"regexp"
	"slices"

	"github.com/rowland/leadtype/colors"
)

type StdPage struct {
	StdContainer
	Scope
	pageStyle      *PageStyle
	marginChanged  bool
	grid           bool
	gridStep       float64
	overflow       bool
	overflowSet    bool
	flowPageIndex  int
	flowItems      []*pageItem
	activeChildren []Widget
}

type pageItem struct {
	Source  Widget
	Current Widget
	Done    bool
}

func (p *StdPage) BeforePrint(w Writer) error {
	p.flowPageIndex = 1
	p.initFlowItems()
	return p.preparePhysicalPage(w, true)
}

func (p *StdPage) Bottom() float64 {
	return p.Height()
}

func (p *StdPage) BottomIsSet() bool {
	return true
}

func (p *StdPage) root() *StdPage {
	if p.container == nil {
		return p
	}
	return p.container.(*StdPage)
}

func (p *StdPage) document() *StdDocument {
	if p.container != nil {
		return p.container.(*StdDocument)
	}
	return nil
}

func (p *StdPage) Height() float64 {
	return p.PageStyle().Height()
}

func (p *StdPage) Left() float64 {
	return 0
}

func (p *StdPage) LeftIsSet() bool {
	return true
}

func (p *StdPage) MarginTop() float64 {
	if p.marginChanged {
		return p.StdContainer.MarginTop()
	}
	if doc := p.document(); doc != nil {
		return doc.MarginTop()
	}
	return 0
}

func (p *StdPage) MarginRight() float64 {
	if p.marginChanged {
		return p.StdContainer.MarginRight()
	}
	if doc := p.document(); doc != nil {
		return doc.MarginRight()
	}
	return 0
}

func (p *StdPage) MarginBottom() float64 {
	if p.marginChanged {
		return p.StdContainer.MarginBottom()
	}
	if doc := p.document(); doc != nil {
		return doc.MarginBottom()
	}
	return 0
}

func (p *StdPage) MarginLeft() float64 {
	if p.marginChanged {
		return p.StdContainer.MarginLeft()
	}
	if doc := p.document(); doc != nil {
		return doc.MarginLeft()
	}
	return 0
}

func (p *StdPage) PageStyle() *PageStyle {
	if p.pageStyle == nil {
		return p.document().PageStyle()
	}
	return p.pageStyle
}

func (p *StdPage) PaintBackground(w Writer) error {
	if err := p.StdContainer.PaintBackground(w); err != nil {
		return err
	}
	if !p.grid {
		return nil
	}
	return p.drawGrid(w)
}

func (p *StdPage) DrawContent(w Writer) error {
	printedOnce, err := p.drawVisibleChildren(w)
	if err != nil {
		return err
	}
	if !p.effectiveOverflow() || !p.supportsOverflowRetry() {
		return nil
	}
	for printedOnce > 0 && p.hasPendingOnceChildren() {
		p.flowPageIndex++
		if err := p.preparePhysicalPage(w, false); err != nil {
			if errors.Is(err, errNoProgressPage) {
				return nil
			}
			return err
		}
		if err := p.PaintBackground(w); err != nil {
			return err
		}
		if err := p.DrawBorder(w); err != nil {
			return err
		}
		printedOnce, err = p.drawVisibleChildren(w)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *StdPage) Right() float64 {
	return p.Width()
}

func (p *StdPage) RightIsSet() bool {
	return true
}

var reMargin = regexp.MustCompile(`^margin(-top|-right|-bottom|-left)?$`)

func (p *StdPage) SetAttrs(attrs map[string]string) {
	p.StdContainer.SetAttrs(attrs)
	if style, ok := attrs["style"]; ok {
		p.pageStyle = PageStyleFor(style, p.scope)
	}
	if grid, ok := attrs["grid"]; ok {
		switch grid {
		case "", "false":
			p.grid = false
		case "true":
			p.grid = true
			p.gridStep = 0.25
		default:
			p.grid = true
			p.gridStep = ParseMeasurement(grid, p.Units())
		}
	}
	if overflow, ok := attrs["overflow"]; ok {
		p.overflowSet = true
		p.overflow = overflow == "true"
	}
	for k, _ := range attrs {
		if reMargin.MatchString(k) {
			p.marginChanged = true
			break
		}
	}
}

var errBadPageContainer = errors.New("page must be child of ltml.")

func (p *StdPage) SetContainer(container Container) error {
	if _, ok := container.(*StdDocument); ok {
		return p.StdContainer.SetContainer(container)
	} else {
		return errBadPageContainer
	}
}

func (p *StdPage) String() string {
	return fmt.Sprintf("StdPage %s", &p.StdContainer)
}

func (p *StdPage) Top() float64 {
	return 0
}

func (p *StdPage) TopIsSet() bool {
	return true
}

func (p *StdPage) Width() float64 {
	return p.PageStyle().Width()
}

func (p *StdPage) Widgets() []Widget {
	if p.activeChildren != nil {
		return p.activeChildren
	}
	return p.children
}

func (p *StdPage) drawGrid(w Writer) error {
	step := p.gridStep
	if step <= 0 {
		step = ParseMeasurement("0.25in", p.Units())
	}
	prevColor := w.SetLineColor(colors.LightGray)
	prevDash := w.SetLineDashPattern("solid")
	prevCap := w.SetLineCapStyle("butt_cap")
	w.SetLineWidth(0.25)
	defer func() {
		w.SetLineColor(prevColor)
		w.SetLineDashPattern(prevDash)
		w.SetLineCapStyle(prevCap)
	}()

	return w.Path(func() {
		for x := 0.0; x <= p.Width(); x += step {
			w.MoveTo(x, 0)
			w.LineTo(x, p.Height())
		}
		for y := 0.0; y <= p.Height(); y += step {
			w.MoveTo(0, y)
			w.LineTo(p.Width(), y)
		}
		_ = w.Stroke()
	})
}

var errNoProgressPage = errors.New("page overflow retry would print no display=once direct children")

func (p *StdPage) drawVisibleChildren(w Writer) (int, error) {
	printedOnce := 0
	children := slices.Clone(p.Widgets())
	slices.SortStableFunc(children, func(a, b Widget) int {
		return a.ZIndex() - b.ZIndex()
	})
	for _, child := range children {
		if !child.Visible() || child.Disabled() {
			if item := p.pageItemForCurrent(child); item != nil && !item.Done {
				progress, err := p.trySplitChild(item, child, w)
				if err != nil {
					return printedOnce, err
				}
				if progress {
					printedOnce++
				}
			}
			continue
		}
		wasPrinted := child.Printed()
		if err := Print(child, w); err != nil {
			return printedOnce, err
		}
		if item := p.pageItemForCurrent(child); item != nil {
			item.Done = true
			item.Current = nil
		}
		if child.Display() == DisplayOnce && !wasPrinted && child.Printed() {
			printedOnce++
		}
	}
	return printedOnce, nil
}

func (p *StdPage) hasPendingOnceChildren() bool {
	if len(p.flowItems) > 0 {
		for _, item := range p.flowItems {
			if !item.Done && item.Current != nil {
				return true
			}
		}
		return false
	}
	for _, child := range p.children {
		if child.Display() == DisplayOnce && !child.Printed() {
			return true
		}
	}
	return false
}

func (p *StdPage) supportsOverflowRetry() bool {
	style := p.layout
	if style == nil {
		if p.scope == nil {
			style = defaultLayouts["vbox"]
		} else {
			style = p.LayoutStyle()
		}
	}
	if style == nil {
		return false
	}
	switch style.manager {
	case "flow", "table", "vbox":
		return true
	default:
		return false
	}
}

func (p *StdPage) effectiveOverflow() bool {
	if p.overflowSet {
		return p.overflow
	}
	return p.supportsOverflowRetry()
}

func (p *StdPage) preparePhysicalPage(w Writer, force bool) error {
	doc := p.document()
	var savedDocPageNo, savedPhysicalPageNo int
	var savedPendingStart *int
	if doc != nil {
		savedDocPageNo = doc.documentPageNo
		savedPhysicalPageNo = doc.physicalPageNo
		savedPendingStart = doc.pendingStart
		if start, ok := p.firstPageNoStartForRender(); ok {
			doc.SetPendingStart(start)
		}
		if doc.pendingStart != nil {
			doc.SetCurrentPageStart(*doc.pendingStart)
			doc.pendingStart = nil
		}
		doc.documentPageNo++
		doc.physicalPageNo++
	}

	if force {
		p.rebuildActiveChildren()
		w.NewPage()
		LayoutContainer(p, w)
	} else {
		probe := newLayoutProbeWriter(w)
		p.rebuildActiveChildren()
		LayoutContainer(p, probe)
		if p.countVisibleOnceChildren() == 0 && !p.hasSplittableOnceProgress(probe) {
			if doc != nil {
				doc.documentPageNo = savedDocPageNo
				doc.physicalPageNo = savedPhysicalPageNo
				doc.pendingStart = savedPendingStart
			}
			return errNoProgressPage
		}
		p.rebuildActiveChildren()
		w.NewPage()
		LayoutContainer(p, w)
	}
	if doc != nil {
		if reset, ok := p.firstPageNoResetForRender(); ok {
			doc.SetPendingStart(reset)
		}
	}
	return nil
}

func (p *StdPage) countVisibleOnceChildren() int {
	count := 0
	for _, child := range p.Widgets() {
		if child.Visible() && !child.Disabled() && child.Display() == DisplayOnce && !child.Printed() {
			count++
		}
	}
	return count
}

func (p *StdPage) firstPageNoResetForRender() (int, bool) {
	var value int
	var found bool
	p.walkDisplayWidgets(p, func(widget Widget) bool {
		if pageNo, ok := widget.(*StdPageNo); ok && pageNo.hasReset() {
			value = pageNo.reset
			found = true
			return false
		}
		return true
	})
	return value, found
}

func (p *StdPage) firstPageNoStartForRender() (int, bool) {
	var value int
	var found bool
	p.walkDisplayWidgets(p, func(widget Widget) bool {
		if pageNo, ok := widget.(*StdPageNo); ok && pageNo.hasStart() {
			value = pageNo.start
			found = true
			return false
		}
		return true
	})
	return value, found
}

func (p *StdPage) walkDisplayWidgets(root Container, fn func(Widget) bool) bool {
	if !fn(root) {
		return false
	}
	physicalPageNo := 0
	if doc := p.document(); doc != nil {
		physicalPageNo = doc.CurrentPhysicalPageNo()
	}
	for _, child := range root.Widgets() {
		parentRepeats := root == p || root.Display() != DisplayOnce
		if !widgetDisplayForRender(child, parentRepeats, p.flowPageIndex, physicalPageNo) {
			continue
		}
		if container, ok := child.(Container); ok {
			if !p.walkDisplayWidgets(container, fn) {
				return false
			}
			continue
		}
		if !fn(child) {
			return false
		}
	}
	return true
}

func (p *StdPage) initFlowItems() {
	p.flowItems = nil
	p.activeChildren = nil
	if !p.effectiveOverflow() {
		return
	}
	for _, child := range p.children {
		if child.Display() == DisplayOnce {
			p.flowItems = append(p.flowItems, &pageItem{Source: child, Current: child})
		}
	}
}

func (p *StdPage) rebuildActiveChildren() {
	if len(p.flowItems) == 0 {
		p.activeChildren = nil
		return
	}
	items := make(map[Widget]*pageItem, len(p.flowItems))
	for _, item := range p.flowItems {
		items[item.Source] = item
	}
	active := make([]Widget, 0, len(p.children))
	for _, child := range p.children {
		if child.Display() != DisplayOnce {
			active = append(active, child)
			continue
		}
		item := items[child]
		if item == nil || item.Done || item.Current == nil {
			continue
		}
		p.resetWidgetRenderState(item.Current)
		if wc, ok := item.Current.(WantsContainer); ok {
			_ = wc.SetContainer(p)
		}
		active = append(active, item.Current)
	}
	p.activeChildren = active
}

func (p *StdPage) resetWidgetRenderState(widget Widget) {
	widget.SetPrinted(false)
	widget.SetVisible(true)
	widget.SetDisabled(false)
	container, ok := widget.(Container)
	if !ok {
		return
	}
	for _, child := range container.Widgets() {
		p.resetWidgetRenderState(child)
	}
}

func (p *StdPage) pageItemForCurrent(widget Widget) *pageItem {
	for _, item := range p.flowItems {
		if item.Current == widget {
			return item
		}
	}
	return nil
}

func (p *StdPage) availableHeightForChild(child Widget) float64 {
	limit := ContentBottom(p)
	for _, sibling := range p.Widgets() {
		if sibling == child || !sibling.Visible() || sibling.Disabled() {
			continue
		}
		if sibling.Align() == AlignBottom {
			limit = min(limit, sibling.Top())
		}
	}
	avail := limit - child.Top()
	if avail < 0 {
		return 0
	}
	return avail
}

func (p *StdPage) hasSplittableOnceProgress(w Writer) bool {
	for _, child := range p.Widgets() {
		if child.Display() != DisplayOnce || child.Visible() || child.Disabled() {
			continue
		}
		item := p.pageItemForCurrent(child)
		if item == nil || item.Done {
			continue
		}
		splittable, ok := child.(Splittable)
		if !ok {
			continue
		}
		result, err := splittable.SplitForHeight(p.availableHeightForChild(child), w)
		if err == nil && result != nil && result.Head != nil {
			return true
		}
	}
	return false
}

func (p *StdPage) trySplitChild(item *pageItem, child Widget, w Writer) (bool, error) {
	splittable, ok := child.(Splittable)
	if !ok {
		return false, nil
	}
	result, err := splittable.SplitForHeight(p.availableHeightForChild(child), w)
	if err != nil || result == nil || result.Head == nil {
		return false, err
	}
	if wc, ok := result.Head.(WantsContainer); ok {
		_ = wc.SetContainer(p)
	}
	p.copySplitGeometry(result.Head, child)
	result.Head.LayoutWidget(w)
	if result.Tail != nil {
		if wc, ok := result.Tail.(WantsContainer); ok {
			_ = wc.SetContainer(p)
		}
		item.Current = result.Tail
	} else {
		item.Current = nil
		item.Done = true
	}
	if err := Print(result.Head, w); err != nil {
		return false, err
	}
	return true, nil
}

func (p *StdPage) copySplitGeometry(dst, src Widget) {
	dst.SetLeft(src.Left())
	dst.SetTop(src.Top())
	dst.SetWidth(src.Width())
	if src.HeightIsSet() {
		dst.SetHeight(src.Height())
	}
	dst.SetPosition(src.Position())
	dst.SetVisible(true)
	dst.SetDisabled(false)
}

func init() {
	registerTag(DefaultSpace, "page", func() any { return &StdPage{} })
}

var _ Container = (*StdPage)(nil)
var _ HasAttrs = (*StdPage)(nil)
var _ Identifier = (*StdPage)(nil)
var _ Printer = (*StdPage)(nil)
var _ WantsContainer = (*StdPage)(nil)
