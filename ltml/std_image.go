// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
	"strings"
)

type StdImage struct {
	StdWidget
	src string
}

func (img *StdImage) LayoutWidget(w Writer) error {
	infoWidth, infoHeight, err := img.imageDimensions(w)
	if err != nil {
		return err
	}
	if infoWidth <= 0 || infoHeight <= 0 {
		return nil
	}
	width, height := imageLikeLayoutSize(&img.StdWidget, float64(infoWidth), float64(infoHeight))
	img.ResolveWidth(width)
	img.ResolveHeight(height)
	return nil
}

func (img *StdImage) DrawContent(w Writer) error {
	return withGraphicAccessibility(w, &img.StdWidget, "Figure", func() error {
		ref, err := img.assetSource()
		if err != nil {
			return err
		}
		if ref.identifier == "" {
			return fmt.Errorf("image src must be specified")
		}
		width, height := img.placementSizeForWriter(w)
		_, _, err = w.PrintImageFile(ref.identifier, ContentLeft(img), ContentTop(img), width, height)
		return err
	})
}

func (img *StdImage) PreferredHeight(w Writer) (float64, error) {
	infoWidth, infoHeight, err := img.imageDimensions(w)
	if err != nil {
		return 0, err
	}
	if infoWidth == 0 {
		return NonContentHeight(img), nil
	}
	_, height := imageLikeLayoutSize(&img.StdWidget, float64(infoWidth), float64(infoHeight))
	return height, nil
}

func (img *StdImage) PreferredWidth(w Writer) (float64, error) {
	infoWidth, infoHeight, err := img.imageDimensions(w)
	if err != nil {
		return 0, err
	}
	if infoHeight == 0 {
		return NonContentWidth(img), nil
	}
	width, _ := imageLikeLayoutSize(&img.StdWidget, float64(infoWidth), float64(infoHeight))
	return width, nil
}

func (img *StdImage) IntrinsicAspectRatio(w Writer) (float64, bool) {
	if w == nil {
		return 0, false
	}
	infoWidth, infoHeight, err := img.imageDimensions(w)
	if err != nil || infoWidth <= 0 || infoHeight <= 0 {
		return 0, false
	}
	return float64(infoWidth) / float64(infoHeight), true
}

func (img *StdImage) imageDimensions(w Writer) (width, height int, err error) {
	ref, err := img.assetSource()
	if err != nil {
		return 0, 0, err
	}
	if ref.identifier == "" {
		return 0, 0, nil
	}
	return w.ImageDimensionsFromFile(ref.identifier)
}

func (img *StdImage) SetAttrs(attrs map[string]string) {
	img.StdWidget.SetAttrs(attrs)
	if src, ok := attrs["src"]; ok {
		img.src = src
	}
}

func (img *StdImage) assetSource() (assetSourceRef, error) {
	if strings.TrimSpace(img.src) == "" {
		return assetSourceRef{}, nil
	}
	if img.doc == nil {
		return assetSourceRef{}, fmt.Errorf("image document is not set")
	}
	return img.doc.resolveAssetSource(img.container, img.src)
}

func (img *StdImage) String() string {
	return fmt.Sprintf("StdImage src=%s %s", img.src, &img.StdWidget)
}

func (img *StdImage) placementSizeForWriter(w Writer) (width, height *float64) {
	infoWidth, infoHeight, err := img.imageDimensions(w)
	if err != nil {
		return imageLikeFallbackPlacementSize(&img.StdWidget)
	}
	return imageLikePlacementSize(&img.StdWidget, float64(infoWidth), float64(infoHeight))
}

func init() {
	registerTag(DefaultSpace, "image", func() any { return &StdImage{} })
}

var _ HasAttrs = (*StdImage)(nil)
var _ Identifier = (*StdImage)(nil)
var _ Printer = (*StdImage)(nil)
var _ IntrinsicAspectRatioProvider = (*StdImage)(nil)
var _ WantsContainer = (*StdImage)(nil)
var _ WantsDoc = (*StdImage)(nil)
var _ WantsScope = (*StdImage)(nil)
