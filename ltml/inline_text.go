package ltml

import (
	"fmt"

	"github.com/rowland/leadtype/options"
)

type inlineText interface {
	Resolve(*StdDocument) string
	Dynamic() bool
}

type inlineTextWithFont interface {
	inlineText
	Font() *FontStyle
}

type inlineTextWithLink interface {
	inlineText
	LinkURI() string
	LinkTarget() string
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

func (p textPiece) Font(fallback *FontStyle) *FontStyle {
	if p.font != nil {
		return p.font
	}
	if content, ok := p.content.(inlineTextWithFont); ok {
		if font := content.Font(); font != nil {
			return font
		}
	}
	return fallback
}

func (p textPiece) RichTextOptions(base options.Options) options.Options {
	if base == nil {
		base = options.Options{}
	}
	linker, ok := p.content.(inlineTextWithLink)
	if !ok {
		return base
	}
	if linker.LinkURI() == "" && linker.LinkTarget() == "" {
		return base
	}
	opts := make(options.Options, len(base)+2)
	for k, v := range base {
		opts[k] = v
	}
	if uri := linker.LinkURI(); uri != "" {
		opts["link_uri"] = uri
	}
	if target := linker.LinkTarget(); target != "" {
		opts["link_target"] = target
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
