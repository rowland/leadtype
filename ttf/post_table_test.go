package ttf

import (
	"bytes"
	"testing"
)

func TestPostTable_readFormat2Names_InvalidCustomNameIndex(t *testing.T) {
	var table postTable

	// numberOfGlyphs = 1
	// glyphNameIndex[0] = 311 => custom-name slot 53
	// numberNewGlyphs inferred = 1, so slot 53 is out of range.
	data := []byte{
		0x00, 0x01,
		0x01, 0x37,
		0x03, 'b', 'a', 'd',
	}

	if err := table.readFormat2Names(bytes.NewReader(data)); err != nil {
		t.Fatalf("readFormat2Names() error = %v, want nil fallback", err)
	}
	if len(table.names) != 1 {
		t.Fatalf("len(table.names) = %d, want 1", len(table.names))
	}
	if table.names[0] != ".notdef" {
		t.Fatalf("table.names[0] = %q, want fallback .notdef", table.names[0])
	}
}
