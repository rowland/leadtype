package ttf

import (
	"fmt"
	"io"
	"os"
)

type os2Table struct {
	// Version 0, per http://developer.apple.com/fonts/TTRefMan/RM06/Chap6OS2.html
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
	ulCharRange         CharRanges
	achVendID           [4]int8
	fsSelection         uint16
	fsFirstCharIndex    uint16
	fsLastCharIndex     uint16
	// Version 0, per http://www.microsoft.com/typography/otspec/os2ver0.htm
	sTypoAscender  int16
	sTypoDescender int16
	sTypoLineGap   int16
	usWinAscent    uint16
	usWinDescent   uint16
	// Version 1, per http://www.microsoft.com/typography/otspec/os2ver1.htm
	ulCodePageRange1 uint32 // Bits 0-31
	ulCodePageRange2 uint32 // Bits 32-63
	// Version 2, per http://www.microsoft.com/typography/otspec/os2ver2.htm
	sxHeight      int16
	sCapHeight    int16
	usDefaultChar uint16
	usBreakChar   uint16
	usMaxContext  uint16
	// Version 3, per http://www.microsoft.com/typography/otspec/os2ver3.htm (no new fields?)
	// Version 4, per http://www.microsoft.com/typography/otspec/os2.htm (no new fields?)
}

func (table *os2Table) init(file io.ReadSeeker, entry *tableDirEntry) (err error) {
	if _, err = file.Seek(int64(entry.offset), os.SEEK_SET); err != nil {
		return
	}
	// No advantage to using a buffered reader here.
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
		&table.sTypoAscender,
		&table.sTypoDescender,
		&table.sTypoLineGap,
		&table.usWinAscent,
		&table.usWinDescent,
	)
	if err == nil && table.version >= 1 {
		err = readValues(file,
			&table.ulCodePageRange1,
			&table.ulCodePageRange2,
		)
	}
	if err == nil && table.version >= 2 {
		err = readValues(file,
			&table.sxHeight,
			&table.sCapHeight,
			&table.usDefaultChar,
			&table.usBreakChar,
			&table.usMaxContext,
		)
	}
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
	fmt.Fprintf(wr, "sTypoAscender       = %d\n", table.sTypoAscender)
	fmt.Fprintf(wr, "sTypoDescender      = %d\n", table.sTypoDescender)
	fmt.Fprintf(wr, "sTypoLineGap        = %d\n", table.sTypoLineGap)
	fmt.Fprintf(wr, "usWinAscent         = %d\n", table.usWinAscent)
	fmt.Fprintf(wr, "usWinDescent        = %d\n", table.usWinDescent)
	fmt.Fprintf(wr, "ulCodePageRange1    = %d\n", table.ulCodePageRange1)
	fmt.Fprintf(wr, "ulCodePageRange2    = %d\n", table.ulCodePageRange2)
	fmt.Fprintf(wr, "sxHeight            = %d\n", table.sxHeight)
	fmt.Fprintf(wr, "sCapHeight          = %d\n", table.sCapHeight)
	fmt.Fprintf(wr, "usDefaultChar       = %d\n", table.usDefaultChar)
	fmt.Fprintf(wr, "usBreakChar         = %d\n", table.usBreakChar)
	fmt.Fprintf(wr, "usMaxContext        = %d\n", table.usMaxContext)
}

type CharRanges [4]uint32

func (cr *CharRanges) IsSet(bit int) bool {
	i := bit >> 5
	b := uint32(1 << uint(bit&31))
	return cr[i]&b != 0
}
