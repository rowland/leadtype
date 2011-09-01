package ttf

import (
	// "encoding/binary"
	"fmt"
	"io"
	"os"
)

type hheaTable struct {
	version             Fixed
	ascent              FWord
	descent             FWord
	lineGap             FWord
	advanceWidthMax     uFWord
	minLeftSideBearing  FWord
	minRightSideBearing FWord
	xMaxExtent          FWord
	caretSlopeRise      int16
	caretSlopeRun       int16
	caretOffset         FWord
	reserved1           int16
	reserved2           int16
	reserved3           int16
	reserved4           int16
	metricDataFormat    int16
	numOfLongHorMetrics uint16
}

func (table *hheaTable) init(file io.ReadSeeker, entry *tableDirEntry) (err os.Error) {
	if _, err = file.Seek(int64(entry.offset), os.SEEK_SET); err != nil {
		return
	}

	if err = table.version.Read(file); err != nil {
		return
	}
	err = readValues(file,
		&table.ascent,
		&table.descent,
		&table.lineGap,
		&table.advanceWidthMax,
		&table.minLeftSideBearing,
		&table.minRightSideBearing,
		&table.xMaxExtent,
		&table.caretSlopeRise,
		&table.caretSlopeRun,
		&table.caretOffset,
		&table.reserved1,
		&table.reserved2,
		&table.reserved3,
		&table.reserved4,
		&table.metricDataFormat,
		&table.numOfLongHorMetrics,
	)
	return
}

func (table *hheaTable) write(wr io.Writer) {
	fmt.Fprintln(wr, "----------")
	fmt.Fprintln(wr, "hhea Table")
	fmt.Fprintf(wr, "version = %g\n", table.version.Tof64())
	fmt.Fprintf(wr, "ascent  = %d\n", table.ascent)
	fmt.Fprintf(wr, "descent  = %d\n", table.descent)
	fmt.Fprintf(wr, "lineGap  = %d\n", table.lineGap)
	fmt.Fprintf(wr, "advanceWidthMax  = %d\n", table.advanceWidthMax)
	fmt.Fprintf(wr, "minLeftSideBearing  = %d\n", table.minLeftSideBearing)
	fmt.Fprintf(wr, "minRightSideBearing  = %d\n", table.minRightSideBearing)
	fmt.Fprintf(wr, "xMaxExtent  = %d\n", table.xMaxExtent)
	fmt.Fprintf(wr, "caretSlopeRise  = %d\n", table.caretSlopeRise)
	fmt.Fprintf(wr, "caretSlopeRun  = %d\n", table.caretSlopeRun)
	fmt.Fprintf(wr, "caretOffset  = %d\n", table.caretOffset)
	fmt.Fprintf(wr, "reserved1  = %d\n", table.reserved1)
	fmt.Fprintf(wr, "reserved2  = %d\n", table.reserved2)
	fmt.Fprintf(wr, "reserved3  = %d\n", table.reserved3)
	fmt.Fprintf(wr, "reserved4  = %d\n", table.reserved4)
	fmt.Fprintf(wr, "metricDataFormat  = %d\n", table.metricDataFormat)
	fmt.Fprintf(wr, "numOfLongHorMetrics  = %d\n", table.numOfLongHorMetrics)
}
