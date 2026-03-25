// Copyright 2024 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import (
	"fmt"
	"sort"
	"strings"
	"unicode/utf16"
)

// toUnicodeCMapData builds the byte content of a ToUnicode CMap stream for a
// simple (8-bit) font. encoding is a 256-element slice where encoding[b] is the
// Unicode rune that byte value b maps to; a zero rune means the byte is unmapped.
//
// The PDF spec limits each beginbfchar/endbfchar block to 100 entries.
func toUnicodeCMapData(encoding []rune) []byte {
	// Collect the non-zero mappings.
	type entry struct {
		b rune
		u rune
	}
	var entries []entry
	for b, u := range encoding {
		if u != 0 {
			entries = append(entries, entry{rune(b), u})
		}
	}

	var sb strings.Builder
	sb.WriteString("/CIDInit /ProcSet findresource begin\n")
	sb.WriteString("12 dict begin\n")
	sb.WriteString("begincmap\n")
	sb.WriteString("/CIDSystemInfo << /Registry (Adobe) /Ordering (UCS) /Supplement 0 >> def\n")
	sb.WriteString("/CMapName /Adobe-Identity-UCS def\n")
	sb.WriteString("/CMapType 1 def\n")
	sb.WriteString("1 begincodespacerange\n")
	sb.WriteString("<00> <FF>\n")
	sb.WriteString("endcodespacerange\n")

	// Emit in blocks of at most 100 entries (PDF spec §5.9.2).
	const blockSize = 100
	for i := 0; i < len(entries); i += blockSize {
		end := i + blockSize
		if end > len(entries) {
			end = len(entries)
		}
		block := entries[i:end]
		fmt.Fprintf(&sb, "%d beginbfchar\n", len(block))
		for _, e := range block {
			fmt.Fprintf(&sb, "<%02X> <%04X>\n", e.b, e.u)
		}
		sb.WriteString("endbfchar\n")
	}

	sb.WriteString("endcmap\n")
	sb.WriteString("CMap end\n")
	sb.WriteString("end\n")
	return []byte(sb.String())
}

// toUnicodeCMapDataComposite builds the ToUnicode CMap stream for a composite
// (Type0 / CIDFontType2) font. glyphToRunes maps each glyph ID used in the
// document back to its Unicode sequence. The codespace range is <0000>–<FFFF>.
//
// The PDF spec limits each beginbfchar/endbfchar block to 100 entries.
func toUnicodeCMapDataComposite(glyphToRunes map[uint16][]rune) []byte {
	type entry struct {
		glyphID uint16
		runes   []rune
	}
	entries := make([]entry, 0, len(glyphToRunes))
	for gid, runes := range glyphToRunes {
		entries = append(entries, entry{gid, append([]rune(nil), runes...)})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].glyphID < entries[j].glyphID })

	var sb strings.Builder
	sb.WriteString("/CIDInit /ProcSet findresource begin\n")
	sb.WriteString("12 dict begin\n")
	sb.WriteString("begincmap\n")
	sb.WriteString("/CIDSystemInfo << /Registry (Adobe) /Ordering (UCS) /Supplement 0 >> def\n")
	sb.WriteString("/CMapName /Adobe-Identity-UCS def\n")
	sb.WriteString("/CMapType 2 def\n")
	sb.WriteString("1 begincodespacerange\n")
	sb.WriteString("<0000> <FFFF>\n")
	sb.WriteString("endcodespacerange\n")

	const blockSize = 100
	for i := 0; i < len(entries); i += blockSize {
		end := i + blockSize
		if end > len(entries) {
			end = len(entries)
		}
		block := entries[i:end]
		fmt.Fprintf(&sb, "%d beginbfchar\n", len(block))
		for _, e := range block {
			fmt.Fprintf(&sb, "<%04X> <%s>\n", e.glyphID, compositeDestinationHex(e.runes))
		}
		sb.WriteString("endbfchar\n")
	}

	sb.WriteString("endcmap\n")
	sb.WriteString("CMap end\n")
	sb.WriteString("end\n")
	return []byte(sb.String())
}

func compositeDestinationHex(runes []rune) string {
	var sb strings.Builder
	for _, unit := range utf16.Encode(runes) {
		fmt.Fprintf(&sb, "%04X", unit)
	}
	return sb.String()
}
