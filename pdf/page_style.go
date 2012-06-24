// Copyright 2011-2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

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

func newPageStyle(options Options) *pageStyle {
	ps := new(pageStyle)
	ps.orientation = options.StringDefault("orientation", "portrait")
	pageSizeName := options.StringDefault("page_size", "letter")
	ps.pageSize = makeSizeRectangle(pageSizeName, ps.orientation)
	cropSizeName := options.StringDefault("crop_size", pageSizeName)
	ps.cropSize = makeSizeRectangle(cropSizeName, ps.orientation)
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
