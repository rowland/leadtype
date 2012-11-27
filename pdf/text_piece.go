// Copyright 2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

type TextPiece struct {
	Text        string
	Font        *Font
	FontSize    float64
	Color       Color
	Underline   bool
	LineThrough bool
	Width       int
	Chars       int
	Tokens      int
}
