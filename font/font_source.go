// Copyright 2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package font

type FontSource interface {
	Select(family, weight, style string, ranges []string) (font FontMetrics, err error)
	SubType() string
}

// ClosestFontSource is an optional extension consulted only when the caller
// explicitly requests nearest-face matching. Implementations must stay within
// family, preserve range restrictions, and return a real face rather than
// synthesizing bold or italic. FontSource remains unchanged so AFM and custom
// sources need not implement approximate matching.
type ClosestFontSource interface {
	SelectClosest(family, weight, style string, ranges []string) (font FontMetrics, err error)
}

type FontSources []FontSource
