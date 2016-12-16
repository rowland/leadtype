// Copyright 2011-2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import (
	"math"
	"testing"
)

func check(t *testing.T, condition bool, msg string) {
	if !condition {
		t.Error(msg)
	}
}

func expect(t *testing.T, name string, condition bool) {
	if !condition {
		t.Errorf("%s: failed condition", name)
	}
}

func checkFatal(t *testing.T, condition bool, msg string) {
	if !condition {
		t.Fatal(msg)
	}
}

func expectB(t *testing.T, expected, actual bool) {
	if expected != actual {
		t.Errorf("Expected %t, got %t", expected, actual)
	}
}

func expectF(t *testing.T, expected, actual float64) {
	if expected != actual {
		t.Errorf("Expected %f, got %f", expected, actual)
	}
}

func expectNF(t *testing.T, name string, expected, actual float64) {
	if expected != actual {
		t.Errorf("%s: expected %f, got %f", name, expected, actual)
	}
}

func expectNFdelta(t *testing.T, name string, expected, actual, delta float64) {
	if math.Abs(expected-actual) > delta {
		t.Errorf("%s: expected %f, got %f", name, expected, actual)
	}
}

func expectFdelta(t *testing.T, expected, actual, delta float64) {
	if math.Abs(expected-actual) > delta {
		t.Errorf("Expected %f, got %f", expected, actual)
	}
}

func expectI(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected %d, got %d", expected, actual)
	}
}

func expectNI(t *testing.T, name string, expected, actual int) {
	if expected != actual {
		t.Errorf("%s: expected %d, got %d", name, expected, actual)
	}
}

func expectS(t *testing.T, expected, actual string) {
	if expected != actual {
		t.Errorf("Expected |%s|, got |%s|", expected, actual)
	}
}

func expectNS(t *testing.T, name string, expected, actual string) {
	if expected != actual {
		t.Errorf("%s: expected %s, got %s", name, expected, actual)
	}
}

func expectV(t *testing.T, expected, actual interface{}) {
	if expected != actual {
		t.Errorf("Expected %v, got %v", expected, actual)
	}
}

func TestIntSlice(t *testing.T) {
	values := []int{1, 20, 3000000}
	expectS(t, "1 20 3000000", intSlice(values).join(" "))
	expectS(t, "1, 20, 3000000", intSlice(values).join(", "))
}

func TestFloat64Slice(t *testing.T) {
	values := []float64{1, 2.5, 3.1415926}
	expectS(t, "1 2.5 3.1416", float64Slice(values).join(" "))
	expectS(t, "1, 2.5, 3.1416", float64Slice(values).join(", "))
}
