// Copyright 2013 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package wordbreaking

import (
	"fmt"
	"unicode"

	"github.com/go-text/typesetting/segmenter"
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
	rcCJK        = runeClass(iota)
)

// unicode.Hyphen includes non-breaking hyphen
func isHyphen(r rune) bool {
	return r == Hyphen || r == HyphenBullet || r == HyphenMinus || r == SoftHyphen
}

func classifyRune(r rune) runeClass {
	if r == ZeroWidthSpace {
		return rcWhiteSpace
	} else if unicode.IsSpace(r) && r != NoBreakSpace {
		return rcWhiteSpace
	} else if isHyphen(r) {
		return rcHyphen
	} else if isCJKBreakRune(r) {
		return rcCJK
	}
	return rcOther
}

func isCJKBreakRune(r rune) bool {
	return unicode.In(r,
		unicode.Han,
		unicode.Hiragana,
		unicode.Katakana,
		unicode.Bopomofo,
	)
}

// MarkRuneAttributes augments flags with UTF-8 byte-indexed text attributes.
// Default line opportunities follow Unicode 17 UAX #14; Thai dictionary
// boundaries are added as a language-specific tailoring.
func MarkRuneAttributes(text string, flags []Flags) {
	if len(flags) < len(text) {
		panic(fmt.Sprintf("flags (len: %d) is smaller than text (len: %d)", len(flags), len(text)))
	}
	if text == "" {
		return
	}

	runes := make([]rune, 0, len(text))
	byteOffsets := make([]int, 0, len(text)+1)
	thaiBreaks := thaiBreakOffsets(text)
	var rc, last runeClass
	for i, r := range text {
		runes = append(runes, r)
		byteOffsets = append(byteOffsets, i)
		flags[i] |= CharStop
		rc = classifyRune(r)
		if i == 0 {
			flags[i] |= WordStop
		} else if _, ok := thaiBreaks[i]; ok {
			flags[i] |= WordStop
		} else if rc == rcWhiteSpace {
			flags[i] |= WhiteSpace
		} else if last == rcWhiteSpace && (rc == rcHyphen || rc == rcOther || rc == rcCJK) {
			flags[i] |= WordStop
		} else if last == rcHyphen && (rc == rcOther || rc == rcCJK) {
			flags[i] |= WordStop
		} else if last == rcCJK && rc == rcCJK {
			flags[i] |= WordStop
		} else if (last == rcOther && rc == rcCJK) || (last == rcCJK && rc == rcOther) {
			flags[i] |= WordStop
		}
		last = rc
	}
	byteOffsets = append(byteOffsets, len(text))

	var seg segmenter.Segmenter
	seg.Init(runes)
	iter := seg.LineIterator()
	for iter.Next() {
		line := iter.Line()
		runeOffset := line.Offset + len(line.Text)
		if runeOffset >= len(runes) {
			continue
		}
		byteOffset := byteOffsets[runeOffset]
		flags[byteOffset] |= SoftBreak
		if line.IsMandatoryBreak {
			flags[byteOffset] |= MandatoryBreak
		}
	}

	for byteOffset := range thaiBreaks {
		flags[byteOffset] |= SoftBreak | WordStop
	}
}
