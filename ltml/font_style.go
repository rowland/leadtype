// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
	"strconv"

	"github.com/rowland/leadtype/colors"
	"github.com/rowland/leadtype/options"
)

type FontStyle struct {
	id   string
	name string
	size float64

	color     colors.Color
	strikeout bool
	style     string
	underline bool
	weight    string

	lineHeight float64
}

func (fs *FontStyle) Apply(w Writer) {
	fmt.Printf("Applying %s\n", fs)
	w.SetFont(fs.name, fs.size, options.Options{
		"color":  colors.Color(fs.color),
		"weight": fs.weight,
		"style":  fs.style})
}

func (fs *FontStyle) Clone() *FontStyle {
	clone := *fs
	return &clone
}

func (fs *FontStyle) ID() string {
	return fs.id
}

const (
	defaultFontName = "Helvetica"
	defaultFontSize = 12
)

var defaultFont = &FontStyle{id: "default", name: defaultFontName, size: defaultFontSize}

func (fs *FontStyle) SetAttrs(prefix string, attrs map[string]string) {
	if id, ok := attrs[prefix+"id"]; ok {
		fs.id = id
	}
	if name, ok := attrs[prefix+"name"]; ok {
		fs.name = name
	}
	if size, ok := attrs[prefix+"size"]; ok {
		var err error
		if fs.size, err = strconv.ParseFloat(size, 64); err != nil {
			fs.size = defaultFontSize
		}
	}
	if color, ok := attrs[prefix+"color"]; ok {
		fs.color = NamedColor(color)
	}
	if strikeout, ok := attrs[prefix+"strikeout"]; ok {
		fs.strikeout = (strikeout == "true")
	}
	if style, ok := attrs[prefix+"style"]; ok {
		fs.style = style
	}
	if underline, ok := attrs[prefix+"underline"]; ok {
		fs.underline = (underline == "true")
	}
	if weight, ok := attrs[prefix+"weight"]; ok {
		fs.weight = weight
	}
	if lineHeight, ok := attrs[prefix+"line-height"]; ok {
		fs.lineHeight, _ = strconv.ParseFloat(lineHeight, 64)
	}
}

func (fs *FontStyle) String() string {
	return fmt.Sprintf("FontStyle id=%s name=%s size=%f color=%s strikeout=%t style=%s underline=%t weight=%s line-height=%f",
		fs.id, fs.name, fs.size, fs.color, fs.strikeout, fs.style, fs.underline, fs.weight, fs.lineHeight)
}

func FontStyleFor(id string, scope HasScope) *FontStyle {
	if style, ok := scope.StyleFor(id); ok {
		fs, _ := style.(*FontStyle)
		return fs
	}
	return nil
}

var _ HasAttrsPrefix = (*FontStyle)(nil)
var _ Styler = (*FontStyle)(nil)

func init() {
	registerTag(DefaultSpace, "font", func() interface{} { return &FontStyle{} })
}
