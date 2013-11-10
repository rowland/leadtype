// Copyright 2013 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package wordbreaking

const (
	SoftBreak  = 1 << iota // Potential linebreak point
	WhiteSpace = 1 << iota // A unicode whitespace character, except NBSP, ZWNBSP
	CharStop   = 1 << iota
	WordStop   = 1 << iota
	Invalid    = 1 << iota // Invalid character sequence
)

type Flags byte
