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
