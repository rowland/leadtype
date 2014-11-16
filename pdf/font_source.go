// Copyright 2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

type FontSource interface {
	Select(family, weight, style string, ranges []string) (font FontMetrics, err error)
}

type FontSources map[string]FontSource
