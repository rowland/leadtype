// Copyright 2011-2014 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import (
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/rowland/leadtype/codepage"
	"github.com/rowland/leadtype/colors"
	"github.com/rowland/leadtype/font"
	"github.com/rowland/leadtype/internal/assetpath"
	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/profile"
	"github.com/rowland/leadtype/rich_text"
	"github.com/rowland/leadtype/svg"
)

type DocWriter struct {
	pages                 []*PageWriter
	nextSeq               func() int
	file                  *file
	catalog               *catalog
	accessibility         *accessibilityState
	resources             *resources
	pagesAcross           int
	pagesDown             int
	curPage               *PageWriter
	options               options.Options
	fontSources           font.FontSources
	selectedFonts         map[string]*font.Font
	fontKeys              map[string]string
	fontEncodings         map[string]*fontEncoding
	glyphRecorders        map[string]*glyphRecorder  // keyed by font PostScript name
	unicodeFonts          map[string]*font.Font      // PostScript name → font, for width lookup at Close
	cidFonts              map[string]*cidFont        // PostScript name → CID font, for /W update at Close
	type0Fonts            map[string]*type0Font      // PostScript name → Type0 font, for ToUnicode at Close
	fontDescriptors       map[string]*fontDescriptor // PostScript name → descriptor, for FontFile2 at Close
	images                map[string]*cachedImage
	svgForms              map[string]*cachedSVGForm
	dimensionCache        map[string]dimensionCacheEntry
	memoForms             map[string]*cachedForm
	gradientShadings      map[string]string // gradient key → shading resource name
	gradientPatterns      map[string]string // gradient key → pattern resource name
	extGStates            map[string]string // graphics-state key → ExtGState resource name
	assetFS               fs.FS
	compressPages         bool
	compressToUnicode     bool
	compressEmbeddedFonts bool
	destinations          map[string]namedDestination
	pendingTargetLinks    []*linkAnnotation
	writeToCleanups       []func()
	profiler              *profile.Profiler
	fontTrace             io.Writer
}

type cachedImage struct {
	image *pdfImage
	name  string
}

type cachedSVGForm struct {
	form   *pdfForm
	name   string
	width  float64
	height float64
}

type dimensionCacheEntry struct {
	width  int
	height int
	err    error
}

type namedDestination struct {
	page *page
	x    float64
	y    float64
}

func NewDocWriter() *DocWriter {
	nextSeq := nextSeqFunc()
	file := newFile()
	pages := newPages(nextSeq(), 0)
	outlines := newOutlines(nextSeq(), 0)
	catalog := newCatalog(nextSeq(), 0, "UseNone", pages, outlines)
	file.body.add(pages, outlines, catalog)
	file.trailer.setRoot(catalog)
	resources := newResources(nextSeq(), 0)
	resources.setProcSet(nameArray("PDF", "Text", "ImageB", "ImageC"))
	file.body.add(resources)
	fontSources := make(font.FontSources, 0, 2)
	selectedFonts := make(map[string]*font.Font)
	fontKeys := make(map[string]string)
	fontEncodings := make(map[string]*fontEncoding)
	return &DocWriter{
		nextSeq:          nextSeq,
		file:             file,
		catalog:          catalog,
		resources:        resources,
		options:          options.Options{},
		fontSources:      fontSources,
		selectedFonts:    selectedFonts,
		fontKeys:         fontKeys,
		fontEncodings:    fontEncodings,
		glyphRecorders:   make(map[string]*glyphRecorder),
		unicodeFonts:     make(map[string]*font.Font),
		cidFonts:         make(map[string]*cidFont),
		type0Fonts:       make(map[string]*type0Font),
		fontDescriptors:  make(map[string]*fontDescriptor),
		images:           make(map[string]*cachedImage),
		svgForms:         make(map[string]*cachedSVGForm),
		dimensionCache:   make(map[string]dimensionCacheEntry),
		memoForms:        make(map[string]*cachedForm),
		gradientShadings: make(map[string]string),
		gradientPatterns: make(map[string]string),
		extGStates:       make(map[string]string),
		destinations:     make(map[string]namedDestination),
	}
}

func (dw *DocWriter) EnableTaggedPDF(value bool) *DocWriter {
	if !value {
		return dw
	}
	if dw.accessibility == nil {
		dw.accessibility = newAccessibilityState(dw)
	}
	return dw
}

func (dw *DocWriter) taggedPDFEnabled() bool {
	return dw.accessibility != nil
}

func (dw *DocWriter) TaggedPDFEnabled() bool {
	return dw.taggedPDFEnabled()
}

func (dw *DocWriter) WithAccessibilityTag(tag string, opts AccessibilityOptions, fn func()) error {
	if dw.curPage != nil {
		return dw.curPage.WithAccessibilityTag(tag, opts, fn)
	}
	dw.EnableTaggedPDF(true)
	elem := dw.accessibility.push(tag, opts)
	if fn != nil {
		fn()
	}
	dw.accessibility.pop(elem)
	return nil
}

func (dw *DocWriter) WithAccessibilityArtifact(fn func()) error {
	return dw.CurPage().WithAccessibilityArtifact(fn)
}

func nextSeqFunc() func() int {
	var nextValue = 0
	return func() int {
		nextValue++
		return nextValue
	}
}

func (dw *DocWriter) AddFont(family string, options options.Options) ([]*font.Font, error) {
	return dw.CurPage().AddFont(family, options)
}

func (dw *DocWriter) AddFontSource(fontSource font.FontSource) {
	dw.fontSources = append(dw.fontSources, fontSource)
	dw.selectedFonts = make(map[string]*font.Font)
}

func (dw *DocWriter) CompressPages(value bool) *DocWriter {
	dw.compressPages = value
	return dw
}

func (dw *DocWriter) CompressToUnicode(value bool) *DocWriter {
	dw.compressToUnicode = value
	return dw
}

func (dw *DocWriter) CompressEmbeddedFonts(value bool) *DocWriter {
	dw.compressEmbeddedFonts = value
	return dw
}

func (dw *DocWriter) CurPage() *PageWriter {
	if dw.curPage == nil {
		return dw.NewPage()
	}
	return dw.curPage
}

func (dw *DocWriter) FontColor() colors.Color {
	return dw.CurPage().FontColor()
}

func (dw *DocWriter) SetAssetFS(assetFS fs.FS) {
	dw.assetFS = assetFS
	dw.dimensionCache = make(map[string]dimensionCacheEntry)
}

func (dw *DocWriter) AssetFS() fs.FS {
	return dw.assetFS
}

func (dw *DocWriter) RegisterWriteToCleanup(fn func()) {
	if fn == nil {
		return
	}
	dw.writeToCleanups = append(dw.writeToCleanups, fn)
}

func (dw *DocWriter) SetProfiler(profiler *profile.Profiler) {
	dw.profiler = profiler
}

func (dw *DocWriter) Profiler() *profile.Profiler {
	return dw.profiler
}

func (dw *DocWriter) runWriteToCleanups() {
	for _, fn := range dw.writeToCleanups {
		if fn != nil {
			fn()
		}
	}
	dw.writeToCleanups = nil
}

func (dw *DocWriter) readImageFile(filename string) ([]byte, error) {
	if filepath.IsAbs(filename) || dw.assetFS == nil {
		return os.ReadFile(filename)
	}
	if !assetpath.Valid(filename) {
		return nil, fmt.Errorf("invalid asset path %q", filename)
	}
	info, err := fs.Stat(dw.assetFS, filename)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return nil, fmt.Errorf("asset path %q is a directory", filename)
	}
	return fs.ReadFile(dw.assetFS, filename)
}

func (dw *DocWriter) openImageFile(filename string) (fs.File, error) {
	var (
		file fs.File
		err  error
	)
	if filepath.IsAbs(filename) || dw.assetFS == nil {
		file, err = os.Open(filename)
	} else {
		if !assetpath.Valid(filename) {
			return nil, fmt.Errorf("invalid asset path %q", filename)
		}
		file, err = dw.assetFS.Open(filename)
	}
	if err != nil {
		return nil, err
	}

	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, err
	}
	if info.IsDir() {
		file.Close()
		return nil, fmt.Errorf("asset path %q is a directory", filename)
	}
	return file, nil
}

func useStandardEncoding(family string) bool {
	switch family {
	case "Symbol", "ZapfDingbats":
		return true
	default:
		return false
	}
}

func (dw *DocWriter) fontKey(f *font.Font, cpi codepage.CodepageIndex) string {
	if f == nil {
		panic("fontKey: No font specified.")
	}
	if !f.HasMetrics() {
		panic("fontKey: font missing metrics.")
	}
	if f.SubType() == "TrueType" {
		return dw.fontKeyUnicode(f)
	}
	name := fmt.Sprintf("%s/%s-%s", f.PostScriptName(), cpi, f.SubType())
	if key, ok := dw.fontKeys[name]; ok {
		return key
	}
	descriptor := newFontDescriptor(
		dw.nextSeq(), 0,
		f.PostScriptName(), f.Family(),
		f.Flags(),
		f.BoundingBox(),
		0, // missingWidth
		f.StemV(),
		0, // stemH
		f.ItalicAngle(),
		f.CapHeight(),
		f.XHeight(),
		f.Ascent(),
		f.Descent(),
		f.Leading(),
		0, 0) // maxWidth, avgWidth
	dw.file.body.add(descriptor)
	key := fmt.Sprintf("F%d", len(dw.fontKeys))
	dw.fontKeys[name] = key
	widths := dw.widthsForFontCodepage(f, cpi)
	dw.file.body.add(widths)
	var font *simpleFont
	switch f.SubType() {
	case "Type1":
		if useStandardEncoding(f.Family()) {
			font = newType1Font(
				dw.nextSeq(), 0,
				f.PostScriptName(),
				32, 255, widths,
				descriptor, nil)
		} else {
			encoding, ok := dw.fontEncodings[cpi.String()]
			if !ok {
				differences := glyphDiffs(codepage.Idx_CP1252, cpi, 32, 255)
				encoding = newFontEncoding(dw.nextSeq(), 0, "WinAnsiEncoding", differences)
				dw.file.body.add(encoding)
				dw.fontEncodings[cpi.String()] = encoding
			}
			font = newType1Font(
				dw.nextSeq(), 0,
				f.PostScriptName(),
				32, 255, widths,
				descriptor, &indirectObjectRef{encoding})
		}
	case "TrueType":
		encoding, ok := dw.fontEncodings[cpi.String()]
		if !ok {
			differences := glyphDiffs(codepage.Idx_CP1252, cpi, 32, 255)
			encoding = newFontEncoding(dw.nextSeq(), 0, "WinAnsiEncoding", differences)
			dw.file.body.add(encoding)
			dw.fontEncodings[cpi.String()] = encoding
		}
		font = newTrueTypeFont(
			dw.nextSeq(), 0,
			f.PostScriptName(),
			32, 255, widths,
			descriptor, &indirectObjectRef{encoding})
	}
	toUnicodeData := toUnicodeCMapData(cpi.Map())
	toUnicodeStream := newStream(dw.nextSeq(), 0, toUnicodeData)
	if dw.compressToUnicode {
		if err := toUnicodeStream.compress(); err != nil {
			panic(err)
		}
	}
	dw.file.body.add(toUnicodeStream)
	font.setToUnicode(&indirectObjectRef{toUnicodeStream})

	dw.file.body.add(font)
	dw.resources.fonts[key] = &indirectObjectRef{font}
	return key
}

// fontKeyUnicode registers a Type0 composite font for the given TrueType or
// CFF-backed OpenType font and returns the PDF resource key (e.g. "F0"). The /W and
// ToUnicode entries are left empty and filled in by flushUnicodeFonts at
// document close, once all glyph usages have been recorded.
func (dw *DocWriter) fontKeyUnicode(f *font.Font) string {
	psName := f.PostScriptName()
	cacheName := psName + "/unicode-Type0"
	if key, ok := dw.fontKeys[cacheName]; ok {
		return key
	}
	descriptor := newFontDescriptor(
		dw.nextSeq(), 0,
		psName, f.Family(),
		f.Flags(),
		f.BoundingBox(),
		0, // missingWidth
		f.StemV(),
		0, // stemH
		f.ItalicAngle(),
		f.CapHeight(),
		f.XHeight(),
		f.Ascent(),
		f.Descent(),
		f.Leading(),
		0, 0) // maxWidth, avgWidth
	dw.file.body.add(descriptor)

	key := fmt.Sprintf("F%d", len(dw.fontKeys))
	dw.fontKeys[cacheName] = key

	// A Type 0 font is a two-stage mapping. Its Encoding maps emitted character
	// codes to descendant-font CIDs; the descendant then maps CIDs to glyphs.
	// TrueType normally uses Identity plus CIDToGIDMap. CID-keyed CFF instead
	// carries its CID-to-glyph relationship in the CFF charset and needs its
	// actual ROS on both PDF font dictionaries.
	cidSubtype := "CIDFontType2"
	systemInfo := font.CIDSystemInfo{Registry: "Adobe", Ordering: "Identity", Supplement: 0}
	cidKeyed := false
	if f.OutlineKind() == "CFF" {
		cidSubtype = "CIDFontType0"
		if info, ok := f.CIDSystemInfo(); ok {
			systemInfo = info
			cidKeyed = true
		}
	}
	cid := newCIDFontWithSystemInfo(dw.nextSeq(), 0, cidSubtype, psName, descriptor, 1000, array{}, systemInfo.Registry, systemInfo.Ordering, systemInfo.Supplement)
	dw.file.body.add(cid)

	t0 := newType0Font(dw.nextSeq(), 0, psName, cid)
	dw.file.body.add(t0)

	dw.resources.fonts[key] = &indirectObjectRef{t0}
	if cidKeyed {
		dw.glyphRecorders[psName] = newCIDKeyedGlyphRecorder(f.CIDForGlyph)
	} else {
		dw.glyphRecorders[psName] = newGlyphRecorder(f.OutlineKind() == "CFF")
	}
	dw.unicodeFonts[psName] = f
	dw.cidFonts[psName] = cid
	dw.type0Fonts[psName] = t0
	dw.fontDescriptors[psName] = descriptor
	return key
}

// subsetTag generates a deterministic 6-character uppercase tag for a font
// subset, per PDF spec §9.6.4. Deriving the tag from stable inputs keeps PDF
// output reproducible for the same document content.
func subsetTag(psName string, glyphIDs []uint16) string {
	h := sha1.New()
	h.Write([]byte(psName))
	var buf [2]byte
	for _, gid := range glyphIDs {
		binary.BigEndian.PutUint16(buf[:], gid)
		h.Write(buf[:])
	}
	sum := h.Sum(nil)
	tag := make([]byte, 6)
	for i := range tag {
		tag[i] = 'A' + (sum[i] % 26)
	}
	return string(tag)
}

// flushUnicodeFonts is called from WriteTo before serialising the PDF.
// It fills in the /W width arrays, ToUnicode CMap streams, and embedded
// subset font streams for every Type0 composite font that was used during
// rendering.
func (dw *DocWriter) flushUnicodeFonts() error {
	defer dw.profiler.Begin("pdf.write.flush_unicode_fonts").End()
	psNames := make([]string, 0, len(dw.glyphRecorders))
	for psName := range dw.glyphRecorders {
		psNames = append(psNames, psName)
	}
	sort.Strings(psNames)

	for _, psName := range psNames {
		gr := dw.glyphRecorders[psName]
		if err := gr.error(); err != nil {
			return fmt.Errorf("font %s: %w", psName, err)
		}
		mapping := gr.mapping()
		glyphIDs := gr.glyphIDs()
		if len(glyphIDs) == 0 {
			continue
		}
		sort.Slice(glyphIDs, func(i, j int) bool { return glyphIDs[i] < glyphIDs[j] })
		f := dw.unicodeFonts[psName]
		cidInfo, cidKeyed := f.CIDSystemInfo()
		upm := f.UnitsPerEm()

		// Width keys belong to the descendant font's CID namespace, whereas
		// ToUnicode keys belong to the emitted character-code namespace.
		cids := gr.cids()
		sort.Slice(cids, func(i, j int) bool { return cids[i] < cids[j] })
		cidWidths := make(map[uint16]int, len(cids))
		cidToGID := make(map[uint16]uint16, len(cids))
		codeToFontCID := make(map[uint16]uint16, len(cids))
		for _, cid := range cids {
			gid := gr.glyphIDForCID(cid)
			widthCID := cid
			if cidKeyed {
				widthCID = gr.fontCIDForCode(cid)
				codeToFontCID[cid] = widthCID
			} else {
				cidToGID[cid] = gid
			}
			w := f.AdvanceWidthForGlyph(gid)
			if upm > 0 {
				w = w * 1000 / upm
			}
			cidWidths[widthCID] = w
		}
		defWidth := mostCommonWidth(cidWidths)
		dw.cidFonts[psName].setDefaultWidth(defWidth)
		dw.cidFonts[psName].setWidths(buildCIDWidthArray(cidWidths, defWidth))

		// Build ToUnicode CMap stream.
		tuData := toUnicodeCMapDataComposite(mapping)
		tuStream := newStream(dw.nextSeq(), 0, tuData)
		if dw.compressToUnicode {
			if err := tuStream.compress(); err != nil {
				panic(err)
			}
		}
		dw.file.body.add(tuStream)
		dw.type0Fonts[psName].setToUnicode(&indirectObjectRef{tuStream})
		if cidKeyed {
			// LeadType allocates compact character codes independently of the
			// font's often sparse CIDs. This CMap joins those namespaces.
			encodingData := codeToCIDCMapData(codeToFontCID, cidSystemInfo{registry: cidInfo.Registry, ordering: cidInfo.Ordering, supplement: cidInfo.Supplement})
			encodingStream := newStream(dw.nextSeq(), 0, encodingData)
			dw.file.body.add(encodingStream)
			dw.type0Fonts[psName].setEncoding(&indirectObjectRef{encodingStream})
		}

		if f.OutlineKind() != "CFF" {
			// CIDFontType2 has no CFF charset, so PDF needs an explicit map from
			// descendant CIDs to subset glyph IDs.
			cidMapStream := newStream(dw.nextSeq(), 0, cidToGIDMapData(cidToGID))
			dw.file.body.add(cidMapStream)
			dw.cidFonts[psName].setCIDToGIDMap(&indirectObjectRef{cidMapStream})
		}

		// Embed a font subset in the descriptor. Some fonts, notably CID-keyed
		// CFF OpenType fonts, are usable for glyph lookup but cannot yet be
		// subset by our local subsetter; for those, embed the full font file so
		// viewers do not substitute a mismatched local font.
		span := dw.profiler.Begin("pdf.font.subset")
		subsetData, pdfFontSubtype, err := f.PDFSubsetBytes(glyphIDs)
		span.End()
		fontData := subsetData
		subsetEmbedded := err == nil
		if err != nil && cidKeyed {
			return fmt.Errorf("subset CID-keyed CFF font %s (%s): %w", psName, f.Filename(), err)
		}
		if err != nil {
			fontData = f.Bytes()
		}
		if len(fontData) > 0 {
			fontStream := newStream(dw.nextSeq(), 0, fontData)
			fontStream.setLength1(len(fontData))
			if f.OutlineKind() == "CFF" {
				if pdfFontSubtype == "" {
					pdfFontSubtype = "OpenType"
				}
				fontStream.setSubtype(pdfFontSubtype)
			}
			if cidKeyed && subsetEmbedded {
				// CIDSet is also indexed by font CID, not by emitted code or GID.
				fontCIDs := make([]uint16, 0, len(codeToFontCID)+1)
				fontCIDs = append(fontCIDs, 0)
				seen := map[uint16]bool{0: true}
				for _, cid := range codeToFontCID {
					if !seen[cid] {
						seen[cid] = true
						fontCIDs = append(fontCIDs, cid)
					}
				}
				cidSetStream := newStream(dw.nextSeq(), 0, cidSetData(fontCIDs))
				dw.file.body.add(cidSetStream)
				dw.fontDescriptors[psName].setCIDSet(&indirectObjectRef{cidSetStream})
			}
			if dw.fontTrace != nil {
				fmt.Fprintf(dw.fontTrace, "font embedded postscript=%q outline=%q cid_keyed=%t glyphs=%d subset_bytes=%d source_bytes=%d\n", psName, f.OutlineKind(), cidKeyed, len(glyphIDs), len(fontData), len(f.Bytes()))
			}
			if dw.compressEmbeddedFonts {
				if err := fontStream.compress(); err != nil {
					panic(err)
				}
			}
			dw.file.body.add(fontStream)
			if f.OutlineKind() == "CFF" {
				dw.fontDescriptors[psName].setFontFile3(&indirectObjectRef{fontStream})
			} else {
				dw.fontDescriptors[psName].setFontFile2(&indirectObjectRef{fontStream})
			}

			if !subsetEmbedded {
				continue
			}
			// Apply the 6-char subset tag to all three name occurrences only
			// when the embedded font file actually is a subset.
			taggedName := subsetTag(psName, glyphIDs) + "+" + psName
			dw.fontDescriptors[psName].dict["FontName"] = name(taggedName)
			dw.cidFonts[psName].setBaseFont(taggedName)
			dw.type0Fonts[psName].setBaseFont(taggedName)
		}
	}
	return nil
}

func (dw *DocWriter) Fonts() []*font.Font {
	return dw.CurPage().Fonts()
}

func (dw *DocWriter) FontSize() float64 {
	return dw.CurPage().FontSize()
}

func (dw *DocWriter) FontSources() font.FontSources {
	return dw.fontSources
}

func (dw *DocWriter) SetFontTrace(w io.Writer) {
	dw.fontTrace = w
}

func (dw *DocWriter) FontTrace() io.Writer {
	return dw.fontTrace
}

func (dw *DocWriter) ShareFontSelectionCacheFrom(other *DocWriter) {
	if other == nil {
		return
	}
	if other.selectedFonts == nil {
		other.selectedFonts = make(map[string]*font.Font)
	}
	dw.selectedFonts = other.selectedFonts
}

func (dw *DocWriter) selectFont(family string, opts options.Options) (*font.Font, error) {
	key, cacheable := fontSelectionCacheKey(family, opts)
	if cacheable {
		if cached := dw.selectedFonts[key]; cached != nil {
			return cached.Clone(), nil
		}
	}
	match := opts.StringDefault("match", "exact")
	var selected *font.Font
	var err error
	if match == "nearest" {
		selected, err = font.NewClosest(family, opts, dw.fontSources)
	} else {
		selected, err = font.New(family, opts, dw.fontSources)
	}
	if err != nil {
		dw.traceFontSelection(family, opts, nil, err)
		return nil, err
	}
	dw.traceFontSelection(family, opts, selected, nil)
	if cacheable {
		if dw.selectedFonts == nil {
			dw.selectedFonts = make(map[string]*font.Font)
		}
		dw.selectedFonts[key] = selected.Clone()
	}
	return selected, nil
}

func (dw *DocWriter) traceFontSelection(family string, opts options.Options, selected *font.Font, err error) {
	if dw.fontTrace == nil {
		return
	}
	status := "selected"
	weight := opts.StringDefault("weight", "")
	style := opts.StringDefault("style", "")
	match := opts.StringDefault("match", "exact")
	ranges := ""
	if rawRanges, ok := opts["ranges"]; ok {
		ranges = fmt.Sprintf(" ranges=%v", rawRanges)
	}
	if err != nil {
		fmt.Fprintf(dw.fontTrace, "font %s match=%s family=%q weight=%q style=%q%s error=%v\n", "error", match, family, weight, style, ranges, err)
		return
	}
	fmt.Fprintf(dw.fontTrace, "font %s match=%s family=%q weight=%q style=%q%s -> family=%q full=%q postscript=%q file=%q\n",
		status, match, family, weight, style, ranges, selected.Family(), selected.FullName(), selected.PostScriptName(), selected.Filename())
}

func fontSelectionCacheKey(family string, opts options.Options) (string, bool) {
	var b strings.Builder
	b.WriteString(family)
	b.WriteByte(0)
	b.WriteString(opts.StringDefault("weight", ""))
	b.WriteByte(0)
	b.WriteString(opts.StringDefault("style", ""))
	b.WriteByte(0)
	b.WriteString(opts.StringDefault("match", "exact"))
	b.WriteByte(0)
	b.WriteString(fmt.Sprintf("%g", opts.FloatDefault("relative_size", 100)))
	b.WriteByte(0)
	if ranges, ok := opts["ranges"]; ok {
		switch ranges := ranges.(type) {
		case nil:
		case []string:
			b.WriteString("[]string")
			for _, value := range ranges {
				b.WriteByte(0)
				b.WriteString(value)
			}
		case font.RuneSet:
			rangesKey, ok := runeSetCacheKey(ranges)
			if !ok {
				return "", false
			}
			b.WriteString(rangesKey)
		default:
			return "", false
		}
	}
	return b.String(), true
}

func runeSetCacheKey(ranges font.RuneSet) (string, bool) {
	value := reflect.ValueOf(ranges)
	if !value.IsValid() {
		return "", true
	}
	if value.Kind() != reflect.Slice {
		return "", false
	}
	var b strings.Builder
	b.WriteString(value.Type().String())
	for i := 0; i < value.Len(); i++ {
		b.WriteByte(0)
		b.WriteString(fmt.Sprintf("%#v", value.Index(i).Interface()))
	}
	return b.String(), true
}

func (dw *DocWriter) FontStyle() string {
	return dw.CurPage().FontStyle()
}

func (dw *DocWriter) indexOfPage(pw *PageWriter) int {
	for i, v := range dw.pages {
		if v == pw {
			return i
		}
	}
	return -1
}

func (dw *DocWriter) inPage() bool {
	return dw.curPage != nil
}

func (dw *DocWriter) registerDestination(name string, page *page, x, y float64) {
	if name == "" || page == nil {
		return
	}
	if _, exists := dw.destinations[name]; exists {
		return
	}
	dw.destinations[name] = namedDestination{page: page, x: x, y: y}
}

func (dw *DocWriter) registerPendingTargetLink(a *linkAnnotation) {
	if a == nil || a.targetName == "" {
		return
	}
	dw.pendingTargetLinks = append(dw.pendingTargetLinks, a)
}

func (dw *DocWriter) RegisterDestination(name string, x, y float64) {
	dw.CurPage().RegisterDestination(name, x, y)
}

func (dw *DocWriter) AddURILink(x, y, width, height float64, uri string) error {
	return dw.CurPage().AddURILink(x, y, width, height, uri)
}

func (dw *DocWriter) AddTargetLink(x, y, width, height float64, target string) error {
	return dw.CurPage().AddTargetLink(x, y, width, height, target)
}

func (dw *DocWriter) resolveTargetLinks() error {
	if len(dw.pendingTargetLinks) == 0 {
		return nil
	}
	var unresolved []string
	seen := make(map[string]struct{})
	for _, link := range dw.pendingTargetLinks {
		dest, ok := dw.destinations[link.targetName]
		if !ok {
			if _, added := seen[link.targetName]; !added {
				seen[link.targetName] = struct{}{}
				unresolved = append(unresolved, link.targetName)
			}
			continue
		}
		link.setDestination(dest.page, dest.x, dest.y)
	}
	if len(unresolved) > 0 {
		sort.Strings(unresolved)
		return fmt.Errorf("unresolved pdf link target(s): %s", strings.Join(unresolved, ", "))
	}
	return nil
}

func (dw *DocWriter) insertPage(pw *PageWriter, index int) {
	pages := make([]*PageWriter, 0, len(dw.pages)+1)
	pages = append(pages, dw.pages[:index+1]...)
	pages = append(pages, pw)
	pages = append(pages, dw.pages[index+1:]...)
	dw.pages = pages
}

func (dw *DocWriter) LineCapStyle() LineCapStyle {
	return dw.CurPage().LineCapStyle()
}

func (dw *DocWriter) LineJoinStyle() LineJoinStyle {
	return dw.CurPage().LineJoinStyle()
}

func (dw *DocWriter) LineColor() colors.Color {
	return dw.CurPage().LineColor()
}

func (dw *DocWriter) LineDashPattern() string {
	return dw.CurPage().LineDashPattern()
}

func (dw *DocWriter) LineSpacing() float64 {
	return dw.CurPage().LineSpacing()
}

func (dw *DocWriter) MiterLimit() float64 {
	return dw.CurPage().MiterLimit()
}

func (dw *DocWriter) LineTo(x, y float64) {
	dw.CurPage().LineTo(x, y)
}

func (dw *DocWriter) Line(x, y, angle, length float64) {
	dw.CurPage().Line(x, y, angle, length)
}

func (dw *DocWriter) LineWidth(units string) float64 {
	return dw.CurPage().LineWidth(units)
}

func (dw *DocWriter) Loc() (x, y float64) {
	return dw.CurPage().Loc()
}

func (dw *DocWriter) MoveTo(x, y float64) {
	dw.CurPage().MoveTo(x, y)
}

func (dw *DocWriter) PointsForCircle(x, y, r float64) []Location {
	return dw.CurPage().PointsForCircle(x, y, r)
}

func (dw *DocWriter) Circle(x, y, r float64, border, fill, reverse bool) error {
	return dw.CurPage().Circle(x, y, r, border, fill, reverse)
}

func (dw *DocWriter) DrawClosedShape(shape ClosedShape, border, fill bool) error {
	return dw.CurPage().DrawClosedShape(shape, border, fill)
}

func (dw *DocWriter) CirclePath(x, y, r float64, reverse bool) error {
	return dw.CurPage().CirclePath(x, y, r, reverse)
}

func (dw *DocWriter) PointsForEllipse(x, y, rx, ry float64) []Location {
	return dw.CurPage().PointsForEllipse(x, y, rx, ry)
}

func (dw *DocWriter) Ellipse(x, y, rx, ry float64, border, fill, reverse bool) error {
	return dw.CurPage().Ellipse(x, y, rx, ry, border, fill, reverse)
}

func (dw *DocWriter) EllipsePath(x, y, rx, ry float64, reverse bool) error {
	return dw.CurPage().EllipsePath(x, y, rx, ry, reverse)
}

func (dw *DocWriter) PointsForArc(x, y, r, startAngle, endAngle float64) []Location {
	return dw.CurPage().PointsForArc(x, y, r, startAngle, endAngle)
}

func (dw *DocWriter) Arc(x, y, r, startAngle, endAngle float64, moveToStart bool) error {
	return dw.CurPage().Arc(x, y, r, startAngle, endAngle, moveToStart)
}

func (dw *DocWriter) Pie(x, y, r, startAngle, endAngle float64, border, fill, reverse bool) error {
	return dw.CurPage().Pie(x, y, r, startAngle, endAngle, border, fill, reverse)
}

func (dw *DocWriter) Arch(x, y, r1, r2, startAngle, endAngle float64, border, fill, reverse bool) error {
	return dw.CurPage().Arch(x, y, r1, r2, startAngle, endAngle, border, fill, reverse)
}

func (dw *DocWriter) PointsForPolygon(x, y, r float64, sides int, rotation float64) []Location {
	return dw.CurPage().PointsForPolygon(x, y, r, sides, rotation)
}

func (dw *DocWriter) Polygon(x, y, r float64, sides int, border, fill, reverse bool, rotation float64) error {
	return dw.CurPage().Polygon(x, y, r, sides, border, fill, reverse, rotation)
}

func (dw *DocWriter) PolygonPath(x, y, r float64, sides int, reverse bool, rotation float64) error {
	return dw.CurPage().PolygonPath(x, y, r, sides, reverse, rotation)
}

func (dw *DocWriter) Star(x, y, r1, r2 float64, points int, border, fill, reverse bool, rotation float64) error {
	return dw.CurPage().Star(x, y, r1, r2, points, border, fill, reverse, rotation)
}

func (dw *DocWriter) StarPath(x, y, r1, r2 float64, points int, reverse bool, rotation float64) error {
	return dw.CurPage().StarPath(x, y, r1, r2, points, reverse, rotation)
}

func (dw *DocWriter) NewPage() *PageWriter {
	if dw.curPage == nil {
		return dw.NewPageWithOptions(options.Options{})
	}
	return dw.NewPageAfter(dw.curPage)
}

func (dw *DocWriter) NewPageAfter(pw *PageWriter) *PageWriter {
	var i int
	if pw == nil {
		i = len(dw.pages)
	} else {
		i = dw.indexOfPage(pw)
	}
	if i >= 0 {
		dw.curPage = clonePageWriter(pw)
		dw.insertPage(dw.curPage, i)
		return dw.curPage
	}
	return nil
}

func (dw *DocWriter) NewPageWithOptions(options options.Options) *PageWriter {
	dw.curPage = newPageWriter(dw, dw.options.Merge(options))
	dw.pages = append(dw.pages, dw.curPage)
	return dw.curPage
}

func (dw *DocWriter) PagesAcross() int {
	if dw.pagesAcross < 1 {
		return 1
	}
	return dw.pagesAcross
}

func (dw *DocWriter) PagesDown() int {
	if dw.pagesDown < 1 {
		return 1
	}
	return dw.pagesDown
}

func (dw *DocWriter) PageHeight() float64 {
	return dw.CurPage().PageHeight()
}

func (dw *DocWriter) PagesUp() int {
	return dw.PagesAcross() * dw.PagesDown()
}

func (dw *DocWriter) PageWidth() float64 {
	return dw.CurPage().PageWidth()
}

func (dw *DocWriter) Print(text string) (err error) {
	return dw.CurPage().Print(text)
}

func (dw *DocWriter) ClipRichText(text *rich_text.RichText, fn func()) error {
	return dw.CurPage().ClipRichText(text, fn)
}

func (dw *DocWriter) FillStrokeClipRichText(text *rich_text.RichText, fn func()) error {
	return dw.CurPage().FillStrokeClipRichText(text, fn)
}

func (dw *DocWriter) ClipText(text string, fn func()) error {
	return dw.CurPage().ClipText(text, fn)
}

func (dw *DocWriter) FillStrokeClipText(text string, fn func()) error {
	return dw.CurPage().FillStrokeClipText(text, fn)
}

func (dw *DocWriter) PrintParagraph(para []*rich_text.RichText, options options.Options) {
	dw.CurPage().PrintParagraph(para, options)
}

func (dw *DocWriter) PrintRichText(text *rich_text.RichText) {
	dw.CurPage().PrintRichText(text)
}

func (dw *DocWriter) PrintWithOptions(text string, options options.Options) (err error) {
	return dw.CurPage().PrintWithOptions(text, options)
}

func (dw *DocWriter) ImageDimensions(data []byte) (width, height int, err error) {
	if dw.profiler != nil {
		defer dw.profiler.Begin("pdf.image.dimensions").End()
	}
	return imageDimensions(data)
}

func (dw *DocWriter) SVGDimensions(data []byte) (width, height int, err error) {
	if dw.profiler != nil {
		defer dw.profiler.Begin("pdf.svg.dimensions").End()
	}
	return svgDimensions(data)
}

func (dw *DocWriter) ImageDimensionsFromFile(filename string) (width, height int, err error) {
	if cached, ok := dw.dimensionCache[filename]; ok {
		return cached.width, cached.height, cached.err
	}
	if dw.profiler != nil {
		defer dw.profiler.Begin("pdf.image.dimensions_file").End()
	}
	file, err := dw.openImageFile(filename)
	if err != nil {
		dw.dimensionCache[filename] = dimensionCacheEntry{err: err}
		return 0, 0, err
	}
	defer file.Close()
	info, err := imageInfoFromReader(file)
	if err != nil {
		dw.dimensionCache[filename] = dimensionCacheEntry{err: err}
		return 0, 0, err
	}
	dw.dimensionCache[filename] = dimensionCacheEntry{width: info.width, height: info.height}
	return info.width, info.height, nil
}

func (dw *DocWriter) SVGDimensionsFromFile(filename string) (width, height int, err error) {
	if cached, ok := dw.dimensionCache[filename]; ok {
		return cached.width, cached.height, cached.err
	}
	if dw.profiler != nil {
		defer dw.profiler.Begin("pdf.svg.dimensions_file").End()
	}
	file, err := dw.openImageFile(filename)
	if err != nil {
		dw.dimensionCache[filename] = dimensionCacheEntry{err: err}
		return 0, 0, err
	}
	defer file.Close()
	width, height, err = svgDimensionsFromReader(file)
	dw.dimensionCache[filename] = dimensionCacheEntry{width: width, height: height, err: err}
	return width, height, err
}

func (dw *DocWriter) Path(fn func()) error {
	return dw.CurPage().Path(fn)
}

func (dw *DocWriter) Fill() error {
	return dw.CurPage().Fill()
}

func (dw *DocWriter) Stroke() error {
	return dw.CurPage().Stroke()
}

func (dw *DocWriter) FillAndStroke() error {
	return dw.CurPage().FillAndStroke()
}

func (dw *DocWriter) Clip(fn func()) error {
	return dw.CurPage().Clip(fn)
}

func (dw *DocWriter) ClosedShapeBounds(shape ClosedShape) (Bounds, error) {
	return dw.CurPage().ClosedShapeBounds(shape)
}

func (dw *DocWriter) ClipClosedShape(shape ClosedShape, fn func()) error {
	return dw.CurPage().ClipClosedShape(shape, fn)
}

func (dw *DocWriter) Rotate(angle, x, y float64, fn func()) error {
	return dw.CurPage().Rotate(angle, x, y, fn)
}

func (dw *DocWriter) Scale(x, y, scaleX, scaleY float64, fn func()) error {
	return dw.CurPage().Scale(x, y, scaleX, scaleY, fn)
}

func (dw *DocWriter) Rectangle(x, y, width, height float64, border bool, fill bool) {
	dw.CurPage().Rectangle(x, y, width, height, border, fill)
}

func (dw *DocWriter) Rectangle2(x, y, width, height float64, border bool, fill bool, corners []float64, path, reverse bool) {
	dw.CurPage().Rectangle2(x, y, width, height, border, fill, corners, path, reverse)
}

func (dw *DocWriter) PrintImage(data []byte, x, y float64, width, height *float64) (actualWidth, actualHeight float64, err error) {
	return dw.CurPage().PrintImage(data, x, y, width, height)
}

func (dw *DocWriter) PaintImage(data []byte, x, y, width, height, opacity float64) error {
	return dw.CurPage().PaintImage(data, x, y, width, height, opacity)
}

func (dw *DocWriter) PrintImageFile(filename string, x, y float64, width, height *float64) (actualWidth, actualHeight float64, err error) {
	return dw.CurPage().PrintImageFile(filename, x, y, width, height)
}

func (dw *DocWriter) PaintImageFile(filename string, x, y, width, height, opacity float64) error {
	return dw.CurPage().PaintImageFile(filename, x, y, width, height, opacity)
}

func (dw *DocWriter) PrintSVG(data []byte, x, y float64, width, height *float64) (actualWidth, actualHeight float64, err error) {
	return dw.CurPage().PrintSVG(data, x, y, width, height)
}

func (dw *DocWriter) PrintSVGFile(filename string, x, y float64, width, height *float64) (actualWidth, actualHeight float64, err error) {
	return dw.CurPage().PrintSVGFile(filename, x, y, width, height)
}

func (dw *DocWriter) loadImage(data []byte, key string) (*pdfImage, string, error) {
	if cached, ok := dw.images[key]; ok {
		return cached.image, cached.name, nil
	}
	if dw.profiler != nil {
		defer dw.profiler.Begin("pdf.image.load").End()
	}
	decoded, err := decodeImage(data)
	if err != nil {
		return nil, "", err
	}
	components := decoded.info.components
	if len(decoded.alphaData) > 0 && (components == imageComponentsGrayAlpha || components == imageComponentsRGBA) {
		components--
	}
	colorSpace, err := imageColorSpace(components)
	if err != nil {
		return nil, "", err
	}
	image := newPDFImage(dw.nextSeq(), 0, decoded.data)
	image.setDimensions(decoded.info.width, decoded.info.height)
	image.setBitsPerComponent(decoded.info.bitsPerComponent)
	image.setColorSpace(colorSpace)
	if decoded.filter == "FlateDecode" {
		if err := image.compress(); err != nil {
			return nil, "", err
		}
	} else if decoded.filter != "" {
		image.setFilter(decoded.filter)
	}
	if len(decoded.alphaData) > 0 {
		mask := newPDFImage(dw.nextSeq(), 0, decoded.alphaData)
		mask.setDimensions(decoded.info.width, decoded.info.height)
		mask.setBitsPerComponent(decoded.info.bitsPerComponent)
		mask.setColorSpace("DeviceGray")
		if err := mask.compress(); err != nil {
			return nil, "", err
		}
		dw.file.body.add(mask)
		image.setSMask(&indirectObjectRef{mask})
	}

	name := fmt.Sprintf("Im%d", len(dw.images))
	dw.file.body.add(image)
	dw.resources.setXObject(name, &indirectObjectRef{image})
	dw.images[key] = &cachedImage{image: image, name: name}
	return image, name, nil
}

func (dw *DocWriter) loadSVGForm(data []byte, key string, renderOptions options.Options) (*cachedSVGForm, error) {
	if cached, ok := dw.svgForms[key]; ok {
		return cached, nil
	}
	if dw.profiler != nil {
		defer dw.profiler.Begin("pdf.svg.load_form").End()
	}

	doc, warnings, err := svg.Parse(data)
	if err != nil {
		return nil, err
	}
	logSVGWarnings(warnings)

	form, err := dw.newSVGForm(doc, renderOptions)
	if err != nil {
		return nil, err
	}

	name := fmt.Sprintf("Fm%d", len(dw.svgForms))
	dw.file.body.add(form.resources, form.form)
	dw.resources.setXObject(name, &indirectObjectRef{form.form})

	cached := &cachedSVGForm{
		form:   form.form,
		name:   name,
		width:  doc.Width,
		height: doc.Height,
	}
	dw.svgForms[key] = cached
	return cached, nil
}

func (dw *DocWriter) loadMemoForm(cacheKey string, base *PageWriter, canvasWidthPts, canvasHeightPts float64, render func(*PageWriter) error) (*cachedForm, error) {
	if cached, ok := dw.memoForms[cacheKey]; ok {
		return cached, nil
	}

	content := newContentWriterFromPage(base, canvasWidthPts, canvasHeightPts)
	if err := render(content); err != nil {
		return nil, err
	}
	content.endText()
	content.endGraph()

	formResources := dw.resources.clone(dw.nextSeq(), 0)
	form := newPDFForm(dw.nextSeq(), 0, content.stream.Bytes())
	form.setBBox(canvasWidthPts, canvasHeightPts)
	form.setResources(formResources)
	if dw.compressPages {
		if err := form.compress(); err != nil {
			return nil, err
		}
	}

	name := fmt.Sprintf("Mf%d", len(dw.memoForms))
	dw.file.body.add(formResources, form)
	dw.resources.setXObject(name, &indirectObjectRef{form})

	cached := &cachedForm{
		form:   form,
		name:   name,
		width:  canvasWidthPts,
		height: canvasHeightPts,
	}
	dw.memoForms[cacheKey] = cached
	return cached, nil
}

func (dw *DocWriter) MemoizeForm(key string, x, y, width, height float64, render func(*PageWriter) error) error {
	return dw.CurPage().MemoizeForm(key, x, y, width, height, render)
}

// MemoizeFormOnCanvas captures a form once on the supplied logical canvas size,
// then places that form at the requested destination size.
func (dw *DocWriter) MemoizeFormOnCanvas(key string, x, y, width, height, canvasWidth, canvasHeight float64, render func(*PageWriter) error) error {
	return dw.CurPage().MemoizeFormOnCanvas(key, x, y, width, height, canvasWidth, canvasHeight, render)
}

func (dw *DocWriter) gradientKey(prefix string, coords []float64, stops []GradientStop) string {
	h := sha1.New()
	fmt.Fprintf(h, "%s:", prefix)
	for _, c := range coords {
		fmt.Fprintf(h, "%s,", g(c))
	}
	for _, s := range stops {
		fmt.Fprintf(h, "%s:%06x,", g(s.Position), int32(s.Color))
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (dw *DocWriter) alphaGradientKey(prefix string, coords []float64, stops []alphaGradientStop) string {
	h := sha1.New()
	fmt.Fprintf(h, "%s:", prefix)
	for _, c := range coords {
		fmt.Fprintf(h, "%s,", g(c))
	}
	for _, s := range stops {
		fmt.Fprintf(h, "%s:%s,", g(s.Position), g(s.Alpha))
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (dw *DocWriter) extGStateKey(fillAlpha, strokeAlpha float64, blendMode string, maskForm seqGen, maskSubtype string, maskBackdrop []float64) string {
	h := sha1.New()
	fmt.Fprintf(h, "fill=%s;stroke=%s;bm=%s;mask=%s;", g(fillAlpha), g(strokeAlpha), blendMode, maskSubtype)
	if maskForm != nil {
		fmt.Fprintf(h, "form=%d/%d", maskForm.Seq(), maskForm.Gen())
	}
	if len(maskBackdrop) > 0 {
		fmt.Fprintf(h, ";bc=")
		for _, value := range maskBackdrop {
			fmt.Fprintf(h, "%s,", g(value))
		}
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (dw *DocWriter) registerExtGState(fillAlpha, strokeAlpha float64, blendMode string, maskForm seqGen, maskSubtype string, maskBackdrop []float64) (string, error) {
	key := dw.extGStateKey(fillAlpha, strokeAlpha, blendMode, maskForm, maskSubtype, maskBackdrop)
	if name, ok := dw.extGStates[key]; ok {
		return name, nil
	}
	gs := newExtGState(dw.nextSeq(), 0)
	if fillAlpha < 1 {
		gs.setFillAlpha(fillAlpha)
	}
	if strokeAlpha < 1 {
		gs.setStrokeAlpha(strokeAlpha)
	}
	if blendMode != "" && blendMode != "Normal" {
		gs.setBlendMode(blendMode)
	}
	if maskForm != nil {
		if maskSubtype == "" {
			maskSubtype = "Luminosity"
		}
		gs.setSoftMask(maskForm, maskSubtype, maskBackdrop)
	}
	dw.file.body.add(gs)
	name := fmt.Sprintf("GS%d", len(dw.extGStates))
	dw.resources.setExtGState(name, &indirectObjectRef{gs})
	dw.extGStates[key] = name
	return name, nil
}

func (dw *DocWriter) registerLinearGradient(lg *LinearGradient, u *units, pageHeight float64) (string, error) {
	x0, y0 := u.toPts(lg.X0), pageHeight-u.toPts(lg.Y0)
	x1, y1 := u.toPts(lg.X1), pageHeight-u.toPts(lg.Y1)
	coords := []float64{x0, y0, x1, y1}
	return dw.registerGradientPattern("axial", coords, lg.Stops, func(fn seqGen) genWriter {
		return newAxialShading(dw.nextSeq(), 0, x0, y0, x1, y1, fn)
	})
}

func (dw *DocWriter) registerRadialGradient(rg *RadialGradient, u *units, pageHeight float64) (string, error) {
	x0, y0 := u.toPts(rg.X0), pageHeight-u.toPts(rg.Y0)
	x1, y1 := u.toPts(rg.X1), pageHeight-u.toPts(rg.Y1)
	r0, r1 := u.toPts(rg.R0), u.toPts(rg.R1)
	coords := []float64{x0, y0, r0, x1, y1, r1}
	return dw.registerGradientPattern("radial", coords, rg.Stops, func(fn seqGen) genWriter {
		return newRadialShading(dw.nextSeq(), 0, x0, y0, r0, x1, y1, r1, fn)
	})
}

func (dw *DocWriter) registerGradientPattern(prefix string, coords []float64, stops []GradientStop, makeShading func(seqGen) genWriter) (string, error) {
	key := dw.gradientKey(prefix, coords, stops)
	if patName, ok := dw.gradientPatterns[key]; ok {
		return patName, nil
	}
	fn, fnObjs := buildGradientFunction(dw.nextSeq, stops)
	for _, obj := range fnObjs {
		dw.file.body.add(obj)
	}
	sh := makeShading(fn)
	shObj := sh.(seqGen)
	dw.file.body.add(sh)
	shName := fmt.Sprintf("Sh%d", len(dw.gradientShadings))
	dw.resources.setShading(shName, &indirectObjectRef{shObj})
	dw.gradientShadings[key] = shName

	pat := newShadingPattern(dw.nextSeq(), 0, shObj)
	dw.file.body.add(pat)
	patName := fmt.Sprintf("P%d", len(dw.gradientPatterns))
	dw.resources.setPattern(patName, &indirectObjectRef{pat})
	dw.gradientPatterns[key] = patName
	return patName, nil
}

func (dw *DocWriter) registerLinearShading(lg *LinearGradient, u *units, pageHeight float64) (string, error) {
	x0, y0 := u.toPts(lg.X0), pageHeight-u.toPts(lg.Y0)
	x1, y1 := u.toPts(lg.X1), pageHeight-u.toPts(lg.Y1)
	coords := []float64{x0, y0, x1, y1}
	return dw.registerShadingOnly("axial", coords, lg.Stops, func(fn seqGen) genWriter {
		return newAxialShading(dw.nextSeq(), 0, x0, y0, x1, y1, fn)
	})
}

func (dw *DocWriter) registerLinearAlphaShading(lg *LinearGradient, stops []alphaGradientStop, u *units, pageHeight float64) (string, error) {
	x0, y0 := u.toPts(lg.X0), pageHeight-u.toPts(lg.Y0)
	x1, y1 := u.toPts(lg.X1), pageHeight-u.toPts(lg.Y1)
	coords := []float64{x0, y0, x1, y1}
	key := dw.alphaGradientKey("alpha_axial", coords, stops)
	if shName, ok := dw.gradientShadings[key]; ok {
		return shName, nil
	}
	fn, fnObjs := buildAlphaGradientFunction(dw.nextSeq, stops)
	for _, obj := range fnObjs {
		dw.file.body.add(obj)
	}
	sh := newGrayAxialShading(dw.nextSeq(), 0, x0, y0, x1, y1, fn)
	dw.file.body.add(sh)
	shName := fmt.Sprintf("Sh%d", len(dw.gradientShadings))
	dw.resources.setShading(shName, &indirectObjectRef{sh})
	dw.gradientShadings[key] = shName
	return shName, nil
}

func (dw *DocWriter) registerRadialShading(rg *RadialGradient, u *units, pageHeight float64) (string, error) {
	x0, y0 := u.toPts(rg.X0), pageHeight-u.toPts(rg.Y0)
	x1, y1 := u.toPts(rg.X1), pageHeight-u.toPts(rg.Y1)
	r0, r1 := u.toPts(rg.R0), u.toPts(rg.R1)
	coords := []float64{x0, y0, r0, x1, y1, r1}
	return dw.registerShadingOnly("radial", coords, rg.Stops, func(fn seqGen) genWriter {
		return newRadialShading(dw.nextSeq(), 0, x0, y0, r0, x1, y1, r1, fn)
	})
}

func (dw *DocWriter) registerShadingOnly(prefix string, coords []float64, stops []GradientStop, makeShading func(seqGen) genWriter) (string, error) {
	key := dw.gradientKey(prefix+"_sh", coords, stops)
	if shName, ok := dw.gradientShadings[key]; ok {
		return shName, nil
	}
	fn, fnObjs := buildGradientFunction(dw.nextSeq, stops)
	for _, obj := range fnObjs {
		dw.file.body.add(obj)
	}
	sh := makeShading(fn)
	shObj := sh.(seqGen)
	dw.file.body.add(sh)
	shName := fmt.Sprintf("Sh%d", len(dw.gradientShadings))
	dw.resources.setShading(shName, &indirectObjectRef{shObj})
	dw.gradientShadings[key] = shName
	return shName, nil
}

func (dw *DocWriter) SetFillLinearGradient(lg *LinearGradient) error {
	return dw.CurPage().SetFillLinearGradient(lg)
}

func (dw *DocWriter) SetFillRadialGradient(rg *RadialGradient) error {
	return dw.CurPage().SetFillRadialGradient(rg)
}

func (dw *DocWriter) ClearFillGradient() {
	dw.CurPage().ClearFillGradient()
}

func (dw *DocWriter) SetLineLinearGradient(lg *LinearGradient) error {
	return dw.CurPage().SetLineLinearGradient(lg)
}

func (dw *DocWriter) SetLineRadialGradient(rg *RadialGradient) error {
	return dw.CurPage().SetLineRadialGradient(rg)
}

func (dw *DocWriter) ClearLineGradient() {
	dw.CurPage().ClearLineGradient()
}

func (dw *DocWriter) PaintLinearGradient(lg *LinearGradient) error {
	return dw.CurPage().PaintLinearGradient(lg)
}

func (dw *DocWriter) PaintRadialGradient(rg *RadialGradient) error {
	return dw.CurPage().PaintRadialGradient(rg)
}

func (dw *DocWriter) PaintSweepBand(sb *SweepBand) error {
	return dw.CurPage().PaintSweepBand(sb)
}

func (dw *DocWriter) PaintSweepArc(x, y, innerRadius, outerRadius float64, segments []SweepBandSegment) error {
	return dw.CurPage().PaintSweepArc(x, y, innerRadius, outerRadius, segments)
}

func (dw *DocWriter) ResetFonts() {
	dw.CurPage().ResetFonts()
}

func (dw *DocWriter) MeasureText(text string) (TextMetrics, error) {
	return dw.CurPage().MeasureText(text)
}

func (dw *DocWriter) SetFillColor(color any) (prev colors.Color) {
	return dw.CurPage().SetFillColor(color)
}

func (dw *DocWriter) SetFont(name string, size float64, options options.Options) ([]*font.Font, error) {
	return dw.CurPage().SetFont(name, size, options)
}

func (dw *DocWriter) SetFontColor(color any) (lastColor colors.Color) {
	return dw.CurPage().SetFontColor(color)
}

func (dw *DocWriter) SetFontSize(size float64) (prev float64) {
	return dw.CurPage().SetFontSize(size)
}

func (dw *DocWriter) SetFontStyle(style string) (prev string, err error) {
	return dw.CurPage().SetFontStyle(style)
}

func (dw *DocWriter) SetLineColor(color colors.Color) (prev colors.Color) {
	return dw.CurPage().SetLineColor(color)
}

func (dw *DocWriter) SetLineJoinStyle(lineJoinStyle LineJoinStyle) (prev LineJoinStyle) {
	return dw.CurPage().SetLineJoinStyle(lineJoinStyle)
}

func (dw *DocWriter) SetLineDashPattern(lineDashPattern string) (prev string) {
	return dw.CurPage().SetLineDashPattern(lineDashPattern)
}

func (dw *DocWriter) SetLineSpacing(lineSpacing float64) (prev float64) {
	return dw.CurPage().SetLineSpacing(lineSpacing)
}

func (dw *DocWriter) SetMiterLimit(limit float64) (prev float64) {
	return dw.CurPage().SetMiterLimit(limit)
}

func (dw *DocWriter) SetStrikeout(strikeout bool) (prev bool) {
	return dw.CurPage().SetStrikeout(strikeout)
}

func (dw *DocWriter) SetLineWidth(width float64, units string) (prev float64) {
	return dw.CurPage().SetLineWidth(width, units)
}

func (dw *DocWriter) SetOptions(options options.Options) {
	dw.options = options
}

func (dw *DocWriter) SetSVGGradientStopOpacityMode(mode SVGGradientStopOpacityMode) (prev SVGGradientStopOpacityMode) {
	prev = svgGradientStopOpacityMode(dw.options)
	if dw.options == nil {
		dw.options = options.Options{}
	}
	dw.options[svgGradientStopOpacityModeOption] = svgGradientStopOpacityMode(options.Options{svgGradientStopOpacityModeOption: mode})
	if dw.curPage != nil {
		dw.curPage.SetSVGGradientStopOpacityMode(svgGradientStopOpacityMode(dw.options))
	}
	return prev
}

func (dw *DocWriter) SetSVGBlendMode(mode SVGBlendMode) (prev SVGBlendMode) {
	prev = svgBlendMode(dw.options)
	if dw.options == nil {
		dw.options = options.Options{}
	}
	dw.options[svgBlendModeOption] = svgBlendMode(options.Options{svgBlendModeOption: mode})
	if dw.curPage != nil {
		dw.curPage.SetSVGBlendMode(svgBlendMode(dw.options))
	}
	return prev
}

func (dw *DocWriter) SetUnderline(underline bool) (prev bool) {
	return dw.CurPage().SetUnderline(underline)
}

func (dw *DocWriter) SetVTextAlign(vTextAlign string) (prev string) {
	return dw.CurPage().SetVTextAlign(vTextAlign)
}

func (dw *DocWriter) SetUnits(units string) {
	dw.options["units"] = units
	if dw.curPage != nil {
		dw.curPage.SetUnits(units)
	}
}

func (dw *DocWriter) Strikeout() bool {
	return dw.CurPage().Strikeout()
}

func (dw *DocWriter) VTextAlign() string {
	return dw.CurPage().VTextAlign()
}

func (dw *DocWriter) Underline() bool {
	return dw.CurPage().Underline()
}

func (dw *DocWriter) widthsForFontCodepage(f *font.Font, cpi codepage.CodepageIndex) *indirectObject {
	var widths [256]int
	upm := f.UnitsPerEm()
	// Avoid divide by zero error for unusual fonts.
	if upm > 0 {
		for i, r := range cpi.Map() {
			designWidth, _ := f.AdvanceWidth(r)
			widths[i] = designWidth * 1000 / upm
		}
	}
	pdfWidths := arrayFromInts(widths[32:])
	ioWidths := &indirectObject{dw.nextSeq(), 0, &pdfWidths}
	return ioWidths
}

func (dw *DocWriter) Write(text []byte) (n int, err error) {
	return dw.CurPage().Write(text)
}

// WriteTo implements io.WriterTo.
func (dw *DocWriter) WriteTo(wr io.Writer) (int64, error) {
	defer dw.profiler.Begin("pdf.write").End()
	defer dw.runWriteToCleanups()
	if len(dw.pages) == 0 {
		dw.NewPage()
	}
	{
		span := dw.profiler.Begin("pdf.write.close_pages")
		for _, pw := range dw.pages {
			pw.close()
		}
		span.End()
	}
	dw.curPage = nil
	{
		span := dw.profiler.Begin("pdf.write.resolve_links")
		err := dw.resolveTargetLinks()
		span.End()
		if err != nil {
			return 0, err
		}
	}
	if err := dw.flushUnicodeFonts(); err != nil {
		return 0, err
	}
	{
		span := dw.profiler.Begin("pdf.write.file")
		dw.file.write(wr)
		span.End()
	}
	return 0, nil
}

func (dw *DocWriter) X() float64 {
	return dw.CurPage().X()
}

func (dw *DocWriter) Y() float64 {
	return dw.CurPage().Y()
}
