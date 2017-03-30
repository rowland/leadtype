// Copyright 2017 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

type SizeSpecification int

const (
	Unspecified = SizeSpecification(iota)
	Specified
	Percent
)

type SpecifiedSize struct {
	How  SizeSpecification
	Size float64
}

type SpecifiedSizes []*SpecifiedSize

func (sss SpecifiedSizes) Partition(pf func(*SpecifiedSize) bool) (trueSS, falseSS SpecifiedSizes) {
	trueSS, falseSS = make(SpecifiedSizes, 0, len(sss)), make(SpecifiedSizes, 0, len(sss))
	for _, ss := range sss {
		if pf(ss) {
			trueSS = append(trueSS, ss)
		} else {
			falseSS = append(falseSS, ss)
		}
	}
	return
}
