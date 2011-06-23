package pdf

import (
	"io"
	"fmt"
)

type Boolean bool

func (b Boolean) Write(w io.Writer) {
	fmt.Fprintf(w, "%t ", b)
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
