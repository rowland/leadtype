package ltml

import (
	"strings"
	"testing"

	"github.com/rowland/leadtype/pdf"
)

func TestResolveSectorSweepBandMapsStopsToSectorGeometry(t *testing.T) {
	opacity := 0.6
	brush := &BrushStyle{
		kind: BrushKindSweepGradient,
		sweepGradient: &sweepGradientStyle{
			Stops: []pdf.GradientStop{
				{Position: 0, Color: NamedColor("Red")},
				{Position: 0.25, Color: NamedColor("Gold")},
				{Position: 1, Color: NamedColor("Blue")},
			},
			Steps: 4,
		},
		opacity: &opacity,
	}
	geometry := radialSectorGeometry{
		CenterX: 42, CenterY: 57,
		InnerRadius: 10, OuterRadius: 30,
		StartAngle: 30, EndAngle: 150,
	}

	band, err := resolveSectorSweepBand(brush, geometry)
	if err != nil {
		t.Fatal(err)
	}
	if band.X != 42 || band.Y != 57 || band.InnerRadius != 10 || band.OuterRadius != 30 {
		t.Fatalf("band geometry = %#v, want sector geometry", band)
	}
	if band.Opacity != 0.6 {
		t.Fatalf("opacity = %v, want 0.6", band.Opacity)
	}
	if len(band.Segments) != 2 {
		t.Fatalf("segments = %d, want 2", len(band.Segments))
	}
	if band.Segments[0].StartAngle != 30 || band.Segments[0].EndAngle != 60 {
		t.Fatalf("segment 0 angles = %v..%v, want 30..60", band.Segments[0].StartAngle, band.Segments[0].EndAngle)
	}
	if band.Segments[1].StartAngle != 60 || band.Segments[1].EndAngle != 150 {
		t.Fatalf("segment 1 angles = %v..%v, want 60..150", band.Segments[1].StartAngle, band.Segments[1].EndAngle)
	}
	for i, segment := range band.Segments {
		if segment.Steps != 4 {
			t.Fatalf("segment %d steps = %d, want 4", i, segment.Steps)
		}
	}
}

func TestResolveSectorSweepBandSupportsClockwiseAndFullCircleSpans(t *testing.T) {
	brush := &BrushStyle{
		kind: BrushKindSweepGradient,
		sweepGradient: &sweepGradientStyle{
			Stops: []pdf.GradientStop{
				{Position: 0, Color: NamedColor("Red")},
				{Position: 0.5, Color: NamedColor("Green")},
				{Position: 1, Color: NamedColor("Blue")},
			},
		},
	}
	tests := []struct {
		name       string
		start, end float64
		wantAngles [][2]float64
	}{
		{
			name:       "clockwise",
			start:      180,
			end:        -180,
			wantAngles: [][2]float64{{180, 0}, {0, -180}},
		},
		{
			name:       "counterclockwise",
			start:      10,
			end:        370,
			wantAngles: [][2]float64{{10, 190}, {190, 370}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			band, err := resolveSectorSweepBand(brush, radialSectorGeometry{
				CenterX: 1, CenterY: 2,
				InnerRadius: 0, OuterRadius: 20,
				StartAngle: tt.start, EndAngle: tt.end,
			})
			if err != nil {
				t.Fatal(err)
			}
			if band.InnerRadius != 0 {
				t.Fatalf("inner radius = %v, want 0", band.InnerRadius)
			}
			for i, want := range tt.wantAngles {
				got := band.Segments[i]
				if got.StartAngle != want[0] || got.EndAngle != want[1] {
					t.Fatalf("segment %d angles = %v..%v, want %v..%v", i, got.StartAngle, got.EndAngle, want[0], want[1])
				}
			}
		})
	}
}

func TestStdSectorPaintBackgroundPaintsSweepBand(t *testing.T) {
	sector := &StdSector{}
	sector.fill = &BrushStyle{
		kind: BrushKindSweepGradient,
		sweepGradient: &sweepGradientStyle{
			Stops: []pdf.GradientStop{
				{Position: 0, Color: NamedColor("Tomato")},
				{Position: 1, Color: NamedColor("Gold")},
			},
			Steps: 6,
		},
	}
	sector.geometry = radialSectorGeometry{
		CenterX: 100, CenterY: 120,
		InnerRadius: 35, OuterRadius: 55,
		StartAngle: -45, EndAngle: 45,
	}
	writer := &labelTestWriter{}

	if err := sector.PaintBackground(writer); err != nil {
		t.Fatal(err)
	}
	if len(writer.sweepPaints) != 1 {
		t.Fatalf("sweep paints = %d, want 1", len(writer.sweepPaints))
	}
	band := writer.sweepPaints[0]
	if band.InnerRadius != 35 || band.OuterRadius != 55 {
		t.Fatalf("band radii = %v..%v, want 35..55", band.InnerRadius, band.OuterRadius)
	}
	if band.Segments[0].StartAngle != -45 || band.Segments[0].EndAngle != 45 {
		t.Fatalf("band angles = %v..%v, want -45..45", band.Segments[0].StartAngle, band.Segments[0].EndAngle)
	}
}

func TestStdSectorPaintBackgroundSweepOpacityZeroSkipsPaint(t *testing.T) {
	opacity := 0.0
	sector := &StdSector{}
	sector.fill = &BrushStyle{
		kind: BrushKindSweepGradient,
		sweepGradient: &sweepGradientStyle{
			Stops: []pdf.GradientStop{
				{Position: 0, Color: NamedColor("Red")},
				{Position: 1, Color: NamedColor("Blue")},
			},
		},
		opacity: &opacity,
	}
	sector.geometry = radialSectorGeometry{OuterRadius: 10, EndAngle: 90}
	writer := &labelTestWriter{}

	if err := sector.PaintBackground(writer); err != nil {
		t.Fatal(err)
	}
	if len(writer.sweepPaints) != 0 {
		t.Fatalf("sweep paints = %d, want 0", len(writer.sweepPaints))
	}
}

func TestResolveSectorSweepBandRejectsInvalidStops(t *testing.T) {
	brush := &BrushStyle{
		kind: BrushKindSweepGradient,
		sweepGradient: &sweepGradientStyle{
			Stops: []pdf.GradientStop{
				{Position: 0.2, Color: NamedColor("Red")},
				{Position: 1, Color: NamedColor("Blue")},
			},
		},
	}
	_, err := resolveSectorSweepBand(brush, radialSectorGeometry{OuterRadius: 10, EndAngle: 90})
	if err == nil || !strings.Contains(err.Error(), "first stop position must be 0") {
		t.Fatalf("error = %v, want invalid first-stop error", err)
	}
}
