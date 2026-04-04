// Copyright 2024 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import (
	"errors"
	"fmt"

	"github.com/rowland/leadtype/colors"
)

// GradientStop defines a colour at a position along the gradient axis.
// Position must be in the range [0,1].
type GradientStop struct {
	Position float64
	Color    colors.Color
}

// LinearGradient describes an axial (Type 2) gradient between two points
// in user-space coordinates.
type LinearGradient struct {
	X0, Y0 float64        // start point
	X1, Y1 float64        // end point
	Stops  []GradientStop // at least two stops required
}

func (lg *LinearGradient) validate() error {
	return validateStops(lg.Stops)
}

// RadialGradient describes a radial (Type 3) gradient between two circles
// in user-space coordinates.
type RadialGradient struct {
	X0, Y0, R0 float64        // start circle centre and radius
	X1, Y1, R1 float64        // end circle centre and radius
	Stops      []GradientStop // at least two stops required
}

func (rg *RadialGradient) validate() error {
	if rg.R0 < 0 || rg.R1 < 0 {
		return errors.New("gradient radii must be non-negative")
	}
	return validateStops(rg.Stops)
}

func validateStops(stops []GradientStop) error {
	if len(stops) < 2 {
		return errors.New("gradient requires at least two stops")
	}
	for i, s := range stops {
		if s.Position < 0 || s.Position > 1 {
			return fmt.Errorf("stop %d position %g out of range [0,1]", i, s.Position)
		}
		if i > 0 && s.Position < stops[i-1].Position {
			return fmt.Errorf("stop %d position %g precedes stop %d position %g", i, s.Position, i-1, stops[i-1].Position)
		}
	}
	return nil
}

// exponentialFunction is a PDF Type 2 (exponential interpolation) function
// that maps t in [0,1] to a linear interpolation from C0 to C1 with N=1.
type exponentialFunction struct {
	dictionaryObject
}

func newExponentialFunction(seq, gen int, c0, c1 [3]float64) *exponentialFunction {
	f := new(exponentialFunction)
	f.dictionaryObject.init(seq, gen)
	f.dict["FunctionType"] = integer(2)
	f.dict["Domain"] = newRealArray(0, 1)
	f.dict["C0"] = newRealArray(c0[0], c0[1], c0[2])
	f.dict["C1"] = newRealArray(c1[0], c1[1], c1[2])
	f.dict["N"] = integer(1)
	return f
}

// stitchingFunction is a PDF Type 3 (stitching) function that chains
// multiple sub-functions across adjacent domains.
type stitchingFunction struct {
	dictionaryObject
}

func newStitchingFunction(seq, gen int, fns []seqGen, bounds, encode []float64) *stitchingFunction {
	f := new(stitchingFunction)
	f.dictionaryObject.init(seq, gen)
	f.dict["FunctionType"] = integer(3)
	f.dict["Domain"] = newRealArray(0, 1)
	refs := make(array, len(fns))
	for i, fn := range fns {
		refs[i] = &indirectObjectRef{fn}
	}
	f.dict["Functions"] = refs
	f.dict["Bounds"] = newRealArray(bounds...)
	f.dict["Encode"] = newRealArray(encode...)
	return f
}

// shadingDict is a PDF shading dictionary.
// Type 2 (axial) for linear gradients, Type 3 (radial) for radial gradients.
type shadingDict struct {
	dictionaryObject
}

func newAxialShading(seq, gen int, x0, y0, x1, y1 float64, fn seqGen) *shadingDict {
	s := new(shadingDict)
	s.dictionaryObject.init(seq, gen)
	s.dict["ShadingType"] = integer(2)
	s.dict["ColorSpace"] = name("DeviceRGB")
	s.dict["Coords"] = newRealArray(x0, y0, x1, y1)
	s.dict["Function"] = &indirectObjectRef{fn}
	s.dict["Extend"] = array{boolean(true), boolean(true)}
	return s
}

func newRadialShading(seq, gen int, x0, y0, r0, x1, y1, r1 float64, fn seqGen) *shadingDict {
	s := new(shadingDict)
	s.dictionaryObject.init(seq, gen)
	s.dict["ShadingType"] = integer(3)
	s.dict["ColorSpace"] = name("DeviceRGB")
	s.dict["Coords"] = newRealArray(x0, y0, r0, x1, y1, r1)
	s.dict["Function"] = &indirectObjectRef{fn}
	s.dict["Extend"] = array{boolean(true), boolean(true)}
	return s
}

// shadingPattern is a PDF Type 2 (shading) pattern that wraps a shading
// dictionary so it can be used as a fill or stroke colour via the Pattern
// colour space.
type shadingPattern struct {
	dictionaryObject
}

func newShadingPattern(seq, gen int, shading seqGen) *shadingPattern {
	p := new(shadingPattern)
	p.dictionaryObject.init(seq, gen)
	p.dict["PatternType"] = integer(2)
	p.dict["Shading"] = &indirectObjectRef{shading}
	return p
}

// newRealArray builds a PDF array of real values.
func newRealArray(values ...float64) array {
	a := make(array, len(values))
	for i, v := range values {
		a[i] = real(v)
	}
	return a
}

// colorToRGB64 returns an [3]float64 with the RGB components of c.
func colorToRGB64(c colors.Color) [3]float64 {
	r, g, b := c.RGB64()
	return [3]float64{r, g, b}
}

// buildGradientFunction creates the interpolation function(s) for a set of
// gradient stops. For two stops it returns a single exponential function.
// For three or more stops it returns a stitching function referencing
// per-segment exponential functions. All created objects are returned so
// the caller can add them to the PDF body.
func buildGradientFunction(nextSeq func() int, stops []GradientStop) (fn seqGen, allObjects []genWriter) {
	if len(stops) == 2 {
		ef := newExponentialFunction(nextSeq(), 0,
			colorToRGB64(stops[0].Color),
			colorToRGB64(stops[1].Color))
		return ef, []genWriter{ef}
	}

	n := len(stops) - 1
	fns := make([]seqGen, n)
	objs := make([]genWriter, 0, n+1)
	bounds := make([]float64, n-1)
	encode := make([]float64, 2*n)
	for i := 0; i < n; i++ {
		ef := newExponentialFunction(nextSeq(), 0,
			colorToRGB64(stops[i].Color),
			colorToRGB64(stops[i+1].Color))
		fns[i] = ef
		objs = append(objs, ef)
		encode[2*i] = 0
		encode[2*i+1] = 1
		if i < n-1 {
			bounds[i] = stops[i+1].Position
		}
	}
	sf := newStitchingFunction(nextSeq(), 0, fns, bounds, encode)
	objs = append(objs, sf)
	return sf, objs
}
