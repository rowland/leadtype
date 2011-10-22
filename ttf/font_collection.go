package ttf

import (
	"fmt"
	"os"
	"path/filepath"
)

type FontCollection struct {
	fonts []*Font
}

func (fc *FontCollection) Add(pattern string) (err os.Error) {
	var pathnames []string
	if pathnames, err = filepath.Glob(pattern); err != nil {
		return
	}
	for _, pathname := range pathnames {
		font, err2 := LoadFont(pathname)
		if err2 != nil {
			err = os.NewError(fmt.Sprintf("Error loading %s: %s", pathname, err2))
			continue
		}
		fc.fonts = append(fc.fonts, font)
	}
	return
}

func (fc *FontCollection) Len() int {
	return len(fc.fonts)
}

func (fc *FontCollection) Select(family, weight, style string, options FontOptions) (*Font, os.Error) {
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
	for _, f := range fc.fonts {
		if f.Family() == family && f.Style() == ws {
			return f, nil
		}
	}
	return nil, os.NewError(fmt.Sprintf("Font %s %s not found", family, ws))
}

type FontOptions struct {
	
}