//go:build ignore

// Command generate produces the minimal TTF and TTC test fixtures used by the
// ttf package tests. Run with:
//
//	go run ttf/testdata/generate/main.go
//
// Output files are written to ttf/testdata/ relative to the module root.
package main

import (
	"encoding/binary"
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"sort"
)

func main() {
	_, file, _, _ := runtime.Caller(0)
	outDir := filepath.Join(filepath.Dir(file), "..")

	ttfData := buildFont("Minimal", "Regular", 400)
	ttfBold := buildFont("Minimal", "Bold", 700)

	if err := os.WriteFile(filepath.Join(outDir, "minimal.ttf"), ttfData, 0644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("wrote minimal.ttf")

	ttcData := buildTTC(ttfData, ttfBold)
	if err := os.WriteFile(filepath.Join(outDir, "minimal.ttc"), ttcData, 0644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("wrote minimal.ttc")
}

// ── codepoint set ────────────────────────────────────────────────────────────

// codepoints returns the Unicode codepoints the fixture font maps, in order.
func codepoints() []rune {
	var cp []rune
	// .notdef placeholder (glyph 0 in TTF is always .notdef)
	// ASCII printable
	for r := rune(0x0020); r <= 0x007E; r++ {
		cp = append(cp, r)
	}
	// Latin supplement
	for r := rune(0x00C0); r <= 0x00FF; r++ {
		cp = append(cp, r)
	}
	// Greek capitals
	for r := rune(0x0391); r <= 0x03A9; r++ {
		cp = append(cp, r)
	}
	// Two CJK characters
	cp = append(cp, 0x4E2D, 0x6587)
	return cp
}

// ── font builder ─────────────────────────────────────────────────────────────

const (
	unitsPerEm    = 1000
	advanceWidth  = 512
	ascent        = 800
	descent       = -200
	lineGap       = 0
	glyphWidth    = 400
	glyphHeight   = 600
)

func buildFont(family, subfamily string, weight uint16) []byte {
	cp := codepoints()
	numGlyphs := uint16(len(cp) + 1) // +1 for .notdef (glyph 0)

	tables := map[string][]byte{
		"cmap": buildCmap(cp),
		"glyf": buildGlyf(numGlyphs),
		"head": buildHead(),
		"hhea": buildHhea(numGlyphs),
		"hmtx": buildHmtx(numGlyphs),
		"loca": buildLoca(numGlyphs),
		"maxp": buildMaxp(numGlyphs),
		"name": buildName(family, subfamily),
		"OS/2": buildOS2(weight, cp),
		"post": buildPost(),
	}

	return assembleTTF(tables)
}

// ── table builders ────────────────────────────────────────────────────────────

func buildCmap(cp []rune) []byte {
	// cmap header: version=0, numTables=1
	// One encoding record: platformID=3 (Windows), encodingID=1 (UCS-2), format 4
	subtable := buildCmapFormat4(cp)
	subtableOffset := uint32(4 + 8) // header(4) + one encodingRecord(8)

	var b bytes.Buffer
	putU16(&b, 0)              // version
	putU16(&b, 1)              // numTables
	putU16(&b, 3)              // platformID: Windows
	putU16(&b, 1)              // encodingID: UCS-2
	putU32(&b, subtableOffset) // offset to subtable
	b.Write(subtable)
	return b.Bytes()
}

func buildCmapFormat4(cp []rune) []byte {
	// Build segments: each contiguous run of codepoints is one segment.
	type segment struct{ start, end rune }
	var segs []segment
	sorted := make([]rune, len(cp))
	copy(sorted, cp)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })

	for i := 0; i < len(sorted); {
		j := i
		for j+1 < len(sorted) && sorted[j+1] == sorted[j]+1 {
			j++
		}
		segs = append(segs, segment{sorted[i], sorted[j]})
		i = j + 1
	}
	// Terminating segment required by spec
	segs = append(segs, segment{0xFFFF, 0xFFFF})

	segCount := uint16(len(segs))
	segCount2 := segCount * 2
	// searchRange: largest power of 2 <= segCount, times 2
	sr := uint16(1)
	for sr*2 <= segCount {
		sr *= 2
	}
	searchRange := sr * 2
	entrySelector := uint16(0)
	for tmp := sr; tmp > 1; tmp >>= 1 {
		entrySelector++
	}
	rangeShift := segCount2 - searchRange

	// idDelta: for each segment, delta = firstGlyphID - startCode
	// glyph 0 is .notdef; actual glyphs start at 1 in order of sorted codepoints
	// Build a map from codepoint to glyph ID
	glyphID := map[rune]uint16{}
	for i, r := range sorted {
		glyphID[r] = uint16(i + 1) // glyph 0 = .notdef
	}

	// For segments where all deltas are the same we can use idDelta.
	// For our fixture all segments have a uniform delta (contiguous cp → contiguous glyph IDs).
	endCodes := make([]uint16, segCount)
	startCodes := make([]uint16, segCount)
	idDeltas := make([]uint16, segCount)
	idRangeOffsets := make([]uint16, segCount)

	glyphPos := uint16(1)
	for i, seg := range segs {
		if seg.start == 0xFFFF {
			endCodes[i] = 0xFFFF
			startCodes[i] = 0xFFFF
			idDeltas[i] = 1
			idRangeOffsets[i] = 0
			continue
		}
		endCodes[i] = uint16(seg.end)
		startCodes[i] = uint16(seg.start)
		// delta = glyphID[seg.start] - seg.start
		delta := int32(glyphPos) - int32(seg.start)
		idDeltas[i] = uint16(delta)
		idRangeOffsets[i] = 0
		glyphPos += uint16(seg.end-seg.start) + 1
	}

	headerSize := uint16(14) // format(2)+length(2)+language(2)+segCountX2(2)+searchRange(2)+entrySelector(2)+rangeShift(2)
	arraysSize := segCount2*4 + 2 // endCode(segCount*2) + reservedPad(2) + startCode + idDelta + idRangeOffset
	length := headerSize + arraysSize

	var b bytes.Buffer
	putU16(&b, 4)            // format
	putU16(&b, length)       // length
	putU16(&b, 0)            // language
	putU16(&b, segCount2)    // segCountX2
	putU16(&b, searchRange)
	putU16(&b, entrySelector)
	putU16(&b, rangeShift)
	for _, v := range endCodes {
		putU16(&b, v)
	}
	putU16(&b, 0) // reservedPad
	for _, v := range startCodes {
		putU16(&b, v)
	}
	for _, v := range idDeltas {
		putU16(&b, v)
	}
	for _, v := range idRangeOffsets {
		putU16(&b, v)
	}
	return b.Bytes()
}

func buildGlyf(numGlyphs uint16) []byte {
	// One simple square glyph repeated for every glyph slot.
	// Format: numberOfContours=1, xMin, yMin, xMax, yMax, endPtsOfContours[0]=3,
	// instructionLength=0, flags×4, xCoordinates×4, yCoordinates×4
	glyph := simpleSquareGlyph()
	var b bytes.Buffer
	for i := uint16(0); i < numGlyphs; i++ {
		b.Write(glyph)
		// Pad to 4-byte boundary
		if len(glyph)%4 != 0 {
			pad := 4 - len(glyph)%4
			b.Write(make([]byte, pad))
		}
	}
	return b.Bytes()
}

func simpleSquareGlyph() []byte {
	var b bytes.Buffer
	putI16(&b, 1)          // numberOfContours
	putI16(&b, 0)          // xMin
	putI16(&b, 0)          // yMin
	putI16(&b, glyphWidth) // xMax
	putI16(&b, glyphHeight)// yMax
	putU16(&b, 3)          // endPtsOfContours[0] = 3 (4 points: 0..3)
	putU16(&b, 0)          // instructionLength
	// flags: 4 on-curve points, each flag = 0x01 (on curve)
	b.WriteByte(0x01)
	b.WriteByte(0x01)
	b.WriteByte(0x01)
	b.WriteByte(0x01)
	// xCoordinates (relative): 0, glyphWidth, 0, -glyphWidth
	putI16(&b, 0)
	putI16(&b, glyphWidth)
	putI16(&b, 0)
	putI16(&b, -glyphWidth)
	// yCoordinates (relative): 0, 0, glyphHeight, -glyphHeight
	putI16(&b, 0)
	putI16(&b, 0)
	putI16(&b, glyphHeight)
	putI16(&b, -glyphHeight)
	return b.Bytes()
}

func buildHead() []byte {
	var b bytes.Buffer
	putU32(&b, 0x00010000) // version 1.0
	putU32(&b, 0x00010000) // fontRevision
	putU32(&b, 0)          // checkSumAdjustment (placeholder; fixed up in assembleTTF)
	putU32(&b, 0x5F0F3CF5) // magicNumber
	putU16(&b, 0x000B)     // flags
	putU16(&b, unitsPerEm)
	putU64(&b, 0) // created
	putU64(&b, 0) // modified
	putI16(&b, 0)          // xMin
	putI16(&b, descent)    // yMin
	putI16(&b, glyphWidth) // xMax
	putI16(&b, ascent)     // yMax
	putU16(&b, 0)          // macStyle
	putU16(&b, 8)          // lowestRecPPEM
	putI16(&b, 2)          // fontDirectionHint
	putI16(&b, 0)          // indexToLocFormat: short offsets
	putI16(&b, 0)          // glyphDataFormat
	return b.Bytes()
}

func buildHhea(numGlyphs uint16) []byte {
	var b bytes.Buffer
	putU32(&b, 0x00010000)  // version
	putI16(&b, ascent)
	putI16(&b, descent)
	putI16(&b, lineGap)
	putU16(&b, advanceWidth) // advanceWidthMax
	putI16(&b, 0)            // minLeftSideBearing
	putI16(&b, 0)            // minRightSideBearing
	putI16(&b, glyphWidth)   // xMaxExtent
	putI16(&b, 1)            // caretSlopeRise
	putI16(&b, 0)            // caretSlopeRun
	putI16(&b, 0)            // caretOffset
	putI16(&b, 0)            // reserved ×4
	putI16(&b, 0)
	putI16(&b, 0)
	putI16(&b, 0)
	putI16(&b, 0)            // metricDataFormat
	putU16(&b, numGlyphs)   // numberOfHMetrics
	return b.Bytes()
}

func buildHmtx(numGlyphs uint16) []byte {
	var b bytes.Buffer
	for i := uint16(0); i < numGlyphs; i++ {
		putU16(&b, advanceWidth)
		putI16(&b, 0) // lsb
	}
	return b.Bytes()
}

func buildLoca(numGlyphs uint16) []byte {
	// Short loca: offsets in units of 2 bytes. Each glyph is the same size.
	glyphBytes := uint16(len(simpleSquareGlyph()))
	// Pad to 4 bytes
	if glyphBytes%4 != 0 {
		glyphBytes += uint16(4 - glyphBytes%4)
	}
	var b bytes.Buffer
	for i := uint16(0); i <= numGlyphs; i++ {
		putU16(&b, (glyphBytes/2)*i) // offset / 2 for short loca
	}
	return b.Bytes()
}

func buildMaxp(numGlyphs uint16) []byte {
	var b bytes.Buffer
	putU32(&b, 0x00010000) // version 1.0
	putU16(&b, numGlyphs)
	putU16(&b, 4)  // maxPoints
	putU16(&b, 1)  // maxContours
	putU16(&b, 0)  // maxCompositePoints
	putU16(&b, 0)  // maxCompositeContours
	putU16(&b, 2)  // maxZones
	putU16(&b, 0)  // maxTwilightPoints
	putU16(&b, 0)  // maxStorage
	putU16(&b, 0)  // maxFunctionDefs
	putU16(&b, 0)  // maxInstructionDefs
	putU16(&b, 0)  // maxStackElements
	putU16(&b, 0)  // maxSizeOfInstructions
	putU16(&b, 0)  // maxComponentElements
	putU16(&b, 0)  // maxComponentDepth
	return b.Bytes()
}

func buildName(family, subfamily string) []byte {
	postscript := family + "-" + subfamily
	if subfamily == "Regular" {
		postscript = family
	}
	type nameEntry struct {
		id  uint16
		val string
	}
	entries := []nameEntry{
		{1, family},
		{2, subfamily},
		{4, family + " " + subfamily},
		{6, postscript},
	}
	if subfamily == "Regular" {
		entries[2] = nameEntry{4, family}
	}

	// Platform 3 (Windows), encoding 1 (UCS-2), language 0x0409 (en-US)
	// Strings stored as UTF-16 BE
	var stringData bytes.Buffer
	type record struct {
		platformID uint16
		encodingID uint16
		languageID uint16
		nameID     uint16
		length     uint16
		offset     uint16
	}
	var records []record
	for _, e := range entries {
		utf16 := toUTF16BE(e.val)
		records = append(records, record{3, 1, 0x0409, e.id, uint16(len(utf16)), uint16(stringData.Len())})
		stringData.Write(utf16)
	}

	count := uint16(len(records))
	stringOffset := uint16(6 + count*12)
	var b bytes.Buffer
	putU16(&b, 0)            // format
	putU16(&b, count)
	putU16(&b, stringOffset)
	for _, r := range records {
		putU16(&b, r.platformID)
		putU16(&b, r.encodingID)
		putU16(&b, r.languageID)
		putU16(&b, r.nameID)
		putU16(&b, r.length)
		putU16(&b, r.offset)
	}
	b.Write(stringData.Bytes())
	return b.Bytes()
}

func buildOS2(weight uint16, cp []rune) []byte {
	// Find first/last char index
	minCP, maxCP := rune(0xFFFF), rune(0)
	for _, r := range cp {
		if r < minCP { minCP = r }
		if r > maxCP { maxCP = r }
	}

	var b bytes.Buffer
	putU16(&b, 4)            // version
	putI16(&b, int16(advanceWidth/2)) // xAvgCharWidth
	putU16(&b, weight)       // usWeightClass
	putU16(&b, 5)            // usWidthClass: medium
	putU16(&b, 0)            // fsType: embeddable
	putI16(&b, 100)          // ySubscriptXSize
	putI16(&b, 100)          // ySubscriptYSize
	putI16(&b, 0)            // ySubscriptXOffset
	putI16(&b, -100)         // ySubscriptYOffset
	putI16(&b, 100)          // ySuperscriptXSize
	putI16(&b, 100)          // ySuperscriptYSize
	putI16(&b, 0)            // ySuperscriptXOffset
	putI16(&b, 200)          // ySuperscriptYOffset
	putI16(&b, 50)           // yStrikeoutSize
	putI16(&b, 250)          // yStrikeoutPosition
	putI16(&b, 0)            // sFamilyClass
	b.Write(make([]byte, 10)) // panose
	// ulUnicodeRange: set bits for Latin (0), Latin supplement (1), Greek (7), CJK (59)
	putU32(&b, (1<<0)|(1<<1)|(1<<7))
	putU32(&b, 1<<(59-32))
	putU32(&b, 0)
	putU32(&b, 0)
	b.Write([]byte("MINI"))  // achVendID
	putU16(&b, 0x0040)       // fsSelection: REGULAR
	putU16(&b, uint16(minCP)) // fsFirstCharIndex
	putU16(&b, uint16(maxCP)) // fsLastCharIndex
	putI16(&b, ascent)       // sTypoAscender
	putI16(&b, descent)      // sTypoDescender
	putI16(&b, lineGap)      // sTypoLineGap
	putU16(&b, uint16(ascent))  // usWinAscent
	putU16(&b, uint16(-descent)) // usWinDescent
	// Version 1+ fields
	putU32(&b, 0) // ulCodePageRange1
	putU32(&b, 0) // ulCodePageRange2
	// Version 2+ fields
	putI16(&b, int16(ascent*7/10)) // sxHeight
	putI16(&b, int16(ascent))      // sCapHeight
	putU16(&b, 0)  // usDefaultChar
	putU16(&b, 32) // usBreakChar (space)
	putU16(&b, 1)  // usMaxContext
	return b.Bytes()
}

func buildPost() []byte {
	var b bytes.Buffer
	putU32(&b, 0x00030000) // format 3.0 (no glyph names)
	putU32(&b, 0)          // italicAngle
	putI16(&b, -100)       // underlinePosition
	putI16(&b, 50)         // underlineThickness
	putU32(&b, 0)          // isFixedPitch
	putU32(&b, 0)          // minMemType42
	putU32(&b, 0)          // maxMemType42
	putU32(&b, 0)          // minMemType1
	putU32(&b, 0)          // maxMemType1
	return b.Bytes()
}

// ── TTF assembler ─────────────────────────────────────────────────────────────

// tableOrder is the recommended table order for TTF files.
var tableOrder = []string{"OS/2", "cmap", "glyf", "head", "hhea", "hmtx", "loca", "maxp", "name", "post"}

func assembleTTF(tables map[string][]byte) []byte {
	// Sort tables by tag (required order for some readers; spec recommends specific order).
	var tags []string
	for t := range tables {
		tags = append(tags, t)
	}
	sort.Slice(tags, func(i, j int) bool { return tags[i] < tags[j] })

	nTables := uint16(len(tags))

	// Offset table (sfVersion + nTables + searchRange + entrySelector + rangeShift)
	sr := uint16(1)
	for sr*2 <= nTables {
		sr *= 2
	}
	searchRange := sr * 16
	entrySelector := uint16(0)
	for tmp := sr; tmp > 1; tmp >>= 1 {
		entrySelector++
	}
	rangeShift := nTables*16 - searchRange

	offsetTableSize := 12
	dirSize := int(nTables) * 16
	headerSize := offsetTableSize + dirSize

	// Compute table offsets (must be 4-byte aligned)
	type tableInfo struct {
		tag    string
		data   []byte
		offset uint32
		padded []byte
	}
	infos := make([]tableInfo, len(tags))
	pos := uint32(headerSize)
	for i, tag := range tags {
		data := tables[tag]
		padded := data
		if len(data)%4 != 0 {
			pad := make([]byte, 4-len(data)%4)
			padded = append(data, pad...)
		}
		infos[i] = tableInfo{tag, data, pos, padded}
		pos += uint32(len(padded))
	}

	// Assemble
	var b bytes.Buffer

	// Offset table
	putU32(&b, 0x00010000) // sfVersion: TrueType
	putU16(&b, nTables)
	putU16(&b, searchRange)
	putU16(&b, entrySelector)
	putU16(&b, rangeShift)

	// Table directory
	for _, info := range infos {
		tag := [4]byte{}
		copy(tag[:], info.tag)
		b.Write(tag[:])
		putU32(&b, checksum(info.data))
		putU32(&b, info.offset)
		putU32(&b, uint32(len(info.data)))
	}

	// Table data
	for _, info := range infos {
		b.Write(info.padded)
	}

	// Fix head.checkSumAdjustment
	result := b.Bytes()
	total := checksumFile(result)
	adj := 0xB1B0AFBA - total
	// Find head table offset
	for _, info := range infos {
		if info.tag == "head" {
			// checkSumAdjustment is at byte 8 within the head table
			off := info.offset + 8
			binary.BigEndian.PutUint32(result[off:], adj)
			break
		}
	}
	return result
}

// ── TTC builder ───────────────────────────────────────────────────────────────

// rebaseFont returns a copy of a TTF font byte slice with all table directory
// offsets incremented by base, so that the offsets are absolute within a TTC file.
// It also recomputes each table's checksum field and the head.checkSumAdjustment
// (which is zeroed; the caller should recompute for the whole TTC if needed).
func rebaseFont(font []byte, base uint32) []byte {
	f := make([]byte, len(font))
	copy(f, font)
	nTables := binary.BigEndian.Uint16(f[4:6])
	for i := uint16(0); i < nTables; i++ {
		entryOff := 12 + int(i)*16
		origOffset := binary.BigEndian.Uint32(f[entryOff+8:])
		binary.BigEndian.PutUint32(f[entryOff+8:], origOffset+base)
	}
	return f
}

func buildTTC(fonts ...[]byte) []byte {
	numFonts := uint32(len(fonts))
	// TTC header: tag(4) + version(4) + numFonts(4) + offsets(4*numFonts)
	headerSize := uint32(12 + 4*numFonts)

	offsets := make([]uint32, numFonts)
	pos := headerSize
	for i, f := range fonts {
		offsets[i] = pos
		pos += uint32(len(f))
	}

	var b bytes.Buffer
	b.WriteString("ttcf")
	putU32(&b, 0x00010000) // version 1.0
	putU32(&b, numFonts)
	for _, off := range offsets {
		putU32(&b, off)
	}
	for i, f := range fonts {
		b.Write(rebaseFont(f, offsets[i]))
	}
	return b.Bytes()
}

// ── checksum helpers ──────────────────────────────────────────────────────────

func checksum(data []byte) uint32 {
	var sum uint32
	for i := 0; i+3 < len(data); i += 4 {
		sum += binary.BigEndian.Uint32(data[i:])
	}
	// Handle trailing bytes
	if rem := len(data) % 4; rem != 0 {
		var buf [4]byte
		copy(buf[:], data[len(data)-rem:])
		sum += binary.BigEndian.Uint32(buf[:])
	}
	return sum
}

func checksumFile(data []byte) uint32 {
	var sum uint32
	for i := 0; i+3 < len(data); i += 4 {
		sum += binary.BigEndian.Uint32(data[i:])
	}
	return sum
}

// ── encoding helpers ──────────────────────────────────────────────────────────

func toUTF16BE(s string) []byte {
	var b bytes.Buffer
	for _, r := range s {
		if r < 0x10000 {
			putU16(&b, uint16(r))
		} else {
			r -= 0x10000
			putU16(&b, uint16(0xD800+(r>>10)))
			putU16(&b, uint16(0xDC00+(r&0x3FF)))
		}
	}
	return b.Bytes()
}

func putU16(b *bytes.Buffer, v uint16) {
	b.WriteByte(byte(v >> 8))
	b.WriteByte(byte(v))
}

func putI16(b *bytes.Buffer, v int16) {
	putU16(b, uint16(v))
}

func putU32(b *bytes.Buffer, v uint32) {
	b.WriteByte(byte(v >> 24))
	b.WriteByte(byte(v >> 16))
	b.WriteByte(byte(v >> 8))
	b.WriteByte(byte(v))
}

func putU64(b *bytes.Buffer, v uint64) {
	putU32(b, uint32(v>>32))
	putU32(b, uint32(v))
}

// Silence unused import warning for rand (used indirectly via build tag ignore).
var _ = rand.New
