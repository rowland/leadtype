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

func (table *vmtxTable) init(file io.ReadSeeker, entry *tableDirEntry, numGlyphs uint16, numOfLongVerMetrics uint16) (err os.Error) {
	if _, err = file.Seek(int64(entry.offset), os.SEEK_SET); err != nil {
		return
	}
	table.vMetrics = make([]longVerMetric, numOfLongVerMetrics)
	for i, _ := range table.vMetrics {
		if err = table.vMetrics[i].read(file); err != nil {
			return
		}
	}
	table.topSideBearing = make([]FWord, numGlyphs-numOfLongVerMetrics)
	err = binary.Read(file, binary.BigEndian, table.topSideBearing)
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

func (m *longVerMetric) read(file io.Reader) os.Error {
	return readValues(file, &m.advanceHeight, &m.topSideBearing)
}
