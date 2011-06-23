package pdf

import (
	"testing"
	"bytes"
)

func TestHeader(t *testing.T) {
	var buf bytes.Buffer
	h := &Header{}
	h.Write(&buf)

	const expected = "%PDF-1.3\n"
	if buf.String() != expected {
		t.Errorf("Expected %s, got %s", expected, buf.String())
	}
}

func TestInUseXRefEntry(t *testing.T) {
	var buf bytes.Buffer
	e := &InUseXRefEntry{500, 0}
	e.Write(&buf)

	const expected = "0000000500 00000 n\n"
	if buf.String() != expected {
		t.Errorf("Expected %s, got %s", expected, buf.String())
	}
}

func TestPdfInteger(t *testing.T) {
	var buf bytes.Buffer
	pi := PdfInteger(7)
	pi.Write(&buf)

	const expected = "7 "
	if buf.String() != expected {
		t.Errorf("Expected %s, got %s", expected, buf.String())
	}
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

	const expected = "7 8 9 10.5 11.5 "
	if buf.String() != expected {
		t.Errorf("Expected %s, got %s", expected, buf.String())
	}
}

func TestPdfBoolean(t *testing.T) {
	var buf bytes.Buffer
	PdfBoolean(true).Write(&buf)
	PdfBoolean(false).Write(&buf)

	const expected = "true false "
	if buf.String() != expected {
		t.Errorf("Expected %s, got %s", expected, buf.String())
	}
}
