package ttf

import (
	"io"
	"os"
)

type FontInfo struct {
	filename      string
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
func LoadFontInfo(filename string) (fi *FontInfo, err os.Error) {
	var file *os.File
	if file, err = os.Open(filename); err != nil {
		return
	}
	defer file.Close()
	fi = new(FontInfo)
	fi.filename = filename
	err = fi.init(file)
	return
}

func (fi *FontInfo) init(file io.ReadSeeker) (err os.Error) {
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

func (fi *FontInfo) Family() string {
	return fi.nameTable.fontFamily
}

func (fi *FontInfo) Style() string {
	return fi.nameTable.fontSubfamily
}
