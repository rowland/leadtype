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
// glyph records. Glyph 0 (.notdef) is always included. The output uses long
// (format 1) loca offsets and trims maxp, hmtx, cmap, and post alongside
// glyf/loca so the embedded subset is materially smaller while preserving
// glyph IDs up to the highest included glyph.
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

	subsetGlyphCount := 1
	for id := range closure {
		if int(id)+1 > subsetGlyphCount {
			subsetGlyphCount = int(id) + 1
		}
	}

	// Build new glyf table and corresponding long-format loca.
	var glyfBuf bytes.Buffer
	newLoca := make([]uint32, subsetGlyphCount+1)
	for i := 0; i < subsetGlyphCount; i++ {
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
	newLoca[subsetGlyphCount] = uint32(glyfBuf.Len())

	// Encode loca as big-endian uint32 values (long / format 1).
	locaBuf := make([]byte, (subsetGlyphCount+1)*4)
	for i, off := range newLoca {
		binary.BigEndian.PutUint32(locaBuf[i*4:], off)
	}

	maxpBuf, err := subsetMaxpTable(raw, font.tableDir.table("maxp"), uint16(subsetGlyphCount))
	if err != nil {
		return nil, err
	}
	hheaBuf, err := subsetHheaTable(raw, font.tableDir.table("hhea"), uint16(subsetGlyphCount))
	if err != nil {
		return nil, err
	}
	hmtxBuf, err := font.subsetHmtxTable(uint16(subsetGlyphCount))
	if err != nil {
		return nil, err
	}
	cmapBuf, err := font.subsetCmapTable(closure)
	if err != nil {
		return nil, err
	}
	postBuf, err := font.subsetPostTable(raw, font.tableDir.table("post"), uint16(subsetGlyphCount))
	if err != nil {
		return nil, err
	}

	// Collect all tables from the original font, replacing glyf and loca.
	type namedTable struct {
		tag  string
		data []byte
	}
	tables := make([]namedTable, 0, len(font.tableDir.entries))
	for _, entry := range font.tableDir.entries {
		if !subsetKeepTable(entry.tag) {
			continue
		}
		var data []byte
		switch entry.tag {
		case "glyf":
			data = glyfBuf.Bytes()
		case "loca":
			data = locaBuf
		case "maxp":
			data = maxpBuf
		case "hhea":
			data = hheaBuf
		case "hmtx":
			data = hmtxBuf
		case "cmap":
			data = cmapBuf
		case "post":
			data = postBuf
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

var subsetRequiredTables = map[string]bool{
	"cmap": true,
	"glyf": true,
	"head": true,
	"hhea": true,
	"hmtx": true,
	"loca": true,
	"maxp": true,
	"name": true,
	"OS/2": true,
	"post": true,
}

var subsetConservativeTables = map[string]bool{
	"cvt ": true,
	"fpgm": true,
	"prep": true,
}

func subsetKeepTable(tag string) bool {
	return subsetRequiredTables[tag] || subsetConservativeTables[tag]
}

func subsetMaxpTable(raw []byte, entry *tableDirEntry, numGlyphs uint16) ([]byte, error) {
	if entry == nil {
		return nil, fmt.Errorf("subset: font has no maxp table")
	}
	if entry.length < 6 {
		return nil, fmt.Errorf("subset: malformed maxp table")
	}
	data := make([]byte, entry.length)
	copy(data, raw[entry.offset:entry.offset+entry.length])
	binary.BigEndian.PutUint16(data[4:], numGlyphs)
	return data, nil
}

func subsetHheaTable(raw []byte, entry *tableDirEntry, numGlyphs uint16) ([]byte, error) {
	if entry == nil {
		return nil, fmt.Errorf("subset: font has no hhea table")
	}
	if entry.length < 36 {
		return nil, fmt.Errorf("subset: malformed hhea table")
	}
	data := make([]byte, entry.length)
	copy(data, raw[entry.offset:entry.offset+entry.length])
	binary.BigEndian.PutUint16(data[34:], numGlyphs)
	return data, nil
}

func (font *Font) subsetHmtxTable(numGlyphs uint16) ([]byte, error) {
	var buf bytes.Buffer
	for gid := uint16(0); gid < numGlyphs; gid++ {
		metric := font.hmtxTable.lookup(int(gid))
		writeUint16(&buf, metric.advanceWidth)
		writeInt16(&buf, metric.leftSideBearing)
	}
	return buf.Bytes(), nil
}

func (font *Font) subsetCmapTable(closure map[uint16]bool) ([]byte, error) {
	mappings, err := font.subsetCodepointMappings(closure)
	if err != nil {
		return nil, err
	}
	allBMP := true
	for codepoint := range mappings {
		if codepoint > 0xFFFF {
			allBMP = false
			break
		}
	}
	if allBMP {
		return buildSubsetFormat4Cmap(mappings), nil
	}
	return buildSubsetFormat12Cmap(mappings), nil
}

func (font *Font) subsetCodepointMappings(closure map[uint16]bool) (map[int]uint16, error) {
	rec := font.cmapTable.preferredRecord()
	if rec == nil || rec.glyphIndexer == nil {
		return nil, fmt.Errorf("subset: font has no preferred cmap")
	}

	mappings := make(map[int]uint16)
	switch enc := rec.glyphIndexer.(type) {
	case *format0EncodingRecord:
		for codepoint, glyphID := range enc.glyphIndexArray {
			if glyphID != 0 && closure[glyphID] {
				mappings[codepoint] = glyphID
			}
		}
	case *format4EncodingRecord:
		for _, codepoint := range enc.codepoints() {
			glyphID := uint16(enc.glyphIndex(codepoint))
			if glyphID != 0 && closure[glyphID] {
				mappings[codepoint] = glyphID
			}
		}
	case *format6EncodingRecord:
		for i, glyphID := range enc.glyphIndexArray {
			if glyphID != 0 && closure[glyphID] {
				mappings[int(enc.firstCode)+i] = glyphID
			}
		}
	case *format12EncodingRecord:
		for _, group := range enc.groups {
			glyphID := group.startGlyphCode
			for codepoint := group.startCharCode; codepoint <= group.endCharCode; codepoint++ {
				if glyphID != 0 && closure[uint16(glyphID)] {
					mappings[int(codepoint)] = uint16(glyphID)
				}
				glyphID++
			}
		}
	default:
		return nil, fmt.Errorf("subset: unsupported preferred cmap format %d", rec.format)
	}
	return mappings, nil
}

func buildSubsetFormat4Cmap(mappings map[int]uint16) []byte {
	type segment struct {
		start  uint16
		end    uint16
		glyphs []uint16
	}

	codepoints := sortedSubsetCodepoints(mappings, 0xFFFF)
	segments := make([]segment, 0, len(codepoints)+1)
	for i := 0; i < len(codepoints); {
		start := uint16(codepoints[i])
		end := start
		glyphs := []uint16{mappings[codepoints[i]]}
		i++
		for i < len(codepoints) && codepoints[i] == int(end)+1 {
			end = uint16(codepoints[i])
			glyphs = append(glyphs, mappings[codepoints[i]])
			i++
		}
		segments = append(segments, segment{start: start, end: end, glyphs: glyphs})
	}
	segments = append(segments, segment{start: 0xFFFF, end: 0xFFFF})

	segCount := uint16(len(segments))
	segCountX2 := segCount * 2
	searchRange, entrySelector, rangeShift := format4SearchFields(segCount)
	glyphArrayCount := 0
	for _, seg := range segments[:len(segments)-1] {
		glyphArrayCount += len(seg.glyphs)
	}
	length := uint16(16 + 8*int(segCount) + 2*glyphArrayCount)

	var subtable bytes.Buffer
	writeUint16(&subtable, 4)
	writeUint16(&subtable, length)
	writeUint16(&subtable, 0)
	writeUint16(&subtable, segCountX2)
	writeUint16(&subtable, searchRange)
	writeUint16(&subtable, entrySelector)
	writeUint16(&subtable, rangeShift)

	for _, seg := range segments {
		writeUint16(&subtable, seg.end)
	}
	writeUint16(&subtable, 0)
	for _, seg := range segments {
		writeUint16(&subtable, seg.start)
	}
	for range segments {
		writeUint16(&subtable, 0)
	}

	glyphArrayIndex := 0
	for i, seg := range segments {
		if i == len(segments)-1 {
			writeUint16(&subtable, 0)
			continue
		}
		offsetWords := len(segments) - i + glyphArrayIndex
		writeUint16(&subtable, uint16(offsetWords*2))
		glyphArrayIndex += len(seg.glyphs)
	}
	for _, seg := range segments[:len(segments)-1] {
		for _, glyphID := range seg.glyphs {
			writeUint16(&subtable, glyphID)
		}
	}

	var cmap bytes.Buffer
	writeUint16(&cmap, 0)
	writeUint16(&cmap, 2)
	writeUint16(&cmap, UnicodePlatformID)
	writeUint16(&cmap, Unicode2PlatformSpecificID)
	writeUint32(&cmap, 20)
	writeUint16(&cmap, MicrosoftPlatformID)
	writeUint16(&cmap, UCS2PlatformSpecificID)
	writeUint32(&cmap, 20)
	cmap.Write(subtable.Bytes())
	return cmap.Bytes()
}

func buildSubsetFormat12Cmap(mappings map[int]uint16) []byte {
	type group struct {
		startCodepoint int
		endCodepoint   int
		startGlyphID   uint16
	}

	codepoints := sortedSubsetCodepoints(mappings, int(^uint32(0)>>1))
	groups := make([]group, 0, len(codepoints))
	for i := 0; i < len(codepoints); {
		start := codepoints[i]
		end := start
		startGlyph := mappings[start]
		nextGlyph := startGlyph + 1
		i++
		for i < len(codepoints) && codepoints[i] == end+1 && mappings[codepoints[i]] == nextGlyph {
			end = codepoints[i]
			nextGlyph++
			i++
		}
		groups = append(groups, group{
			startCodepoint: start,
			endCodepoint:   end,
			startGlyphID:   startGlyph,
		})
	}

	length := uint32(16 + len(groups)*12)
	var subtable bytes.Buffer
	writeUint16(&subtable, 12)
	writeUint16(&subtable, 0)
	writeUint32(&subtable, length)
	writeUint32(&subtable, 0)
	writeUint32(&subtable, uint32(len(groups)))
	for _, group := range groups {
		writeUint32(&subtable, uint32(group.startCodepoint))
		writeUint32(&subtable, uint32(group.endCodepoint))
		writeUint32(&subtable, uint32(group.startGlyphID))
	}

	var cmap bytes.Buffer
	writeUint16(&cmap, 0)
	writeUint16(&cmap, 1)
	writeUint16(&cmap, UnicodePlatformID)
	writeUint16(&cmap, Unicode2FullPlatformSpecificID)
	writeUint32(&cmap, 12)
	cmap.Write(subtable.Bytes())
	return cmap.Bytes()
}

func sortedSubsetCodepoints(mappings map[int]uint16, maxCodepoint int) []int {
	codepoints := make([]int, 0, len(mappings))
	for codepoint := range mappings {
		if codepoint >= 0 && codepoint <= maxCodepoint {
			codepoints = append(codepoints, codepoint)
		}
	}
	sort.Ints(codepoints)
	return codepoints
}

func format4SearchFields(segCount uint16) (searchRange, entrySelector, rangeShift uint16) {
	power := uint16(1)
	for power<<1 <= segCount {
		power <<= 1
		entrySelector++
	}
	searchRange = power * 2
	rangeShift = segCount*2 - searchRange
	return
}

func (font *Font) subsetPostTable(raw []byte, entry *tableDirEntry, numGlyphs uint16) ([]byte, error) {
	if entry == nil {
		return nil, fmt.Errorf("subset: font has no post table")
	}
	if font.postTable.format.Tof64() != 2.0 {
		data := make([]byte, entry.length)
		copy(data, raw[entry.offset:entry.offset+entry.length])
		return data, nil
	}
	return font.buildSubsetPostFormat2(numGlyphs)
}

func (font *Font) buildSubsetPostFormat2(numGlyphs uint16) ([]byte, error) {
	standardNames := make(map[string]uint16, len(MacRomanPostNames))
	for i, name := range MacRomanPostNames {
		standardNames[name] = uint16(i)
	}

	var names []string
	if len(font.postTable.names) >= int(numGlyphs) {
		names = font.postTable.names[:numGlyphs]
	} else {
		names = make([]string, numGlyphs)
		copy(names, font.postTable.names)
	}

	var buf bytes.Buffer
	writeFixed(&buf, font.postTable.format)
	writeFixed(&buf, font.postTable.italicAngle)
	writeInt16(&buf, font.postTable.underlinePosition)
	writeInt16(&buf, font.postTable.underlineThickness)
	writeUint32(&buf, font.postTable.isFixedPitch)
	writeUint32(&buf, font.postTable.minMemType42)
	writeUint32(&buf, font.postTable.maxMemType42)
	writeUint32(&buf, font.postTable.minMemType1)
	writeUint32(&buf, font.postTable.maxMemType1)
	writeUint16(&buf, numGlyphs)

	newNames := make([]string, 0)
	newNameIndexes := make(map[string]uint16)
	for _, name := range names {
		if idx, ok := standardNames[name]; ok {
			writeUint16(&buf, idx)
			continue
		}
		idx, ok := newNameIndexes[name]
		if !ok {
			idx = uint16(258 + len(newNames))
			newNameIndexes[name] = idx
			newNames = append(newNames, name)
		}
		writeUint16(&buf, idx)
	}
	for _, name := range newNames {
		if len(name) > 255 {
			return nil, fmt.Errorf("subset: post glyph name too long: %q", name)
		}
		buf.WriteByte(byte(len(name)))
		buf.WriteString(name)
	}
	return buf.Bytes(), nil
}

func (enc *format4EncodingRecord) codepoints() []int {
	var codepoints []int
	for i := range enc.startCode {
		start := enc.startCode[i]
		end := enc.endCode[i]
		if start == 0xFFFF && end == 0xFFFF {
			continue
		}
		for codepoint := start; codepoint <= end; codepoint++ {
			if glyphID := enc.glyphIndex(int(codepoint)); glyphID > 0 {
				codepoints = append(codepoints, int(codepoint))
			}
		}
	}
	return codepoints
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

func writeInt16(w *bytes.Buffer, v int16) {
	var b [2]byte
	binary.BigEndian.PutUint16(b[:], uint16(v))
	w.Write(b[:])
}

func writeFixed(w *bytes.Buffer, v Fixed) {
	writeInt16(w, v.base)
	writeUint16(w, v.frac)
}
