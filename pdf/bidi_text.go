// Copyright 2026 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import (
	"github.com/rowland/leadtype/rich_text"
	xbidi "golang.org/x/text/unicode/bidi"
)

type leafSlice struct {
	piece     *rich_text.RichText
	leafStart int
	leafEnd   int
	clipStart int
	clipEnd   int
}

func bidiDisplayPieces(text *rich_text.RichText, baseDirection xbidi.Direction) []*rich_text.RichText {
	merged := text.Merge()
	logical := merged.String()
	if logical == "" {
		return nil
	}

	leaves := flattenLeafSlices(merged)
	if len(leaves) == 0 {
		return leafPieces(leaves)
	}

	var paragraph xbidi.Paragraph
	if _, err := paragraph.SetString(logical, xbidi.DefaultDirection(baseDirection)); err != nil {
		return leafPieces(leaves)
	}
	order, err := paragraph.Order()
	if err != nil || order.NumRuns() == 0 {
		return leafPieces(leaves)
	}

	runIndexes := make([]int, 0, order.NumRuns())
	if baseDirection == xbidi.RightToLeft {
		for i := order.NumRuns() - 1; i >= 0; i-- {
			runIndexes = append(runIndexes, i)
		}
	} else {
		for i := 0; i < order.NumRuns(); i++ {
			runIndexes = append(runIndexes, i)
		}
	}

	out := make([]*rich_text.RichText, 0, len(leaves))
	for _, runIndex := range runIndexes {
		run := order.Run(runIndex)
		start, end := run.Pos()
		end++
		segments := intersectingLeafSlices(leaves, start, end)
		if run.Direction() == xbidi.RightToLeft {
			for i, j := 0, len(segments)-1; i < j; i, j = i+1, j-1 {
				segments[i], segments[j] = segments[j], segments[i]
			}
		}
		out = append(out, slicePieces(segments)...)
	}
	if len(out) == 0 {
		return leafPieces(leaves)
	}
	return out
}

func flattenLeafSlices(text *rich_text.RichText) []leafSlice {
	slices := []leafSlice{}
	runePos := 0
	text.VisitAll(func(p *rich_text.RichText) {
		if !p.IsLeaf() || p.Text == "" || p.Font == nil {
			return
		}
		runeLen := len([]rune(p.Text))
		slices = append(slices, leafSlice{
			piece:     p,
			leafStart: runePos,
			leafEnd:   runePos + runeLen,
			clipStart: runePos,
			clipEnd:   runePos + runeLen,
		})
		runePos += runeLen
	})
	return slices
}

func leafPieces(slices []leafSlice) []*rich_text.RichText {
	out := make([]*rich_text.RichText, 0, len(slices))
	for _, s := range slices {
		out = append(out, s.piece)
	}
	return out
}

func intersectingLeafSlices(slices []leafSlice, start, end int) []leafSlice {
	out := make([]leafSlice, 0, len(slices))
	for _, s := range slices {
		if s.leafEnd <= start || s.leafStart >= end {
			continue
		}
		out = append(out, leafSlice{
			piece:     s.piece,
			leafStart: s.leafStart,
			leafEnd:   s.leafEnd,
			clipStart: max(s.leafStart, start),
			clipEnd:   min(s.leafEnd, end),
		})
	}
	return out
}

func slicePieces(slices []leafSlice) []*rich_text.RichText {
	out := make([]*rich_text.RichText, 0, len(slices))
	for _, s := range slices {
		text := sliceRunes(s.piece.Text, s.clipStart-s.leafStart, s.clipEnd-s.leafStart)
		if text == "" {
			continue
		}
		if text == s.piece.Text {
			out = append(out, s.piece)
			continue
		}
		clone := s.piece.Clone()
		clone.Text = text
		out = append(out, clone)
	}
	return out
}

func sliceRunes(s string, start, end int) string {
	if start == 0 && end >= len([]rune(s)) {
		return s
	}
	runes := []rune(s)
	if start < 0 {
		start = 0
	}
	if end > len(runes) {
		end = len(runes)
	}
	if start >= end {
		return ""
	}
	return string(runes[start:end])
}

func (pw *PageWriter) bidiBaseDirection(text string) xbidi.Direction {
	if pw.textDirectionSet {
		if pw.textDirection == TextDirectionRTL {
			return xbidi.RightToLeft
		}
		return xbidi.LeftToRight
	}
	return firstStrongDirection(text)
}

func firstStrongDirection(text string) xbidi.Direction {
	var paragraph xbidi.Paragraph
	if _, err := paragraph.SetString(text); err == nil {
		if order, err := paragraph.Order(); err == nil && order.NumRuns() > 0 {
			if dir := paragraph.Direction(); dir == xbidi.LeftToRight || dir == xbidi.RightToLeft {
				return dir
			}
		}
	}
	return xbidi.LeftToRight
}
