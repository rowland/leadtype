// Copyright 2017 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package haru

import "fmt"

type linePattern struct {
	pattern []float64
	phase   int
}

func (pat *linePattern) String() string {
	return fmt.Sprintf("[%s] %d", float64Slice(pat.pattern).join(" "), pat.phase)
}

type LinePatternMap map[string]*linePattern

func (lpm LinePatternMap) Add(name string, pattern []float64, phase int) {
	lpm[name] = &linePattern{pattern, phase}
}

var LinePatterns = LinePatternMap{
	"solid":  &linePattern{[]float64{}, 0},
	"dotted": &linePattern{[]float64{1, 2}, 0},
	"dashed": &linePattern{[]float64{4, 2}, 0},
}
