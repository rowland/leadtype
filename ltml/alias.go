// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
)

type Alias struct {
	ID    string
	Tag   string
	Attrs map[string]string
}

// func (a Alias) Clone() (clone Alias) {
// 	clone.ID = a.ID
// 	clone.Tag = a.Tag
// 	clone.Attrs = make(map[string]string, len(a.Attrs))
// 	for k, v := range a.Attrs {
// 		clone.Attrs[k] = v
// 	}
// 	return
// }

func (a *Alias) SetAttrs(attrs map[string]string) {
	if a.Attrs == nil {
		a.Attrs = make(map[string]string, len(attrs))
	}
	if id, ok := attrs["id"]; ok {
		a.ID = id
	}
	if tag, ok := attrs["tag"]; ok {
		a.Tag = tag
	}
	for k, v := range attrs {
		a.Attrs[k] = v
	}
	delete(a.Attrs, "id")
	delete(a.Attrs, "tag")
}

func (a *Alias) String() string {
	return fmt.Sprintf("Alias id=%s tag=%s %v",
		a.ID, a.Tag, a.Attrs)
}

var StdAliases = map[string]*Alias{
	"h":     {"h", "p", map[string]string{"font.weight": "Bold", "text_align": "center", "width": "100%"}},
	"b":     {"b", "span", map[string]string{"font.weight": "Bold"}},
	"i":     {"i", "span", map[string]string{"font.style": "Italic"}},
	"u":     {"u", "span", map[string]string{"font.underline": "true"}},
	"hbox":  {"hbox", "div", map[string]string{"layout": "hbox"}},
	"vbox":  {"vbox", "div", map[string]string{"layout": "vbox"}},
	"table": {"table", "div", map[string]string{"layout": "table"}},
	"layer": {"layer", "div", map[string]string{"position": "relative", "width": "100%", "height": "100%"}},
	"br":    {"br", "label", map[string]string{}},
}

func init() {
	registerTag(DefaultSpace, "define", func() interface{} { return &Alias{} })
}

var _ HasAttrs = (*Alias)(nil)
