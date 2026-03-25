// Copyright 2024 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

//go:build arabic && !harfbuzz

package shaping

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/go-text/typesetting/di"
	"github.com/go-text/typesetting/font"
	"github.com/go-text/typesetting/language"
	gotextshaping "github.com/go-text/typesetting/shaping"
	"golang.org/x/image/math/fixed"
)

// NewShaper returns a pure-Go Arabic shaper backed by go-text/typesetting.
func NewShaper() Shaper { return &goTextShaper{} }

type goTextShaper struct {
	mu         sync.Mutex
	shaper     gotextshaping.HarfbuzzShaper
	lastKey    string     // FontKey of the cached font
	lastParsed *font.Font // go-text Font; thread-safe for reads
}

// parsedFont returns the go-text *font.Font for fr, using the size-1 cache.
// fr.Bytes() is called only on a cache miss, keeping disk I/O out of the
// hot path for documents that use a single Arabic font.
func (s *goTextShaper) parsedFont(fr FontReader) (*font.Font, error) {
	key := fr.FontKey()

	s.mu.Lock()
	if key != "" && key == s.lastKey && s.lastParsed != nil {
		f := s.lastParsed
		s.mu.Unlock()
		return f, nil
	}
	s.mu.Unlock()

	// Cache miss: read and parse without holding the lock so other goroutines
	// can still hit the cache concurrently. Worst case: two goroutines both
	// miss and parse the same font; the second write is harmless.
	data := fr.Bytes()
	if len(data) == 0 {
		return nil, fmt.Errorf("shaping: no font data for %q", key)
	}
	loaded, err := font.ParseTTF(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("shaping: parse font %q: %w", key, err)
	}

	s.mu.Lock()
	s.lastKey = key
	s.lastParsed = loaded.Font
	s.mu.Unlock()

	return loaded.Font, nil
}

// Shape shapes a run of Arabic text using go-text/typesetting's pure-Go
// HarfBuzz port. The returned glyphs are in visual (display) order.
func (s *goTextShaper) Shape(text []rune, fr FontReader, ppem float32) ([]GlyphPosition, error) {
	parsed, err := s.parsedFont(fr)
	if err != nil {
		return nil, err
	}

	// font.Face wraps Font with per-call glyph-extent caching; not thread-safe.
	face := font.NewFace(parsed)

	input := gotextshaping.Input{
		Text:      text,
		RunStart:  0,
		RunEnd:    len(text),
		Direction: di.DirectionRTL,
		Face:      face,
		Size:      fixed.Int26_6(ppem * 64),
		Script:    language.Arabic,
		Language:  language.NewLanguage("ar"),
	}

	s.mu.Lock()
	output := s.shaper.Shape(input)
	s.mu.Unlock()

	glyphs := make([]GlyphPosition, len(output.Glyphs))
	for i, g := range output.Glyphs {
		glyphs[i] = GlyphPosition{
			GlyphID:      uint16(g.GlyphID),
			XAdvance:     int32(g.XAdvance),
			YAdvance:     int32(g.YAdvance),
			XOffset:      int32(g.XOffset),
			YOffset:      int32(g.YOffset),
			ClusterIndex: g.ClusterIndex,
		}
	}
	return glyphs, nil
}
