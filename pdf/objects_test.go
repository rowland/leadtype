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
