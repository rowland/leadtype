// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/rowland/leadtype/colors"
	"github.com/rowland/leadtype/options"
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
	id      string
	entries []fontEntry
	size    float64

	color      colors.Color
	strikeout  bool
	style      string
	underline  bool
	weight     string
	lineHeight float64
}

func (fs *FontStyle) Apply(w Writer) {
	if len(fs.entries) == 0 {
		return
	}
	baseOpts := options.Options{
		"color":  colors.Color(fs.color),
		"weight": fs.weight,
		"style":  fs.style,
	}
	loadedPrimary := false
	for _, entry := range fs.entries {
		opts := applyEntryOptions(entry, baseOpts)
		if !loadedPrimary {
			if fonts, err := w.SetFont(entry.name, fs.size, opts); err == nil && len(fonts) > 0 {
				loadedPrimary = true
			}
			continue
		}
		w.AddFont(entry.name, opts)
	}
	if !loadedPrimary {
		// Keep LTML renderable on machines that lack requested system fonts.
		w.SetFont(defaultFontName, fs.size, baseOpts)
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
}

// SetAttrs applies XML attributes to the FontStyle.  The prefix is prepended to
// every attribute key before lookup (e.g. "font." when attrs are inline on
// another element, "" when the element is a <font> element itself).
//
// Supported attributes (shown without prefix):
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
func (fs *FontStyle) SetAttrs(prefix string, attrs map[string]string) {
	if id, ok := attrs[prefix+"id"]; ok {
		fs.id = id
	}
	// Process "name" first so that "ranges" and "sizes" have entries to target.
	if name, ok := attrs[prefix+"name"]; ok {
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
	if rangesAttr, ok := attrs[prefix+"ranges"]; ok {
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
	if sizesAttr, ok := attrs[prefix+"sizes"]; ok {
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
	if size, ok := attrs[prefix+"size"]; ok {
		var err error
		if fs.size, err = strconv.ParseFloat(size, 64); err != nil {
			fs.size = defaultFontSize
		}
	}
	if color, ok := attrs[prefix+"color"]; ok {
		fs.color = NamedColor(color)
	}
	if strikeout, ok := attrs[prefix+"strikeout"]; ok {
		fs.strikeout = (strikeout == "true")
	}
	if style, ok := attrs[prefix+"style"]; ok {
		fs.style = style
	}
	if underline, ok := attrs[prefix+"underline"]; ok {
		fs.underline = (underline == "true")
	}
	if weight, ok := attrs[prefix+"weight"]; ok {
		fs.weight = weight
	}
	if lineHeight, ok := attrs[prefix+"line-height"]; ok {
		fs.lineHeight, _ = strconv.ParseFloat(lineHeight, 64)
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
	return fmt.Sprintf("FontStyle id=%s name=%s size=%f color=%v strikeout=%t style=%s underline=%t weight=%s line-height=%f",
		fs.id, strings.Join(names, ","), fs.size, fs.color, fs.strikeout, fs.style, fs.underline, fs.weight, fs.lineHeight)
}

func (fs *FontStyle) RichTextOptions() options.Options {
	return options.Options{
		"color":     fs.color,
		"strikeout": fs.strikeout,
		"underline": fs.underline,
	}
}

func FontStyleFor(id string, scope HasScope) *FontStyle {
	if style, ok := scope.StyleFor(id); ok {
		fs, _ := style.(*FontStyle)
		return fs
	}
	return nil
}

var _ HasAttrsPrefix = (*FontStyle)(nil)
var _ Styler = (*FontStyle)(nil)

func init() {
	registerTag(DefaultSpace, "font", func() any { return &FontStyle{} })
}
