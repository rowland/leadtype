// Copyright 2026 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import (
	"testing"

	"github.com/rowland/leadtype/colors"
)

func TestSweepBand_Validate(t *testing.T) {
	valid := &SweepBand{
		X: 10, Y: 10,
		InnerRadius: 2,
		OuterRadius: 4,
		Segments: []SweepBandSegment{
			{StartAngle: 0, EndAngle: 45, StartColor: colors.Red, EndColor: colors.Red},
			{StartAngle: 45, EndAngle: 90, StartColor: colors.Red, EndColor: colors.Yellow},
		},
	}
	check(t, valid.validate() == nil, "valid sweep band should pass validation")

	check(t, (*SweepBand)(nil).validate() != nil, "nil sweep band should fail validation")
	check(t, (&SweepBand{InnerRadius: -1, OuterRadius: 2, Segments: []SweepBandSegment{{StartAngle: 0, EndAngle: 45, StartColor: colors.Red, EndColor: colors.Red}}}).validate() != nil, "negative inner radius should fail")
	check(t, (&SweepBand{InnerRadius: 2, OuterRadius: 2, Segments: []SweepBandSegment{{StartAngle: 0, EndAngle: 45, StartColor: colors.Red, EndColor: colors.Red}}}).validate() != nil, "outer radius must exceed inner radius")
	check(t, (&SweepBand{InnerRadius: 0, OuterRadius: 2}).validate() != nil, "empty segment list should fail")
	check(t, (&SweepBand{InnerRadius: 0, OuterRadius: 2, Segments: []SweepBandSegment{{StartAngle: 10, EndAngle: 10, StartColor: colors.Red, EndColor: colors.Red}}}).validate() != nil, "zero-span segment should fail")
	check(t, (&SweepBand{InnerRadius: 0, OuterRadius: 2, Segments: []SweepBandSegment{{StartAngle: 0, EndAngle: 400, StartColor: colors.Red, EndColor: colors.Red}}}).validate() != nil, "spans over 360 should fail")
	check(t, (&SweepBand{
		InnerRadius: 1,
		OuterRadius: 2,
		Segments: []SweepBandSegment{
			{StartAngle: 0, EndAngle: 90, StartColor: colors.Red, EndColor: colors.Red},
			{StartAngle: 45, EndAngle: 135, StartColor: colors.Red, EndColor: colors.Blue},
		},
	}).validate() != nil, "overlapping segments should fail")
}

func TestSweepBandGradientAxis_Directions(t *testing.T) {
	p0, p1 := sweepBandGradientAxis(10, 10, 2, 4, 45, 135)
	expectFdelta(t, p0.Y, p1.Y, 0.0001)

	p0, p1 = sweepBandGradientAxis(10, 10, 2, 4, -45, 45)
	expectFdelta(t, p0.X, p1.X, 0.0001)

	p0, p1 = sweepBandGradientAxis(10, 10, 2, 4, 0, 45)
	check(t, p0.X != p1.X, "diagonal axis should change x")
	check(t, p0.Y != p1.Y, "diagonal axis should change y")
}

func TestSweepBandConstantColorSegment_AllowsEqualEndpoints(t *testing.T) {
	sb := &SweepBand{
		X: 10, Y: 10,
		InnerRadius: 1,
		OuterRadius: 2,
		Segments: []SweepBandSegment{
			{StartAngle: 0, EndAngle: 45, StartColor: colors.Red, EndColor: colors.Red},
		},
	}
	check(t, sb.validate() == nil, "equal endpoint colors should produce a valid constant-color segment")
}

func TestSweepBandSegment_NormalizedSteps(t *testing.T) {
	expectI(t, 1, (SweepBandSegment{}).normalizedSteps())
	expectI(t, 1, (SweepBandSegment{Steps: -3}).normalizedSteps())
	expectI(t, 1, (SweepBandSegment{Steps: 0}).normalizedSteps())
	expectI(t, 8, (SweepBandSegment{Steps: 8}).normalizedSteps())
}

func TestSweepBandSegment_Expand(t *testing.T) {
	seg := SweepBandSegment{
		StartAngle: 0,
		EndAngle:   90,
		StartColor: colors.Red,
		EndColor:   colors.Blue,
		Steps:      3,
	}
	expanded := seg.expand()
	expectI(t, 3, len(expanded))
	expectFdelta(t, 0, expanded[0].StartAngle, 0.0001)
	expectFdelta(t, 30, expanded[0].EndAngle, 0.0001)
	expectFdelta(t, 30, expanded[1].StartAngle, 0.0001)
	expectFdelta(t, 60, expanded[1].EndAngle, 0.0001)
	expectFdelta(t, 60, expanded[2].StartAngle, 0.0001)
	expectFdelta(t, 90, expanded[2].EndAngle, 0.0001)
	expectV(t, colors.Red, expanded[0].StartColor)
	expectV(t, colors.Blue, expanded[2].EndColor)
}

func TestSweepBandSegment_ExpandConstantColor(t *testing.T) {
	seg := SweepBandSegment{
		StartAngle: 0,
		EndAngle:   45,
		StartColor: colors.Green,
		EndColor:   colors.Green,
		Steps:      4,
	}
	expanded := seg.expand()
	expectI(t, 4, len(expanded))
	for _, sub := range expanded {
		expectV(t, colors.Green, sub.StartColor)
		expectV(t, colors.Green, sub.EndColor)
	}
}

func TestExpandSweepBandSegments_OverlapsContiguousBoundaries(t *testing.T) {
	expanded := expandSweepBandSegments([]SweepBandSegment{
		{StartAngle: 0, EndAngle: 45, StartColor: colors.Red, EndColor: colors.Red},
		{StartAngle: 45, EndAngle: 90, StartColor: colors.Red, EndColor: colors.Blue},
	})
	expectI(t, 2, len(expanded))
	check(t, expanded[0].EndAngle > 45, "first segment should slightly overlap a touching neighbor")
	check(t, expanded[1].StartAngle < 45, "second segment should slightly overlap a touching neighbor")
}

func TestExpandSweepBandSegments_DoesNotExpandAcrossGaps(t *testing.T) {
	expanded := expandSweepBandSegments([]SweepBandSegment{
		{StartAngle: 0, EndAngle: 45, StartColor: colors.Red, EndColor: colors.Red},
		{StartAngle: 50, EndAngle: 90, StartColor: colors.Red, EndColor: colors.Blue},
	})
	expectI(t, 2, len(expanded))
	expectFdelta(t, 45, expanded[0].EndAngle, 0.0001)
	expectFdelta(t, 50, expanded[1].StartAngle, 0.0001)
}

func TestExpandSweepBandSegments_OverlapsWrappedBoundary(t *testing.T) {
	expanded := expandSweepBandSegments([]SweepBandSegment{
		{StartAngle: 315, EndAngle: 360, StartColor: colors.Blue, EndColor: colors.Red},
		{StartAngle: 0, EndAngle: 45, StartColor: colors.Red, EndColor: colors.Yellow},
	})
	expectI(t, 2, len(expanded))
	check(t, expanded[0].EndAngle > 360, "wrapped segment should overlap past 360 degrees")
	check(t, expanded[1].StartAngle < 0, "wrapped neighbor should overlap below 0 degrees")
}

func TestInterpolateSweepBandColor(t *testing.T) {
	expectV(t, colors.Red, interpolateSweepBandColor(colors.Red, colors.Blue, 0))
	expectV(t, colors.Blue, interpolateSweepBandColor(colors.Red, colors.Blue, 1))
	expectV(t, colors.Color(0x800080), interpolateSweepBandColor(colors.Red, colors.Blue, 0.5))
}
