package ttf

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

type hmtxTable struct {
	hMetrics        []longHorMetric
	leftSideBearing []FWord
}

func (table *hmtxTable) init(file io.ReadSeeker, entry *tableDirEntry, numGlyphs uint16, numOfLongHorMetrics uint16) (err os.Error) {
	if _, err = file.Seek(int64(entry.offset), os.SEEK_SET); err != nil {
		return
	}
	table.hMetrics = make([]longHorMetric, numOfLongHorMetrics)
	for i, _ := range table.hMetrics {
		if err = table.hMetrics[i].read(file); err != nil {
			return
		}
	}
	table.leftSideBearing = make([]FWord, numGlyphs-numOfLongHorMetrics)
	err = binary.Read(file, binary.BigEndian, table.leftSideBearing)
	return
}

func (table *hmtxTable) write(wr io.Writer) {
	fmt.Fprintln(wr, "----------")
	fmt.Fprintln(wr, "hmtx Table")
	fmt.Fprintf(wr, "hMetrics (%d)\n", len(table.hMetrics))
	for i, m := range table.hMetrics {
		fmt.Fprintf(wr, "[%d] advanceWidth: %d, leftSideBearing: %d\n", i, m.advanceWidth, m.leftSideBearing)
	}
	fmt.Fprintf(wr, "leftSideBearing (%d)\n", len(table.leftSideBearing))
	for i, b := range table.leftSideBearing {
		fmt.Fprintf(wr, "[%d] %d\n", i, b)
	}
}

type longHorMetric struct {
	advanceWidth    uint16
	leftSideBearing int16
}

func (m *longHorMetric) read(file io.Reader) os.Error {
	return readValues(file, &m.advanceWidth, &m.leftSideBearing)
}
