// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import "sync/atomic"

type StdComponent struct {
	StdContainer
	body   string
	source textSource
}

var networkAssetsDisabled atomic.Bool

func DisableNetworkAssets() {
	networkAssetsDisabled.Store(true)
}

func (c *StdComponent) Body() string {
	body, err := c.source.Text(c.doc, c.container, c.body, "component")
	if err != nil {
		return ""
	}
	return body
}

func (c *StdComponent) SetBody(body string) {
	if c.source.Explicit() {
		return
	}
	c.body = body
}

func (c *StdComponent) SetAttrs(attrs map[string]string) {
	c.StdContainer.SetAttrs(attrs)
	c.source.SetAttrs(attrs)
}

func (c *StdComponent) ensureBody() error {
	if !c.source.Explicit() {
		return nil
	}
	_, err := c.assetSource()
	return err
}

func (c *StdComponent) assetSource() (assetSourceRef, error) {
	return c.source.assetSource(c.doc, c.container, "component")
}

var _ Component = (*StdComponent)(nil)
var _ HasAttrs = (*StdComponent)(nil)
var _ WantsDoc = (*StdComponent)(nil)
