package ltml

import (
	"errors"
	"fmt"
	"math"
	"sort"
)

var errRadialNeedsDimension = errors.New("radial layout requires rows or cols")
var errRadialNeedsAngles = errors.New("radial layout angles must contain at least one boundary")

const radialAngleEpsilon = 1e-9

type radialSweep uint8

const (
	radialSweepCCW radialSweep = iota
	radialSweepCW
)

type radialAngleSpan struct {
	StartAngle float64
	EndAngle   float64
}

func isRadialLayoutManager(manager string) bool {
	return manager == "radial" || manager == "radial-out"
}

func isRadialLayoutStyle(style *LayoutStyle) bool {
	return style != nil && isRadialLayoutManager(style.manager)
}

func radialRowsGrowOutward(style *LayoutStyle) bool {
	return style != nil && style.manager == "radial-out"
}

type radialPoint struct {
	X float64
	Y float64
}

type radialBounds struct {
	MinX float64
	MinY float64
	MaxX float64
	MaxY float64
}

type radialInterval struct {
	MinX float64
	MaxX float64
}

type radialSectorGeometry struct {
	CenterX     float64
	CenterY     float64
	InnerRadius float64
	OuterRadius float64
	StartAngle  float64
	EndAngle    float64
	AnchorAngle float64
	AnchorX     float64
	AnchorY     float64
}

func LayoutRadialTable(container Container, style *LayoutStyle, writer Writer) {
	inferRadialContainerDimensions(container)
	cells := radialCellWidgets(container)
	if len(cells) == 0 {
		return
	}

	rowsHint, colsHint := container.Rows(), container.Cols()
	if base, ok := container.(*StdContainer); ok && len(base.Angles()) > 0 {
		colsFromAngles := radialExplicitSectorCount(base.Angles())
		if colsHint > 0 && colsHint != colsFromAngles {
			panic(fmt.Errorf("radial cols=%d conflicts with %d normalized angle spans", colsHint, colsFromAngles))
		}
		colsHint = colsFromAngles
	}

	grid, rows, cols, err := deriveRadialGrid(cells, container.Order(), rowsHint, colsHint)
	if err != nil {
		panic(err)
	}
	spans, err := radialAngleSpans(container, cols)
	if err != nil {
		panic(err)
	}
	centerX, centerY, innerRadius, outerRadius := radialContainerGeometry(container)
	if outerRadius <= innerRadius {
		debugf("invalid radial geometry: outer radius %g must be greater than inner radius %g\n", outerRadius, innerRadius)
		return
	}

	rowsGrowOutward := radialRowsGrowOutward(style)
	for row := 0; row < grid.Rows(); row++ {
		for col := 0; col < grid.Cols(); col++ {
			widget := grid.Cell(col, row)
			if widget == nil {
				continue
			}
			sector, ok := widget.(*StdSector)
			if !ok {
				panic(fmt.Errorf("radial layout child %T is not a sector", widget))
			}
			rowSpan := sector.RowSpan()
			colSpan := sector.ColSpan()
			inner, outer := radialTrackBounds(innerRadius, outerRadius, rows, row, rowSpan, rowsGrowOutward)
			startAngle := spans[col].StartAngle
			endAngle := spans[col+colSpan-1].EndAngle
			geometry := radialSectorGeometry{
				CenterX:     centerX,
				CenterY:     centerY,
				InnerRadius: max(inner, innerRadius),
				OuterRadius: outer,
				StartAngle:  startAngle,
				EndAngle:    endAngle,
				AnchorAngle: startAngle + (endAngle-startAngle)/2,
			}
			geometry.AnchorX, geometry.AnchorY = radialPointAt(centerX, centerY, (geometry.InnerRadius+geometry.OuterRadius)/2, geometry.AnchorAngle)
			sector.setGeometry(geometry)
			sector.LayoutWidget(writer)
		}
	}
	if !container.HeightIsSet() {
		container.SetHeight((outerRadius * 2) + NonContentHeight(container))
	}
	if !container.WidthIsSet() {
		container.SetWidth((outerRadius * 2) + NonContentWidth(container))
	}
}

func inferRadialContainerDimensions(container Container) {
	base, ok := container.(*StdContainer)
	if !ok || !isRadialLayoutStyle(base.layout) {
		return
	}
	if !base.WidthIsSet() {
		if width, ok := base.radialInferredWidth(); ok {
			base.SetWidth(width)
		}
	}
	if !base.HeightIsSet() {
		if height, ok := base.radialInferredHeight(); ok {
			base.SetHeight(height)
		}
	}
}

func radialTrackBounds(innerRadius, outerRadius float64, rows, row, rowSpan int, rowsGrowOutward bool) (float64, float64) {
	step := (outerRadius - innerRadius) / float64(rows)
	if rowsGrowOutward {
		inner := innerRadius + (step * float64(row))
		outer := innerRadius + (step * float64(row+rowSpan))
		return inner, outer
	}
	outer := outerRadius - (step * float64(row))
	inner := outerRadius - (step * float64(row+rowSpan))
	return inner, outer
}

func boundsForPoints(points []radialPoint) radialBounds {
	if len(points) == 0 {
		return radialBounds{}
	}
	bounds := radialBounds{
		MinX: points[0].X,
		MinY: points[0].Y,
		MaxX: points[0].X,
		MaxY: points[0].Y,
	}
	for _, point := range points[1:] {
		bounds.MinX = min(bounds.MinX, point.X)
		bounds.MinY = min(bounds.MinY, point.Y)
		bounds.MaxX = max(bounds.MaxX, point.X)
		bounds.MaxY = max(bounds.MaxY, point.Y)
	}
	return bounds
}

func deriveRadialGrid(widgets []Widget, order TableOrder, rowsHint, colsHint int) (*WidgetGrid, int, int, error) {
	if rowsHint <= 0 && colsHint <= 0 {
		return nil, 0, 0, errRadialNeedsDimension
	}
	switch order {
	case TableOrderRows:
		if colsHint > 0 {
			grid, err := radialRowGrid(widgets, colsHint)
			if err != nil {
				return nil, 0, 0, err
			}
			rows := max(rowsHint, grid.Rows())
			if rowsHint > 0 && grid.Rows() > rowsHint {
				return nil, 0, 0, fmt.Errorf("radial rows=%d is too small for the data", rowsHint)
			}
			return grid, rows, colsHint, nil
		}
		maxCols := totalColSpan(widgets)
		for cols := max(maxColSpan(widgets), 1); cols <= maxCols; cols++ {
			grid, err := radialRowGrid(widgets, cols)
			if err != nil {
				return nil, 0, 0, err
			}
			if grid.Rows() <= rowsHint {
				return grid, rowsHint, cols, nil
			}
		}
	case TableOrderCols:
		if rowsHint > 0 {
			grid, err := radialColGrid(widgets, rowsHint)
			if err != nil {
				return nil, 0, 0, err
			}
			cols := max(colsHint, grid.Cols())
			if colsHint > 0 && grid.Cols() > colsHint {
				return nil, 0, 0, fmt.Errorf("radial cols=%d is too small for the data", colsHint)
			}
			return grid, rowsHint, cols, nil
		}
		maxRows := totalRowSpan(widgets)
		for rows := max(maxRowSpan(widgets), 1); rows <= maxRows; rows++ {
			grid, err := radialColGrid(widgets, rows)
			if err != nil {
				return nil, 0, 0, err
			}
			if grid.Cols() <= colsHint {
				return grid, rows, colsHint, nil
			}
		}
	}
	return nil, 0, 0, fmt.Errorf("unable to derive radial grid")
}

func maxColSpan(widgets []Widget) int {
	value := 1
	for _, widget := range widgets {
		value = max(value, widget.ColSpan())
	}
	return value
}

func maxRowSpan(widgets []Widget) int {
	value := 1
	for _, widget := range widgets {
		value = max(value, widget.RowSpan())
	}
	return value
}

func totalColSpan(widgets []Widget) int {
	total := 0
	for _, widget := range widgets {
		total += widget.ColSpan()
	}
	return max(total, 1)
}

func totalRowSpan(widgets []Widget) int {
	total := 0
	for _, widget := range widgets {
		total += widget.RowSpan()
	}
	return max(total, 1)
}

func radialRowGrid(widgets []Widget, cols int) (*WidgetGrid, error) {
	if cols < 1 {
		return nil, errors.New("radial cols must be specified")
	}
	used := NewBoolGrid(cols, 0)
	grid := NewWidgetGrid(cols, 0)
	row, col := 0, 0
	for _, widget := range widgets {
		for used.Cell(col, row) {
			col++
			if col >= cols {
				row++
				col = 0
			}
		}
		grid.SetCell(col, row, widget)
		markGrid(used, col, row, widget.ColSpan(), widget.RowSpan(), true)
		col += widget.ColSpan()
		if col > cols {
			return nil, errors.New("radial colspan exceeds the number of columns")
		}
		if col == cols {
			row++
			col = 0
		}
	}
	return grid, nil
}

func radialColGrid(widgets []Widget, rows int) (*WidgetGrid, error) {
	if rows < 1 {
		return nil, errors.New("radial rows must be specified")
	}
	used := NewBoolGrid(0, rows)
	grid := NewWidgetGrid(0, rows)
	row, col := 0, 0
	for _, widget := range widgets {
		for used.Cell(col, row) {
			row++
			if row >= rows {
				col++
				row = 0
			}
		}
		if row >= rows {
			col++
			row = 0
		}
		grid.SetCell(col, row, widget)
		markGrid(used, col, row, widget.ColSpan(), widget.RowSpan(), true)
		row += widget.RowSpan()
		if row > rows {
			return nil, errors.New("radial rowspan exceeds the number of rows")
		}
	}
	return grid, nil
}

func radialAngleSpans(container Container, cols int) ([]radialAngleSpan, error) {
	baseAngle := 0.0
	explicit := []float64(nil)
	sweep := radialSweepCCW
	if base, ok := container.(*StdContainer); ok {
		baseAngle = base.BaseAngle()
		explicit = append(explicit, base.Angles()...)
		sweep = base.RadialSweep()
	}
	if len(explicit) > 0 {
		boundaries := radialNormalizedBoundaries(explicit)
		if len(boundaries) == 0 {
			return nil, errRadialNeedsAngles
		}
		return radialSpansFromBoundaries(boundaries, baseAngle, sweep), nil
	}
	if cols < 1 {
		return nil, errRadialNeedsDimension
	}
	boundaries := make([]float64, cols)
	step := 360.0 / float64(cols)
	for i := 0; i < cols; i++ {
		boundaries[i] = step * float64(i)
	}
	return radialSpansFromBoundaries(boundaries, baseAngle, sweep), nil
}

func radialExplicitSectorCount(explicit []float64) int {
	return len(radialNormalizedBoundaries(explicit))
}

func radialNormalizedBoundaries(explicit []float64) []float64 {
	if len(explicit) == 0 {
		return nil
	}
	normalized := make([]float64, 0, len(explicit))
	for _, angle := range explicit {
		normalized = append(normalized, normalizeRadialAngle(angle))
	}
	sort.Float64s(normalized)
	deduped := normalized[:0]
	for _, angle := range normalized {
		if len(deduped) == 0 || math.Abs(angle-deduped[len(deduped)-1]) > radialAngleEpsilon {
			deduped = append(deduped, angle)
		}
	}
	return append([]float64(nil), deduped...)
}

func normalizeRadialAngle(angle float64) float64 {
	value := math.Mod(angle, 360)
	if value < 0 {
		value += 360
	}
	if math.Abs(value-360) <= radialAngleEpsilon {
		return 0
	}
	return value
}

func radialSpansFromBoundaries(boundaries []float64, baseAngle float64, sweep radialSweep) []radialAngleSpan {
	if len(boundaries) == 0 {
		return nil
	}
	if len(boundaries) == 1 {
		start := baseAngle + boundaries[0]
		end := start + 360
		if sweep == radialSweepCW {
			end = start - 360
		}
		return []radialAngleSpan{{StartAngle: start, EndAngle: end}}
	}

	spans := make([]radialAngleSpan, len(boundaries))
	for i, boundary := range boundaries {
		start := baseAngle + boundary
		if sweep == radialSweepCW {
			prev := boundaries[(i+len(boundaries)-1)%len(boundaries)]
			end := baseAngle + prev
			if prev >= boundary-radialAngleEpsilon {
				end -= 360
			}
			spans[i] = radialAngleSpan{StartAngle: start, EndAngle: end}
			continue
		}
		next := boundaries[(i+1)%len(boundaries)]
		end := baseAngle + next
		if next <= boundary+radialAngleEpsilon {
			end += 360
		}
		spans[i] = radialAngleSpan{StartAngle: start, EndAngle: end}
	}
	return spans
}

func radialContainerGeometry(container Container) (centerX, centerY, innerRadius, outerRadius float64) {
	centerX = ContentLeft(container) + ContentWidth(container)/2
	centerY = ContentTop(container) + ContentHeight(container)/2
	outerRadius = min(ContentWidth(container), ContentHeight(container)) / 2
	if base, ok := container.(*StdContainer); ok {
		if value, set := base.CenterX(); set {
			centerX = ContentLeft(container) + value
		}
		if value, set := base.CenterY(); set {
			centerY = ContentTop(container) + value
		}
		if value := base.OuterRadius(); value > 0 {
			outerRadius = value
		}
		if value := base.InnerRadius(); value > 0 {
			innerRadius = value
		}
	}
	return
}

func radialCellWidgets(c Container) []Widget {
	root := rootPageForContainer(c)
	physicalPageNo := 0
	flowPageIndex := 1
	parentRepeats := c.Display() != DisplayOnce
	if root != nil {
		parentRepeats = c == root || c.Display() != DisplayOnce
		flowPageIndex = root.flowPageIndex
		if doc := root.document(); doc != nil {
			physicalPageNo = doc.CurrentPhysicalPageNo()
		}
	}
	var widgets []Widget
	for _, child := range c.Widgets() {
		if _, ok := child.(*StdSector); !ok {
			continue
		}
		if widgetDisplayForRender(child, parentRepeats, flowPageIndex, physicalPageNo) {
			widgets = append(widgets, child)
		} else {
			child.SetVisible(false)
		}
	}
	return widgets
}

func radialPointAt(centerX, centerY, radius, angle float64) (float64, float64) {
	theta := angle * math.Pi / 180
	return centerX + radius*math.Cos(theta), centerY - radius*math.Sin(theta)
}

func rotatePagePoint(x, y, centerX, centerY, angle float64) (float64, float64) {
	theta := angle * math.Pi / 180
	dx := x - centerX
	dy := y - centerY
	return centerX + (dx * math.Cos(theta)) - (dy * math.Sin(theta)),
		centerY + (dx * math.Sin(theta)) + (dy * math.Cos(theta))
}

func init() {
	RegisterLayoutManager("radial", LayoutRadialTable)
	RegisterLayoutManager("radial-out", LayoutRadialTable)
}
