// Copyright 2011-2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ttf

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

type postTable struct {
	format             Fixed
	italicAngle        Fixed
	underlinePosition  int16
	underlineThickness int16
	isFixedPitch       uint32
	minMemType42       uint32
	maxMemType42       uint32
	minMemType1        uint32
	maxMemType1        uint32
	names              []string
	charCodes          []uint16
}

func (table *postTable) init(rs io.ReadSeeker, entry *tableDirEntry, numGlyphs uint16) (err error) {
	if _, err = rs.Seek(int64(entry.offset), os.SEEK_SET); err != nil {
		return
	}
	file := bufio.NewReaderSize(rs, int(entry.length))
	if err = table.format.Read(file); err != nil {
		return
	}
	if err = table.italicAngle.Read(file); err != nil {
		return
	}
	if err = readValues(file,
		&table.underlinePosition,
		&table.underlineThickness,
		&table.isFixedPitch,
		&table.minMemType42,
		&table.maxMemType42,
		&table.minMemType1,
		&table.maxMemType1); err != nil {
		return
	}

	switch table.format.Tof64() {
	case 1.0:
		table.names = MacRomanPostNames
	case 2.0:
		err = table.readFormat2Names(file)
	case 3.0:
		// no subtable for format 3
	case 4.0:
		table.charCodes = make([]uint16, numGlyphs)
		err = readValues(file, &table.charCodes)
	default:
		return fmt.Errorf("Unsupported post table format: %g", table.format.Tof64())
	}

	return
}

func (table *postTable) readFormat2Names(file io.Reader) (err error) {
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
	fmt.Fprintf(wr, "format = %g\n", table.format.Tof64())
	fmt.Fprintf(wr, "italicAngle = %g\n", table.italicAngle.Tof64())
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
