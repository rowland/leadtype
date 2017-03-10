// Copyright 2017 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package haru

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type linePattern struct {
	pattern []int
	phase   int
}

var reLinePattern = regexp.MustCompile(`\[\s*((\d+(\s+)?)*)\s*\]\s*(\d+)`)

func newLinePatternFromString(s string) *linePattern {
	var lp = linePattern{[]int{}, 0}
	if matches := reLinePattern.FindStringSubmatch(s); matches != nil {
		for _, p := range strings.Split(matches[1], " ") {
			if i, err := strconv.Atoi(p); err == nil {
				lp.pattern = append(lp.pattern, i)
			}
		}
		lp.phase, _ = strconv.Atoi(matches[4])
	}
	return &lp
}

func (pat *linePattern) String() string {
	return fmt.Sprintf("[%s] %d", intSlice(pat.pattern).join(" "), pat.phase)
}

type LinePatternMap map[string]*linePattern

func (lpm LinePatternMap) Add(name string, pattern []int, phase int) {
	lpm[name] = &linePattern{pattern, phase}
}

var LinePatterns = LinePatternMap{
	"solid":  &linePattern{[]int{}, 0},
	"dotted": &linePattern{[]int{1, 2}, 0},
	"dashed": &linePattern{[]int{4, 2}, 0},
}
