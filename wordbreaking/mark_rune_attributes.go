// Copyright 2013 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package wordbreaking

import (
	"fmt"
	"unicode"
)

const (
	Hyphen             = 0x2010
	HyphenBullet       = 0x2043
	HyphenMinus        = 0x002D
	LineSeparator      = 0x2028
	NarrowNoBreakSpace = 0x202F
	NoBreakSpace       = 0x00A0
	NonBreakingHyphen  = 0x2011
	ParagraphSeparator = 0x2029
	SoftHyphen         = 0x00AD
	WordJoiner         = 0x2060
	ZeroWidthSpace     = 0x200B
)

type runeClass int

const (
	rcOther      = runeClass(iota)
	rcHyphen     = runeClass(iota)
	rcWhiteSpace = runeClass(iota)
)

// unicode.Hyphen includes non-breaking hyphen
func isHyphen(r rune) bool {
	return r == Hyphen || r == HyphenBullet || r == HyphenMinus || r == SoftHyphen
}

func classifyRune(r rune) runeClass {
	if unicode.IsSpace(r) && r != NoBreakSpace {
		return rcWhiteSpace
	} else if isHyphen(r) {
		return rcHyphen
	}
	return rcOther
}

func MarkRuneAttributes(text string, flags []Flags) {
	if len(flags) < len(text) {
		panic(fmt.Sprintf("flags (len: %d) is smaller than text (len: %d)", len(flags), len(text)))
	}
	var rc, last runeClass
	for i, r := range text {
		flags[i] |= CharStop
		rc = classifyRune(r)
		if i == 0 {
			flags[i] |= SoftBreak | WordStop
		} else if rc == rcWhiteSpace {
			flags[i] |= SoftBreak | WhiteSpace
		} else if last == rcWhiteSpace && (rc == rcHyphen || rc == rcOther) {
			flags[i] |= SoftBreak | WordStop
		} else if last == rcHyphen && rc == rcOther {
			flags[i] |= SoftBreak | WordStop
		}
		last = rc
	}
}
