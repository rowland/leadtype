package ttf

import (
	"errors"
	"fmt"
	"path/filepath"
)

type FontCollection struct {
	fontInfos []*FontInfo
	fonts     map[string]*Font
}

// 81,980,000 ns
func (fc *FontCollection) Add(pattern string) (err error) {
	var pathnames []string
	if pathnames, err = filepath.Glob(pattern); err != nil {
		return
	}
	for _, pathname := range pathnames {
		fi, err2 := LoadFontInfo(pathname)
		if err2 != nil {
			err = errors.New(fmt.Sprintf("Error loading %s: %s", pathname, err2))
			continue
		}
		fc.fontInfos = append(fc.fontInfos, fi)
	}
	return
}

func (fc *FontCollection) Len() int {
	return len(fc.fontInfos)
}

func (fc *FontCollection) Select(family, weight, style string, options FontOptions) (font *Font, err error) {
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
		fc.fonts = make(map[string]*Font)
	}
search:
	for _, f := range fc.fontInfos {
		if f.Family() == family && f.Style() == ws {
			for _, r := range options.CharRanges {
				if !f.CharRanges().IsSet(r) {
					continue search
				}
			}
			font = fc.fonts[f.filename]
			if font == nil {
				font, err = LoadFont(f.filename)
				fc.fonts[f.filename] = font
			}
			return
		}
	}
	err = errors.New(fmt.Sprintf("Font %s %s not found", family, ws))
	return
}

type FontOptions struct {
	CharRanges []int
}
