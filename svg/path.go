package svg

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"
)

type pathScanner struct {
	text string
	pos  int
}

func parsePathData(text string) ([]Segment, error) {
	s := &pathScanner{text: text}
	var segments []Segment
	var cmd rune
	cur := Point{}
	start := Point{}
	lastCubicControl := Point{}
	lastQuadraticControl := Point{}
	hasCubicControl := false
	hasQuadraticControl := false

	for {
		s.skipSpace()
		if s.done() {
			break
		}
		if r := rune(s.peek()); unicode.IsLetter(r) {
			cmd = r
			s.pos++
		} else if cmd == 0 {
			return nil, fmt.Errorf("path missing command")
		}
		switch cmd {
		case 'M', 'm':
			points, err := s.readPointSequence()
			if err != nil {
				return nil, err
			}
			for i, pt := range points {
				if cmd == 'm' {
					pt = Point{X: cur.X + pt.X, Y: cur.Y + pt.Y}
				}
				if i == 0 {
					segments = append(segments, Segment{Type: SegmentMoveTo, Points: []Point{pt}})
					start = pt
				} else {
					segments = append(segments, Segment{Type: SegmentLineTo, Points: []Point{pt}})
				}
				cur = pt
			}
			if cmd == 'm' {
				cmd = 'l'
			} else {
				cmd = 'L'
			}
			hasCubicControl = false
			hasQuadraticControl = false
		case 'L', 'l':
			points, err := s.readPointSequence()
			if err != nil {
				return nil, err
			}
			for _, pt := range points {
				if cmd == 'l' {
					pt = Point{X: cur.X + pt.X, Y: cur.Y + pt.Y}
				}
				segments = append(segments, Segment{Type: SegmentLineTo, Points: []Point{pt}})
				cur = pt
			}
			hasCubicControl = false
			hasQuadraticControl = false
		case 'H', 'h':
			values, err := s.readNumbers()
			if err != nil {
				return nil, err
			}
			for _, x := range values {
				if cmd == 'h' {
					x += cur.X
				}
				cur = Point{X: x, Y: cur.Y}
				segments = append(segments, Segment{Type: SegmentLineTo, Points: []Point{cur}})
			}
			hasCubicControl = false
			hasQuadraticControl = false
		case 'V', 'v':
			values, err := s.readNumbers()
			if err != nil {
				return nil, err
			}
			for _, y := range values {
				if cmd == 'v' {
					y += cur.Y
				}
				cur = Point{X: cur.X, Y: y}
				segments = append(segments, Segment{Type: SegmentLineTo, Points: []Point{cur}})
			}
			hasCubicControl = false
			hasQuadraticControl = false
		case 'C', 'c':
			for s.moreNumbers() {
				p1, err := s.readPoint()
				if err != nil {
					return nil, err
				}
				p2, err := s.readPoint()
				if err != nil {
					return nil, err
				}
				p3, err := s.readPoint()
				if err != nil {
					return nil, err
				}
				if cmd == 'c' {
					p1 = translatePoint(cur, p1)
					p2 = translatePoint(cur, p2)
					p3 = translatePoint(cur, p3)
				}
				segments = append(segments, Segment{Type: SegmentCubicTo, Points: []Point{p1, p2, p3}})
				cur = p3
				lastCubicControl = p2
				hasCubicControl = true
				hasQuadraticControl = false
			}
		case 'S', 's':
			for s.moreNumbers() {
				p2, err := s.readPoint()
				if err != nil {
					return nil, err
				}
				p3, err := s.readPoint()
				if err != nil {
					return nil, err
				}
				p1 := cur
				if hasCubicControl {
					p1 = reflectPoint(lastCubicControl, cur)
				}
				if cmd == 's' {
					p2 = translatePoint(cur, p2)
					p3 = translatePoint(cur, p3)
				}
				segments = append(segments, Segment{Type: SegmentCubicTo, Points: []Point{p1, p2, p3}})
				cur = p3
				lastCubicControl = p2
				hasCubicControl = true
				hasQuadraticControl = false
			}
		case 'Q', 'q':
			for s.moreNumbers() {
				qc, err := s.readPoint()
				if err != nil {
					return nil, err
				}
				p3, err := s.readPoint()
				if err != nil {
					return nil, err
				}
				if cmd == 'q' {
					qc = translatePoint(cur, qc)
					p3 = translatePoint(cur, p3)
				}
				p1, p2 := quadraticToCubic(cur, qc, p3)
				segments = append(segments, Segment{Type: SegmentCubicTo, Points: []Point{p1, p2, p3}})
				cur = p3
				lastQuadraticControl = qc
				hasQuadraticControl = true
				hasCubicControl = false
			}
		case 'T', 't':
			for s.moreNumbers() {
				p3, err := s.readPoint()
				if err != nil {
					return nil, err
				}
				qc := cur
				if hasQuadraticControl {
					qc = reflectPoint(lastQuadraticControl, cur)
				}
				if cmd == 't' {
					p3 = translatePoint(cur, p3)
				}
				p1, p2 := quadraticToCubic(cur, qc, p3)
				segments = append(segments, Segment{Type: SegmentCubicTo, Points: []Point{p1, p2, p3}})
				cur = p3
				lastQuadraticControl = qc
				hasQuadraticControl = true
				hasCubicControl = false
			}
		case 'A', 'a':
			for s.moreNumbers() {
				rx, err := s.readNumber()
				if err != nil {
					return nil, err
				}
				ry, err := s.readNumber()
				if err != nil {
					return nil, err
				}
				rot, err := s.readNumber()
				if err != nil {
					return nil, err
				}
				large, err := s.readFlag()
				if err != nil {
					return nil, err
				}
				sweep, err := s.readFlag()
				if err != nil {
					return nil, err
				}
				end, err := s.readPoint()
				if err != nil {
					return nil, err
				}
				if cmd == 'a' {
					end = translatePoint(cur, end)
				}
				curves := arcToCubic(cur, end, rx, ry, rot, large, sweep)
				for _, curve := range curves {
					segments = append(segments, Segment{Type: SegmentCubicTo, Points: []Point{curve[0], curve[1], curve[2]}})
					cur = curve[2]
					lastCubicControl = curve[1]
				}
				hasCubicControl = len(curves) > 0
				hasQuadraticControl = false
			}
		case 'Z', 'z':
			segments = append(segments, Segment{Type: SegmentClosePath})
			cur = start
			hasCubicControl = false
			hasQuadraticControl = false
		default:
			return nil, fmt.Errorf("unsupported path command %q", string(cmd))
		}
	}
	return segments, nil
}

func (s *pathScanner) done() bool {
	return s.pos >= len(s.text)
}

func (s *pathScanner) peek() byte {
	return s.text[s.pos]
}

func (s *pathScanner) skipSpace() {
	for !s.done() {
		switch s.text[s.pos] {
		case ' ', '\n', '\t', '\r', ',':
			s.pos++
		default:
			return
		}
	}
}

func (s *pathScanner) moreNumbers() bool {
	s.skipSpace()
	if s.done() {
		return false
	}
	return !unicode.IsLetter(rune(s.peek()))
}

func (s *pathScanner) readNumber() (float64, error) {
	s.skipSpace()
	start := s.pos
	if !s.done() && (s.text[s.pos] == '+' || s.text[s.pos] == '-') {
		s.pos++
	}
	dot := false
	exp := false
	for !s.done() {
		ch := s.text[s.pos]
		switch {
		case ch >= '0' && ch <= '9':
			s.pos++
		case ch == '.' && !dot:
			dot = true
			s.pos++
		case (ch == 'e' || ch == 'E') && !exp:
			exp = true
			s.pos++
			if !s.done() && (s.text[s.pos] == '+' || s.text[s.pos] == '-') {
				s.pos++
			}
		default:
			value, err := strconv.ParseFloat(s.text[start:s.pos], 64)
			if err != nil {
				return 0, err
			}
			return value, nil
		}
	}
	value, err := strconv.ParseFloat(s.text[start:s.pos], 64)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func (s *pathScanner) readNumbers() ([]float64, error) {
	values := []float64{}
	for s.moreNumbers() {
		value, err := s.readNumber()
		if err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	return values, nil
}

func (s *pathScanner) readPoint() (Point, error) {
	x, err := s.readNumber()
	if err != nil {
		return Point{}, err
	}
	y, err := s.readNumber()
	if err != nil {
		return Point{}, err
	}
	return Point{X: x, Y: y}, nil
}

func (s *pathScanner) readPointSequence() ([]Point, error) {
	points := []Point{}
	for s.moreNumbers() {
		point, err := s.readPoint()
		if err != nil {
			return nil, err
		}
		points = append(points, point)
	}
	if len(points) == 0 {
		return nil, fmt.Errorf("expected point sequence")
	}
	return points, nil
}

func (s *pathScanner) readFlag() (bool, error) {
	value, err := s.readNumber()
	if err != nil {
		return false, err
	}
	return value != 0, nil
}

func translatePoint(base, delta Point) Point {
	return Point{X: base.X + delta.X, Y: base.Y + delta.Y}
}

func reflectPoint(control, around Point) Point {
	return Point{X: (2 * around.X) - control.X, Y: (2 * around.Y) - control.Y}
}

func quadraticToCubic(start, control, end Point) (Point, Point) {
	return Point{
			X: start.X + ((2.0 / 3.0) * (control.X - start.X)),
			Y: start.Y + ((2.0 / 3.0) * (control.Y - start.Y)),
		}, Point{
			X: end.X + ((2.0 / 3.0) * (control.X - end.X)),
			Y: end.Y + ((2.0 / 3.0) * (control.Y - end.Y)),
		}
}

func arcToCubic(start, end Point, rx, ry, xAxisRotation float64, largeArc, sweep bool) [][3]Point {
	if rx == 0 || ry == 0 || (start == end) {
		return [][3]Point{{start, end, end}}[:0]
	}
	phi := xAxisRotation * math.Pi / 180.0
	cphi := math.Cos(phi)
	sphi := math.Sin(phi)

	dx2 := (start.X - end.X) / 2.0
	dy2 := (start.Y - end.Y) / 2.0
	x1p := (cphi * dx2) + (sphi * dy2)
	y1p := (-sphi * dx2) + (cphi * dy2)

	rx = math.Abs(rx)
	ry = math.Abs(ry)
	lambda := (x1p*x1p)/(rx*rx) + (y1p*y1p)/(ry*ry)
	if lambda > 1 {
		scale := math.Sqrt(lambda)
		rx *= scale
		ry *= scale
	}

	sign := -1.0
	if largeArc != sweep {
		sign = 1.0
	}
	num := (rx * rx * ry * ry) - (rx*rx*y1p*y1p) - (ry*ry*x1p*x1p)
	den := (rx*rx*y1p*y1p + ry*ry*x1p*x1p)
	if den == 0 {
		return nil
	}
	coef := sign * math.Sqrt(math.Max(0, num/den))
	cxp := coef * ((rx * y1p) / ry)
	cyp := coef * (-(ry * x1p) / rx)

	cx := (cphi * cxp) - (sphi * cyp) + ((start.X + end.X) / 2.0)
	cy := (sphi * cxp) + (cphi * cyp) + ((start.Y + end.Y) / 2.0)

	theta1 := angleBetween(Point{X: 1, Y: 0}, Point{X: (x1p - cxp) / rx, Y: (y1p - cyp) / ry})
	delta := angleBetween(
		Point{X: (x1p - cxp) / rx, Y: (y1p - cyp) / ry},
		Point{X: (-x1p - cxp) / rx, Y: (-y1p - cyp) / ry},
	)
	if !sweep && delta > 0 {
		delta -= 2 * math.Pi
	} else if sweep && delta < 0 {
		delta += 2 * math.Pi
	}

	segments := int(math.Ceil(math.Abs(delta) / (math.Pi / 2)))
	step := delta / float64(segments)
	cur := theta1
	curves := make([][3]Point, 0, segments)
	for i := 0; i < segments; i++ {
		next := cur + step
		curves = append(curves, approximateArcSegment(cx, cy, rx, ry, phi, cur, next))
		cur = next
	}
	return curves
}

func angleBetween(u, v Point) float64 {
	sign := 1.0
	if (u.X*v.Y - u.Y*v.X) < 0 {
		sign = -1
	}
	dot := (u.X * v.X) + (u.Y * v.Y)
	den := math.Sqrt((u.X*u.X)+(u.Y*u.Y)) * math.Sqrt((v.X*v.X)+(v.Y*v.Y))
	if den == 0 {
		return 0
	}
	value := dot / den
	if value < -1 {
		value = -1
	} else if value > 1 {
		value = 1
	}
	return sign * math.Acos(value)
}

func approximateArcSegment(cx, cy, rx, ry, phi, start, end float64) [3]Point {
	alpha := (4.0 / 3.0) * math.Tan((end-start)/4.0)
	x1 := math.Cos(start)
	y1 := math.Sin(start)
	x2 := math.Cos(end)
	y2 := math.Sin(end)

	p1 := Point{X: x1 - (alpha * y1), Y: y1 + (alpha * x1)}
	p2 := Point{X: x2 + (alpha * y2), Y: y2 - (alpha * x2)}
	p := Point{X: x2, Y: y2}

	return [3]Point{
		mapArcPoint(cx, cy, rx, ry, phi, p1),
		mapArcPoint(cx, cy, rx, ry, phi, p2),
		mapArcPoint(cx, cy, rx, ry, phi, p),
	}
}

func mapArcPoint(cx, cy, rx, ry, phi float64, p Point) Point {
	cphi := math.Cos(phi)
	sphi := math.Sin(phi)
	return Point{
		X: cx + (rx*p.X*cphi) - (ry*p.Y*sphi),
		Y: cy + (rx*p.X*sphi) + (ry*p.Y*cphi),
	}
}

func applyTransformToSegments(segments []Segment, t Transform) []Segment {
	out := make([]Segment, len(segments))
	for i, seg := range segments {
		out[i].Type = seg.Type
		if len(seg.Points) == 0 {
			continue
		}
		out[i].Points = make([]Point, len(seg.Points))
		for j, p := range seg.Points {
			out[i].Points[j] = t.Apply(p)
		}
	}
	return out
}

func ellipsePath(cx, cy, rx, ry float64) []Segment {
	const kappa = 0.5522847498307936
	return []Segment{
		{Type: SegmentMoveTo, Points: []Point{{X: cx + rx, Y: cy}}},
		{Type: SegmentCubicTo, Points: []Point{{X: cx + rx, Y: cy + kappa*ry}, {X: cx + kappa*rx, Y: cy + ry}, {X: cx, Y: cy + ry}}},
		{Type: SegmentCubicTo, Points: []Point{{X: cx - kappa*rx, Y: cy + ry}, {X: cx - rx, Y: cy + kappa*ry}, {X: cx - rx, Y: cy}}},
		{Type: SegmentCubicTo, Points: []Point{{X: cx - rx, Y: cy - kappa*ry}, {X: cx - kappa*rx, Y: cy - ry}, {X: cx, Y: cy - ry}}},
		{Type: SegmentCubicTo, Points: []Point{{X: cx + kappa*rx, Y: cy - ry}, {X: cx + rx, Y: cy - kappa*ry}, {X: cx + rx, Y: cy}}},
		{Type: SegmentClosePath},
	}
}

func rectPath(x, y, width, height, rx, ry float64) []Segment {
	if rx <= 0 && ry <= 0 {
		return []Segment{
			{Type: SegmentMoveTo, Points: []Point{{X: x, Y: y}}},
			{Type: SegmentLineTo, Points: []Point{{X: x + width, Y: y}}},
			{Type: SegmentLineTo, Points: []Point{{X: x + width, Y: y + height}}},
			{Type: SegmentLineTo, Points: []Point{{X: x, Y: y + height}}},
			{Type: SegmentClosePath},
		}
	}
	if rx <= 0 {
		rx = ry
	}
	if ry <= 0 {
		ry = rx
	}
	if rx > width/2 {
		rx = width / 2
	}
	if ry > height/2 {
		ry = height / 2
	}
	kx := rx * 0.5522847498307936
	ky := ry * 0.5522847498307936
	return []Segment{
		{Type: SegmentMoveTo, Points: []Point{{X: x + rx, Y: y}}},
		{Type: SegmentLineTo, Points: []Point{{X: x + width - rx, Y: y}}},
		{Type: SegmentCubicTo, Points: []Point{{X: x + width - rx + kx, Y: y}, {X: x + width, Y: y + ry - ky}, {X: x + width, Y: y + ry}}},
		{Type: SegmentLineTo, Points: []Point{{X: x + width, Y: y + height - ry}}},
		{Type: SegmentCubicTo, Points: []Point{{X: x + width, Y: y + height - ry + ky}, {X: x + width - rx + kx, Y: y + height}, {X: x + width - rx, Y: y + height}}},
		{Type: SegmentLineTo, Points: []Point{{X: x + rx, Y: y + height}}},
		{Type: SegmentCubicTo, Points: []Point{{X: x + rx - kx, Y: y + height}, {X: x, Y: y + height - ry + ky}, {X: x, Y: y + height - ry}}},
		{Type: SegmentLineTo, Points: []Point{{X: x, Y: y + ry}}},
		{Type: SegmentCubicTo, Points: []Point{{X: x, Y: y + ry - ky}, {X: x + rx - kx, Y: y}, {X: x + rx, Y: y}}},
		{Type: SegmentClosePath},
	}
}

func openPolylinePath(points []Point, close bool) []Segment {
	if len(points) == 0 {
		return nil
	}
	segments := []Segment{{Type: SegmentMoveTo, Points: []Point{points[0]}}}
	for _, point := range points[1:] {
		segments = append(segments, Segment{Type: SegmentLineTo, Points: []Point{point}})
	}
	if close {
		segments = append(segments, Segment{Type: SegmentClosePath})
	}
	return segments
}

func parseDashArray(text string) ([]float64, error) {
	text = strings.TrimSpace(text)
	if text == "" || text == "none" {
		return nil, nil
	}
	values, err := parseFloatList(text)
	if err != nil {
		return nil, err
	}
	return values, nil
}
