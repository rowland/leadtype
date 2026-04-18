// Copyright 2011-2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import "github.com/rowland/leadtype/options"

var PageSizes = SizeMap{
	"letter": Size{612, 792},
	"legal":  Size{612, 1008},
	"A4":     Size{595, 842},
	"B5":     Size{499, 708},
	"C5":     Size{459, 649},
}

type pageStyle struct {
	orientation string
	landscape   bool
	pageSize    rectangle
	cropSize    rectangle
	rotate      int
}

func newPageStyle(options options.Options) *pageStyle {
	ps := new(pageStyle)
	ps.orientation = options.StringDefault("orientation", "portrait")
	pageSizeName := options.StringDefault("page_size", "letter")
	ps.pageSize = rectangleFromOptions(options, pageSizeName, "page_width", "page_height", ps.orientation)
	if options.HasKey("crop_size") || options.HasKey("crop_width") || options.HasKey("crop_height") {
		cropSizeName := options.StringDefault("crop_size", pageSizeName)
		ps.cropSize = rectangleFromOptions(options, cropSizeName, "crop_width", "crop_height", ps.orientation)
	} else {
		ps.cropSize = ps.pageSize
	}
	ps.rotate = lookupRotation(options.StringDefault("rotate", "portrait"))
	return ps
}

func makeSizeRectangle(size, orientation string) (r rectangle) {
	sz := PageSizes[size]
	if orientation == "landscape" {
		r.x2, r.y2 = sz.Height, sz.Width
	} else {
		r.x2, r.y2 = sz.Width, sz.Height
	}
	return
}

func rectangleFromOptions(options options.Options, sizeName, widthKey, heightKey, orientation string) rectangle {
	r := makeSizeRectangle(sizeName, orientation)
	if options.HasKey(widthKey) {
		r.x2 = options.FloatDefault(widthKey, r.x2)
	}
	if options.HasKey(heightKey) {
		r.y2 = options.FloatDefault(heightKey, r.y2)
	}
	return r
}

const (
	PORTRAIT  = 0
	LANDSCAPE = 270
)

func lookupRotation(rotate string) int {
	if rotate == "landscape" {
		return LANDSCAPE
	}
	return PORTRAIT
}
