// Copyright 2011-2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package afm

import (
	"bufio"
	"io"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type Font struct {
	FontInfo
	serif       bool
	CharMetrics CharMetrics
}

// 3,956,700 ns/3.957 ms
// 2,446,984 ns/2.447 ms
// 2,284,165 ns/2.284 ms
// 2,294,880 ns/2.294 ms go1
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
	font.CharMetrics = NewCharMetrics(font.numGlyphs)
	var line []byte
	line, err = file.ReadSlice('\n')
	for i := 0; err == nil; i += 1 {
		if m := reEndCharMetrics.Find(line); m != nil {
			break
		}
		cm := &font.CharMetrics[i]
		if m := reCharMetrics.FindSubmatch(line); m != nil {
			w, _ := strconv.Atoi(string(m[2]))
			cm.Width = int32(w)
			cm.Name = string(m[3])
		} else {
			fields := strings.Split(string(line), ";")
			for _, f := range fields {
				f = strings.TrimSpace(f)
				kv := strings.Split(f, " ")
				switch kv[0] {
				// case "C":
				// 	r, _ := strconv.Atoi(kv[1])
				// 	cm.code = rune(r)
				// case "CH":
				// 	n, _ := strconv.ParseUint(kv[1][1:len(kv[1])-1], 16, 64)
				// 	cm.code = rune(n)
				case "WX", "W0X":
					w, _ := strconv.Atoi(kv[1])
					cm.Width = int32(w)
				case "N":
					cm.Name = kv[1]
				}
			}
		}
		cm.Code = GlyphCodepoints[cm.Name]
		line, err = file.ReadSlice('\n')
	}
	sort.Sort(font.CharMetrics)
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

// 42.5 ns go1 (returning bool err)
func (font *Font) AdvanceWidth(codepoint rune) (width int, err bool) {
	if cm := font.CharMetrics.ForRune(codepoint); cm != nil {
		return int(cm.Width), false
	}
	return 0, true
}

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
	flags = 0
	if font.isFixedPitch {
		flags |= 1 << flagFixedPitch
	}
	if font.serif {
		flags |= 1 << flagSerif
	}
	if font.italicAngle != 0 {
		flags |= 1 << flagItalic
	}
	if strings.Contains(font.weight, "Bold") {
		flags |= 1 << flagForceBold
	}
	return
	// TODO: Set remainder of flags
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
