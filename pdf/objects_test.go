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
