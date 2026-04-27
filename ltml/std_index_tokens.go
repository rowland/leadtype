package ltml

import "fmt"

type StdIndexTitle struct {
	StdSpan
}

func (t *StdIndexTitle) AddText(string) {
	// Attribute-only inline placeholder.
}

func (t *StdIndexTitle) Dynamic() bool {
	return true
}

func (t *StdIndexTitle) Resolve(doc *StdDocument) string {
	if doc == nil || doc.renderContext == nil || doc.renderContext.activeIndexEntry == nil || documentVisualCaptureActive(doc) {
		return ""
	}
	return doc.renderContext.activeIndexEntry.Label
}

func (t *StdIndexTitle) SetContainer(container Container) error {
	if err := t.StdSpan.SetContainer(container); err != nil {
		return err
	}
	inlineContainer, ok := container.(InlineContainer)
	if !ok {
		return fmt.Errorf("index-title must be child of an inline text container")
	}
	inlineContainer.AddInlineWithFont(t, t.explicitFont())
	return nil
}

type StdIndexPage struct {
	StdSpan
}

func (p *StdIndexPage) AddText(string) {
	// Attribute-only inline placeholder.
}

func (p *StdIndexPage) Dynamic() bool {
	return true
}

func (p *StdIndexPage) Resolve(doc *StdDocument) string {
	if doc == nil || doc.renderContext == nil || doc.renderContext.activeIndexEntry == nil || documentVisualCaptureActive(doc) {
		return ""
	}
	return formatPageNo(doc.renderContext.activeIndexEntry.PageNo)
}

func (p *StdIndexPage) SetContainer(container Container) error {
	if err := p.StdSpan.SetContainer(container); err != nil {
		return err
	}
	inlineContainer, ok := container.(InlineContainer)
	if !ok {
		return fmt.Errorf("index-page must be child of an inline text container")
	}
	inlineContainer.AddInlineWithFont(p, p.explicitFont())
	return nil
}

func init() {
	registerTag(DefaultSpace, "index-title", func() any { return &StdIndexTitle{} })
	registerTag(DefaultSpace, "index-page", func() any { return &StdIndexPage{} })
}

var _ HasAttrs = (*StdIndexTitle)(nil)
var _ HasText = (*StdIndexTitle)(nil)
var _ HasFont = (*StdIndexTitle)(nil)

var _ HasAttrs = (*StdIndexPage)(nil)
var _ HasText = (*StdIndexPage)(nil)
var _ HasFont = (*StdIndexPage)(nil)
