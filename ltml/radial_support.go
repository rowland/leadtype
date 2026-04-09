package ltml

import "math"

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

type radialCell interface {
	Widget
	setGeometry(radialSectorGeometry)
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
