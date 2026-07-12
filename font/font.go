// Copyright 2012-2014 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package font

import (
	"fmt"

	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/shaping"
)

// ShaperSource is an optional interface that FontSource implementations may
// satisfy to provide an Arabic (complex-script) text shaper. When the font
// source selected by New implements ShaperSource, the returned Font has its
// Shaper field set automatically.
type ShaperSource interface {
	Shaper() shaping.Shaper
}

type Font struct {
	family       string
	Weight       string
	style        string
	subType      string
	Ranges       []string
	RuneSet      RuneSet
	RelativeSize float64
	metrics      FontMetrics
	// Shaper is non-nil for fonts whose source supports complex-script shaping
	// (e.g. Arabic). It is set automatically by New when the winning FontSource
	// implements ShaperSource.
	Shaper shaping.Shaper
}

func (font *Font) Clone() *Font {
	if font == nil {
		return nil
	}
	clone := *font
	if font.Ranges != nil {
		clone.Ranges = append([]string(nil), font.Ranges...)
	}
	return &clone
}

func New(family string, options options.Options, fontSources FontSources) (*Font, error) {
	return newFont(family, options, fontSources, false)
}

func NewClosest(family string, options options.Options, fontSources FontSources) (*Font, error) {
	return newFont(family, options, fontSources, true)
}

func newFont(family string, options options.Options, fontSources FontSources, closest bool) (*Font, error) {
	font := &Font{
		family:       family,
		Weight:       options.StringDefault("weight", ""),
		style:        options.StringDefault("style", ""),
		RelativeSize: options.FloatDefault("relative_size", 100) / 100.0,
	}
	if Ranges, ok := options["ranges"]; ok {
		switch Ranges := Ranges.(type) {
		case []string:
			font.Ranges = Ranges
		case RuneSet:
			font.RuneSet = Ranges
		}
	}
	var err error
	for _, fontSource := range fontSources {
		if closest {
			if source, ok := fontSource.(ClosestFontSource); ok {
				font.metrics, err = source.SelectClosest(font.family, font.Weight, font.style, font.Ranges)
			} else {
				font.metrics, err = fontSource.Select(font.family, font.Weight, font.style, font.Ranges)
			}
		} else {
			font.metrics, err = fontSource.Select(font.family, font.Weight, font.style, font.Ranges)
		}
		if err == nil {
			font.subType = fontSource.SubType()
			if ss, ok := fontSource.(ShaperSource); ok {
				font.Shaper = ss.Shaper()
			}
			return font, nil
		}
	}
	return nil, err
}

func (font *Font) AdvanceWidth(codepoint rune) (width int, err bool) {
	return font.metrics.AdvanceWidth(codepoint)
}

func (font *Font) AdvanceWidthForGlyph(glyphID uint16) int {
	return font.metrics.AdvanceWidthForGlyph(glyphID)
}

func (font *Font) GlyphIndex(r rune) uint16 {
	return font.metrics.GlyphIndex(r)
}

func (font *Font) Ascent() int {
	return font.metrics.Ascent()
}

func (font *Font) BoundingBox() [4]int {
	return font.metrics.BoundingBox()
}

func (font *Font) CapHeight() int {
	return font.metrics.CapHeight()
}

func (font *Font) Copyright() string {
	return font.metrics.Copyright()
}

func (font *Font) Descent() int {
	return font.metrics.Descent()
}

func (font *Font) Family() string {
	return font.metrics.Family()
}

func (font *Font) Filename() string {
	return font.metrics.Filename()
}

func (font *Font) Flags() (flags uint32) {
	return font.metrics.Flags()
}

func (font *Font) FullName() string {
	return font.metrics.FullName()
}

func (font *Font) HasRune(rune rune) bool {
	if font.RuneSet == nil {
		_, err := font.metrics.AdvanceWidth(rune)
		return !err
	}
	if font.RuneSet.HasRune(rune) {
		_, err := font.metrics.AdvanceWidth(rune)
		return !err
	}
	return false
}

func (font *Font) HasMetrics() bool {
	return font.metrics != nil
}

func (font *Font) Height() int {
	return font.metrics.Ascent() + -font.metrics.Descent()
}

func (font *Font) ItalicAngle() float64 {
	return font.metrics.ItalicAngle()
}

func (font *Font) Leading() int {
	return font.metrics.Leading()
}

func (font *Font) LineGap() int {
	return font.metrics.LineGap()
}

func (font *Font) Matches(other *Font) bool {
	return font.family == other.family &&
		font.Weight == other.Weight &&
		font.style == other.style &&
		font.subType == other.subType &&
		font.RuneSet == other.RuneSet &&
		font.RelativeSize == other.RelativeSize &&
		stringSlicesEqual(font.Ranges, other.Ranges)
}

func (font *Font) NumGlyphs() int {
	return font.metrics.NumGlyphs()
}

func (font *Font) PostScriptName() string {
	return font.metrics.PostScriptName()
}

func (font *Font) StemV() int {
	return font.metrics.StemV()
}

func (font *Font) SupportsArabic() bool {
	if font == nil || font.metrics == nil {
		return false
	}
	return font.Shaper != nil && font.metrics.SupportsArabic()
}

func (font *Font) StrikeoutPosition() int {
	return font.metrics.StrikeoutPosition()
}

func (font *Font) StrikeoutThickness() int {
	return font.metrics.StrikeoutThickness()
}

func (font *Font) Style() string {
	return font.style
}

func (font *Font) SubType() string {
	return font.subType
}

func (font *Font) UnderlinePosition() int {
	return font.metrics.UnderlinePosition()
}

func (font *Font) UnderlineThickness() int {
	return font.metrics.UnderlineThickness()
}

func (font *Font) UnitsPerEm() int {
	return font.metrics.UnitsPerEm()
}

func (font *Font) Version() string {
	return font.metrics.Version()
}

func (font *Font) XHeight() int {
	return font.metrics.XHeight()
}

// ByteReader is an optional interface implemented by FontMetrics backends that
// can return their raw font file bytes for use by external shapers.
// FontKey must be cheap (no I/O); Bytes may perform I/O and should be called
// only when the shaper's cache misses.
type ByteReader interface {
	FontKey() string
	Bytes() []byte
}

// CIDSystemInfo identifies the character collection used by a CID-keyed font.
// Registry, Ordering, and Supplement are the CFF ROS values and are copied into
// the corresponding PDF dictionaries and Encoding CMap.
type CIDSystemInfo struct {
	Registry   string
	Ordering   string
	Supplement int
}

// CIDMapper is implemented by metrics backends for CID-keyed fonts whose
// glyph IDs and character identifiers are distinct namespaces.
type CIDMapper interface {
	CIDSystemInfo() (registry, ordering string, supplement int, ok bool)
	CIDForGlyph(glyphID uint16) (uint16, bool)
}

func (font *Font) CIDSystemInfo() (CIDSystemInfo, bool) {
	if mapper, ok := font.metrics.(CIDMapper); ok {
		registry, ordering, supplement, found := mapper.CIDSystemInfo()
		return CIDSystemInfo{Registry: registry, Ordering: ordering, Supplement: supplement}, found
	}
	return CIDSystemInfo{}, false
}

// CIDForGlyph translates the source font's glyph index to its character
// identifier. Callers need this when a CID-keyed CFF charset is non-identity;
// the returned CID is not a Unicode value and is not a PDF character code.
func (font *Font) CIDForGlyph(glyphID uint16) (uint16, bool) {
	if mapper, ok := font.metrics.(CIDMapper); ok {
		return mapper.CIDForGlyph(glyphID)
	}
	return 0, false
}

// IsCIDKeyed reports whether the metrics backend exposes a CFF ROS and
// GID-to-CID mapping.
func (font *Font) IsCIDKeyed() bool {
	_, ok := font.CIDSystemInfo()
	return ok
}

// FontKey returns a stable string identifying the underlying font file,
// or "" if the backend does not support it (e.g. AFM fonts).
func (font *Font) FontKey() string {
	if br, ok := font.metrics.(ByteReader); ok {
		return br.FontKey()
	}
	return ""
}

// Bytes returns the raw bytes of the underlying font file, or nil if the font
// backend does not support it (e.g. AFM fonts or TTC collection members).
func (font *Font) Bytes() []byte {
	if br, ok := font.metrics.(ByteReader); ok {
		return br.Bytes()
	}
	return nil
}

// OutlineKind reports the font-program technology used for PDF embedding and
// diagnostics. It does not imply whether a CFF font is name- or CID-keyed; use
// IsCIDKeyed for that distinction.
func (font *Font) OutlineKind() string {
	return font.metrics.OutlineKind()
}

// Subsetter is an optional interface implemented by FontMetrics backends that
// support font subsetting. When implemented, SubsetBytes returns a
// self-consistent font binary containing only the requested glyphs.
type Subsetter interface {
	Subset(glyphIDs []uint16) ([]byte, error)
}

// PDFSubsetter optionally supplies a font program in the exact stream format
// required by PDF. This is used by CID-keyed CFF fonts, whose raw CFF program
// has normative CID charset semantics that an OpenType wrapper does not.
type PDFSubsetter interface {
	PDFSubset(glyphIDs []uint16) (data []byte, subtype string, err error)
}

// SubsetBytes returns a font binary containing only the supplied glyph IDs, or
// an error if the underlying font type does not support subsetting.
func (font *Font) SubsetBytes(glyphIDs []uint16) ([]byte, error) {
	if s, ok := font.metrics.(Subsetter); ok {
		return s.Subset(glyphIDs)
	}
	return nil, fmt.Errorf("SubsetBytes: font type %T does not support subsetting", font.metrics)
}

// PDFSubsetBytes returns the font-program bytes and optional FontFile3 subtype
// appropriate for direct PDF embedding. Most backends return an sfnt subset;
// CID-keyed CFF returns its raw CFF program with subtype CIDFontType0C.
func (font *Font) PDFSubsetBytes(glyphIDs []uint16) ([]byte, string, error) {
	if s, ok := font.metrics.(PDFSubsetter); ok {
		return s.PDFSubset(glyphIDs)
	}
	data, err := font.SubsetBytes(glyphIDs)
	return data, "", err
}

func stringSlicesEqual(sl1, sl2 []string) bool {
	if len(sl1) != len(sl2) {
		return false
	}
	for i, s := range sl1 {
		if s != sl2[i] {
			return false
		}
	}
	return true
}
