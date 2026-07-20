package ltml

import "testing"

func TestSetBrushStyleAssignsNamedStyle(t *testing.T) {
	scope := &Scope{}
	base := &BrushStyle{id: "accent", color: NamedColor("Gold")}
	if err := scope.AddStyle(base); err != nil {
		t.Fatal(err)
	}

	var field *BrushStyle
	SetBrushStyle(&field, "text-fill", map[string]string{
		"text-fill": "accent",
	}, scope, "pt")

	if field != base {
		t.Fatalf("field = %p, want named style %p", field, base)
	}
}

func TestSetBrushStyleClonesNamedStyleBeforeOverrides(t *testing.T) {
	scope := &Scope{}
	base := &BrushStyle{id: "accent", color: NamedColor("Gold")}
	if err := scope.AddStyle(base); err != nil {
		t.Fatal(err)
	}

	var field *BrushStyle
	SetBrushStyle(&field, "text-fill", map[string]string{
		"text-fill":       "accent",
		"text-fill.color": "Blue",
	}, scope, "pt")

	if field == base {
		t.Fatal("field reused named style, want clone")
	}
	if field.color != NamedColor("Blue") {
		t.Fatalf("field color = %v, want Blue", field.color)
	}
	if base.color != NamedColor("Gold") {
		t.Fatalf("named style color = %v, want unchanged Gold", base.color)
	}
}

func TestSetBrushStyleCreatesStyleFromOverridesForArbitraryAttribute(t *testing.T) {
	var field *BrushStyle
	SetBrushStyle(&field, "fill", map[string]string{
		"fill.kind": "linear-gradient",
		"fill.x0":   "2",
	}, nil, "mm")

	if field == nil || field.linearGradient == nil {
		t.Fatal("field has no linear gradient, want override-created brush")
	}
	if got, want := field.linearGradient.X0, FromUnits(2, "mm"); got != want {
		t.Fatalf("x0 = %v, want %v", got, want)
	}
}

func TestSetBrushStylePrefixedUnitsOverrideWidgetUnits(t *testing.T) {
	var field *BrushStyle
	SetBrushStyle(&field, "text-fill", map[string]string{
		"text-fill.kind":  "linear-gradient",
		"text-fill.units": "cm",
		"text-fill.x0":    "2",
	}, nil, "mm")

	if got, want := field.linearGradient.X0, FromUnits(2, "cm"); got != want {
		t.Fatalf("x0 = %v, want %v", got, want)
	}
}

func TestSetBrushStyleLeavesFieldUnchangedWithoutMatchingAttrs(t *testing.T) {
	base := &BrushStyle{color: NamedColor("Gold")}
	field := base

	SetBrushStyle(&field, "text-fill", map[string]string{"fill": "Blue"}, nil, "pt")

	if field != base {
		t.Fatalf("field = %p, want unchanged %p", field, base)
	}
}

func TestBrushStyleSetAttrsSolidDefaults(t *testing.T) {
	var bs BrushStyle
	bs.SetAttrs(map[string]string{
		"id":    "primary",
		"color": "Gold",
	})

	if bs.ID() != "primary" {
		t.Fatalf("id = %q, want primary", bs.ID())
	}
	if bs.Kind() != BrushKindSolid {
		t.Fatalf("kind = %q, want solid", bs.Kind())
	}
	if bs.color != NamedColor("Gold") {
		t.Fatalf("color = %#v, want Gold", bs.color)
	}
	if bs.linearGradient != nil || bs.radialGradient != nil || bs.image != nil {
		t.Fatalf("solid brush should not create gradient/image state: %#v", bs)
	}
}

func TestBrushStyleSetAttrsLinearGradient(t *testing.T) {
	var bs BrushStyle
	bs.SetAttrs(map[string]string{
		"kind":  "linear-gradient",
		"x0":    "1.5",
		"y0":    "2.5",
		"x1":    "31.5",
		"y1":    "42.5",
		"stops": "0:#112233, 0.5:Gold, 1:#445566",
	})

	if bs.Kind() != BrushKindLinearGradient {
		t.Fatalf("kind = %q, want linear-gradient", bs.Kind())
	}
	if bs.linearGradient == nil {
		t.Fatal("expected linear gradient to be parsed")
	}
	if got := len(bs.linearGradient.Stops); got != 3 {
		t.Fatalf("len(stops) = %d, want 3", got)
	}
	if bs.linearGradient.X0 != 1.5 || bs.linearGradient.Y0 != 2.5 || bs.linearGradient.X1 != 31.5 || bs.linearGradient.Y1 != 42.5 {
		t.Fatalf("coords = %#v, want parsed values", bs.linearGradient)
	}
	if bs.linearGradient.Stops[0].Color != NamedColor("#112233") {
		t.Fatalf("stop[0].color = %#v, want #112233", bs.linearGradient.Stops[0].Color)
	}
	if bs.linearGradient.Stops[1].Color != NamedColor("Gold") {
		t.Fatalf("stop[1].color = %#v, want Gold", bs.linearGradient.Stops[1].Color)
	}
}

func TestBrushStyleSetAttrsLinearGradientPercentageCoordinates(t *testing.T) {
	var bs BrushStyle
	bs.SetAttrs(map[string]string{
		"kind":  "linear-gradient",
		"x0":    "0%",
		"y0":    "50%",
		"x1":    "100%",
		"y1":    "50%",
		"stops": "0:#112233,1:#445566",
	})

	if bs.linearGradient == nil {
		t.Fatal("expected linear gradient to be parsed")
	}
	if bs.linearPct == nil {
		t.Fatal("expected percentage geometry to be stored")
	}
	if bs.linearGradient.X0 != 0 || bs.linearGradient.Y0 != 0 || bs.linearGradient.X1 != 0 || bs.linearGradient.Y1 != 0 {
		t.Fatalf("absolute coords = %#v, want zeroed percentage placeholders", bs.linearGradient)
	}
	if bs.linearPct.X0 == nil || *bs.linearPct.X0 != 0 {
		t.Fatalf("x0 pct = %v, want 0", bs.linearPct.X0)
	}
	if bs.linearPct.Y0 == nil || *bs.linearPct.Y0 != 50 {
		t.Fatalf("y0 pct = %v, want 50", bs.linearPct.Y0)
	}
	if bs.linearPct.X1 == nil || *bs.linearPct.X1 != 100 {
		t.Fatalf("x1 pct = %v, want 100", bs.linearPct.X1)
	}
	if bs.linearPct.Y1 == nil || *bs.linearPct.Y1 != 50 {
		t.Fatalf("y1 pct = %v, want 50", bs.linearPct.Y1)
	}
}

func TestBrushStyleSetAttrsRadialGradient(t *testing.T) {
	var bs BrushStyle
	bs.SetAttrs(map[string]string{
		"kind":  "radial-gradient",
		"x0":    "10",
		"y0":    "11",
		"r0":    "12",
		"x1":    "20",
		"y1":    "21",
		"r1":    "22",
		"stops": "0:#000000,1:#ffffff",
	})

	if bs.Kind() != BrushKindRadialGradient {
		t.Fatalf("kind = %q, want radial-gradient", bs.Kind())
	}
	if bs.radialGradient == nil {
		t.Fatal("expected radial gradient to be parsed")
	}
	if bs.radialGradient.X0 != 10 || bs.radialGradient.Y0 != 11 || bs.radialGradient.R0 != 12 {
		t.Fatalf("start circle = %#v, want parsed values", bs.radialGradient)
	}
	if bs.radialGradient.X1 != 20 || bs.radialGradient.Y1 != 21 || bs.radialGradient.R1 != 22 {
		t.Fatalf("end circle = %#v, want parsed values", bs.radialGradient)
	}
	if got := len(bs.radialGradient.Stops); got != 2 {
		t.Fatalf("len(stops) = %d, want 2", got)
	}
}

func TestBrushStyleSetAttrsRadialGradientPercentageRadii(t *testing.T) {
	var bs BrushStyle
	bs.SetAttrs(map[string]string{
		"kind":  "radial-gradient",
		"x0":    "50%",
		"y0":    "50%",
		"r0":    "10%",
		"x1":    "50%",
		"y1":    "50%",
		"r1":    "60%",
		"stops": "0:#000000,1:#ffffff",
	})

	if bs.radialGradient == nil {
		t.Fatal("expected radial gradient to be parsed")
	}
	if bs.radialPct == nil {
		t.Fatal("expected radial percentage geometry to be stored")
	}
	if bs.radialGradient.R0 != 0 || bs.radialGradient.R1 != 0 {
		t.Fatalf("absolute radii = %#v, want zeroed percentage placeholders", bs.radialGradient)
	}
	if bs.radialPct.R0 == nil || *bs.radialPct.R0 != 10 {
		t.Fatalf("r0 pct = %v, want 10", bs.radialPct.R0)
	}
	if bs.radialPct.R1 == nil || *bs.radialPct.R1 != 60 {
		t.Fatalf("r1 pct = %v, want 60", bs.radialPct.R1)
	}
}

func TestBrushStyleSetAttrsImage(t *testing.T) {
	var bs BrushStyle
	bs.SetAttrs(map[string]string{
		"kind":        "image",
		"src":         "docs/assets/metal-movable-type-banner.jpg",
		"fit":         "cover",
		"anchor":      "center",
		"repeat":      "tile",
		"opacity":     "0.35",
		"tile-width":  "12",
		"tile-height": "25%",
	})

	if bs.Kind() != BrushKindImage {
		t.Fatalf("kind = %q, want image", bs.Kind())
	}
	if bs.image == nil {
		t.Fatal("expected image brush config to be parsed")
	}
	if bs.image.Src != "docs/assets/metal-movable-type-banner.jpg" {
		t.Fatalf("src = %q, want parsed value", bs.image.Src)
	}
	if bs.image.Fit != "cover" || bs.image.Anchor != "center" || bs.image.Repeat != "tile" {
		t.Fatalf("image options = %#v, want parsed values", bs.image)
	}
	if bs.image.Opacity != 0.35 {
		t.Fatalf("opacity = %v, want 0.35", bs.image.Opacity)
	}
	if bs.image.TileWidth != 12 || bs.image.TileHeightPct != 25 {
		t.Fatalf("tile size = %#v, want tile width 12 and tile height pct 25", bs.image)
	}
}

func TestBrushStyleSetAttrsImageOpacityPercent(t *testing.T) {
	var bs BrushStyle
	bs.SetAttrs(map[string]string{
		"kind":    "image",
		"src":     "fixture.png",
		"opacity": "60%",
	})

	if bs.image == nil {
		t.Fatal("expected image brush config to be parsed")
	}
	if bs.image.Opacity != 0.6 {
		t.Fatalf("opacity = %v, want 0.6", bs.image.Opacity)
	}
}

func TestBrushStyleSetAttrsGradientOpacity(t *testing.T) {
	var bs BrushStyle
	bs.SetAttrs(map[string]string{
		"kind":    "linear-gradient",
		"stops":   "0:#000000,1:#ffffff",
		"opacity": "0.6",
	})

	if bs.Kind() != BrushKindLinearGradient {
		t.Fatalf("kind = %q, want linear-gradient", bs.Kind())
	}
	if bs.opacity == nil || *bs.opacity != 0.6 {
		t.Fatalf("opacity = %v, want 0.6", bs.opacity)
	}
	if bs.image != nil {
		t.Fatalf("image config = %#v, want nil for gradient opacity", bs.image)
	}
}

func TestBrushStyleSetAttrsGradientOpacityPercent(t *testing.T) {
	var bs BrushStyle
	bs.SetAttrs(map[string]string{
		"kind":    "linear-gradient",
		"stops":   "0:#000000,1:#ffffff",
		"opacity": "60%",
	})

	if bs.opacity == nil || *bs.opacity != 0.6 {
		t.Fatalf("opacity = %v, want 0.6", bs.opacity)
	}
	if bs.image != nil {
		t.Fatalf("image config = %#v, want nil for gradient opacity", bs.image)
	}
}

func TestBrushStyleSetAttrsSweepGradient(t *testing.T) {
	var bs BrushStyle
	bs.SetAttrs(map[string]string{
		"kind":    "sweep-gradient",
		"stops":   "0:#112233,0.4:Gold,1:#445566",
		"steps":   "12",
		"opacity": "60%",
	})

	if bs.Kind() != BrushKindSweepGradient {
		t.Fatalf("kind = %q, want sweep-gradient", bs.Kind())
	}
	if bs.sweepGradient == nil {
		t.Fatal("expected sweep gradient to be parsed")
	}
	if got := len(bs.sweepGradient.Stops); got != 3 {
		t.Fatalf("len(stops) = %d, want 3", got)
	}
	if bs.sweepGradient.Steps != 12 {
		t.Fatalf("steps = %d, want 12", bs.sweepGradient.Steps)
	}
	if bs.opacity == nil || *bs.opacity != 0.6 {
		t.Fatalf("opacity = %v, want 0.6", bs.opacity)
	}
	if bs.linearGradient != nil || bs.radialGradient != nil {
		t.Fatalf("sweep brush should not create axial or radial gradient state: %#v", bs)
	}
}

func TestBrushStyleCloneDeepCopiesSweepGradient(t *testing.T) {
	original := &BrushStyle{}
	original.SetAttrs(map[string]string{
		"kind":  "sweep-gradient",
		"stops": "0:#111111,1:#999999",
		"steps": "8",
	})

	clone := original.Clone()
	clone.sweepGradient.Stops[0].Position = 0.25
	clone.sweepGradient.Steps = 2

	if original.sweepGradient.Stops[0].Position != 0 {
		t.Fatalf("original stop position mutated to %v", original.sweepGradient.Stops[0].Position)
	}
	if original.sweepGradient.Steps != 8 {
		t.Fatalf("original steps mutated to %d", original.sweepGradient.Steps)
	}
}

func TestBrushStyleCloneDeepCopiesNestedBrushData(t *testing.T) {
	original := &BrushStyle{}
	original.SetAttrs(map[string]string{
		"kind":    "radial-gradient",
		"x0":      "1",
		"y0":      "2",
		"r0":      "3",
		"x1":      "4",
		"y1":      "5",
		"r1":      "6",
		"stops":   "0:#111111,1:#999999",
		"src":     "ignored-by-kind.png",
		"fit":     "contain",
		"opacity": "0.8",
	})
	original.image = &BrushImageStyle{
		Src:           "docs/assets/metal-movable-type-banner.jpg",
		Fit:           "contain",
		Anchor:        "top",
		Repeat:        "none",
		Opacity:       0.8,
		TileWidth:     18,
		TileHeightPct: 100,
	}

	clone := original.Clone()
	clone.radialGradient.Stops[0].Position = 0.25
	*clone.opacity = 0.2
	clone.image.Opacity = 0.1
	clone.image.TileWidth = 4
	clone.image.TileHeightPct = 50

	if original.radialGradient.Stops[0].Position != 0 {
		t.Fatalf("original stop position mutated to %v", original.radialGradient.Stops[0].Position)
	}
	if original.image.Opacity != 0.8 {
		t.Fatalf("original image opacity mutated to %v", original.image.Opacity)
	}
	if original.opacity == nil || *original.opacity != 0.8 {
		t.Fatalf("original gradient opacity = %v, want 0.8", original.opacity)
	}
	if original.image.TileWidth != 18 {
		t.Fatalf("original image tile width mutated to %v", original.image.TileWidth)
	}
	if original.image.TileHeightPct != 100 {
		t.Fatalf("original image tile height pct mutated to %v", original.image.TileHeightPct)
	}
}
