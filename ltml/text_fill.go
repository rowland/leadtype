package ltml

import (
	"math"

	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/rich_text"
)

// paragraphAlignedLine prepares one paragraph line for clip-based text fill
// using the same alignment rules as pdf.PageWriter.PrintParagraph.
//
// This intentionally duplicates just the alignment/justification portion of
// PrintParagraph because the LTML text-fill path needs the positioned line
// without asking PageWriter to emit visible text. For clipped paragraph fill we
// must:
//  1. compute the same x offset and justification spacing as normal paragraph
//     rendering,
//  2. position the line,
//  3. clip the line's text, and then
//  4. paint the paragraph brush through that stencil.
//
// PageWriter.PrintParagraph currently combines paragraph layout decisions with
// visible text emission and cursor advancement, so the clip-based path cannot
// reuse it directly. Keeping this helper narrowly scoped avoids reimplementing
// the rest of paragraph rendering while preserving layout parity.
func paragraphAlignedLine(line *rich_text.RichText, width float64, align HAlign) (float64, *rich_text.RichText) {
	if line == nil {
		return 0, nil
	}
	switch align {
	case HAlignCenter:
		return (width - line.Width()) / 2, line
	case HAlignRight:
		return width - line.Width(), line
	case HAlignJustify:
		delta := width - line.Width()
		spaces := 0
		for _, r := range line.String() {
			if r == rune(32) {
				spaces++
			}
		}
		words := spaces + 1
		if width > 0 && math.Abs(delta)/width < 0.4 {
			var charSpacing, wordSpacing float64
			if words == 1 {
				wordSpacing = 0
				charSpacing = delta / float64(line.Chars()-1)
			} else if math.Abs(delta)/float64(words) > 3 {
				wordSpacing = 3
				delta -= float64(words-1) * wordSpacing
				charSpacing = delta / float64(line.Chars()-1)
			} else {
				wordSpacing = delta / float64(words-1)
				charSpacing = 0
			}
			line = line.DeepClone()
			line.VisitAll(func(p *rich_text.RichText) {
				p.CharSpacing = charSpacing
				p.WordSpacing = wordSpacing
			})
		}
	}
	return 0, line
}

func paragraphTextFillOptions(p *StdParagraph) options.Options {
	return options.Options{
		"text-align": p.ParagraphStyle().ResolvedTextAlign(p).String(),
		"width":      ContentWidth(p) - p.textIndent(),
	}
}
