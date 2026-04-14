// Copyright 2012-2014 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package font

import (
	"fmt"
	"strconv"

	"github.com/rowland/leadtype/colors"
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

	strikeoutPosition  int16
	strikeoutThickness int16
	strikeoutColor     colors.Color
	strikeoutCapStyle  string
	underlinePosition  int16
	underlineThickness int16
	underlineColor     colors.Color
	underlineCapStyle  string

	strikeoutPositionSet  bool
	strikeoutThicknessSet bool
	strikeoutColorSet     bool
	strikeoutCapStyleSet  bool
	underlinePositionSet  bool
	underlineThicknessSet bool
	underlineColorSet     bool
	underlineCapStyleSet  bool
}

func New(family string, options options.Options, fontSources FontSources) (*Font, error) {
	strikeoutPosition, strikeoutPositionSet := decorationInt16Option(options, "strikeout_position")
	strikeoutThickness, strikeoutThicknessSet := decorationInt16Option(options, "strikeout_thickness")
	strikeoutColor, strikeoutColorSet := decorationColorOption(options, "strikeout_color")
	strikeoutCapStyle, strikeoutCapStyleSet := decorationLineCapStyleOption(options, "strikeout_cap")

	underlinePosition, underlinePositionSet := decorationInt16Option(options, "underline_position")
	underlineThickness, underlineThicknessSet := decorationInt16Option(options, "underline_thickness")
	underlineColor, underlineColorSet := decorationColorOption(options, "underline_color")
	underlineCapStyle, underlineCapStyleSet := decorationLineCapStyleOption(options, "underline_cap")

	font := &Font{
		family:                family,
		Weight:                options.StringDefault("weight", ""),
		style:                 options.StringDefault("style", ""),
		RelativeSize:          options.FloatDefault("relative_size", 100) / 100.0,
		strikeoutPosition:     strikeoutPosition,
		strikeoutThickness:    strikeoutThickness,
		strikeoutColor:        strikeoutColor,
		strikeoutCapStyle:     strikeoutCapStyle,
		underlinePosition:     underlinePosition,
		underlineThickness:    underlineThickness,
		underlineColor:        underlineColor,
		underlineCapStyle:     underlineCapStyle,
		strikeoutPositionSet:  strikeoutPositionSet,
		strikeoutThicknessSet: strikeoutThicknessSet,
		strikeoutColorSet:     strikeoutColorSet,
		strikeoutCapStyleSet:  strikeoutCapStyleSet,
		underlinePositionSet:  underlinePositionSet,
		underlineThicknessSet: underlineThicknessSet,
		underlineColorSet:     underlineColorSet,
		underlineCapStyleSet:  underlineCapStyleSet,
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
		if font.metrics, err = fontSource.Select(font.family, font.Weight, font.style, font.Ranges); err == nil {
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
		font.strikeoutPosition == other.strikeoutPosition &&
		font.strikeoutThickness == other.strikeoutThickness &&
		font.strikeoutColor == other.strikeoutColor &&
		font.strikeoutCapStyle == other.strikeoutCapStyle &&
		font.underlinePosition == other.underlinePosition &&
		font.underlineThickness == other.underlineThickness &&
		font.underlineColor == other.underlineColor &&
		font.underlineCapStyle == other.underlineCapStyle &&
		font.strikeoutPositionSet == other.strikeoutPositionSet &&
		font.strikeoutThicknessSet == other.strikeoutThicknessSet &&
		font.strikeoutColorSet == other.strikeoutColorSet &&
		font.strikeoutCapStyleSet == other.strikeoutCapStyleSet &&
		font.underlinePositionSet == other.underlinePositionSet &&
		font.underlineThicknessSet == other.underlineThicknessSet &&
		font.underlineColorSet == other.underlineColorSet &&
		font.underlineCapStyleSet == other.underlineCapStyleSet &&
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
	if font.strikeoutPositionSet {
		return int(font.strikeoutPosition)
	}
	return font.metrics.StrikeoutPosition()
}

func (font *Font) StrikeoutThickness() int {
	if font.strikeoutThicknessSet {
		return int(font.strikeoutThickness)
	}
	return font.metrics.StrikeoutThickness()
}

func (font *Font) StrikeoutColor() (colors.Color, bool) {
	if font.strikeoutColorSet {
		return font.strikeoutColor, true
	}
	return 0, false
}

func (font *Font) StrikeoutCapStyle() string {
	if font.strikeoutCapStyleSet {
		return font.strikeoutCapStyle
	}
	return ""
}

func (font *Font) Style() string {
	return font.style
}

func (font *Font) SubType() string {
	return font.subType
}

func (font *Font) UnderlinePosition() int {
	if font.underlinePositionSet {
		return int(font.underlinePosition)
	}
	return font.metrics.UnderlinePosition()
}

func (font *Font) UnderlineThickness() int {
	if font.underlineThicknessSet {
		return int(font.underlineThickness)
	}
	return font.metrics.UnderlineThickness()
}

func (font *Font) UnderlineColor() (colors.Color, bool) {
	if font.underlineColorSet {
		return font.underlineColor, true
	}
	return 0, false
}

func (font *Font) UnderlineCapStyle() string {
	if font.underlineCapStyleSet {
		return font.underlineCapStyle
	}
	return ""
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

// Subsetter is an optional interface implemented by FontMetrics backends that
// support font subsetting (currently only TTF). When implemented, SubsetBytes
// returns a self-consistent TTF binary containing only the requested glyphs.
type Subsetter interface {
	Subset(glyphIDs []uint16) ([]byte, error)
}

// SubsetBytes returns a TTF binary containing only the supplied glyph IDs, or
// an error if the underlying font type does not support subsetting.
func (font *Font) SubsetBytes(glyphIDs []uint16) ([]byte, error) {
	if s, ok := font.metrics.(Subsetter); ok {
		return s.Subset(glyphIDs)
	}
	return nil, fmt.Errorf("SubsetBytes: font type %T does not support subsetting", font.metrics)
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

func normalizeDecorationLineCapStyle(style string) string {
	switch style {
	case "butt_cap", "round_cap", "projecting_square_cap":
		return style
	default:
		return ""
	}
}

func decorationColorOption(opts options.Options, key string) (colors.Color, bool) {
	value, ok := opts[key]
	if !ok {
		return 0, false
	}
	switch value := value.(type) {
	case colors.Color:
		return value, true
	case string:
		color, err := colors.NamedColor(value)
		if err != nil {
			return 0, false
		}
		return color, true
	case int:
		return colors.Color(value), true
	case int32:
		return colors.Color(value), true
	default:
		return 0, false
	}
}

func decorationLineCapStyleOption(opts options.Options, key string) (string, bool) {
	value, ok := opts[key]
	if !ok {
		return "", false
	}
	style, ok := value.(string)
	if !ok {
		return "", false
	}
	style = normalizeDecorationLineCapStyle(style)
	if style == "" {
		return "", false
	}
	return style, true
}

func decorationInt16Option(opts options.Options, key string) (int16, bool) {
	value, ok := opts[key]
	if !ok {
		return 0, false
	}
	switch value := value.(type) {
	case int:
		return int16InRange(value)
	case int16:
		return value, true
	case int32:
		return int16InRange(int(value))
	case float64:
		return int16InRange(int(value))
	case string:
		i, err := strconv.Atoi(value)
		if err != nil {
			return 0, false
		}
		return int16InRange(i)
	default:
		return 0, false
	}
}

func int16InRange(value int) (int16, bool) {
	if value < -32768 || value > 32767 {
		return 0, false
	}
	return int16(value), true
}
