package ttf

import (
	"bytes"
	"io"
	"testing"
)

type stubGlyphIndexer struct{}

func (stubGlyphIndexer) init(file io.Reader) error    { return nil }
func (stubGlyphIndexer) glyphIndex(codepoint int) int { return codepoint }
func (stubGlyphIndexer) write(wr io.Writer)           {}

func TestCmapEncodingRecord_ReadMapping_Format14IsSkippable(t *testing.T) {
	rec := &cmapEncodingRecord{}
	err := rec.readMapping(bytes.NewReader([]byte{0x00, 0x0E}))
	if err == nil {
		t.Fatal("readMapping() error = nil, want unsupported format error")
	}
	if !isSkippableCmapFormatError(err) {
		t.Fatalf("expected skippable unsupported-format error, got %v", err)
	}
}

func TestCmapTable_PreferredRecord_PrefersMicrosoftUCS4(t *testing.T) {
	want := stubGlyphIndexer{}
	table := cmapTable{
		encodingRecords: []cmapEncodingRecord{
			{platformID: MicrosoftPlatformID, platformSpecificID: UCS4PlatformSpecificID, format: 12, glyphIndexer: want},
			{platformID: MacintoshPlatformID, platformSpecificID: 0, format: 0, glyphIndexer: stubGlyphIndexer{}},
		},
	}

	got := table.preferredRecord()
	if got == nil {
		t.Fatal("preferredRecord() = nil, want non-nil")
	}
	if got.platformID != MicrosoftPlatformID || got.platformSpecificID != UCS4PlatformSpecificID {
		t.Fatalf("preferredRecord() = platform %d/%d, want Microsoft UCS4", got.platformID, got.platformSpecificID)
	}
}

func TestCmapTable_PreferredRecord_SkipsNilUnsupportedRecords(t *testing.T) {
	want := stubGlyphIndexer{}
	table := cmapTable{
		encodingRecords: []cmapEncodingRecord{
			{platformID: UnicodePlatformID, platformSpecificID: Unicode2FullPlatformSpecificID, format: 14, glyphIndexer: nil},
			{platformID: MicrosoftPlatformID, platformSpecificID: UCS2PlatformSpecificID, format: 4, glyphIndexer: want},
		},
	}

	got := table.preferredRecord()
	if got == nil {
		t.Fatal("preferredRecord() = nil, want non-nil")
	}
	if got.platformID != MicrosoftPlatformID || got.platformSpecificID != UCS2PlatformSpecificID {
		t.Fatalf("preferredRecord() picked platform %d/%d, want Microsoft UCS2 fallback", got.platformID, got.platformSpecificID)
	}
}

func TestCmapTable_PreferredRecord_AcceptsUnicodeDefaultEncoding(t *testing.T) {
	want := stubGlyphIndexer{}
	table := cmapTable{
		encodingRecords: []cmapEncodingRecord{
			{platformID: UnicodePlatformID, platformSpecificID: DefaultPlatformSpecificID, format: 4, glyphIndexer: want},
		},
	}

	got := table.preferredRecord()
	if got == nil {
		t.Fatal("preferredRecord() = nil, want non-nil")
	}
	if got.platformID != UnicodePlatformID || got.platformSpecificID != DefaultPlatformSpecificID {
		t.Fatalf("preferredRecord() = platform %d/%d, want Unicode default encoding", got.platformID, got.platformSpecificID)
	}
}

func TestCmapTable_PreferredRecord_FallsBackToAnyParsedIndexer(t *testing.T) {
	want := stubGlyphIndexer{}
	table := cmapTable{
		encodingRecords: []cmapEncodingRecord{
			{platformID: 5, platformSpecificID: 7, format: 6, glyphIndexer: want},
		},
	}

	got := table.preferredRecord()
	if got == nil {
		t.Fatal("preferredRecord() = nil, want non-nil")
	}
	if got.platformID != 5 || got.platformSpecificID != 7 {
		t.Fatalf("preferredRecord() = platform %d/%d, want fallback record 5/7", got.platformID, got.platformSpecificID)
	}
}
