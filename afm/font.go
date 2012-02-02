package afm

import (
	"bufio"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type Font struct {
	FontInfo
	serif            bool
	charMetrics      []CharMetric
	charsByGlyphName map[string]*CharMetric
	charsByCodepoint map[int]*CharMetric
}

// 3,956,700 ns/3.957 ms
func LoadFont(filename string) (font *Font, err error) {
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
	reSerif          = regexp.MustCompile("^Serif[ ]+([A-Za-z]+)")
	reAfmExt         = regexp.MustCompile("\\.afm$")
	reCharMetrics    = regexp.MustCompile("^C[ ]+(-?[0-9]+)[ ]*;[ ]*WX[ ]+([0-9]+)[ ]*;[ ]*N[ ]+([A-Za-z0-9]+)")
)

func (font *Font) init(file *bufio.Reader) (err error) {
	if err = font.FontInfo.init(file); err != nil {
		return
	}
	font.charMetrics = make([]CharMetric, font.numGlyphs)
	font.charsByGlyphName = make(map[string]*CharMetric, font.numGlyphs)
	font.charsByCodepoint = make(map[int]*CharMetric, font.numGlyphs)
	var line []byte
	line, err = file.ReadSlice('\n')
	for i := 0; err == nil; i += 1 {
		if m := reEndCharMetrics.Find(line); m != nil {
			break
		}
		cm := &font.charMetrics[i]
		if m := reCharMetrics.FindSubmatch(line); m != nil {
			cm.code, _ = strconv.Atoi(string(m[1]))
			cm.width, _ = strconv.Atoi(string(m[2]))
			cm.name = string(m[3])
		} else {
			fields := strings.Split(string(line), ";")
			for _, f := range fields {
				f = strings.TrimSpace(f)
				kv := strings.Split(f, " ")
				switch kv[0] {
				case "C":
					cm.code, _ = strconv.Atoi(kv[1])
				case "CH":
					n, _ := strconv.ParseUint(kv[1][1:len(kv[1])-1], 16, 64)
					cm.code = int(n)
				case "WX", "W0X":
					cm.width, _ = strconv.Atoi(kv[1])
				case "N":
					cm.name = kv[1]
				}
			}
		}
		cp := GlyphCodepoints[cm.name]
		font.charsByGlyphName[cm.name] = cm
		font.charsByCodepoint[cp] = cm
		line, err = file.ReadSlice('\n')
	}
	return
}

func (font *Font) initInf(filename string) (err error) {
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
	if err == io.EOF {
		return nil
	}
	return
}

func (font *Font) AdvanceWidth(codepoint int) int {
	if cm := font.charsByCodepoint[codepoint]; cm != nil {
		return cm.width
	}
	return 0
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
	code  int
	width int
	name  string
}
