package afm

import (
	"bufio"
	"os"
	"regexp"
)

type Font struct {
	FontInfo
}

func LoadFont(filename string) (font *Font, err os.Error) {
	var file *os.File
	if file, err = os.Open(filename); err != nil {
		return
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	font = new(Font)
	err = font.init(reader)
	return
}

var (
	reEndCharMetrics = regexp.MustCompile("^EndCharMetrics")
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

func (font *Font) NumGlyphs() int {
	return font.numGlyphs
}

func (font *Font) UnitsPerEm() int {
	return 1000
}

type CharMetric struct {
	
}
