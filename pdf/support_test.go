package pdf

import (
	"testing"
	"math"
)

func check(t *testing.T, condition bool, msg string) {
	if !condition {
		t.Error(msg)
	}
}

func expectF(t *testing.T, expected, actual float64) {
	if expected != actual {
		t.Errorf("Expected %f, got %f", expected, actual)
	}
}

func expectFdelta(t *testing.T, expected, actual, delta float64) {
	if math.Fabs(expected - actual) > delta {
		t.Errorf("Expected %f, got %f", expected, actual)
	}
}

func expectI(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected %d, got %d", expected, actual)
	}
}

func expectS(t *testing.T, expected, actual string) {
	if expected != actual {
		t.Errorf("Expected %s, got %s", expected, actual)
	}
}

func expectV(t *testing.T, expected, actual interface{}) {
	if expected != actual {
		t.Errorf("Expected %v, got %v", expected, actual)
	}
}

func TestMerge(t *testing.T) {
	a := Options{"a": "a", "b": 1}
	b := Options{"c": 3.5, "d": "d2"}
	c := a.Merge(b)
	// a and b should be unchanged
	expectI(t, 2, len(a))
	expectI(t, 2, len(b))
	// result should include all keys and values
	expectI(t, 4, len(c))
	expectV(t, "a", c["a"])
	expectV(t, 1, c["b"])
	expectV(t, 3.5, c["c"])
	expectV(t, "d2", c["d"])
}

func TestRGBfromColor(t *testing.T) {
	r, g, b := rgbFromColor(0x030507)
	expectI(t, 3, int(r))
	expectI(t, 5, int(g))
	expectI(t, 7, int(b))
}
