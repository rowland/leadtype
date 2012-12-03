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

func TestMerge(t *testing.T) {
	a := Options{"a": "a", "b": 1}
	b := Options{"c": 3.5, "d": "d2"}
	c := a.Merge(b)
	// a and b should be unchanged
	expectNI(t, "length of a", 2, len(a))
	expectNI(t, "length of b", 2, len(b))
	// result should include all keys and values
	expectNI(t, "length of c", 4, len(c))
	expectV(t, "a", c["a"])
	expectV(t, 1, c["b"])
	expectV(t, 3.5, c["c"])
	expectV(t, "d2", c["d"])
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

func TestOptions_BoolDefault(t *testing.T) {
	o := Options{
		"true_value": true, "1": "1", "t": "t", "T": "T", "TRUE": "TRUE", "true": "true", "True": "True",
		"false_value": false, "0": "0", "f": "f", "F": "F", "FALSE": "FALSE", "false": "false", "False": "False",
	}
	expectB(t, false, o.BoolDefault("missing", false))
	expectB(t, true, o.BoolDefault("missing", true))

	expectB(t, true, o.BoolDefault("true_value", false))
	expectB(t, true, o.BoolDefault("1", false))
	expectB(t, true, o.BoolDefault("t", false))
	expectB(t, true, o.BoolDefault("T", false))
	expectB(t, true, o.BoolDefault("TRUE", false))
	expectB(t, true, o.BoolDefault("true", false))
	expectB(t, true, o.BoolDefault("True", false))

	expectB(t, false, o.BoolDefault("false_value", true))
	expectB(t, false, o.BoolDefault("0", true))
	expectB(t, false, o.BoolDefault("f", true))
	expectB(t, false, o.BoolDefault("F", true))
	expectB(t, false, o.BoolDefault("FALSE", true))
	expectB(t, false, o.BoolDefault("false", true))
	expectB(t, false, o.BoolDefault("False", true))
}

func TestOptions_FloatDefault(t *testing.T) {
	o := Options{"1st": "6.54", "2nd": 3.21, "3rd": 7, "4th": `33%`}
	expectF(t, 98.7, o.FloatDefault("missing", 98.7))
	expectF(t, 6.54, o.FloatDefault("1st", 0))
	expectF(t, 3.21, o.FloatDefault("2nd", -1))
	expectF(t, 7, o.FloatDefault("3rd", 100.0))
	expectF(t, 100.0, o.FloatDefault("4th", 100))
}

func TestOptions_StringDefault(t *testing.T) {
	o := Options{"i": 3, "s": "something", "f": 3.14}
	expectS(t, "3", o.StringDefault("i", ""))
	expectS(t, "something", o.StringDefault("s", ""))
	expectS(t, "3.14", o.StringDefault("f", ""))
	expectS(t, "missing", o.StringDefault("bogus", "missing"))
}
