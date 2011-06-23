package pdf

import (
	"testing"
	"bytes"
)

func expect(t *testing.T, expected, actual string) {
	if expected != actual {
		t.Errorf("Expected %s, got %s", expected, actual)
	}
}

func TestHeader(t *testing.T) {
	var buf bytes.Buffer
	h := &Header{}
	h.Write(&buf)

	expect(t, "%PDF-1.3\n", buf.String())
}

func TestInUseXRefEntry(t *testing.T) {
	var buf bytes.Buffer
	e := &InUseXRefEntry{500, 0}
	e.Write(&buf)

	expect(t, "0000000500 00000 n\n", buf.String())
}

func TestPdfInteger(t *testing.T) {
	var buf bytes.Buffer
	pi := PdfInteger(7)
	pi.Write(&buf)

	expect(t, "7 ", buf.String())
}

func TestPdfNumber(t *testing.T) {
	var buf bytes.Buffer
	ni := int(7)
	ni32 := int32(8)
	ni64 := int64(9)
	f32 := float32(10.5)
	f64 := float64(11.5)
	PdfNumber{ni}.Write(&buf)
	PdfNumber{ni32}.Write(&buf)
	PdfNumber{ni64}.Write(&buf)
	PdfNumber{f32}.Write(&buf)
	PdfNumber{f64}.Write(&buf)

	expect(t, "7 8 9 10.5 11.5 ", buf.String())
}

func TestPdfBoolean(t *testing.T) {
	var buf bytes.Buffer
	PdfBoolean(true).Write(&buf)
	PdfBoolean(false).Write(&buf)

	expect(t, "true false ", buf.String())
}

func TestPdfName(t *testing.T) {
	var buf bytes.Buffer
	PdfName("name").Write(&buf)

	expect(t, "/name ", buf.String())
}

func TestRectangle(t *testing.T) {
	var buf bytes.Buffer
	r := &Rectangle{1, 2, 3, 4}
	r.Write(&buf)

	expect(t, "[1 2 3 4 ] ", buf.String())
}
