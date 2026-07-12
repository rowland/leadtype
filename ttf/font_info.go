// Copyright 2011-2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ttf

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
)

type FontInfo struct {
	filename      string
	ttcOffset     int64
	scalar        uint32
	nTables       uint16
	searchRange   uint16
	entrySelector uint16
	rangeShift    uint16
	tableDir      tableDir
	nameTable     nameTable
	os2Table      os2Table
}

// 1,077,216 ns
func LoadFontInfo(filename string) (fi *FontInfo, err error) {
	return LoadFontInfoAtOffset(filename, 0)
}

func LoadFontInfoAtOffset(filename string, offset int64) (fi *FontInfo, err error) {
	var file *os.File
	if file, err = os.Open(filename); err != nil {
		return
	}
	defer file.Close()
	fi = new(FontInfo)
	fi.filename = filename
	fi.ttcOffset = offset
	err = fi.init(file, offset)
	return
}

func (fi *FontInfo) TTCOffset() int64 {
	return fi.ttcOffset
}

func (fi *FontInfo) init(file io.ReadSeeker, offset int64) (err error) {
	if offset != 0 {
		if _, err = file.Seek(offset, io.SeekStart); err != nil {
			return
		}
	}
	if err = readValues(file,
		&fi.scalar,
		&fi.nTables,
		&fi.searchRange,
		&fi.entrySelector,
		&fi.rangeShift); err != nil {
		return
	}

	if err = fi.tableDir.read(file, fi.nTables); err != nil {
		return
	}
	if entry := fi.tableDir.table("name"); entry != nil {
		if err = fi.nameTable.init(file, entry); err != nil {
			return
		}
	}
	if entry := fi.tableDir.table("OS/2"); entry != nil {
		if err = fi.os2Table.init(file, entry); err != nil {
			return
		}
	}
	return
}

func (fi *FontInfo) AvgWidth() int {
	return int(fi.os2Table.xAvgCharWidth)
}

func (fi *FontInfo) CapHeight() int {
	if fi.os2Table.version >= 2 {
		return int(fi.os2Table.sCapHeight)
	}
	return 0
}

func (fi *FontInfo) CharRanges() *CharRanges {
	return &fi.os2Table.ulCharRange
}

func (fi *FontInfo) Copyright() string {
	return fi.nameTable.copyrightNotice
}

func (fi *FontInfo) Designer() string {
	return fi.nameTable.designer
}

func (fi *FontInfo) Embeddable() bool {
	return fi.os2Table.fsType&RestrictedLicenseEmbedding == 0
}

func (fi *FontInfo) Family() string {
	return fi.nameTable.fontFamily
}

func (fi *FontInfo) HasTable(tag string) bool {
	return fi.tableDir.table(tag) != nil
}

// NumGlyphs reads maxp without loading the full font. It is used by font
// catalog diagnostics; zero means the table was absent or unreadable.
func (fi *FontInfo) NumGlyphs() int {
	entry := fi.tableDir.table("maxp")
	if entry == nil {
		return 0
	}
	data, err := fi.readTable(entry)
	if err != nil || len(data) < 6 {
		return 0
	}
	return int(binary.BigEndian.Uint16(data[4:6]))
}

func (fi *FontInfo) HasTrueTypeOutlines() bool {
	return fi.HasTable("glyf") && fi.HasTable("loca")
}

func (fi *FontInfo) HasCFFOutlines() bool {
	return fi.HasTable("CFF ")
}

func (fi *FontInfo) HasSupportedCFFOutlines() bool {
	return fi.HasCFFOutlines()
}

func (fi *FontInfo) HasCIDKeyedCFFOutlines() (bool, error) {
	entry := fi.tableDir.table("CFF ")
	if entry == nil {
		return false, nil
	}
	data, err := fi.readTable(entry)
	if err != nil {
		return false, err
	}
	return cffDataIsCIDKeyed(data)
}

func (fi *FontInfo) HasSupportedOutlines() bool {
	return fi.HasTrueTypeOutlines() || fi.HasSupportedCFFOutlines()
}

func (fi *FontInfo) HasOpenTypeShaping() bool {
	return fi.HasTable("GSUB") || fi.HasTable("GPOS")
}

func (fi *FontInfo) Filename() string {
	return fi.filename
}

func (fi *FontInfo) FullName() string {
	return fi.nameTable.fullName
}

func (fi *FontInfo) License() string {
	return fi.nameTable.licenseDescription
}

func (fi *FontInfo) Manufacturer() string {
	return fi.nameTable.manufacturerName
}

func (fi *FontInfo) PostScriptName() string {
	return fi.nameTable.postScriptName
}

func (fi *FontInfo) StemV() int {
	return 50 + int(math.Pow(float64(fi.os2Table.usWeightClass)/65.0, 2))
}

func (fi *FontInfo) Style() string {
	return fi.nameTable.fontSubfamily
}

func (fi *FontInfo) Trademark() string {
	return fi.nameTable.trademarkNotice
}

func (fi *FontInfo) UniqueName() string {
	return fi.nameTable.uniqueSubfamily
}

func (fi *FontInfo) Version() string {
	return fi.nameTable.version
}

func (fi *FontInfo) XHeight() int {
	if fi.os2Table.version >= 2 {
		return int(fi.os2Table.sxHeight)
	}
	return 0
}

// WeightClass returns OS/2.usWeightClass, normally a CSS-compatible value from
// 100 through 900. Closest-face selection treats zero as Regular (400).
func (fi *FontInfo) WeightClass() int {
	return int(fi.os2Table.usWeightClass)
}

func (fi *FontInfo) readTable(entry *tableDirEntry) ([]byte, error) {
	file, err := os.Open(fi.filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data := make([]byte, entry.length)
	if _, err := file.ReadAt(data, int64(entry.offset)); err != nil {
		return nil, err
	}
	return data, nil
}

func cffDataIsCIDKeyed(data []byte) (bool, error) {
	if len(data) < 4 {
		return false, fmt.Errorf("cff: truncated header")
	}
	if data[0] != 1 || data[1] != 0 {
		return false, fmt.Errorf("cff: unsupported version %d.%d", data[0], data[1])
	}
	offset := int(data[2])
	if offset > len(data) {
		return false, fmt.Errorf("cff: invalid header size")
	}
	_, next, err := cffIndexItems(data, offset)
	if err != nil {
		return false, err
	}
	topDicts, _, err := cffIndexItems(data, next)
	if err != nil {
		return false, err
	}
	if len(topDicts) != 1 {
		return false, fmt.Errorf("cff: expected one top dict, got %d", len(topDicts))
	}
	return cffDictHasOperator(topDicts[0], []byte{12, 30})
}

func cffIndexItems(data []byte, offset int) ([][]byte, int, error) {
	if offset < 0 || offset+2 > len(data) {
		return nil, offset, fmt.Errorf("cff: invalid index offset")
	}
	count := int(binary.BigEndian.Uint16(data[offset:]))
	offset += 2
	if count == 0 {
		return nil, offset, nil
	}
	if offset >= len(data) {
		return nil, offset, fmt.Errorf("cff: truncated index")
	}
	offSize := int(data[offset])
	offset++
	if offSize < 1 || offSize > 4 {
		return nil, offset, fmt.Errorf("cff: invalid index offset size")
	}
	if offset+(count+1)*offSize > len(data) {
		return nil, offset, fmt.Errorf("cff: truncated index offsets")
	}
	offsets := make([]int, count+1)
	for i := range offsets {
		offsets[i] = cffOffset(data[offset+i*offSize:offset+(i+1)*offSize], offSize)
		if offsets[i] <= 0 {
			return nil, offset, fmt.Errorf("cff: invalid index object offset")
		}
	}
	objectStart := offset + (count+1)*offSize
	end := objectStart + offsets[count] - 1
	if end > len(data) {
		return nil, offset, fmt.Errorf("cff: index data out of bounds")
	}
	items := make([][]byte, count)
	for i := range items {
		start := objectStart + offsets[i] - 1
		stop := objectStart + offsets[i+1] - 1
		if start > stop || stop > len(data) {
			return nil, offset, fmt.Errorf("cff: invalid index object bounds")
		}
		items[i] = data[start:stop]
	}
	return items, end, nil
}

func cffOffset(data []byte, size int) int {
	n := 0
	for i := 0; i < size; i++ {
		n = n<<8 | int(data[i])
	}
	return n
}

func cffDictHasOperator(data []byte, target []byte) (bool, error) {
	for pos := 0; pos < len(data); {
		b := data[pos]
		if b <= 21 {
			if b == 12 {
				if pos+1 >= len(data) {
					return false, fmt.Errorf("cff: truncated escaped operator")
				}
				if bytes.Equal(data[pos:pos+2], target) {
					return true, nil
				}
				pos += 2
				continue
			}
			if bytes.Equal(data[pos:pos+1], target) {
				return true, nil
			}
			pos++
			continue
		}
		n, err := cffDictNumberLen(data[pos:])
		if err != nil {
			return false, err
		}
		pos += n
	}
	return false, nil
}

func cffDictNumberLen(data []byte) (int, error) {
	if len(data) == 0 {
		return 0, fmt.Errorf("cff: empty number")
	}
	b := data[0]
	switch {
	case b == 28:
		if len(data) < 3 {
			return 0, fmt.Errorf("cff: truncated int16")
		}
		return 3, nil
	case b == 29:
		if len(data) < 5 {
			return 0, fmt.Errorf("cff: truncated int32")
		}
		return 5, nil
	case b == 30:
		for i := 1; i < len(data); i++ {
			if data[i]>>4 == 0x0f || data[i]&0x0f == 0x0f {
				return i + 1, nil
			}
		}
		return 0, fmt.Errorf("cff: unterminated real")
	case b == 255:
		if len(data) < 5 {
			return 0, fmt.Errorf("cff: truncated fixed")
		}
		return 5, nil
	case b >= 32 && b <= 246:
		return 1, nil
	case b >= 247 && b <= 254:
		if len(data) < 2 {
			return 0, fmt.Errorf("cff: truncated two-byte number")
		}
		return 2, nil
	default:
		return 0, fmt.Errorf("cff: invalid number byte %d", b)
	}
}

type tableDir struct {
	entries    []*tableDirEntry
	entriesMap map[string]*tableDirEntry
}

func (dir *tableDir) read(file io.Reader, nTables uint16) (err error) {
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

func (entry *tableDirEntry) read(file io.Reader) (err error) {
	tag := make([]byte, 4)
	if _, err = file.Read(tag); err != nil {
		return
	}
	entry.tag = string(tag)
	err = readValues(file, &entry.checkSum, &entry.offset, &entry.length)
	return
}

func (font *Font) StrikeoutPosition() int {
	return int(font.os2Table.yStrikeoutPosition)
}

func (font *Font) StrikeoutThickness() int {
	return int(font.os2Table.yStrikeoutSize)
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
