package ltml

import (
	"fmt"
	"strings"
	"unicode"
)

type StdA struct {
	StdSpan
	uri    string
	target string
}

// linkedInlineText decorates normal inline text with LTML link semantics while
// leaving styling decisions to the surrounding paragraph/label/span logic.
type linkedInlineText struct {
	inlineText
	uri    string
	target string
	id     string
}

func (t linkedInlineText) LinkURI() string    { return t.uri }
func (t linkedInlineText) LinkTarget() string { return t.target }
func (t linkedInlineText) LinkID() string     { return t.id }

func (t linkedInlineText) Font() *FontStyle {
	if withFont, ok := t.inlineText.(inlineTextWithFont); ok {
		return withFont.Font()
	}
	return nil
}

func (a *StdA) AddText(text string) {
	a.AddTextWithFont(text, a.Font())
}

func (a *StdA) AddTextWithFont(text string, font *FontStyle) {
	text = normalizeLinkedText(a.container, text)
	if text == "" {
		return
	}
	a.AddInlineWithFont(staticInlineText(text), font)
}

func (a *StdA) AddInlineWithFont(content inlineText, font *FontStyle) {
	if content == nil {
		return
	}
	a.container.(AddInlineWithFonter).AddInlineWithFont(linkedInlineText{
		inlineText: content,
		uri:        a.uri,
		target:     a.target,
		id:         a.AccessibilityLogicalID(),
	}, font)
}

func (a *StdA) LinkURI() string {
	return a.uri
}

func (a *StdA) LinkTarget() string {
	return a.target
}

func (a *StdA) SetAttrs(attrs map[string]string) {
	a.StdSpan.SetAttrs(attrs)
	a.uri = attrs["uri"]
	a.target = attrs["target"]
}

func (a *StdA) SetContainer(container Container) error {
	switch container.(type) {
	case *StdA:
	case *StdSpan:
	case *StdParagraph:
	case *StdLabel:
	default:
		return fmt.Errorf("a must be child of p, label, span or another a")
	}
	a.container = container
	return nil
}

func normalizeLinkedText(container Container, text string) string {
	if container == nil {
		return normalizeXMLText(text)
	}
	root := inlineRootContainer(container)
	switch root.(type) {
	case *StdLabel:
		text = normalizeLabelXMLText(text)
	default:
		text = normalizeXMLText(text)
	}
	if text == "" {
		return ""
	}
	last := lastInlineResolvedText(root)
	if last == "" {
		return strings.TrimLeftFunc(text, unicode.IsSpace)
	}
	if strings.HasSuffix(last, " ") && strings.HasPrefix(text, " ") {
		return text[1:]
	}
	return text
}

func inlineRootContainer(container Container) Container {
	for container != nil {
		switch value := container.(type) {
		case *StdA:
			container = value.container
		case *StdSpan:
			container = value.container
		default:
			return container
		}
	}
	return nil
}

func lastInlineResolvedText(container Container) string {
	switch value := container.(type) {
	case *StdParagraph:
		if len(value.textPieces) == 0 {
			return ""
		}
		return value.textPieces[len(value.textPieces)-1].ResolvedText(nil)
	case *StdLabel:
		if len(value.textPieces) == 0 {
			return ""
		}
		return value.textPieces[len(value.textPieces)-1].ResolvedText(nil)
	default:
		return ""
	}
}

func init() {
	registerTag(DefaultSpace, "a", func() any { return &StdA{} })
}
