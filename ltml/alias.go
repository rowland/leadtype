// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

type Alias struct {
	Tag   string
	Attrs map[string]string
}

func (a Alias) Clone() (clone Alias) {
	clone.Tag = a.Tag
	clone.Attrs = make(map[string]string, len(a.Attrs))
	for k, v := range a.Attrs {
		clone.Attrs[k] = v
	}
	return
}

var StdAliases = map[string]*Alias{
	"h":     {"p", map[string]string{"font.weight": "Bold", "text_align": "center", "width": "100%"}},
	"b":     {"span", map[string]string{"font.weight": "Bold"}},
	"i":     {"span", map[string]string{"font.style": "Italic"}},
	"u":     {"span", map[string]string{"underline": "true"}},
	"hbox":  {"div", map[string]string{"layout": "hbox"}},
	"vbox":  {"div", map[string]string{"layout": "vbox"}},
	"table": {"div", map[string]string{"layout": "table"}},
	"layer": {"div", map[string]string{"position": "relative", "width": "100%", "height": "100%"}},
	"br":    {"label", map[string]string{}},
}

var defaultScope = Scope{aliases: StdAliases}
