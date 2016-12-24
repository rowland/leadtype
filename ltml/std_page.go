// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"errors"
	"fmt"
	"regexp"
)

type StdPage struct {
	StdContainer
	Scope
	pageStyle     *PageStyle
	marginChanged bool
}

func (p *StdPage) BeforePrint(w Writer) error {
	fmt.Printf("Printing %s\n", p)
	fmt.Print(&p.Scope)
	w.NewPage()
	LayoutContainer(p, w)
	return nil
}

func (p *StdPage) Bottom() float64 {
	return p.Height()
}

func (p *StdPage) BottomIsSet() bool {
	return true
}

func (p *StdPage) root() *StdPage {
	if p.container == nil {
		return p
	}
	return p.container.(*StdPage)
}

func (p *StdPage) document() *StdDocument {
	if p.container != nil {
		return p.container.(*StdDocument)
	}
	return nil
}

func (p *StdPage) Height() float64 {
	return p.PageStyle().Height()
}

func (p *StdPage) Left() float64 {
	return 0
}

func (p *StdPage) LeftIsSet() bool {
	return true
}

func (p *StdPage) MarginTop() float64 {
	if p.marginChanged {
		return p.StdContainer.MarginTop()
	}
	if doc := p.document(); doc != nil {
		return doc.MarginTop()
	}
	return 0
}

func (p *StdPage) MarginRight() float64 {
	if p.marginChanged {
		return p.StdContainer.MarginRight()
	}
	if doc := p.document(); doc != nil {
		return doc.MarginRight()
	}
	return 0
}

func (p *StdPage) MarginBottom() float64 {
	if p.marginChanged {
		return p.StdContainer.MarginBottom()
	}
	if doc := p.document(); doc != nil {
		return doc.MarginBottom()
	}
	return 0
}

func (p *StdPage) MarginLeft() float64 {
	if p.marginChanged {
		return p.StdContainer.MarginLeft()
	}
	if doc := p.document(); doc != nil {
		return doc.MarginLeft()
	}
	return 0
}

func (p *StdPage) PageStyle() *PageStyle {
	if p.pageStyle == nil {
		return p.document().PageStyle()
	}
	return p.pageStyle
}

func (p *StdPage) Right() float64 {
	return p.Width()
}

func (p *StdPage) RightIsSet() bool {
	return true
}

var reMargin = regexp.MustCompile(`^margin(-top|-right|-bottom|-left)?$`)

func (p *StdPage) SetAttrs(attrs map[string]string) {
	p.StdContainer.SetAttrs(attrs)
	if style, ok := attrs["style"]; ok {
		p.pageStyle = PageStyleFor(style, p.scope)
	}
	for k, _ := range attrs {
		if reMargin.MatchString(k) {
			p.marginChanged = true
			break
		}
	}
}

var errBadPageContainer = errors.New("page must be child of ltml.")

func (p *StdPage) SetContainer(container Container) error {
	if _, ok := container.(*StdDocument); ok {
		return p.StdContainer.SetContainer(container)
	} else {
		return errBadPageContainer
	}
}

func (p *StdPage) String() string {
	return fmt.Sprintf("StdPage %s", &p.StdContainer)
}

func (p *StdPage) Top() float64 {
	return 0
}

func (p *StdPage) TopIsSet() bool {
	return true
}

func (p *StdPage) Width() float64 {
	return p.PageStyle().Width()
}

func init() {
	registerTag(DefaultSpace, "page", func() interface{} { return &StdPage{} })
}

var _ Container = (*StdPage)(nil)
var _ HasAttrs = (*StdPage)(nil)
var _ Identifier = (*StdPage)(nil)
var _ Printer = (*StdPage)(nil)
var _ WantsContainer = (*StdPage)(nil)
