package svg

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

type Transform struct {
	A float64
	B float64
	C float64
	D float64
	E float64
	F float64
}

func IdentityTransform() Transform {
	return Transform{A: 1, D: 1}
}

func (t Transform) Mul(other Transform) Transform {
	return Transform{
		A: (t.A * other.A) + (t.C * other.B),
		B: (t.B * other.A) + (t.D * other.B),
		C: (t.A * other.C) + (t.C * other.D),
		D: (t.B * other.C) + (t.D * other.D),
		E: (t.A * other.E) + (t.C * other.F) + t.E,
		F: (t.B * other.E) + (t.D * other.F) + t.F,
	}
}

func (t Transform) Apply(p Point) Point {
	return Point{
		X: (t.A * p.X) + (t.C * p.Y) + t.E,
		Y: (t.B * p.X) + (t.D * p.Y) + t.F,
	}
}

func parseTransform(text string) (Transform, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return IdentityTransform(), nil
	}
	result := IdentityTransform()
	for text != "" {
		open := strings.IndexByte(text, '(')
		close := strings.IndexByte(text, ')')
		if open <= 0 || close < open {
			return IdentityTransform(), fmt.Errorf("bad transform %q", text)
		}
		name := strings.TrimSpace(text[:open])
		args, err := parseFloatList(text[open+1 : close])
		if err != nil {
			return IdentityTransform(), err
		}
		next, err := transformFromParts(name, args)
		if err != nil {
			return IdentityTransform(), err
		}
		result = result.Mul(next)
		text = strings.TrimSpace(text[close+1:])
	}
	return result, nil
}

func transformFromParts(name string, args []float64) (Transform, error) {
	switch name {
	case "matrix":
		if len(args) != 6 {
			return IdentityTransform(), fmt.Errorf("matrix requires 6 args")
		}
		return Transform{A: args[0], B: args[1], C: args[2], D: args[3], E: args[4], F: args[5]}, nil
	case "translate":
		if len(args) == 1 {
			return Transform{A: 1, D: 1, E: args[0]}, nil
		}
		if len(args) == 2 {
			return Transform{A: 1, D: 1, E: args[0], F: args[1]}, nil
		}
	case "scale":
		if len(args) == 1 {
			return Transform{A: args[0], D: args[0]}, nil
		}
		if len(args) == 2 {
			return Transform{A: args[0], D: args[1]}, nil
		}
	case "rotate":
		if len(args) == 1 {
			theta := args[0] * math.Pi / 180.0
			c, s := math.Cos(theta), math.Sin(theta)
			return Transform{A: c, B: s, C: -s, D: c}, nil
		}
		if len(args) == 3 {
			theta := args[0] * math.Pi / 180.0
			c, s := math.Cos(theta), math.Sin(theta)
			cx, cy := args[1], args[2]
			return Transform{A: c, B: s, C: -s, D: c, E: cx - (c * cx) + (s * cy), F: cy - (s * cx) - (c * cy)}, nil
		}
	case "skewX":
		if len(args) == 1 {
			return Transform{A: 1, C: math.Tan(args[0] * math.Pi / 180.0), D: 1}, nil
		}
	case "skewY":
		if len(args) == 1 {
			return Transform{A: 1, B: math.Tan(args[0] * math.Pi / 180.0), D: 1}, nil
		}
	}
	return IdentityTransform(), fmt.Errorf("unsupported transform %q", name)
}

func parseFloatList(text string) ([]float64, error) {
	replacer := strings.NewReplacer(",", " ", "\n", " ", "\t", " ")
	fields := strings.Fields(replacer.Replace(text))
	values := make([]float64, 0, len(fields))
	for _, field := range fields {
		value, err := strconv.ParseFloat(field, 64)
		if err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	return values, nil
}
