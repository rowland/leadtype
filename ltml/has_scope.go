// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

type HasScope interface {
	AddAlias(alias *Alias) error
	AddLayout(layout *LayoutStyle) error
	AddPageStyle(style *PageStyle) error
	AddStyle(style Styler) error
	AliasFor(name string) (alias *Alias, ok bool)
	LayoutFor(id string) (layout *LayoutStyle, ok bool)
	PageStyleFor(id string) (style *PageStyle, ok bool)
	SetParentScope(parent HasScope)
	StyleFor(id string) (style Styler, ok bool)
}
