// Read some of the tables specified in http://developer.apple.com/fonts/ttrefman/

package ttf

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

type Font struct {
	FontInfo
	cmapTable cmapTable
	headTable headTable
	hheaTable hheaTable
	hmtxTable hmtxTable
	maxpTable maxpTable
	postTable postTable
	vheaTable vheaTable
	vmtxTable vmtxTable
}

// 9,151,820 ns
func LoadFont(filename string) (font *Font, err error) {
	var file *os.File
	if file, err = os.Open(filename); err != nil {
		return
	}
	defer file.Close()
	font = new(Font)
	err = font.init(file)
	return
}

func (font *Font) init(file io.ReadSeeker) (err error) {
	if err = font.FontInfo.init(file); err != nil {
		return
	}

	if entry := font.tableDir.table("cmap"); entry != nil {
		if err = font.cmapTable.init(file, entry); err != nil {
			return
		}
	}
	if entry := font.tableDir.table("head"); entry != nil {
		if err = font.headTable.init(file, entry); err != nil {
			return
		}
	}
	if entry := font.tableDir.table("hhea"); entry != nil {
		if err = font.hheaTable.init(file, entry); err != nil {
			return
		}
	}
	if entry := font.tableDir.table("post"); entry != nil {
		if err = font.postTable.init(file, entry, font.maxpTable.numGlyphs); err != nil {
			return
		}
	}
	if entry := font.tableDir.table("maxp"); entry != nil {
		if err = font.maxpTable.init(file, entry); err != nil {
			return
		}
	}
	if entry := font.tableDir.table("hmtx"); entry != nil {
		if err = font.hmtxTable.init(file, entry, font.maxpTable.numGlyphs, font.hheaTable.numOfLongHorMetrics); err != nil {
			return
		}
	}
	if entry := font.tableDir.table("vhea"); entry != nil {
		if err = font.vheaTable.init(file, entry); err != nil {
			return
		}
	}
	if entry := font.tableDir.table("vmtx"); entry != nil {
		if err = font.vmtxTable.init(file, entry, font.maxpTable.numGlyphs, font.vheaTable.numOfLongVerMetrics); err != nil {
			return
		}
	}
	return
}

// 74.6 ns
func (font *Font) AdvanceWidth(codepoint int) (width int, err error) {
	if index := font.cmapTable.glyphIndex(codepoint); index < 0 {
		err = missingCodepoint(codepoint)
	} else {
		width = int(font.hmtxTable.lookupAdvanceWidth(index))
	}
	return
}

func (font *Font) Ascent() int {
	return int(font.hheaTable.ascent)
}

func (font *Font) BoundingBox() BoundingBox {
	return BoundingBox{int(font.headTable.xMin), int(font.headTable.yMin), int(font.headTable.xMax), int(font.headTable.yMax)}
}

func (font *Font) Descent() int {
	return int(font.hheaTable.descent)
}

var features = []string{"header", "dir", "name", "post", "cmap", "head", "hhea", "maxp", "hmtx", "vhea", "vmtx", "OS/2"}

func (font *Font) Dump(wr io.Writer, feature string) {
	switch feature {
	case "header":
		font.writeHeader(wr)
	case "dir":
		font.tableDir.write(wr)
	case "name":
		font.nameTable.write(wr)
	case "post":
		font.postTable.write(wr)
	case "cmap":
		font.cmapTable.write(wr)
	case "head":
		font.headTable.write(wr)
	case "hhea":
		font.hheaTable.write(wr)
	case "hmtx":
		font.hmtxTable.write(wr)
	case "maxp":
		font.maxpTable.write(wr)
	case "OS/2":
		font.os2Table.write(wr)
	case "vhea":
		font.vheaTable.write(wr)
	case "vmtx":
		font.vmtxTable.write(wr)
	default:
		for _, feature2 := range features {
			font.Dump(wr, feature2)
		}
	}
}

const RestrictedLicenseEmbedding = 0x002

const (
	flagFixedPitch  = 1 - 1
	flagSerif       = 2 - 1
	flagSymbolic    = 3 - 1
	flagScript      = 4 - 1
	flagNonsymbolic = 6 - 1
	flagItalic      = 7 - 1
	flagAllCap      = 17 - 1
	flagSmallCap    = 18 - 1
	flagForceBold   = 19 - 1
)

func (font *Font) Flags() (flags uint32) {
	flags = font.postTable.isFixedPitch << flagFixedPitch
	if font.ItalicAngle() != 0 {
		flags |= 1 << flagItalic
	}
	if font.os2Table.usWeightClass > 500 {
		flags |= 1 << flagForceBold
	}
	return
	// TODO: Set remainder of flags
}

func (font *Font) ItalicAngle() float64 {
	return font.postTable.italicAngle.Tof64()
}

func (font *Font) Leading() int {
	return int(font.hheaTable.ascent - font.hheaTable.descent + font.hheaTable.lineGap)
}

func (font *Font) MaxWidth() int {
	return int(font.hheaTable.advanceWidthMax)
}

/*
// Optional. Not found in TTF.
func (font *Font) MissingWidth() int {
	return 0
}
*/

func (font *Font) NumGlyphs() int {
	return int(font.maxpTable.numGlyphs)
}

/*
// Optional. Not found in TTF.
func (font *font) StemH() int {
	return 0
}
*/

func (font *Font) String() string {
	var buf bytes.Buffer
	font.Dump(&buf, "all")
	return buf.String()
}

func (font *Font) UnitsPerEm() int {
	return int(font.headTable.unitsPerEm)
}

func (font *Font) writeHeader(wr io.Writer) {
	fmt.Fprintf(wr, "scalar = %d\n", font.scalar)
	fmt.Fprintf(wr, "nTables = %d\n", font.nTables)
	fmt.Fprintf(wr, "searchRange = %d\n", font.searchRange)
	fmt.Fprintf(wr, "entrySelector = %d\n", font.entrySelector)
	fmt.Fprintf(wr, "rangeShift = %d\n", font.rangeShift)
}

type BoundingBox struct {
	XMin, YMin, XMax, YMax int
}

type missingCodepoint int

func (cp missingCodepoint) Error() string {
	return fmt.Sprintf("Font missing codepoint %d.", cp)
}
