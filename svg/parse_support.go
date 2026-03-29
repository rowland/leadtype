package svg

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/rowland/leadtype/colors"
)

func parseColor(text string) (Paint, error) {
	text = strings.TrimSpace(strings.ToLower(text))
	if text == "" {
		return Paint{}, fmt.Errorf("empty color")
	}
	if text == "none" {
		return Paint{Set: true, None: true}, nil
	}
	if strings.HasPrefix(text, "#") {
		switch len(text) {
		case 4:
			r, err := strconv.ParseUint(strings.Repeat(string(text[1]), 2), 16, 8)
			if err != nil {
				return Paint{}, err
			}
			g, err := strconv.ParseUint(strings.Repeat(string(text[2]), 2), 16, 8)
			if err != nil {
				return Paint{}, err
			}
			b, err := strconv.ParseUint(strings.Repeat(string(text[3]), 2), 16, 8)
			if err != nil {
				return Paint{}, err
			}
			return Paint{Set: true, Color: colors.Color((r << 16) | (g << 8) | b)}, nil
		case 7:
			value, err := strconv.ParseUint(text[1:], 16, 32)
			if err != nil {
				return Paint{}, err
			}
			return Paint{Set: true, Color: colors.Color(value)}, nil
		}
	}
	if strings.HasPrefix(text, "rgb(") && strings.HasSuffix(text, ")") {
		body := strings.TrimSuffix(strings.TrimPrefix(text, "rgb("), ")")
		parts := strings.Split(body, ",")
		if len(parts) != 3 {
			return Paint{}, fmt.Errorf("rgb color requires 3 channels")
		}
		var rgb [3]uint64
		for i, part := range parts {
			part = strings.TrimSpace(part)
			if strings.HasSuffix(part, "%") {
				value, err := strconv.ParseFloat(strings.TrimSuffix(part, "%"), 64)
				if err != nil {
					return Paint{}, err
				}
				rgb[i] = uint64((value * 255.0 / 100.0) + 0.5)
			} else {
				value, err := strconv.ParseUint(part, 10, 8)
				if err != nil {
					return Paint{}, err
				}
				rgb[i] = value
			}
		}
		return Paint{Set: true, Color: colors.Color((rgb[0] << 16) | (rgb[1] << 8) | rgb[2])}, nil
	}
	if named, err := colors.NamedColor(text); err == nil {
		return Paint{Set: true, Color: named}, nil
	}
	return Paint{}, fmt.Errorf("unsupported color %q", text)
}

func parseNumber(text string) (float64, error) {
	return strconv.ParseFloat(strings.TrimSpace(text), 64)
}

func parseLength(text string, relative float64) (float64, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return 0, fmt.Errorf("empty length")
	}
	if strings.HasSuffix(text, "%") {
		value, err := strconv.ParseFloat(strings.TrimSuffix(text, "%"), 64)
		if err != nil {
			return 0, err
		}
		return relative * value / 100.0, nil
	}
	for _, unit := range []string{"px", "pt", "in", "cm", "mm"} {
		if strings.HasSuffix(text, unit) {
			value, err := strconv.ParseFloat(strings.TrimSpace(strings.TrimSuffix(text, unit)), 64)
			if err != nil {
				return 0, err
			}
			switch unit {
			case "px", "pt":
				return value, nil
			case "in":
				return value * 72.0, nil
			case "cm":
				return value * 72.0 / 2.54, nil
			case "mm":
				return value * 72.0 / 25.4, nil
			}
		}
	}
	return strconv.ParseFloat(text, 64)
}

func parsePoints(text string) ([]Point, error) {
	values, err := parseFloatList(strings.NewReplacer(",", " ", "\n", " ", "\t", " ").Replace(text))
	if err != nil {
		return nil, err
	}
	if len(values)%2 != 0 {
		return nil, fmt.Errorf("points must have even coordinate count")
	}
	points := make([]Point, 0, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		points = append(points, Point{X: values[i], Y: values[i+1]})
	}
	return points, nil
}
