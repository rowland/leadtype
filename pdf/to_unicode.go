// Copyright 2024 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import (
	"fmt"
	"strings"
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
