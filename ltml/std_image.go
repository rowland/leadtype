// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
	"io"
)

// dataImageWriter is an optional extension of Writer whose implementations
// can render image data supplied as raw bytes rather than a filename. This
// allows the asset filesystem seam to resolve files without altering the
// Writer interface.
type dataImageWriter interface {
	ImageDimensions(data []byte) (width, height int, err error)
	PrintImage(data []byte, x, y float64, width, height *float64) (actualWidth, actualHeight float64, err error)
}

type StdImage struct {
	StdWidget
	src string
}

// readAsset opens img.src via the scope's asset filesystem when one is set,
// returning the raw bytes. Returns nil, nil when no asset filesystem is
// attached; callers should then fall back to file-path–based Writer methods.
func (img *StdImage) readAsset() ([]byte, error) {
	if img.scope == nil {
		return nil, nil
	}
	fsys := img.scope.AssetFS()
	if fsys == nil {
		return nil, nil
	}
	f, err := fsys.Open(img.src)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(f)
}

func (img *StdImage) DrawContent(w Writer) error {
	if img.src == "" {
		return fmt.Errorf("image src must be specified")
	}
	if dw, ok := w.(dataImageWriter); ok {
		if data, err := img.readAsset(); err != nil {
			return err
		} else if data != nil {
			_, _, err = dw.PrintImage(data, ContentLeft(img), ContentTop(img), img.widthForWriter(), img.heightForWriter())
			return err
		}
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

// imageDimensions returns the pixel dimensions of img.src, using the scope's
// asset filesystem when available and falling back to the Writer's file-based
// method otherwise.
func (img *StdImage) imageDimensions(w Writer) (width, height int, err error) {
	if dw, ok := w.(dataImageWriter); ok {
		if data, readErr := img.readAsset(); readErr != nil {
			return 0, 0, readErr
		} else if data != nil {
			return dw.ImageDimensions(data)
		}
	}
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
