package afm

import (
	"bufio"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type FontInfo struct {
	filename    string
	ascender    int
	capHeight   int
	copyright   string
	descender   int
	familyName  string
	fontBBox    BoundingBox
	fontName    string
	fullName    string
	italicAngle float64
	numGlyphs   int
	stdVW       int
	version     string
	weight      string
	xHeight     int
}

// 323,331 ns
func LoadFontInfo(filename string) (fi *FontInfo, err os.Error) {
	var file *os.File
	if file, err = os.Open(filename); err != nil {
		return
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	fi = new(FontInfo)
	fi.filename = filename
	err = fi.init(reader)
	return
}

var (
	reAscender         = regexp.MustCompile("^Ascender[ ]+([0-9]+)")
	reCapHeight        = regexp.MustCompile("^CapHeight[ ]+([0-9]+)")
	reCopyright        = regexp.MustCompile("^Notice Copyright[ ]+(.*)")
	reDescender        = regexp.MustCompile("^Descender[ ]+(-?[0-9]+)")
	reFamilyName       = regexp.MustCompile("^FamilyName[ ]+(.*)")
	reFontBBox         = regexp.MustCompile("^FontBBox(([ ]+-?[0-9]+)([ ]+-?[0-9]+)([ ]+-?[0-9]+)([ ]+-?[0-9]+))")
	reFontName         = regexp.MustCompile("^FontName[ ]+(.*)")
	reFullName         = regexp.MustCompile("^FullName[ ]+(.*)")
	reItalicAngle      = regexp.MustCompile("^ItalicAngle[ ]+(-?[0-9]+(\\.[0-9]+)?)")
	reStartCharMetrics = regexp.MustCompile("^StartCharMetrics[ ]+([0-9]+)")
	reStdVW            = regexp.MustCompile("^StdVW[ ]+([0-9]+)")
	reVersion          = regexp.MustCompile("^Version[ ]+(.*)")
	reWeight           = regexp.MustCompile("^Weight[ ]+([A-Za-z0-9]+)")
	reXHeight          = regexp.MustCompile("^XHeight[ ]+([0-9]+)")
)

func (fi *FontInfo) init(file *bufio.Reader) (err os.Error) {
	// Symbolic fonts don't specify ascender and descender, so default to reasonable numbers.
        fi.ascender = 750
        fi.descender = -188

	var line []byte
	line, err = file.ReadSlice('\n')
	for err == nil {
		if m := reStartCharMetrics.FindSubmatch(line); m != nil {
			fi.numGlyphs, _ = strconv.Atoi(string(m[1]))
			break
		} else if m := reFontName.FindSubmatch(line); m != nil {
			fi.fontName = strings.TrimSpace(string(m[1]))
		} else if m := reAscender.FindSubmatch(line); m != nil {
			fi.ascender, _ = strconv.Atoi(string(m[1]))
		} else if m := reFontBBox.FindSubmatch(line); m != nil {
			sa := strings.Split(strings.TrimSpace(string(m[1])), " ")
			fi.fontBBox.XMin, _ = strconv.Atoi(sa[0])
			fi.fontBBox.YMin, _ = strconv.Atoi(sa[1])
			fi.fontBBox.XMax, _ = strconv.Atoi(sa[2])
			fi.fontBBox.YMax, _ = strconv.Atoi(sa[3])
		} else if m := reDescender.FindSubmatch(line); m != nil {
			fi.descender, _ = strconv.Atoi(string(m[1]))
		} else if m := reFamilyName.FindSubmatch(line); m != nil {
			fi.familyName = strings.TrimSpace(string(m[1]))
		} else if m := reItalicAngle.FindSubmatch(line); m != nil {
			fi.italicAngle, _ = strconv.Atof64(string(m[1]))
		} else if m := reWeight.FindSubmatch(line); m != nil {
			fi.weight = strings.TrimSpace(string(m[1]))
		} else if m := reCapHeight.FindSubmatch(line); m != nil {
			fi.capHeight, _ = strconv.Atoi(string(m[1]))
		} else if m := reCopyright.FindSubmatch(line); m != nil {
			fi.copyright = strings.TrimSpace(string(m[1]))
		} else if m := reFullName.FindSubmatch(line); m != nil {
			fi.fullName = strings.TrimSpace(string(m[1]))
		} else if m := reStdVW.FindSubmatch(line); m != nil {
			fi.stdVW, _ = strconv.Atoi(string(m[1]))
		} else if m := reVersion.FindSubmatch(line); m != nil {
			fi.version = strings.TrimSpace(string(m[1]))
		} else if m := reXHeight.FindSubmatch(line); m != nil {
			fi.xHeight, _ = strconv.Atoi(string(m[1]))
		}
		line, err = file.ReadSlice('\n')
	}
	return
}

func (fi *FontInfo) Ascent() int {
	return fi.ascender
}

func (fi *FontInfo) BoundingBox() BoundingBox {
	return fi.fontBBox
}

func (fi *FontInfo) CapHeight() int {
	return fi.capHeight
}

func (fi *FontInfo) Copyright() string {
	return fi.copyright
}

func (fi *FontInfo) Descent() int {
	return fi.descender
}

func (fi *FontInfo) Family() string {
	return fi.familyName
}

func (fi *FontInfo) FullName() string {
	return fi.fullName
}

func (fi *FontInfo) ItalicAngle() float64 {
	return fi.italicAngle
}

func (fi *FontInfo) Leading() int {
	return int(float64(fi.ascender - fi.descender) * 1.15)
}

func (fi *FontInfo) PostScriptName() string {
	return fi.fontName
}

func (fi *FontInfo) StemV() int {
	return fi.stdVW
}

func (fi *FontInfo) Style() string {
	return fi.weight
}

func (fi *FontInfo) Version() string {
	return fi.version
}

func (fi *FontInfo) XHeight() int {
	return fi.xHeight
}

type BoundingBox struct {
	XMin, YMin, XMax, YMax int
}