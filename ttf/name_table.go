package ttf

import (
	"bufio"
	"bytes"
	"encoding/binary"
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
	DefaultPlatformSpecificID      = 0
	Unicode2PlatformSpecificID     = 3
	Unicode2FullPlatformSpecificID = 4
)

// Microsoft Platform Specific ID's http://www.microsoft.com/typography/otspec/name.htm
const (
	UCS2PlatformSpecificID = 1
)

const (
	copyrightNoticeID       = 0
	fontFamilyID            = 1
	fontSubfamilyID         = 2
	uniqueSubfamilyID       = 3
	fullNameID              = 4
	versionID               = 5
	postScriptNameID        = 6
	trademarkNoticeID       = 7
	manufacturerNameID      = 8
	designerID              = 9
	descriptionID           = 10
	urlFontVendorID         = 11
	urlFontDesignerID       = 12
	licenseDescriptionID    = 13
	urlLicenseInformationID = 14
	preferredFamilyID       = 16
	preferredSubfamilyID    = 17
	compatibleFullID        = 18
	sampleTextID            = 19
)

type nameTable struct {
	format                uint16
	count                 uint16
	stringOffset          uint16
	nameRecords           []nameRecord
	copyrightNotice       string
	fontFamily            string
	fontSubfamily         string
	uniqueSubfamily       string
	fullName              string
	version               string
	postScriptName        string
	trademarkNotice       string
	manufacturerName      string
	designer              string
	description           string
	urlFontVendor         string
	urlFontDesigner       string
	licenseDescription    string
	urlLicenseInformation string
	preferredFamily       string
	preferredSubfamily    string
	compatibleFull        string
	sampleText            string
}

func (table *nameTable) init(rs io.ReadSeeker, entry *tableDirEntry) (err error) {
	if _, err = rs.Seek(int64(entry.offset), os.SEEK_SET); err != nil {
		return
	}
	file, _ := bufio.NewReaderSize(rs, int(entry.length))
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
	names := make([]byte, int(entry.length)-int(table.stringOffset))
	if _, err = file.Read(names); err != nil {
		return
	}
	for i := uint16(0); i < table.count; i++ {
		rec := &table.nameRecords[i]
		buf := bytes.NewBuffer(names[rec.offset : rec.offset+rec.length])
		var s string
		if s, err = table.readField(rec, buf); err != nil {
			return
		}
		table.setField(rec.nameID, s, rec.platformID != MacintoshPlatformID && (rec.languageID == 0 || rec.languageID == 1033))
	}
	return
}

func (table *nameTable) getField(nameID uint16) string {
	switch nameID {
	case copyrightNoticeID:
		return table.copyrightNotice
	case fontFamilyID:
		return table.fontFamily
	case fontSubfamilyID:
		return table.fontSubfamily
	case uniqueSubfamilyID:
		return table.uniqueSubfamily
	case fullNameID:
		return table.fullName
	case versionID:
		return table.version
	case postScriptNameID:
		return table.postScriptName
	case trademarkNoticeID:
		return table.trademarkNotice
	case manufacturerNameID:
		return table.manufacturerName
	case designerID:
		return table.designer
	case descriptionID:
		return table.description
	case urlFontVendorID:
		return table.urlFontVendor
	case urlFontDesignerID:
		return table.urlFontDesigner
	case licenseDescriptionID:
		return table.licenseDescription
	case urlLicenseInformationID:
		return table.urlLicenseInformation
	case preferredFamilyID:
		return table.preferredFamily
	case preferredSubfamilyID:
		return table.preferredSubfamily
	case compatibleFullID:
		return table.compatibleFull
	case sampleTextID:
		return table.sampleText
	}
	return ""
}

func (table *nameTable) readField(rec *nameRecord, file io.Reader) (s string, err error) {
	switch rec.platformID {
	case UnicodePlatformID, MicrosoftPlatformID:
		u := make([]uint16, rec.length/2)
		if err = binary.Read(file, binary.BigEndian, u); err != nil {
			return
		}
		s = utf16ToString(u)
	case MacintoshPlatformID:
		b := make([]byte, rec.length)
		if _, err = file.Read(b); err != nil {
			return
		}
		s = string(b)
	}
	return
}

func (table *nameTable) setField(nameID uint16, s string, overwrite bool) {
	if !overwrite && table.getField(nameID) != "" {
		return
	}
	switch nameID {
	case copyrightNoticeID:
		table.copyrightNotice = s
	case fontFamilyID:
		table.fontFamily = s
	case fontSubfamilyID:
		table.fontSubfamily = s
	case uniqueSubfamilyID:
		table.uniqueSubfamily = s
	case fullNameID:
		table.fullName = s
	case versionID:
		table.version = s
	case postScriptNameID:
		table.postScriptName = s
	case trademarkNoticeID:
		table.trademarkNotice = s
	case manufacturerNameID:
		table.manufacturerName = s
	case designerID:
		table.designer = s
	case descriptionID:
		table.description = s
	case urlFontVendorID:
		table.urlFontVendor = s
	case urlFontDesignerID:
		table.urlFontDesigner = s
	case licenseDescriptionID:
		table.licenseDescription = s
	case urlLicenseInformationID:
		table.urlLicenseInformation = s
	case preferredFamilyID:
		table.preferredFamily = s
	case preferredSubfamilyID:
		table.preferredSubfamily = s
	case compatibleFullID:
		table.compatibleFull = s
	case sampleTextID:
		table.sampleText = s
	}
}

func (table *nameTable) write(wr io.Writer) {
	fmt.Fprintln(wr, "----------")
	fmt.Fprintln(wr, "name Table")
	fmt.Fprintf(wr, "format = %d\n", table.format)
	fmt.Fprintf(wr, "count = %d\n", table.count)
	fmt.Fprintf(wr, "stringOffset = %d\n", table.stringOffset)
	fmt.Fprintln(wr, "platformID\tplatformSpecificID\tlanguageID\tnameID\tlength\toffset")
	for i := 0; i < len(table.nameRecords); i++ {
		table.nameRecords[i].write(wr)
	}
	fmt.Fprintf(wr, "copyrightNotice       = %s\n", table.copyrightNotice)
	fmt.Fprintf(wr, "fontFamily            = %s\n", table.fontFamily)
	fmt.Fprintf(wr, "fontSubfamily         = %s\n", table.fontSubfamily)
	fmt.Fprintf(wr, "uniqueSubfamily       = %s\n", table.uniqueSubfamily)
	fmt.Fprintf(wr, "fullName              = %s\n", table.fullName)
	fmt.Fprintf(wr, "version               = %s\n", table.version)
	fmt.Fprintf(wr, "postScriptName        = %s\n", table.postScriptName)
	fmt.Fprintf(wr, "trademarkNotice       = %s\n", table.trademarkNotice)
	fmt.Fprintf(wr, "manufacturerName      = %s\n", table.manufacturerName)
	fmt.Fprintf(wr, "designer              = %s\n", table.designer)
	fmt.Fprintf(wr, "description           = %s\n", table.description)
	fmt.Fprintf(wr, "urlFontVendor         = %s\n", table.urlFontVendor)
	fmt.Fprintf(wr, "urlFontDesigner       = %s\n", table.urlFontDesigner)
	fmt.Fprintf(wr, "licenseDescription    = %s\n", table.licenseDescription)
	fmt.Fprintf(wr, "urlLicenseInformation = %s\n", table.urlLicenseInformation)
	fmt.Fprintf(wr, "preferredFamily       = %s\n", table.preferredFamily)
	fmt.Fprintf(wr, "preferredSubfamily    = %s\n", table.preferredSubfamily)
	fmt.Fprintf(wr, "compatibleFull        = %s\n", table.compatibleFull)
	fmt.Fprintf(wr, "sampleText            = %s\n", table.sampleText)
}

type nameRecord struct {
	platformID         uint16
	platformSpecificID uint16
	languageID         uint16
	nameID             uint16
	length             uint16
	offset             uint16
}

func (rec *nameRecord) read(file io.Reader) error {
	return readValues(file,
		&rec.platformID,
		&rec.platformSpecificID,
		&rec.languageID,
		&rec.nameID,
		&rec.length,
		&rec.offset,
	)
}

func (rec *nameRecord) write(wr io.Writer) {
	fmt.Fprintf(wr, "%d\t%d\t%d\t%d\t%d\t%d\n",
		rec.platformID,
		rec.platformSpecificID,
		rec.languageID,
		rec.nameID,
		rec.length,
		rec.offset)
}
