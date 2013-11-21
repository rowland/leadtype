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
			if p.ascent > line.ascent {
				line.ascent = p.ascent
			}
		}
	}
	return line.ascent
}

func (line *TextLine) Chars() int {
	if line.chars == 0 {
		for _, p := range line.RichText {
			line.chars += p.chars
		}
	}
	return line.chars
}

func (line *TextLine) Descent() float64 {
	if line.descent == 0.0 {
		for _, p := range line.RichText {
			if p.descent < line.descent {
				line.descent = p.descent
			}
		}
	}
	return line.descent
}

func (line *TextLine) Height() float64 {
	if line.height == 0.0 {
		for _, p := range line.RichText {
			if p.height > line.height {
				line.height = p.height
			}
		}
	}
	return line.height
}

func (line *TextLine) Width() float64 {
	if line.width == 0.0 {
		for _, p := range line.RichText {
			line.width += p.width
		}
	}
	return line.width
}
