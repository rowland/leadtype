// Copyright 2011-2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ttf_fonts

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/rowland/leadtype/font"
	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/ttf"
)

type TtfFonts struct {
	FontInfos []*ttf.FontInfo
	fonts     map[string]*ttf.Font
}

func New(pattern string) (*TtfFonts, error) {
	var fc TtfFonts
	if err := fc.Add(pattern); err != nil {
		return nil, err
	}
	return &fc, nil
}

// NewFromSystemFonts creates a TtfFonts collection loaded from all standard
// font directories for the current platform.
func NewFromSystemFonts() (*TtfFonts, error) {
	var fc TtfFonts
	if err := fc.AddSystemFonts(); err != nil {
		return nil, err
	}
	return &fc, nil
}

// AddSystemFonts adds all TTF and TTC fonts found in the platform's standard
// font directories.
func (fc *TtfFonts) AddSystemFonts() error {
	for _, dir := range SystemFontDirs() {
		for _, ext := range []string{"*.ttf", "*.TTF", "*.ttc", "*.TTC"} {
			// Ignore errors from individual patterns (directory may not exist).
			fc.Add(filepath.Join(dir, ext))
		}
	}
	return nil
}

func Families(families ...string) (fonts []*font.Font) {
	fc, err := NewFromSystemFonts()
	if err != nil {
		panic(err)
	}
	for _, family := range families {
		if family == "" {
			fonts = append(fonts, nil)
			continue
		}
		f, err := font.New(family, options.Options{}, font.FontSources{fc})
		if err != nil {
			panic(err)
		}
		fonts = append(fonts, f)
	}
	return
}

func (fc *TtfFonts) Add(pattern string) (err error) {
	var pathnames []string
	if pathnames, err = filepath.Glob(pattern); err != nil {
		return
	}
	for _, pathname := range pathnames {
		if strings.EqualFold(filepath.Ext(pathname), ".ttc") {
			infos, err2 := ttf.LoadFontInfosFromTTC(pathname)
			if err2 != nil {
				err = fmt.Errorf("Error loading %s: %s", pathname, err2)
				continue
			}
			fc.FontInfos = append(fc.FontInfos, infos...)
		} else {
			fi, err2 := ttf.LoadFontInfo(pathname)
			if err2 != nil {
				err = fmt.Errorf("Error loading %s: %s", pathname, err2)
				continue
			}
			fc.FontInfos = append(fc.FontInfos, fi)
		}
	}
	return
}

func (fc *TtfFonts) Len() int {
	return len(fc.FontInfos)
}

func (fc *TtfFonts) Select(family, weight, style string, ranges []string) (fontMetrics font.FontMetrics, err error) {
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
	for _, f := range fc.FontInfos {
		if strings.EqualFold(f.Family(), family) && strings.EqualFold(f.Style(), ws) {
			for _, r := range ranges {
				cpr, ok := ttf.CodepointRangesByName[r]
				if !ok || !f.CharRanges().IsSet(int(cpr.Bit)) {
					continue search
				}
			}
			cacheKey := fmt.Sprintf("%s@%d", f.Filename(), f.TTCOffset())
			font := fc.fonts[cacheKey]
			if font == nil {
				font, err = ttf.LoadFontAtOffset(f.Filename(), f.TTCOffset())
				fc.fonts[cacheKey] = font
			}
			fontMetrics = font
			return
		}
	}
	err = fmt.Errorf("Font %s %s not found", family, ws)
	return
}

func (fc *TtfFonts) SubType() string {
	return "TrueType"
}
