// Copyright 2024 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package wordbreaking

import (
	_ "embed"
	"strings"
	"sync"
)

//go:embed words_th.txt
var thaiWordsData string

var (
	thaiDict     map[string]struct{}
	thaiDictOnce sync.Once
)

func getThaiDict() map[string]struct{} {
	thaiDictOnce.Do(func() {
		lines := strings.Split(thaiWordsData, "\n")
		thaiDict = make(map[string]struct{}, len(lines))
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			// Only include entries composed entirely of Thai characters.
			allThai := true
			for _, r := range line {
				if !isThai(r) {
					allThai = false
					break
				}
			}
			if allThai {
				thaiDict[line] = struct{}{}
			}
		}
	})
	return thaiDict
}

func isThai(r rune) bool {
	return r >= 0x0E00 && r <= 0x0E7F
}

// SegmentThai inserts zero-width spaces between Thai words in text,
// enabling the standard word-breaking algorithm to wrap Thai text correctly.
// Non-Thai portions of the string are passed through unchanged.
// It should be called before MarkRuneAttributes.
func SegmentThai(text string) string {
	hasThai := false
	for _, r := range text {
		if isThai(r) {
			hasThai = true
			break
		}
	}
	if !hasThai {
		return text
	}

	dict := getThaiDict()
	runes := []rune(text)
	var buf strings.Builder
	buf.Grow(len(text))

	i := 0
	for i < len(runes) {
		if !isThai(runes[i]) {
			buf.WriteRune(runes[i])
			i++
			continue
		}
		// Collect a run of consecutive Thai characters.
		start := i
		for i < len(runes) && isThai(runes[i]) {
			i++
		}
		run := runes[start:i]

		// Apply forward maximal matching within the Thai run.
		pos := 0
		first := true
		for pos < len(run) {
			// Find the longest dictionary word starting at pos.
			end := len(run)
			matched := false
			for end > pos {
				if _, ok := dict[string(run[pos:end])]; ok {
					matched = true
					break
				}
				end--
			}
			if !matched {
				end = pos + 1 // single-character fallback
			}
			if !first {
				buf.WriteRune(ZeroWidthSpace)
			}
			buf.WriteString(string(run[pos:end]))
			pos = end
			first = false
		}
	}
	return buf.String()
}
