package svg

import "github.com/rowland/leadtype/colors"

type Warning struct {
	Element   string
	Attribute string
	Message   string
}

type Document struct {
	Width     float64
	Height    float64
	ViewBox   ViewBox
	Children  []Node
	ClipPaths map[string]*ClipPath
	Classes   map[string]StyleSpec
	Warnings  []Warning
}

type ViewBox struct {
	MinX   float64
	MinY   float64
	Width  float64
	Height float64
}

type Node interface {
	node()
	NodeCommon() *Common
}

type Common struct {
	ID          string
	Transform   Transform
	Style       StyleSpec
	ClipPathRef string
}

func (c *Common) NodeCommon() *Common {
	return c
}

type Group struct {
	Common
	Children []Node
}

func (*Group) node() {}

type Defs struct {
	Common
	Children []Node
}

func (*Defs) node() {}

type ClipPath struct {
	Common
	Units    string
	Children []Node
}

func (*ClipPath) node() {}

type Line struct {
	Common
	X1 float64
	Y1 float64
	X2 float64
	Y2 float64
}

func (*Line) node() {}

type Rect struct {
	Common
	X      float64
	Y      float64
	Width  float64
	Height float64
	RX     float64
	RY     float64
}

func (*Rect) node() {}

type Circle struct {
	Common
	CX float64
	CY float64
	R  float64
}

func (*Circle) node() {}

type Ellipse struct {
	Common
	CX float64
	CY float64
	RX float64
	RY float64
}

func (*Ellipse) node() {}

type Polyline struct {
	Common
	Points []Point
}

func (*Polyline) node() {}

type Polygon struct {
	Common
	Points []Point
}

func (*Polygon) node() {}

type Path struct {
	Common
	Segments []Segment
}

func (*Path) node() {}

type Text struct {
	Common
	X    float64
	Y    float64
	Body string
}

func (*Text) node() {}

type Point struct {
	X float64
	Y float64
}

type SegmentType uint8

const (
	SegmentMoveTo SegmentType = iota
	SegmentLineTo
	SegmentCubicTo
	SegmentClosePath
)

type Segment struct {
	Type   SegmentType
	Points []Point
}

type Paint struct {
	Set   bool
	None  bool
	Color colors.Color
}

type StyleSpec struct {
	Fill          *Paint
	Stroke        *Paint
	StrokeWidth   *float64
	FillOpacity   *float64
	StrokeOpacity *float64
	Opacity       *float64
	LineCap       *string
	LineJoin      *string
	MiterLimit    *float64
	DashArray     *[]float64
	DashOffset    *float64
	FillRule      *string
	FontFamily    *string
	FontSize      *float64
	FontStyle     *string
	FontWeight    *string
	TextAnchor    *string
	Display       *string
}

type Style struct {
	Fill          Paint
	Stroke        Paint
	StrokeWidth   float64
	FillOpacity   float64
	StrokeOpacity float64
	Opacity       float64
	LineCap       string
	LineJoin      string
	MiterLimit    float64
	DashArray     []float64
	DashOffset    float64
	FillRule      string
	FontFamily    string
	FontSize      float64
	FontStyle     string
	FontWeight    string
	TextAnchor    string
	Display       string
}

func DefaultStyle() Style {
	return Style{
		Fill: Paint{
			Set:   true,
			Color: colors.Black,
		},
		Stroke: Paint{
			Set:  true,
			None: true,
		},
		StrokeWidth:   1,
		FillOpacity:   1,
		StrokeOpacity: 1,
		Opacity:       1,
		LineCap:       "butt",
		LineJoin:      "miter",
		MiterLimit:    4,
		FillRule:      "nonzero",
		FontFamily:    "Helvetica",
		FontSize:      16,
		FontStyle:     "normal",
		FontWeight:    "normal",
		TextAnchor:    "start",
		Display:       "inline",
	}
}

func (spec StyleSpec) Resolve(parent Style) Style {
	out := parent
	if spec.Fill != nil {
		out.Fill = *spec.Fill
	}
	if spec.Stroke != nil {
		out.Stroke = *spec.Stroke
	}
	if spec.StrokeWidth != nil {
		out.StrokeWidth = *spec.StrokeWidth
	}
	if spec.FillOpacity != nil {
		out.FillOpacity = *spec.FillOpacity
	}
	if spec.StrokeOpacity != nil {
		out.StrokeOpacity = *spec.StrokeOpacity
	}
	if spec.Opacity != nil {
		out.Opacity = *spec.Opacity
	}
	if spec.LineCap != nil {
		out.LineCap = *spec.LineCap
	}
	if spec.LineJoin != nil {
		out.LineJoin = *spec.LineJoin
	}
	if spec.MiterLimit != nil {
		out.MiterLimit = *spec.MiterLimit
	}
	if spec.DashArray != nil {
		out.DashArray = append([]float64(nil), (*spec.DashArray)...)
	}
	if spec.DashOffset != nil {
		out.DashOffset = *spec.DashOffset
	}
	if spec.FillRule != nil {
		out.FillRule = *spec.FillRule
	}
	if spec.FontFamily != nil {
		out.FontFamily = *spec.FontFamily
	}
	if spec.FontSize != nil {
		out.FontSize = *spec.FontSize
	}
	if spec.FontStyle != nil {
		out.FontStyle = *spec.FontStyle
	}
	if spec.FontWeight != nil {
		out.FontWeight = *spec.FontWeight
	}
	if spec.TextAnchor != nil {
		out.TextAnchor = *spec.TextAnchor
	}
	if spec.Display != nil {
		out.Display = *spec.Display
	}
	return out
}
