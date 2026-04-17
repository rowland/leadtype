// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
	"strings"
)

// textSource centralizes src parsing and asset-resolution behavior for widgets
// that can resolve body text from either inline content or an external source.
type textSource struct {
	src         string
	srcExplicit bool
}

func (s *textSource) SetAttrs(attrs map[string]string) {
	src, ok := attrs["src"]
	if !ok || strings.TrimSpace(src) == "" {
		return
	}
	s.src = strings.TrimSpace(src)
	s.srcExplicit = true
}

func (s *textSource) Explicit() bool {
	return s.srcExplicit && strings.TrimSpace(s.src) != ""
}

func (s *textSource) assetSource(doc *Doc, container Container, kind string) (assetSourceRef, error) {
	if !s.Explicit() {
		return assetSourceRef{}, nil
	}
	if doc == nil {
		return assetSourceRef{}, fmt.Errorf("%s document is not set", kind)
	}
	return doc.resolveAssetSource(container, s.src)
}

func (s *textSource) Text(doc *Doc, container Container, fallback string, kind string) (string, error) {
	if !s.Explicit() {
		return fallback, nil
	}
	ref, err := s.assetSource(doc, container, kind)
	if err != nil {
		return "", err
	}
	if ref.identifier == "" {
		return "", nil
	}
	data, err := readAssetSource(doc, ref)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
