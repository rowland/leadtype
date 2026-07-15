package ltml

import (
	"strings"
	"unicode"
)

type StdIndexEntry struct {
	StdWidget
	indexID string
	target  string
	text    string
}

func (e *StdIndexEntry) AddText(text string) {
	text = normalizeXMLText(text)
	if text == "" {
		return
	}
	if e.text == "" {
		text = strings.TrimLeftFunc(text, unicode.IsSpace)
	} else if strings.HasSuffix(e.text, " ") && strings.HasPrefix(text, " ") {
		text = text[1:]
	}
	if text == "" {
		return
	}
	e.text += text
}

func (e *StdIndexEntry) PreferredHeight(Writer) (float64, error) {
	return 0, nil
}

func (e *StdIndexEntry) PreferredWidth(Writer) (float64, error) {
	return 0, nil
}

func (e *StdIndexEntry) ResolveText(*StdDocument) string {
	return e.text
}

func (e *StdIndexEntry) SetAttrs(attrs map[string]string) {
	e.StdWidget.SetAttrs(attrs)
	e.indexID = attrs["index"]
	e.target = attrs["target"]
	e.SetWidth(0)
	e.SetHeight(0)
}

func (e *StdIndexEntry) ZeroFootprint() bool {
	return true
}

func init() {
	registerTag(DefaultSpace, "index-entry", func() any { return &StdIndexEntry{} })
}

var _ HasAttrs = (*StdIndexEntry)(nil)
var _ HasText = (*StdIndexEntry)(nil)
var _ Identifier = (*StdIndexEntry)(nil)
var _ WantsContainer = (*StdIndexEntry)(nil)
var _ ZeroFootprint = (*StdIndexEntry)(nil)
