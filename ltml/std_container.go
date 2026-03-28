// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strconv"
)

type StdContainer struct {
	StdWidget
	Children
	cols            int
	layout          *LayoutStyle
	order           TableOrder
	paragraphStyle  *ParagraphStyle
	preferredHeight float64
	preferredWidth  float64
	rows            int
	activeChildren  []Widget
	splitEnabled    bool
	splitExplicit   bool
	headerRows      int
	footerRows      int
}

func (c *StdContainer) Cols() int {
	return c.cols
}

func (c *StdContainer) Container() Container {
	return c.container
}

func (c *StdContainer) DrawContent(w Writer) error {
	// fmt.Printf("DrawContent %s\n", c)
	children := slices.Clone(c.Widgets())
	slices.SortStableFunc(children, func(a, b Widget) int {
		return a.ZIndex() - b.ZIndex()
	})
	for _, child := range children {
		if !child.Visible() || child.Disabled() {
			continue
		}
		if err := Print(child, w); err != nil {
			return err
		}
	}
	return nil
}

func (c *StdContainer) LayoutStyle() *LayoutStyle {
	if c.layout == nil {
		return LayoutStyleFor("vbox", c.scope)
	}
	return c.layout
}

func (c *StdContainer) LayoutWidget(w Writer) {
	LayoutContainer(c, w)
}

func (c *StdContainer) Widgets() []Widget {
	if c.activeChildren != nil {
		return c.activeChildren
	}
	return c.children
}

func (c *StdContainer) PreferredHeight(w Writer) float64 {
	if c.height != 0 {
		return c.height
	}
	savedHeight, savedHeightPct, savedHeightRel, savedHeightSet :=
		c.height, c.heightPct, c.heightRel, c.heightSet
	LayoutContainer(c, newLayoutProbeWriter(w))
	height := c.Height()
	c.height, c.heightPct, c.heightRel, c.heightSet =
		savedHeight, savedHeightPct, savedHeightRel, savedHeightSet
	return height
}

func (c *StdContainer) Order() TableOrder {
	return c.order
}

func (c *StdContainer) ParagraphStyle() *ParagraphStyle {
	if c.paragraphStyle == nil {
		return c.container.ParagraphStyle()
	}
	return c.paragraphStyle
}

func (c *StdContainer) Rows() int {
	return c.rows
}

func (c *StdContainer) SetAttrs(attrs map[string]string) {
	c.StdWidget.SetAttrs(attrs)
	if layout, ok := attrs["layout"]; ok {
		c.layout = LayoutStyleFor(layout, c.scope)
	}
	if order, ok := attrs["order"]; ok {
		if order == "rows" {
			c.order = TableOrderRows
		} else if order == "cols" {
			c.order = TableOrderCols
		}
	}
	if rows, ok := attrs["rows"]; ok {
		if value, err := strconv.Atoi(rows); err == nil {
			c.rows = value
		}
	}
	if cols, ok := attrs["cols"]; ok {
		if value, err := strconv.Atoi(cols); err == nil {
			c.cols = value
		}
	}
	if split, ok := attrs["split"]; ok {
		c.splitExplicit = true
		c.splitEnabled = split != "false"
	}
	if headerRows, ok := attrs["header-rows"]; ok {
		if value, err := strconv.Atoi(headerRows); err == nil {
			c.headerRows = value
		}
	}
	if footerRows, ok := attrs["footer-rows"]; ok {
		if value, err := strconv.Atoi(footerRows); err == nil {
			c.footerRows = value
		}
	}
	if ps, ok := attrs["paragraph-style"]; ok {
		c.paragraphStyle = ParagraphStyleFor(ps, c.scope)
	}
	if MapHasKeyPrefix(attrs, "paragraph-style.") {
		c.paragraphStyle = c.ParagraphStyle().Clone()
		c.paragraphStyle.SetAttrs("paragraph-style.", attrs)
	}
}

func (c *StdContainer) String() string {
	return fmt.Sprintf("StdContainer layout=%v paragraphStyle=%v %s", c.layout, c.paragraphStyle, &c.StdWidget)
}

var errTableSplitUnsupportedRowSpan = errors.New("table splitting does not support rowspan > 1")

func (c *StdContainer) SplitForHeight(avail float64, w Writer) (*SplitResult, error) {
	if c.LayoutStyle() == nil || c.LayoutStyle().manager != "table" || !c.tableSplitEnabled() {
		return nil, nil
	}
	metrics, err := c.tableSplitMetrics(w)
	if err != nil {
		return nil, err
	}
	bodyCount := metrics.bodyEnd - metrics.bodyStart
	if bodyCount < 2 {
		return nil, nil
	}
	fitBodies := 0
	for n := 1; n < bodyCount; n++ {
		if c.tableFragmentHeight(metrics, metrics.bodyStart, metrics.bodyStart+n) <= avail {
			fitBodies = n
			continue
		}
		break
	}
	if fitBodies == 0 {
		return nil, nil
	}
	headRows := append([]int{}, metrics.headerRows...)
	for r := metrics.bodyStart; r < metrics.bodyStart+fitBodies; r++ {
		headRows = append(headRows, r)
	}
	headRows = append(headRows, metrics.footerRows...)

	tailRows := append([]int{}, metrics.headerRows...)
	for r := metrics.bodyStart + fitBodies; r < metrics.bodyEnd; r++ {
		tailRows = append(tailRows, r)
	}
	tailRows = append(tailRows, metrics.footerRows...)

	head := c.cloneTableFragment(metrics, headRows)
	tail := c.cloneTableFragment(metrics, tailRows)
	return &SplitResult{Head: head, Tail: tail}, nil
}

func (c *StdContainer) tableSplitEnabled() bool {
	if c.splitExplicit {
		return c.splitEnabled
	}
	return c.LayoutStyle() != nil && c.LayoutStyle().manager == "table"
}

type tableSplitMetrics struct {
	grid       *WidgetGrid
	rowHeights []float64
	headerRows []int
	footerRows []int
	bodyStart  int
	bodyEnd    int
}

func (c *StdContainer) tableSplitMetrics(w Writer) (*tableSplitMetrics, error) {
	var grid *WidgetGrid
	var err error
	if c.Order() == TableOrderRows {
		grid, err = rowGrid(c)
	} else {
		grid, err = colGrid(c)
	}
	if err != nil {
		return nil, err
	}
	for _, widget := range c.Widgets() {
		if widget.RowSpan() > 1 {
			return nil, errTableSplitUnsupportedRowSpan
		}
	}
	widths := detectWidths(grid, w)
	percents, others := widths.Partition(func(w *SpecifiedSize) bool { return w.How == Percent })
	specified, others := others.Partition(func(w *SpecifiedSize) bool { return w.How == Specified })
	widthAvail := ContentWidth(c)
	widthAvail = allocateSpecifiedWidths(widthAvail, specified, c.LayoutStyle())
	widthAvail = allocatePercentWidths(widthAvail, percents, c.LayoutStyle())
	_ = allocateOtherWidths(widthAvail, others, c.LayoutStyle())

	rowHeights := make([]float64, grid.Rows())
	for r := 0; r < grid.Rows(); r++ {
		maxHeight := 0.0
		for col := 0; col < grid.Cols(); col++ {
			widget := grid.Cell(col, r)
			if widget == nil {
				continue
			}
			if widths[col].Size <= 0 {
				continue
			}
			width := 0.0
			for i := 0; i < widget.ColSpan(); i++ {
				width += widths[col+i].Size
			}
			widget.SetWidth(width + float64(widget.ColSpan()-1)*c.LayoutStyle().HPadding())
			height := widget.Height()
			if !widget.HeightIsSet() {
				height = widget.PreferredHeight(w)
			}
			if height > maxHeight {
				maxHeight = height
			}
		}
		rowHeights[r] = maxHeight
	}
	headerCount := minInt(c.headerRows, grid.Rows())
	footerCount := minInt(c.footerRows, max(0, grid.Rows()-headerCount))
	headerRows := make([]int, 0, headerCount)
	for i := 0; i < headerCount; i++ {
		headerRows = append(headerRows, i)
	}
	footerRows := make([]int, 0, footerCount)
	for i := grid.Rows() - footerCount; i < grid.Rows(); i++ {
		if i >= headerCount {
			footerRows = append(footerRows, i)
		}
	}
	return &tableSplitMetrics{
		grid:       grid,
		rowHeights: rowHeights,
		headerRows: headerRows,
		footerRows: footerRows,
		bodyStart:  headerCount,
		bodyEnd:    grid.Rows() - footerCount,
	}, nil
}

func (c *StdContainer) tableFragmentHeight(metrics *tableSplitMetrics, bodyStart, bodyEnd int) float64 {
	rows := make([]int, 0, len(metrics.headerRows)+(bodyEnd-bodyStart)+len(metrics.footerRows))
	rows = append(rows, metrics.headerRows...)
	for r := bodyStart; r < bodyEnd; r++ {
		rows = append(rows, r)
	}
	rows = append(rows, metrics.footerRows...)
	height := NonContentHeight(c)
	for i, row := range rows {
		height += metrics.rowHeights[row]
		if i > 0 {
			height += c.LayoutStyle().VPadding()
		}
	}
	return height
}

func (c *StdContainer) cloneTableFragment(metrics *tableSplitMetrics, rows []int) *StdContainer {
	clone := *c
	clone.activeChildren = c.cloneTableWidgetsForRows(metrics.grid, rows, &clone)
	clone.height = 0
	clone.heightPct = 0
	clone.heightRel = 0
	clone.heightSet = false
	clone.printed = false
	clone.invisible = false
	clone.disabled = false
	clone.path = ""
	for _, child := range clone.activeChildren {
		child.SetPrinted(false)
		child.SetVisible(true)
		child.SetDisabled(false)
	}
	return &clone
}

func (c *StdContainer) cloneTableWidgetsForRows(grid *WidgetGrid, rows []int, parent Container) []Widget {
	var widgets []Widget
	seen := map[Widget]bool{}
	for _, r := range rows {
		for col := 0; col < grid.Cols(); col++ {
			widget := grid.Cell(col, r)
			if widget == nil || seen[widget] {
				continue
			}
			seen[widget] = true
			clone := cloneWidgetShallow(widget)
			if wc, ok := clone.(WantsContainer); ok {
				_ = wc.SetContainer(parent)
			}
			widgets = append(widgets, clone)
		}
	}
	return widgets
}

func cloneWidgetShallow(widget Widget) Widget {
	value := reflect.ValueOf(widget)
	if value.Kind() != reflect.Pointer || value.IsNil() {
		panic("cloneWidgetShallow expects non-nil pointer widget")
	}
	clone := reflect.New(value.Elem().Type())
	clone.Elem().Set(value.Elem())
	w, ok := clone.Interface().(Widget)
	if !ok {
		panic("cloneWidgetShallow produced non-widget clone")
	}
	w.SetPrinted(false)
	w.SetVisible(true)
	w.SetDisabled(false)
	return w
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func init() {
	registerTag(DefaultSpace, "div", func() any { return &StdContainer{} })
}

var _ Container = (*StdContainer)(nil)
var _ HasAttrs = (*StdContainer)(nil)
var _ Identifier = (*StdContainer)(nil)
var _ Printer = (*StdContainer)(nil)
var _ WantsContainer = (*StdContainer)(nil)
