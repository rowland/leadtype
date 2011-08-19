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
	nameTable     nameTable
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
	binary.Read(file, binary.BigEndian, &font.scalar)
	binary.Read(file, binary.BigEndian, &font.nTables)
	binary.Read(file, binary.BigEndian, &font.searchRange)
	binary.Read(file, binary.BigEndian, &font.entrySelector)
	binary.Read(file, binary.BigEndian, &font.rangeShift)

	font.tableDir.read(file, font.nTables)
	err = font.readNameTable(file)
	return
}

func (font *Font) readNameTable(file io.ReadSeeker) (err os.Error) {
	entry := font.tableDir.table("name")
	if entry != nil {
		err = font.nameTable.init(file, entry)
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

func (entry *tableDirEntry) read(file io.Reader) {
	tag := make([]byte, 4)
	file.Read(tag)
	entry.tag = string(tag)
	binary.Read(file, binary.BigEndian, &entry.checkSum)
	binary.Read(file, binary.BigEndian, &entry.offset)
	binary.Read(file, binary.BigEndian, &entry.length)
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

type nameTable struct {
	format       uint16
	count        uint16
	stringOffset uint16
	nameRecords  []nameRecord
}

func (table *nameTable) init(file io.ReadSeeker, entry *tableDirEntry) (err os.Error) {
	file.Seek(int64(entry.offset), os.SEEK_SET)
	bs := make([]byte, entry.length)
	_, err = file.Read(bs)
	if err != nil {
		return err
	}
	
	buf := bytes.NewBuffer(bs)
	binary.Read(buf, binary.BigEndian, &table.format)
	binary.Read(buf, binary.BigEndian, &table.count)
	binary.Read(buf, binary.BigEndian, &table.stringOffset)
	table.nameRecords = make([]nameRecord, table.count)
	for i := uint16(0); i < table.count; i++ {
		table.nameRecords[i].read(buf)
	}

	return
}

func (table *nameTable) write(wr io.Writer) {
	fmt.Fprintln(wr, "----------")
	fmt.Fprintln(wr, "Name Table")
	fmt.Fprintf(wr, "format = %d\n", table.format)
	fmt.Fprintf(wr, "count = %d\n", table.format)
	fmt.Fprintf(wr, "stringOffset = %d\n", table.stringOffset)
}

type nameRecord struct {
	platformID         uint16
	platformSpecificID uint16
	languageID         uint16
	nameID             uint16
	length             uint16
	offset             uint16
}

func (rec *nameRecord) read(file io.Reader) {
	binary.Read(file, binary.BigEndian, &rec.platformID)
	binary.Read(file, binary.BigEndian, &rec.platformSpecificID)
	binary.Read(file, binary.BigEndian, &rec.languageID)
	binary.Read(file, binary.BigEndian, &rec.nameID)
	binary.Read(file, binary.BigEndian, &rec.length)
	binary.Read(file, binary.BigEndian, &rec.offset)
}
