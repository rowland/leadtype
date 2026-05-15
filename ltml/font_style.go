// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/rowland/leadtype/colors"
	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/rich_text"
	"github.com/rowland/leadtype/ttf"
)

// fontEntry holds the name and optional per-font settings for one font in a
// fallback chain defined by a FontStyle.
type fontEntry struct {
	name         string
	ranges       []string // Unicode range names restricting which codepoints this font serves.
	relativeSize float64  // Size multiplier to normalise rendered size; 0 is treated as 1.0.
}

type FontStyle struct {
	scope    HasScope
	id       string
	entries  []fontEntry
	size     float64
	sizeSpec fontSizeSpec

	color      colors.Color
	strikeout  bool
	style      string
	underline  bool
	weight     string
	lineHeight float64

	strokeColor colors.Color
	strokeWidth *float64

	underlinePenID string
	strikeoutPenID string
	underlinePos   *float64
	strikeoutPos   *float64

	decorationPayload *rich_text.DecorationOverrides
	decorationCached  bool
}

func (fs *FontStyle) Apply(w Writer) {
	fs.applyWithSize(w, fs.ResolveAgainstBase(defaultFontSize))
}

func (fs *FontStyle) applyWithSize(w Writer, size float64) {
	if len(fs.entries) == 0 {
		return
	}
	baseOpts := options.Options{
		"color":  colors.Color(fs.color),
		"weight": fs.weight,
		"style":  fs.style,
	}
	primaryIndex := -1
	for i, entry := range fs.entries {
		opts := applyEntryOptions(entry, baseOpts)
		if fonts, err := w.SetFont(entry.name, size, opts); err == nil && len(fonts) > 0 {
			primaryIndex = i
			break
		}
	}
	if primaryIndex >= 0 {
		needsRelaxed := make([]fontEntry, 0, len(fs.entries)-1)
		for i, entry := range fs.entries {
			if i == primaryIndex {
				continue
			}
			opts := applyEntryOptions(entry, baseOpts)
			if fonts, err := w.AddFont(entry.name, opts); err == nil && len(fonts) > 0 {
				continue
			}
			needsRelaxed = append(needsRelaxed, entry)
		}
		for _, entry := range needsRelaxed {
			if relaxed, ok := relaxedEntryOptions(entry, baseOpts); ok {
				w.AddFont(entry.name, relaxed)
			}
		}
	}
	if primaryIndex < 0 {
		// Keep LTML renderable on machines that lack requested system fonts.
		w.SetFont(defaultFontName, size, baseOpts)
	}
	if fs.lineHeight == 0 {
		fs.lineHeight = 1.0
	}
	w.SetStrikeout(fs.strikeout)
	w.SetUnderline(fs.underline)
	w.SetLineSpacing(fs.lineHeight)
}

// applyEntryOptions copies base and adds per-entry ranges and relative_size.
func applyEntryOptions(entry fontEntry, base options.Options) options.Options {
	opts := make(options.Options, len(base)+2)
	for k, v := range base {
		opts[k] = v
	}
	if len(entry.ranges) > 0 {
		if rs, err := ttf.NewCodepointRangeSet(entry.ranges...); err == nil {
			// Pass as RuneSet so the font restricts which codepoints it serves.
			opts["ranges"] = rs
		} else {
			// Unknown names: fall back to string slice for font-source selection.
			opts["ranges"] = entry.ranges
		}
	}
	if entry.relativeSize != 0 {
		// font.New expects relative_size as a percentage (100 = no adjustment).
		opts["relative_size"] = entry.relativeSize * 100
	}
	return opts
}

// relaxedEntryOptions removes weight/style so fallback fonts can still supply
// missing glyphs when the exact face (e.g. Bold) is unavailable.
func relaxedEntryOptions(entry fontEntry, base options.Options) (options.Options, bool) {
	if base.StringDefault("weight", "") == "" && base.StringDefault("style", "") == "" {
		return nil, false
	}
	opts := make(options.Options, len(base)+2)
	for k, v := range base {
		opts[k] = v
	}
	delete(opts, "weight")
	delete(opts, "style")
	if len(entry.ranges) > 0 {
		if rs, err := ttf.NewCodepointRangeSet(entry.ranges...); err == nil {
			opts["ranges"] = rs
		} else {
			opts["ranges"] = entry.ranges
		}
	}
	if entry.relativeSize != 0 {
		opts["relative_size"] = entry.relativeSize * 100
	}
	return opts, true
}

func (fs *FontStyle) Clone() *FontStyle {
	clone := *fs
	clone.entries = make([]fontEntry, len(fs.entries))
	for i, e := range fs.entries {
		clone.entries[i] = e
		if len(e.ranges) > 0 {
			clone.entries[i].ranges = make([]string, len(e.ranges))
			copy(clone.entries[i].ranges, e.ranges)
		}
	}
	return &clone
}

func (fs *FontStyle) ID() string {
	return fs.id
}

const (
	defaultFontName = "Helvetica"
	defaultFontSize = 12
)

var defaultFont = &FontStyle{
	id:      "default",
	entries: []fontEntry{{name: defaultFontName}},
	size:    defaultFontSize,
	sizeSpec: fontSizeSpec{
		kind:  fontSizeAbsolute,
		value: defaultFontSize,
	},
}

type fontSizeKind int

const (
	fontSizeUnset fontSizeKind = iota
	fontSizeAbsolute
	fontSizeRem
)

type fontSizeSpec struct {
	kind  fontSizeKind
	value float64
}

var (
	reFontSizeRem = regexp.MustCompile(`^\s*([+-]?\d+(\.\d+)?)rem\s*$`)
	reFontSizePt  = regexp.MustCompile(`^\s*([+-]?\d+(\.\d+)?)pt\s*$`)
)

func parseFontSizeSpec(value string) (fontSizeSpec, bool) {
	if matches := reFontSizeRem.FindStringSubmatch(value); len(matches) >= 2 {
		if v, err := strconv.ParseFloat(matches[1], 64); err == nil {
			return fontSizeSpec{kind: fontSizeRem, value: v}, true
		}
		return fontSizeSpec{}, false
	}
	if matches := reFontSizePt.FindStringSubmatch(value); len(matches) >= 2 {
		if v, err := strconv.ParseFloat(matches[1], 64); err == nil {
			return fontSizeSpec{kind: fontSizeAbsolute, value: v}, true
		}
		return fontSizeSpec{}, false
	}
	if v, err := strconv.ParseFloat(strings.TrimSpace(value), 64); err == nil {
		return fontSizeSpec{kind: fontSizeAbsolute, value: v}, true
	}
	return fontSizeSpec{}, false
}

func (spec fontSizeSpec) String() string {
	switch spec.kind {
	case fontSizeRem:
		return strconv.FormatFloat(spec.value, 'f', -1, 64) + "rem"
	case fontSizeAbsolute:
		return strconv.FormatFloat(spec.value, 'f', -1, 64)
	default:
		return ""
	}
}

func (spec fontSizeSpec) ResolveAgainstBase(base float64) float64 {
	switch spec.kind {
	case fontSizeRem:
		return base * spec.value
	case fontSizeAbsolute:
		return spec.value
	default:
		return base
	}
}

// SetAttrs applies XML attributes to the FontStyle.
//
// Supported attributes:
//
//	name     – comma-separated list of font-family names, defining the fallback
//	           chain.  A single name is backward-compatible with the old API.
//	ranges   – pipe-separated groups of comma-separated Unicode range names, one
//	           group per font in the name list.  An empty group means no range
//	           restriction for that position.  Example:
//	             "Basic Latin, Latin-1 Supplement | CJK Unified Ideographs"
//	sizes    – pipe-separated relative-size multipliers, one per font.  Use to
//	           normalise fonts that render at visually different sizes for the
//	           same point size.  Example: "1.0 | 0.9"
//	size     – shared point size for all fonts in the chain.
//	color, weight, style, strikeout, underline, line-height – as before.
func (fs *FontStyle) SetAttrs(attrs map[string]string) {
	fs.decorationCached = false
	fs.decorationPayload = nil
	if id, ok := attrs["id"]; ok {
		fs.id = id
	}
	// Process "name" first so that "ranges" and "sizes" have entries to target.
	if name, ok := attrs["name"]; ok {
		names := splitCommaTrimmed(name)
		newEntries := make([]fontEntry, len(names))
		for i, n := range names {
			if i < len(fs.entries) {
				newEntries[i] = fs.entries[i] // preserve any existing ranges/size
			}
			newEntries[i].name = n
		}
		fs.entries = newEntries
	}
	if rangesAttr, ok := attrs["ranges"]; ok {
		groups := splitPipe(rangesAttr)
		for i, group := range groups {
			if i >= len(fs.entries) {
				break
			}
			if group == "" {
				fs.entries[i].ranges = nil
			} else {
				fs.entries[i].ranges = splitCommaTrimmed(group)
			}
		}
	}
	if sizesAttr, ok := attrs["sizes"]; ok {
		groups := splitPipe(sizesAttr)
		for i, group := range groups {
			if i >= len(fs.entries) {
				break
			}
			if f, err := strconv.ParseFloat(group, 64); err == nil {
				fs.entries[i].relativeSize = f
			}
		}
	}
	if size, ok := attrs["size"]; ok {
		if spec, ok := parseFontSizeSpec(size); ok {
			fs.sizeSpec = spec
			if spec.kind == fontSizeAbsolute {
				fs.size = spec.value
			}
		} else {
			fs.size = defaultFontSize
			fs.sizeSpec = fontSizeSpec{kind: fontSizeAbsolute, value: defaultFontSize}
		}
	}
	if color, ok := attrs["color"]; ok {
		fs.color = NamedColor(color)
	}

	var units Units = "pt"
	units.SetAttrs(attrs)

	if strokeColor, ok := attrs["stroke-color"]; ok {
		fs.strokeColor = NamedColor(strings.TrimSpace(strokeColor))
	}
	if strokeWidth, ok := attrs["stroke-width"]; ok {
		fs.strokeWidth = ParseOptionalMeasurement(strings.TrimSpace(strokeWidth), units)
	}
	if strikeout, ok := attrs["strikeout"]; ok {
		fs.strikeout = (strikeout == "true")
	}
	if style, ok := attrs["style"]; ok {
		fs.style = style
	}
	if underline, ok := attrs["underline"]; ok {
		fs.underline = (underline == "true")
	}
	if weight, ok := attrs["weight"]; ok {
		fs.weight = weight
	}
	if lineHeight, ok := attrs["line-height"]; ok {
		if h, err := strconv.ParseFloat(lineHeight, 64); err == nil {
			fs.lineHeight = h
		}
	}
	if underlinePenID, ok := attrs["underline-pen"]; ok {
		fs.underlinePenID = strings.TrimSpace(underlinePenID)
	}
	if strikeoutPenID, ok := attrs["strikeout-pen"]; ok {
		fs.strikeoutPenID = strings.TrimSpace(strikeoutPenID)
	}
	if underlinePos, ok := attrs["underline-pos"]; ok {
		fs.underlinePos = ParseOptionalMeasurement(strings.TrimSpace(underlinePos), units)
	}
	if strikeoutPos, ok := attrs["strikeout-pos"]; ok {
		fs.strikeoutPos = ParseOptionalMeasurement(strings.TrimSpace(strikeoutPos), units)
	}
}

// splitCommaTrimmed splits s by commas and trims whitespace, omitting empties.
func splitCommaTrimmed(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			result = append(result, t)
		}
	}
	return result
}

// splitPipe splits s by pipes, trimming whitespace from each element.
// Empty elements are preserved so that positions remain meaningful.
func splitPipe(s string) []string {
	parts := strings.Split(s, "|")
	for i, p := range parts {
		parts[i] = strings.TrimSpace(p)
	}
	return parts
}

func (fs *FontStyle) String() string {
	names := make([]string, len(fs.entries))
	for i, e := range fs.entries {
		names[i] = e.name
	}
	return fmt.Sprintf("FontStyle id=%s name=%s size=%s color=%v strikeout=%t strikeout-pen=%s strikeout-pos=%s style=%s underline=%t underline-pen=%s underline-pos=%s weight=%s line-height=%f",
		fs.id, strings.Join(names, ","), fs.sizeString(), fs.color, fs.strikeout, fs.strikeoutPenID, formatOptionalFloat(fs.strikeoutPos),
		fs.style, fs.underline, fs.underlinePenID, formatOptionalFloat(fs.underlinePos), fs.weight, fs.lineHeight)
}

func (fs *FontStyle) RichTextOptions() options.Options {
	opts := options.Options{
		"color":     fs.color,
		"strikeout": fs.strikeout,
		"underline": fs.underline,
	}
	if decoration := fs.decorationOverrides(); decoration != nil {
		opts["decoration"] = decoration
	}
	return opts
}

func (fs *FontStyle) SetScope(scope HasScope) {
	fs.scope = scope
	fs.decorationCached = false
	fs.decorationPayload = nil
}

func (fs *FontStyle) ResolveAgainstBase(base float64) float64 {
	if fs == nil {
		return base
	}
	switch fs.sizeSpec.kind {
	case fontSizeAbsolute, fontSizeRem:
		return fs.sizeSpec.ResolveAgainstBase(base)
	default:
		return fs.size
	}
}

func (fs *FontStyle) decorationOverrides() *rich_text.DecorationOverrides {
	if fs.decorationCached {
		return fs.decorationPayload
	}
	fs.decorationCached = true
	var payload rich_text.DecorationOverrides
	hasAny := false
	if line, ok := fs.decorationLineOverrides(fs.underlinePenID, fs.underlinePos); ok {
		payload.Underline = line
		hasAny = true
	}
	if line, ok := fs.decorationLineOverrides(fs.strikeoutPenID, fs.strikeoutPos); ok {
		payload.Strikeout = line
		hasAny = true
	}
	if stroke, ok := fs.decorationStrokeOverrides(); ok {
		payload.TextStroke = stroke
		hasAny = true
	}
	if hasAny {
		fs.decorationPayload = &payload
	}
	return fs.decorationPayload
}

func (fs *FontStyle) decorationLineOverrides(penID string, pos *float64) (rich_text.DecorationLineOverrides, bool) {
	var line rich_text.DecorationLineOverrides
	hasAny := false
	if pen := fs.resolveDecorationPen(penID); pen != nil {
		line.Color = pen.color
		line.HasColor = true
		line.Width = pen.width
		line.HasWidth = true
		line.Pattern = pen.pattern
		if line.Pattern == "" {
			line.Pattern = defaultPenPattern
		}
		line.HasPattern = true
		line.CapStyle = pen.Cap()
		hasAny = true
	}
	if pos != nil {
		line.Position = *pos
		line.HasPosition = true
		hasAny = true
	}
	return line, hasAny
}

func (fs *FontStyle) decorationStrokeOverrides() (rich_text.DecorationStrokeOverrides, bool) {
	var stroke rich_text.DecorationStrokeOverrides
	hasAny := false
	if fs.strokeWidth != nil && *fs.strokeWidth > 0 {
		stroke.Width = *fs.strokeWidth
		stroke.HasWidth = true
		stroke.Color = fs.strokeColor
		stroke.HasColor = true
		hasAny = true
	}
	return stroke, hasAny
}

func (fs *FontStyle) resolveDecorationPen(id string) *PenStyle {
	if fs.scope == nil || strings.TrimSpace(id) == "" {
		return nil
	}
	if style, ok := fs.scope.StyleFor(id); ok {
		pen, _ := style.(*PenStyle)
		return pen
	}
	if style, ok := fs.scope.StyleFor("pen_" + id); ok {
		pen, _ := style.(*PenStyle)
		return pen
	}
	return nil
}

func formatOptionalFloat(value *float64) string {
	if value == nil {
		return ""
	}
	return strconv.FormatFloat(*value, 'f', -1, 64)
}

func (fs *FontStyle) sizeString() string {
	if fs == nil {
		return ""
	}
	if text := fs.sizeSpec.String(); text != "" {
		return text
	}
	return strconv.FormatFloat(fs.size, 'f', -1, 64)
}

func FontStyleFor(id string, scope HasScope) *FontStyle {
	if style, ok := scope.StyleFor(id); ok {
		fs, _ := style.(*FontStyle)
		return fs
	}
	return nil
}

var _ HasAttrs = (*FontStyle)(nil)
var _ Styler = (*FontStyle)(nil)
var _ WantsScope = (*FontStyle)(nil)

func init() {
	registerTag(DefaultSpace, "font", func() any { return &FontStyle{} })
}
