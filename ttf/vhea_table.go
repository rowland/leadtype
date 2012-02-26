package ttf

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

type vheaTable struct {
	version              Fixed
	vertTypoAscender     int16
	vertTypoDescender    int16
	vertTypoLineGap      int16
	advanceHeightMax     int16
	minTopSideBearing    int16
	minBottomSideBearing int16
	yMaxExtent           int16
	caretSlopeRise       int16
	caretSlopeRun        int16
	caretOffset          int16
	reserved1            int16
	reserved2            int16
	reserved3            int16
	reserved4            int16
	metricDataFormat     int16
	numOfLongVerMetrics  uint16
}

func (table *vheaTable) init(rs io.ReadSeeker, entry *tableDirEntry) (err error) {
	if _, err = rs.Seek(int64(entry.offset), os.SEEK_SET); err != nil {
		return
	}
	file := bufio.NewReaderSize(rs, int(entry.length))
	if err = table.version.Read(file); err != nil {
		return
	}
	err = readValues(file,
		&table.vertTypoAscender,
		&table.vertTypoDescender,
		&table.vertTypoLineGap,
		&table.advanceHeightMax,
		&table.minTopSideBearing,
		&table.minBottomSideBearing,
		&table.yMaxExtent,
		&table.caretSlopeRise,
		&table.caretSlopeRun,
		&table.caretOffset,
		&table.reserved1,
		&table.reserved2,
		&table.reserved3,
		&table.reserved4,
		&table.metricDataFormat,
		&table.numOfLongVerMetrics,
	)
	return
}

func (table *vheaTable) write(wr io.Writer) {
	fmt.Fprintln(wr, "----------")
	fmt.Fprintln(wr, "vhea Table")
	fmt.Fprintf(wr, "version              = %g\n", table.version.Tof64())
	fmt.Fprintf(wr, "vertTypoAscender     = %d\n", table.vertTypoAscender)
	fmt.Fprintf(wr, "vertTypoDescender    = %d\n", table.vertTypoDescender)
	fmt.Fprintf(wr, "vertTypoLineGap      = %d\n", table.vertTypoLineGap)
	fmt.Fprintf(wr, "advanceHeightMax     = %d\n", table.advanceHeightMax)
	fmt.Fprintf(wr, "minTopSideBearing    = %d\n", table.minTopSideBearing)
	fmt.Fprintf(wr, "minBottomSideBearing = %d\n", table.minBottomSideBearing)
	fmt.Fprintf(wr, "yMaxExtent           = %d\n", table.yMaxExtent)
	fmt.Fprintf(wr, "caretSlopeRise       = %d\n", table.caretSlopeRise)
	fmt.Fprintf(wr, "caretSlopeRun        = %d\n", table.caretSlopeRun)
	fmt.Fprintf(wr, "caretOffset          = %d\n", table.caretOffset)
	fmt.Fprintf(wr, "reserved1            = %d\n", table.reserved1)
	fmt.Fprintf(wr, "reserved2            = %d\n", table.reserved2)
	fmt.Fprintf(wr, "reserved3            = %d\n", table.reserved3)
	fmt.Fprintf(wr, "reserved4            = %d\n", table.reserved4)
	fmt.Fprintf(wr, "metricDataFormat     = %d\n", table.metricDataFormat)
	fmt.Fprintf(wr, "numOfLongVerMetrics  = %d\n", table.numOfLongVerMetrics)
}
