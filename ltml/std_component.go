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
	"path/filepath"
	"strings"
	"sync/atomic"
)

type StdComponent struct {
	StdContainer
	body              string
	src               string
	srcExplicit       bool
	doc               *Doc
	bodyLoadAttempted bool
	bodyLoadErr       error
}

var networkAssetsDisabled atomic.Bool

func DisableNetworkAssets() {
	networkAssetsDisabled.Store(true)
}

func (c *StdComponent) Body() string {
	if err := c.ensureBody(); err != nil {
		return ""
	}
	return c.body
}

func (c *StdComponent) SetBody(body string) {
	if c.srcExplicit {
		return
	}
	c.body = body
	c.bodyLoadAttempted = true
	c.bodyLoadErr = nil
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
	src = strings.TrimSpace(src)
	if c.src != src || !c.srcExplicit {
		c.body = ""
		c.bodyLoadAttempted = false
		c.bodyLoadErr = nil
	}
	c.src = src
	c.srcExplicit = true
}

func (c *StdComponent) loadBody(src string) (string, error) {
	data, err := loadAssetBytes(c.doc, c.container, src)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (c *StdComponent) ensureBody() error {
	if !c.srcExplicit || strings.TrimSpace(c.src) == "" {
		return nil
	}
	if c.bodyLoadAttempted {
		return c.bodyLoadErr
	}
	c.bodyLoadAttempted = true
	body, err := c.loadBody(c.src)
	if err != nil {
		c.bodyLoadErr = err
		return err
	}
	c.body = body
	c.bodyLoadErr = nil
	return nil
}

func loadAssetBytes(doc *Doc, container Container, src string) ([]byte, error) {
	if assetURL, err := url.Parse(src); err == nil && assetURL.Scheme != "" {
		return loadURLAssetBytes(doc, container, src, assetURL)
	}
	return loadFileAssetBytes(doc, src)
}

func loadFileAssetBytes(doc *Doc, src string) ([]byte, error) {
	if doc != nil && doc.assetFS != nil {
		if !validAssetPath(src) {
			return nil, fmt.Errorf("invalid asset path %q", src)
		}
		data, err := fs.ReadFile(doc.assetFS, src)
		if err != nil {
			return nil, err
		}
		return data, nil
	}
	if doc != nil && doc.sourceDir != "" && !filepath.IsAbs(src) {
		src = filepath.Join(doc.sourceDir, src)
	}
	return os.ReadFile(src)
}

func loadURLAssetBytes(doc *Doc, container Container, src string, assetURL *url.URL) ([]byte, error) {
	if assetURL == nil || !assetURL.IsAbs() {
		return nil, fmt.Errorf("component src URL must be absolute: %q", src)
	}
	switch assetURL.Scheme {
	case "http", "https":
	default:
		return nil, fmt.Errorf("unsupported component src URL scheme %q", assetURL.Scheme)
	}
	if networkAssetsDisabled.Load() {
		return nil, fmt.Errorf("network assets are globally disabled")
	}
	stdDoc := (*StdDocument)(nil)
	if container != nil {
		stdDoc = documentForContainer(container)
	}
	if stdDoc == nil || !stdDoc.NetworkAssetsEnabled() {
		return nil, fmt.Errorf("network assets are disabled for this document")
	}

	resp, err := http.Get(assetURL.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("fetching %q returned %s", assetURL.String(), resp.Status)
	}
	return io.ReadAll(resp.Body)
}

var _ Component = (*StdComponent)(nil)
var _ HasAttrs = (*StdComponent)(nil)
var _ WantsDoc = (*StdComponent)(nil)
