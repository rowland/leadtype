// Copyright 2026 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

// glyphRecorder accumulates the glyph IDs used for a single font face
// during document rendering, along with the Unicode codepoints they represent.
// It is used in Unicode mode to build the /W width array and the ToUnicode
// CMap stream at document close.
type glyphRecorder struct {
	glyphToRunes map[uint16][]rune
	usedGlyphs   map[uint16]struct{}
}

func newGlyphRecorder() *glyphRecorder {
	return &glyphRecorder{
		glyphToRunes: make(map[uint16][]rune),
		usedGlyphs:   make(map[uint16]struct{}),
	}
}

// record notes that glyphID was used to render the given rune.
// If glyphID is already recorded it is silently ignored (first rune wins).
func (gr *glyphRecorder) record(glyphID uint16, r rune) {
	gr.recordRunes(glyphID, []rune{r})
}

// recordRunes notes that glyphID was used to render the given Unicode sequence.
// If glyphID is already recorded it is silently ignored (first sequence wins).
func (gr *glyphRecorder) recordRunes(glyphID uint16, runes []rune) {
	gr.use(glyphID)
	if _, ok := gr.glyphToRunes[glyphID]; !ok {
		gr.glyphToRunes[glyphID] = append([]rune(nil), runes...)
	}
}

// use notes that glyphID was used during rendering even if it has no direct
// Unicode mapping in the ToUnicode CMap (for example, a secondary glyph in a
// shaped cluster).
func (gr *glyphRecorder) use(glyphID uint16) {
	gr.usedGlyphs[glyphID] = struct{}{}
}

// mapping returns a copy of the glyph-ID → Unicode-sequence map.
func (gr *glyphRecorder) mapping() map[uint16][]rune {
	m := make(map[uint16][]rune, len(gr.glyphToRunes))
	for k, v := range gr.glyphToRunes {
		m[k] = append([]rune(nil), v...)
	}
	return m
}

// glyphIDs returns all glyph IDs used during rendering, whether or not they
// carry a Unicode mapping in the ToUnicode CMap.
func (gr *glyphRecorder) glyphIDs() []uint16 {
	glyphIDs := make([]uint16, 0, len(gr.usedGlyphs))
	for gid := range gr.usedGlyphs {
		glyphIDs = append(glyphIDs, gid)
	}
	return glyphIDs
}
