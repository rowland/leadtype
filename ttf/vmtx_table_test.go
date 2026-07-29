// Copyright 2011-2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ttf

import (
	"bytes"
	"testing"
)

func TestVmtxTableInitBulkDecodesMetricsAndBearings(t *testing.T) {
	data := []byte{
		0x03, 0x20, 0xff, 0xec, // advance 800, bearing -20
		0x03, 0x84, 0x00, 0x1e, // advance 900, bearing 30
		0xff, 0xd8, // bearing -40
	}
	source := append([]byte{0xaa, 0xbb}, data...)
	var table vmtxTable
	if err := table.init(bytes.NewReader(source), &tableDirEntry{offset: 2, length: uint32(len(data))}, 3, 2); err != nil {
		t.Fatal(err)
	}

	want := []longVerMetric{
		{advanceHeight: 800, topSideBearing: -20},
		{advanceHeight: 900, topSideBearing: 30},
	}
	for i := range want {
		if table.vMetrics[i] != want[i] {
			t.Fatalf("metric %d = %#v, want %#v", i, table.vMetrics[i], want[i])
		}
	}
	if got := table.topSideBearing; len(got) != 1 || got[0] != -40 {
		t.Fatalf("top side bearings = %v, want [-40]", got)
	}
}

func TestVmtxTableInitRejectsMalformedLengthsAndCounts(t *testing.T) {
	tests := []struct {
		name       string
		data       []byte
		tableLen   uint32
		numGlyphs  uint16
		numMetrics uint16
	}{
		{name: "metrics exceed glyphs", tableLen: 8, numGlyphs: 1, numMetrics: 2},
		{name: "declared table too short", data: make([]byte, 10), tableLen: 9, numGlyphs: 3, numMetrics: 2},
		{name: "source data truncated", data: make([]byte, 9), tableLen: 10, numGlyphs: 3, numMetrics: 2},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var table vmtxTable
			if err := table.init(bytes.NewReader(tc.data), &tableDirEntry{length: tc.tableLen}, tc.numGlyphs, tc.numMetrics); err == nil {
				t.Fatal("init succeeded, want malformed table error")
			}
		})
	}
}
