package ttf

import (
	"encoding/binary"
	"io"
	"os"
)

type Font struct {
	scalar        uint32
	nTables       uint16
	searchRange   uint16
	entrySelector uint16
	rangeShift    uint16
	tableDir      tableDir
}

func LoadFont(filename string) (font *Font, err os.Error) {
	var file *os.File
	file, err = os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()
	font = new(Font)

	binary.Read(file, binary.BigEndian, &font.scalar)
	binary.Read(file, binary.BigEndian, &font.nTables)
	binary.Read(file, binary.BigEndian, &font.searchRange)
	binary.Read(file, binary.BigEndian, &font.entrySelector)
	binary.Read(file, binary.BigEndian, &font.rangeShift)

	font.tableDir.read(file, font.nTables)
	return
}

type tableDir struct {
	entries    []*tableDirEntry
	entriesMap map[string]*tableDirEntry
}

func (dir *tableDir) read(file io.Reader, nTables uint16) {
	dir.entries = make([]*tableDirEntry, nTables)
	dir.entriesMap = make(map[string]*tableDirEntry, nTables)
	for i := uint16(0); i < nTables; i++ {
		var entry tableDirEntry
		entry.read(file)
		dir.entries[i] = &entry
		dir.entriesMap[entry.tag] = &entry
	}
}

func (dir *tableDir) table(tag string) *tableDirEntry {
	return dir.entriesMap[tag]
}

type tableDirEntry struct {
	tag      string
	checkSum uint32
	offset   uint32
	length   uint32
}

func (entry *tableDirEntry) read(file io.Reader) {
	tag := make([]byte, 4)
	file.Read(tag)
	entry.tag = string(tag)
	binary.Read(file, binary.BigEndian, &entry.checkSum)
	binary.Read(file, binary.BigEndian, &entry.offset)
	binary.Read(file, binary.BigEndian, &entry.length)
}
