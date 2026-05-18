// Copyright 2011-2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ttf_fonts

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/rowland/leadtype/font"
	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/shaping"
	"github.com/rowland/leadtype/ttf"
)

type TtfFonts struct {
	FontInfos  []*ttf.FontInfo
	fonts      map[string]*ttf.Font
	shaper     shaping.Shaper // optional override
	shaperOnce sync.Once
}

var systemFontCache struct {
	mu     sync.Mutex
	infos  []*ttf.FontInfo
	err    error
	loaded bool
}

var loadSystemFontInfos = func() ([]*ttf.FontInfo, error) {
	var fc TtfFonts
	var err error
	for _, dir := range SystemFontDirs() {
		for _, ext := range []string{"*.ttf", "*.TTF", "*.ttc", "*.TTC", "*.otf", "*.OTF"} {
			// Ignore errors from individual patterns (directory may not exist).
			if err2 := fc.Add(filepath.Join(dir, ext)); err2 != nil {
				err = err2
			}
		}
	}
	return cloneFontInfos(fc.FontInfos), err
}

func cloneFontInfos(infos []*ttf.FontInfo) []*ttf.FontInfo {
	if len(infos) == 0 {
		return nil
	}
	cloned := make([]*ttf.FontInfo, len(infos))
	copy(cloned, infos)
	return cloned
}

func cachedSystemFontInfos() ([]*ttf.FontInfo, error) {
	systemFontCache.mu.Lock()
	defer systemFontCache.mu.Unlock()

	if !systemFontCache.loaded {
		systemFontCache.infos, systemFontCache.err = loadSystemFontInfos()
		systemFontCache.loaded = true
	}
	return cloneFontInfos(systemFontCache.infos), systemFontCache.err
}

// ClearCache releases the cached system font inventory used by
// NewFromSystemFonts and AddSystemFonts. Future calls will rescan the
// platform's standard font directories.
func ClearCache() {
	systemFontCache.mu.Lock()
	defer systemFontCache.mu.Unlock()
	systemFontCache.infos = nil
	systemFontCache.err = nil
	systemFontCache.loaded = false
}

// SetShaper attaches a complex-script shaper override to this font collection.
// Fonts subsequently selected from this collection will share s instead of the
// default lazily-created shaper.
func (fc *TtfFonts) SetShaper(s shaping.Shaper) { fc.shaper = s }

// Shaper implements font.ShaperSource, allowing font.New to automatically
// propagate a shared shaper to each Font it creates from this collection.
func (fc *TtfFonts) Shaper() shaping.Shaper {
	fc.shaperOnce.Do(func() {
		if fc.shaper == nil {
			fc.shaper = shaping.NewShaper()
		}
	})
	return fc.shaper
}

func New(pattern string) (*TtfFonts, error) {
	var fc TtfFonts
	if err := fc.Add(pattern); err != nil {
		return nil, err
	}
	return &fc, nil
}

// NewFromSystemFonts creates a TtfFonts collection loaded from all standard
// font directories for the current platform.
func NewFromSystemFonts() (*TtfFonts, error) {
	var fc TtfFonts
	if err := fc.AddSystemFonts(); err != nil {
		return nil, err
	}
	return &fc, nil
}

// AddDir adds all TTF, TTC, and OTF fonts found in dir. It returns an error if dir
// does not exist or is not a directory; errors loading individual font files
// are silently ignored, matching the behaviour of AddSystemFonts.
func (fc *TtfFonts) AddDir(dir string) error {
	info, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("font directory %q: %w", dir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("font directory %q is not a directory", dir)
	}
	for _, ext := range []string{"*.ttf", "*.TTF", "*.ttc", "*.TTC", "*.otf", "*.OTF"} {
		fc.Add(filepath.Join(dir, ext)) // errors loading individual fonts are non-fatal
	}
	return nil
}

// AddSystemFonts adds all TTF, TTC, and OTF fonts found in the platform's
// standard font directories.
func (fc *TtfFonts) AddSystemFonts() error {
	infos, err := cachedSystemFontInfos()
	fc.FontInfos = append(fc.FontInfos, infos...)
	return err
}

func Families(families ...string) (fonts []*font.Font) {
	fc, err := NewFromSystemFonts()
	if err != nil {
		panic(err)
	}
	for _, family := range families {
		if family == "" {
			fonts = append(fonts, nil)
			continue
		}
		f, err := font.New(family, options.Options{}, font.FontSources{fc})
		if err != nil {
			panic(err)
		}
		fonts = append(fonts, f)
	}
	return
}

func (fc *TtfFonts) Add(pattern string) (err error) {
	var pathnames []string
	if pathnames, err = filepath.Glob(pattern); err != nil {
		return
	}
	for _, pathname := range pathnames {
		if strings.EqualFold(filepath.Ext(pathname), ".ttc") {
			infos, err2 := ttf.LoadFontInfosFromTTC(pathname)
			if err2 != nil {
				err = fmt.Errorf("Error loading %s: %s", pathname, err2)
				continue
			}
			for _, fi := range infos {
				if fi.HasSupportedOutlines() {
					fc.FontInfos = append(fc.FontInfos, fi)
				}
			}
		} else {
			fi, err2 := ttf.LoadFontInfo(pathname)
			if err2 != nil {
				err = fmt.Errorf("Error loading %s: %s", pathname, err2)
				continue
			}
			if !fi.HasSupportedOutlines() {
				continue
			}
			fc.FontInfos = append(fc.FontInfos, fi)
		}
	}
	return
}

func (fc *TtfFonts) Len() int {
	return len(fc.FontInfos)
}

func (fc *TtfFonts) Select(family, weight, style string, ranges []string) (fontMetrics font.FontMetrics, err error) {
	var ws string
	if weight != "" && style != "" {
		ws = weight + " " + style
	} else if weight == "" && style == "" {
		ws = "Regular"
	} else if style == "" {
		ws = weight
	} else if weight == "" {
		ws = style
	}
	if fc.fonts == nil {
		fc.fonts = make(map[string]*ttf.Font)
	}
search:
	for _, f := range fc.FontInfos {
		if ttfFontInfoMatches(f, family, ws) {
			for _, r := range ranges {
				cpr, ok := ttf.CodepointRangesByName[r]
				if !ok || !f.CharRanges().IsSet(int(cpr.Bit)) {
					continue search
				}
			}
			cacheKey := fmt.Sprintf("%s@%d", f.Filename(), f.TTCOffset())
			font := fc.fonts[cacheKey]
			if font == nil {
				font, err = ttf.LoadFontAtOffset(f.Filename(), f.TTCOffset())
				fc.fonts[cacheKey] = font
			}
			fontMetrics = font
			return
		}
	}
	err = fmt.Errorf("Font %s %s not found", family, ws)
	return
}

func ttfFontInfoMatches(f *ttf.FontInfo, family, weightStyle string) bool {
	if f == nil {
		return false
	}
	if strings.EqualFold(f.Family(), family) && fontStyleNameEqual(f.Style(), weightStyle) {
		return true
	}
	if weightStyle != "" && !strings.EqualFold(weightStyle, "Regular") {
		if strings.EqualFold(f.FullName(), strings.TrimSpace(family+" "+weightStyle)) || strings.EqualFold(f.PostScriptName(), strings.TrimSpace(family+"-"+strings.ReplaceAll(weightStyle, " ", ""))) {
			return true
		}
	}
	_, _, embedsStyle := splitPostScriptStyleName(family)
	if (strings.EqualFold(f.PostScriptName(), family) || strings.EqualFold(f.FullName(), family)) && (styleCompatible(f.Style(), weightStyle) || embedsStyle) {
		return true
	}
	base, inferredStyle, ok := splitPostScriptStyleName(family)
	if !ok {
		return false
	}
	return normalizedFontName(f.Family()) == normalizedFontName(base) && fontStyleNameEqual(f.Style(), inferredStyle)
}

func styleCompatible(fontStyle, requestedStyle string) bool {
	return requestedStyle == "" || strings.EqualFold(requestedStyle, "Regular") || fontStyleNameEqual(fontStyle, requestedStyle)
}

func fontStyleNameEqual(a, b string) bool {
	return normalizedFontName(a) == normalizedFontName(b)
}

func splitPostScriptStyleName(name string) (base, style string, ok bool) {
	name = strings.TrimSpace(name)
	for _, sep := range []string{"-", " "} {
		for _, suffix := range []struct {
			text  string
			style string
		}{
			{"ExtraBlackItalic", "Extra Black Italic"},
			{"ExtraBlack", "Extra Black"},
			{"ExtraBoldItalic", "Extra Bold Italic"},
			{"ExtraBold", "Extra Bold"},
			{"SemiBoldItalic", "Semi Bold Italic"},
			{"SemiBold", "Semi Bold"},
			{"DemiBoldItalic", "Demi Bold Italic"},
			{"DemiBold", "Demi Bold"},
			{"UltraLightItalic", "Ultra Light Italic"},
			{"UltraLight", "Ultra Light"},
			{"ExtraLightItalic", "Extra Light Italic"},
			{"ExtraLight", "Extra Light"},
			{"ThinItalic", "Thin Italic"},
			{"Thin", "Thin"},
			{"LightItalic", "Light Italic"},
			{"Light", "Light"},
			{"MediumItalic", "Medium Italic"},
			{"Medium", "Medium"},
			{"BlackItalic", "Black Italic"},
			{"Black", "Black"},
			{"HeavyItalic", "Heavy Italic"},
			{"Heavy", "Heavy"},
			{"BoldItalic", "Bold Italic"},
			{"BoldOblique", "Bold Oblique"},
			{"Regular", "Regular"},
			{"Italic", "Italic"},
			{"Oblique", "Oblique"},
			{"Bold", "Bold"},
		} {
			token := sep + suffix.text
			if len(name) <= len(token) || !strings.EqualFold(name[len(name)-len(token):], token) {
				continue
			}
			base = strings.TrimSpace(name[:len(name)-len(token)])
			if base == "" {
				return "", "", false
			}
			return base, suffix.style, true
		}
	}
	return "", "", false
}

func normalizedFontName(name string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(name) {
		switch r {
		case ' ', '-', '_':
			continue
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

func (fc *TtfFonts) SubType() string {
	return "TrueType"
}
