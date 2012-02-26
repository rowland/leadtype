package pdf

import "testing"

func TestMakeSizeRectangle(t *testing.T) {
	r := makeSizeRectangle("letter", "portrait")
	expectF(t, 0, r.x1)
	expectF(t, 0, r.y1)
	expectF(t, 612, r.x2)
	expectF(t, 792, r.y2)
}

func TestNewPageStyle(t *testing.T) {
	opt1 := Options{}
	ps1 := newPageStyle(opt1)
	expectF(t, 0, ps1.pageSize.x1)
	expectF(t, 0, ps1.pageSize.y1)
	expectF(t, 612, ps1.pageSize.x2)
	expectF(t, 792, ps1.pageSize.y2)
	expectF(t, 0, ps1.cropSize.x1)
	expectF(t, 0, ps1.cropSize.y1)
	expectF(t, 612, ps1.cropSize.x2)
	expectF(t, 792, ps1.cropSize.y2)

	opt2 := Options{"orientation": "landscape"}
	ps2 := newPageStyle(opt2)
	expectF(t, 0, ps2.pageSize.x1)
	expectF(t, 0, ps2.pageSize.y1)
	expectF(t, 792, ps2.pageSize.x2)
	expectF(t, 612, ps2.pageSize.y2)
	expectF(t, 0, ps2.cropSize.x1)
	expectF(t, 0, ps2.cropSize.y1)
	expectF(t, 792, ps2.cropSize.x2)
	expectF(t, 612, ps2.cropSize.y2)

	opt3 := Options{"page_size": "A4"}
	ps3 := newPageStyle(opt3)
	expectF(t, 0, ps3.pageSize.x1)
	expectF(t, 0, ps3.pageSize.y1)
	expectF(t, 595, ps3.pageSize.x2)
	expectF(t, 842, ps3.pageSize.y2)
	expectF(t, 0, ps3.cropSize.x1)
	expectF(t, 0, ps3.cropSize.y1)
	expectF(t, 595, ps3.cropSize.x2)
	expectF(t, 842, ps3.cropSize.y2)

	opt4 := Options{"page_size": "A4", "orientation": "landscape"}
	ps4 := newPageStyle(opt4)
	expectF(t, 0, ps4.pageSize.x1)
	expectF(t, 0, ps4.pageSize.y1)
	expectF(t, 842, ps4.pageSize.x2)
	expectF(t, 595, ps4.pageSize.y2)
	expectF(t, 0, ps4.cropSize.x1)
	expectF(t, 0, ps4.cropSize.y1)
	expectF(t, 842, ps4.cropSize.x2)
	expectF(t, 595, ps4.cropSize.y2)

	// TODO: Handle crop_size differently to allow non-zero x1 and y1.

	opt5 := Options{"rotate": "portrait"}
	ps5 := newPageStyle(opt5)
	expectI(t, 0, ps5.rotate)

	opt6 := Options{"rotate": "landscape"}
	ps6 := newPageStyle(opt6)
	expectI(t, 270, ps6.rotate)
}
