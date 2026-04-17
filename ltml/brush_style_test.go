package ltml

import "testing"

func TestBrushStyleSetAttrsSolidDefaults(t *testing.T) {
	var bs BrushStyle
	bs.SetAttrs("", map[string]string{
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
	bs.SetAttrs("fill.", map[string]string{
		"fill.kind":  "linear-gradient",
		"fill.x0":    "1.5",
		"fill.y0":    "2.5",
		"fill.x1":    "31.5",
		"fill.y1":    "42.5",
		"fill.stops": "0:#112233, 0.5:Gold, 1:#445566",
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

func TestBrushStyleSetAttrsRadialGradient(t *testing.T) {
	var bs BrushStyle
	bs.SetAttrs("fill.", map[string]string{
		"fill.kind":  "radial-gradient",
		"fill.x0":    "10",
		"fill.y0":    "11",
		"fill.r0":    "12",
		"fill.x1":    "20",
		"fill.y1":    "21",
		"fill.r1":    "22",
		"fill.stops": "0:#000000,1:#ffffff",
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

func TestBrushStyleSetAttrsImage(t *testing.T) {
	var bs BrushStyle
	bs.SetAttrs("fill.", map[string]string{
		"fill.kind":    "image",
		"fill.src":     "docs/assets/metal-movable-type-banner.jpg",
		"fill.fit":     "cover",
		"fill.anchor":  "center",
		"fill.repeat":  "tile",
		"fill.opacity": "0.35",
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
}

func TestBrushStyleCloneDeepCopiesNestedBrushData(t *testing.T) {
	original := &BrushStyle{}
	original.SetAttrs("fill.", map[string]string{
		"fill.kind":    "radial-gradient",
		"fill.x0":      "1",
		"fill.y0":      "2",
		"fill.r0":      "3",
		"fill.x1":      "4",
		"fill.y1":      "5",
		"fill.r1":      "6",
		"fill.stops":   "0:#111111,1:#999999",
		"fill.src":     "ignored-by-kind.png",
		"fill.fit":     "contain",
		"fill.opacity": "0.8",
	})
	original.image = &BrushImageStyle{
		Src:     "docs/assets/metal-movable-type-banner.jpg",
		Fit:     "contain",
		Anchor:  "top",
		Repeat:  "none",
		Opacity: 0.8,
	}

	clone := original.Clone()
	clone.radialGradient.Stops[0].Position = 0.25
	clone.image.Opacity = 0.1

	if original.radialGradient.Stops[0].Position != 0 {
		t.Fatalf("original stop position mutated to %v", original.radialGradient.Stops[0].Position)
	}
	if original.image.Opacity != 0.8 {
		t.Fatalf("original image opacity mutated to %v", original.image.Opacity)
	}
}
