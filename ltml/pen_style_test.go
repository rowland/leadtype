package ltml

import (
	"testing"

	"github.com/rowland/leadtype/pdf"
)

func TestSetPenStyleAssignsNamedStyle(t *testing.T) {
	scope := &Scope{}
	base := &PenStyle{id: "accent", color: NamedColor("Gold"), pattern: "dashed", cap: "round_cap"}
	if err := scope.AddStyle(base); err != nil {
		t.Fatal(err)
	}

	var field *PenStyle
	SetPenStyle(&field, "border", map[string]string{"border": "accent"}, scope, "pt")

	if field != base {
		t.Fatalf("pen = %p, want named style %p", field, base)
	}
}

func TestSetPenStyleClonesNamedStyleBeforeOverrides(t *testing.T) {
	scope := &Scope{}
	base := &PenStyle{id: "accent", color: NamedColor("Gold"), pattern: "dashed", cap: "round_cap"}
	if err := scope.AddStyle(base); err != nil {
		t.Fatal(err)
	}

	var field *PenStyle
	SetPenStyle(&field, "border", map[string]string{
		"border":       "accent",
		"border.color": "Blue",
	}, scope, "pt")

	if field == base {
		t.Fatal("pen reused named style, want clone")
	}
	if field.color != NamedColor("Blue") {
		t.Fatalf("pen color = %v, want Blue", field.color)
	}
	if base.color != NamedColor("Gold") {
		t.Fatalf("named pen color = %v, want unchanged Gold", base.color)
	}
}

func TestSetPenStyleCreatesDefaultPenForArbitraryAttribute(t *testing.T) {
	var field *PenStyle
	SetPenStyle(&field, "stroke", map[string]string{
		"stroke.width": "2",
	}, nil, "mm")

	if field == nil {
		t.Fatal("pen is nil, want override-created pen")
	}
	if field.pattern != defaultPenPattern || field.cap != defaultPenCap {
		t.Fatalf("pen defaults = pattern %q cap %q", field.pattern, field.cap)
	}
	if got, want := field.width, FromUnits(2, "mm"); got != want {
		t.Fatalf("pen width = %v, want %v", got, want)
	}
}

func TestSetPenStylePrefixedUnitsOverrideWidgetUnits(t *testing.T) {
	var field *PenStyle
	SetPenStyle(&field, "border", map[string]string{
		"border.units": "cm",
		"border.width": "2",
	}, nil, "mm")

	if got, want := field.width, FromUnits(2, "cm"); got != want {
		t.Fatalf("pen width = %v, want %v", got, want)
	}
}

func TestSetPenStyleLeavesFieldUnchangedWithoutMatchingAttrs(t *testing.T) {
	base := &PenStyle{pattern: defaultPenPattern, cap: defaultPenCap}
	field := base

	SetPenStyle(&field, "border", map[string]string{"stroke": "Blue"}, nil, "pt")

	if field != base {
		t.Fatalf("pen = %p, want unchanged %p", field, base)
	}
}

func TestPenStyleSetAttrsLinearGradient(t *testing.T) {
	var ps PenStyle
	ps.SetAttrs(map[string]string{
		"kind":    "linear-gradient",
		"width":   "5pt",
		"pattern": "dashed",
		"cap":     "round_cap",
		"x0":      "0%",
		"y0":      "50%",
		"x1":      "100%",
		"y1":      "50%",
		"stops":   "0:#ef5148,1:#4f93ad",
	})

	if ps.Kind() != PenKindLinearGradient {
		t.Fatalf("kind = %q, want linear-gradient", ps.Kind())
	}
	if ps.width != 5 || ps.pattern != "dashed" || ps.Cap() != "round_cap" {
		t.Fatalf("pen basics = width %v pattern %q cap %q", ps.width, ps.pattern, ps.Cap())
	}
	if ps.linearGradient == nil || len(ps.linearGradient.Stops) != 2 {
		t.Fatalf("linear gradient = %#v, want two stops", ps.linearGradient)
	}
	if ps.linearPct == nil || ps.linearPct.X0 == nil || *ps.linearPct.X0 != 0 || ps.linearPct.Y0 == nil || *ps.linearPct.Y0 != 50 ||
		ps.linearPct.X1 == nil || *ps.linearPct.X1 != 100 || ps.linearPct.Y1 == nil || *ps.linearPct.Y1 != 50 {
		t.Fatalf("linear pct = %#v, want 0/50/100/50", ps.linearPct)
	}
}

func TestPenStyleSetAttrsRadialGradient(t *testing.T) {
	var ps PenStyle
	ps.SetAttrs(map[string]string{
		"kind":  "radial-gradient",
		"x0":    "50%",
		"y0":    "50%",
		"r0":    "0%",
		"x1":    "50%",
		"y1":    "50%",
		"r1":    "60%",
		"stops": "0:White,1:Black",
	})

	if ps.Kind() != PenKindRadialGradient {
		t.Fatalf("kind = %q, want radial-gradient", ps.Kind())
	}
	if ps.radialGradient == nil || len(ps.radialGradient.Stops) != 2 {
		t.Fatalf("radial gradient = %#v, want two stops", ps.radialGradient)
	}
	if ps.radialPct == nil || ps.radialPct.R0 == nil || *ps.radialPct.R0 != 0 || ps.radialPct.R1 == nil || *ps.radialPct.R1 != 60 {
		t.Fatalf("radial pct = %#v, want r0=0 r1=60", ps.radialPct)
	}
}

func TestPenStyleCloneDeepCopiesGradientState(t *testing.T) {
	original := &PenStyle{}
	original.SetAttrs(map[string]string{
		"kind":  "radial-gradient",
		"x0":    "1",
		"y0":    "2",
		"r0":    "3",
		"x1":    "4",
		"y1":    "5",
		"r1":    "60%",
		"stops": "0:#111111,1:#999999",
	})

	clone := original.Clone()
	clone.radialGradient.Stops[0].Position = 0.25
	*clone.radialPct.R1 = 75

	if original.radialGradient.Stops[0].Position != 0 {
		t.Fatalf("original stop position mutated to %v", original.radialGradient.Stops[0].Position)
	}
	if *original.radialPct.R1 != 60 {
		t.Fatalf("original radial r1 pct mutated to %v", *original.radialPct.R1)
	}
}

func TestPenStyleApplySolidClearsLineGradient(t *testing.T) {
	style := &PenStyle{id: "solid", color: NamedColor("Tomato"), width: 2, pattern: "solid"}
	w := &labelTestWriter{}

	style.Apply(w)

	if w.lineClears != 1 {
		t.Fatalf("line gradient clears = %d, want 1", w.lineClears)
	}
	if len(w.lineColors) != 1 || w.lineColors[0] != NamedColor("Tomato") {
		t.Fatalf("line colors = %v, want Tomato", w.lineColors)
	}
}

func TestPenStyleApplyInRectResolvesLinearGradient(t *testing.T) {
	style := &PenStyle{
		kind: PenKindLinearGradient,
		linearGradient: &pdf.LinearGradient{
			Stops: []pdf.GradientStop{
				{Position: 0, Color: NamedColor("Tomato")},
				{Position: 1, Color: NamedColor("SteelBlue")},
			},
		},
		linearPct: &linearGradientPct{
			X0: float64Ptr(0),
			Y0: float64Ptr(50),
			X1: float64Ptr(100),
			Y1: float64Ptr(50),
		},
	}
	w := &labelTestWriter{}

	if err := style.ApplyInRect(w, 10, 20, 120, 40); err != nil {
		t.Fatal(err)
	}
	if len(w.lineLinear) != 1 {
		t.Fatalf("line linear gradient count = %d, want 1", len(w.lineLinear))
	}
	got := w.lineLinear[0]
	if got.X0 != 10 || got.Y0 != 40 || got.X1 != 130 || got.Y1 != 40 {
		t.Fatalf("gradient coords = %#v, want x0=10 y0=40 x1=130 y1=40", got)
	}
}

func TestSample_GradientPens_PrintsGradientStrokes(t *testing.T) {
	doc, err := ParseFile(sampleFile("test_062_gradient_pens.ltml"))
	if err != nil {
		t.Fatal(err)
	}
	writer := &labelTestWriter{t: t}

	if err := doc.Print(writer); err != nil {
		t.Fatal(err)
	}
	if len(writer.lineLinear) == 0 {
		t.Fatal("expected sample to exercise linear gradient pens")
	}
	if len(writer.lineRadial) == 0 {
		t.Fatal("expected sample to exercise radial gradient pens")
	}
}
