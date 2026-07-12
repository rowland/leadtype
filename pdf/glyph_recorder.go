// Copyright 2026 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import "fmt"

// glyphRecorder keeps three namespaces separate for one composite PDF font:
//
//   - emitted character code: the uint16 written by text-showing operators;
//   - source GID: the glyph selected by cmap or shaping in the source font;
//   - font CID: the identifier stored in a CID-keyed CFF charset.
//
// These values happen to coincide for some fonts, which is why confusing them
// can survive for years. CID-keyed CFF fonts may have non-identity GID-to-CID
// mappings. For ordinary composite fonts LeadType may allocate sequential
// codes; for CID-keyed CFF it emits the font CID directly so Identity-H works
// across PDF consumers. At document close this data drives ToUnicode, widths,
// CIDSet, and font subsetting.
type glyphRecorder struct {
	keyToCID  map[glyphUseKey]uint16
	cidUses   map[uint16]glyphUse
	nextCID   uint32
	identity  bool
	cidForGID func(uint16) (uint16, bool)
	err       error
}

// glyphUseKey makes ToUnicode semantics part of code allocation. Reusing a GID
// is safe only when it represents the same authored text sequence.
type glyphUseKey struct {
	glyphID uint16
	text    string
}

// glyphUse is the complete meaning of one emitted character code.
type glyphUse struct {
	glyphID uint16
	fontCID uint16
	runes   []rune
}

func newGlyphRecorder(identity bool) *glyphRecorder {
	return &glyphRecorder{
		keyToCID: make(map[glyphUseKey]uint16),
		cidUses:  make(map[uint16]glyphUse),
		nextCID:  1, // reserve character code 0 for .notdef
		identity: identity,
	}
}

func newCIDKeyedGlyphRecorder(cidForGID func(uint16) (uint16, bool)) *glyphRecorder {
	gr := newGlyphRecorder(false)
	gr.cidForGID = cidForGID
	return gr
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

// cidFor returns the character code to emit. CID-keyed CFF uses the font CID as
// that code, avoiding a custom Encoding CMap that several otherwise capable
// PDF consumers fail to apply to native CFF fonts. A ToUnicode entry can give
// that CID only one meaning; emitters use requiresActualText to preserve less
// common Unicode aliases of the same glyph at the point where they occur.
func (gr *glyphRecorder) cidFor(glyphID uint16, runes []rune) uint16 {
	if gr.identity {
		cid := glyphID
		if use, ok := gr.cidUses[cid]; ok {
			if len(use.runes) == 0 && len(runes) > 0 {
				use.runes = append([]rune(nil), runes...)
				gr.cidUses[cid] = use
			}
			return cid
		}
		gr.cidUses[cid] = glyphUse{
			glyphID: glyphID,
			fontCID: cid,
			runes:   append([]rune(nil), runes...),
		}
		return cid
	}
	key := glyphUseKey{
		glyphID: glyphID,
		text:    string(runes),
	}
	if cid, ok := gr.keyToCID[key]; ok {
		return cid
	}
	if gr.cidForGID != nil {
		fontCID, ok := gr.cidForGID(glyphID)
		if !ok {
			if gr.err == nil {
				gr.err = fmt.Errorf("glyph %d has no CID mapping", glyphID)
			}
			return 0
		}
		if _, used := gr.cidUses[fontCID]; used {
			gr.keyToCID[key] = fontCID
			return fontCID
		}
		gr.keyToCID[key] = fontCID
		gr.cidUses[fontCID] = glyphUse{
			glyphID: glyphID,
			fontCID: fontCID,
			runes:   append([]rune(nil), runes...),
		}
		return fontCID
	}
	if gr.nextCID > 0xffff {
		if gr.err == nil {
			gr.err = fmt.Errorf("composite font uses more than 65535 character codes")
		}
		return 0
	}
	cid := uint16(gr.nextCID)
	gr.nextCID++
	gr.keyToCID[key] = cid
	gr.cidUses[cid] = glyphUse{
		glyphID: glyphID,
		fontCID: cid,
		runes:   append([]rune(nil), runes...),
	}
	return cid
}

// requiresActualText reports whether this occurrence has a different Unicode
// meaning from the one retained in ToUnicode for code. CID-keyed fonts commonly
// share outlines among compatibility characters, spaces, and punctuation.
// Keeping Identity-H makes their normal text selectable in more PDF consumers;
// a small ActualText span is needed only for those ambiguous occurrences.
func (gr *glyphRecorder) requiresActualText(code uint16, runes []rune) bool {
	use, ok := gr.cidUses[code]
	return ok && string(use.runes) != string(runes)
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

// fontCIDForCode maps a text-showing character code to the CID understood by
// the descendant CIDFont. It is used to construct a custom Encoding CMap.
func (gr *glyphRecorder) fontCIDForCode(code uint16) uint16 {
	if use, ok := gr.cidUses[code]; ok {
		return use.fontCID
	}
	return 0
}

func (gr *glyphRecorder) error() error { return gr.err }
