package pdf

import (
	"io"
)

type DocWriter struct {
	wr          io.Writer
	pages       []*PageWriter
	nextSeq     func() int
	file        *file
	catalog     *catalog
	resources   *resources
	pagesAcross int
	pagesDown   int
	curPage     *PageWriter
}

func NewDocWriter(wr io.Writer) *DocWriter {
	nextSeq := nextSeqFunc()
	file := newFile()
	pages := newPages(nextSeq(), 0)
	outlines := newOutlines(nextSeq(), 0)
	catalog := newCatalog(nextSeq(), 0, "UseNone", pages, outlines)
	file.body.add(pages, outlines, catalog)
	file.trailer.setRoot(catalog)
	resources := newResources(nextSeq(), 0)
	resources.setProcSet(nameArray("PDF", "Text", "ImageB", "ImageC"))
	file.body.add(resources)
	return &DocWriter{wr: wr, nextSeq: nextSeq, file: file, catalog: catalog, resources: resources}
}

func nextSeqFunc() func() int {
	var nextValue = 0
	return func() int {
		nextValue++
		return nextValue
	}
}

func (dw *DocWriter) Close() {
	if len(dw.pages) == 0 {
		dw.OpenPage()
	}
	if dw.inPage() {
		dw.ClosePage()
	}
}

func (dw *DocWriter) ClosePage() {
	if dw.inPage() {
		dw.curPage.Close()
		dw.curPage = nil
	}
}

func (dw *DocWriter) inPage() bool {
	return dw.curPage != nil
}

func (dw *DocWriter) Open() {
	// TODO: assign options
}

func (dw *DocWriter) OpenPage() *PageWriter {
	if dw.inPage() {
		dw.ClosePage()
	}
	dw.curPage = newPageWriter()
	dw.pages = append(dw.pages, dw.curPage)
	return dw.curPage
}

func (dw *DocWriter) PagesAcross() int {
	if dw.pagesAcross < 1 {
		return 1
	}
	return dw.pagesAcross
}

func (dw *DocWriter) PagesDown() int {
	if dw.pagesDown < 1 {
		return 1
	}
	return dw.pagesDown
}

func (dw *DocWriter) PagesUp() int {
	return dw.PagesAcross() * dw.PagesDown()
}
