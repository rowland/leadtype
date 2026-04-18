package ltml

import (
	"fmt"
	"strings"

	"github.com/rowland/leadtype/options"
)

type inlineText interface {
	Resolve(*StdDocument) string
	Dynamic() bool
}

type inlineTextWithFont interface {
	inlineText
	HasFont
}

type inlineTextWithLink interface {
	inlineText
	LinkURI() string
	LinkTarget() string
	LinkID() string
}

type staticInlineText string

func (t staticInlineText) Resolve(*StdDocument) string { return string(t) }
func (t staticInlineText) Dynamic() bool               { return false }

type textPiece struct {
	content inlineText
	font    *FontStyle
}

func newStaticTextPiece(text string, font *FontStyle) textPiece {
	return textPiece{content: staticInlineText(text), font: font}
}

func (p textPiece) ResolvedText(doc *StdDocument) string {
	if p.content == nil {
		return ""
	}
	return p.content.Resolve(doc)
}

func (p textPiece) Dynamic() bool {
	return p.content != nil && p.content.Dynamic()
}

func (p textPiece) Font(fallback *FontStyle) (*FontStyle, bool) {
	if p.font != nil {
		return p.font, true
	}
	if content, ok := p.content.(inlineTextWithFont); ok {
		if font := content.Font(); font != nil {
			return font, true
		}
	}
	return fallback, false
}

func (p textPiece) RichTextOptions(base options.Options) options.Options {
	if base == nil {
		base = options.Options{}
	}
	opts := make(options.Options, len(base)+3)
	for k, v := range base {
		opts[k] = v
	}
	if linker, ok := p.content.(inlineTextWithLink); ok {
		if uri := linker.LinkURI(); uri != "" {
			opts["link_uri"] = uri
		}
		if target := linker.LinkTarget(); target != "" {
			opts["link_target"] = target
		}
		if id := linker.LinkID(); id != "" {
			opts["link_id"] = id
		}
	}
	if len(opts) == 0 {
		return base
	}
	return opts
}

type AddTextWithFonter interface {
	AddTextWithFont(text string, font *FontStyle)
}

type AddInlineWithFonter interface {
	AddInlineWithFont(content inlineText, font *FontStyle)
}

type InlineContainer interface {
	Container
	AddInlineWithFonter
}

func documentForContainer(c Container) *StdDocument {
	for c != nil {
		switch value := c.(type) {
		case *StdDocument:
			return value
		case *StdPage:
			return value.document()
		}
		c = c.Container()
	}
	return nil
}

func walkWidgets(root Widget, fn func(Widget) bool) bool {
	if root == nil {
		return false
	}
	if !fn(root) {
		return false
	}
	container, ok := root.(Container)
	if !ok {
		return true
	}
	for _, child := range container.Widgets() {
		if !walkWidgets(child, fn) {
			return false
		}
	}
	return true
}

func formatPageNo(value int) string {
	return fmt.Sprintf("%d", value)
}

func resolvedTextPieces(pieces []textPiece, doc *StdDocument) string {
	if len(pieces) == 0 {
		return ""
	}
	var b strings.Builder
	lastText := ""
	for _, piece := range pieces {
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
		b.WriteString(text)
		lastText = text
	}
	return b.String()
}
