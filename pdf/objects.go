package pdf

import (
	"io"
	"fmt"
	"strings"
	"sort"
)

type writer interface {
	write(io.Writer)
}

type array []writer

func (a array) write(w io.Writer) {
	fmt.Fprintf(w, "[")
	for _, v := range a {
		v.write(w)
	}
	fmt.Fprintf(w, "] ")
}

type boolean bool

func (b boolean) write(w io.Writer) {
	fmt.Fprintf(w, "%t ", b)
}

type dictionary map[string]writer

func (d dictionary) keys() []string {
	sa := make([]string, len(d))
	i := 0
	for k, _ := range d {
		sa[i] = k
		i++
	}
	sort.SortStrings(sa)
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

func newDictionaryObject(seq, gen int, dict dictionary) *dictionaryObject {
	return &dictionaryObject{indirectObject{seq, gen}, dict}
}

func (d *dictionaryObject) write(w io.Writer) {
	d.writeHeader(w)
	d.dict.write(w)
	d.writeFooter(w)
}

type freeXRefEntry indirectObject

func (e *freeXRefEntry) write(w io.Writer) {
	fmt.Fprintf(w, "%.10d %.5d f\n", e.seq, e.gen)
}

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
}

func (obj *indirectObject) writeHeader(w io.Writer) {
	fmt.Fprintf(w, "%d %d obj\n", obj.seq, obj.gen)
}

func (obj *indirectObject) writeFooter(w io.Writer) {
	fmt.Fprintf(w, "endobj\n")
}

type indirectObjectRef struct {
	obj *indirectObject
}

func (ref *indirectObjectRef) write(w io.Writer) {
	fmt.Fprintf(w, "%d %d R ", ref.obj.seq, ref.obj.gen)
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

type rectangle struct {
	x1, y1, x2, y2 float32
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
	fonts dictionary
	xObjects dictionary
}

func newResources(seq, gen int) *resources {
	return &resources{dictionaryObject{indirectObject{seq, gen}, dictionary{}}, nil, nil}
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

type trailer struct {
	dict           dictionary
	xrefTableStart int
}

func newTrailer() *trailer {
	return &trailer{dictionary{}, 0}
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

type xRefSubSection struct {
	list array
}

func newXRefSubSection() *xRefSubSection {
	return &xRefSubSection{array{&freeXRefEntry{0, 65535}}}
}

func (ss *xRefSubSection) add(w writer) {
	ss.list = append(ss.list, w)
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
