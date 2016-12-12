// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
)

type LayoutStyle struct {
	id       string
	units    Units
	hpadding float64
	vpadding float64
	manager  string
}

func (ls *LayoutStyle) Clone() *LayoutStyle {
	clone := *ls
	return &clone
}

func (ls *LayoutStyle) ID() string {
	return ls.id
}

func (ls *LayoutStyle) Layout(c Container, w Writer) {
	fmt.Println("In Layout")
	// f := LayoutManagerFor(ls.manager)
	// f(c, ls, w)
	LayoutVBox(c, ls, w)
}

func (ls *LayoutStyle) SetAttrs(attrs map[string]string) {
	if id, ok := attrs["id"]; ok {
		ls.id = id
	}
	if units, ok := attrs["units"]; ok {
		ls.units = Units(units)
	}
	if padding, ok := attrs["padding"]; ok {
		hvpadding := ParseMeasurement(padding, ls.units)
		ls.hpadding = hvpadding
		ls.hpadding = hvpadding
	}
	if hpadding, ok := attrs["hpadding"]; ok {
		ls.hpadding = ParseMeasurement(hpadding, ls.units)
	}
	if vpadding, ok := attrs["vpadding"]; ok {
		ls.vpadding = ParseMeasurement(vpadding, ls.units)
	}
	if manager, ok := attrs["manager"]; ok {
		ls.manager = manager
	}
}

func (ls *LayoutStyle) String() string {
	return fmt.Sprintf("LayoutStyle id=%s units=%s hpadding=%f vpadding=%f manager=%s",
		ls.id, ls.units, ls.hpadding, ls.vpadding, ls.manager)
}

func LayoutStyleFor(id string, scope HasScope) *LayoutStyle {
	fmt.Println("In LayoutStyleFor", id)
	if ls, ok := scope.Layout(id); ok {
		fmt.Println("Found LayoutStyle", ls)
		return ls
	}
	return nil
}

var _ HasAttrs = (*LayoutStyle)(nil)

func init() {
	registerTag(DefaultSpace, "layout", func() interface{} { return &LayoutStyle{} })
}
