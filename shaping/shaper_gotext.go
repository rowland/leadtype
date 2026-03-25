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
	mu     sync.Mutex
	shaper gotextshaping.HarfbuzzShaper
	fonts  sync.Map // key: fontKey, value: *font.Font
}

// fontKey is a cheap fingerprint for a font's byte slice.
// It is not cryptographically strong, but collision probability for
// distinct font files is negligible in practice.
type fontKey struct {
	head [16]byte
	size int
}

func makeFontKey(data []byte) fontKey {
	var k fontKey
	k.size = len(data)
	copy(k.head[:], data)
	return k
}

// face returns a *font.Face for the given font bytes.
// *font.Font is cached (thread-safe); a new *font.Face wrapping the shared
// Font is created per call because face caches glyph extents and is not
// safe for concurrent use.
func (s *goTextShaper) face(fontBytes []byte) (*font.Face, error) {
	k := makeFontKey(fontBytes)
	if v, ok := s.fonts.Load(k); ok {
		return font.NewFace(v.(*font.Font)), nil
	}
	loaded, err := font.ParseTTF(bytes.NewReader(fontBytes))
	if err != nil {
		return nil, fmt.Errorf("shaping: parse font: %w", err)
	}
	// Store the underlying *Font (thread-safe); discard the wrapper Face.
	actual, _ := s.fonts.LoadOrStore(k, loaded.Font)
	return font.NewFace(actual.(*font.Font)), nil
}

// Shape shapes a run of Arabic text using go-text/typesetting's pure-Go
// HarfBuzz port. The returned glyphs are in visual (display) order.
func (s *goTextShaper) Shape(text []rune, fontBytes []byte, ppem float32) ([]GlyphPosition, error) {
	face, err := s.face(fontBytes)
	if err != nil {
		return nil, err
	}

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

	// HarfbuzzShaper is documented as safe for concurrent use via its
	// internal LRU cache; guard with a mutex to be conservative.
	s.mu.Lock()
	output := s.shaper.Shape(input)
	s.mu.Unlock()

	glyphs := make([]GlyphPosition, len(output.Glyphs))
	for i, g := range output.Glyphs {
		glyphs[i] = GlyphPosition{
			GlyphID:  uint16(g.GlyphID),
			XAdvance: int32(g.XAdvance),
			YAdvance: int32(g.YAdvance),
			XOffset:  int32(g.XOffset),
			YOffset:  int32(g.YOffset),
		}
	}
	return glyphs, nil
}
