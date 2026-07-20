// Copyright 2026 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import (
	"errors"
	"fmt"
	"math"

	"github.com/rowland/leadtype/colors"
)

// SweepBandSegment defines one explicit angular segment of a sweep band.
type SweepBandSegment struct {
	StartAngle float64
	EndAngle   float64
	StartColor colors.Color
	EndColor   colors.Color
	Steps      int // Steps <= 0 defaults to 1 equal-angle painted subsegment
}

// SweepBand describes an annular band centered at (X, Y) and divided into
// explicit angular segments. Each segment is painted as a chord-aligned
// linear gradient from StartColor to EndColor. Identical start and end colors
// produce a visually solid segment.
type SweepBand struct {
	X, Y                     float64
	InnerRadius, OuterRadius float64
	Segments                 []SweepBandSegment
	Opacity                  float64 // optional uniform paint opacity; zero means opaque
}

type sweepAngleInterval struct {
	start float64
	end   float64
}

const sweepAngleEpsilon = 1e-9
const sweepPaintOverlapDegrees = 0.2

func (sb *SweepBand) validate() error {
	if sb == nil {
		return errors.New("sweep band must not be nil")
	}
	if sb.InnerRadius < 0 {
		return fmt.Errorf("inner radius %g must be non-negative", sb.InnerRadius)
	}
	if sb.OuterRadius <= sb.InnerRadius {
		return fmt.Errorf("outer radius %g must be greater than inner radius %g", sb.OuterRadius, sb.InnerRadius)
	}
	if len(sb.Segments) == 0 {
		return errors.New("sweep band requires at least one segment")
	}
	for i, seg := range sb.Segments {
		if err := seg.validate(); err != nil {
			return fmt.Errorf("segment %d: %w", i, err)
		}
	}
	for i := range sb.Segments {
		for j := i + 1; j < len(sb.Segments); j++ {
			if sweepSegmentsOverlap(sb.Segments[i], sb.Segments[j]) {
				return fmt.Errorf("segment %d overlaps segment %d", i, j)
			}
		}
	}
	return nil
}

func (seg SweepBandSegment) validate() error {
	span := seg.EndAngle - seg.StartAngle
	if math.Abs(span) <= sweepAngleEpsilon {
		return errors.New("segment must have non-zero angular span")
	}
	if math.Abs(span) > 360+sweepAngleEpsilon {
		return fmt.Errorf("segment span %g exceeds 360 degrees", span)
	}
	return nil
}

func sweepSegmentsOverlap(a, b SweepBandSegment) bool {
	ai := sweepSegmentIntervals(a)
	bi := sweepSegmentIntervals(b)
	for _, ia := range ai {
		for _, ib := range bi {
			if ia.start < ib.end-sweepAngleEpsilon && ib.start < ia.end-sweepAngleEpsilon {
				return true
			}
		}
	}
	return false
}

func sweepSegmentIntervals(seg SweepBandSegment) []sweepAngleInterval {
	span := seg.EndAngle - seg.StartAngle
	if math.Abs(span) >= 360-sweepAngleEpsilon {
		return []sweepAngleInterval{{start: 0, end: 360}}
	}

	var start, end float64
	if span > 0 {
		start = normalizeSweepAngle(seg.StartAngle)
		end = normalizeSweepAngle(seg.EndAngle)
	} else {
		start = normalizeSweepAngle(seg.EndAngle)
		end = normalizeSweepAngle(seg.StartAngle)
	}
	if start < end {
		return []sweepAngleInterval{{start: start, end: end}}
	}
	if start > end {
		return []sweepAngleInterval{
			{start: start, end: 360},
			{start: 0, end: end},
		}
	}
	return []sweepAngleInterval{{start: 0, end: 360}}
}

func normalizeSweepAngle(angle float64) float64 {
	value := math.Mod(angle, 360)
	if value < 0 {
		value += 360
	}
	if math.Abs(value-360) <= sweepAngleEpsilon {
		return 0
	}
	return value
}

func sweepBandGradientAxis(x, y, innerRadius, outerRadius, startAngle, endAngle float64) (Location, Location) {
	rm := (innerRadius + outerRadius) / 2
	return sweepBandPoint(x, y, rm, startAngle), sweepBandPoint(x, y, rm, endAngle)
}

func sweepBandPoint(x, y, radius, angle float64) Location {
	theta := angle * math.Pi / 180.0
	return Location{
		X: x + (radius * math.Cos(theta)),
		Y: y - (radius * math.Sin(theta)),
	}
}

// PaintSweepBand paints an annular band composed of explicit clipped,
// chord-aligned linear-gradient segments.
func (pw *PageWriter) PaintSweepBand(sb *SweepBand) error {
	if err := sb.validate(); err != nil {
		return err
	}
	for _, seg := range expandSweepBandSegments(sb.Segments) {
		if err := pw.paintSweepBandSegment(sb.X, sb.Y, sb.InnerRadius, sb.OuterRadius, sb.Opacity, seg); err != nil {
			return err
		}
	}
	return nil
}

func (seg SweepBandSegment) normalizedSteps() int {
	if seg.Steps <= 0 {
		return 1
	}
	return seg.Steps
}

func (seg SweepBandSegment) expand() []SweepBandSegment {
	steps := seg.normalizedSteps()
	if steps == 1 {
		seg.Steps = 1
		return []SweepBandSegment{seg}
	}

	expanded := make([]SweepBandSegment, 0, steps)
	angleStep := (seg.EndAngle - seg.StartAngle) / float64(steps)
	for i := 0; i < steps; i++ {
		t0 := float64(i) / float64(steps)
		t1 := float64(i+1) / float64(steps)
		expanded = append(expanded, SweepBandSegment{
			StartAngle: seg.StartAngle + (angleStep * float64(i)),
			EndAngle:   seg.StartAngle + (angleStep * float64(i+1)),
			StartColor: interpolateSweepBandColor(seg.StartColor, seg.EndColor, t0),
			EndColor:   interpolateSweepBandColor(seg.StartColor, seg.EndColor, t1),
			Steps:      1,
		})
	}
	return expanded
}

func expandSweepBandSegments(segments []SweepBandSegment) []SweepBandSegment {
	expanded := make([]SweepBandSegment, 0, len(segments))
	for _, seg := range segments {
		expanded = append(expanded, seg.expand()...)
	}
	applySweepBandPaintOverlap(expanded)
	return expanded
}

func applySweepBandPaintOverlap(segments []SweepBandSegment) {
	if len(segments) < 2 {
		return
	}
	for i := 1; i < len(segments); i++ {
		overlapSweepBandBoundary(&segments[i-1], &segments[i])
	}
	overlapSweepBandBoundary(&segments[len(segments)-1], &segments[0])
}

func overlapSweepBandBoundary(prev, curr *SweepBandSegment) {
	if !sweepAnglesTouch(prev.EndAngle, curr.StartAngle) {
		return
	}
	prevSpan := prev.EndAngle - prev.StartAngle
	currSpan := curr.EndAngle - curr.StartAngle
	if math.Abs(prevSpan) <= sweepAngleEpsilon || math.Abs(currSpan) <= sweepAngleEpsilon {
		return
	}
	overlap := math.Min(sweepPaintOverlapDegrees, math.Abs(prevSpan)/4)
	overlap = math.Min(overlap, math.Abs(currSpan)/4)
	if overlap <= 0 {
		return
	}
	prev.EndAngle += (overlap / 2) * math.Copysign(1, prevSpan)
	curr.StartAngle -= (overlap / 2) * math.Copysign(1, currSpan)
}

func sweepAnglesTouch(a, b float64) bool {
	delta := math.Mod(a-b, 360)
	if delta < -180 {
		delta += 360
	} else if delta > 180 {
		delta -= 360
	}
	return math.Abs(delta) <= sweepAngleEpsilon
}

func interpolateSweepBandColor(start, end colors.Color, t float64) colors.Color {
	sr, sg, sb := start.RGB64()
	er, eg, eb := end.RGB64()
	r := interpolateSweepBandChannel(sr, er, t)
	g := interpolateSweepBandChannel(sg, eg, t)
	b := interpolateSweepBandChannel(sb, eb, t)
	return colors.Color(
		(int32(math.Round(r*255)) << 16) |
			(int32(math.Round(g*255)) << 8) |
			int32(math.Round(b*255)),
	)
}

func interpolateSweepBandChannel(start, end, t float64) float64 {
	if t <= 0 {
		return start
	}
	if t >= 1 {
		return end
	}
	return start + ((end - start) * t)
}

func (pw *PageWriter) PaintSweepArc(x, y, innerRadius, outerRadius float64, segments []SweepBandSegment) error {
	return pw.PaintSweepBand(&SweepBand{
		X:           x,
		Y:           y,
		InnerRadius: innerRadius,
		OuterRadius: outerRadius,
		Segments:    segments,
	})
}

func (pw *PageWriter) paintSweepBandSegment(x, y, innerRadius, outerRadius, opacity float64, seg SweepBandSegment) (err error) {
	pw.flushText()
	savedState := pw.drawState
	savedLast := pw.last
	defer func() {
		pw.drawState = savedState
		pw.last = savedLast
	}()

	err = pw.Path(func() {
		if innerRadius == 0 {
			pw.MoveTo(x, y)
			pw.LineTo(x+(outerRadius*math.Cos(seg.StartAngle*math.Pi/180.0)), y-(outerRadius*math.Sin(seg.StartAngle*math.Pi/180.0)))
			err = pw.Arc(x, y, outerRadius, seg.StartAngle, seg.EndAngle, false)
			if err != nil {
				return
			}
			pw.LineTo(x, y)
		} else {
			err = pw.buildSweepBandPath(x, y, innerRadius, outerRadius, seg.StartAngle, seg.EndAngle)
			if err != nil {
				return
			}
		}
		err = pw.Clip(func() {
			p0, p1 := sweepBandGradientAxis(x, y, innerRadius, outerRadius, seg.StartAngle, seg.EndAngle)
			err = pw.PaintLinearGradient(&LinearGradient{
				X0:      p0.X,
				Y0:      p0.Y,
				X1:      p1.X,
				Y1:      p1.Y,
				Opacity: opacity,
				Stops: []GradientStop{
					{Position: 0, Color: seg.StartColor},
					{Position: 1, Color: seg.EndColor},
				},
			})
		})
	})
	return err
}

func (pw *PageWriter) buildSweepBandPath(x, y, innerRadius, outerRadius, startAngle, endAngle float64) error {
	if startAngle == endAngle {
		return errors.New("sweep band path requires non-zero arc segments")
	}
	return pw.annularArcPath(x, y, outerRadius, innerRadius, startAngle, endAngle)
}
