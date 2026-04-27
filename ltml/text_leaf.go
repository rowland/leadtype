package ltml

import (
	"strings"

	"github.com/rowland/leadtype/rich_text"
)

func addNormalizedTextPiece(pieces *[]textPiece, cache **rich_text.RichText, text string, font *FontStyle, normalize func(string) string) {
	if normalize != nil {
		text = normalize(text)
	}
	if text == "" {
		return
	}
	if len(*pieces) == 0 {
		text = strings.TrimLeft(text, " ")
		if text == "" {
			return
		}
	} else if last := &(*pieces)[len(*pieces)-1]; strings.HasSuffix(last.ResolvedText(nil), " ") && strings.HasPrefix(text, " ") {
		text = text[1:]
	}
	*cache = nil
	if len(*pieces) > 0 && (*pieces)[len(*pieces)-1].font == font && !(*pieces)[len(*pieces)-1].Dynamic() {
		lastText := (*pieces)[len(*pieces)-1].ResolvedText(nil)
		(*pieces)[len(*pieces)-1].content = staticInlineText(lastText + text)
		return
	}
	*pieces = append(*pieces, newStaticTextPiece(text, font))
}

func addInlineTextPiece(pieces *[]textPiece, cache **rich_text.RichText, content inlineText, font *FontStyle) {
	*cache = nil
	*pieces = append(*pieces, textPiece{content: content, font: font})
}

func textPiecesHaveDynamicText(pieces []textPiece) bool {
	for _, piece := range pieces {
		if piece.Dynamic() {
			return true
		}
	}
	return false
}

func applyTextPieceFonts(w Writer, container Container, pieces []textPiece, fallback *FontStyle) {
	for _, piece := range pieces {
		applyTextPieceFontForContainer(w, container, piece, fallback)
	}
}

func richTextForTextPieces(w Writer, container Container, pieces []textPiece, cache **rich_text.RichText, fallback *FontStyle) *rich_text.RichText {
	doc := documentForContainer(container)
	if cache != nil && *cache != nil && !textPiecesHaveDynamicText(pieces) {
		applyTextPieceFonts(w, container, pieces, fallback)
		return *cache
	}
	rt := &rich_text.RichText{}
	lastText := ""
	for _, piece := range pieces {
		font := applyTextPieceFontForContainer(w, container, piece, fallback)
		text := piece.ResolvedText(doc)
		if text == "" {
			continue
		}
		if strings.HasSuffix(lastText, " ") && strings.HasPrefix(text, " ") {
			text = text[1:]
		}
		if text == "" {
			continue
		}
		next, err := rt.Add(text, w.Fonts(), w.FontSize(), piece.RichTextOptions(font.RichTextOptions()))
		if err != nil {
			debugf("richTextForTextPieces: %v", err)
			continue
		}
		rt = next
		lastText = text
	}
	if cache != nil && !textPiecesHaveDynamicText(pieces) {
		*cache = rt
	}
	return rt
}

func lineMaxWidth(lines []*rich_text.RichText) float64 {
	width := 0.0
	for _, line := range lines {
		if line != nil && line.Width() > width {
			width = line.Width()
		}
	}
	return width
}

func combineRichText(parts ...*rich_text.RichText) *rich_text.RichText {
	var result *rich_text.RichText
	for _, part := range parts {
		if part == nil || part.Len() == 0 {
			continue
		}
		if result == nil {
			result = part.DeepClone()
			continue
		}
		result = result.AddPiece(part.DeepClone())
	}
	if result == nil {
		return &rich_text.RichText{}
	}
	return result
}
