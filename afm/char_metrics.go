// Copyright 2011-2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package afm

type CharMetric struct {
	Code  rune
	Width int32
	Name  string
}

type CharMetrics []CharMetric

func NewCharMetrics(count int) CharMetrics {
	return make(CharMetrics, count)
}

func (cms CharMetrics) Len() int {
	return len(cms)
}

// Less returns whether the element with index i should sort
// before the element with index j.
func (cms CharMetrics) Less(i, j int) bool {
	return cms[i].Code < cms[j].Code
}

// Swap swaps the elements with indexes i and j.
func (cms CharMetrics) Swap(i, j int) {
	cms[i], cms[j] = cms[j], cms[i]
}

func (cms CharMetrics) ForRune(codepoint rune) *CharMetric {
	low, high := 0, len(cms)-1
	for low <= high {
		i := (low + high) / 2
		if codepoint > cms[i].Code {
			low = i + 1
			continue
		}
		if codepoint < cms[i].Code {
			high = i - 1
			continue
		}
		return &cms[i]
	}
	return nil
}
