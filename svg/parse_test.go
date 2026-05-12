package svg

import (
	"strings"
	"testing"
)

func TestParseRootUsesViewBoxWhenSizeMissing(t *testing.T) {
	doc, warnings, err := Parse([]byte(`<svg viewBox="0 0 200 100"><rect width="200" height="100"/></svg>`))
	if err != nil {
		t.Fatal(err)
	}
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %+v", warnings)
	}
	if doc.Width != 200 || doc.Height != 100 {
		t.Fatalf("dimensions = %.0fx%.0f, want 200x100", doc.Width, doc.Height)
	}
}

func TestParseRootUsesDefaultSizeWhenSizingMissing(t *testing.T) {
	doc, warnings, err := Parse([]byte(`<svg><rect width="100" height="100"/></svg>`))
	if err != nil {
		t.Fatal(err)
	}
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %+v", warnings)
	}
	if doc.Width != defaultSVGWidth || doc.Height != defaultSVGHeight {
		t.Fatalf("dimensions = %.0fx%.0f, want %.0fx%.0f", doc.Width, doc.Height, defaultSVGWidth, defaultSVGHeight)
	}
}

func TestParseRootUsesDefaultSizeWhenPercentSizingHasNoViewport(t *testing.T) {
	doc, warnings, err := Parse([]byte(`<svg width="100%" height="100%"><rect width="100" height="100"/></svg>`))
	if err != nil {
		t.Fatal(err)
	}
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %+v", warnings)
	}
	if doc.Width != defaultSVGWidth || doc.Height != defaultSVGHeight {
		t.Fatalf("dimensions = %.0fx%.0f, want %.0fx%.0f", doc.Width, doc.Height, defaultSVGWidth, defaultSVGHeight)
	}
}

func TestDimensionsUsesRootAttributesOnly(t *testing.T) {
	width, height, err := Dimensions([]byte(`<svg width="120" height="80"><g><broken></svg>`))
	if err != nil {
		t.Fatal(err)
	}
	if width != 120 || height != 80 {
		t.Fatalf("dimensions = %.0fx%.0f, want 120x80", width, height)
	}
}

func TestDimensionsUsesViewBoxWhenSizeMissing(t *testing.T) {
	width, height, err := Dimensions([]byte(`<svg viewBox="0 0 200 100"><g><broken></svg>`))
	if err != nil {
		t.Fatal(err)
	}
	if width != 200 || height != 100 {
		t.Fatalf("dimensions = %.0fx%.0f, want 200x100", width, height)
	}
}

func TestDimensionsUsesDefaultSizeWhenSizingMissing(t *testing.T) {
	width, height, err := Dimensions([]byte(`<svg><g><broken></svg>`))
	if err != nil {
		t.Fatal(err)
	}
	if width != defaultSVGWidth || height != defaultSVGHeight {
		t.Fatalf("dimensions = %.0fx%.0f, want %.0fx%.0f", width, height, defaultSVGWidth, defaultSVGHeight)
	}
}

func TestParseColorFormats(t *testing.T) {
	tests := []struct {
		in   string
		want int
	}{
		{"#123", 0x112233},
		{"#112233", 0x112233},
		{"rgb(17, 34, 51)", 0x112233},
		{"rgb(6.6667%, 13.3333%, 20%)", 0x112233},
	}
	for _, test := range tests {
		paint, err := parseColor(test.in)
		if err != nil {
			t.Fatalf("parseColor(%q): %v", test.in, err)
		}
		if int(paint.Color) != test.want {
			t.Fatalf("parseColor(%q) = %#x, want %#x", test.in, paint.Color, test.want)
		}
	}
}

func TestParseLengthUnits(t *testing.T) {
	if got, err := parseLength("1in", 100); err != nil || got != 72 {
		t.Fatalf("1in = %v, %v", got, err)
	}
	if got, err := parseLength("50%", 200); err != nil || got != 100 {
		t.Fatalf("50%% = %v, %v", got, err)
	}
}

func TestParseTransform(t *testing.T) {
	transform, err := parseTransform("translate(10,20) rotate(90)")
	if err != nil {
		t.Fatal(err)
	}
	point := transform.Apply(Point{X: 5, Y: 0})
	if point.X < 9.999 || point.X > 10.001 || point.Y < 24.999 || point.Y > 25.001 {
		t.Fatalf("unexpected transformed point %+v", point)
	}
}

func TestParsePathDataArc(t *testing.T) {
	segments, err := parsePathData("M 0 0 A 25 25 0 0 1 50 0")
	if err != nil {
		t.Fatal(err)
	}
	if len(segments) < 2 {
		t.Fatalf("expected arc to expand into cubic segments, got %d segments", len(segments))
	}
	if segments[0].Type != SegmentMoveTo {
		t.Fatalf("first segment type = %v, want move", segments[0].Type)
	}
	last := segments[len(segments)-1]
	if last.Type != SegmentCubicTo {
		t.Fatalf("last segment type = %v, want cubic", last.Type)
	}
	end := last.Points[len(last.Points)-1]
	if end.X < 49.999 || end.X > 50.001 || end.Y < -0.001 || end.Y > 0.001 {
		t.Fatalf("arc end = %+v, want (50,0)", end)
	}
}

func TestParseClipPath(t *testing.T) {
	doc, warnings, err := Parse([]byte(`<svg width="100" height="100"><defs><clipPath id="c"><rect x="0" y="0" width="10" height="10"/></clipPath></defs><rect clip-path="url(#c)" width="100" height="100"/></svg>`))
	if err != nil {
		t.Fatal(err)
	}
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %+v", warnings)
	}
	if doc.ClipPaths["c"] == nil {
		t.Fatal("expected clip path to be registered")
	}
}

func TestParseInternalStyleClassFill(t *testing.T) {
	doc, warnings, err := Parse([]byte(`<svg width="100" height="100"><style>.st0{fill:#21495A;}</style><path class="st0" d="M0 0 L10 0 L10 10 Z"/></svg>`))
	if err != nil {
		t.Fatal(err)
	}
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %+v", warnings)
	}
	if len(doc.Children) != 1 {
		t.Fatalf("child count = %d, want 1", len(doc.Children))
	}
	path, ok := doc.Children[0].(*Path)
	if !ok {
		t.Fatalf("child type = %T, want *Path", doc.Children[0])
	}
	style := path.Style.Resolve(DefaultStyle())
	if !style.Fill.Set || style.Fill.None {
		t.Fatal("expected class style to set a visible fill")
	}
	if int(style.Fill.Color) != 0x21495A {
		t.Fatalf("fill color = %#x, want %#x", style.Fill.Color, 0x21495A)
	}
}

func TestParseFontFamilyCandidates(t *testing.T) {
	doc, warnings, err := Parse([]byte(`<svg width="100" height="20"><text font-family="'OpenSans-Regular', Minimal, &quot;serif&quot;">Hi</text></svg>`))
	if err != nil {
		t.Fatal(err)
	}
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %+v", warnings)
	}
	text, ok := doc.Children[0].(*Text)
	if !ok {
		t.Fatalf("child type = %T, want *Text", doc.Children[0])
	}
	style := text.Style.Resolve(DefaultStyle())
	want := []string{"OpenSans-Regular", "Minimal", "serif"}
	if style.FontFamily != want[0] {
		t.Fatalf("FontFamily = %q, want %q", style.FontFamily, want[0])
	}
	if len(style.FontFamilies) != len(want) {
		t.Fatalf("FontFamilies = %#v, want %#v", style.FontFamilies, want)
	}
	for i := range want {
		if style.FontFamilies[i] != want[i] {
			t.Fatalf("FontFamilies = %#v, want %#v", style.FontFamilies, want)
		}
	}
}

func TestParseFontFamilyDoubleWrappedQuotes(t *testing.T) {
	doc, warnings, err := Parse([]byte(`<svg width="100" height="20"><text font-family="&quot;'Montserrat-Bold'&quot;">Hi</text></svg>`))
	if err != nil {
		t.Fatal(err)
	}
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %+v", warnings)
	}
	text := doc.Children[0].(*Text)
	style := text.Style.Resolve(DefaultStyle())
	if style.FontFamily != "Montserrat-Bold" {
		t.Fatalf("FontFamily = %q, want Montserrat-Bold", style.FontFamily)
	}
}

func TestParseFontFamilyEmptyEntriesWarn(t *testing.T) {
	doc, warnings, err := Parse([]byte(`<svg width="100" height="20"><text font-family="'Missing', , Minimal">Hi</text></svg>`))
	if err != nil {
		t.Fatal(err)
	}
	if len(warnings) != 1 {
		t.Fatalf("warnings = %+v, want one empty-family warning", warnings)
	}
	if warnings[0].Attribute != "font-family" {
		t.Fatalf("warning attribute = %q, want font-family", warnings[0].Attribute)
	}
	text := doc.Children[0].(*Text)
	style := text.Style.Resolve(DefaultStyle())
	want := []string{"Missing", "Minimal"}
	if len(style.FontFamilies) != len(want) {
		t.Fatalf("FontFamilies = %#v, want %#v", style.FontFamilies, want)
	}
	for i := range want {
		if style.FontFamilies[i] != want[i] {
			t.Fatalf("FontFamilies = %#v, want %#v", style.FontFamilies, want)
		}
	}
}

func TestParseFontFamilyMalformedQuoteWarns(t *testing.T) {
	_, warnings, err := Parse([]byte(`<svg width="100" height="20"><text font-family="'Missing, Minimal">Hi</text></svg>`))
	if err != nil {
		t.Fatal(err)
	}
	if len(warnings) != 1 {
		t.Fatalf("warnings = %+v, want one malformed-family warning", warnings)
	}
	if warnings[0].Attribute != "font-family" || !strings.Contains(warnings[0].Message, "unterminated quote") {
		t.Fatalf("warning = %+v, want unterminated font-family warning", warnings[0])
	}
}

func TestParseLinearGradientAndURLPaint(t *testing.T) {
	doc, warnings, err := Parse([]byte(`<svg width="100" height="20"><defs><style>.grad{fill:url(#SVGID_1_);opacity:0.8;clip-path:url(#clip);}</style><clipPath id="clip"><rect width="100" height="20"/></clipPath><linearGradient id="SVGID_1_" x1="0" y1="0" x2="100" y2="0" gradientUnits="userSpaceOnUse"><stop offset="0" stop-color="#112233"/><stop offset="1" style="stop-color:#445566;stop-opacity:0.4"/></linearGradient></defs><rect class="grad" width="100" height="20"/></svg>`))
	if err != nil {
		t.Fatal(err)
	}
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %+v", warnings)
	}
	gradient := doc.Gradients["SVGID_1_"]
	if gradient == nil {
		t.Fatal("expected linear gradient to be registered")
	}
	if gradient.Kind != GradientLinear {
		t.Fatalf("gradient kind = %v, want linear", gradient.Kind)
	}
	if len(gradient.Stops) != 2 {
		t.Fatalf("stop count = %d, want 2", len(gradient.Stops))
	}
	if gradient.Stops[1].Opacity != 0.4 {
		t.Fatalf("stop opacity = %v, want 0.4", gradient.Stops[1].Opacity)
	}
	var rect *Rect
	for _, child := range doc.Children {
		if node, ok := child.(*Rect); ok {
			rect = node
			break
		}
	}
	if rect == nil {
		t.Fatalf("expected a drawable rect child, got %+v", doc.Children)
	}
	style := rect.Style.Resolve(DefaultStyle())
	if style.Fill.Ref != "SVGID_1_" {
		t.Fatalf("fill ref = %q, want SVGID_1_", style.Fill.Ref)
	}
	if rect.ClipPathRef != "clip" {
		t.Fatalf("clip-path ref = %q, want clip", rect.ClipPathRef)
	}
	if style.Opacity != 0.8 {
		t.Fatalf("opacity = %v, want 0.8", style.Opacity)
	}
}

func TestParseUseMaskAndBlendMode(t *testing.T) {
	doc, warnings, err := Parse([]byte(`<svg width="20" height="20" xmlns:xlink="http://www.w3.org/1999/xlink"><defs><rect id="shape" width="10" height="10"/><mask id="mask" maskUnits="userSpaceOnUse"><rect width="20" height="20" fill="#ffffff"/></mask></defs><use xlink:href="#shape" x="5" y="6" style="mix-blend-mode:hard-light;mask:url(#mask)"/></svg>`))
	if err != nil {
		t.Fatal(err)
	}
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %+v", warnings)
	}
	if doc.Masks["mask"] == nil {
		t.Fatal("expected mask definition to be registered")
	}
	var use *Use
	for _, child := range doc.Children {
		if node, ok := child.(*Use); ok {
			use = node
			break
		}
	}
	if use == nil {
		t.Fatalf("expected a use node child, got %+v", doc.Children)
	}
	if use.Href != "shape" {
		t.Fatalf("href = %q, want shape", use.Href)
	}
	if use.MaskRef != "mask" {
		t.Fatalf("mask ref = %q, want mask", use.MaskRef)
	}
	style := use.Style.Resolve(DefaultStyle())
	if style.BlendMode != "hard-light" {
		t.Fatalf("blend mode = %q, want hard-light", style.BlendMode)
	}
}

func TestParseMetadataElementsAreIgnoredWithoutWarning(t *testing.T) {
	doc, warnings, err := Parse([]byte(`<svg width="10" height="10"><title>Example</title><desc>Details</desc><rect width="10" height="10"/></svg>`))
	if err != nil {
		t.Fatal(err)
	}
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %+v", warnings)
	}
	if len(doc.Children) != 1 {
		t.Fatalf("child count = %d, want 1", len(doc.Children))
	}
	if _, ok := doc.Children[0].(*Rect); !ok {
		t.Fatalf("child type = %T, want *Rect", doc.Children[0])
	}
}

func TestParseUnsupportedElementWarning(t *testing.T) {
	_, warnings, err := Parse([]byte(`<svg width="10" height="10"><foreignObject /></svg>`))
	if err != nil {
		t.Fatal(err)
	}
	if len(warnings) == 0 {
		t.Fatal("expected unsupported element warning")
	}
}

func TestParseISO88591XML(t *testing.T) {
	data := []byte{
		'<', '?', 'x', 'm', 'l', ' ', 'v', 'e', 'r', 's', 'i', 'o', 'n', '=', '"', '1', '.', '0', '"', ' ',
		'e', 'n', 'c', 'o', 'd', 'i', 'n', 'g', '=', '"', 'I', 'S', 'O', '-', '8', '8', '5', '9', '-', '1', '"', '?', '>',
		'<', 's', 'v', 'g', ' ', 'w', 'i', 'd', 't', 'h', '=', '"', '1', '0', '0', '"', ' ', 'h', 'e', 'i', 'g', 'h', 't', '=', '"', '4', '0', '"', '>',
		'<', 't', 'e', 'x', 't', ' ', 'x', '=', '"', '1', '0', '"', ' ', 'y', '=', '"', '2', '0', '"', '>', 'c', 'a', 'f', 0xe9, '<', '/', 't', 'e', 'x', 't', '>',
		'<', '/', 's', 'v', 'g', '>',
	}
	doc, warnings, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %+v", warnings)
	}
	if len(doc.Children) != 1 {
		t.Fatalf("child count = %d, want 1", len(doc.Children))
	}
	text, ok := doc.Children[0].(*Text)
	if !ok {
		t.Fatalf("child type = %T, want *Text", doc.Children[0])
	}
	if text.Body != "cafe" && text.Body != "café" {
		t.Fatalf("text body = %q, want decoded Latin-1 text", text.Body)
	}
	if text.Body != "café" {
		t.Fatalf("text body = %q, want café", text.Body)
	}
}
