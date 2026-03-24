// Copyright 2024 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import (
	"fmt"
	"io"
	"sort"
)

// ── CIDSystemInfo ─────────────────────────────────────────────────────────────

type cidSystemInfo struct{}

func (c *cidSystemInfo) write(w io.Writer) {
	fmt.Fprintf(w, "<< /Registry (Adobe) /Ordering (Identity) /Supplement 0 >> ")
}

// ── /W sparse width array ─────────────────────────────────────────────────────

// mostCommonWidth returns the most frequently occurring width value in
// glyphWidths. It is used to pick /DW so that /W only needs to list
// exceptions. Returns 0 if the map is empty.
func mostCommonWidth(glyphWidths map[uint16]int) int {
	counts := make(map[int]int, len(glyphWidths))
	for _, w := range glyphWidths {
		counts[w]++
	}
	best, bestCount := 0, 0
	for w, c := range counts {
		if c > bestCount {
			best, bestCount = w, c
		}
	}
	return best
}

// buildCIDWidthArray constructs a /W array in the sparse PDF format:
//
//	[firstGlyph [w0 w1 …] nextGlyph [w …] …]
//
// Consecutive glyph IDs with widths are grouped into a single sub-array.
// Entries whose width equals defaultWidth are omitted (they are covered by
// the /DW entry). Pass 0 for defaultWidth to include all entries.
// glyphWidths maps glyphID → advance width scaled to 1000-unit PDF space.
func buildCIDWidthArray(glyphWidths map[uint16]int, defaultWidth int) writer {
	if len(glyphWidths) == 0 {
		return array{}
	}
	// Sort glyph IDs, skipping those that match the default width.
	ids := make([]int, 0, len(glyphWidths))
	for id, w := range glyphWidths {
		if w != defaultWidth {
			ids = append(ids, int(id))
		}
	}
	if len(ids) == 0 {
		return array{}
	}
	sort.Ints(ids)

	// Build runs of consecutive IDs.
	type run struct {
		start  int
		widths []int
	}
	var runs []run
	cur := run{start: ids[0], widths: []int{glyphWidths[uint16(ids[0])]}}
	for _, id := range ids[1:] {
		if id == cur.start+len(cur.widths) {
			cur.widths = append(cur.widths, glyphWidths[uint16(id)])
		} else {
			runs = append(runs, cur)
			cur = run{start: id, widths: []int{glyphWidths[uint16(id)]}}
		}
	}
	runs = append(runs, cur)

	// Serialise as [start [w…] start [w…] …]
	var elems []writer
	for _, r := range runs {
		elems = append(elems, integer(r.start))
		sub := make(array, len(r.widths))
		for i, w := range r.widths {
			sub[i] = integer(w)
		}
		elems = append(elems, sub)
	}
	return array(elems)
}

// ── CIDFont (descendant of Type0) ────────────────────────────────────────────

type cidFont struct {
	dictionaryObject
}

func (f *cidFont) init(seq, gen int,
	baseFont string,
	fontDescriptor *fontDescriptor,
	defaultWidth int,
	widths writer) *cidFont {
	f.dictionaryObject.init(seq, gen)
	f.dict["Type"] = name("Font")
	f.dict["Subtype"] = name("CIDFontType2")
	f.dict["BaseFont"] = name(baseFont)
	f.dict["CIDSystemInfo"] = &cidSystemInfo{}
	f.dict["FontDescriptor"] = &indirectObjectRef{fontDescriptor}
	f.dict["DW"] = integer(defaultWidth)
	f.dict["W"] = widths
	f.dict["CIDToGIDMap"] = name("Identity")
	return f
}

func newCIDFont(seq, gen int,
	baseFont string,
	fontDescriptor *fontDescriptor,
	defaultWidth int,
	widths writer) *cidFont {
	return new(cidFont).init(seq, gen, baseFont, fontDescriptor, defaultWidth, widths)
}

// setWidths replaces the /W entry after initial construction, used when
// glyph usage is not known until document close.
func (f *cidFont) setWidths(w writer) {
	f.dict["W"] = w
}

// setDefaultWidth updates the /DW entry with the most-common glyph width,
// called at document close once all glyph usage is known.
func (f *cidFont) setDefaultWidth(w int) {
	f.dict["DW"] = integer(w)
}

// setBaseFont updates /BaseFont, used to apply the subset tag at close.
func (f *cidFont) setBaseFont(n string) {
	f.dict["BaseFont"] = name(n)
}

// ── Type0 (composite) font ────────────────────────────────────────────────────

type type0Font struct {
	dictionaryObject
}

func (f *type0Font) init(seq, gen int,
	baseFont string,
	descendant *cidFont) *type0Font {
	f.dictionaryObject.init(seq, gen)
	f.dict["Type"] = name("Font")
	f.dict["Subtype"] = name("Type0")
	f.dict["BaseFont"] = name(baseFont)
	f.dict["Encoding"] = name("Identity-H")
	f.dict["DescendantFonts"] = array{&indirectObjectRef{descendant}}
	return f
}

func newType0Font(seq, gen int, baseFont string, descendant *cidFont) *type0Font {
	return new(type0Font).init(seq, gen, baseFont, descendant)
}

// setBaseFont updates /BaseFont, used to apply the subset tag at close.
func (f *type0Font) setBaseFont(n string) {
	f.dict["BaseFont"] = name(n)
}

func (f *type0Font) setToUnicode(ref *indirectObjectRef) {
	f.dict["ToUnicode"] = ref
}
