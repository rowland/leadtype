package pdf

import (
	"fmt"
	"math"
)

type ClosedShapeKind string

const (
	ClosedShapeCircle  ClosedShapeKind = "circle"
	ClosedShapeEllipse ClosedShapeKind = "ellipse"
	ClosedShapePolygon ClosedShapeKind = "polygon"
	ClosedShapeStar    ClosedShapeKind = "star"
)

type ClosedShape struct {
	Kind        ClosedShapeKind
	Center      Location
	Radius      float64
	RadiusX     float64
	RadiusY     float64
	InnerRadius float64
	Sides       int
	Points      int
	Rotation    float64
	Reverse     bool
}

func (s ClosedShape) normalized() ClosedShape {
	switch s.Kind {
	case ClosedShapeCircle:
		if s.RadiusX == 0 {
			s.RadiusX = s.Radius
		}
		if s.RadiusY == 0 {
			s.RadiusY = s.Radius
		}
	case ClosedShapeEllipse:
		if s.RadiusX == 0 {
			s.RadiusX = s.Radius
		}
		if s.RadiusY == 0 {
			s.RadiusY = s.Radius
		}
	case ClosedShapeStar:
		if s.InnerRadius == 0 {
			s.InnerRadius = s.Radius * 0.5
		}
	}
	return s
}

func (s ClosedShape) validate() error {
	s = s.normalized()
	switch s.Kind {
	case ClosedShapeCircle:
		if s.RadiusX <= 0 || s.RadiusY <= 0 {
			return fmt.Errorf("circle radius must be positive")
		}
	case ClosedShapeEllipse:
		if s.RadiusX <= 0 || s.RadiusY <= 0 {
			return fmt.Errorf("ellipse radii must be positive")
		}
	case ClosedShapePolygon:
		if s.Radius <= 0 {
			return fmt.Errorf("polygon radius must be positive")
		}
		if s.Sides < 3 {
			return errInvalidPolygonSides
		}
	case ClosedShapeStar:
		if s.Radius <= 0 || s.InnerRadius <= 0 {
			return fmt.Errorf("star radii must be positive")
		}
		if s.Points < 5 {
			return errInvalidStarPoints
		}
	default:
		return fmt.Errorf("unsupported closed shape %q", s.Kind)
	}
	return nil
}

func (s ClosedShape) Bounds() (Bounds, error) {
	points, err := s.pathPoints()
	if err != nil {
		return Bounds{}, err
	}
	return boundsForLocations(points), nil
}

func (s ClosedShape) pathPoints() ([]Location, error) {
	s = s.normalized()
	if err := s.validate(); err != nil {
		return nil, err
	}
	switch s.Kind {
	case ClosedShapeCircle:
		return circlePoints(s.Center.X, s.Center.Y, s.RadiusX), nil
	case ClosedShapeEllipse:
		return ellipsePoints(s.Center.X, s.Center.Y, s.RadiusX, s.RadiusY), nil
	case ClosedShapePolygon:
		return polygonPoints(s.Center.X, s.Center.Y, s.Radius, s.Sides, s.Rotation), nil
	case ClosedShapeStar:
		return starPoints(s.Center.X, s.Center.Y, s.Radius, s.InnerRadius, s.Points, s.Rotation), nil
	default:
		return nil, fmt.Errorf("unsupported closed shape %q", s.Kind)
	}
}

func circlePoints(x, y, r float64) []Location {
	points := make([]Location, 0, 13)
	for q := 1; q <= 4; q++ {
		points = append(points, quadrantBezierPoints(q, x, y, r, r)...)
	}
	for _, i := range []int{12, 8, 4} {
		points = append(points[:i], points[i+1:]...)
	}
	return points
}

func ellipsePoints(x, y, rx, ry float64) []Location {
	points := make([]Location, 0, 13)
	for q := 1; q <= 4; q++ {
		points = append(points, quadrantBezierPoints(q, x, y, rx, ry)...)
	}
	for _, i := range []int{12, 8, 4} {
		points = append(points[:i], points[i+1:]...)
	}
	return points
}

func polygonPoints(x, y, r float64, sides int, rotation float64) []Location {
	if sides < 3 {
		return nil
	}
	step := 360.0 / float64(sides)
	angle := step/2.0 + 90.0
	points := make([]Location, 0, sides+1)
	for i := 0; i <= sides; i++ {
		dx, dy := rotateXY(1, 0, angle)
		points = append(points, Location{x + dx*r, y - dy*r})
		angle += step
	}
	if rotation != 0 {
		center := Location{x, y}
		for i := range points {
			points[i] = rotatePoint(center, points[i], -rotation)
		}
	}
	return points
}

func starPoints(x, y, r1, r2 float64, points int, rotation float64) []Location {
	outer := polygonPoints(x, y, r1, points, rotation)
	inner := polygonPoints(x, y, r2, points, rotation+(360.0/float64(points)/2.0))
	if len(outer) == 0 || len(inner) == 0 {
		return nil
	}
	path := make([]Location, 0, points*2+1)
	path = append(path, inner[0])
	for i := 0; i < points; i++ {
		path = append(path, outer[i], inner[i+1])
	}
	return path
}

func boundsForLocations(points []Location) Bounds {
	if len(points) == 0 {
		return Bounds{}
	}
	bounds := Bounds{
		MinX: points[0].X,
		MaxX: points[0].X,
		MinY: points[0].Y,
		MaxY: points[0].Y,
	}
	for _, point := range points[1:] {
		bounds.MinX = math.Min(bounds.MinX, point.X)
		bounds.MaxX = math.Max(bounds.MaxX, point.X)
		bounds.MinY = math.Min(bounds.MinY, point.Y)
		bounds.MaxY = math.Max(bounds.MaxY, point.Y)
	}
	return bounds
}
