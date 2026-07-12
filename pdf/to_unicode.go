// Copyright 2026 Brent Rowland.
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
// (Type0 / CIDFontType2) font. cidToRunes maps each emitted CID used in the
// document back to its Unicode sequence; a nil or empty slice emits an empty
// destination (`<>`). The codespace range is <0000>–<FFFF>.
//
// The PDF spec limits each beginbfchar/endbfchar block to 100 entries.
func toUnicodeCMapDataComposite(cidToRunes map[uint16][]rune) []byte {
	type entry struct {
		cid   uint16
		runes []rune
	}
	entries := make([]entry, 0, len(cidToRunes))
	for cid, runes := range cidToRunes {
		entries = append(entries, entry{cid, append([]rune(nil), runes...)})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].cid < entries[j].cid })

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
			fmt.Fprintf(&sb, "<%04X> <%s>\n", e.cid, compositeDestinationHex(e.runes))
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

func cidToGIDMapData(cidToGID map[uint16]uint16) []byte {
	if len(cidToGID) == 0 {
		return nil
	}
	maxCID := uint16(0)
	for cid := range cidToGID {
		if cid > maxCID {
			maxCID = cid
		}
	}
	data := make([]byte, int(maxCID+1)*2)
	for cid, gid := range cidToGID {
		offset := int(cid) * 2
		data[offset] = byte(gid >> 8)
		data[offset+1] = byte(gid)
	}
	return data
}

// codeToCIDCMapData creates the Type 0 Encoding CMap used by a CID-keyed CFF
// font. ToUnicode answers "what authored text did this code represent?"; this
// CMap instead answers "which descendant-font CID should render this code?".
// Entries are sorted for deterministic PDFs and split into blocks because PDF
// consumers conventionally limit each begincidchar section to 100 entries.
func codeToCIDCMapData(codeToCID map[uint16]uint16, systemInfo cidSystemInfo) []byte {
	codes := make([]int, 0, len(codeToCID))
	for code := range codeToCID {
		codes = append(codes, int(code))
	}
	sort.Ints(codes)
	var sb strings.Builder
	sb.WriteString("/CIDInit /ProcSet findresource begin\n")
	sb.WriteString("12 dict begin\n")
	sb.WriteString("begincmap\n")
	fmt.Fprintf(&sb, "/CIDSystemInfo << /Registry (%s) /Ordering (%s) /Supplement %d >> def\n", systemInfo.registry, systemInfo.ordering, systemInfo.supplement)
	sb.WriteString("/CMapName /LeadType-CID-Encoding def\n")
	sb.WriteString("/CMapType 1 def\n")
	sb.WriteString("1 begincodespacerange\n<0000> <FFFF>\nendcodespacerange\n")
	const blockSize = 100
	for i := 0; i < len(codes); i += blockSize {
		end := i + blockSize
		if end > len(codes) {
			end = len(codes)
		}
		fmt.Fprintf(&sb, "%d begincidchar\n", end-i)
		for _, rawCode := range codes[i:end] {
			code := uint16(rawCode)
			fmt.Fprintf(&sb, "<%04X> %d\n", code, codeToCID[code])
		}
		sb.WriteString("endcidchar\n")
	}
	sb.WriteString("endcmap\nCMapName currentdict /CMap defineresource pop\nend\nend\n")
	return []byte(sb.String())
}

// cidSetData builds the FontDescriptor CIDSet bit field. Bit n says that CID n
// is present in the embedded subset; bit order within each byte is high first,
// as required by PDF.
func cidSetData(cids []uint16) []byte {
	maxCID := uint16(0)
	for _, cid := range cids {
		if cid > maxCID {
			maxCID = cid
		}
	}
	data := make([]byte, int(maxCID)/8+1)
	for _, cid := range cids {
		data[int(cid)/8] |= byte(0x80 >> (cid % 8))
	}
	return data
}
