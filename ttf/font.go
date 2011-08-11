package ttf

import (
	"encoding/binary"
	"os"
)

type Font struct {
	scalar        uint32
	nTables       uint16
	searchRange   uint16
	entrySelector uint16
	rangeShift    uint16
}

func NewFont(filename string) (font *Font, err os.Error) {
	font = new(Font)
	var file *os.File
	file, err = os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()

	binary.Read(file, binary.BigEndian, &font.scalar)
	binary.Read(file, binary.BigEndian, &font.nTables)

	return
}
