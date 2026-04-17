// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
	"strings"
	"sync/atomic"
)

type StdComponent struct {
	StdContainer
	body        string
	src         string
	srcExplicit bool
}

var networkAssetsDisabled atomic.Bool

func DisableNetworkAssets() {
	networkAssetsDisabled.Store(true)
}

func (c *StdComponent) Body() string {
	if !c.srcExplicit || strings.TrimSpace(c.src) == "" {
		return c.body
	}
	ref, err := c.assetSource()
	if err != nil {
		return ""
	}
	data, err := readAssetSource(c.doc, ref)
	if err != nil {
		return ""
	}
	return string(data)
}

func (c *StdComponent) SetBody(body string) {
	if c.srcExplicit {
		return
	}
	c.body = body
}

func (c *StdComponent) SetAttrs(attrs map[string]string) {
	c.StdContainer.SetAttrs(attrs)
	src, ok := attrs["src"]
	if !ok || strings.TrimSpace(src) == "" {
		return
	}
	src = strings.TrimSpace(src)
	c.src = src
	c.srcExplicit = true
}

func (c *StdComponent) ensureBody() error {
	if !c.srcExplicit || strings.TrimSpace(c.src) == "" {
		return nil
	}
	_, err := c.assetSource()
	return err
}

func (c *StdComponent) assetSource() (assetSourceRef, error) {
	if !c.srcExplicit || strings.TrimSpace(c.src) == "" {
		return assetSourceRef{}, nil
	}
	if c.doc == nil {
		return assetSourceRef{}, fmt.Errorf("component document is not set")
	}
	return c.doc.resolveAssetSource(c.container, c.src)
}

var _ Component = (*StdComponent)(nil)
var _ HasAttrs = (*StdComponent)(nil)
var _ WantsDoc = (*StdComponent)(nil)
