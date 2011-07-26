package pdf

import (
	"testing"
	"bytes"
)

func TestPageWriter(t *testing.T) {
	var buf bytes.Buffer
	dw := NewDocWriter(&buf)
	pw1 := newPageWriter(dw, Options{})
	expectS(t, "pt", pw1.units.name)
	expectF(t, 1, pw1.units.ratio)
	pw2 := newPageWriter(dw, Options{"units": "in"})
	expectS(t, "in", pw2.units.name)
	expectF(t, 72, pw2.units.ratio)
	pw3 := newPageWriter(dw, Options{"units": "cm"})
	expectS(t, "cm", pw3.units.name)
	expectF(t, 28.35, pw3.units.ratio)
	UnitConversions.Add("dp", 0.072)
	pw4 := newPageWriter(dw, Options{"units": "dp"})
	expectS(t, "dp", pw4.units.name)
	expectF(t, 0.072, pw4.units.ratio)
}
