// Copyright 2011-2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ttf_fonts

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
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

// NewFromDirs creates a TtfFonts collection loaded only from dirs, in the
// order supplied. Earlier directories win when multiple files expose the same
// font identifiers. No system directories are added implicitly.
func NewFromDirs(dirs []string) (*TtfFonts, error) {
	var fc TtfFonts
	for _, dir := range dirs {
		if err := fc.AddDir(dir); err != nil {
			return nil, err
		}
	}
	return &fc, nil
}

// ResolveFontDirs parses an ordered, comma-delimited font search path. The
// token "auto" expands in place to the current platform's standard font
// directories. Missing automatic directories are ignored because not every
// standard location exists on every host; explicit directories remain strict.
func ResolveFontDirs(spec string) ([]string, error) {
	return resolveFontDirs(spec, SystemFontDirs())
}

func resolveFontDirs(spec string, systemDirs []string) ([]string, error) {
	parts := strings.Split(spec, ",")
	dirs := make([]string, 0, len(parts)+len(systemDirs))
	seen := make(map[string]struct{})
	add := func(dir string, automatic bool) error {
		info, err := os.Stat(dir)
		if err != nil {
			if automatic {
				return nil
			}
			return fmt.Errorf("font directory %q: %w", dir, err)
		}
		if !info.IsDir() {
			if automatic {
				return nil
			}
			return fmt.Errorf("font directory %q is not a directory", dir)
		}
		key, err := filepath.Abs(filepath.Clean(dir))
		if err != nil {
			key = filepath.Clean(dir)
		}
		if _, ok := seen[key]; ok {
			return nil
		}
		seen[key] = struct{}{}
		dirs = append(dirs, dir)
		return nil
	}

	for _, part := range parts {
		entry := strings.TrimSpace(part)
		if entry == "" {
			return nil, fmt.Errorf("font directory list contains an empty entry")
		}
		if entry == "auto" {
			for _, dir := range systemDirs {
				if err := add(dir, true); err != nil {
					return nil, err
				}
			}
			continue
		}
		if err := add(entry, false); err != nil {
			return nil, err
		}
	}
	return dirs, nil
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

// WriteCatalog writes a stable, tab-separated inventory of the font identifiers
// that Select can match, plus outline and CID-keying diagnostics.
func (fc *TtfFonts) WriteCatalog(w io.Writer) error {
	rows := make([]string, 0, len(fc.FontInfos))
	for _, f := range fc.FontInfos {
		if f == nil {
			continue
		}
		outline := font.OutlineTrueType
		if f.HasCFFOutlines() {
			outline = font.OutlineCFF
		}
		cidKeyed := "false"
		if cid, err := f.HasCIDKeyedCFFOutlines(); err != nil {
			cidKeyed = "error"
		} else if cid {
			cidKeyed = "true"
		}
		rows = append(rows, fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%d\t%s\t%d",
			f.Family(), f.Style(), f.FullName(), f.PostScriptName(), outline, cidKeyed, f.NumGlyphs(), f.Filename(), f.TTCOffset()))
	}
	sort.Strings(rows)
	if _, err := fmt.Fprintln(w, "family\tstyle\tfull_name\tpostscript_name\toutline\tcid_keyed\tglyphs\tfile\tttc_offset"); err != nil {
		return err
	}
	for _, row := range rows {
		if _, err := fmt.Fprintln(w, row); err != nil {
			return err
		}
	}
	return nil
}

func (fc *TtfFonts) CatalogString() string {
	var b bytes.Buffer
	_ = fc.WriteCatalog(&b)
	return b.String()
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

// SelectClosest chooses a real face from the requested family when Select
// cannot satisfy weight/style exactly. It never crosses into another family
// and never synthesizes bold or slant. FontInfos retain insertion order, so an
// equal score preserves directory-first precedence and deterministic results.
func (fc *TtfFonts) SelectClosest(family, weight, style string, ranges []string) (font.FontMetrics, error) {
	type candidate struct {
		info  *ttf.FontInfo
		score int
	}
	var best *candidate
	for _, info := range fc.FontInfos {
		if !fontIdentifierMatches(info, family) || !fontInfoSupportsRanges(info, ranges) {
			continue
		}
		score := closestFaceScore(info, weight, style)
		if best == nil || score < best.score {
			best = &candidate{info: info, score: score}
		}
	}
	if best == nil {
		return nil, fmt.Errorf("Font %s has no available face", family)
	}
	if fc.fonts == nil {
		fc.fonts = make(map[string]*ttf.Font)
	}
	cacheKey := fmt.Sprintf("%s@%d", best.info.Filename(), best.info.TTCOffset())
	selected := fc.fonts[cacheKey]
	var err error
	if selected == nil {
		selected, err = ttf.LoadFontAtOffset(best.info.Filename(), best.info.TTCOffset())
		if err != nil {
			return nil, err
		}
		fc.fonts[cacheKey] = selected
	}
	return selected, nil
}

func fontInfoSupportsRanges(info *ttf.FontInfo, ranges []string) bool {
	for _, r := range ranges {
		cpr, ok := ttf.CodepointRangesByName[r]
		if !ok || !info.CharRanges().IsSet(int(cpr.Bit)) {
			return false
		}
	}
	return true
}

// fontIdentifierMatches accepts every identifier exposed by WriteCatalog. A
// PostScript base name is also accepted so a family whose records expose only
// names such as SourceHanSans-Light remains selectable as SourceHanSans.
func fontIdentifierMatches(info *ttf.FontInfo, requested string) bool {
	if info == nil {
		return false
	}
	want := normalizedFontName(requested)
	if normalizedFontName(info.Family()) == want || normalizedFontName(info.FullName()) == want || normalizedFontName(info.PostScriptName()) == want {
		return true
	}
	base, _, ok := splitPostScriptStyleName(info.PostScriptName())
	return ok && normalizedFontName(base) == want
}

// closestFaceScore gives upright/slanted agreement precedence over weight,
// then applies the CSS Fonts weight search order. This keeps an upright Light
// face ahead of an italic Bold face for an upright Bold request.
func closestFaceScore(info *ttf.FontInfo, weight, style string) int {
	requestedSlanted := isSlantedStyle(weight + " " + style)
	candidateSlanted := isSlantedStyle(info.Style() + " " + info.FullName() + " " + info.PostScriptName())
	stylePenalty := 0
	if requestedSlanted != candidateSlanted {
		stylePenalty = 100
	}
	requestedWeight := requestedWeightClass(weight)
	return stylePenalty + cssWeightRank(requestedWeight, info.WeightClass())
}

func isSlantedStyle(value string) bool {
	value = strings.ToLower(value)
	return strings.Contains(value, "italic") || strings.Contains(value, "oblique")
}

// requestedWeightClass converts common OpenType style names to CSS's numeric
// weight classes. Unknown, empty, and "Regular" requests all mean 400.
func requestedWeightClass(value string) int {
	v := normalizedFontName(value)
	switch {
	case strings.Contains(v, "thin"):
		return 100
	case strings.Contains(v, "extralight"), strings.Contains(v, "ultralight"):
		return 200
	case strings.Contains(v, "light"):
		return 300
	case strings.Contains(v, "medium"):
		return 500
	case strings.Contains(v, "semibold"), strings.Contains(v, "demibold"):
		return 600
	case strings.Contains(v, "extrabold"), strings.Contains(v, "ultrabold"):
		return 800
	case strings.Contains(v, "black"), strings.Contains(v, "heavy"):
		return 900
	case strings.Contains(v, "bold"):
		return 700
	default:
		return 400
	}
}

// cssWeightRank implements the directional fallback ordering from CSS Fonts.
// The 400/500 interval is special: try weights through 500 first, then lighter
// weights, then 600 and above. Outside that interval, search toward the nearest
// edge before reversing direction. The return value is an ordering rank rather
// than a numeric distance because 400→500 is preferred over 400→300.
func cssWeightRank(requested, candidate int) int {
	if candidate < 1 {
		candidate = 400
	}
	order := make([]int, 0, 9)
	if requested >= 400 && requested <= 500 {
		for w := requested; w <= 500; w += 100 {
			order = append(order, w)
		}
		for w := requested - 100; w >= 100; w -= 100 {
			order = append(order, w)
		}
		for w := 600; w <= 900; w += 100 {
			order = append(order, w)
		}
	} else if requested < 400 {
		for w := requested; w >= 100; w -= 100 {
			order = append(order, w)
		}
		for w := requested + 100; w <= 900; w += 100 {
			order = append(order, w)
		}
	} else {
		for w := requested; w <= 900; w += 100 {
			order = append(order, w)
		}
		for w := requested - 100; w >= 100; w -= 100 {
			order = append(order, w)
		}
	}
	bestRank := len(order)
	bestDistance := 1000
	for rank, weight := range order {
		distance := candidate - weight
		if distance < 0 {
			distance = -distance
		}
		if distance < bestDistance {
			bestDistance = distance
			bestRank = rank
		}
	}
	return bestRank
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
	if strings.EqualFold(f.PostScriptName(), family) || strings.EqualFold(f.FullName(), family) {
		if weightStyle == "" || strings.EqualFold(weightStyle, "Regular") {
			return true
		}
		return fontStyleNameEqual(f.Style(), weightStyle)
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
