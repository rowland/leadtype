// Copyright 2026 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"bytes"
	"strings"
	"testing"

	"github.com/rowland/leadtype/ltml/ltpdf"
)

var _ CanvasDrawer = (*ltpdf.DocWriter)(nil)

type canvasDrawCall struct {
	key          string
	x            float64
	y            float64
	width        float64
	height       float64
	canvasWidth  float64
	canvasHeight float64
	capture      *labelTestWriter
}

type canvasTestWriter struct {
	labelTestWriter
	drawCalls []canvasDrawCall
}

func (w *canvasTestWriter) DrawCanvas(key string, x, y, width, height, canvasWidth, canvasHeight float64, draw func(any) error) error {
	capture := &labelTestWriter{t: w.t}
	if draw != nil {
		if err := draw(capture); err != nil {
			return err
		}
	}
	w.drawCalls = append(w.drawCalls, canvasDrawCall{
		key:          key,
		x:            x,
		y:            y,
		width:        width,
		height:       height,
		canvasWidth:  canvasWidth,
		canvasHeight: canvasHeight,
		capture:      capture,
	})
	return nil
}

func parseCanvasDoc(t *testing.T, input string) *Doc {
	t.Helper()
	doc, err := Parse([]byte(input))
	if err != nil {
		t.Fatal(err)
	}
	return doc
}

func renderCanvasPDF(t *testing.T, input string) string {
	t.Helper()
	doc := parseCanvasDoc(t, input)
	writer := newLTMLPDFWriter(t)
	if err := doc.Print(writer); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if _, err := writer.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}

func captureTexts(w *labelTestWriter) string {
	if w == nil {
		return ""
	}
	var b strings.Builder
	for _, rt := range w.printed {
		if rt != nil {
			b.WriteString(rt.String())
		}
	}
	for _, text := range w.plainPrinted {
		b.WriteString(text)
	}
	return b.String()
}

func TestParse_CanvasMustBeDirectChildOfDocument(t *testing.T) {
	_, err := Parse([]byte(`
<ltml>
  <page>
    <canvas key="board" width="120" height="80" />
  </page>
</ltml>`))
	if err == nil || !strings.Contains(err.Error(), "direct child") {
		t.Fatalf("Parse error = %v, want direct-child canvas error", err)
	}
}

func TestParse_CanvasAndDrawRequireKeysAndCanvasSize(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{
			name: "<canvas> key",
			input: `
<ltml>
  <canvas width="120" height="80" />
  <page />
</ltml>`,
			want: "requires a key",
		},
		{
			name: "<canvas> width",
			input: `
<ltml>
  <canvas key="board" height="80" />
  <page />
</ltml>`,
			want: "requires a positive width",
		},
		{
			name: "<canvas> height",
			input: `
<ltml>
  <canvas key="board" width="120" />
  <page />
</ltml>`,
			want: "requires a positive height",
		},
		{
			name: "<draw> key",
			input: `
<ltml>
  <page><draw /></page>
  <canvas key="board" width="120" height="80" />
</ltml>`,
			want: "<draw> requires a key",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			_, err := Parse([]byte(tc.input))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("Parse error = %v, want %q", err, tc.want)
			}
		})
	}
}

func TestParse_DuplicateCanvasKeysRejected(t *testing.T) {
	_, err := Parse([]byte(`
<ltml>
  <canvas key="board" width="120" height="80" />
  <canvas key="board" width="120" height="80" />
  <page />
</ltml>`))
	if err == nil || !strings.Contains(err.Error(), "duplicate canvas key") {
		t.Fatalf("Parse error = %v, want duplicate key error", err)
	}
}

func TestStdDocument_DrawCanvasOrderIndependentLookup(t *testing.T) {
	doc := parseCanvasDoc(t, `
<ltml units="pt">
  <page>
    <draw key="board" />
  </page>
  <canvas key="board" width="120" height="80" />
</ltml>`)

	writer := &canvasTestWriter{labelTestWriter: labelTestWriter{t: t}}
	if err := doc.Print(writer); err != nil {
		t.Fatal(err)
	}
	if len(writer.drawCalls) != 1 {
		t.Fatalf("draw call count = %d, want 1", len(writer.drawCalls))
	}
	call := writer.drawCalls[0]
	if call.key != "board" || call.canvasWidth != 120 || call.canvasHeight != 80 {
		t.Fatalf("draw call = %#v, want board 120x80", call)
	}
}

func TestStdDraw_PlacementSizingUsesNaturalSizeAndAspectRatio(t *testing.T) {
	doc := parseCanvasDoc(t, `
<ltml units="pt">
  <page layout="vbox">
    <draw key="board" />
    <draw key="board" width="60" />
    <draw key="board" height="20" />
    <draw key="board" width="50" height="50" />
  </page>
  <canvas key="board" width="120" height="80" />
</ltml>`)

	writer := &canvasTestWriter{labelTestWriter: labelTestWriter{t: t}}
	if err := doc.Print(writer); err != nil {
		t.Fatal(err)
	}
	if len(writer.drawCalls) != 4 {
		t.Fatalf("draw call count = %d, want 4", len(writer.drawCalls))
	}

	got := [][2]float64{
		{writer.drawCalls[0].width, writer.drawCalls[0].height},
		{writer.drawCalls[1].width, writer.drawCalls[1].height},
		{writer.drawCalls[2].width, writer.drawCalls[2].height},
		{writer.drawCalls[3].width, writer.drawCalls[3].height},
	}
	want := [][2]float64{
		{120, 80},
		{60, 40},
		{30, 20},
		{50, 50},
	}
	if len(got) != len(want) {
		t.Fatalf("unexpected sizing results: %#v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("draw %d size = %v, want %v", i, got[i], want[i])
		}
	}
}

func TestStdDraw_CanvasCaptureUsesLocalScopeAndCanvasChrome(t *testing.T) {
	doc := parseCanvasDoc(t, `
<ltml units="pt">
  <page>
    <draw key="badge" alt="Badge" />
  </page>
  <canvas key="badge" width="120" height="60" padding="8" fill.color="#ffeeaa" border.color="black" border.width="1">
    <font id="canvasfont" name="Helvetica" size="9" />
    <label left="0" top="0" width="40" height="12" font="canvasfont">Local</label>
  </canvas>
</ltml>`)

	writer := &canvasTestWriter{labelTestWriter: labelTestWriter{t: t}}
	if err := doc.Print(writer); err != nil {
		t.Fatal(err)
	}
	if len(writer.drawCalls) != 1 {
		t.Fatalf("draw call count = %d, want 1", len(writer.drawCalls))
	}
	call := writer.drawCalls[0]
	if call.canvasWidth != 120 || call.canvasHeight != 60 {
		t.Fatalf("canvas size = %vx%v, want 120x60", call.canvasWidth, call.canvasHeight)
	}
	if captureTexts(call.capture) != "Local" {
		t.Fatalf("capture text = %q, want Local", captureTexts(call.capture))
	}
	if call.capture.fontSize != 9 {
		t.Fatalf("capture fontSize = %v, want 9", call.capture.fontSize)
	}
	if len(call.capture.rectPages) == 0 {
		t.Fatalf("expected canvas fill/border rectangle output, got %#v", call.capture.rectPages)
	}
}

func TestStdDocument_Print_RejectsUnknownCanvasKey(t *testing.T) {
	doc := parseCanvasDoc(t, `
<ltml>
  <page><draw key="missing" /></page>
</ltml>`)

	err := doc.Print(&labelTestWriter{t: t})
	if err == nil || !strings.Contains(err.Error(), "missing canvas definition") {
		t.Fatalf("Print error = %v, want missing canvas definition", err)
	}
}

func TestDrawCanvas_ReusesMemoizedFormAcrossPagesAndSizes(t *testing.T) {
	pdfText := renderCanvasPDF(t, `
<ltml units="pt">
  <canvas key="board" width="160" height="160">
    <circle left="20" top="20" width="120" height="120" border.color="#1f4ea8" border.width="2" />
    <circle left="52" top="52" width="56" height="56" fill.color="#cfe4ff" border.color="#1f4ea8" border.width="1.5" />
    <label left="66" top="68">92</label>
  </canvas>
  <page layout="absolute">
    <draw key="board" left="36" top="48" />
    <draw key="board" left="240" top="48" width="80" />
  </page>
  <page layout="absolute">
    <draw key="board" left="72" top="96" width="56" />
  </page>
</ltml>`)

	if count := strings.Count(pdfText, "/Subtype /Form"); count != 1 {
		t.Fatalf("expected one reusable canvas form, got %d\n%s", count, pdfText)
	}
	if count := strings.Count(pdfText, "/Mf0 Do"); count != 3 {
		t.Fatalf("expected three placements of the shared canvas form, got %d\n%s", count, pdfText)
	}
}

func TestDrawCanvas_SuppressesInnerLinksPageNumbersAndTaggedContent(t *testing.T) {
	pdfText := renderCanvasPDF(t, `
<ltml ua="true" units="pt">
  <canvas key="badge" width="180" height="60">
    <label id="inner">Page <pageno/></label>
    <p top="26"><a target="inner">Jump</a></p>
  </canvas>
  <page layout="absolute">
    <p><pageno hidden="true" start="777" /></p>
    <draw key="badge" left="36" top="48" alt="Badge" />
  </page>
</ltml>`)

	if strings.Contains(pdfText, "/Subtype /Link") {
		t.Fatalf("canvas capture should suppress inner link annotations, got:\n%s", pdfText)
	}
	if strings.Contains(pdfText, "/S /GoTo") {
		t.Fatalf("canvas capture should suppress inner target annotations, got:\n%s", pdfText)
	}
	if strings.Contains(pdfText, "(777) Tj") {
		t.Fatalf("canvas capture should suppress inner page-number text, got:\n%s", pdfText)
	}
	if !strings.Contains(pdfText, "/StructTreeRoot") || !strings.Contains(pdfText, "/S /Figure") {
		t.Fatalf("expected tagged outer draw placement, got:\n%s", pdfText)
	}
	if count := strings.Count(pdfText, "BDC\n"); count != 1 {
		t.Fatalf("expected one outer tagged placement and no inner tagged text, got %d\n%s", count, pdfText)
	}
}
