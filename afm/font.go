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
	charsByCodepoint map[rune]*CharMetric
}

// 3,956,700 ns/3.957 ms
// 2,446,984 ns/2.447 ms
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
	font.charsByCodepoint = make(map[rune]*CharMetric, font.numGlyphs)
	var line []byte
	line, err = file.ReadSlice('\n')
	for i := 0; err == nil; i += 1 {
		if m := reEndCharMetrics.Find(line); m != nil {
			break
		}
		cm := &font.charMetrics[i]
		if m := reCharMetrics.FindSubmatch(line); m != nil {
			r, _ := strconv.Atoi(string(m[1]))
			cm.code = rune(r)
			w, _ := strconv.Atoi(string(m[2]))
			cm.width = int32(w)
			cm.name = string(m[3])
		} else {
			fields := strings.Split(string(line), ";")
			for _, f := range fields {
				f = strings.TrimSpace(f)
				kv := strings.Split(f, " ")
				switch kv[0] {
				case "C":
					r, _ := strconv.Atoi(kv[1])
					cm.code = rune(r)
				case "CH":
					n, _ := strconv.ParseUint(kv[1][1:len(kv[1])-1], 16, 64)
					cm.code = rune(n)
				case "WX", "W0X":
					w, _ := strconv.Atoi(kv[1])
					cm.width = int32(w)
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

// 58.4 ns
// 71.3 ns
func (font *Font) AdvanceWidth(codepoint rune) int {
	if cm := font.charsByCodepoint[codepoint]; cm != nil {
		return int(cm.width)
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
	code  rune
	width int32
	name  string
}
