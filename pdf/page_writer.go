package pdf

import (
	"bytes"
)

type PageWriter struct {
	dw     *DocWriter
	page   *page
	stream bytes.Buffer
	units  *units
}

func newPageWriter(dw *DocWriter, options Options) *PageWriter {
	ps := newPageStyle(options)
	units := UnitConversions[options.StringDefault("units", "pt")]
	page := newPage(dw.nextSeq(), 0, dw.catalog.pages)
	page.setMediaBox(ps.pageSize)
	page.setCropBox(ps.cropSize)
	page.setRotate(ps.rotate)
	page.setResources(dw.resources)
	dw.file.body.add(page)
	return &PageWriter{dw: dw, page: page, units: units}
}

func (pw *PageWriter) Close() {
	// end margins
	// end sub page
	// end text
	// end graph
	// end misc
	// compress stream
	pdf_stream := newStream(pw.dw.nextSeq(), 0, pw.stream.Bytes())
	pw.dw.file.body.add(pdf_stream)
	// set annots
	pw.page.add(pdf_stream)
	pw.dw.catalog.pages.add(pw.page) // unless reusing page
	pw.stream.Reset()
}
