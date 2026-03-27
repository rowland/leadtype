// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
	"strconv"
)

type StdShape struct {
	StdContainer
	reverse bool
}

func (s *StdShape) DrawBorder(w Writer) error {
	return nil
}

func (s *StdShape) PaintBackground(w Writer) error {
	return nil
}

func (s *StdShape) SetAttrs(attrs map[string]string) {
	s.StdContainer.SetAttrs(attrs)
	if reverse, ok := attrs["reverse"]; ok {
		s.reverse = reverse == "true"
	}
}

func (s *StdShape) drawChildren(w Writer) error {
	return s.StdContainer.DrawContent(w)
}

func (s *StdShape) hasBorderOrFill() (bool, bool) {
	return s.border != nil, s.fill != nil
}

func (s *StdShape) applyBorderAndFill(w Writer) {
	if s.border != nil {
		s.border.Apply(w)
	}
	if s.fill != nil {
		s.fill.Apply(w)
	}
}

func (s *StdShape) center() (float64, float64) {
	return s.shapeLeft() + s.shapeWidth()/2, s.shapeTop() + s.shapeHeight()/2
}

func (s *StdShape) radius() float64 {
	shapeWidth := s.shapeWidth()
	shapeHeight := s.shapeHeight()
	if shapeWidth <= 0 && shapeHeight <= 0 {
		return 0
	}
	if shapeWidth <= 0 {
		return shapeHeight / 2
	}
	if shapeHeight <= 0 {
		return shapeWidth / 2
	}
	if shapeWidth < shapeHeight {
		return shapeWidth / 2
	}
	return shapeHeight / 2
}

func (s *StdShape) shapeHeight() float64 {
	return s.Height() - s.MarginTop() - s.MarginBottom()
}

func (s *StdShape) shapeLeft() float64 {
	return s.Left() + s.MarginLeft()
}

func (s *StdShape) shapeTop() float64 {
	return s.Top() + s.MarginTop()
}

func (s *StdShape) shapeWidth() float64 {
	return s.Width() - s.MarginLeft() - s.MarginRight()
}

type StdCircle struct {
	StdShape
	radiusValue float64
}

func (c *StdCircle) DrawContent(w Writer) error {
	c.applyBorderAndFill(w)
	x, y := c.center()
	if err := w.Circle(x, y, c.radius(), c.border != nil, c.fill != nil, c.reverse); err != nil {
		return err
	}
	return c.drawChildren(w)
}

func (c *StdCircle) PreferredHeight(Writer) float64 {
	if c.height != 0 {
		return c.height
	}
	if c.width != 0 {
		return c.width
	}
	if c.radiusValue != 0 {
		return c.radiusValue*2 + NonContentHeight(c)
	}
	return NonContentHeight(c)
}

func (c *StdCircle) PreferredWidth(Writer) float64 {
	if c.width != 0 {
		return c.width
	}
	if c.height != 0 {
		return c.height
	}
	if c.radiusValue != 0 {
		return c.radiusValue*2 + NonContentWidth(c)
	}
	return NonContentWidth(c)
}

func (c *StdCircle) SetAttrs(attrs map[string]string) {
	c.StdShape.SetAttrs(attrs)
	if r, ok := attrs["r"]; ok {
		c.radiusValue = ParseMeasurement(r, c.Units())
	}
}

func (c *StdCircle) radius() float64 {
	if c.radiusValue != 0 {
		return c.radiusValue
	}
	return c.StdShape.radius()
}

type StdEllipse struct {
	StdShape
	rx float64
	ry float64
}

func (e *StdEllipse) DrawContent(w Writer) error {
	e.applyBorderAndFill(w)
	x, y := e.center()
	if err := w.Ellipse(x, y, e.radiusX(), e.radiusY(), e.border != nil, e.fill != nil, e.reverse); err != nil {
		return err
	}
	return e.drawChildren(w)
}

func (e *StdEllipse) SetAttrs(attrs map[string]string) {
	e.StdShape.SetAttrs(attrs)
	if rx, ok := attrs["rx"]; ok {
		e.rx = ParseMeasurement(rx, e.Units())
	}
	if ry, ok := attrs["ry"]; ok {
		e.ry = ParseMeasurement(ry, e.Units())
	}
}

func (e *StdEllipse) radiusX() float64 {
	if e.rx != 0 {
		return e.rx
	}
	return e.shapeWidth() / 2
}

func (e *StdEllipse) radiusY() float64 {
	if e.ry != 0 {
		return e.ry
	}
	return e.shapeHeight() / 2
}

type StdPolygon struct {
	StdShape
	r        float64
	sides    int
	rotation float64
}

func (p *StdPolygon) DrawContent(w Writer) error {
	p.applyBorderAndFill(w)
	x, y := p.center()
	if err := w.Polygon(x, y, p.radius(), p.Sides(), p.border != nil, p.fill != nil, p.reverse, p.rotation); err != nil {
		return err
	}
	return p.drawChildren(w)
}

func (p *StdPolygon) SetAttrs(attrs map[string]string) {
	p.StdShape.SetAttrs(attrs)
	if r, ok := attrs["r"]; ok {
		p.r = ParseMeasurement(r, p.Units())
	}
	if sides, ok := attrs["sides"]; ok {
		p.sides, _ = strconv.Atoi(sides)
	}
	if rotation, ok := attrs["rotation"]; ok {
		p.rotation, _ = strconv.ParseFloat(rotation, 64)
	}
}

func (p *StdPolygon) PreferredHeight(Writer) float64 {
	if p.height != 0 {
		return p.height
	}
	if p.r != 0 {
		return p.r*2 + NonContentHeight(p)
	}
	return NonContentHeight(p)
}

func (p *StdPolygon) PreferredWidth(Writer) float64 {
	if p.width != 0 {
		return p.width
	}
	if p.r != 0 {
		return p.r*2 + NonContentWidth(p)
	}
	return NonContentWidth(p)
}

func (p *StdPolygon) radius() float64 {
	if p.r != 0 {
		return p.r
	}
	return p.StdShape.radius()
}

func (p *StdPolygon) Sides() int {
	if p.sides >= 3 {
		return p.sides
	}
	return 3
}

type StdStar struct {
	StdShape
	r1       float64
	r2       float64
	points   int
	rotation float64
}

func (s *StdStar) DrawContent(w Writer) error {
	s.applyBorderAndFill(w)
	x, y := s.center()
	if err := w.Star(x, y, s.outerRadius(), s.innerRadius(), s.Points(), s.border != nil, s.fill != nil, s.reverse, s.rotation); err != nil {
		return err
	}
	return s.drawChildren(w)
}

func (s *StdStar) SetAttrs(attrs map[string]string) {
	s.StdShape.SetAttrs(attrs)
	if r1, ok := attrs["r1"]; ok {
		s.r1 = ParseMeasurement(r1, s.Units())
	}
	if r, ok := attrs["r"]; ok && s.r1 == 0 {
		s.r1 = ParseMeasurement(r, s.Units())
	}
	if r2, ok := attrs["r2"]; ok {
		s.r2 = ParseMeasurement(r2, s.Units())
	}
	if points, ok := attrs["points"]; ok {
		s.points, _ = strconv.Atoi(points)
	}
	if rotation, ok := attrs["rotation"]; ok {
		s.rotation, _ = strconv.ParseFloat(rotation, 64)
	}
}

func (s *StdStar) PreferredHeight(Writer) float64 {
	if s.height != 0 {
		return s.height
	}
	if s.r1 != 0 {
		return s.r1*2 + NonContentHeight(s)
	}
	return NonContentHeight(s)
}

func (s *StdStar) PreferredWidth(Writer) float64 {
	if s.width != 0 {
		return s.width
	}
	if s.r1 != 0 {
		return s.r1*2 + NonContentWidth(s)
	}
	return NonContentWidth(s)
}

func (s *StdStar) outerRadius() float64 {
	if s.r1 != 0 {
		return s.r1
	}
	return s.StdShape.radius()
}

func (s *StdStar) innerRadius() float64 {
	if s.r2 != 0 {
		return s.r2
	}
	return s.outerRadius() * 0.5
}

func (s *StdStar) Points() int {
	if s.points >= 5 {
		return s.points
	}
	return 5
}

type StdArc struct {
	StdShape
	r          float64
	startAngle float64
	endAngle   float64
}

func (a *StdArc) DrawContent(w Writer) error {
	a.applyBorderAndFill(w)
	x, y := a.center()
	var err error
	if err = w.Path(func() {
		if e := w.Arc(x, y, a.radius(), a.startAngle, a.endAngle, true); e != nil {
			err = e
			return
		}
		if e := w.Stroke(); e != nil {
			err = e
		}
	}); err != nil {
		return err
	}
	if err != nil {
		return err
	}
	return a.drawChildren(w)
}

func (a *StdArc) PreferredHeight(Writer) float64 {
	if a.height != 0 {
		return a.height
	}
	if a.r != 0 {
		return a.r*2 + NonContentHeight(a)
	}
	return NonContentHeight(a)
}

func (a *StdArc) PreferredWidth(Writer) float64 {
	if a.width != 0 {
		return a.width
	}
	if a.r != 0 {
		return a.r*2 + NonContentWidth(a)
	}
	return NonContentWidth(a)
}

func (a *StdArc) SetAttrs(attrs map[string]string) {
	a.StdShape.SetAttrs(attrs)
	if r, ok := attrs["r"]; ok {
		a.r = ParseMeasurement(r, a.Units())
	}
	if startAngle, ok := attrs["start_angle"]; ok {
		a.startAngle, _ = strconv.ParseFloat(startAngle, 64)
	}
	if endAngle, ok := attrs["end_angle"]; ok {
		a.endAngle, _ = strconv.ParseFloat(endAngle, 64)
	}
}

func (a *StdArc) radius() float64 {
	if a.r != 0 {
		return a.r
	}
	return a.StdShape.radius()
}

type StdPie struct {
	StdArc
}

func (p *StdPie) DrawContent(w Writer) error {
	p.applyBorderAndFill(w)
	x, y := p.center()
	if err := w.Pie(x, y, p.radius(), p.startAngle, p.endAngle, p.border != nil, p.fill != nil, p.reverse); err != nil {
		return err
	}
	return p.drawChildren(w)
}

type StdArch struct {
	StdShape
	r1         float64
	r2         float64
	startAngle float64
	endAngle   float64
}

func (a *StdArch) DrawContent(w Writer) error {
	a.applyBorderAndFill(w)
	x, y := a.center()
	if err := w.Arch(x, y, a.outerRadius(), a.innerRadius(), a.startAngle, a.endAngle, a.border != nil, a.fill != nil, a.reverse); err != nil {
		return err
	}
	return a.drawChildren(w)
}

func (a *StdArch) PreferredHeight(Writer) float64 {
	if a.height != 0 {
		return a.height
	}
	if a.r1 != 0 {
		return a.r1*2 + NonContentHeight(a)
	}
	return NonContentHeight(a)
}

func (a *StdArch) PreferredWidth(Writer) float64 {
	if a.width != 0 {
		return a.width
	}
	if a.r1 != 0 {
		return a.r1*2 + NonContentWidth(a)
	}
	return NonContentWidth(a)
}

func (a *StdArch) SetAttrs(attrs map[string]string) {
	a.StdShape.SetAttrs(attrs)
	if r1, ok := attrs["r1"]; ok {
		a.r1 = ParseMeasurement(r1, a.Units())
	}
	if r2, ok := attrs["r2"]; ok {
		a.r2 = ParseMeasurement(r2, a.Units())
	}
	if startAngle, ok := attrs["start_angle"]; ok {
		a.startAngle, _ = strconv.ParseFloat(startAngle, 64)
	}
	if endAngle, ok := attrs["end_angle"]; ok {
		a.endAngle, _ = strconv.ParseFloat(endAngle, 64)
	}
}

func (a *StdArch) outerRadius() float64 {
	if a.r1 != 0 {
		return a.r1
	}
	return a.StdShape.radius()
}

func (a *StdArch) innerRadius() float64 {
	if a.r2 != 0 {
		return a.r2
	}
	return a.outerRadius() * 0.5
}

func init() {
	registerTag(DefaultSpace, "circle", func() interface{} { return &StdCircle{} })
	registerTag(DefaultSpace, "ellipse", func() interface{} { return &StdEllipse{} })
	registerTag(DefaultSpace, "polygon", func() interface{} { return &StdPolygon{} })
	registerTag(DefaultSpace, "star", func() interface{} { return &StdStar{} })
	registerTag(DefaultSpace, "arc", func() interface{} { return &StdArc{} })
	registerTag(DefaultSpace, "pie", func() interface{} { return &StdPie{} })
	registerTag(DefaultSpace, "arch", func() interface{} { return &StdArch{} })
}

func (c *StdCircle) String() string  { return fmt.Sprintf("StdCircle %s", &c.StdShape) }
func (e *StdEllipse) String() string { return fmt.Sprintf("StdEllipse %s", &e.StdShape) }
func (p *StdPolygon) String() string { return fmt.Sprintf("StdPolygon %s", &p.StdShape) }
func (s *StdStar) String() string    { return fmt.Sprintf("StdStar %s", &s.StdShape) }
func (a *StdArc) String() string     { return fmt.Sprintf("StdArc %s", &a.StdShape) }
func (p *StdPie) String() string     { return fmt.Sprintf("StdPie %s", &p.StdArc) }
func (a *StdArch) String() string    { return fmt.Sprintf("StdArch %s", &a.StdShape) }

var _ Container = (*StdCircle)(nil)
var _ Container = (*StdEllipse)(nil)
var _ Container = (*StdPolygon)(nil)
var _ Container = (*StdStar)(nil)
var _ Container = (*StdArc)(nil)
var _ Container = (*StdPie)(nil)
var _ Container = (*StdArch)(nil)
var _ HasAttrs = (*StdCircle)(nil)
var _ HasAttrs = (*StdEllipse)(nil)
var _ HasAttrs = (*StdPolygon)(nil)
var _ HasAttrs = (*StdStar)(nil)
var _ HasAttrs = (*StdArc)(nil)
var _ HasAttrs = (*StdPie)(nil)
var _ HasAttrs = (*StdArch)(nil)
var _ Identifier = (*StdCircle)(nil)
var _ Identifier = (*StdEllipse)(nil)
var _ Identifier = (*StdPolygon)(nil)
var _ Identifier = (*StdStar)(nil)
var _ Identifier = (*StdArc)(nil)
var _ Identifier = (*StdPie)(nil)
var _ Identifier = (*StdArch)(nil)
var _ Printer = (*StdCircle)(nil)
var _ Printer = (*StdEllipse)(nil)
var _ Printer = (*StdPolygon)(nil)
var _ Printer = (*StdStar)(nil)
var _ Printer = (*StdArc)(nil)
var _ Printer = (*StdPie)(nil)
var _ Printer = (*StdArch)(nil)
var _ WantsContainer = (*StdCircle)(nil)
var _ WantsContainer = (*StdEllipse)(nil)
var _ WantsContainer = (*StdPolygon)(nil)
var _ WantsContainer = (*StdStar)(nil)
var _ WantsContainer = (*StdArc)(nil)
var _ WantsContainer = (*StdPie)(nil)
var _ WantsContainer = (*StdArch)(nil)
