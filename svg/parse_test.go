package svg

import "testing"

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
