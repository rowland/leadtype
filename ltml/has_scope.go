// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

type HasScope interface {
	AddLayout(layout *LayoutStyle) error
	AddStyle(style Styler) error
	Alias(name string) (alias *Alias, ok bool)
	Layout(id string) (layout *LayoutStyle, ok bool)
	SetParentScope(parent HasScope)
	Style(id string) (style Styler, ok bool)
}
