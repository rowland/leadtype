package ltml

import "testing"

func TestNamedColorExpandsShortHex(t *testing.T) {
	tests := map[string]string{
		"#123": "#112233",
		"#c3a": "#cc33aa",
		"#AbF": "#AAbbFF",
	}
	for short, full := range tests {
		t.Run(short, func(t *testing.T) {
			if got, want := NamedColor(short), NamedColor(full); got != want {
				t.Fatalf("NamedColor(%q) = %#v, want %#v", short, got, want)
			}
		})
	}
}

func TestNamedColorShortHexAppliesToStylesAndGradientStops(t *testing.T) {
	var pen PenStyle
	pen.SetAttrs(map[string]string{"color": "#c33"})
	if got, want := pen.color, NamedColor("#cc3333"); got != want {
		t.Fatalf("pen color = %#v, want %#v", got, want)
	}

	var brush BrushStyle
	brush.SetAttrs(map[string]string{"color": "#fea"})
	if got, want := brush.color, NamedColor("#ffeeaa"); got != want {
		t.Fatalf("brush color = %#v, want %#v", got, want)
	}

	var font FontStyle
	font.SetAttrs(map[string]string{"color": "#08f"})
	if got, want := font.color, NamedColor("#0088ff"); got != want {
		t.Fatalf("font color = %#v, want %#v", got, want)
	}

	brush.SetAttrs(map[string]string{"kind": "linear-gradient", "stops": "0:#123,1:#AbF"})
	if brush.linearGradient == nil || len(brush.linearGradient.Stops) != 2 {
		t.Fatalf("gradient stops = %#v, want two stops", brush.linearGradient)
	}
	if got, want := brush.linearGradient.Stops[0].Color, NamedColor("#112233"); got != want {
		t.Fatalf("first stop = %#v, want %#v", got, want)
	}
	if got, want := brush.linearGradient.Stops[1].Color, NamedColor("#aabbff"); got != want {
		t.Fatalf("second stop = %#v, want %#v", got, want)
	}
}
