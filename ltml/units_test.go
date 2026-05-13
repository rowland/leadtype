// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"math"
	"testing"
)

func TestParseMeasurement_dp(t *testing.T) {
	got := ParseMeasurement("200dp", "pt")
	want := 200 * (72.0 / 1000)
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("ParseMeasurement(200dp) = %v, want %v", got, want)
	}
	if want != 0.2*72 {
		t.Fatalf("200dp should equal 0.2in in points")
	}
}

func TestParseMeasurement_mm_matchesPageCatalog(t *testing.T) {
	got := ParseMeasurement("210mm", "pt")
	want := mmToPoints(210)
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("ParseMeasurement(210mm) = %v, want mmToPoints(210)=%v", got, want)
	}
}

func TestFromUnits_bareNumberUsesDefaultUnits(t *testing.T) {
	if g := ParseMeasurement("1000", "dp"); math.Abs(g-72) > 1e-9 {
		t.Fatalf(`ParseMeasurement("1000", dp) = %v, want 72 (one inch)`, g)
	}
}

func TestParseMeasurement_cm_matchesTenMm(t *testing.T) {
	cm := ParseMeasurement("2.5cm", "pt")
	mm := ParseMeasurement("25mm", "pt")
	if math.Abs(cm-mm) > 1e-9 {
		t.Fatalf("2.5cm = %v, 25mm = %v", cm, mm)
	}
}
