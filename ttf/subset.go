// Copyright 2024 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ttf

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"sort"
)

// Composite glyph component flag bits (OpenType spec §2.5.2 / Table 31).
const (
	compositeArgOneAndTwoAreWords uint16 = 1 << 0
	compositeWeHaveAScale         uint16 = 1 << 3
	compositeMoreComponents       uint16 = 1 << 5
	compositeWeHaveAnXAndYScale   uint16 = 1 << 6
	compositeWeHaveATwoByTwo      uint16 = 1 << 7
)

// Subset returns a valid TTF binary containing only the glyphs in glyphIDs
// plus any component glyphs they reference transitively through composite
// glyph records. Glyph 0 (.notdef) is always included. All non-glyph tables
// (head, hhea, hmtx, name, OS/2, post, etc.) are preserved verbatim.
// The output uses long (format 1) loca offsets and is suitable for
// embedding in a PDF as a FontFile2 stream.
func (font *Font) Subset(glyphIDs []uint16) ([]byte, error) {
	raw, err := os.ReadFile(font.filename)
	if err != nil {
		return nil, fmt.Errorf("subset: reading %s: %w", font.filename, err)
	}

	glyfEntry := font.tableDir.table("glyf")
	locaEntry := font.tableDir.table("loca")
	if glyfEntry == nil {
		return nil, fmt.Errorf("subset: font has no glyf table")
	}
	if locaEntry == nil {
		return nil, fmt.Errorf("subset: font has no loca table")
	}

	numGlyphs := int(font.maxpTable.numGlyphs)
	isLong := font.headTable.indexToLocFormat == 1
	glyfOff, glyfLen := parseLocaOffsets(raw, locaEntry, numGlyphs, isLong)

	// Build glyph closure: always include .notdef (glyph 0), then
	// recursively add composite component IDs.
	closure := make(map[uint16]bool, len(glyphIDs)+1)
	closure[0] = true
	var expand func(id uint16)
	expand = func(id uint16) {
		if closure[id] || int(id) >= numGlyphs {
			return
		}
		closure[id] = true
		if glyfLen[id] < 10 {
			return // empty or malformed glyph, no components
		}
		absOff := int(glyfEntry.offset) + int(glyfOff[id])
		numContours := int16(binary.BigEndian.Uint16(raw[absOff:]))
		if numContours >= 0 {
			return // simple glyph, no components
		}
		// Composite glyph: iterate component records.
		// Glyph header: numberOfContours(2) + xMin(2)+yMin(2)+xMax(2)+yMax(2) = 10 bytes.
		pos := absOff + 10
		end := absOff + int(glyfLen[id])
		for pos+4 <= end {
			flags := binary.BigEndian.Uint16(raw[pos:])
			compID := binary.BigEndian.Uint16(raw[pos+2:])
			pos += 4
			expand(compID)
			if flags&compositeArgOneAndTwoAreWords != 0 {
				pos += 4 // two int16 arguments
			} else {
				pos += 2 // two int8 arguments
			}
			if flags&compositeWeHaveATwoByTwo != 0 {
				pos += 8 // 2×2 F2Dot14 transformation matrix
			} else if flags&compositeWeHaveAnXAndYScale != 0 {
				pos += 4 // x-scale + y-scale
			} else if flags&compositeWeHaveAScale != 0 {
				pos += 2 // uniform scale
			}
			if flags&compositeMoreComponents == 0 {
				break
			}
		}
	}
	for _, id := range glyphIDs {
		expand(id)
	}

	// Build new glyf table and corresponding long-format loca.
	var glyfBuf bytes.Buffer
	newLoca := make([]uint32, numGlyphs+1)
	for i := 0; i < numGlyphs; i++ {
		newLoca[i] = uint32(glyfBuf.Len())
		if closure[uint16(i)] && glyfLen[i] > 0 {
			absOff := int(glyfEntry.offset) + int(glyfOff[i])
			glyfBuf.Write(raw[absOff : absOff+int(glyfLen[i])])
			// Pad each glyph slot to a 4-byte boundary.
			if pad := glyfBuf.Len() % 4; pad != 0 {
				glyfBuf.Write(make([]byte, 4-pad))
			}
		}
	}
	newLoca[numGlyphs] = uint32(glyfBuf.Len())

	// Encode loca as big-endian uint32 values (long / format 1).
	locaBuf := make([]byte, (numGlyphs+1)*4)
	for i, off := range newLoca {
		binary.BigEndian.PutUint32(locaBuf[i*4:], off)
	}

	// Collect all tables from the original font, replacing glyf and loca.
	type namedTable struct {
		tag  string
		data []byte
	}
	tables := make([]namedTable, 0, len(font.tableDir.entries))
	for _, entry := range font.tableDir.entries {
		var data []byte
		switch entry.tag {
		case "glyf":
			data = glyfBuf.Bytes()
		case "loca":
			data = locaBuf
		default:
			data = make([]byte, entry.length)
			copy(data, raw[entry.offset:uint32(entry.offset)+entry.length])
		}
		tables = append(tables, namedTable{entry.tag, data})
	}
	// Table directory entries must be in ascending tag order.
	sort.Slice(tables, func(i, j int) bool { return tables[i].tag < tables[j].tag })

	// Compute table offsets. The font starts with:
	//   12-byte offset table + 16*nTables-byte directory
	// followed by table data aligned to a 4-byte boundary.
	nTables := uint16(len(tables))
	dirEnd := 12 + 16*int(nTables)
	dataStart := dirEnd
	if rem := dataStart % 4; rem != 0 {
		dataStart += 4 - rem
	}

	type record struct {
		tag      [4]byte
		checkSum uint32
		offset   uint32
		length   uint32
		data     []byte
	}
	records := make([]record, len(tables))
	off := uint32(dataStart)
	for i, t := range tables {
		copy(records[i].tag[:], t.tag)
		records[i].length = uint32(len(t.data))
		records[i].offset = off
		records[i].checkSum = ttfTableChecksum(t.data)
		records[i].data = t.data
		padded := uint32(len(t.data))
		if rem := padded % 4; rem != 0 {
			padded += 4 - rem
		}
		off += padded
	}

	// Patch the head table:
	//   • zero checkSumAdjustment (byte offset 8, 4 bytes)
	//   • set indexToLocFormat = 1 (byte offset 50, 2 bytes)
	for i := range records {
		if string(records[i].tag[:]) == "head" {
			patched := make([]byte, len(records[i].data))
			copy(patched, records[i].data)
			binary.BigEndian.PutUint32(patched[8:], 0)  // checkSumAdjustment = 0
			binary.BigEndian.PutUint16(patched[50:], 1) // indexToLocFormat = 1 (long)
			records[i].data = patched
			records[i].checkSum = ttfTableChecksum(patched)
		}
	}

	// Assemble the output font binary.
	var out bytes.Buffer
	sr, es, rs := ttfSearchRange(nTables)
	writeUint32(&out, font.scalar)
	writeUint16(&out, nTables)
	writeUint16(&out, sr)
	writeUint16(&out, es)
	writeUint16(&out, rs)

	for _, r := range records {
		out.Write(r.tag[:])
		writeUint32(&out, r.checkSum)
		writeUint32(&out, r.offset)
		writeUint32(&out, r.length)
	}
	for out.Len() < dataStart {
		out.WriteByte(0)
	}
	for _, r := range records {
		out.Write(r.data)
		if pad := len(r.data) % 4; pad != 0 {
			out.Write(make([]byte, 4-pad))
		}
	}

	// Compute and store the head.checkSumAdjustment.
	// Spec: sum of all uint32 words in the file must equal 0xB1B0AFBA.
	result := out.Bytes()
	fileSum := ttfTableChecksum(result)
	adj := uint32(0xB1B0AFBA) - fileSum
	for _, r := range records {
		if string(r.tag[:]) == "head" {
			binary.BigEndian.PutUint32(result[r.offset+8:], adj)
			break
		}
	}

	return result, nil
}

// parseLocaOffsets returns per-glyph (offset, length) pairs within the glyf
// table. Offsets are relative to the start of the glyf table data.
// There are numGlyphs+1 entries in loca; the last gives the end of the final glyph.
func parseLocaOffsets(raw []byte, entry *tableDirEntry, numGlyphs int, isLong bool) (offsets, lengths []uint32) {
	offsets = make([]uint32, numGlyphs)
	lengths = make([]uint32, numGlyphs)
	all := make([]uint32, numGlyphs+1)
	if isLong {
		for i := range all {
			pos := int(entry.offset) + i*4
			all[i] = binary.BigEndian.Uint32(raw[pos:])
		}
	} else {
		// Short loca: stored value × 2 = actual byte offset.
		for i := range all {
			pos := int(entry.offset) + i*2
			all[i] = uint32(binary.BigEndian.Uint16(raw[pos:])) * 2
		}
	}
	for i := 0; i < numGlyphs; i++ {
		offsets[i] = all[i]
		if all[i+1] >= all[i] {
			lengths[i] = all[i+1] - all[i]
		}
	}
	return
}

// ttfTableChecksum computes the TTF/OpenType table checksum: the arithmetic
// sum of all big-endian uint32 words, with the last incomplete word padded
// with zero bytes.
func ttfTableChecksum(data []byte) uint32 {
	var sum uint32
	for len(data) >= 4 {
		sum += binary.BigEndian.Uint32(data)
		data = data[4:]
	}
	if len(data) > 0 {
		var last [4]byte
		copy(last[:], data)
		sum += binary.BigEndian.Uint32(last[:])
	}
	return sum
}

// ttfSearchRange computes the offset table searchRange, entrySelector, and
// rangeShift fields from numTables per the OpenType spec.
func ttfSearchRange(numTables uint16) (searchRange, entrySelector, rangeShift uint16) {
	e := uint16(0)
	for (1 << (e + 1)) <= numTables {
		e++
	}
	searchRange = (1 << e) * 16
	entrySelector = e
	rangeShift = numTables*16 - searchRange
	return
}

func writeUint16(w *bytes.Buffer, v uint16) {
	var b [2]byte
	binary.BigEndian.PutUint16(b[:], v)
	w.Write(b[:])
}

func writeUint32(w *bytes.Buffer, v uint32) {
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], v)
	w.Write(b[:])
}
