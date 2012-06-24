// Copyright 2011-2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"strings"
)

type array []writer

func (a array) write(w io.Writer) {
	fmt.Fprintf(w, "[")
	for _, v := range a {
		v.write(w)
	}
	fmt.Fprintf(w, "] ")
}

type body struct {
	list genWriterArray
}

func (b *body) add(w ...genWriter) {
	b.list = append(b.list, w...)
}

func (b *body) write(w lenWriter, ss *xRefSubSection) {
	for _, e := range b.list {
		xe := &inUseXRefEntry{w.Len(), e.Gen()}
		ss.add(xe)
		e.write(w)
	}
}

type boolean bool

func (b boolean) write(w io.Writer) {
	fmt.Fprintf(w, "%t ", b)
}

type catalog struct {
	dictionaryObject
	pageMode string
	pages    *pages
	outlines *outlines
}

func (c *catalog) init(seq, gen int, pageMode string, pages *pages, outlines *outlines) *catalog {
	c.dictionaryObject.init(seq, gen)
	c.dict["Type"] = name("Catalog")
	c.pageMode = pageMode
	c.dict["PageMode"] = name(pageMode)
	c.pages = pages
	c.dict["Pages"] = &indirectObjectRef{pages}
	c.outlines = outlines
	c.dict["Outlines"] = &indirectObjectRef{outlines}
	return c
}

func newCatalog(seq, gen int, pageMode string, pages *pages, outlines *outlines) *catalog {
	return new(catalog).init(seq, gen, pageMode, pages, outlines)
}

type dictionary map[string]writer

func (d dictionary) keys() []string {
	sa := make([]string, len(d))
	i := 0
	for k, _ := range d {
		sa[i] = k
		i++
	}
	sort.Strings(sa)
	return sa
}

func (d dictionary) write(w io.Writer) {
	fmt.Fprintf(w, "<<\n")
	for _, k := range d.keys() {
		name(k).write(w)
		d[k].write(w)
		fmt.Fprintf(w, "\n")
	}
	fmt.Fprintf(w, ">>\n")
}

type dictionaryObject struct {
	indirectObject
	dict dictionary
}

func (d *dictionaryObject) init(seq, gen int) *dictionaryObject {
	d.indirectObject.init(seq, gen, nil)
	d.dict = dictionary{}
	return d
}

func newDictionaryObject(seq, gen int) *dictionaryObject {
	return new(dictionaryObject).init(seq, gen)
}

func (d *dictionaryObject) write(w io.Writer) {
	d.writeHeader(w)
	d.dict.write(w)
	d.writeFooter(w)
}

type file struct {
	header  header
	body    body
	trailer trailer
}

func newFile() *file {
	return &file{header{}, body{}, trailer{dictionary{}, 0}}
}

func (f *file) write(w io.Writer) {
	var buf bytes.Buffer // TODO: replace with io.Writer wrapper that implements lenWriter interface
	var table xRefTable
	ss := newXRefSubSection()
	table.add(ss)
	f.header.write(&buf)
	f.body.write(&buf, ss)
	f.trailer.xrefTableStart = buf.Len()
	f.trailer.setXrefTableSize(ss.len())
	table.write(&buf)
	f.trailer.write(&buf)
	buf.WriteTo(w)
}

type freeXRefEntry indirectObject

func (e *freeXRefEntry) write(w io.Writer) {
	fmt.Fprintf(w, "%.10d %.5d f\n", e.seq, e.gen)
}

type genWriter interface {
	writer
	Gen() int
}

type genWriterArray []genWriter

type header struct {
	Version float32
}

func (h *header) write(w io.Writer) {
	v := h.Version
	if v == 0.0 {
		v = 1.3
	}
	fmt.Fprintf(w, "%%PDF-%1.1f\n", v)
}

type indirectObject struct {
	seq, gen int
	obj      writer
}

func (indObj *indirectObject) init(seq, gen int, obj writer) {
	indObj.seq = seq
	indObj.gen = gen
	indObj.obj = obj
}

func (indObj *indirectObject) Gen() int {
	return indObj.gen
}

func (indObj *indirectObject) Seq() int {
	return indObj.seq
}

func (indObj *indirectObject) write(w io.Writer) {
	indObj.writeHeader(w)
	if indObj.obj != nil {
		indObj.obj.write(w)
		fmt.Fprintf(w, "\n")
	}
	indObj.writeFooter(w)
}

func (indObj *indirectObject) writeHeader(w io.Writer) {
	fmt.Fprintf(w, "%d %d obj\n", indObj.seq, indObj.gen)
}

func (indObj *indirectObject) writeFooter(w io.Writer) {
	fmt.Fprintf(w, "endobj\n")
}

type indirectObjectRef struct {
	obj seqGen
}

func (ref *indirectObjectRef) write(w io.Writer) {
	fmt.Fprintf(w, "%d %d R ", ref.obj.Seq(), ref.obj.Gen())
}

type integer int

func (i integer) write(w io.Writer) {
	fmt.Fprintf(w, "%d ", i)
}

type inUseXRefEntry struct {
	byteOffset, gen int
}

func (e *inUseXRefEntry) write(w io.Writer) {
	fmt.Fprintf(w, "%.10d %.5d n\n", e.byteOffset, e.gen)
}

type lenWriter interface {
	io.Writer
	Len() int
}

type name string

func (n name) write(w io.Writer) {
	fmt.Fprintf(w, "/%s ", n)
}

func nameArray(names ...string) array {
	na := make(array, len(names))
	for i, n := range names {
		na[i] = name(n)
	}
	return na
}

type number struct {
	value interface{}
}

func (n number) write(w io.Writer) {
	fmt.Fprintf(w, "%v ", n.value)
}

type outlines struct {
	dictionaryObject
}

func (o *outlines) init(seq, gen int) *outlines {
	o.dictionaryObject.init(seq, gen)
	o.dict["Type"] = name("Outlines")
	return o
}

func newOutlines(seq, gen int) *outlines {
	return new(outlines).init(seq, gen)
}

func (o *outlines) write(w io.Writer) {
	o.dict["Count"] = integer(0)
	o.dictionaryObject.write(w)
}

type page struct {
	pageBase
	contents []*stream
}

func (p *page) init(seq, gen int, parent seqGen) *page {
	p.pageBase.init(seq, gen, parent)
	p.dict["Type"] = name("Page")
	return p
}

func newPage(seq, gen int, parent seqGen) *page {
	return new(page).init(seq, gen, parent)
}

func (p *page) add(s *stream) {
	p.contents = append(p.contents, s)
}

func (p *page) contentLength() (result int) {
	for _, s := range p.contents {
		result += s.len()
	}
	return
}

func (p *page) contentRefs() (result array) {
	result = make(array, len(p.contents))
	for i, s := range p.contents {
		result[i] = &indirectObjectRef{s}
	}
	return
}

// TODO: setAnnots
// TODO: setBeads

func (p *page) setThumb(thumb seqGen) {
	p.dict["Thumb"] = &indirectObjectRef{thumb}
}

func (p *page) write(w io.Writer) {
	p.writeHeader(w)
	p.writeBody(w)
	p.writeFooter(w)
}

func (p *page) writeBody(w io.Writer) {
	if len(p.contents) > 1 {
		p.dict["Contents"] = p.contentRefs()
	} else if len(p.contents) == 1 {
		p.dict["Contents"] = &indirectObjectRef{p.contents[0]}
	}
	p.dict["Length"] = integer(p.contentLength())
	p.dict.write(w)
}

type pageBase struct {
	dictionaryObject
}

func (pb *pageBase) init(seq, gen int, parent seqGen) *pageBase {
	pb.dictionaryObject.init(seq, gen)
	if parent != nil {
		pb.dict["Parent"] = &indirectObjectRef{parent}
	}
	return pb
}

func newPageBase(seq, gen int, parent *indirectObject) *pageBase {
	return new(pageBase).init(seq, gen, parent)
}

func (pb *pageBase) setCropBox(r rectangle) {
	pb.dict["CropBox"] = &r
}

func (pb *pageBase) setDuration(duration interface{}) {
	pb.dict["Dur"] = number{duration}
}

func (pb *pageBase) setHidden(hidden bool) {
	pb.dict["Hid"] = boolean(hidden)
}

func (pb *pageBase) setMediaBox(r rectangle) {
	pb.dict["MediaBox"] = &r
}

func (pb *pageBase) setResources(r *resources) {
	pb.dict["Resources"] = &indirectObjectRef{r}
}

func (pb *pageBase) setRotate(rotate int) {
	pb.dict["Rotate"] = integer(rotate)
}

type pages struct {
	pageBase
	kids []*page
}

func (ps *pages) init(seq, gen int) *pages {
	ps.pageBase.init(seq, gen, nil)
	ps.dict["Type"] = name("Pages")
	return ps
}

func newPages(seq, gen int) *pages {
	return new(pages).init(seq, gen)
}

func (ps *pages) add(p *page) {
	ps.kids = append(ps.kids, p)
}

func (ps *pages) write(w io.Writer) {
	ps.dict["Count"] = integer(len(ps.kids))
	kidsRefs := make(array, len(ps.kids))
	for i, kid := range ps.kids {
		kidsRefs[i] = &indirectObjectRef{kid}
	}
	ps.dict["Kids"] = kidsRefs
	ps.pageBase.write(w)
}

type rectangle struct {
	x1, y1, x2, y2 float64
}

func (r *rectangle) write(w io.Writer) {
	fmt.Fprintf(w, "[")
	number{r.x1}.write(w)
	number{r.y1}.write(w)
	number{r.x2}.write(w)
	number{r.y2}.write(w)
	fmt.Fprintf(w, "] ")
}

type resources struct {
	dictionaryObject
	fonts    dictionary
	xObjects dictionary
}

func (r *resources) init(seq, gen int) *resources {
	r.dictionaryObject.init(seq, gen)
	return r
}

func newResources(seq, gen int) *resources {
	return new(resources).init(seq, gen)
}

func (r *resources) setFont(name string, ref *indirectObjectRef) {
	if r.fonts == nil {
		r.fonts = dictionary{}
		r.dict["Font"] = r.fonts
	}
	r.fonts[name] = ref
}

func (r *resources) setProcSet(w writer) {
	r.dict["ProcSet"] = w
}

func (r *resources) setXObject(name string, ref *indirectObjectRef) {
	if r.xObjects == nil {
		r.xObjects = dictionary{}
		r.dict["XObject"] = r.xObjects
	}
	r.xObjects[name] = ref
}

type seqGen interface {
	Seq() int
	Gen() int
}

type str string

func (s str) escape() string {
	s1 := strings.Replace(string(s), "\\", "\\\\", -1)
	s2 := strings.Replace(s1, "(", "\\(", -1)
	s3 := strings.Replace(s2, ")", "\\)", -1)
	return s3
}

func (s str) write(w io.Writer) {
	fmt.Fprintf(w, "(%s) ", s.escape())
}

type stream struct {
	dictionaryObject
	data []byte
}

func (s *stream) init(seq, gen int, data []byte) *stream {
	s.dictionaryObject.init(seq, gen)
	s.data = make([]byte, len(data))
	copy(s.data, data)
	s.dict["Length"] = integer(len(data))
	return s
}

func newStream(seq, gen int, data []byte) *stream {
	return new(stream).init(seq, gen, data)
}

func (s *stream) len() int {
	return len(s.data)
}

func (s *stream) setFilter(filter string) {
	s.dict["Filter"] = name(filter)
}

func (s *stream) write(w io.Writer) {
	s.writeHeader(w)
	s.writeBody(w)
	s.writeFooter(w)
}

func (s *stream) writeBody(w io.Writer) {
	s.dict["Length"] = integer(len(s.data))
	s.dict.write(w)
	fmt.Fprintf(w, "stream\n")
	if s.data != nil {
		w.Write(s.data)
	}
	fmt.Fprintf(w, "endstream\n")
}

type trailer struct {
	dict           dictionary
	xrefTableStart int
}

func newTrailer() *trailer {
	return &trailer{dictionary{}, 0}
}

func (tr *trailer) setRoot(root seqGen) {
	tr.dict["Root"] = &indirectObjectRef{root}
}

func (tr *trailer) setXrefTableSize(size int) {
	tr.dict["Size"] = integer(size)
}

func (tr *trailer) xrefTableSize() int {
	if size, ok := tr.dict["Size"]; ok {
		return int(size.(integer))
	}
	return 0
}

func (tr *trailer) write(w io.Writer) {
	fmt.Fprintf(w, "trailer\n")
	tr.dict.write(w)
	fmt.Fprintf(w, "startxref\n")
	fmt.Fprintf(w, "%d\n", tr.xrefTableStart)
	fmt.Fprintf(w, "%%%%EOF\n")
}

type writer interface {
	write(io.Writer)
}

type xRefSubSection struct {
	list array
}

func newXRefSubSection() *xRefSubSection {
	return &xRefSubSection{array{&freeXRefEntry{0, 65535, nil}}}
}

func (ss *xRefSubSection) add(w writer) {
	ss.list = append(ss.list, w)
}

func (ss *xRefSubSection) len() int {
	return len(ss.list)
}

func (ss *xRefSubSection) write(w io.Writer) {
	fmt.Fprintf(w, "%d %d\n", 0, len(ss.list))
	for _, e := range ss.list {
		e.write(w)
	}
}

type xRefTable struct {
	// TODO: Can list be just a writer?
	list []*xRefSubSection
}

func (table *xRefTable) add(ss *xRefSubSection) {
	table.list = append(table.list, ss)
}

func (table *xRefTable) write(w io.Writer) {
	fmt.Fprintf(w, "xref\n")
	for _, e := range table.list {
		e.write(w)
	}
}
