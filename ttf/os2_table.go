package ttf

import (
	"fmt"
	"io"
	"os"
)

type os2Table struct {
	version             uint16
	xAvgCharWidth       int16
	usWeightClass       uint16
	usWidthClass        uint16
	fsType              int16
	ySubscriptXSize     int16
	ySubscriptYSize     int16
	ySubscriptXOffset   int16
	ySubscriptYOffset   int16
	ySuperscriptXSize   int16
	ySuperscriptYSize   int16
	ySuperscriptXOffset int16
	ySuperscriptYOffset int16
	yStrikeoutSize      int16
	yStrikeoutPosition  int16
	sFamilyClass        int16
	panose              PANOSE
	ulCharRange         [4]uint32
	achVendID           [4]int8
	fsSelection         uint16
	fsFirstCharIndex    uint16
	fsLastCharIndex     uint16
}

func (table *os2Table) init(file io.ReadSeeker, entry *tableDirEntry) (err os.Error) {
	if _, err = file.Seek(int64(entry.offset), os.SEEK_SET); err != nil {
		return
	}

	err = readValues(file,
		&table.version,
		&table.xAvgCharWidth,
		&table.usWeightClass,
		&table.usWidthClass,
		&table.fsType,
		&table.ySubscriptXSize,
		&table.ySubscriptYSize,
		&table.ySubscriptXOffset,
		&table.ySubscriptYOffset,
		&table.ySuperscriptXSize,
		&table.ySuperscriptYSize,
		&table.ySuperscriptXOffset,
		&table.ySuperscriptYOffset,
		&table.yStrikeoutSize,
		&table.yStrikeoutPosition,
		&table.sFamilyClass,
		&table.panose,
		&table.ulCharRange,
		&table.achVendID,
		&table.fsSelection,
		&table.fsFirstCharIndex,
		&table.fsLastCharIndex,
	)
	return
}

func (table *os2Table) write(wr io.Writer) {
	fmt.Fprintln(wr, "----------")
	fmt.Fprintln(wr, "OS/2 Table")
	fmt.Fprintf(wr, "version             = %d\n", table.version)
	fmt.Fprintf(wr, "xAvgCharWidth       = %d\n", table.xAvgCharWidth)
	fmt.Fprintf(wr, "usWeightClass       = %d\n", table.usWeightClass)
	fmt.Fprintf(wr, "usWidthClass        = %d\n", table.usWidthClass)
	fmt.Fprintf(wr, "fsType              = %d\n", table.fsType)
	fmt.Fprintf(wr, "ySubscriptXSize     = %d\n", table.ySubscriptXSize)
	fmt.Fprintf(wr, "ySubscriptYSize     = %d\n", table.ySubscriptYSize)
	fmt.Fprintf(wr, "ySubscriptXOffset   = %d\n", table.ySubscriptXOffset)
	fmt.Fprintf(wr, "ySubscriptYOffset   = %d\n", table.ySubscriptYOffset)
	fmt.Fprintf(wr, "ySuperscriptXSize   = %d\n", table.ySuperscriptXSize)
	fmt.Fprintf(wr, "ySuperscriptYSize   = %d\n", table.ySuperscriptYSize)
	fmt.Fprintf(wr, "ySuperscriptXOffset = %d\n", table.ySuperscriptXOffset)
	fmt.Fprintf(wr, "ySuperscriptYOffset = %d\n", table.ySuperscriptYOffset)
	fmt.Fprintf(wr, "yStrikeoutSize      = %d\n", table.yStrikeoutSize)
	fmt.Fprintf(wr, "yStrikeoutPosition  = %d\n", table.yStrikeoutPosition)
	fmt.Fprintf(wr, "sFamilyClass        = %d\n", table.sFamilyClass)
	fmt.Fprintf(wr, "panose              = %d\n", table.panose)
	fmt.Fprintf(wr, "ulCharRange         = %d\n", table.ulCharRange)
	fmt.Fprintf(wr, "achVendID           = %d\n", table.achVendID)
	fmt.Fprintf(wr, "fsSelection         = %d\n", table.fsSelection)
	fmt.Fprintf(wr, "fsFirstCharIndex    = %d\n", table.fsFirstCharIndex)
	fmt.Fprintf(wr, "fsLastCharIndex     = %d\n", table.fsLastCharIndex)
}
