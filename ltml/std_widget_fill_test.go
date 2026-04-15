package ltml

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rowland/leadtype/pdf"
)

type paintedImageCall struct {
	filename string
	x        float64
	y        float64
	width    float64
	height   float64
}

type backgroundFillTestWriter struct {
	labelTestWriter
	pathCount      int
	clipCount      int
	pathRects      int
	linearPaints   []*pdf.LinearGradient
	radialPaints   []*pdf.RadialGradient
	imagePaints    []paintedImageCall
	fileDimensions map[string][2]int
}

func (w *backgroundFillTestWriter) Path(fn func()) error {
	w.pathCount++
	if fn != nil {
		fn()
	}
	return nil
}

func (w *backgroundFillTestWriter) Clip(fn func()) error {
	w.clipCount++
	if fn != nil {
		fn()
	}
	return nil
}

func (w *backgroundFillTestWriter) Rectangle2(x, y, width, height float64, border bool, fill bool, corners []float64, path, reverse bool) {
	w.labelTestWriter.Rectangle2(x, y, width, height, border, fill, corners, path, reverse)
	if path {
		w.pathRects++
	}
}

func (w *backgroundFillTestWriter) PaintLinearGradient(lg *pdf.LinearGradient) error {
	w.linearPaints = append(w.linearPaints, lg)
	return nil
}

func (w *backgroundFillTestWriter) PaintRadialGradient(rg *pdf.RadialGradient) error {
	w.radialPaints = append(w.radialPaints, rg)
	return nil
}

func (w *backgroundFillTestWriter) PaintImageFile(filename string, x, y, width, height float64) error {
	w.imagePaints = append(w.imagePaints, paintedImageCall{
		filename: filename,
		x:        x,
		y:        y,
		width:    width,
		height:   height,
	})
	return nil
}

func (w *backgroundFillTestWriter) ImageDimensionsFromFile(filename string) (width, height int, err error) {
	if dims, ok := w.fileDimensions[filename]; ok {
		return dims[0], dims[1], nil
	}
	return 0, 0, nil
}

func TestStdWidgetPaintBackground_LinearGradientUsesClipPath(t *testing.T) {
	widget := &StdWidget{}
	widget.SetLeft(10)
	widget.SetTop(20)
	widget.SetWidth(120)
	widget.SetHeight(50)
	widget.fill = &BrushStyle{
		kind: BrushKindLinearGradient,
		linearGradient: &pdf.LinearGradient{
			X0: 10, Y0: 20, X1: 130, Y1: 20,
			Stops: []pdf.GradientStop{
				{Position: 0, Color: NamedColor("Tomato")},
				{Position: 1, Color: NamedColor("Gold")},
			},
		},
	}
	writer := &backgroundFillTestWriter{}

	if err := widget.PaintBackground(writer); err != nil {
		t.Fatal(err)
	}
	if writer.pathCount != 1 {
		t.Fatalf("path count = %d, want 1", writer.pathCount)
	}
	if writer.clipCount != 1 {
		t.Fatalf("clip count = %d, want 1", writer.clipCount)
	}
	if writer.pathRects != 1 {
		t.Fatalf("path rect count = %d, want 1", writer.pathRects)
	}
	if len(writer.linearPaints) != 1 {
		t.Fatalf("linear paint count = %d, want 1", len(writer.linearPaints))
	}
	if got := writer.linearPaints[0].X0; got != 20 {
		t.Fatalf("linear x0 = %v, want 20", got)
	}
	if got := writer.linearPaints[0].Y0; got != 40 {
		t.Fatalf("linear y0 = %v, want 40", got)
	}
	if got := writer.linearPaints[0].X1; got != 140 {
		t.Fatalf("linear x1 = %v, want 140", got)
	}
	if got := writer.linearPaints[0].Y1; got != 40 {
		t.Fatalf("linear y1 = %v, want 40", got)
	}
	if len(writer.fillRectPages) != 0 {
		t.Fatalf("solid fill rect count = %d, want 0", len(writer.fillRectPages))
	}
}

func TestStdWidgetPaintBackground_RadialGradientUsesBoxLocalCoordinates(t *testing.T) {
	widget := &StdWidget{}
	widget.SetLeft(100)
	widget.SetTop(200)
	widget.SetWidth(120)
	widget.SetHeight(80)
	widget.fill = &BrushStyle{
		kind: BrushKindRadialGradient,
		radialGradient: &pdf.RadialGradient{
			X0: 20, Y0: 15, R0: 0,
			X1: 20, Y1: 15, R1: 45,
			Stops: []pdf.GradientStop{
				{Position: 0, Color: NamedColor("White")},
				{Position: 1, Color: NamedColor("Tomato")},
			},
		},
	}
	writer := &backgroundFillTestWriter{}

	if err := widget.PaintBackground(writer); err != nil {
		t.Fatal(err)
	}
	if len(writer.radialPaints) != 1 {
		t.Fatalf("radial paint count = %d, want 1", len(writer.radialPaints))
	}
	if got := writer.radialPaints[0].X0; got != 120 {
		t.Fatalf("radial x0 = %v, want 120", got)
	}
	if got := writer.radialPaints[0].Y0; got != 215 {
		t.Fatalf("radial y0 = %v, want 215", got)
	}
	if got := writer.radialPaints[0].X1; got != 120 {
		t.Fatalf("radial x1 = %v, want 120", got)
	}
	if got := writer.radialPaints[0].Y1; got != 215 {
		t.Fatalf("radial y1 = %v, want 215", got)
	}
	if got := writer.radialPaints[0].R1; got != 45 {
		t.Fatalf("radial r1 = %v, want 45", got)
	}
}

func TestStdWidgetSetAttrs_NormalizesBrushGradientMeasurements(t *testing.T) {
	var widget StdWidget
	widget.SetAttrs(map[string]string{
		"units":      "in",
		"fill.kind":  "radial-gradient",
		"fill.x0":    "0.5",
		"fill.y0":    "1.25in",
		"fill.r0":    "6pt",
		"fill.x1":    "2",
		"fill.y1":    "2.5",
		"fill.r1":    "1.5in",
		"fill.stops": "0:#ffffff,1:#000000",
	})

	if widget.fill == nil || widget.fill.radialGradient == nil {
		t.Fatal("expected normalized radial gradient")
	}
	if got := widget.fill.radialGradient.X0; got != 36 {
		t.Fatalf("x0 = %v, want 36", got)
	}
	if got := widget.fill.radialGradient.Y0; got != 90 {
		t.Fatalf("y0 = %v, want 90", got)
	}
	if got := widget.fill.radialGradient.R0; got != 6 {
		t.Fatalf("r0 = %v, want 6", got)
	}
	if got := widget.fill.radialGradient.X1; got != 144 {
		t.Fatalf("x1 = %v, want 144", got)
	}
	if got := widget.fill.radialGradient.Y1; got != 180 {
		t.Fatalf("y1 = %v, want 180", got)
	}
	if got := widget.fill.radialGradient.R1; got != 108 {
		t.Fatalf("r1 = %v, want 108", got)
	}
}

func TestStdWidgetPaintBackground_ImageRepeatTilesWithinClip(t *testing.T) {
	doc, err := Parse([]byte(`<ltml><page /></ltml>`), WithAssetFS(testingMapFS("fixture.png", "image-data")))
	if err != nil {
		t.Fatal(err)
	}
	page := doc.Root().Page(0)
	widget := &StdWidget{}
	if err := widget.SetContainer(page); err != nil {
		t.Fatal(err)
	}
	widget.SetDoc(doc)
	widget.SetLeft(5)
	widget.SetTop(7)
	widget.SetWidth(100)
	widget.SetHeight(40)
	widget.fill = &BrushStyle{
		kind: BrushKindImage,
		image: &BrushImageStyle{
			Src:    "fixture.png",
			Fit:    "tile",
			Anchor: "top-left",
			Repeat: "repeat-x",
		},
	}
	writer := &backgroundFillTestWriter{
		fileDimensions: map[string][2]int{
			"fixture.png": {20, 10},
		},
	}

	if err := widget.PaintBackground(writer); err != nil {
		t.Fatal(err)
	}
	if writer.pathCount != 1 || writer.clipCount != 1 {
		t.Fatalf("path/clip counts = %d/%d, want 1/1", writer.pathCount, writer.clipCount)
	}
	if got := len(writer.imagePaints); got != 5 {
		t.Fatalf("image paint count = %d, want 5", got)
	}
	if writer.imagePaints[0].x != 5 || writer.imagePaints[0].y != 7 {
		t.Fatalf("first tile origin = (%v,%v), want (5,7)", writer.imagePaints[0].x, writer.imagePaints[0].y)
	}
	if writer.imagePaints[0].width != 20 || writer.imagePaints[0].height != 10 {
		t.Fatalf("tile size = %vx%v, want 20x10", writer.imagePaints[0].width, writer.imagePaints[0].height)
	}
}

func TestSample_WidgetBrushBackgrounds_PrintsWithoutWidgetSpecificLogic(t *testing.T) {
	doc, err := ParseFile(sampleFile("test_045_widget_brush_backgrounds.ltml"))
	if err != nil {
		t.Fatal(err)
	}
	writer := &backgroundFillTestWriter{
		labelTestWriter: labelTestWriter{t: t},
		fileDimensions: map[string][2]int{
			filepath.Join(filepath.Dir(sampleFile("test_045_widget_brush_backgrounds.ltml")), "../../docs/assets/metal-movable-type-banner.jpg"): {1200, 400},
		},
	}

	if err := doc.Print(writer); err != nil {
		t.Fatal(err)
	}
	if len(writer.linearPaints) == 0 {
		t.Fatal("expected sample to exercise gradient background painting")
	}
	if len(writer.imagePaints) == 0 {
		t.Fatal("expected sample to exercise image background painting")
	}
	foundDocAsset := false
	for _, call := range writer.imagePaints {
		if strings.HasSuffix(call.filename, filepath.Join("docs", "assets", "metal-movable-type-banner.jpg")) {
			foundDocAsset = true
			break
		}
	}
	if !foundDocAsset {
		t.Fatalf("sample image paints = %#v, want docs/assets/metal-movable-type-banner.jpg", writer.imagePaints)
	}
}

func TestSample_WidgetBrushBackgrounds_FileExists(t *testing.T) {
	if _, err := os.Stat(sampleFile("test_045_widget_brush_backgrounds.ltml")); err != nil {
		t.Fatal(err)
	}
}
