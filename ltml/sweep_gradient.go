// Copyright 2026 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"

	"github.com/rowland/leadtype/pdf"
)

func resolveSectorSweepBand(brush *BrushStyle, geometry radialSectorGeometry) (*pdf.SweepBand, error) {
	if brush == nil || brush.sweepGradient == nil {
		return nil, nil
	}
	opacity := normalizeBrushOpacityValue(brush.opacity)
	if opacity <= 0 {
		return nil, nil
	}
	stops := brush.sweepGradient.Stops
	if err := validateSweepGradientStops(stops); err != nil {
		return nil, err
	}
	steps := brush.sweepGradient.Steps
	if steps <= 0 {
		steps = 1
	}
	span := geometry.EndAngle - geometry.StartAngle
	segments := make([]pdf.SweepBandSegment, 0, len(stops)-1)
	for i := 0; i < len(stops)-1; i++ {
		start, end := stops[i], stops[i+1]
		segments = append(segments, pdf.SweepBandSegment{
			StartAngle: geometry.StartAngle + span*start.Position,
			EndAngle:   geometry.StartAngle + span*end.Position,
			StartColor: start.Color,
			EndColor:   end.Color,
			Steps:      steps,
		})
	}
	return &pdf.SweepBand{
		X:           geometry.CenterX,
		Y:           geometry.CenterY,
		InnerRadius: geometry.InnerRadius,
		OuterRadius: geometry.OuterRadius,
		Segments:    segments,
		Opacity:     opacity,
	}, nil
}

func validateSweepGradientStops(stops []pdf.GradientStop) error {
	if len(stops) < 2 {
		return fmt.Errorf("sweep gradient requires at least two stops")
	}
	for i, stop := range stops {
		if stop.Position < 0 || stop.Position > 1 {
			return fmt.Errorf("sweep gradient stop %d position %g out of range [0,1]", i, stop.Position)
		}
		if i > 0 && stop.Position <= stops[i-1].Position {
			return fmt.Errorf("sweep gradient stop %d position %g must be greater than stop %d position %g", i, stop.Position, i-1, stops[i-1].Position)
		}
	}
	if stops[0].Position != 0 {
		return fmt.Errorf("sweep gradient first stop position must be 0, got %g", stops[0].Position)
	}
	if stops[len(stops)-1].Position != 1 {
		return fmt.Errorf("sweep gradient last stop position must be 1, got %g", stops[len(stops)-1].Position)
	}
	return nil
}
