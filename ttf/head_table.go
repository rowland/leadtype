package ttf

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

type headTable struct {
	version            Fixed
	fontRevision       Fixed
	checkSumAdjustment uint32
	magicNumber        uint32
	flags              uint16
	unitsPerEm         uint16
	created            longDateTime
	modified           longDateTime
	xMin               FWord
	yMin               FWord
	xMax               FWord
	yMax               FWord
	macStyle           uint16
	lowestRecPPEM      uint16
	fontDirectionHint  int16
	indexToLocFormat   int16
	glyphDataFormat    int16
}

func (table *headTable) init(file io.ReadSeeker, entry *tableDirEntry) (err os.Error) {
	_, err = file.Seek(int64(entry.offset), os.SEEK_SET)
	if err != nil {
		return
	}

	if err = table.version.Read(file); err != nil {
		return
	}
	if err = table.fontRevision.Read(file); err != nil {
		return
	}
	if err = binary.Read(file, binary.BigEndian, &table.checkSumAdjustment); err != nil {
		return
	}
	if err = binary.Read(file, binary.BigEndian, &table.magicNumber); err != nil {
		return
	}
	if err = binary.Read(file, binary.BigEndian, &table.flags); err != nil {
		return
	}
	if err = binary.Read(file, binary.BigEndian, &table.unitsPerEm); err != nil {
		return
	}
	if err = binary.Read(file, binary.BigEndian, &table.created); err != nil {
		return
	}
	if err = binary.Read(file, binary.BigEndian, &table.modified); err != nil {
		return
	}
	if err = binary.Read(file, binary.BigEndian, &table.xMin); err != nil {
		return
	}
	if err = binary.Read(file, binary.BigEndian, &table.yMin); err != nil {
		return
	}
	if err = binary.Read(file, binary.BigEndian, &table.xMax); err != nil {
		return
	}
	if err = binary.Read(file, binary.BigEndian, &table.yMax); err != nil {
		return
	}
	if err = binary.Read(file, binary.BigEndian, &table.macStyle); err != nil {
		return
	}
	if err = binary.Read(file, binary.BigEndian, &table.lowestRecPPEM); err != nil {
		return
	}
	if err = binary.Read(file, binary.BigEndian, &table.fontDirectionHint); err != nil {
		return
	}
	if err = binary.Read(file, binary.BigEndian, &table.indexToLocFormat); err != nil {
		return
	}
	if err = binary.Read(file, binary.BigEndian, &table.glyphDataFormat); err != nil {
		return
	}
	return
}

func (table *headTable) write(wr io.Writer) {
	fmt.Fprintln(wr, "----------")
	fmt.Fprintln(wr, "head Table")
	fmt.Fprintf(wr, "version = %g\n", table.version.Tof64())
	fmt.Fprintf(wr, "fontRevision = %g\n", table.fontRevision.Tof64())
	fmt.Fprintf(wr, "checkSumAdjustment  = %d\n", table.checkSumAdjustment)
	fmt.Fprintf(wr, "magicNumber = %X\n", table.magicNumber)
	fmt.Fprintf(wr, "flags = %b\n", table.flags)
	fmt.Fprintf(wr, "unitsPerEm = %d\n", table.unitsPerEm)

	fmt.Fprintf(wr, "created = %d\n", table.created)
	fmt.Fprintf(wr, "modified = %d\n", table.modified)

	fmt.Fprintf(wr, "xMin = %d\n", table.xMin)
	fmt.Fprintf(wr, "yMin = %d\n", table.yMin)
	fmt.Fprintf(wr, "xMax = %d\n", table.xMax)
	fmt.Fprintf(wr, "yMax = %d\n", table.yMax)

	fmt.Fprintf(wr, "macStyle = %d\n", table.macStyle)
	fmt.Fprintf(wr, "lowestRecPPEM = %d\n", table.lowestRecPPEM)
	fmt.Fprintf(wr, "fontDirectionHint = %d\n", table.fontDirectionHint)
	fmt.Fprintf(wr, "indexToLocFormat = %d\n", table.indexToLocFormat)
	fmt.Fprintf(wr, "glyphDataFormat = %d\n", table.glyphDataFormat)
}
