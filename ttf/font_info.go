// Copyright 2011-2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ttf

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"os"
)

type FontInfo struct {
	Filename      string
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
	var file *os.File
	if file, err = os.Open(filename); err != nil {
		return
	}
	defer file.Close()
	fi = new(FontInfo)
	fi.Filename = filename
	err = fi.init(file)
	return
}

func (fi *FontInfo) init(file io.ReadSeeker) (err error) {
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
