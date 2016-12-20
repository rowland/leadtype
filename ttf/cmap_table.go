// Copyright 2011-2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ttf

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
)

type cmapTable struct {
	version          uint16
	numberSubtables  uint16
	encodingRecords  []cmapEncodingRecord
	format0Indexer   glyphIndexer
	format2Indexer   glyphIndexer
	format4Indexer   glyphIndexer
	format6Indexer   glyphIndexer
	format12Indexer  glyphIndexer
	preferredIndexer glyphIndexer
}

func (table *cmapTable) init(rs io.ReadSeeker, entry *tableDirEntry) (err error) {
	if _, err = rs.Seek(int64(entry.offset), os.SEEK_SET); err != nil {
		return
	}
	file := bufio.NewReaderSize(rs, int(entry.length))
	if err = readValues(file, &table.version, &table.numberSubtables); err != nil {
		return
	}
	table.encodingRecords = make([]cmapEncodingRecord, table.numberSubtables)
	for i := uint16(0); i < table.numberSubtables; i++ {
		if err = table.encodingRecords[i].read(file); err != nil {
			return
		}
	}
	for i := uint16(0); i < table.numberSubtables; i++ {
		if err = table.encodingRecords[i].readMapping(file); err != nil {
			return
		}
	}
	preferredEncoding := -1
	for i := 0; uint16(i) < table.numberSubtables; i++ {
		enc := &table.encodingRecords[i]
		switch enc.format {
		case 0:
			table.format0Indexer = enc.glyphIndexer
		case 2:
			table.format2Indexer = enc.glyphIndexer
		case 4:
			table.format4Indexer = enc.glyphIndexer
		case 6:
			table.format6Indexer = enc.glyphIndexer
		case 12:
			table.format12Indexer = enc.glyphIndexer
		}
		if enc.platformID == UnicodePlatformID &&
			(enc.platformSpecificID == Unicode2PlatformSpecificID || enc.platformSpecificID == Unicode2FullPlatformSpecificID) {
			preferredEncoding = i
			break
		}
		if enc.platformID == MicrosoftPlatformID && enc.platformSpecificID == UCS2PlatformSpecificID {
			preferredEncoding = i
		}
		if enc.platformID == MacintoshPlatformID && preferredEncoding < 0 {
			preferredEncoding = i
		}
	}
	if preferredEncoding >= 0 {
		table.preferredIndexer = table.encodingRecords[preferredEncoding].glyphIndexer
	} else {
		err = errors.New("Unable to select preferred cmap glyph indexer.")
	}
	return
}

// 41.9 ns
func (table *cmapTable) glyphIndex(codepoint int) int {
	return table.preferredIndexer.glyphIndex(codepoint)
}

func (table *cmapTable) write(wr io.Writer) {
	fmt.Fprintln(wr, "----------")
	fmt.Fprintln(wr, "cmap Table")
	fmt.Fprintf(wr, "version = %d\n", table.version)
	fmt.Fprintf(wr, "numberSubtables = %d\n", table.numberSubtables)
	for _, rec := range table.encodingRecords {
		rec.write(wr)
	}
}

type cmapEncodingRecord struct {
	platformID         uint16
	platformSpecificID uint16
	offset             uint32
	format             uint16
	glyphIndexer       glyphIndexer
}

func (rec *cmapEncodingRecord) read(file io.Reader) (err error) {
	return readValues(file, &rec.platformID, &rec.platformSpecificID, &rec.offset)
}

func (rec *cmapEncodingRecord) readMapping(file io.Reader) (err error) {
	if err = binary.Read(file, binary.BigEndian, &rec.format); err != nil {
		return
	}
	switch rec.format {
	case 0:
		rec.glyphIndexer = new(format0EncodingRecord)
		err = rec.glyphIndexer.init(file)
	case 2:
		rec.glyphIndexer = new(format2EncodingRecord)
		err = rec.glyphIndexer.init(file)
	case 4:
		rec.glyphIndexer = new(format4EncodingRecord)
		err = rec.glyphIndexer.init(file)
	case 6:
		rec.glyphIndexer = new(format6EncodingRecord)
		err = rec.glyphIndexer.init(file)
	case 12:
		rec.glyphIndexer = new(format12EncodingRecord)
		err = rec.glyphIndexer.init(file)
	default:
		return fmt.Errorf("Unsupported mapping table format: %d", rec.format)
	}
	return
}

func (rec *cmapEncodingRecord) write(wr io.Writer) {
	fmt.Fprintln(wr, "----------")
	fmt.Fprintln(wr, "cmap encoding")
	fmt.Fprintf(wr, "platformID = %d\n", rec.platformID)
	fmt.Fprintf(wr, "platformSpecificID = %d\n", rec.platformSpecificID)
	fmt.Fprintf(wr, "offset = %d\n", rec.offset)
	fmt.Fprintf(wr, "format = %d\n", rec.format)
	rec.glyphIndexer.write(wr)
}

type glyphIndexer interface {
	init(file io.Reader) (err error)
	glyphIndex(codepoint int) int
	write(wr io.Writer)
}

type format0EncodingRecord struct {
	length          uint16
	language        uint16
	segCountX2      uint16
	searchRange     uint16
	entrySelector   uint16
	rangeShift      uint16
	endCode         []uint16
	reservedPad     uint16
	startCode       []uint16
	idDelta         []uint16
	idRangeOffset   []uint16
	glyphIndexArray []uint16
}

func (rec *format0EncodingRecord) init(file io.Reader) (err error) {
	if err = readValues(file, &rec.length, &rec.language); err != nil {
		return
	}
	glyphIndexArray := make([]uint8, 256)
	if err = binary.Read(file, binary.BigEndian, glyphIndexArray); err != nil {
		return
	}
	rec.glyphIndexArray = make([]uint16, 256)
	for i, v := range glyphIndexArray {
		rec.glyphIndexArray[i] = uint16(v)
	}
	return
}

func (enc *format0EncodingRecord) glyphIndex(codepoint int) int {
	low, high := 0, len(enc.endCode)-1
	for low <= high {
		i := (low + high) / 2
		if uint16(codepoint) > enc.endCode[i] {
			low = i + 1
			continue
		}
		if uint16(codepoint) < enc.startCode[i] {
			high = i - 1
			continue
		}
		if enc.idRangeOffset[i] == 0 {
			return int(enc.idDelta[i]) + codepoint
		}
		gi := enc.idRangeOffset[i]/2 + (uint16(codepoint) - enc.startCode[i]) + uint16(i-len(enc.idRangeOffset))
		return int(enc.glyphIndexArray[gi])
	}
	return -1
}

func (rec *format0EncodingRecord) write(wr io.Writer) {
	fmt.Fprintf(wr, "length = %d\n", rec.length)
	fmt.Fprintf(wr, "language = %d\n", rec.language)
	fmt.Fprintf(wr, "segCountX2 = %d\n", rec.segCountX2)
	fmt.Fprintf(wr, "searchRange = %d\n", rec.searchRange)
	fmt.Fprintf(wr, "entrySelector = %d\n", rec.entrySelector)
	fmt.Fprintf(wr, "rangeShift = %d\n", rec.rangeShift)
	fmt.Fprint(wr, "endCode = ", rec.endCode, "\n")
	fmt.Fprintf(wr, "reservedPad = %d\n", rec.reservedPad)
	fmt.Fprint(wr, "startCode = ", rec.startCode, "\n")
	fmt.Fprint(wr, "idDelta = ", rec.idDelta, "\n")
	fmt.Fprint(wr, "idRangeOffset = ", rec.idRangeOffset, "\n")
	fmt.Fprint(wr, "glyphIndexArray = ", rec.glyphIndexArray, "\n")
}

type format2EncodingRecord struct {
	length          uint16
	language        uint16
	subHeaderKeys   [256]uint16
	subHeaders      []subHeader
	glyphIndexArray []uint16
}

func (rec *format2EncodingRecord) init(file io.Reader) (err error) {
	var subHeaderKeys [256]uint16
	if err = readValues(file, &rec.length, &rec.language, &subHeaderKeys); err != nil {
		return
	}
	for i, v := range subHeaderKeys {
		rec.subHeaderKeys[i] = v << 3 // divide by size of subHeader in bytes
	}
	rec.subHeaders = make([]subHeader, countUniqueUint16Values(rec.subHeaderKeys[:]))
	for i, _ := range rec.subHeaders {
		if err = rec.subHeaders[i].read(file); err != nil {
			return
		}
	}
	l := int(rec.length) - (2 + 2 + 2 + (256 * 2)) - len(rec.subHeaders)*(2+2+2+2)
	if l <= 0 {
		return fmt.Errorf("Invalid glyph index array size: %d", l)
	}
	rec.glyphIndexArray = make([]uint16, l)
	if err = readValues(file, &rec.glyphIndexArray); err != nil {
		return
	}
	return
}

func (rec *format2EncodingRecord) glyphIndex(codepoint int) int {
	return -1 // FIXME: Find sample font, write test and implement.
}

func (rec *format2EncodingRecord) write(wr io.Writer) {
	fmt.Fprintf(wr, "length = %d\n", rec.length)
	fmt.Fprintf(wr, "language = %d\n", rec.language)
	fmt.Fprintf(wr, "subHeaderKeys = %v\n", rec.subHeaderKeys)
	fmt.Fprintf(wr, "subHeaders\n")
	fmt.Fprintf(wr, "index\tfirstCode\tentryCount\tidDelta\tidRangeOffset\n")
	for i, _ := range rec.subHeaders {
		fmt.Fprintf(wr, "[%d]\t%d\t%d\t%d\t%d\n", i,
			rec.subHeaders[i].firstCode,
			rec.subHeaders[i].entryCount,
			rec.subHeaders[i].idDelta,
			rec.subHeaders[i].idRangeOffset)
	}
	fmt.Fprintf(wr, "glyphIndexArray = %v\n", rec.glyphIndexArray)
}

type subHeader struct {
	firstCode     uint16
	entryCount    uint16
	idDelta       int16
	idRangeOffset uint16
}

func (sh *subHeader) read(file io.Reader) (err error) {
	return readValues(file, &sh.firstCode, &sh.entryCount, &sh.idDelta, &sh.idRangeOffset)
}

type format4EncodingRecord struct {
	length          uint16
	language        uint16
	segCountX2      uint16
	searchRange     uint16
	entrySelector   uint16
	rangeShift      uint16
	endCode         []uint16
	reservedPad     uint16
	startCode       []uint16
	idDelta         []uint16
	idRangeOffset   []uint16
	glyphIndexArray []uint16
}

func (rec *format4EncodingRecord) init(file io.Reader) (err error) {
	if err = readValues(file,
		&rec.length,
		&rec.language,
		&rec.segCountX2,
		&rec.searchRange,
		&rec.entrySelector,
		&rec.rangeShift); err != nil {
		return
	}
	rec.endCode = make([]uint16, rec.segCountX2/2)
	if err = binary.Read(file, binary.BigEndian, rec.endCode); err != nil {
		return
	}
	if err = binary.Read(file, binary.BigEndian, &rec.reservedPad); err != nil {
		return
	}
	rec.startCode = make([]uint16, rec.segCountX2/2)
	if err = binary.Read(file, binary.BigEndian, rec.startCode); err != nil {
		return
	}
	rec.idDelta = make([]uint16, rec.segCountX2/2)
	if err = binary.Read(file, binary.BigEndian, rec.idDelta); err != nil {
		return
	}
	rec.idRangeOffset = make([]uint16, rec.segCountX2/2)
	if err = binary.Read(file, binary.BigEndian, rec.idRangeOffset); err != nil {
		return
	}
	v := rec.length / 2           // bytes to words
	v -= 8                        // 1 word each for format, length, language, segCountX2, searchRange, entrySelector, rangeShift, reservedPad
	v -= 4 * (rec.segCountX2 / 2) // segCount * 1 word each for endCode, startCode, idDelta, idRangeOffset
	rec.glyphIndexArray = make([]uint16, v)
	if err = binary.Read(file, binary.BigEndian, rec.glyphIndexArray); err != nil {
		return
	}
	return
}

// 35.9 ns
func (enc *format4EncodingRecord) glyphIndex(codepoint int) int {
	low, high := 0, len(enc.endCode)-1
	for low <= high {
		i := (low + high) / 2
		if uint16(codepoint) > enc.endCode[i] {
			low = i + 1
			continue
		}
		if uint16(codepoint) < enc.startCode[i] {
			high = i - 1
			continue
		}
		if enc.idRangeOffset[i] == 0 {
			return int(enc.idDelta[i] + uint16(codepoint))
		}
		gi := enc.idRangeOffset[i]/2 + (uint16(codepoint) - enc.startCode[i]) + uint16(i-len(enc.idRangeOffset))
		return int(enc.glyphIndexArray[gi])
	}
	return -1
}

func (rec *format4EncodingRecord) write(wr io.Writer) {
	fmt.Fprintf(wr, "length = %d\n", rec.length)
	fmt.Fprintf(wr, "language = %d\n", rec.language)
	fmt.Fprintf(wr, "segCountX2 = %d\n", rec.segCountX2)
	fmt.Fprintf(wr, "searchRange = %d\n", rec.searchRange)
	fmt.Fprintf(wr, "entrySelector = %d\n", rec.entrySelector)
	fmt.Fprintf(wr, "rangeShift = %d\n", rec.rangeShift)
	fmt.Fprint(wr, "endCode = ", rec.endCode, "\n")
	fmt.Fprintf(wr, "reservedPad = %d\n", rec.reservedPad)
	fmt.Fprint(wr, "startCode = ", rec.startCode, "\n")
	fmt.Fprint(wr, "idDelta = ", rec.idDelta, "\n")
	fmt.Fprint(wr, "idRangeOffset = ", rec.idRangeOffset, "\n")
	fmt.Fprint(wr, "glyphIndexArray = ", rec.glyphIndexArray, "\n")
}

type format6EncodingRecord struct {
	length          uint16
	language        uint16
	firstCode       uint16
	entryCount      uint16
	glyphIndexArray []uint16
}

func (enc *format6EncodingRecord) init(file io.Reader) (err error) {
	if err = readValues(file,
		&enc.length,
		&enc.language,
		&enc.firstCode,
		&enc.entryCount); err != nil {
		return
	}
	enc.glyphIndexArray = make([]uint16, enc.entryCount)
	return readValues(file, &enc.glyphIndexArray)
}

func (enc *format6EncodingRecord) glyphIndex(codepoint int) int {
	if codepoint < int(enc.firstCode) || codepoint >= int(enc.firstCode+enc.entryCount) {
		return -1
	}
	return int(enc.glyphIndexArray[codepoint-int(enc.firstCode)])
}

func (enc *format6EncodingRecord) write(wr io.Writer) {
	fmt.Fprintf(wr, "length = %d\n", enc.length)
	fmt.Fprintf(wr, "language = %d\n", enc.language)
	fmt.Fprintf(wr, "firstCode = %d\n", enc.firstCode)
	fmt.Fprintf(wr, "firstCode = %d\n", enc.entryCount)
	fmt.Fprintf(wr, "glyphIndexArray = %v\n", enc.glyphIndexArray)
}

type format12EncodingRecord struct {
	frac     uint16
	length   uint32
	language uint32
	nGroups  uint32
	groups   []format12Group
}

func (enc *format12EncodingRecord) init(file io.Reader) (err error) {
	if err = readValues(file,
		&enc.frac,
		&enc.length,
		&enc.language,
		&enc.nGroups); err != nil {
		return
	}
	enc.groups = make([]format12Group, enc.nGroups)
	for i := uint32(0); i < enc.nGroups; i++ {
		if err = enc.groups[i].read(file); err != nil {
			return
		}
	}
	return
}

// 51.7 ns
func (enc *format12EncodingRecord) glyphIndex(codepoint int) int {
	low, high := uint32(0), enc.nGroups-1
	for low <= high {
		i := (low + high) / 2
		group := &enc.groups[i]
		if uint32(codepoint) > group.endCharCode {
			low = i + 1
			continue
		}
		if uint32(codepoint) < group.startCharCode {
			high = i - 1
			continue
		}
		return int(group.startGlyphCode + (uint32(codepoint) - group.startCharCode))
	}
	return -1
}

func (enc *format12EncodingRecord) write(wr io.Writer) {
	fmt.Fprintf(wr, "length = %d\n", enc.length)
	fmt.Fprintf(wr, "language = %d\n", enc.language)
	fmt.Fprintf(wr, "nGroups = %d\n", enc.nGroups)
	for i, group := range enc.groups {
		fmt.Fprintf(wr, "[%d] starCharCode = %d, endCharCode = %d, startGlyphCode = %d\n",
			i, group.startCharCode, group.endCharCode, group.startGlyphCode)
	}
}

type format12Group struct {
	startCharCode  uint32
	endCharCode    uint32
	startGlyphCode uint32
}

func (group *format12Group) read(file io.Reader) (err error) {
	return readValues(file, &group.startCharCode, &group.endCharCode, &group.startGlyphCode)
}
