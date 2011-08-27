package ttf

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

type cmapTable struct {
	version         uint16
	numberSubtables uint16
	encodingRecords []cmapEncodingRecord
}

func (table *cmapTable) init(file io.ReadSeeker, entry *tableDirEntry) (err os.Error) {
	_, err = file.Seek(int64(entry.offset), os.SEEK_SET)
	if err != nil {
		return
	}

	if err = binary.Read(file, binary.BigEndian, &table.version); err != nil {
		return
	}
	if err = binary.Read(file, binary.BigEndian, &table.numberSubtables); err != nil {
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
	return
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
	length             uint16
	language           uint16
	segCountX2         uint16
	searchRange        uint16
	entrySelector      uint16
	rangeShift         uint16
	endCode            []uint16
	reservedPad        uint16
	startCode          []uint16
	idDelta            []uint16
	idRangeOffset      []uint16
	glyphIndexArray    []uint16
}

func (rec *cmapEncodingRecord) read(file io.Reader) (err os.Error) {
	if err = binary.Read(file, binary.BigEndian, &rec.platformID); err != nil {
		return
	}
	if err = binary.Read(file, binary.BigEndian, &rec.platformSpecificID); err != nil {
		return
	}
	if err = binary.Read(file, binary.BigEndian, &rec.offset); err != nil {
		return
	}
	return
}

func (rec *cmapEncodingRecord) readMapping(file io.Reader) (err os.Error) {
	if err = binary.Read(file, binary.BigEndian, &rec.format); err != nil {
		return
	}
	switch rec.format {
	case 0:
		err = rec.readMappingFormat0(file)
	case 4:
		err = rec.readMappingFormat4(file)
	default:
		return os.NewError(fmt.Sprintf("Unsupported mapping table format: %d", rec.format))
	}
	return
}

func (rec *cmapEncodingRecord) readMappingFormat0(file io.Reader) (err os.Error) {
	if err = binary.Read(file, binary.BigEndian, &rec.length); err != nil {
		return
	}
	if err = binary.Read(file, binary.BigEndian, &rec.language); err != nil {
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

func (rec *cmapEncodingRecord) readMappingFormat4(file io.Reader) (err os.Error) {
	if err = binary.Read(file, binary.BigEndian, &rec.length); err != nil {
		return
	}
	if err = binary.Read(file, binary.BigEndian, &rec.language); err != nil {
		return
	}
	if err = binary.Read(file, binary.BigEndian, &rec.segCountX2); err != nil {
		return
	}
	if err = binary.Read(file, binary.BigEndian, &rec.searchRange); err != nil {
		return
	}
	if err = binary.Read(file, binary.BigEndian, &rec.entrySelector); err != nil {
		return
	}
	if err = binary.Read(file, binary.BigEndian, &rec.rangeShift); err != nil {
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

func (rec *cmapEncodingRecord) write(wr io.Writer) {
	fmt.Fprintln(wr, "----------")
	fmt.Fprintln(wr, "cmap encoding")
	fmt.Fprintf(wr, "platformID = %d\n", rec.platformID)
	fmt.Fprintf(wr, "platformSpecificID = %d\n", rec.platformSpecificID)
	fmt.Fprintf(wr, "offset = %d\n", rec.offset)
	fmt.Fprintf(wr, "format = %d\n", rec.format)
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
