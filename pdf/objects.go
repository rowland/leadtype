package pdf

import (
	"io"
	"fmt"
	"strings"
	"sort"
)

type Writer interface {
	Write(io.Writer)
}

type Array []Writer

func (a Array) Write(w io.Writer) {
	fmt.Fprintf(w, "[")
	for _, v := range a {
		v.Write(w)
	}
	fmt.Fprintf(w, "] ")
}

type Boolean bool

func (b Boolean) Write(w io.Writer) {
	fmt.Fprintf(w, "%t ", b)
}

type Dictionary map[string]Writer

func (d Dictionary) keys() []string {
	sa := make([]string, len(d))
	i := 0
	for k, _ := range d {
		sa[i] = k
		i++
	}
	sort.SortStrings(sa)
	return sa
}

func (d Dictionary) Write(w io.Writer) {
	fmt.Fprintf(w, "<<\n")
	for _, k := range d.keys() {
		Name(k).Write(w)
		d[k].Write(w)
		fmt.Fprintf(w, "\n")
	}
	fmt.Fprintf(w, ">>\n")
}

type DictionaryObject struct {
	IndirectObject
	dict Dictionary
}

func newDictionaryObject(seq, gen int, dict Dictionary) *DictionaryObject {
	return &DictionaryObject{IndirectObject{seq, gen}, dict}
}

func (d *DictionaryObject) Write(w io.Writer) {
	d.WriteHeader(w)
	d.dict.Write(w)
	d.WriteFooter(w)
}

type FreeXRefEntry IndirectObject

func (e *FreeXRefEntry) Write(w io.Writer) {
	fmt.Fprintf(w, "%.10d %.5d f\n", e.seq, e.gen)
}

type Header struct {
	Version float32
}

func (h *Header) Write(w io.Writer) {
	v := h.Version
	if v == 0.0 {
		v = 1.3
	}
	fmt.Fprintf(w, "%%PDF-%1.1f\n", v)
}

type IndirectObject struct {
	seq, gen int
}

func (obj *IndirectObject) WriteHeader(w io.Writer) {
	fmt.Fprintf(w, "%d %d obj\n", obj.seq, obj.gen)
}

func (obj *IndirectObject) WriteFooter(w io.Writer) {
	fmt.Fprintf(w, "endobj\n")
}

type IndirectObjectRef struct {
	obj *IndirectObject
}

func (ref *IndirectObjectRef) Write(w io.Writer) {
	fmt.Fprintf(w, "%d %d R ", ref.obj.seq, ref.obj.gen)
}

type Integer int

func (i Integer) Write(w io.Writer) {
	fmt.Fprintf(w, "%d ", i)
}

type InUseXRefEntry struct {
	byteOffset, gen int
}

func (e *InUseXRefEntry) Write(w io.Writer) {
	fmt.Fprintf(w, "%.10d %.5d n\n", e.byteOffset, e.gen)
}

type Name string

func (n Name) Write(w io.Writer) {
	fmt.Fprintf(w, "/%s ", n)
}

func NameArray(names ...string) Array {
	na := make(Array, len(names))
	for i, n := range names {
		na[i] = Name(n)
	}
	return na
}

type Number struct {
	value interface{}
}

func (n Number) Write(w io.Writer) {
	fmt.Fprintf(w, "%v ", n.value)
}

type Rectangle struct {
	x1, y1, x2, y2 float32
}

func (r *Rectangle) Write(w io.Writer) {
	fmt.Fprintf(w, "[")
	Number{r.x1}.Write(w)
	Number{r.y1}.Write(w)
	Number{r.x2}.Write(w)
	Number{r.y2}.Write(w)
	fmt.Fprintf(w, "] ")
}

type Resources struct {
	DictionaryObject
	fonts Dictionary
	xObjects Dictionary
}

func newResources(seq, gen int) *Resources {
	return &Resources{DictionaryObject{IndirectObject{seq, gen}, Dictionary{}}, nil, nil}
}

func (r *Resources) setFont(name string, ref *IndirectObjectRef) {
	if r.fonts == nil {
		r.fonts = Dictionary{}
		r.dict["Font"] = r.fonts
	}
	r.fonts[name] = ref
}

func (r *Resources) setProcSet(w Writer) {
	r.dict["ProcSet"] = w
}

func (r *Resources) setXObject(name string, ref *IndirectObjectRef) {
	if r.xObjects == nil {
		r.xObjects = Dictionary{}
		r.dict["XObject"] = r.xObjects
	}
	r.xObjects[name] = ref
}

type String string

func (s String) escape() string {
	s1 := strings.Replace(string(s), "\\", "\\\\", -1)
	s2 := strings.Replace(s1, "(", "\\(", -1)
	s3 := strings.Replace(s2, ")", "\\)", -1)
	return s3
}

func (s String) Write(w io.Writer) {
	fmt.Fprintf(w, "(%s) ", s.escape())
}

type Trailer struct {
	dict           Dictionary
	xrefTableStart int
}

func newTrailer() *Trailer {
	return &Trailer{Dictionary{}, 0}
}

func (tr *Trailer) setXrefTableSize(size int) {
	tr.dict["Size"] = Integer(size)
}

func (tr *Trailer) xrefTableSize() int {
	if size, ok := tr.dict["Size"]; ok {
		return int(size.(Integer))
	}
	return 0
}

func (tr *Trailer) Write(w io.Writer) {
	fmt.Fprintf(w, "trailer\n")
	tr.dict.Write(w)
	fmt.Fprintf(w, "startxref\n")
	fmt.Fprintf(w, "%d\n", tr.xrefTableStart)
	fmt.Fprintf(w, "%%%%EOF\n")
}

type XRefSubSection struct {
	list Array
}

func newXRefSubSection() *XRefSubSection {
	return &XRefSubSection{Array{&FreeXRefEntry{0, 65535}}}
}

func (ss *XRefSubSection) add(w Writer) {
	ss.list = append(ss.list, w)
}

func (ss *XRefSubSection) Write(w io.Writer) {
	fmt.Fprintf(w, "%d %d\n", 0, len(ss.list))
	for _, e := range ss.list {
		e.Write(w)
	}
}

type XRefTable struct {
	// TODO: Can list be just a Writer?
	list []*XRefSubSection
}

func (table *XRefTable) add(ss *XRefSubSection) {
	table.list = append(table.list, ss)
}

func (table *XRefTable) Write(w io.Writer) {
	fmt.Fprintf(w, "xref\n")
	for _, e := range table.list {
		e.Write(w)
	}
}
