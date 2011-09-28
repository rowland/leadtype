package ttf

import (
	"fmt"
	"io"
	"os"
)

// Platform ID's http://developer.apple.com/fonts/TTRefMan/RM06/Chap6name.html
const (
	UnicodePlatformID   = 0
	MacintoshPlatformID = 1
	MicrosoftPlatformID = 3
)

// Unicode Platform Specific ID's http://developer.apple.com/fonts/TTRefMan/RM06/Chap6name.html
const (
	DefaultPlatformSpecificID  = 0
	Unicode2PlatformSpecificID = 3
)

// Microsoft Platform Specific ID's http://www.microsoft.com/typography/otspec/name.htm
const (
	UCS2PlatformSpecificID = 1
)

type nameTable struct {
	format       uint16
	count        uint16
	stringOffset uint16
	nameRecords  []nameRecord
}

func (table *nameTable) init(file io.ReadSeeker, entry *tableDirEntry) (err os.Error) {
	if _, err = file.Seek(int64(entry.offset), os.SEEK_SET); err != nil {
		return
	}
	if err = readValues(file,
		&table.format,
		&table.count,
		&table.stringOffset); err != nil {
		return
	}
	table.nameRecords = make([]nameRecord, table.count)
	for i := uint16(0); i < table.count; i++ {
		if err = table.nameRecords[i].read(file); err != nil {
			return
		}
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

func (rec *nameRecord) read(file io.Reader) os.Error {
	return readValues(file,
		&rec.platformID,
		&rec.platformSpecificID,
		&rec.languageID,
		&rec.nameID,
		&rec.length,
		&rec.offset,
	)
}
