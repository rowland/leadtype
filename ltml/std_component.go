// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync/atomic"
)

type StdComponent struct {
	StdContainer
	body        string
	src         string
	srcExplicit bool
	doc         *Doc
}

var networkAssetsDisabled atomic.Bool

func DisableNetworkAssets() {
	networkAssetsDisabled.Store(true)
}

func (c *StdComponent) Body() string {
	return c.body
}

func (c *StdComponent) SetBody(body string) {
	if c.srcExplicit {
		return
	}
	c.body = body
}

func (c *StdComponent) SetDoc(doc *Doc) {
	c.doc = doc
}

func (c *StdComponent) SetAttrs(attrs map[string]string) {
	c.StdContainer.SetAttrs(attrs)
	src, ok := attrs["src"]
	if !ok || strings.TrimSpace(src) == "" {
		return
	}
	c.src = strings.TrimSpace(src)
	c.srcExplicit = true
	body, err := c.loadBody(c.src)
	if err != nil {
		if c.doc != nil && c.doc.parseErr == nil {
			c.doc.parseErr = err
		}
		return
	}
	c.body = body
}

func (c *StdComponent) loadBody(src string) (string, error) {
	if assetURL, err := url.Parse(src); err == nil && assetURL.Scheme != "" {
		return c.loadURLBody(assetURL)
	}
	return c.loadFileBody(src)
}

func (c *StdComponent) loadFileBody(src string) (string, error) {
	if c.doc != nil && c.doc.assetFS != nil {
		if !validAssetPath(src) {
			return "", fmt.Errorf("invalid asset path %q", src)
		}
		data, err := fs.ReadFile(c.doc.assetFS, src)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}
	data, err := os.ReadFile(src)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (c *StdComponent) loadURLBody(assetURL *url.URL) (string, error) {
	if assetURL == nil || !assetURL.IsAbs() {
		return "", fmt.Errorf("component src URL must be absolute: %q", c.src)
	}
	switch assetURL.Scheme {
	case "http", "https":
	default:
		return "", fmt.Errorf("unsupported component src URL scheme %q", assetURL.Scheme)
	}
	if networkAssetsDisabled.Load() {
		return "", fmt.Errorf("network assets are globally disabled")
	}
	doc := documentForContainer(c.container)
	if doc == nil || !doc.NetworkAssetsEnabled() {
		return "", fmt.Errorf("network assets are disabled for this document")
	}

	resp, err := http.Get(assetURL.String())
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("fetching %q returned %s", assetURL.String(), resp.Status)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

var _ Component = (*StdComponent)(nil)
var _ HasAttrs = (*StdComponent)(nil)
var _ WantsDoc = (*StdComponent)(nil)
