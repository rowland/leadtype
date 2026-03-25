// Copyright 2024 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

// glyphRecorder accumulates the glyph IDs used for a single font face
// during document rendering, along with the Unicode codepoints they represent.
// It is used in Unicode mode to build the /W width array and the ToUnicode
// CMap stream at document close.
type glyphRecorder struct {
	glyphToRunes map[uint16][]rune
}

func newGlyphRecorder() *glyphRecorder {
	return &glyphRecorder{glyphToRunes: make(map[uint16][]rune)}
}

// record notes that glyphID was used to render the given rune.
// If glyphID is already recorded it is silently ignored (first rune wins).
func (gr *glyphRecorder) record(glyphID uint16, r rune) {
	gr.recordRunes(glyphID, []rune{r})
}

// recordRunes notes that glyphID was used to render the given Unicode sequence.
// If glyphID is already recorded it is silently ignored (first sequence wins).
func (gr *glyphRecorder) recordRunes(glyphID uint16, runes []rune) {
	if _, ok := gr.glyphToRunes[glyphID]; !ok {
		gr.glyphToRunes[glyphID] = append([]rune(nil), runes...)
	}
}

// mapping returns a copy of the glyph-ID → Unicode-sequence map.
func (gr *glyphRecorder) mapping() map[uint16][]rune {
	m := make(map[uint16][]rune, len(gr.glyphToRunes))
	for k, v := range gr.glyphToRunes {
		m[k] = append([]rune(nil), v...)
	}
	return m
}
