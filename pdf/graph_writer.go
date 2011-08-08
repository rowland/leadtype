package pdf

import (
	"io"
	"fmt"
)

type graphWriter struct {
	wr io.Writer
}

func newGraphWriter(wr io.Writer) *graphWriter {
	return &graphWriter{wr}
}

// use nonzero winding number rule
func (gw *graphWriter) clip() {
	fmt.Fprintf(gw.wr, "W\n")
}

func (gw *graphWriter) closePath() {
	fmt.Fprintf(gw.wr, "h\n")
}

func (gw *graphWriter) closePathAndStroke() {
	fmt.Fprintf(gw.wr, "s\n")
}

func (gw *graphWriter) closePathEoFillAndStroke() {
	fmt.Fprintf(gw.wr, "b*\n")
}

func (gw *graphWriter) closePathFillAndStroke() {
	fmt.Fprintf(gw.wr, "b\n")
}

func (gw *graphWriter) concatMatrix(a, b, c, d, x, y float64) {
	fmt.Fprintf(gw.wr, "%s %s %s %s %s %s cm\n", g(a), g(b), g(c), g(d), g(x), g(y))
}

func (gw *graphWriter) curveTo(x1, y1, x2, y2, x3, y3 float64) {
	fmt.Fprintf(gw.wr, "%s %s %s %s %s %s c\n", g(x1), g(y1), g(x2), g(y2), g(x3), g(y3))
}

// use even-odd rule
func (gw *graphWriter) eoClip() {
	fmt.Fprintf(gw.wr, "W*\n")
}

func (gw *graphWriter) eoFill() {
	fmt.Fprintf(gw.wr, "f*\n")
}

func (gw *graphWriter) eoFillAndStroke() {
	fmt.Fprintf(gw.wr, "B*\n")
}

func (gw *graphWriter) fill() {
	fmt.Fprintf(gw.wr, "f\n")
}

func (gw *graphWriter) fillAndStroke() {
	fmt.Fprintf(gw.wr, "B\n")
}

func (gw *graphWriter) lineTo(x, y float64) {
	fmt.Fprintf(gw.wr, "%s %s l\n", g(x), g(y))
}

func (gw *graphWriter) moveTo(x, y float64) {
	fmt.Fprintf(gw.wr, "%s %s m\n", g(x), g(y))
}

func (gw *graphWriter) newPath() {
	fmt.Fprintf(gw.wr, "n\n")
}

func (gw *graphWriter) rectangle(x, y, width, height float64) {
	fmt.Fprintf(gw.wr, "%s %s %s %s re\n", g(x), g(y), g(width), g(height))
}

func (gw *graphWriter) restoreGraphicsState() {
	fmt.Fprintf(gw.wr, "Q\n")
}

func (gw *graphWriter) saveGraphicsState() {
	fmt.Fprintf(gw.wr, "q\n")
}

func (gw *graphWriter) setFlatness(flatness int) {
	fmt.Fprintf(gw.wr, "%d i\n", flatness)
}

func (gw *graphWriter) setLineCapStyle(lineCapStyle int) {
	fmt.Fprintf(gw.wr, "%d J\n", lineCapStyle)
}

func (gw *graphWriter) setLineDashPattern(lineDashPattern string) {
	fmt.Fprintf(gw.wr, "%s d\n", lineDashPattern)
}

func (gw *graphWriter) setLineJoinStyle(lineJoinStyle int) {
	fmt.Fprintf(gw.wr, "%d j\n", lineJoinStyle)
}

func (gw *graphWriter) setLineWidth(lineWidth float64) {
	fmt.Fprintf(gw.wr, "%s w\n", g(lineWidth))
}

func (gw *graphWriter) setMiterLimit(miterLimit float64) {
	fmt.Fprintf(gw.wr, "%s M\n", g(miterLimit))
}

func (gw *graphWriter) stroke() {
	fmt.Fprintf(gw.wr, "S\n")
}
