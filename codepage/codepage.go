// Copyright 2011-2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package codepage

type CharRange struct {
	firstCode  rune
	lastCode   rune
	entryCount int
	delta      int
}

type Codepage []CharRange

func (list Codepage) CharForCodepoint(cp rune) (ch rune, found bool) {
	low, high := 0, len(list)-1
	for low <= high {
		i := (low + high) / 2
		r := &list[i]
		if cp < r.firstCode {
			high = i - 1
			continue
		}
		if cp > r.lastCode {
			low = i + 1
			continue
		}
		return cp + rune(r.delta), true
	}
	return 0, false
}

type CodepageRange struct {
	firstCode  rune
	lastCode   rune
	entryCount int
	codepage   int
}

type CodepageRanges []CodepageRange

func (list CodepageRanges) CodepageForCodepoint(cp rune) (codepage int, found bool) {
	low, high := 0, len(list)-1
	for low <= high {
		i := (low + high) / 2
		r := &list[i]
		if cp < r.firstCode {
			high = i - 1
			continue
		}
		if cp > r.lastCode {
			low = i + 1
			continue
		}
		return r.codepage, true
	}
	return
}
