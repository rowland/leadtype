package ttf

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

type cmapTable struct {
	version          uint16
	numberSubtables  uint16
	encodingRecords  []cmapEncodingRecord
	format0Indexer   glyphIndexer
	format4Indexer   glyphIndexer
	preferredIndexer glyphIndexer
}

func (table *cmapTable) init(rs io.ReadSeeker, entry *tableDirEntry) (err os.Error) {
	if _, err = rs.Seek(int64(entry.offset), os.SEEK_SET); err != nil {
		return
	}
	file, _ := bufio.NewReaderSize(rs, int(entry.length))
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
		case 4:
			table.format4Indexer = enc.glyphIndexer
		}
		if enc.platformID == UnicodePlatformID && enc.platformSpecificID == Unicode2PlatformSpecificID {
			preferredEncoding = i
			break
		}
		if enc.platformID == MicrosoftPlatformID && enc.platformSpecificID == UCS2PlatformSpecificID {
			preferredEncoding = i
		}
	}
	if preferredEncoding >= 0 {
		table.preferredIndexer = table.encodingRecords[preferredEncoding].glyphIndexer
	} else {
		err = os.NewError("Unable to select preferred cmap glyph indexer.")
	}
	return
}

// 76.7 ns
func (table *cmapTable) glyphIndex(codepoint int) uint16 {
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

func (rec *cmapEncodingRecord) read(file io.Reader) (err os.Error) {
	return readValues(file, &rec.platformID, &rec.platformSpecificID, &rec.offset)
}

func (rec *cmapEncodingRecord) readMapping(file io.Reader) (err os.Error) {
	if err = binary.Read(file, binary.BigEndian, &rec.format); err != nil {
		return
	}
	switch rec.format {
	case 0:
		rec.glyphIndexer = new(format0EncodingRecord)
		err = rec.glyphIndexer.init(file)
	case 4:
		rec.glyphIndexer = new(format4EncodingRecord)
		err = rec.glyphIndexer.init(file)
	default:
		return os.NewError(fmt.Sprintf("Unsupported mapping table format: %d", rec.format))
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
	init(file io.Reader) (err os.Error)
	glyphIndex(codepoint int) uint16
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

func (rec *format0EncodingRecord) init(file io.Reader) (err os.Error) {
	if readValues(file, &rec.length, &rec.language); err != nil {
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

func (enc *format0EncodingRecord) glyphIndex(codepoint int) uint16 {
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
			return enc.idDelta[i] + uint16(codepoint)
		}
		// fmt.Printf("codepoint: %d, i: %d, idRangeOffset: %d, startCode: %d, diff: %d, offset: %d\n", codepoint, i, enc.idRangeOffset[i], enc.startCode[i], codepoint - enc.startCode[i], i - len(enc.idRangeOffset))
		gi := enc.idRangeOffset[i]/2 + (uint16(codepoint) - enc.startCode[i]) + uint16(i-len(enc.idRangeOffset))
		return enc.glyphIndexArray[gi]
	}
	return 0
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

func (rec *format4EncodingRecord) init(file io.Reader) (err os.Error) {
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

// 72.0 ns
func (enc *format4EncodingRecord) glyphIndex(codepoint int) uint16 {
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
			return enc.idDelta[i] + uint16(codepoint)
		}
		// fmt.Printf("codepoint: %d, i: %d, idRangeOffset: %d, startCode: %d, diff: %d, offset: %d\n", codepoint, i, enc.idRangeOffset[i], enc.startCode[i], codepoint - enc.startCode[i], i - len(enc.idRangeOffset))
		gi := enc.idRangeOffset[i]/2 + (uint16(codepoint) - enc.startCode[i]) + uint16(i-len(enc.idRangeOffset))
		return enc.glyphIndexArray[gi]
	}
	return 0
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
