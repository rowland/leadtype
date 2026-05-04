package ltml

import (
	"errors"
	"math"
)

func markGrid(grid *BoolGrid, a, b, c, d int, value bool) {
	for aa := 0; aa < c; aa++ {
		for bb := 0; bb < d; bb++ {
			if aa > 0 || bb > 0 {
				grid.SetCell(a+aa, b+bb, value)
			}
		}
	}
}

func rowGrid(container Container) (*WidgetGrid, error) {
	if container.Cols() < 1 {
		return nil, errors.New("cols must be specified")
	}
	static, _ := printableWidgets(container, Static)
	used := NewBoolGrid(container.Cols(), 0)
	grid := NewWidgetGrid(container.Cols(), 0)
	row, col := 0, 0
	for _, widget := range static {
		for used.Cell(col, row) {
			col += 1
			if col >= container.Cols() {
				row += 1
				col = 0
			}
		}
		grid.SetCell(col, row, widget)
		markGrid(used, col, row, widget.ColSpan(), widget.RowSpan(), true)
		col += widget.ColSpan()
		if col > container.Cols() {
			return nil, errors.New("colspan causes number of columns to exceed table size")
		}
		if col == container.Cols() {
			row += 1
			col = 0
		}
	}
	return grid, nil
}

func colGrid(container Container) (*WidgetGrid, error) {
	if container.Rows() < 1 {
		return nil, errors.New("rows must be specified")
	}
	static, _ := printableWidgets(container, Static)
	used := NewBoolGrid(container.Cols(), 0)
	grid := NewWidgetGrid(0, container.Rows())
	row, col := 0, 0
	for _, widget := range static {
		for used.Cell(col, row) {
			row += 1
			if row >= container.Rows() {
				col += 1
				row = 0
			}
		}
		if row >= container.Rows() {
			col += 1
			row = 0
		}
		grid.SetCell(col, row, widget)
		markGrid(used, col, row, widget.ColSpan(), widget.RowSpan(), true)
		row += widget.RowSpan()
		if row > container.Rows() {
			return nil, errors.New("rowspan causes number of rows to exceed table size")
		}
	}
	return grid, nil
}

func detectWidths(grid *WidgetGrid, writer Writer) SpecifiedSizes {
	widths := make(SpecifiedSizes, grid.Cols())
	for c := 0; c < grid.Cols(); c++ {
		var widget Widget
		for r := 0; r < grid.Rows(); r++ {
			if w := grid.Cell(c, r); w != nil && w.ColSpan() == 1 {
				widget = w
				break
			}
		}
		if widget == nil {
			widths[c] = &SpecifiedSize{How: Unspecified, Size: 0}
		} else if widget.WidthPctIsSet() {
			widths[c] = &SpecifiedSize{How: Percent, Size: widget.Width()}
		} else if widgetWidthSpecified(widget) {
			widths[c] = &SpecifiedSize{How: Specified, Size: widget.Width()}
		} else {
			max := 0.0
			for r := 0; r < grid.Rows(); r++ {
				if w := grid.Cell(c, r); w != nil {
					pw := w.PreferredWidth(writer)
					if pw > max {
						max = pw
					}
				}
			}
			widths[c] = &SpecifiedSize{How: Unspecified, Size: max}
		}
	}
	return widths
}

func allocateSpecifiedWidths(widthAvail float64, specified SpecifiedSizes, style *LayoutStyle) float64 {
	for _, w := range specified {
		if widthAvail >= w.Size {
			widthAvail -= (w.Size + style.HPadding())
		}
	}
	return widthAvail
}

func allocatePercentWidths(widthAvail float64, percents SpecifiedSizes, style *LayoutStyle) float64 {
	if widthAvail-(float64(len(percents)-1))*style.HPadding() >= float64(len(percents)) {
		widthAvail -= float64((len(percents) - 1)) * style.HPadding()
		totalPercents := 0.0
		for i := range percents {
			totalPercents += percents[i].Size
		}
		ratio := widthAvail / totalPercents
		for i := range percents {
			if ratio < 1.0 {
				percents[i].Size *= ratio
			}
			widthAvail -= percents[i].Size
		}
	} else {
		for i := range percents {
			percents[i].Size = 0
		}
	}
	widthAvail -= style.HPadding()
	return widthAvail
}

func allocateOtherWidths(widthAvail float64, others SpecifiedSizes, style *LayoutStyle) float64 {
	if widthAvail-(float64(len(others)-1))*style.HPadding() >= float64(len(others)) {
		widthAvail -= float64(len(others)-1) * style.HPadding()
		othersWidth := widthAvail / float64(len(others))
		for i := range others {
			others[i].Size = othersWidth
		}
	} else {
		for i := range others {
			others[i].Size = 0
		}
	}
	return widthAvail
}

func LayoutTable(container Container, style *LayoutStyle, writer Writer) {
	var grid *WidgetGrid
	var err error

	if container.Order() == TableOrderRows {
		grid, err = rowGrid(container)
	} else if container.Order() == TableOrderCols {
		grid, err = colGrid(container)
	} else {
		panic("invalid order")
	}
	if err != nil {
		panic(err)
	}

	containerFull := false
	widths := detectWidths(grid, writer)
	if container.Width() <= 0 {
		panic("container width not set")
	}
	percents, others := widths.Partition(func(w *SpecifiedSize) bool { return w.How == Percent })
	specified, others := others.Partition(func(w *SpecifiedSize) bool { return w.How == Specified })

	widthAvail := ContentWidth(container)
	widthAvail = allocateSpecifiedWidths(widthAvail, specified, style)
	widthAvail = allocatePercentWidths(widthAvail, percents, style)
	widthAvail = allocateOtherWidths(widthAvail, others, style)

	heights := NewSpanSizeGrid(grid.Cols(), grid.Rows())
	for c := 0; c < grid.Cols(); c++ {
		for r := 0; r < grid.Rows(); r++ {
			widget := grid.Cell(c, r)
			if widget == nil {
				continue
			}
			if widths[c].Size > 0 {
				width := 0.0
				for i := 0; i < widget.ColSpan(); i++ {
					width += widths[c+i].Size
				}
				widget.ResolveWidth(width + float64(widget.ColSpan()-1)*style.HPadding())
				var height float64
				if widget.HeightIsSet() {
					height = widget.Height()
				} else {
					height = widget.PreferredHeight(writer)
				}
				heights.SetCell(c, r, SpanSize{Span: widget.RowSpan(), Size: height})
			} else {
				widget.SetDisabled(true)
			}
		}
	}

	for r := 0; r < heights.Rows(); r++ {
		minRowSpan := math.MaxInt64
		for c := 0; c < heights.Cols(); c++ {
			if ss := heights.Cell(c, r); ss.Span > 0 && ss.Span < minRowSpan {
				minRowSpan = ss.Span
			}
		}
		maxHeight := 0.0
		for c := 0; c < heights.Cols(); c++ {
			if ss := heights.Cell(c, r); ss.Span == minRowSpan && ss.Size > maxHeight {
				maxHeight = ss.Size
			}
		}
		for c := 0; c < heights.Cols(); c++ {
			ss := heights.Cell(c, r)
			if ss.Span > minRowSpan {
				heights.SetCell(c, r+1, SpanSize{Span: ss.Span - 1, Size: max(ss.Size-maxHeight, 0)})
			}
			ss.Size = maxHeight
			heights.SetCell(c, r, ss)
		}
	}

	top := ContentTop(container)
	bottom := top + MaxContentHeight(container)
	externalSplit := false
	if table, ok := container.(*StdContainer); ok && table.tableSplitEnabled() {
		if _, ok := table.Container().(*StdPage); ok {
			externalSplit = true
		}
	}
	rtl := IsRTL(container)
	for r := 0; r < grid.Rows(); r++ {
		maxHeight := 0.0
		left := ContentLeft(container)
		right := ContentRight(container)
		for c := 0; c < grid.Cols(); c++ {
			if widget := grid.Cell(c, r); widget != nil {
				widget.SetVisible(!containerFull)
				if containerFull {
					continue
				}
				ss := heights.Cell(c, r)
				widget.SetTop(top)
				if rtl {
					colWidth := widths[c].Size
					for i := 1; i < widget.ColSpan(); i++ {
						colWidth += widths[c+i].Size + style.HPadding()
					}
					widget.SetLeft(right - colWidth)
				} else {
					widget.SetLeft(left)
				}
				height := float64(ss.Span-1) * style.VPadding()
				for rowOffset := 0; rowOffset < ss.Span; rowOffset++ {
					height += heights.Cell(c, r+rowOffset).Size
				}
				widget.ResolveHeight(height)
				if ss.Span == 1 && ss.Size > maxHeight {
					maxHeight = ss.Size
				}
			}
			left += widths[c].Size + style.HPadding()
			right -= widths[c].Size + style.HPadding()
		}
		if containerFull {
			continue
		}
		if !externalSplit && top+maxHeight > bottom {
			containerFull = true
			for c := 0; c < grid.Cols(); c++ {
				if widget := grid.Cell(c, r); widget != nil {
					widget.SetVisible(false)
				}
			}
		}
		if externalSplit || !containerFull {
			top += maxHeight + style.VPadding()
		}
	}
	if !container.HeightIsSet() {
		container.ResolveHeight(top - ContentTop(container) + NonContentHeight(container) - style.VPadding())
	}
	static, remaining := printableWidgets(container, Static)
	for _, widget := range remaining {
		widget.SetVisible(false)
	}
	for _, widget := range static {
		widget.LayoutWidget(writer)
	}
	layoutPositionedChildren(container, writer)
}
