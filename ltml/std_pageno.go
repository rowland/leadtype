package ltml

import (
	"fmt"
	"strconv"
)

type StdPageNo struct {
	StdSpan
	start  int
	reset  int
	hidden bool
}

func (p *StdPageNo) AddText(text string) {
	// Inner text is ignored. Page-number control semantics are attribute-based.
}

func (p *StdPageNo) Dynamic() bool {
	return true
}

func (p *StdPageNo) Resolve(doc *StdDocument) string {
	if p.hidden || doc == nil {
		return ""
	}
	return formatPageNo(doc.CurrentPageNo())
}

func (p *StdPageNo) SetAttrs(attrs map[string]string) {
	p.StdContainer.SetAttrs(attrs)
	if start, ok := attrs["start"]; ok {
		if value, err := strconv.Atoi(start); err == nil {
			p.start = value
		}
	}
	if reset, ok := attrs["reset"]; ok {
		if value, err := strconv.Atoi(reset); err == nil {
			p.reset = value
		}
	}
	if hidden, ok := attrs["hidden"]; ok {
		p.hidden = hidden == "true"
	}
}

func (p *StdPageNo) SetContainer(container Container) error {
	if err := p.StdSpan.SetContainer(container); err != nil {
		return err
	}
	inlineContainer, ok := container.(InlineContainer)
	if !ok {
		return fmt.Errorf("pageno must be child of an inline text container")
	}
	inlineContainer.AddInlineWithFont(p, nil)
	return nil
}

func (p *StdPageNo) String() string {
	return fmt.Sprintf("StdPageNo start=%d reset=%d hidden=%t %s", p.start, p.reset, p.hidden, &p.StdSpan)
}

func (p *StdPageNo) hasReset() bool {
	return p.reset > 0
}

func (p *StdPageNo) hasStart() bool {
	return p.start > 0
}

func init() {
	registerTag(DefaultSpace, "pageno", func() interface{} { return &StdPageNo{} })
}

var _ HasAttrs = (*StdPageNo)(nil)
var _ HasText = (*StdPageNo)(nil)
