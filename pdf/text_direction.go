// Copyright 2026 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

// TextDirection is the base direction used to lay out bidirectional text.
type TextDirection uint8

const (
	TextDirectionLTR TextDirection = iota
	TextDirectionRTL
)
