// Copyright 2012, 2013, 2014 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import (
	"bytes"
	"fmt"
	"leadtype/wordbreaking"
	"sort"
	"strings"
	"unicode"
)

type RichText struct {
	Text               string
	Font               *Font
	FontSize           float64
	Color              Color
	Underline          bool
	LineThrough        bool
	ascent             float64
	descent            float64
	height             float64
	UnderlinePosition  float64
	UnderlineThickness float64
	width              float64
	chars              int
	CharSpacing        float64
	WordSpacing        float64
	NoBreak            bool
	pieces             []*RichText
}

func NewRichText(s string, fonts []*Font, fontSize float64, options Options) (*RichText, error) {
	piece := &RichText{
		Text:        s,
		FontSize:    fontSize,
		Color:       options.ColorDefault("color", Black),
		Underline:   options.BoolDefault("underline", false),
		LineThrough: options.BoolDefault("line_through", false),
		CharSpacing: options.FloatDefault("char_spacing", 0),
		WordSpacing: options.FloatDefault("word_spacing", 0),
		NoBreak:     options.BoolDefault("nobreak", false),
	}
	pieces, err := piece.splitByFont(fonts)
	if err != nil {
		return nil, err
	}
	if len(pieces) == 1 {
		return pieces[0], nil
	}
	return &RichText{pieces: pieces}, nil
}

func (piece *RichText) Add(s string, fonts []*Font, fontSize float64, options Options) (*RichText, error) {
	p, err := NewRichText(s, fonts, fontSize, options)
	if err != nil {
		return piece, err
	}
	return piece.AddPiece(p), nil
}

func (piece *RichText) AddPiece(p *RichText) *RichText {
	if len(piece.pieces) > 0 {
		piece.pieces = append(piece.pieces, p)
		piece.chars = 0
		piece.width = 0.0
	} else {
		piece = &RichText{pieces: []*RichText{piece, p}}
	}
	return piece
}

func (piece *RichText) Ascent() float64 {
	if piece.ascent == 0.0 {
		if piece.IsLeaf() {
			return piece.measure().ascent
		}
		for _, p := range piece.pieces {
			ascent := p.Ascent()
			if ascent > piece.ascent {
				piece.ascent = ascent
			}
		}
	}
	return piece.ascent
}

func (piece *RichText) Chars() int {
	if piece.chars == 0 {
		if piece.IsLeaf() {
			return piece.measure().chars
		}
		for _, p := range piece.pieces {
			piece.chars += p.Chars()
		}
	}
	return piece.chars
}

func (piece *RichText) clone() *RichText {
	p := *piece
	p.chars = 0
	p.width = 0.0
	return &p
}

func (piece *RichText) Count() int {
	result := 1
	for _, p := range piece.pieces {
		result += p.Count()
	}
	return result
}

func (piece *RichText) Descent() float64 {
	if piece.descent == 0.0 {
		if piece.IsLeaf() {
			return piece.measure().descent
		}
		for _, p := range piece.pieces {
			descent := p.Descent()
			if descent < piece.descent {
				piece.descent = descent
			}
		}
	}
	return piece.descent
}

func (piece *RichText) EachRune(fn func(r rune, p *RichText, offset int) bool) bool {
	offset := 0
	return piece.eachRune(fn, &offset)
}

func (piece *RichText) eachRune(fn func(rune rune, p *RichText, offset int) bool, offset *int) bool {
	for o, rune := range piece.Text {
		if fn(rune, piece, *offset+o) {
			return true
		}
	}
	*offset += len(piece.Text)
	for _, p := range piece.pieces {
		if p.eachRune(fn, offset) {
			return true
		}
	}
	return false
}

func (piece *RichText) Height() float64 {
	if piece.height == 0.0 {
		if piece.IsLeaf() {
			return piece.measure().descent
		}
		for _, p := range piece.pieces {
			height := p.Height()
			if height > piece.height {
				piece.height = height
			}
		}
	}
	return piece.height
}

func (piece *RichText) InsertStringAtOffsets(s string, offsets []int) *RichText {
	sort.Ints(offsets)
	current := 0
	i := 0
	res := new(RichText)
	piece.insertStringAtOffsets(s, offsets, &i, &current, res)
	return res
}

func (piece *RichText) insertStringAtOffsets(s string, offsets []int, i *int, current *int, res *RichText) {
	if piece.IsLeaf() {
		n := len(piece.Text)
		if *i < len(offsets) && *current < offsets[*i] {
			if offsets[*i]-*current >= n {
				res.pieces = append(res.pieces, piece)
			} else {
				j := *i
				for j < len(offsets) && offsets[j] < *current+n {
					j++
				}
				buf := make([]byte, n+len(s)*(j-*i))
				src, dst := 0, 0
				for ; *i < j; *i++ {
					d := copy(buf[dst:], piece.Text[src:offsets[*i]-*current])
					copy(buf[dst+d:], s)
					src += d
					dst += d + len(s)
				}
				copy(buf[dst:], piece.Text[src:])
				newPiece := *piece
				newPiece.Text = string(buf)
				res.pieces = append(res.pieces, &newPiece)
			}
		}
		*current += n
		return
	}

	newPiece := *piece
	newPiece.pieces = make([]*RichText, 0, len(piece.pieces))
	res.pieces = append(res.pieces, &newPiece)

	for _, p := range piece.pieces {
		p.insertStringAtOffsets(s, offsets, i, current, &newPiece)
	}
}

func (piece *RichText) IsLeaf() bool {
	return len(piece.pieces) == 0
}

func (piece *RichText) IsNewLine() bool {
	return piece.Text == "\n"
}

func (piece *RichText) IsWhiteSpace() bool {
	for _, rune := range piece.Text {
		if !unicode.IsSpace(rune) {
			return false
		}
	}
	return len(piece.Text) > 0
}

func (piece *RichText) lastPiece() *RichText {
	for !piece.IsLeaf() {
		piece = piece.pieces[len(piece.pieces)-1]
	}
	return piece
}

func (piece *RichText) Len() int {
	result := len(piece.Text)
	for _, p := range piece.pieces {
		result += p.Len()
	}
	return result
}

func (piece *RichText) MarkNoBreak(wordFlags []wordbreaking.Flags) {
	offset := 0
	piece.VisitAll(func(p *RichText) {
		if p.NoBreak {
			l := p.Len()
			for i := offset + 1; i < offset+l; i++ {
				wordFlags[i] |= wordbreaking.NoBreak
			}
		}
		offset += len(p.Text)
	})
}

func (piece *RichText) MatchesAttributes(other *RichText) bool {
	return (piece.Font != nil && other.Font != nil) && (piece.Font == other.Font || piece.Font.Matches(other.Font)) &&
		piece.FontSize == other.FontSize &&
		piece.Color == other.Color &&
		piece.Underline == other.Underline &&
		piece.LineThrough == other.LineThrough &&
		piece.CharSpacing == other.CharSpacing &&
		piece.WordSpacing == other.WordSpacing
}

func (piece *RichText) measure() *RichText {
	if piece.Font == nil {
		return piece
	}
	piece.chars, piece.width = 0, 0
	metrics := piece.Font.metrics
	fsize := piece.FontSize / float64(metrics.UnitsPerEm())
	piece.ascent = float64(metrics.Ascent()) * fsize
	piece.descent = float64(metrics.Descent()) * fsize
	piece.height = float64(piece.Font.Height()) * fsize
	piece.UnderlinePosition = float64(metrics.UnderlinePosition()) * fsize
	piece.UnderlineThickness = float64(metrics.UnderlineThickness()) * fsize
	for _, rune := range piece.Text {
		if rune == wordbreaking.SoftHyphen {
			continue
		}
		piece.chars += 1
		runeWidth, _ := metrics.AdvanceWidth(rune)
		piece.width += (fsize * float64(runeWidth)) + piece.CharSpacing
		if unicode.IsSpace(rune) {
			piece.width += piece.WordSpacing
		}
	}
	return piece
}

func (piece *RichText) Merge() *RichText {
	if len(piece.pieces) == 0 {
		return piece
	}

	flattened := make([]*RichText, 0, len(piece.pieces))
	for _, p := range piece.pieces {
		pm := p.Merge()
		if pm.IsLeaf() || pm.NoBreak {
			flattened = append(flattened, pm)
		} else {
			flattened = append(flattened, pm.pieces...)
		}
	}

	mergedText := *piece
	mergedText.pieces = make([]*RichText, 0, len(piece.pieces))
	var last *RichText
	for _, p := range flattened {
		if last != nil && p.MatchesAttributes(last) {
			last.Text += p.Text
		} else {
			newPiece := *p
			mergedText.pieces = append(mergedText.pieces, &newPiece)
			last = &newPiece
		}
	}
	if len(mergedText.pieces) == 1 {
		return mergedText.pieces[0]
	}
	return &mergedText
}

func (piece *RichText) Split(offset int) (left, right *RichText) {
	if offset < 0 {
		offset = 0
	}
	var current int = 0
	left, right = new(RichText), new(RichText)
	piece.split(offset, left, right, &current)
	for len(left.pieces) == 1 {
		left = left.pieces[0]
	}
	for len(right.pieces) == 1 {
		right = right.pieces[0]
	}
	return
}

func (piece *RichText) split(offset int, left, right *RichText, current *int) {
	if piece.IsLeaf() {
		if *current < offset {
			if (offset - *current) >= len(piece.Text) {
				left.pieces = append(left.pieces, piece)
			} else {
				newLeft := *piece
				newLeft.Text = piece.Text[:offset-*current]
				newLeft.width, newLeft.chars = 0, 0
				left.pieces = append(left.pieces, &newLeft)

				newRight := *piece
				newRight.Text = piece.Text[offset-*current:]
				newRight.width, newRight.chars = 0, 0
				right.pieces = append(right.pieces, &newRight)
			}
		} else {
			right.pieces = append(right.pieces, piece)
		}
		*current += len(piece.Text)
		return
	}

	newPiece := *piece
	newPiece.pieces = make([]*RichText, 0, len(piece.pieces))
	if *current < offset {
		left.pieces = append(left.pieces, &newPiece)
		left = &newPiece
	} else {
		right.pieces = append(right.pieces, &newPiece)
		right = &newPiece
	}

	for _, p := range piece.pieces {
		p.split(offset, left, right, current)
	}
}

func (piece *RichText) splitByFont(fonts []*Font) (pieces []*RichText, err error) {
	if len(fonts) == 0 {
		return nil, fmt.Errorf("No font found for %s.", piece.Text)
	}
	font := fonts[0]
	start := 0
	inFont := false
	for index, rune := range piece.Text {
		runeInFont := (rune == '\t') || (rune == '\n') || (rune == wordbreaking.SoftHyphen) || font.HasRune(rune)
		if runeInFont != inFont {
			if index > start {
				newPiece := *piece
				newPiece.Text = piece.Text[start:index]
				if inFont {
					newPiece.Font = font
					newPiece.measure()
					pieces = append(pieces, &newPiece)
				} else {
					var newPieces []*RichText
					newPieces, err = newPiece.splitByFont(fonts[1:])
					pieces = append(pieces, newPieces...)
				}
			}
			inFont = runeInFont
			start = index
		}
	}
	if len(piece.Text) > start {
		newPiece := *piece
		newPiece.Text = piece.Text[start:]
		if inFont {
			newPiece.Font = font
			newPiece.measure()
			pieces = append(pieces, &newPiece)
		} else {
			var newPieces []*RichText
			newPieces, err = newPiece.splitByFont(fonts[1:])
			pieces = append(pieces, newPieces...)
		}
	}
	return
}

func (piece *RichText) String() string {
	var buf bytes.Buffer
	piece.VisitAll(func(p *RichText) {
		buf.WriteString(p.Text)
	})
	return buf.String()
}

func (piece *RichText) TrimFunc(f func(rune) bool) *RichText {
	return piece.TrimLeftFunc(f).TrimRightFunc(f)
}

func (piece *RichText) TrimLeftFunc(f func(rune) bool) *RichText {
	s := piece.String()
	trimmed := strings.TrimLeftFunc(s, f)
	if len(trimmed) == len(s) {
		return piece
	}
	_, right := piece.Split(len(s) - len(trimmed))
	return right
}

func (piece *RichText) TrimLeftSpace() *RichText {
	return piece.TrimLeftFunc(unicode.IsSpace)
}

func (piece *RichText) TrimRightFunc(f func(rune) bool) *RichText {
	s := piece.String()
	trimmed := strings.TrimRightFunc(s, f)
	if len(trimmed) == len(s) {
		return piece
	}
	left, _ := piece.Split(len(trimmed))
	return left
}

func (piece *RichText) TrimRightSpace() *RichText {
	return piece.TrimRightFunc(unicode.IsSpace)
}

func (piece *RichText) TrimSpace() *RichText {
	return piece.TrimFunc(unicode.IsSpace)
}

func (piece *RichText) VisitAll(fn func(*RichText)) {
	fn(piece)
	for _, p := range piece.pieces {
		p.VisitAll(fn)
	}
}

func (piece *RichText) Width() float64 {
	if piece.width == 0.0 {
		if piece.IsLeaf() {
			return piece.measure().width
		}
		for _, p := range piece.pieces {
			piece.width += p.Width()
		}
	}
	return piece.width
}

func (piece *RichText) WordsToWidth(
	width float64, wordFlags []wordbreaking.Flags, hardBreak bool) (
	line, remainder *RichText, lineFlags, remainderFlags []wordbreaking.Flags) {
	if width < 0 {
		width = 0
	}
	current := 0
	currentWidth := 0.0
	words := 0
	wordWidth := 0.0
	extra := 0.0
	lastOffset := 0

	var lastPiece *RichText
	var metrics FontMetrics
	var fsize float64
	var lastRune rune

	fn := func(rune rune, p *RichText, offset int) bool {
		if words > 0 && currentWidth+extra+wordWidth > width {
			return true
		} else if words == 0 && hardBreak && wordWidth > width {
			current = lastOffset
			return true
		}
		if offset > 0 &&
			wordFlags[offset]&wordbreaking.SoftBreak == wordbreaking.SoftBreak &&
			wordFlags[offset]&wordbreaking.NoBreak != wordbreaking.NoBreak {
			current = offset
			currentWidth += wordWidth
			wordWidth = 0.0
			words++
			if lastRune == wordbreaking.SoftHyphen {
				extraRuneWidth, _ := metrics.AdvanceWidth(wordbreaking.HyphenMinus)
				extra = (fsize * float64(extraRuneWidth)) + p.CharSpacing
			} else {
				extra = 0.0
			}
		}
		if p != lastPiece {
			metrics = p.Font.metrics
			fsize = p.FontSize / float64(metrics.UnitsPerEm())
			lastPiece = p
		}
		if rune != wordbreaking.SoftHyphen {
			runeWidth, _ := metrics.AdvanceWidth(rune)
			wordWidth += (fsize * float64(runeWidth)) + p.CharSpacing
			if unicode.IsSpace(rune) {
				wordWidth += p.WordSpacing
			}
		}
		lastRune = rune
		lastOffset = offset
		return false
	}

	finished := !piece.EachRune(fn)
	if current == 0 || finished {
		return piece, nil, wordFlags, nil
	}
	line, remainder = piece.Split(current)
	lineFlags, remainderFlags = wordFlags[:current], wordFlags[current:]
	if extra > 0.0 {
		p := piece.lastPiece().clone()
		p.Text = "-"
		line = line.AddPiece(p)
	}
	return
}

func (piece *RichText) WrapToWidth(width float64, wordFlags []wordbreaking.Flags, hardBreak bool) (lines []*RichText) {
	line, remainder, _, remainderFlags := piece.WordsToWidth(width, wordFlags, hardBreak)
	for remainder != nil {
		lines = append(lines, line.TrimSpace())
		line, remainder, _, remainderFlags = remainder.WordsToWidth(width, remainderFlags, hardBreak)
	}
	lines = append(lines, line.TrimSpace())
	return
}
