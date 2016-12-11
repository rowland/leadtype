// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
)

type StdParagraph struct {
	Widget
	Children
	richText []string
}

func (p *StdParagraph) AddText(text string) {
	p.richText = append(p.richText, text)
}

func (p *StdParagraph) Print(w Writer) error {
	fmt.Printf("Printing %s\n", p)
	if err := p.Widget.Print(w); err != nil {
		return err
	}
	w.MoveTo(p.Left(), p.Top())
	p.Font().Apply(w)
	for _, s := range p.richText {
		fmt.Println(s)
		if err := w.Print(s); err != nil {
			return err
		}
	}
	return nil
}

func (p *StdParagraph) SetAttrs(attrs map[string]string) {
	p.Widget.SetAttrs(attrs)
}

func (p *StdParagraph) String() string {
	return fmt.Sprintf("StdParagraph %s", &p.Widget)
}

func init() {
	registerTag(DefaultSpace, "p", func() interface{} { return &StdParagraph{} })
}

var _ Container = (*StdParagraph)(nil)
var _ HasAttrs = (*StdParagraph)(nil)
var _ Identifier = (*StdParagraph)(nil)
var _ Printer = (*StdParagraph)(nil)
var _ WantsContainer = (*StdParagraph)(nil)
