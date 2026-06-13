package ltml

import (
	"errors"
	"math"
)

type tableTrackKind int8

const (
	tableTrackOmitted tableTrackKind = iota
	tableTrackSpecified
	tableTrackPercent
	tableTrackAuto
)

type tableTrackSize struct {
	kind      tableTrackKind
	size      float64
	preferred float64
}

type tableTrackPlan []tableTrackSize

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

func detectTableColumnTracks(grid *WidgetGrid, writer Writer) tableTrackPlan {
	tracks := make(tableTrackPlan, grid.Cols())
	for c := 0; c < grid.Cols(); c++ {
		var specifiedWidget Widget
		auto := false
		preferred := 0.0
		for r := 0; r < grid.Rows(); r++ {
			if w := grid.Cell(c, r); w != nil && w.ColSpan() == 1 {
				if specifiedWidget == nil && (w.WidthPctIsSet() || widgetWidthSpecified(w)) {
					specifiedWidget = w
				}
				if !widgetWidthSpecified(w) && !w.WidthPctIsSet() {
					preferred = max(preferred, w.PreferredWidth(writer))
				}
				if widgetAutoWidth(w) {
					auto = true
				}
			}
		}
		if specifiedWidget != nil && specifiedWidget.WidthPctIsSet() {
			tracks[c] = tableTrackSize{kind: tableTrackPercent, size: specifiedWidget.Width()}
		} else if specifiedWidget != nil {
			tracks[c] = tableTrackSize{kind: tableTrackSpecified, size: specifiedWidget.Width()}
		} else if auto {
			tracks[c] = tableTrackSize{kind: tableTrackAuto, preferred: preferred}
		} else {
			tracks[c] = tableTrackSize{kind: tableTrackOmitted, preferred: preferred}
		}
	}
	return tracks
}

func (tracks tableTrackPlan) resolvedSizes() []float64 {
	sizes := make([]float64, len(tracks))
	for i, track := range tracks {
		sizes[i] = track.size
	}
	return sizes
}

func allocateTableColumnTracks(widthAvail float64, tracks tableTrackPlan, style *LayoutStyle) {
	for i := range tracks {
		if tracks[i].kind == tableTrackSpecified && widthAvail >= tracks[i].size {
			widthAvail -= tracks[i].size + style.HPadding()
		}
	}

	var percentIndexes []int
	for i, track := range tracks {
		if track.kind == tableTrackPercent {
			percentIndexes = append(percentIndexes, i)
		}
	}
	if len(percentIndexes) > 0 && widthAvail-(float64(len(percentIndexes)-1))*style.HPadding() >= float64(len(percentIndexes)) {
		widthAvail -= float64(len(percentIndexes)-1) * style.HPadding()
		totalPercents := 0.0
		for _, i := range percentIndexes {
			totalPercents += tracks[i].size
		}
		ratio := widthAvail / totalPercents
		for _, i := range percentIndexes {
			if ratio < 1.0 {
				tracks[i].size *= ratio
			}
			widthAvail -= tracks[i].size
		}
		widthAvail -= style.HPadding()
	} else if len(percentIndexes) > 0 {
		for _, i := range percentIndexes {
			tracks[i].size = 0
		}
		widthAvail -= style.HPadding()
	}

	var omittedIndexes, autoIndexes []int
	for i, track := range tracks {
		switch track.kind {
		case tableTrackOmitted:
			omittedIndexes = append(omittedIndexes, i)
		case tableTrackAuto:
			autoIndexes = append(autoIndexes, i)
		}
	}
	otherCount := len(omittedIndexes) + len(autoIndexes)
	if otherCount == 0 {
		return
	}
	paddingCost := float64(otherCount-1) * style.HPadding()
	if len(autoIndexes) > 0 {
		preferredTotal := 0.0
		omittedPreferredTotal := 0.0
		for _, i := range omittedIndexes {
			preferredTotal += tracks[i].preferred
			omittedPreferredTotal += tracks[i].preferred
		}
		for _, i := range autoIndexes {
			preferredTotal += tracks[i].preferred
		}
		if widthAvail > preferredTotal+paddingCost {
			widthAvail -= paddingCost
			for _, i := range omittedIndexes {
				tracks[i].size = tracks[i].preferred
				widthAvail -= tracks[i].size
			}
			autoWidth := widthAvail / float64(len(autoIndexes))
			for _, i := range autoIndexes {
				tracks[i].size = autoWidth
			}
			return
		}
		if widthAvail-paddingCost-omittedPreferredTotal >= float64(len(autoIndexes)) {
			widthAvail -= paddingCost
			for _, i := range omittedIndexes {
				tracks[i].size = tracks[i].preferred
				widthAvail -= tracks[i].size
			}
			autoWidth := widthAvail / float64(len(autoIndexes))
			for _, i := range autoIndexes {
				tracks[i].size = autoWidth
			}
			return
		}
	}
	if widthAvail-paddingCost >= float64(otherCount) {
		widthAvail -= paddingCost
		otherWidth := widthAvail / float64(otherCount)
		for _, i := range omittedIndexes {
			tracks[i].size = otherWidth
		}
		for _, i := range autoIndexes {
			tracks[i].size = otherWidth
		}
	} else {
		for _, i := range omittedIndexes {
			tracks[i].size = 0
		}
		for _, i := range autoIndexes {
			tracks[i].size = 0
		}
	}
}

func planTableColumnWidths(grid *WidgetGrid, container Container, style *LayoutStyle, writer Writer) tableTrackPlan {
	tracks := detectTableColumnTracks(grid, writer)
	allocateTableColumnTracks(ContentWidth(container), tracks, style)
	return tracks
}

func tableCellWidth(widths []float64, startCol, colSpan int, hpadding float64) float64 {
	width := 0.0
	for i := 0; i < colSpan; i++ {
		width += widths[startCol+i]
	}
	return width + float64(colSpan-1)*hpadding
}

func tableBaseHeights(grid *WidgetGrid, widths []float64, style *LayoutStyle, writer Writer) (*SpanSizeGrid, []bool) {
	heights := NewSpanSizeGrid(grid.Cols(), grid.Rows())
	autoRows := make([]bool, grid.Rows())
	for c := 0; c < grid.Cols(); c++ {
		for r := 0; r < grid.Rows(); r++ {
			widget := grid.Cell(c, r)
			if widget == nil {
				continue
			}
			if widths[c] <= 0 {
				widget.SetDisabled(true)
				continue
			}
			widget.ResolveWidth(tableCellWidth(widths, c, widget.ColSpan(), style.HPadding()))
			var height float64
			if widget.HeightIsSet() {
				height = widget.Height()
			} else {
				height = widget.PreferredHeight(writer)
				if widgetAutoHeight(widget) && widget.RowSpan() == 1 {
					autoRows[r] = true
				}
			}
			heights.SetCell(c, r, SpanSize{Span: widget.RowSpan(), Size: height})
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
	return heights, autoRows
}

func applyTableAutoRowHeights(container Container, style *LayoutStyle, heights *SpanSizeGrid, autoRows []bool) {
	if !container.HeightIsSet() {
		return
	}
	autoCount := 0
	baselineHeight := 0.0
	for r := 0; r < heights.Rows(); r++ {
		if r > 0 {
			baselineHeight += style.VPadding()
		}
		baselineHeight += heights.Cell(0, r).Size
		if autoRows[r] {
			autoCount++
		}
	}
	if autoCount == 0 {
		return
	}
	surplus := ContentHeight(container) - baselineHeight
	if surplus <= 0 {
		return
	}
	extra := surplus / float64(autoCount)
	for r := 0; r < heights.Rows(); r++ {
		if !autoRows[r] {
			continue
		}
		for c := 0; c < heights.Cols(); c++ {
			ss := heights.Cell(c, r)
			ss.Size += extra
			heights.SetCell(c, r, ss)
		}
	}
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
	if container.Width() <= 0 {
		panic("container width not set")
	}
	widths := planTableColumnWidths(grid, container, style, writer).resolvedSizes()
	heights, autoRows := tableBaseHeights(grid, widths, style, writer)
	applyTableAutoRowHeights(container, style, heights, autoRows)

	top := ContentTop(container)
	bottom := top + MaxContentHeight(container)
	externalSplit := false
	if table, ok := container.(*StdContainer); ok && table.SplitEnabled() {
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
					widget.SetLeft(right - tableCellWidth(widths, c, widget.ColSpan(), style.HPadding()))
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
			left += widths[c] + style.HPadding()
			right -= widths[c] + style.HPadding()
		}
		if containerFull {
			continue
		}
		if !externalSplit && top+maxHeight > bottom+layoutFitEpsilon {
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
