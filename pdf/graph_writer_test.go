package pdf

import (
	"testing"
	"bytes"
)

func TestGraphWriter_clip(t *testing.T) {
	var buf bytes.Buffer
	gw := newGraphWriter(&buf)
	gw.clip()
	expectS(t, "W\n", buf.String())
}

func TestGraphWriter_closePath(t *testing.T) {
	var buf bytes.Buffer
	gw := newGraphWriter(&buf)
	gw.closePath()
	expectS(t, "h\n", buf.String())
}

func TestGraphWriter_closePathAndStroke(t *testing.T) {
	var buf bytes.Buffer
	gw := newGraphWriter(&buf)
	gw.closePathAndStroke()
	expectS(t, "s\n", buf.String())
}

func TestGraphWriter_closePathEoFillAndStroke(t *testing.T) {
	var buf bytes.Buffer
	gw := newGraphWriter(&buf)
	gw.closePathEoFillAndStroke()
	expectS(t, "b*\n", buf.String())
}

func TestGraphWriter_closePathFillAndStroke(t *testing.T) {
	var buf bytes.Buffer
	gw := newGraphWriter(&buf)
	gw.closePathFillAndStroke()
	expectS(t, "b\n", buf.String())
}

func TestGraphWriter_concatMatrix(t *testing.T) {
	var buf bytes.Buffer
	gw := newGraphWriter(&buf)
	gw.concatMatrix(1.1, 2.2, 3.3, 4.4, 5.5, 6.6)
	expectS(t, "1.1 2.2 3.3 4.4 5.5 6.6 cm\n", buf.String())
}

func TestGraphWriter_curveTo(t *testing.T) {
	var buf bytes.Buffer
	gw := newGraphWriter(&buf)
	gw.curveTo(1.1, 2.2, 3.3, 4.4, 5.5, 6.6)
	expectS(t, "1.1 2.2 3.3 4.4 5.5 6.6 c\n", buf.String())
}

func TestGraphWriter_eoClip(t *testing.T) {
	var buf bytes.Buffer
	gw := newGraphWriter(&buf)
	gw.eoClip()
	expectS(t, "W*\n", buf.String())
}

func TestGraphWriter_eoFill(t *testing.T) {
	var buf bytes.Buffer
	gw := newGraphWriter(&buf)
	gw.eoFill()
	expectS(t, "f*\n", buf.String())
}

func TestGraphWriter_eoFillAndStroke(t *testing.T) {
	var buf bytes.Buffer
	gw := newGraphWriter(&buf)
	gw.eoFillAndStroke()
	expectS(t, "B*\n", buf.String())
}

func TestGraphWriter_fill(t *testing.T) {
	var buf bytes.Buffer
	gw := newGraphWriter(&buf)
	gw.fill()
	expectS(t, "f\n", buf.String())
}

func TestGraphWriter_fillAndStroke(t *testing.T) {
	var buf bytes.Buffer
	gw := newGraphWriter(&buf)
	gw.fillAndStroke()
	expectS(t, "B\n", buf.String())
}

func TestGraphWriter_lineTo(t *testing.T) {
	var buf bytes.Buffer
	gw := newGraphWriter(&buf)
	gw.lineTo(5.55, 4)
	expectS(t, "5.55 4 l\n", buf.String())
}

func TestGraphWriter_moveTo(t *testing.T) {
	var buf bytes.Buffer
	gw := newGraphWriter(&buf)
	gw.moveTo(4, 5.55)
	expectS(t, "4 5.55 m\n", buf.String())
}

func TestGraphWriter_newPath(t *testing.T) {
	var buf bytes.Buffer
	gw := newGraphWriter(&buf)
	gw.newPath()
	expectS(t, "n\n", buf.String())
}

func TestGraphWriter_rectangle(t *testing.T) {
	var buf bytes.Buffer
	gw := newGraphWriter(&buf)
	gw.rectangle(5.5, 5.5, 4, 6)
	expectS(t, "5.5 5.5 4 6 re\n", buf.String())
}

func TestGraphWriter_restoreGraphicsState(t *testing.T) {
	var buf bytes.Buffer
	gw := newGraphWriter(&buf)
	gw.restoreGraphicsState()
	expectS(t, "Q\n", buf.String())
}

func TestGraphWriter_saveGraphicsState(t *testing.T) {
	var buf bytes.Buffer
	gw := newGraphWriter(&buf)
	gw.saveGraphicsState()
	expectS(t, "q\n", buf.String())
}

func TestGraphWriter_setFlatness(t *testing.T) {
	var buf bytes.Buffer
	gw := newGraphWriter(&buf)
	gw.setFlatness(50)
	expectS(t, "50 i\n", buf.String())
}

func TestGraphWriter_setLineCapStyle(t *testing.T) {
	var buf bytes.Buffer
	gw := newGraphWriter(&buf)
	gw.setLineCapStyle(0)
	expectS(t, "0 J\n", buf.String())
}

func TestGraphWriter_setLineDashPattern(t *testing.T) {
	var buf bytes.Buffer
	gw := newGraphWriter(&buf)
	gw.setLineDashPattern("[2 3] 11")
	expectS(t, "[2 3] 11 d\n", buf.String())
}

func TestGraphWriter_setLineJoinStyle(t *testing.T) {
	var buf bytes.Buffer
	gw := newGraphWriter(&buf)
	gw.setLineJoinStyle(0)
	expectS(t, "0 j\n", buf.String())
}

func TestGraphWriter_setLineWidth(t *testing.T) {
	var buf bytes.Buffer
	gw := newGraphWriter(&buf)
	gw.setLineWidth(3)
	expectS(t, "3 w\n", buf.String())
}

func TestGraphWriter_setMiterLimit(t *testing.T) {
	var buf bytes.Buffer
	gw := newGraphWriter(&buf)
	gw.setMiterLimit(3.6)
	expectS(t, "3.6 M\n", buf.String())
}

func TestGraphWriter_stroke(t *testing.T) {
	var buf bytes.Buffer
	gw := newGraphWriter(&buf)
	gw.stroke()
	expectS(t, "S\n", buf.String())
}

func TestMakeLineDashPattern(t *testing.T) {
	expectS(t, "[1 2 3] 2", makeLineDashPattern([]int{1, 2, 3}, 2))
}


