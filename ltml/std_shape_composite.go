package ltml

import (
	"slices"

	"github.com/rowland/leadtype/pdf"
)

type shapePathContributor interface {
	shape() *StdShape
	appendShapePath(w Writer) error
}

type compositeShapeWriter struct {
	Writer
	contributors map[*StdShape]struct{}
}

func newCompositeShapeWriter(w Writer) *compositeShapeWriter {
	return &compositeShapeWriter{
		Writer:       w,
		contributors: make(map[*StdShape]struct{}),
	}
}

func (w *compositeShapeWriter) markContributed(s *StdShape) {
	w.contributors[s] = struct{}{}
}

func (w *compositeShapeWriter) contributed(s *StdShape) bool {
	_, ok := w.contributors[s]
	return ok
}

func (s *StdShape) shape() *StdShape {
	return s
}

func (s *StdShape) hasShapeChild() bool {
	return s.findShapeChild(s.Widgets())
}

func (s *StdShape) findShapeChild(widgets []Widget) bool {
	for _, child := range widgets {
		if !child.Visible() || child.Disabled() {
			continue
		}
		if _, ok := child.(shapePathContributor); ok {
			return true
		}
		if container, ok := child.(Container); ok {
			if s.findShapeChild(container.Widgets()) {
				return true
			}
		}
	}
	return false
}

func (s *StdShape) needsCompositePath(border, fill bool) bool {
	return (border || fill) && s.hasShapeChild()
}

func (s *StdShape) finalizeShapePath(w Writer, border, fill bool) error {
	switch {
	case border && fill:
		return w.FillAndStroke()
	case fill:
		return w.Fill()
	case border:
		return w.Stroke()
	default:
		return nil
	}
}

func (s *StdShape) appendDirectShapeChildPaths(w Writer, widgets []Widget) error {
	children := slices.Clone(widgets)
	slices.SortStableFunc(children, func(a, b Widget) int {
		return a.ZIndex() - b.ZIndex()
	})
	cw, _ := w.(*compositeShapeWriter)
	for _, child := range children {
		if !child.Visible() || child.Disabled() {
			continue
		}
		contributor, ok := child.(shapePathContributor)
		if !ok {
			continue
		}
		if err := contributor.appendShapePath(w); err != nil {
			return err
		}
		if cw != nil {
			cw.markContributed(contributor.shape())
		}
	}
	return nil
}

func (s *StdShape) runCompositeShapePath(w Writer, appendPath func(Writer) error, border, fill bool) error {
	composite := newCompositeShapeWriter(w)
	var pathErr, finalizeErr error
	if err := w.Path(func() {
		pathErr = appendPath(composite)
		if pathErr != nil {
			return
		}
		pathErr = s.appendDirectShapeChildPaths(composite, s.Widgets())
		if pathErr != nil {
			return
		}
		finalizeErr = s.finalizeShapePath(composite, border, fill)
	}); err != nil {
		return err
	}
	if pathErr != nil {
		return pathErr
	}
	if finalizeErr != nil {
		return finalizeErr
	}
	return s.drawChildren(composite)
}

func (s *StdShape) drawShapeContent(w Writer, widget *StdWidget, appendPath func(Writer) error, drawImmediate func(Writer, bool, bool) error) error {
	return withGraphicAccessibility(w, widget, "Figure", func() error {
		border := s.border != nil
		fill := s.fill != nil

		if cw, ok := w.(*compositeShapeWriter); ok && cw.contributed(s) {
			if s.needsCompositePath(border, fill) {
				if err := s.applyBorderAndFill(w); err != nil {
					return err
				}
				return s.runCompositeShapePath(w, appendPath, border, fill)
			}
			if border || fill {
				if err := s.applyBorderAndFill(w); err != nil {
					return err
				}
				if err := drawImmediate(w, border, fill); err != nil {
					return err
				}
			}
			return s.drawChildren(w)
		}

		if err := s.applyBorderAndFill(w); err != nil {
			return err
		}

		if s.needsCompositePath(border, fill) {
			return s.runCompositeShapePath(w, appendPath, border, fill)
		}

		if err := drawImmediate(w, border, fill); err != nil {
			return err
		}
		return s.drawChildren(w)
	})
}

func (c *StdCircle) appendShapePath(w Writer) error {
	x, y := c.center()
	return w.AppendClosedShapePath(pdf.ClosedShape{
		Kind:    pdf.ClosedShapeCircle,
		Center:  pdf.Location{X: x, Y: y},
		Radius:  c.radius(),
		Reverse: c.reverse,
	})
}

func (c *StdCircle) drawCircleImmediate(w Writer, border, fill bool) error {
	x, y := c.center()
	return w.Circle(x, y, c.radius(), border, fill, c.reverse)
}

func (e *StdEllipse) appendShapePath(w Writer) error {
	x, y := e.center()
	return w.AppendClosedShapePath(pdf.ClosedShape{
		Kind:    pdf.ClosedShapeEllipse,
		Center:  pdf.Location{X: x, Y: y},
		RadiusX: e.radiusX(),
		RadiusY: e.radiusY(),
		Reverse: e.reverse,
	})
}

func (e *StdEllipse) drawEllipseImmediate(w Writer, border, fill bool) error {
	x, y := e.center()
	return w.Ellipse(x, y, e.radiusX(), e.radiusY(), border, fill, e.reverse)
}

func (p *StdPolygon) appendShapePath(w Writer) error {
	x, y := p.center()
	return w.AppendClosedShapePath(pdf.ClosedShape{
		Kind:     pdf.ClosedShapePolygon,
		Center:   pdf.Location{X: x, Y: y},
		Radius:   p.radius(),
		Sides:    p.Sides(),
		Rotation: p.rotation,
		Reverse:  p.reverse,
	})
}

func (p *StdPolygon) drawPolygonImmediate(w Writer, border, fill bool) error {
	x, y := p.center()
	return w.Polygon(x, y, p.radius(), p.Sides(), border, fill, p.reverse, p.rotation)
}

func (s *StdStar) appendShapePath(w Writer) error {
	x, y := s.center()
	return w.AppendClosedShapePath(pdf.ClosedShape{
		Kind:        pdf.ClosedShapeStar,
		Center:      pdf.Location{X: x, Y: y},
		Radius:      s.outerRadius(),
		InnerRadius: s.innerRadius(),
		Points:      s.Points(),
		Rotation:    s.effectiveRotation(),
		Reverse:     s.reverse,
	})
}

func (s *StdStar) drawStarImmediate(w Writer, border, fill bool) error {
	x, y := s.center()
	return w.Star(x, y, s.outerRadius(), s.innerRadius(), s.Points(), border, fill, s.reverse, s.effectiveRotation())
}

func (p *StdPie) appendShapePath(w Writer) error {
	x, y := p.center()
	return w.AppendPiePath(x, y, p.radius(), p.startAngle, p.endAngle, p.reverse)
}

func (p *StdPie) drawPieImmediate(w Writer, border, fill bool) error {
	x, y := p.center()
	return w.Pie(x, y, p.radius(), p.startAngle, p.endAngle, border, fill, p.reverse)
}

func (a *StdArch) appendShapePath(w Writer) error {
	x, y := a.center()
	return w.AppendArchPath(x, y, a.outerRadius(), a.innerRadius(), a.startAngle, a.endAngle, a.reverse)
}

func (a *StdArch) drawArchImmediate(w Writer, border, fill bool) error {
	x, y := a.center()
	return w.Arch(x, y, a.outerRadius(), a.innerRadius(), a.startAngle, a.endAngle, border, fill, a.reverse)
}
