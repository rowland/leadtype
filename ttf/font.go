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

type nameTable struct {
	format       uint16
	count        uint16
	stringOffset uint16
	nameRecords  []nameRecord
}

func (table *nameTable) init(file io.ReadSeeker, entry *tableDirEntry) (err os.Error) {
	_, err = file.Seek(int64(entry.offset), os.SEEK_SET)
	if err != nil {
		return
	}
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
	fmt.Fprintln(wr, "name Table")
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

func (rec *nameRecord) read(file io.Reader) (err os.Error) {
	err = binary.Read(file, binary.BigEndian, &rec.platformID)
	if err != nil {
		return
	}
	err = binary.Read(file, binary.BigEndian, &rec.platformSpecificID)
	if err != nil {
		return
	}
	err = binary.Read(file, binary.BigEndian, &rec.languageID)
	if err != nil {
		return
	}
	err = binary.Read(file, binary.BigEndian, &rec.nameID)
	if err != nil {
		return
	}
	err = binary.Read(file, binary.BigEndian, &rec.length)
	if err != nil {
		return
	}
	err = binary.Read(file, binary.BigEndian, &rec.offset)
	return
}

type postTable struct {
	formatBase         int16
	formatFrac         uint16
	italicAngleBase    int16
	italicAngleFrac    uint16
	underlinePosition  int16
	underlineThickness int16
	isFixedPitch       uint32
	minMemType42       uint32
	maxMemType42       uint32
	minMemType1        uint32
	maxMemType1        uint32
	names              []string
}

func float64FromFixed(base int16, frac uint16) float64 {
	return float64(base) + float64(frac)/65536
}
func (table *postTable) format() float64 {
	return float64FromFixed(table.formatBase, table.formatFrac)
}

func (table *postTable) italicAngle() float64 {
	return float64FromFixed(table.italicAngleBase, table.italicAngleFrac)
}

func (table *postTable) init(file io.ReadSeeker, entry *tableDirEntry) (err os.Error) {
	_, err = file.Seek(int64(entry.offset), os.SEEK_SET)
	if err != nil {
		return
	}

	if err = binary.Read(file, binary.BigEndian, &table.formatBase); err != nil {
		return
	}
	if err = binary.Read(file, binary.BigEndian, &table.formatFrac); err != nil {
		return
	}
	if err = binary.Read(file, binary.BigEndian, &table.italicAngleBase); err != nil {
		return
	}
	if err = binary.Read(file, binary.BigEndian, &table.italicAngleFrac); err != nil {
		return
	}
	if err = binary.Read(file, binary.BigEndian, &table.underlinePosition); err != nil {
		return
	}
	if err = binary.Read(file, binary.BigEndian, &table.underlineThickness); err != nil {
		return
	}
	if err = binary.Read(file, binary.BigEndian, &table.isFixedPitch); err != nil {
		return
	}
	if err = binary.Read(file, binary.BigEndian, &table.minMemType42); err != nil {
		return
	}
	if err = binary.Read(file, binary.BigEndian, &table.maxMemType42); err != nil {
		return
	}
	if err = binary.Read(file, binary.BigEndian, &table.minMemType1); err != nil {
		return
	}
	if err = binary.Read(file, binary.BigEndian, &table.maxMemType1); err != nil {
		return
	}

	switch table.format() {
	case 1.0:
		table.names = MacRomanPostNames
	case 2.0:
		err = table.readFormat2Names(file)
	case 3.0:
		// no subtable for format 3
	default:
		return os.NewError(fmt.Sprintf("Unsupported post table format: %g", table.format()))
	}

	return
}

func (table *postTable) readFormat2Names(file io.ReadSeeker) (err os.Error) {
	var numberOfGlyphs uint16
	if err = binary.Read(file, binary.BigEndian, &numberOfGlyphs); err != nil {
		return
	}
	glyphNameIndex := make([]uint16, numberOfGlyphs)
	table.names = make([]string, numberOfGlyphs)
	if err = binary.Read(file, binary.BigEndian, glyphNameIndex); err != nil {
		return
	}
	numberNewGlyphs := numberOfGlyphs
	for i := uint16(0); i < numberOfGlyphs; i++ {
		if glyphNameIndex[i] < 258 {
			table.names[i] = MacRomanPostNames[glyphNameIndex[i]]
			numberNewGlyphs--
		}
	}
	newNames := make([]string, numberNewGlyphs)
	for i := uint16(0); i < numberNewGlyphs; i++ {
		var slen byte
		if err = binary.Read(file, binary.BigEndian, &slen); err != nil {
			return
		}
		s := make([]byte, slen)
		if _, err = file.Read(s); err != nil {
			return
		}
		newNames[i] = string(s)
	}
	for i := uint16(0); i < numberOfGlyphs; i++ {
		if glyphNameIndex[i] >= 258 {
			table.names[i] = newNames[glyphNameIndex[i]-258]
		}
	}
	return
}

func (table *postTable) write(wr io.Writer) {
	fmt.Fprintln(wr, "----------")
	fmt.Fprintln(wr, "post Table")
	fmt.Fprintf(wr, "format = %g\n", table.format())
	fmt.Fprintf(wr, "italicAngle = %g\n", table.italicAngle())
	fmt.Fprintf(wr, "underlinePosition  = %d\n", table.underlinePosition)
	fmt.Fprintf(wr, "underlineThickness = %d\n", table.underlineThickness)
	fmt.Fprintf(wr, "isFixedPitch = %d\n", table.isFixedPitch)
	fmt.Fprintf(wr, "minMemType42 = %d\n", table.minMemType42)
	fmt.Fprintf(wr, "maxMemType42 = %d\n", table.maxMemType42)
	fmt.Fprintf(wr, "minMemType1 = %d\n", table.minMemType1)
	fmt.Fprintf(wr, "maxMemType1 = %d\n", table.maxMemType1)
	fmt.Fprintf(wr, "numberOfGlyphs = %d\n", len(table.names))
	for i, name := range table.names {
		fmt.Fprintf(wr, "[%d] %s\n", i, name)
	}
}
