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

func (d *DictionaryObject) Write(w io.Writer) {
	d.WriteHeader(w)
	d.dict.Write(w)
	d.WriteFooter(w)
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
