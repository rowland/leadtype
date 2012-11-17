// Copyright 2011-2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import (
	"fmt"
	"leadtype/ttf"
	"path/filepath"
)

type TtfFontCollection struct {
	fontInfos []*ttf.FontInfo
	fonts     map[string]*ttf.Font
}

func NewTtfFontCollection(pattern string) (*TtfFontCollection, error) {
	var fc TtfFontCollection
	if err := fc.Add(pattern); err != nil {
		return nil, err
	}
	return &fc, nil
}

// 81,980,000 ns
func (fc *TtfFontCollection) Add(pattern string) (err error) {
	var pathnames []string
	if pathnames, err = filepath.Glob(pattern); err != nil {
		return
	}
	for _, pathname := range pathnames {
		fi, err2 := ttf.LoadFontInfo(pathname)
		if err2 != nil {
			err = fmt.Errorf("Error loading %s: %s", pathname, err2)
			continue
		}
		fc.fontInfos = append(fc.fontInfos, fi)
	}
	return
}

func (fc *TtfFontCollection) Len() int {
	return len(fc.fontInfos)
}

func (fc *TtfFontCollection) Select(family, weight, style string, ranges []string) (fontMetrics FontMetrics, err error) {
	var ws string
	if weight != "" && style != "" {
		ws = weight + " " + style
	} else if weight == "" && style == "" {
		ws = "Regular"
	} else if style == "" {
		ws = weight
	} else if weight == "" {
		ws = style
	}
	if fc.fonts == nil {
		fc.fonts = make(map[string]*ttf.Font)
	}
search:
	for _, f := range fc.fontInfos {
		if f.Family() == family && f.Style() == ws {
			for _, r := range ranges {
				cpr, ok := ttf.CodepointRangesByName[r]
				if !ok || !f.CharRanges().IsSet(int(cpr.Bit)) {
					continue search
				}
			}
			font := fc.fonts[f.Filename]
			if font == nil {
				font, err = ttf.LoadFont(f.Filename)
				fc.fonts[f.Filename] = font
			}
			fontMetrics = font
			return
		}
	}
	err = fmt.Errorf("Font %s %s not found", family, ws)
	return
}
