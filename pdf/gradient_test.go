// Copyright 2026 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import (
	"bytes"
	"strings"
	"testing"

	"github.com/rowland/leadtype/colors"
)

func TestLinearGradient_Validate(t *testing.T) {
	valid := &LinearGradient{
		X0: 0, Y0: 0, X1: 100, Y1: 0,
		Stops: []GradientStop{
			{Position: 0, Color: colors.Red},
			{Position: 1, Color: colors.Blue},
		},
	}
	check(t, valid.validate() == nil, "valid gradient should pass validation")

	tooFew := &LinearGradient{
		X0: 0, Y0: 0, X1: 100, Y1: 0,
		Stops: []GradientStop{
			{Position: 0, Color: colors.Red},
		},
	}
	check(t, tooFew.validate() != nil, "single stop should fail validation")

	outOfRange := &LinearGradient{
		X0: 0, Y0: 0, X1: 100, Y1: 0,
		Stops: []GradientStop{
			{Position: -0.1, Color: colors.Red},
			{Position: 1, Color: colors.Blue},
		},
	}
	check(t, outOfRange.validate() != nil, "negative position should fail validation")

	unsorted := &LinearGradient{
		X0: 0, Y0: 0, X1: 100, Y1: 0,
		Stops: []GradientStop{
			{Position: 0.5, Color: colors.Red},
			{Position: 0.2, Color: colors.Blue},
		},
	}
	check(t, unsorted.validate() != nil, "unsorted positions should fail validation")

	missingStart := &LinearGradient{
		X0: 0, Y0: 0, X1: 100, Y1: 0,
		Stops: []GradientStop{
			{Position: 0.25, Color: colors.Red},
			{Position: 1, Color: colors.Blue},
		},
	}
	check(t, missingStart.validate() != nil, "first stop must be at 0")

	missingEnd := &LinearGradient{
		X0: 0, Y0: 0, X1: 100, Y1: 0,
		Stops: []GradientStop{
			{Position: 0, Color: colors.Red},
			{Position: 0.75, Color: colors.Blue},
		},
	}
	check(t, missingEnd.validate() != nil, "last stop must be at 1")

	repeated := &LinearGradient{
		X0: 0, Y0: 0, X1: 100, Y1: 0,
		Stops: []GradientStop{
			{Position: 0, Color: colors.Red},
			{Position: 0, Color: colors.White},
			{Position: 1, Color: colors.Blue},
		},
	}
	check(t, repeated.validate() != nil, "equal adjacent positions should fail validation")
}

func TestRadialGradient_Validate(t *testing.T) {
	valid := &RadialGradient{
		X0: 50, Y0: 50, R0: 0,
		X1: 50, Y1: 50, R1: 50,
		Stops: []GradientStop{
			{Position: 0, Color: colors.White},
			{Position: 1, Color: colors.Black},
		},
	}
	check(t, valid.validate() == nil, "valid radial gradient should pass validation")

	negRadius := &RadialGradient{
		X0: 50, Y0: 50, R0: -1,
		X1: 50, Y1: 50, R1: 50,
		Stops: []GradientStop{
			{Position: 0, Color: colors.White},
			{Position: 1, Color: colors.Black},
		},
	}
	check(t, negRadius.validate() != nil, "negative radius should fail validation")

	missingEnd := &RadialGradient{
		X0: 50, Y0: 50, R0: 0,
		X1: 50, Y1: 50, R1: 50,
		Stops: []GradientStop{
			{Position: 0, Color: colors.White},
			{Position: 0.8, Color: colors.Black},
		},
	}
	check(t, missingEnd.validate() != nil, "last stop must be at 1")
}

func TestExponentialFunction_Serialisation(t *testing.T) {
	ef := newExponentialFunction(1, 0, [3]float64{1, 0, 0}, [3]float64{0, 0, 1})
	s := stringFromWriter(ef)
	check(t, strings.Contains(s, "/FunctionType 2 "), "should contain FunctionType 2")
	check(t, strings.Contains(s, "/N 1 "), "should contain N 1")
	check(t, strings.Contains(s, "/C0 [1 0 0 ]"), "should contain C0")
	check(t, strings.Contains(s, "/C1 [0 0 1 ]"), "should contain C1")
	check(t, strings.Contains(s, "/Domain [0 1 ]"), "should contain Domain")
}

func TestStitchingFunction_Serialisation(t *testing.T) {
	ef1 := newExponentialFunction(1, 0, [3]float64{1, 0, 0}, [3]float64{1, 1, 1})
	ef2 := newExponentialFunction(2, 0, [3]float64{1, 1, 1}, [3]float64{0, 0, 1})
	sf := newStitchingFunction(3, 0,
		[]seqGen{ef1, ef2},
		[]float64{0.5},
		[]float64{0, 1, 0, 1})
	s := stringFromWriter(sf)
	check(t, strings.Contains(s, "/FunctionType 3 "), "should contain FunctionType 3")
	check(t, strings.Contains(s, "/Bounds [0.5 ]"), "should contain Bounds")
	check(t, strings.Contains(s, "/Encode [0 1 0 1 ]"), "should contain Encode")
	check(t, strings.Contains(s, "1 0 R "), "should reference first function")
	check(t, strings.Contains(s, "2 0 R "), "should reference second function")
}

func TestAxialShading_Serialisation(t *testing.T) {
	ef := newExponentialFunction(1, 0, [3]float64{1, 0, 0}, [3]float64{0, 0, 1})
	sh := newAxialShading(2, 0, 0, 0, 100, 0, ef)
	s := stringFromWriter(sh)
	check(t, strings.Contains(s, "/ShadingType 2 "), "should contain ShadingType 2")
	check(t, strings.Contains(s, "/ColorSpace /DeviceRGB "), "should contain ColorSpace")
	check(t, strings.Contains(s, "/Coords [0 0 100 0 ]"), "should contain Coords")
	check(t, strings.Contains(s, "/Extend [true true ]"), "should contain Extend")
	check(t, strings.Contains(s, "1 0 R "), "should reference function")
}

func TestRadialShading_Serialisation(t *testing.T) {
	ef := newExponentialFunction(1, 0, [3]float64{1, 1, 0}, [3]float64{0, 0.5, 0})
	sh := newRadialShading(2, 0, 50, 50, 0, 50, 50, 75, ef)
	s := stringFromWriter(sh)
	check(t, strings.Contains(s, "/ShadingType 3 "), "should contain ShadingType 3")
	check(t, strings.Contains(s, "/ColorSpace /DeviceRGB "), "should contain ColorSpace")
	check(t, strings.Contains(s, "/Coords [50 50 0 50 50 75 ]"), "should contain Coords")
	check(t, strings.Contains(s, "/Extend [true true ]"), "should contain Extend")
}

func TestShadingPattern_Serialisation(t *testing.T) {
	ef := newExponentialFunction(1, 0, [3]float64{1, 0, 0}, [3]float64{0, 0, 1})
	sh := newAxialShading(2, 0, 0, 0, 100, 0, ef)
	pat := newShadingPattern(3, 0, sh)
	s := stringFromWriter(pat)
	check(t, strings.Contains(s, "/PatternType 2 "), "should contain PatternType 2")
	check(t, strings.Contains(s, "/Shading 2 0 R "), "should reference shading")
}

func TestNewRealArray(t *testing.T) {
	var buf bytes.Buffer
	a := newRealArray(1.5, 0, 0.25)
	a.write(&buf)
	expectS(t, "[1.5 0 0.25 ] ", buf.String())
}

func TestBuildGradientFunction_TwoStop(t *testing.T) {
	seq := 0
	nextSeq := func() int { seq++; return seq }
	stops := []GradientStop{
		{Position: 0, Color: colors.Red},
		{Position: 1, Color: colors.Blue},
	}
	fn, objs := buildGradientFunction(nextSeq, stops)
	check(t, fn != nil, "function should not be nil")
	expectI(t, 1, len(objs))
	s := stringFromWriter(objs[0])
	check(t, strings.Contains(s, "/FunctionType 2 "), "two-stop should produce exponential function")
}

func TestBuildGradientFunction_MultiStop(t *testing.T) {
	seq := 0
	nextSeq := func() int { seq++; return seq }
	stops := []GradientStop{
		{Position: 0, Color: colors.Red},
		{Position: 0.5, Color: colors.White},
		{Position: 1, Color: colors.Blue},
	}
	fn, objs := buildGradientFunction(nextSeq, stops)
	check(t, fn != nil, "function should not be nil")
	expectI(t, 3, len(objs)) // 2 exponential + 1 stitching
	s := stringFromWriter(objs[2])
	check(t, strings.Contains(s, "/FunctionType 3 "), "multi-stop should produce stitching function")
	check(t, strings.Contains(s, "/Bounds [0.5 ]"), "should have bounds at 0.5")
}
