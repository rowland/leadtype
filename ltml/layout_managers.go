package ltml

import (
	"errors"
	"fmt"
	"math"
)

type LayoutFunc func(container Container, style *LayoutStyle, writer Writer)

var layoutManagers = make(map[string]LayoutFunc)

func RegisterLayoutManager(name string, f LayoutFunc) {
	layoutManagers[name] = f
}

func LayoutManagerFor(name string) LayoutFunc {
	// fmt.Println("In LayoutManagerFor", name)
	if f, ok := layoutManagers[name]; ok {
		// fmt.Printf("%#v", f)
		return f
	}
	fmt.Println("couldn't find ", name)
	return LayoutVBox
}

type Position int

const (
	Static = Position(iota)
	Relative
	Absolute
)

func LayoutAbsolute(container Container, style *LayoutStyle, writer Writer) {

}

func LayoutFlow(container Container, style *LayoutStyle, writer Writer) {
	var cx, cy, maxY float64
	containerFull := false
	bottom := ContentTop(container) + MaxContentHeight(container)
	widgets, remaining := printableWidgets(container, Static)
	for _, widget := range remaining {
		if widget.Printed() {
			widget.SetVisible(false)
		}
	}
	for _, widget := range widgets {
		widget.SetVisible(!containerFull)
		if containerFull {
			continue
		}
		//   widget.before_layout
		if w := widget.Width(); w == 0 {
			pw := widget.PreferredWidth(writer)
			cw := ContentWidth(container)
			if pw == 0 {
				pw = cw
			}
			w = math.Min(pw, cw)
			widget.SetWidth(w)
		}
		if cx != 0 && (cx+widget.Width()) > ContentWidth(container) {
			cy += maxY + style.VPadding()
			cx, maxY = 0, 0
		}
		widget.SetLeft(ContentLeft(container) + cx)
		widget.SetTop(ContentTop(container) + cy)
		if h := widget.Height(); h == 0 {
			widget.SetHeight(widget.PreferredHeight(writer))
		}
		widget.LayoutWidget(writer)
		if widget.Bottom() > bottom {
			containerFull = true
			//     # widget.visible = (cy == 0)
			//     widget.visible = container.root_page.positioned_widgets[:static] == 0
			//     # $stderr.puts "+++flow+++ #{container.root_page.positioned_widgets[:static]}, visible: #{widget.visible}"
			continue
		}
		//   container.root_page.positioned_widgets[widget.position] += 1
		cx += widget.Width() + style.HPadding()
		maxY = math.Max(maxY, widget.Height())
	}
	// container.more(true) if container_full and container.overflow
	if container.Height() == 0 && maxY > 0 {
		container.SetHeight(cy + maxY + NonContentHeight(container))
	}
	// super(container, writer)
}

func LayoutHBox(container Container, style *LayoutStyle, writer Writer) {
	// fmt.Println("In LayoutHBox")
	containerFull := false

	static, remaining := printableWidgets(container, Static)
	for _, widget := range remaining {
		if widget.Printed() {
			widget.SetVisible(false)
		}
	}

	var lpanels, rpanels, unaligned []Widget
	for _, widget := range static {
		switch widget.Align() {
		case AlignLeft:
			lpanels = append(lpanels, widget)
		case AlignRight:
			rpanels = append(rpanels, widget)
		default:
			unaligned = append(unaligned, widget)
		}
	}

	var percents, specified, others []Widget
	for _, widget := range static {
		if widget.WidthPctIsSet() {
			percents = append(percents, widget)
		} else if widget.WidthIsSet() {
			specified = append(specified, widget)
		} else {
			others = append(others, widget)
		}
	}

	widthAvail := ContentWidth(container)

	// allocate specified widths first
	for _, widget := range specified {
		widthAvail -= widget.Width()
		containerFull = widthAvail < 0
		widget.SetDisabled(containerFull)
		widthAvail -= style.HPadding()
	}

	// allocate percent widths next, with a minimum width of 1 point
	if widthAvail-float64(len(percents)-1)*style.HPadding() >= float64(len(percents)) {
		widthAvail -= float64(len(percents)-1) * style.HPadding()
		totalPercents := 0.0
		for _, widget := range percents {
			totalPercents += widget.Width()
		}
		ratio := widthAvail / totalPercents
		for _, widget := range percents {
			if ratio < 1.0 {
				widget.SetWidth(widget.Width() * ratio)
			}
			widthAvail -= widget.Width()
		}
	} else {
		containerFull = true
		for _, widget := range percents {
			widget.SetDisabled(true)
		}
	}
	widthAvail -= style.HPadding()

	// divide remaining width equally among widgets with unspecified widths
	if widthAvail-float64(len(others)-1)*style.HPadding() >= float64(len(others)) {
		widthAvail -= float64(len(others)-1) * style.HPadding()
		othersWidth := widthAvail / float64(len(others))
		for _, widget := range others {
			widget.SetWidth(othersWidth)
		}
	} else {
		containerFull = true
		for _, widget := range others {
			widget.SetDisabled(true)
		}
	}

	for _, widget := range static {
		if container.Align() == AlignBottom {
			widget.SetBottom(ContentBottom(container))
		} else {
			containerFull = true
			widget.SetTop(ContentTop(container))
		}
		if !widget.HeightIsSet() {
			widget.SetHeight(widget.PreferredHeight(writer))
		}
	}

	left := ContentLeft(container)
	right := ContentRight(container)
	// fmt.Println("left:", left, "right:", right)

	for _, widget := range lpanels {
		if widget.Disabled() {
			continue
		}
		widget.SetLeft(left)
		left += (widget.Width() + style.HPadding())
	}
	for i := len(rpanels) - 1; i >= 0; i-- {
		widget := rpanels[i]
		if widget.Disabled() {
			continue
		}
		widget.SetRight(right)
		right -= (widget.Width() + style.HPadding())
	}
	for _, widget := range unaligned {
		if widget.Disabled() {
			continue
		}
		widget.SetLeft(left)
		left += (widget.Width() + style.HPadding())
	}

	if !container.HeightIsSet() {
		contentHeight := 0.0
		for _, widget := range static {
			if widget.Height() > contentHeight {
				contentHeight = widget.Height()
			}
		}
		container.SetHeight(contentHeight + NonContentHeight(container))
	}
	for _, widget := range static {
		if widget.Visible() && !widget.Disabled() {
			widget.LayoutWidget(writer)
		}
	}
	// super(container, writer)
}

func LayoutRelative(container Container, style *LayoutStyle, writer Writer) {

}

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
		} else if widget.WidthIsSet() {
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
	// allocate percent widths with a minimum width of 1 point
	if widthAvail-(float64(len(percents)-1))*style.HPadding() >= float64(len(percents)) {
		widthAvail -= float64((len(percents) - 1)) * style.HPadding()
		totalPercents := 0.0
		for i := 0; i < len(percents); i++ {
			totalPercents += percents[i].Size
		}
		ratio := widthAvail / totalPercents
		for i := 0; i < len(percents); i++ {
			if ratio < 1.0 {
				percents[i].Size *= ratio
			}
			widthAvail -= percents[i].Size
		}
	} else {
		for i := 0; i < len(percents); i++ {
			percents[i].Size = 0
		}
	}
	widthAvail -= style.HPadding()
	return widthAvail
}

func allocateOtherWidths(widthAvail float64, others SpecifiedSizes, style *LayoutStyle) float64 {
	// divide remaining width equally among widgets with unspecified widths
	if widthAvail-(float64(len(others)-1))*style.HPadding() >= float64(len(others)) {
		widthAvail -= float64(len(others)-1) * style.HPadding()
		othersWidth := widthAvail / float64(len(others))
		for i := 0; i < len(others); i++ {
			others[i].Size = othersWidth
		}
	} else {
		for i := 0; i < len(others); i++ {
			others[i].Size = 0
		}
	}
	return widthAvail
}

func LayoutTable(container Container, style *LayoutStyle, writer Writer) {
	// fmt.Println("In LayoutTable")
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
				widget.SetWidth(width + float64(widget.ColSpan()-1)*style.HPadding())
				var height float64
				if widget.HeightIsSet() {
					height = widget.Height()
				} else {
					height = widget.PreferredHeight(writer)
				}
				heights.SetCell(c, r, SpanSize{Span: widget.RowSpan(), Size: height})
			} else {
				// widget.SetVisible(false)
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
				heights.SetCell(c, r+1, SpanSize{Span: ss.Span - 1, Size: math.Max(ss.Size-maxHeight, 0)})
			}
			ss.Size = maxHeight
			heights.SetCell(c, r, ss)
		}
	}

	top := ContentTop(container)
	bottom := top + MaxContentHeight(container)
	for r := 0; r < grid.Rows(); r++ {
		maxHeight := 0.0
		left := ContentLeft(container)
		for c := 0; c < grid.Cols(); c++ {
			if widget := grid.Cell(c, r); widget != nil {
				widget.SetVisible(!containerFull)
				if containerFull {
					continue
				}
				ss := heights.Cell(c, r)
				widget.SetTop(top)
				widget.SetLeft(left)
				height := float64(ss.Span-1) * style.VPadding()
				for rowOffset := 0; rowOffset < ss.Span; rowOffset++ {
					height += heights.Cell(c, r+rowOffset).Size
				}
				widget.SetHeight(height)
				if ss.Span == 1 && ss.Size > maxHeight {
					maxHeight = ss.Size
				}
			}
			left += widths[c].Size + style.HPadding()
		}
		if containerFull {
			continue
		}
		if top+maxHeight > bottom {
			containerFull = true
			for c := 0; c < grid.Cols(); c++ {
				if widget := grid.Cell(c, r); widget != nil {
					widget.SetVisible(r == 0)
				}
			}
			// container.more(true) if container.overflow and (r > 0)
		}
		if !containerFull {
			top += maxHeight + style.VPadding()
		}
	}
	if !container.HeightIsSet() {
		container.SetHeight(top - ContentTop(container) + NonContentHeight(container) - style.VPadding())
	}
	static, remaining := printableWidgets(container, Static)
	for _, widget := range remaining {
		if widget.Printed() {
			widget.SetVisible(false)
		}
	}
	for _, widget := range static {
		widget.LayoutWidget(writer)
	}

	// super(container, writer)
}

func LayoutVBox(container Container, style *LayoutStyle, writer Writer) {
	// fmt.Println("In LayoutVBox")
	containerFull := false
	static, remaining := printableWidgets(container, Static)
	for _, widget := range remaining {
		if widget.Printed() {
			widget.SetVisible(false)
		}
	}
	var headers, footers, unaligned []Widget
	for _, widget := range static {
		switch widget.Align() {
		case AlignTop:
			headers = append(headers, widget)
		case AlignBottom:
			footers = append(footers, widget)
		default:
			unaligned = append(unaligned, widget)
		}
	}
	left := ContentLeft(container)
	// fmt.Println("left:", left)
	for _, widget := range static {
		if !widget.WidthIsSet() {
			pw := widget.PreferredWidth(writer)
			// fmt.Println("pw:", pw, widget)
			cw := ContentWidth(container)
			// fmt.Println("cw:", cw, container)
			if pw == 0 {
				pw = cw
			}
			w := math.Min(pw, cw)
			// fmt.Println("w:", w)
			// panic("foo")
			widget.SetWidth(w)
		}
		widget.SetLeft(left)
	}
	top, dy := ContentTop(container), 0.0
	// fmt.Println("top:", top)
	bottom := ContentTop(container) + MaxContentHeight(container)

	for i, widget := range headers {
		widget.SetTop(top)
		widget.LayoutWidget(writer)
		if !widget.HeightIsSet() {
			widget.SetHeight(widget.PreferredHeight(writer))
		}
		top += widget.Height() + style.VPadding()
		dy += widget.Height()
		if i > 0 {
			dy += style.VPadding()
		}
		widget.SetVisible(widget.Bottom() <= bottom)
	}

	if len(footers) > 0 {
		if !container.HeightIsSet() {
			container.SetHeightPct(100)
		}
		for i := len(footers) - 1; i >= 0; i-- {
			widget := footers[i]
			widget.SetBottom(bottom)
			widget.LayoutWidget(writer)
			if !widget.HeightIsSet() {
				widget.SetHeight(widget.PreferredHeight(writer))
			}
			widget.SetVisible(widget.Top() >= top)
		}
	}

	widgetsVisible := 0
	for i, widget := range unaligned {
		widget.SetVisible(!containerFull)
		if containerFull {
			continue
		}
		widget.SetTop(top)
		widget.LayoutWidget(writer)
		if !widget.HeightIsSet() {
			widget.SetHeight(widget.PreferredHeight(writer))
		}
		top += widget.Height()
		dy += widget.Height()
		if i > 0 {
			dy += style.VPadding()
		}
		if top > bottom {
			containerFull = true
			widget.SetVisible(widgetsVisible == 0)
			// widget.visible = widget.leaves > 0 and container.root_page.positioned_widgets[:static] == 0
		}
		if widget.Visible() {
			widgetsVisible += 1
		}
		top += style.VPadding()
	}
}

func printableWidgets(c Container, p Position) (widgets, remaining []Widget) {
	for _, w := range c.Widgets() {
		if w.Position() == p {
			widgets = append(widgets, w)
		} else {
			remaining = append(remaining, w)
		}
	}
	return
}

func init() {
	RegisterLayoutManager("absolute", LayoutAbsolute)
	RegisterLayoutManager("flow", LayoutFlow)
	RegisterLayoutManager("hbox", LayoutHBox)
	RegisterLayoutManager("relative", LayoutRelative)
	RegisterLayoutManager("table", LayoutTable)
	RegisterLayoutManager("vbox", LayoutVBox)
}
