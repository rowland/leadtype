// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
	"strings"
)

type StdImage struct {
	StdWidget
	src               string
	data              []byte
	doc               *Doc
	dataLoadAttempted bool
	dataLoadErr       error
}

func (img *StdImage) DrawContent(w Writer) error {
	return withGraphicAccessibility(w, &img.StdWidget, "Figure", func() error {
		if err := img.ensureData(); err != nil {
			return err
		}
		if len(img.data) == 0 {
			return fmt.Errorf("image src must be specified")
		}
		_, _, err := w.PrintImage(img.data, ContentLeft(img), ContentTop(img), img.widthForWriter(), img.heightForWriter())
		return err
	})
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
	if err := img.ensureData(); err != nil {
		return 0, 0, err
	}
	if len(img.data) == 0 {
		return 0, 0, nil
	}
	return w.ImageDimensions(img.data)
}

func (img *StdImage) SetDoc(doc *Doc) {
	img.doc = doc
}

func (img *StdImage) SetAttrs(attrs map[string]string) {
	img.StdWidget.SetAttrs(attrs)
	if src, ok := attrs["src"]; ok {
		if img.src != src {
			img.data = nil
			img.dataLoadAttempted = false
			img.dataLoadErr = nil
		}
		img.src = src
	}
}

func (img *StdImage) ensureData() error {
	if strings.TrimSpace(img.src) == "" {
		return nil
	}
	if img.dataLoadAttempted {
		return img.dataLoadErr
	}
	img.dataLoadAttempted = true
	data, err := loadAssetBytes(img.doc, img.container, img.src)
	if err != nil {
		img.dataLoadErr = err
		return err
	}
	img.data = data
	img.dataLoadErr = nil
	return nil
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
	registerTag(DefaultSpace, "image", func() any { return &StdImage{} })
}

var _ HasAttrs = (*StdImage)(nil)
var _ Identifier = (*StdImage)(nil)
var _ Printer = (*StdImage)(nil)
var _ WantsContainer = (*StdImage)(nil)
var _ WantsDoc = (*StdImage)(nil)
var _ WantsScope = (*StdImage)(nil)
