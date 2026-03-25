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

// NewShaper returns a pure-Go Arabic shaper backed by go-text/typesetting
// with a default font cache size of 1.
func NewShaper() Shaper { return &goTextShaper{maxSize: 1} }

type fontCacheEntry struct {
	key    string
	parsed *font.Font // go-text Font; thread-safe for reads
}

type goTextShaper struct {
	mu      sync.Mutex
	shaper  gotextshaping.HarfbuzzShaper
	cache   []fontCacheEntry // MRU order: cache[0] is most recently used
	maxSize int
}

// SetFontCacheSize sets the maximum number of parsed fonts held in the LRU.
// The default is 1, which is optimal for documents using a single Arabic font.
// Increase it when a document alternates between several Arabic fonts to avoid
// repeated file reads and re-parses.
func (s *goTextShaper) SetFontCacheSize(n int) {
	if n < 1 {
		n = 1
	}
	s.mu.Lock()
	s.maxSize = n
	if len(s.cache) > n {
		s.cache = s.cache[:n]
	}
	s.mu.Unlock()
}

// cacheLookup returns the cached *font.Font for key, promoting it to the
// front of the cache (MRU position). Must be called with s.mu held.
func (s *goTextShaper) cacheLookup(key string) (*font.Font, bool) {
	for i, e := range s.cache {
		if e.key == key {
			if i > 0 {
				copy(s.cache[1:i+1], s.cache[:i])
				s.cache[0] = e
			}
			return s.cache[0].parsed, true
		}
	}
	return nil, false
}

// cacheStore inserts a new entry at the MRU position, evicting the LRU entry
// if the cache is at capacity. Must be called with s.mu held.
func (s *goTextShaper) cacheStore(key string, f *font.Font) {
	if len(s.cache) < s.maxSize {
		s.cache = append(s.cache, fontCacheEntry{})
	}
	copy(s.cache[1:], s.cache[:len(s.cache)-1])
	s.cache[0] = fontCacheEntry{key: key, parsed: f}
}

// parsedFont returns the go-text *font.Font for fr, consulting the LRU cache.
// fr.Bytes() is called only on a cache miss, keeping disk I/O out of the
// hot path for documents that reuse the same Arabic font.
func (s *goTextShaper) parsedFont(fr FontReader) (*font.Font, error) {
	key := fr.FontKey()

	s.mu.Lock()
	if f, ok := s.cacheLookup(key); ok {
		s.mu.Unlock()
		return f, nil
	}
	s.mu.Unlock()

	// Cache miss: read and parse outside the lock so other goroutines can
	// still hit the cache concurrently. Worst case: two goroutines both miss
	// on the same key and parse it twice; the second store is harmless.
	data := fr.Bytes()
	if len(data) == 0 {
		return nil, fmt.Errorf("shaping: no font data for %q", key)
	}
	loaded, err := font.ParseTTF(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("shaping: parse font %q: %w", key, err)
	}

	s.mu.Lock()
	s.cacheStore(key, loaded.Font)
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
