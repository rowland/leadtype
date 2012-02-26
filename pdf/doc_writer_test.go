package pdf

import (
	"bytes"
	"testing"
)

func TestNewDocWriter(t *testing.T) {
	var buf bytes.Buffer
	dw := NewDocWriter(&buf)
	check(t, dw.wr == &buf, "DocWriter should have reference to io.Writer")
	check(t, dw.nextSeq != nil, "DocWriter should have nextSeq func")
	check(t, dw.file != nil, "DocWriter should have file")
	check(t, dw.catalog != nil, "DocWriter should have catalog")
	check(t, len(dw.file.body.list) == 4, "DocWriter file body should be initialized")
	check(t, dw.file.trailer.dict["Root"] != nil, "DocWriter file trailer root should be set")
	check(t, dw.resources != nil, "DocWriter resources should be initialized")
	check(t, dw.PagesAcross() == 1, "DocWriter PagesAcross should default to 1")
	check(t, dw.PagesDown() == 1, "DocWriter PagesDown should default to 1")
	check(t, dw.PagesUp() == 1, "DocWriter PagesUp should default to 1")
	check(t, dw.pages == nil, "DocWriter pages should be nil")
	check(t, dw.curPage == nil, "DocWriter curPage should be nil")
	check(t, !dw.inPage(), "DocWriter should not be in page yet")
}

func TestClose(t *testing.T) {
	var buf bytes.Buffer
	dw := NewDocWriter(&buf)
	dw.OpenPage(Options{})
	dw.Close()
	check(t, !dw.inPage(), "DocWriter should not be in page anymore")

	dw2 := NewDocWriter(&buf)
	dw2.Close()
	check(t, len(dw.pages) == 1, "DocWriter pages should have minimum of 1 page after Close")
}

func TestClosePage(t *testing.T) {
	var buf bytes.Buffer
	dw := NewDocWriter(&buf)
	dw.ClosePage()         // ignored when not in a page
	dw.OpenPage(Options{}) // TODO: save PageWriter reference and test for being closed
	dw.ClosePage()
	check(t, !dw.inPage(), "DocWriter should not be in page anymore")
	check(t, dw.curPage == nil, "DocWriter curPage should be nil again")
}

func TestOpen(t *testing.T) {
	var buf bytes.Buffer
	dw := NewDocWriter(&buf)
	dw.Open(Options{})
	// TODO: test options
}

func TestOpenPage(t *testing.T) {
	var buf bytes.Buffer
	dw := NewDocWriter(&buf)
	dw.OpenPage(Options{})
	check(t, dw.curPage != nil, "DocWriter curPage should not be nil")
	check(t, dw.inPage(), "DocWriter should be in page now")
	check(t, len(dw.pages) == 1, "DocWriter should have 1 page now")
}

// TODO: TestPagesAcross
// TODO: TestPagesDown
// TODO: TestPagesUp
