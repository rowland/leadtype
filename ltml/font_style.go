// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
	"strconv"
)

type FontStyle struct {
	id   string
	name string
	size float64

	color     Color
	strikeout bool
	style     string
	underline bool
	weight    string

	lineHeight float64
}

func (fs *FontStyle) Apply(w Writer) {
	fmt.Printf("Applying %s\n", fs)
}

func (fs *FontStyle) ID() string {
	return fs.id
}

const defaultFontSize = 12

func (fs *FontStyle) SetAttrs(attrs map[string]string) {
	if id, ok := attrs["id"]; ok {
		fs.id = id
	}
	if name, ok := attrs["name"]; ok {
		fs.name = name
	}
	if size, ok := attrs["size"]; ok {
		var err error
		if fs.size, err = strconv.ParseFloat(size, 64); err != nil {
			fs.size = defaultFontSize
		}
	}
	fs.color.SetAttrs(attrs)
	if strikeout, ok := attrs["strikeout"]; ok {
		fs.strikeout = (strikeout == "true")
	}
	if style, ok := attrs["style"]; ok {
		fs.style = style
	}
	if underline, ok := attrs["underline"]; ok {
		fs.underline = (underline == "true")
	}
	if weight, ok := attrs["weight"]; ok {
		fs.weight = weight
	}
	if lineHeight, ok := attrs["line-height"]; ok {
		fs.lineHeight, _ = strconv.ParseFloat(lineHeight, 64)
	}
}

func (fs *FontStyle) String() string {
	return fmt.Sprintf("FontStyle id=%s name=%s size=%f color=%s strikeout=%t style=%s underline=%t weight=%s line-height=%f",
		fs.id, fs.name, fs.size, fs.color, fs.strikeout, fs.style, fs.underline, fs.weight, fs.lineHeight)
}

func FontStyleFor(id string, scope HasScope) *FontStyle {
	style, ok := scope.Style(id)
	if !ok {
		style, ok = scope.Style("font_" + id)
	}
	if ok {
		fs, _ := style.(*FontStyle)
		return fs
	}
	return nil
}

func init() {
	registerTag(DefaultSpace, "font", func() interface{} { return &FontStyle{} })
}
