package ltml

import (
	"math"
	"testing"

	"github.com/rowland/leadtype/pdf"
)

func TestBulletStyle_SetAttrs_ExtendedModes(t *testing.T) {
	scope := &Scope{parent: &defaultScope}
	if err := scope.AddStyle(&FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 12}); err != nil {
		t.Fatal(err)
	}
	style := &BulletStyle{scope: scope, width: 36}

	style.SetAttrs(map[string]string{
		"id":       "brand",
		"font":     "body",
		"text":     "*",
		"format":   "%d.",
		"src":      "fixture.svg",
		"shape":    "star",
		"width":    "18",
		"height":   "12",
		"r":        "5",
		"rx":       "6",
		"ry":       "7",
		"pen":      "solid",
		"brush":    "Gold",
		"align-x":  "center",
		"align-y":  "middle",
		"sides":    "6",
		"points":   "7",
		"rotation": "30",
		"r0":       "4",
	})

	if style.id != "brand" {
		t.Fatalf("id = %q, want brand", style.id)
	}
	if style.font == nil {
		t.Fatal("font is nil, want resolved font style")
	}
	if style.text != "*" {
		t.Fatalf("text = %q, want *", style.text)
	}
	if style.format != "%d." {
		t.Fatalf("format = %q, want %%d.", style.format)
	}
	if !style.IsFormatted() {
		t.Fatal("IsFormatted() = false, want true")
	}
	if style.src != "fixture.svg" {
		t.Fatalf("src = %q, want fixture.svg", style.src)
	}
	if style.shape != "star" {
		t.Fatalf("shape = %q, want star", style.shape)
	}
	if style.width != 18 || style.height != 12 {
		t.Fatalf("width/height = %v/%v, want 18/12", style.width, style.height)
	}
	if style.r != 5 {
		t.Fatalf("r = %v, want 5", style.r)
	}
	if style.rx != 6 || style.ry != 7 {
		t.Fatalf("rx/ry = %v/%v, want 6/7", style.rx, style.ry)
	}
	if style.AlignX() != "center" || style.AlignY() != "middle" {
		t.Fatalf("align-x/align-y = %q/%q, want center/middle", style.AlignX(), style.AlignY())
	}
	if style.pen == nil || style.pen.id != "solid" {
		t.Fatalf("pen = %#v, want solid", style.pen)
	}
	if style.brush == nil {
		t.Fatal("brush is nil, want resolved brush style")
	}
	if style.sides != 6 || style.points != 7 {
		t.Fatalf("sides/points = %d/%d, want 6/7", style.sides, style.points)
	}
	if style.rotation != 30 {
		t.Fatalf("rotation = %v, want 30", style.rotation)
	}
	if style.r0 != 4 {
		t.Fatalf("r0 = %v, want 4", style.r0)
	}
}

func TestBulletStyle_SetAttrs_ClonesNamedFontForOverrides(t *testing.T) {
	scope := &Scope{}
	base := &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, weight: "Regular"}
	if err := scope.AddStyle(base); err != nil {
		t.Fatal(err)
	}
	style := &BulletStyle{scope: scope}

	style.SetAttrs(map[string]string{
		"font":        "body",
		"font.weight": "Bold",
	})

	if style.font == base {
		t.Fatal("bullet font reused named style, want clone")
	}
	if style.font == nil || style.font.weight != "Bold" {
		t.Fatalf("bullet font = %#v, want Bold", style.font)
	}
	if base.weight != "Regular" {
		t.Fatalf("named font weight = %q, want unchanged Regular", base.weight)
	}
}

func TestBulletStyle_SetAttrs_FontOverridesCloneDefaultWithBulletUnits(t *testing.T) {
	style := &BulletStyle{units: "mm"}

	style.SetAttrs(map[string]string{
		"font.weight":       "Bold",
		"font.stroke-width": "2",
	})

	if style.font == nil || style.font == defaultFont {
		t.Fatalf("bullet font = %p, want clone of default", style.font)
	}
	if style.font.weight != "Bold" {
		t.Fatalf("bullet font weight = %q, want Bold", style.font.weight)
	}
	if style.font.strokeWidth == nil || *style.font.strokeWidth != FromUnits(2, "mm") {
		t.Fatalf("bullet stroke width = %v, want 2mm", style.font.strokeWidth)
	}
	if defaultFont.weight != "" || defaultFont.strokeWidth != nil {
		t.Fatalf("default font mutated: %#v", defaultFont)
	}
}

func TestBulletStyleFor_BuiltInListMarkers(t *testing.T) {
	ordered := BulletStyleFor("ordered", &defaultScope)
	if ordered == nil || ordered.Format() != "%d." {
		t.Fatalf("ordered = %#v, want formatted %%d. bullet", ordered)
	}
	unordered := BulletStyleFor("unordered", &defaultScope)
	if unordered == nil || unordered.Shape() != "circle" || unordered.Width() != defaultListBulletSize || unordered.AlignY() != "middle" {
		t.Fatalf("unordered = %#v, want default circle bullet", unordered)
	}
}

func TestBulletStyle_ModePrecedence(t *testing.T) {
	style := &BulletStyle{src: "fixture.jpg", shape: "circle", text: "*"}

	if !style.IsImage() {
		t.Fatal("IsImage() = false, want true")
	}
	if style.IsShape() {
		t.Fatal("IsShape() = true, want false when src is present")
	}
}

func TestParse_BulletTag_ExtendedAttrs(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml>
  <brush id="goldfill" color="Gold" />
  <page>
    <bullet id="brand" src="fixture.svg" width="18pt" height="12pt" />
    <bullet id="shape" shape="star" width="18pt" height="12pt" brush="goldfill" points="6" r0="4pt" />
  </page>
</ltml>`), WithAssetFS(testingMapFS("fixture.svg", "<svg/>")))
	if err != nil {
		t.Fatal(err)
	}

	page := doc.Root().Page(0)
	brand := BulletStyleFor("brand", page)
	if brand == nil || brand.src != "fixture.svg" {
		t.Fatalf("brand bullet = %#v, want src fixture.svg", brand)
	}
	shape := BulletStyleFor("shape", page)
	if shape == nil {
		t.Fatal("shape bullet not found")
	}
	if shape.shape != "star" || shape.points != 6 || shape.r0 != 4 {
		t.Fatalf("shape bullet = %#v", shape)
	}
	if shape.brush == nil || shape.brush.id != "goldfill" {
		t.Fatalf("shape brush = %#v, want goldfill", shape.brush)
	}
}

func TestClosedShapeForBullet_TriangleAndSquareAliases(t *testing.T) {
	triangle := &BulletStyle{shape: "triangle", sides: 9}
	gotTriangle := closedShapeForBullet(triangle, 10, 20, 18)
	if gotTriangle.Kind != pdf.ClosedShapePolygon {
		t.Fatalf("triangle kind = %q, want polygon", gotTriangle.Kind)
	}
	if gotTriangle.Sides != 3 {
		t.Fatalf("triangle sides = %d, want 3", gotTriangle.Sides)
	}
	if gotTriangle.Rotation != 180 {
		t.Fatalf("triangle rotation = %v, want 180", gotTriangle.Rotation)
	}

	square := &BulletStyle{shape: "square", sides: 9}
	gotSquare := closedShapeForBullet(square, 10, 20, 18)
	if gotSquare.Kind != pdf.ClosedShapePolygon {
		t.Fatalf("square kind = %q, want polygon", gotSquare.Kind)
	}
	if gotSquare.Sides != 4 {
		t.Fatalf("square sides = %d, want 4", gotSquare.Sides)
	}
	if gotSquare.Rotation != 0 {
		t.Fatalf("square rotation = %v, want 0", gotSquare.Rotation)
	}
}

func TestClosedShapeForBullet_TriangleExplicitRotationWins(t *testing.T) {
	triangle := &BulletStyle{shape: "triangle", rotation: 30, rotationSet: true}
	got := closedShapeForBullet(triangle, 10, 20, 18)
	if got.Rotation != 210 {
		t.Fatalf("triangle rotation = %v, want 210", got.Rotation)
	}
}

func TestClosedShapeForBullet_LowPointPolygonAndStarDefaultPointUpRotation(t *testing.T) {
	polygon5 := &BulletStyle{shape: "polygon", sides: 5}
	if got := closedShapeForBullet(polygon5, 10, 20, 18).Rotation; got != 36 {
		t.Fatalf("5-point polygon rotation = %v, want 36", got)
	}

	polygon6 := &BulletStyle{shape: "polygon", sides: 6}
	if got := closedShapeForBullet(polygon6, 10, 20, 18).Rotation; got != 30 {
		t.Fatalf("6-point polygon rotation = %v, want 30", got)
	}

	star5 := &BulletStyle{shape: "star", points: 5}
	if got := closedShapeForBullet(star5, 10, 20, 18).Rotation; got != 36 {
		t.Fatalf("5-point star rotation = %v, want 36", got)
	}

	star6 := &BulletStyle{shape: "star", points: 6}
	if got := closedShapeForBullet(star6, 10, 20, 18).Rotation; got != 30 {
		t.Fatalf("6-point star rotation = %v, want 30", got)
	}

	star3 := &BulletStyle{shape: "star", points: 3}
	if got := closedShapeForBullet(star3, 10, 20, 18).Rotation; got != 60 {
		t.Fatalf("3-point star rotation = %v, want 60", got)
	}

	star7 := &BulletStyle{shape: "star", points: 7}
	if got := closedShapeForBullet(star7, 10, 20, 18).Rotation; math.Abs(got-180.0/7.0) > 0.0001 {
		t.Fatalf("7-point star rotation = %v, want %v", got, 180.0/7.0)
	}
}

func TestClosedShapeForBullet_UserRotationAddsToDefault(t *testing.T) {
	polygon5 := &BulletStyle{shape: "polygon", sides: 5, rotation: 10, rotationSet: true}
	if got := closedShapeForBullet(polygon5, 10, 20, 18).Rotation; got != 46 {
		t.Fatalf("5-point polygon rotation = %v, want 46", got)
	}

	star5 := &BulletStyle{shape: "star", points: 5, rotation: 10, rotationSet: true}
	if got := closedShapeForBullet(star5, 10, 20, 18).Rotation; got != 46 {
		t.Fatalf("5-point star rotation = %v, want 46", got)
	}
}
