// Copyright 2011-2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ttf

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

type vmtxTable struct {
	vMetrics       []longVerMetric
	topSideBearing []FWord
}

func (table *vmtxTable) init(file io.ReadSeeker, entry *tableDirEntry, numGlyphs uint16, numOfLongVerMetrics uint16) (err error) {
	if numOfLongVerMetrics > numGlyphs {
		return fmt.Errorf("vmtx: number of long metrics %d exceeds glyph count %d", numOfLongVerMetrics, numGlyphs)
	}
	numBearings := int(numGlyphs - numOfLongVerMetrics)
	required := int(numOfLongVerMetrics)*4 + numBearings*2
	if uint64(required) > uint64(entry.length) {
		return fmt.Errorf("vmtx: table length %d is shorter than required %d", entry.length, required)
	}
	if _, err = file.Seek(int64(entry.offset), os.SEEK_SET); err != nil {
		return
	}
	data := make([]byte, required)
	if _, err = io.ReadFull(file, data); err != nil {
		return
	}

	table.vMetrics = make([]longVerMetric, numOfLongVerMetrics)
	pos := 0
	for i := range table.vMetrics {
		table.vMetrics[i] = longVerMetric{
			advanceHeight:  binary.BigEndian.Uint16(data[pos:]),
			topSideBearing: int16(binary.BigEndian.Uint16(data[pos+2:])),
		}
		pos += 4
	}
	table.topSideBearing = make([]FWord, numBearings)
	for i := range table.topSideBearing {
		table.topSideBearing[i] = FWord(int16(binary.BigEndian.Uint16(data[pos:])))
		pos += 2
	}
	return
}

func (table *vmtxTable) write(wr io.Writer) {
	fmt.Fprintln(wr, "----------")
	fmt.Fprintln(wr, "vmtx Table")
	fmt.Fprintf(wr, "vMetrics (%d)\n", len(table.vMetrics))
	for i, m := range table.vMetrics {
		fmt.Fprintf(wr, "[%d] advanceHeight: %d, topSideBearing: %d\n", i, m.advanceHeight, m.topSideBearing)
	}
	fmt.Fprintf(wr, "topSideBearing (%d)\n", len(table.topSideBearing))
	for i, b := range table.topSideBearing {
		fmt.Fprintf(wr, "[%d] %d\n", i, b)
	}
}

type longVerMetric struct {
	advanceHeight  uint16
	topSideBearing int16
}
