// Copyright 2012-2014 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/rowland/leadtype/codepage"
	"github.com/rowland/leadtype/color"
	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/wordbreaking"
)

// Hierarchical structure of text and attributes, allowing text to be described, measured and wrapped.
type RichText struct {
	Text               string
	Font               *Font
	FontSize           float64
	Color              color.Color
	Underline          bool
	LineThrough        bool
	ascent             float64
	descent            float64
	height             float64
	lineGap            float64
	UnderlinePosition  float64
	UnderlineThickness float64
	width              float64
	chars              int
	CharSpacing        float64
	WordSpacing        float64
	NoBreak            bool
	pieces             []*RichText
}

var errNoFontSet = errors.New("No font set")

// NewRichText returns a structure containing text with the specified attributes.
//
// Fonts in the list will be applied in order. That is, the first font will be used for every rune for which it has a glyph.
// Successive fonts will only be used where they have glyphs lacking in previous fonts.
// If NewRichText finds a nil in fonts before finding a font with a glyph for a rune, an error is returned.
// Otherwise, the runes are replaced with question marks in the first font in the list.
//
// Font size is in points.
//
// Options:
//   color:        Fill text with color.
//                 A value of type Color, a string with a color name from NamedColors map, or an RGB color as int, int32 or hexadecimal string.
//   underline:    Draw a line under text.
//                 A bool, a string that evalutes to bool via strconv.ParseBool, a non-zero int or float64.
//   line_through: Draw a line through (or strikethrough) text.
//                 A bool, a string that evalutes to bool via strconv.ParseBool, a non-zero int or float64.
//   char_spacing: Add extra space between characters, expressed in points.
//   word_spacing: Add extra space between words, expressed in points.
//   nobreak:      Prevent WordsToWidth or WrapToWidth from breaking within this stretch of text.
//                 A bool, a string that evalutes to bool via strconv.ParseBool, a non-zero int or float64.
func NewRichText(s string, fonts []*Font, fontSize float64, options options.Options) (*RichText, error) {
	piece := &RichText{
		Text:        s,
		FontSize:    fontSize,
		Color:       options.ColorDefault("color", color.Black),
		Underline:   options.BoolDefault("underline", false),
		LineThrough: options.BoolDefault("line_through", false),
		CharSpacing: options.FloatDefault("char_spacing", 0),
		WordSpacing: options.FloatDefault("word_spacing", 0),
		NoBreak:     options.BoolDefault("nobreak", false),
	}
	var defaultFont *Font
	if len(fonts) == 0 {
		return nil, errNoFontSet
	}
	defaultFont = fonts[0]
	pieces, err := piece.splitByFont(fonts, defaultFont)
	if err != nil {
		return nil, err
	}
	if len(pieces) == 1 {
		return pieces[0], nil
	}
	return &RichText{pieces: pieces}, nil
}

// Add appends text with specified attributes to existing structure, possibly returning a new root.
// Options are the same as for NewRichText.
func (piece *RichText) Add(s string, fonts []*Font, fontSize float64, options options.Options) (*RichText, error) {
	p, err := NewRichText(s, fonts, fontSize, options)
	if err != nil {
		return piece, err
	}
	return piece.AddPiece(p), nil
}

// AddPiece appends one piece of text to another, possibly returning a new root.
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

// Ascent calculates, caches and returns the maximum ascent from the text, expressed in points.
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

// Chars counts, caches and returns the number of discrete characters (not bytes) in the text.
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

func (piece *RichText) Clone() *RichText {
	p := *piece
	p.chars = 0
	p.width = 0.0
	return &p
}

// Count counts and returns the number of text pieces in this structure.
func (piece *RichText) Count() int {
	result := 1
	for _, p := range piece.pieces {
		result += p.Count()
	}
	return result
}

// Descent calculates, caches and returns the maximum descent from the text, expressed in points.
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

func (piece *RichText) EachCodepage(fn func(cpi codepage.CodepageIndex, text string, p *RichText)) {
	piece.VisitAll(func(p *RichText) {
		if p == nil {
			panic("Whoa there!")
		}
		var lastIdx codepage.CodepageIndex = -1
		start := 0
		for i, r := range p.Text {
			if lastIdx >= 0 {
				if _, found := lastIdx.Codepage().CharForCodepoint(r); found {
					continue
				} else {
					if i > start {
						fn(lastIdx, p.Text[start:i], p)
						start = i
					}
					lastIdx, _ = codepage.IndexForCodepoint(r)
				}
			} else {
				if idx, found := codepage.IndexForCodepoint(r); found {
					if i > start {
						fn(lastIdx, p.Text[start:i], p)
						start = i
					}
					lastIdx = idx
				}
			}
		}
		if start < len(p.Text) {
			fn(lastIdx, p.Text[start:], p)
		}
	})
}

// EachRune invokes the callback function for each rune within the text, in sequence.
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

// Height calculates, caches and returns the maximum height of the text, expressed in points.
func (piece *RichText) Height() float64 {
	if piece.height == 0.0 {
		if piece.IsLeaf() {
			return piece.measure().height
		}
		piece.height = piece.Ascent() + -piece.Descent()
	}
	return piece.height
}

// InsertStringAtOffsets returns a new structure with the given string inserted at the specified offsets. This is meant to facilitate dictionary-based hyphenation or similar functionality.
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
				newPiece := piece.Clone()
				newPiece.Text = string(buf)
				res.pieces = append(res.pieces, newPiece)
			}
		}
		*current += n
		return
	}

	newPiece := piece.Clone()
	newPiece.pieces = make([]*RichText, 0, len(piece.pieces))
	res.pieces = append(res.pieces, newPiece)

	for _, p := range piece.pieces {
		p.insertStringAtOffsets(s, offsets, i, current, newPiece)
	}
}

// IsLeaf returns true if this piece of the structure is a leaf node.
func (piece *RichText) IsLeaf() bool {
	return len(piece.pieces) == 0
}

// IsNewLine returns true if this piece contains a single new line character.
func (piece *RichText) IsNewLine() bool {
	return piece.Text == "\n"
}

// IsWhiteSpace returns true if this piece contains whitespace and only whitespace.
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

// Leading calculates and returns the maximum leading of the text, expressed in points.
func (piece *RichText) Leading() float64 {
	return piece.Height() + piece.LineGap()
}

// Len returns the length in bytes of the text in the entire structure.
func (piece *RichText) Len() int {
	result := len(piece.Text)
	for _, p := range piece.pieces {
		result += p.Len()
	}
	return result
}

// LineGap calculates, caches and returns the maximum line gap from the text, expressed in points.
func (piece *RichText) LineGap() float64 {
	// piece.lineGap will frequently be zero, so use piece.chars to detect if the piece has been measured.
	if piece.chars == 0 {
		if piece.IsLeaf() {
			return piece.measure().lineGap
		}
		for _, p := range piece.pieces {
			lineGap := p.LineGap()
			if lineGap > piece.lineGap {
				piece.lineGap = lineGap
			}
		}
	}
	return piece.lineGap
}

// MarkNoBreak sets the NoBreak bits in the previously-allocated flags at every character offset within piece hierarchies
// with NoBreak set to true, except for the offsets corresponding to the initial characters. MarkNoBreak must be called
// before WordsToWidth or WrapToWidth or the NoBreak flags on text pieces will have no effect.
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

// MatchesAttributes returns true if this piece shares the same attributes as the other piece.
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
	piece.chars, piece.width = 0, 0.0
	metrics := piece.Font.metrics
	fsize := piece.FontSize / float64(metrics.UnitsPerEm())
	piece.ascent = float64(metrics.Ascent()) * fsize
	piece.descent = float64(metrics.Descent()) * fsize
	piece.height = float64(piece.Font.Height()) * fsize
	piece.lineGap = float64(metrics.LineGap()) * fsize
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

// Merge returns a new structure where text pieces with matching attributes have been merged together for more efficient printing.
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

	mergedText := piece.Clone()
	mergedText.pieces = make([]*RichText, 0, len(piece.pieces))
	var last *RichText
	for _, p := range flattened {
		if last != nil && p.MatchesAttributes(last) {
			last.Text += p.Text
		} else {
			newPiece := p.Clone()
			mergedText.pieces = append(mergedText.pieces, newPiece)
			last = newPiece
		}
	}
	if len(mergedText.pieces) == 1 {
		return mergedText.pieces[0]
	}
	return mergedText
}

// Split returns the two structures resulting from splitting this text at the specified offset.
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
				newLeft := piece.Clone()
				newLeft.Text = piece.Text[:offset-*current]
				left.pieces = append(left.pieces, newLeft)

				newRight := piece.Clone()
				newRight.Text = piece.Text[offset-*current:]
				right.pieces = append(right.pieces, newRight)
			}
		} else {
			right.pieces = append(right.pieces, piece)
		}
		*current += len(piece.Text)
		return
	}

	newPiece := piece.Clone()
	newPiece.pieces = make([]*RichText, 0, len(piece.pieces))
	if *current < offset {
		left.pieces = append(left.pieces, newPiece)
		left = newPiece
	} else {
		right.pieces = append(right.pieces, newPiece)
		right = newPiece
	}

	for _, p := range piece.pieces {
		p.split(offset, left, right, current)
	}
}

func (piece *RichText) splitByFont(fonts []*Font, defaultFont *Font) (pieces []*RichText, err error) {
	if len(fonts) == 0 {
		newPiece := piece.Clone()
		newPiece.Text = strings.Repeat("?", utf8.RuneCountInString(piece.Text))
		newPiece.Font = defaultFont
		pieces = append(pieces, newPiece)
		return
	}
	font := fonts[0]
	if font == nil {
		return nil, fmt.Errorf("No font found for %s.", piece.Text)
	}
	start := 0
	inFont := false
	for index, rune := range piece.Text {
		runeInFont := (rune == '\t') || (rune == '\n') || (rune == '\r') || (rune == wordbreaking.SoftHyphen) || font.HasRune(rune)
		if runeInFont != inFont {
			if index > start {
				newPiece := piece.Clone()
				newPiece.Text = piece.Text[start:index]
				if inFont {
					newPiece.Font = font
					// newPiece.measure()
					pieces = append(pieces, newPiece)
				} else {
					var newPieces []*RichText
					newPieces, err = newPiece.splitByFont(fonts[1:], defaultFont)
					pieces = append(pieces, newPieces...)
				}
			}
			inFont = runeInFont
			start = index
		}
	}
	if len(piece.Text) > start {
		newPiece := piece.Clone()
		newPiece.Text = piece.Text[start:]
		if inFont {
			newPiece.Font = font
			// newPiece.measure()
			pieces = append(pieces, newPiece)
		} else {
			var newPieces []*RichText
			newPieces, err = newPiece.splitByFont(fonts[1:], defaultFont)
			pieces = append(pieces, newPieces...)
		}
	}
	return
}

// String returns the simple concatenation of the text from each piece in the structure in order.
func (piece *RichText) String() string {
	var buf bytes.Buffer
	piece.VisitAll(func(p *RichText) {
		buf.WriteString(p.Text)
	})
	return buf.String()
}

// TrimFunc trims the runes from both ends of the text where the specified callback function returns true,
// returning a new structure and leaving the original unchanged.
func (piece *RichText) TrimFunc(f func(rune) bool) *RichText {
	return piece.TrimLeftFunc(f).TrimRightFunc(f)
}

// TrimLeftFunc trims the runes from the left end of the text where the specified callback function returns true,
// returning a new structure and leaving the original unchanged.
func (piece *RichText) TrimLeftFunc(f func(rune) bool) *RichText {
	s := piece.String()
	trimmed := strings.TrimLeftFunc(s, f)
	if len(trimmed) == len(s) {
		return piece
	}
	_, right := piece.Split(len(s) - len(trimmed))
	return right
}

// TrimLeftSpace trims the runes from the left end of the text where unicode.IsSpace returns true,
// returning a new structure and leaving the original unchanged.
func (piece *RichText) TrimLeftSpace() *RichText {
	return piece.TrimLeftFunc(unicode.IsSpace)
}

// TrimRightFunc trims the runes from the right end of the text where the specified callback function returns true,
// returning a new structure and leaving the original unchanged.
func (piece *RichText) TrimRightFunc(f func(rune) bool) *RichText {
	s := piece.String()
	trimmed := strings.TrimRightFunc(s, f)
	if len(trimmed) == len(s) {
		return piece
	}
	left, _ := piece.Split(len(trimmed))
	return left
}

// TrimRightSpace trims the runes from the right end of the text where unicode.IsSpace returns true,
// returning a new structure and leaving the original unchanged.
func (piece *RichText) TrimRightSpace() *RichText {
	return piece.TrimRightFunc(unicode.IsSpace)
}

// TrimSpace trims the runes from both ends of the text where unicode.IsSpace returns true,
// returning a new structure and leaving the original unchanged.
func (piece *RichText) TrimSpace() *RichText {
	return piece.TrimFunc(unicode.IsSpace)
}

// VisitAll invokes the callback function for each piece of text within the structure, in order.
func (piece *RichText) VisitAll(fn func(*RichText)) {
	fn(piece)
	for _, p := range piece.pieces {
		p.VisitAll(fn)
	}
}

// Width calculates, caches and returns the width of the text, expressed in points.
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

// WordsToWidth splits the text at the last breaking point before width is exceeded, making use of the previously-allocated and marked flags
// and returning new structures representing both halves and the remaining flags. The original structure is unchanged.
// A minimum of one word is split off, even if width is exceeded, unless hardBreak is true, in which case the text is split along character
// boundaries.
func (piece *RichText) WordsToWidth(
	width float64, wordFlags []wordbreaking.Flags, hardBreak bool) (
	line, remainder *RichText, remainderFlags []wordbreaking.Flags) {
	if width < 0.0 {
		width = 0.0
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
			if lastOffset > 0 {
				current = lastOffset
			} else {
				current = offset
			}
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
		return piece, nil, nil
	}
	line, remainder = piece.Split(current)
	remainderFlags = wordFlags[current:]
	if extra > 0.0 {
		p := piece.lastPiece().Clone()
		p.Text = "-"
		line = line.AddPiece(p)
	}
	return
}

// WrapToWidth returns one or more lines of text resulting from repeatedly invoking WordsToWidth.
func (piece *RichText) WrapToWidth(width float64, wordFlags []wordbreaking.Flags, hardBreak bool) (lines []*RichText) {
	line, remainder, remainderFlags := piece.WordsToWidth(width, wordFlags, hardBreak)
	for remainder != nil {
		lines = append(lines, line.TrimSpace())
		line, remainder, remainderFlags = remainder.WordsToWidth(width, remainderFlags, hardBreak)
	}
	lines = append(lines, line.TrimSpace())
	return
}
