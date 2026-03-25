package ttf

import (
	"bytes"
	"strings"
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

	err := table.readFormat2Names(bytes.NewReader(data))
	if err == nil {
		t.Fatal("readFormat2Names() error = nil, want non-nil")
	}
	if !strings.Contains(err.Error(), "invalid post format 2 glyph name index") {
		t.Fatalf("unexpected error: %v", err)
	}
}
