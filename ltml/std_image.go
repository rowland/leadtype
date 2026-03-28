// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
)

type StdImage struct {
	StdWidget
	src string
}

func (img *StdImage) DrawContent(w Writer) error {
	if img.src == "" {
		return fmt.Errorf("image src must be specified")
	}
	_, _, err := w.PrintImageFile(img.src, ContentLeft(img), ContentTop(img), img.widthForWriter(), img.heightForWriter())
	return err
}

func (img *StdImage) PreferredHeight(w Writer) float64 {
	if img.height != 0 {
		return img.height
	}
	infoWidth, infoHeight, err := img.imageDimensions(w)
	if err != nil || infoWidth == 0 {
		return NonContentHeight(img)
	}
	if img.width != 0 {
		return img.width*float64(infoHeight)/float64(infoWidth) + NonContentHeight(img)
	}
	return float64(infoHeight) + NonContentHeight(img)
}

func (img *StdImage) PreferredWidth(w Writer) float64 {
	if img.width != 0 {
		return img.width
	}
	infoWidth, infoHeight, err := img.imageDimensions(w)
	if err != nil || infoHeight == 0 {
		return NonContentWidth(img)
	}
	if img.height != 0 {
		return img.height*float64(infoWidth)/float64(infoHeight) + NonContentWidth(img)
	}
	return float64(infoWidth) + NonContentWidth(img)
}

func (img *StdImage) imageDimensions(w Writer) (width, height int, err error) {
	return w.ImageDimensionsFromFile(img.src)
}

func (img *StdImage) SetAttrs(attrs map[string]string) {
	img.StdWidget.SetAttrs(attrs)
	if src, ok := attrs["src"]; ok {
		img.src = src
	}
}

func (img *StdImage) String() string {
	return fmt.Sprintf("StdImage src=%s %s", img.src, &img.StdWidget)
}

func (img *StdImage) widthForWriter() *float64 {
	if img.WidthIsSet() {
		width := ContentWidth(img)
		return &width
	}
	return nil
}

func (img *StdImage) heightForWriter() *float64 {
	if img.HeightIsSet() {
		height := ContentHeight(img)
		return &height
	}
	return nil
}

func init() {
	registerTag(DefaultSpace, "image", func() interface{} { return &StdImage{} })
}

var _ HasAttrs = (*StdImage)(nil)
var _ Identifier = (*StdImage)(nil)
var _ Printer = (*StdImage)(nil)
var _ WantsContainer = (*StdImage)(nil)
var _ WantsScope = (*StdImage)(nil)
