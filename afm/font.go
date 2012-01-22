package afm

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

type Font struct {
	FontInfo
	serif bool
}

func LoadFont(filename string) (font *Font, err os.Error) {
	var file *os.File
	if file, err = os.Open(filename); err != nil {
		return
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	font = new(Font)
	if err = font.init(reader); err != nil {
		return
	}
	inf := reAfmExt.ReplaceAllString(filename, ".inf")
	err = font.initInf(inf)
	return
}

var (
	reEndCharMetrics = regexp.MustCompile("^EndCharMetrics")
	reSerif = regexp.MustCompile("^Serif[ ]+([A-Za-z]+)")
	reAfmExt = regexp.MustCompile("\\.afm$")
)

func (font *Font) init(file *bufio.Reader) (err os.Error) {
	if err = font.FontInfo.init(file); err != nil {
		return
	}
	
	var line []byte
	line, err = file.ReadSlice('\n')
	for err == nil {
		if m := reEndCharMetrics.Find(line); m != nil {
			break;
		}
		line, err = file.ReadSlice('\n')
	}
	return
}

func (font *Font) initInf(filename string) (err os.Error) {
	var file *os.File
	if file, err = os.Open(filename); err != nil {
		return
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	var line []byte
	line, err = reader.ReadSlice('\n')
	for err == nil {
		if m := reSerif.FindSubmatch(line); m != nil {
			font.serif = (strings.TrimSpace(string(m[1])) == "true")
		}
		line, err = reader.ReadSlice('\n')
	}
	if err == os.EOF {
		return nil
	}
	return
}

func (font *Font) NumGlyphs() int {
	return font.numGlyphs
}

func (font *Font) Serif() bool {
	return font.serif
}

func (font *Font) UnitsPerEm() int {
	return 1000
}

type CharMetric struct {
	
}
