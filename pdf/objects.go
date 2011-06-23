package pdf

import (
	"io"
	"fmt"
)

type Header struct {
	Version float32
}

func (header *Header) Write(w io.Writer) {
	v := header.Version
	if v == 0.0 {
		v = 1.3
	}
	fmt.Fprintf(w, "%%PDF-%1.1f\n", v)
}

type InUseXRefEntry struct {
	byteOffset, gen int
}

func (entry *InUseXRefEntry) Write(w io.Writer) {
	fmt.Fprintf(w, "%.10d %.5d n\n", entry.byteOffset, entry.gen)
}

type PdfInteger int

func (i PdfInteger) Write(w io.Writer) {
	fmt.Fprintf(w, "%d ", i)
}

type PdfNumber struct {
	value interface{}
}

func (n PdfNumber) Write(w io.Writer) {
	fmt.Fprintf(w, "%v ", n.value)
}

type PdfBoolean bool

func (b PdfBoolean) Write(w io.Writer) {
	fmt.Fprintf(w, "%t ", b)
}
