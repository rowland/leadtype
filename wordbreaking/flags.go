// Copyright 2013 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package wordbreaking

type Flags byte

const (
	SoftBreak  = Flags(1 << iota) // potential linebreak point
	WhiteSpace = Flags(1 << iota) // a unicode whitespace character, except NBSP
	CharStop   = Flags(1 << iota) // valid cursor position
	WordStop   = Flags(1 << iota) // start of a word
	Invalid    = Flags(1 << iota) // invalid character sequence
)
