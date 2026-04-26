// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
	"strings"
)

type StdCanvas struct {
	StdContainer
	Scope
	key string
}

func (c *StdCanvas) DefaultAttrs(HasScope) map[string]string {
	return map[string]string{
		"layout": "absolute",
	}
}

func (c *StdCanvas) Key() string {
	return c.key
}

func (c *StdCanvas) SetAttrs(attrs map[string]string) {
	c.StdContainer.SetAttrs(attrs)
	c.key = strings.TrimSpace(attrs["key"])
}

func (c *StdCanvas) SetContainer(container Container) error {
	if _, ok := container.(*StdDocument); !ok {
		return fmt.Errorf("canvas must be direct child of ltml")
	}
	return c.StdContainer.SetContainer(container)
}

func (c *StdCanvas) validateDefinition() error {
	if strings.TrimSpace(c.key) == "" {
		return fmt.Errorf("<canvas> requires a key")
	}
	if !c.WidthIsSet() || c.Width() <= 0 {
		return fmt.Errorf("<canvas key=%q> requires a positive width", c.key)
	}
	if !c.HeightIsSet() || c.Height() <= 0 {
		return fmt.Errorf("<canvas key=%q> requires a positive height", c.key)
	}
	return nil
}

func (c *StdCanvas) String() string {
	return fmt.Sprintf("StdCanvas key=%s %s", c.key, &c.StdContainer)
}

func init() {
	registerTag(DefaultSpace, "canvas", func() any { return &StdCanvas{} })
}

var _ Container = (*StdCanvas)(nil)
var _ HasAttrs = (*StdCanvas)(nil)
var _ HasDefaultAttrs = (*StdCanvas)(nil)
var _ HasScope = (*StdCanvas)(nil)
var _ WantsContainer = (*StdCanvas)(nil)
var _ WantsDoc = (*StdCanvas)(nil)
var _ WantsScope = (*StdCanvas)(nil)
