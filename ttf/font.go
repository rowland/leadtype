// Read some of the tables specified in http://developer.apple.com/fonts/ttrefman/

package ttf

import (
	"bytes"
	"encoding/binary"
	"fmt"
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
	cmapTable     cmapTable
	headTable     headTable
	nameTable     nameTable
	postTable     postTable
}

func LoadFont(filename string) (font *Font, err os.Error) {
	var file *os.File
	file, err = os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()
	font = new(Font)
	err = font.init(file)
	return
}

func (font *Font) init(file io.ReadSeeker) (err os.Error) {
	if err = binary.Read(file, binary.BigEndian, &font.scalar); err != nil {
		return
	}
	if err = binary.Read(file, binary.BigEndian, &font.nTables); err != nil {
		return
	}
	if err = binary.Read(file, binary.BigEndian, &font.searchRange); err != nil {
		return
	}
	if err = binary.Read(file, binary.BigEndian, &font.entrySelector); err != nil {
		return
	}
	if err = binary.Read(file, binary.BigEndian, &font.rangeShift); err != nil {
		return
	}

	if err = font.tableDir.read(file, font.nTables); err != nil {
		return
	}
	if entry := font.tableDir.table("cmap"); entry != nil {
		if err = font.cmapTable.init(file, entry); err != nil {
			return
		}
	}
	if entry := font.tableDir.table("head"); entry != nil {
		if err = font.headTable.init(file, entry); err != nil {
			return
		}
	}
	if entry := font.tableDir.table("name"); entry != nil {
		if err = font.nameTable.init(file, entry); err != nil {
			return
		}
	}
	if entry := font.tableDir.table("post"); entry != nil {
		if err = font.postTable.init(file, entry); err != nil {
			return
		}
	}
	return
}

func (font *Font) String() string {
	var buf bytes.Buffer
	font.write(&buf)
	return buf.String()
}

func (font *Font) write(wr io.Writer) {
	fmt.Fprintf(wr, "scalar = %d\n", font.scalar)
	fmt.Fprintf(wr, "nTables = %d\n", font.nTables)
	fmt.Fprintf(wr, "searchRange = %d\n", font.searchRange)
	fmt.Fprintf(wr, "entrySelector = %d\n", font.entrySelector)
	fmt.Fprintf(wr, "rangeShift = %d\n", font.rangeShift)
	font.tableDir.write(wr)
	font.nameTable.write(wr)
	font.postTable.write(wr)
	font.cmapTable.write(wr)
	font.headTable.write(wr)
}

type tableDir struct {
	entries    []*tableDirEntry
	entriesMap map[string]*tableDirEntry
}

func (dir *tableDir) read(file io.Reader, nTables uint16) (err os.Error) {
	dir.entries = make([]*tableDirEntry, nTables)
	dir.entriesMap = make(map[string]*tableDirEntry, nTables)
	for i := uint16(0); i < nTables; i++ {
		var entry tableDirEntry
		if err = entry.read(file); err != nil {
			return
		}
		dir.entries[i] = &entry
		dir.entriesMap[entry.tag] = &entry
	}
	return
}

func (dir *tableDir) String() string {
	var buf bytes.Buffer
	dir.write(&buf)
	return buf.String()
}

func (dir *tableDir) write(wr io.Writer) {
	for _, e := range dir.entries {
		e.write(wr)
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

func (entry *tableDirEntry) read(file io.Reader) (err os.Error) {
	tag := make([]byte, 4)
	_, err = file.Read(tag)
	entry.tag = string(tag)
	if err = binary.Read(file, binary.BigEndian, &entry.checkSum); err != nil {
		return
	}
	if err = binary.Read(file, binary.BigEndian, &entry.offset); err != nil {
		return
	}
	if err = binary.Read(file, binary.BigEndian, &entry.length); err != nil {
		return
	}
	return
}

func (entry *tableDirEntry) String() string {
	var buf bytes.Buffer
	entry.write(&buf)
	return buf.String()
}

func (entry *tableDirEntry) write(wr io.Writer) {
	fmt.Fprintln(wr, "----------")
	fmt.Fprintln(wr, "Table")
	fmt.Fprintf(wr, "tag = %s\n", entry.tag)
	fmt.Fprintf(wr, "checkSum = %d\n", entry.checkSum)
	fmt.Fprintf(wr, "offset = %d\n", entry.offset)
	fmt.Fprintf(wr, "length = %d\n", entry.length)
}
