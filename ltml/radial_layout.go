package ltml

import (
	"errors"
	"fmt"
	"math"
)

var errRadialNeedsDimension = errors.New("radial layout requires rows or cols")
var errRadialNeedsAngles = errors.New("radial layout angles must contain at least two boundaries")

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
	cells := radialCellWidgets(container)
	if len(cells) == 0 {
		return
	}

	rowsHint, colsHint := container.Rows(), container.Cols()
	if base, ok := container.(*StdContainer); ok && len(base.Angles()) > 0 {
		colsFromAngles := len(base.Angles()) - 1
		if colsHint > 0 && colsHint != colsFromAngles {
			panic(fmt.Errorf("radial cols=%d conflicts with %d angle spans", colsHint, colsFromAngles))
		}
		colsHint = colsFromAngles
	}

	grid, rows, cols, err := deriveRadialGrid(cells, container.Order(), rowsHint, colsHint)
	if err != nil {
		panic(err)
	}
	angles, err := radialBoundaries(container, cols)
	if err != nil {
		panic(err)
	}
	centerX, centerY, innerRadius, outerRadius := radialContainerGeometry(container)
	if outerRadius <= innerRadius {
		panic(fmt.Errorf("radial outer radius %g must be greater than inner radius %g", outerRadius, innerRadius))
	}

	step := (outerRadius - innerRadius) / float64(rows)
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
			outer := outerRadius - (step * float64(row))
			inner := outerRadius - (step * float64(row+rowSpan))
			startAngle := angles[col]
			endAngle := angles[col+colSpan]
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

func radialBoundaries(container Container, cols int) ([]float64, error) {
	baseAngle := 0.0
	explicit := []float64(nil)
	if base, ok := container.(*StdContainer); ok {
		baseAngle = base.BaseAngle()
		explicit = append(explicit, base.Angles()...)
	}
	if len(explicit) > 0 {
		if len(explicit) < 2 {
			return nil, errRadialNeedsAngles
		}
		angles := make([]float64, len(explicit))
		prev := explicit[0]
		angles[0] = baseAngle + prev
		for i := 1; i < len(explicit); i++ {
			if explicit[i] <= prev {
				return nil, errors.New("radial angles must be strictly increasing")
			}
			prev = explicit[i]
			angles[i] = baseAngle + explicit[i]
		}
		return angles, nil
	}
	if cols < 1 {
		return nil, errRadialNeedsDimension
	}
	angles := make([]float64, cols+1)
	step := 360.0 / float64(cols)
	for i := 0; i <= cols; i++ {
		angles[i] = baseAngle + (step * float64(i))
	}
	return angles, nil
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
		if value := base.RadiusValue(); value > 0 {
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
}
