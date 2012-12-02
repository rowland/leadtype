// Copyright 2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

type TextLine struct {
	RichText
	ascent  float64
	chars   int
	descent float64
	height  float64
	tokens  int
	width   float64
}

func (line *TextLine) Ascent() float64 {
	if line.ascent == 0.0 {
		for _, p := range line.RichText {
			a := p.Ascent()
			if a > line.ascent {
				line.ascent = a
			}
		}
	}
	return line.ascent
}

func (line *TextLine) Chars() int {
	if line.chars == 0 {
		for _, p := range line.RichText {
			line.chars += p.Chars
		}
	}
	return line.chars
}

func (line *TextLine) Descent() float64 {
	if line.descent == 0.0 {
		for _, p := range line.RichText {
			d := p.Descent()
			if d < line.descent {
				line.descent = d
			}
		}
	}
	return line.descent
}

func (line *TextLine) Height() float64 {
	if line.height == 0.0 {
		for _, p := range line.RichText {
			h := p.Height()
			if h > line.height {
				line.height = h
			}
		}
	}
	return line.height
}

func (line *TextLine) Tokens() int {
	if line.tokens == 0 {
		for _, p := range line.RichText {
			line.tokens += p.Tokens
		}
	}
	return line.tokens
}

func (line *TextLine) Width() float64 {
	if line.width == 0.0 {
		for _, p := range line.RichText {
			line.width += p.Width
		}
	}
	return line.width
}
