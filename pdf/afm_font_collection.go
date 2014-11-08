// Copyright 2011-2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import (
	"errors"
	"fmt"
	"leadtype/afm"
	"path/filepath"
	"regexp"
	"strings"
)

type AfmFontCollection struct {
	FontInfos []*afm.FontInfo
	fonts     map[string]*afm.Font
}

func NewAfmFontCollection(pattern string) (*AfmFontCollection, error) {
	var fc AfmFontCollection
	if err := fc.Add(pattern); err != nil {
		return nil, err
	}
	return &fc, nil
}

// 81,980,000 ns
func (fc *AfmFontCollection) Add(pattern string) (err error) {
	var pathnames []string
	if pathnames, err = filepath.Glob(pattern); err != nil {
		return
	}
	for _, pathname := range pathnames {
		fi, err2 := afm.LoadFontInfo(pathname)
		if err2 != nil {
			err = fmt.Errorf("Error loading %s: %s", pathname, err2)
			continue
		}
		fc.FontInfos = append(fc.FontInfos, fi)
	}
	return
}

func (fc *AfmFontCollection) Len() int {
	return len(fc.FontInfos)
}

func makeFontSelectRegexp(family, weight, style string) (re *regexp.Regexp, err error) {
	family = regexp.QuoteMeta(family)
	family = strings.Replace(family, " ", `\-?`, -1)

	style = strings.ToLower(style)
	if style == "italic" || style == "oblique" {
		style = "(Italic|Obl(ique)?)"
	} else {
		style = ""
	}

	ws := regexp.QuoteMeta(weight) + style
	var reWS string
	if ws == "" {
		reWS = "(-(Roman|Medium))?"
	} else {
		reWS = "-" + ws
	}

	reS := "(?i)^" + family + reWS + "$"
	re, err = regexp.Compile(reS)
	return
}

func (fc *AfmFontCollection) Select(family, weight, style string, ranges []string) (fontMetrics FontMetrics, err error) {
	if len(ranges) > 0 {
		return nil, errors.New("Named ranges not supported for Type1 fonts.")
	}
	var re *regexp.Regexp
	if re, err = makeFontSelectRegexp(family, weight, style); err != nil {
		return
	}
	if fc.fonts == nil {
		fc.fonts = make(map[string]*afm.Font)
	}
	for _, f := range fc.FontInfos {
		if re.MatchString(f.PostScriptName()) {
			font := fc.fonts[f.Filename]
			if font == nil {
				font, err = afm.LoadFont(f.Filename)
				fc.fonts[f.Filename] = font
			}
			fontMetrics = font
			return
		}
	}
	err = fmt.Errorf("Font '%s %s %s' not found", family, weight, style)
	return
}
