// Copyright 2026 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

// glyphRecorder accumulates the glyph IDs used for a single font face
// during document rendering, along with the Unicode codepoints they represent.
// It is used in Unicode mode to build the /W width array and the ToUnicode
// CMap stream at document close.
type glyphRecorder struct {
	keyToCID map[glyphUseKey]uint16
	cidUses  map[uint16]glyphUse
	nextCID  uint16
}

type glyphUseKey struct {
	glyphID uint16
	text    string
}

type glyphUse struct {
	glyphID uint16
	runes   []rune
}

func newGlyphRecorder() *glyphRecorder {
	return &glyphRecorder{
		keyToCID: make(map[glyphUseKey]uint16),
		cidUses:  make(map[uint16]glyphUse),
		nextCID:  1, // reserve CID 0 for .notdef
	}
}

// record notes that glyphID was used to render the given rune.
func (gr *glyphRecorder) record(glyphID uint16, r rune) uint16 {
	return gr.recordRunes(glyphID, []rune{r})
}

// recordRunes notes that glyphID was used to render the given Unicode sequence.
func (gr *glyphRecorder) recordRunes(glyphID uint16, runes []rune) uint16 {
	return gr.cidFor(glyphID, runes)
}

// recordEmpty notes that glyphID was used during rendering but does not
// correspond to authored Unicode text (for example, a secondary glyph in a
// shaped cluster). The ToUnicode CMap still gets an entry for the CID, but the
// destination sequence is empty.
func (gr *glyphRecorder) recordEmpty(glyphID uint16) uint16 {
	return gr.cidFor(glyphID, nil)
}

func (gr *glyphRecorder) cidFor(glyphID uint16, runes []rune) uint16 {
	key := glyphUseKey{
		glyphID: glyphID,
		text:    string(runes),
	}
	if cid, ok := gr.keyToCID[key]; ok {
		return cid
	}
	cid := gr.nextCID
	gr.nextCID++
	gr.keyToCID[key] = cid
	gr.cidUses[cid] = glyphUse{
		glyphID: glyphID,
		runes:   append([]rune(nil), runes...),
	}
	return cid
}

// mapping returns a copy of the CID → Unicode-sequence map.
func (gr *glyphRecorder) mapping() map[uint16][]rune {
	m := make(map[uint16][]rune, len(gr.cidUses))
	for cid, use := range gr.cidUses {
		m[cid] = append([]rune(nil), use.runes...)
	}
	return m
}

// glyphIDs returns all source glyph IDs used during rendering.
func (gr *glyphRecorder) glyphIDs() []uint16 {
	seen := make(map[uint16]struct{}, len(gr.cidUses))
	glyphIDs := make([]uint16, 0, len(gr.cidUses))
	for _, use := range gr.cidUses {
		if _, ok := seen[use.glyphID]; ok {
			continue
		}
		seen[use.glyphID] = struct{}{}
		glyphIDs = append(glyphIDs, use.glyphID)
	}
	return glyphIDs
}

// cids returns all PDF character codes used during rendering.
func (gr *glyphRecorder) cids() []uint16 {
	cids := make([]uint16, 0, len(gr.cidUses))
	for cid := range gr.cidUses {
		cids = append(cids, cid)
	}
	return cids
}

// glyphIDForCID reports the source glyph ID for a emitted PDF CID.
func (gr *glyphRecorder) glyphIDForCID(cid uint16) uint16 {
	if use, ok := gr.cidUses[cid]; ok {
		return use.glyphID
	}
	return 0
}
