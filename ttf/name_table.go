package ttf

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

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
