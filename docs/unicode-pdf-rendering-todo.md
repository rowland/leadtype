# Unicode PDF Rendering — Task Checklist

See [unicode-pdf-rendering.md](unicode-pdf-rendering.md) for full design notes.
See [development-process.md](development-process.md) for testing strategy and fixture details.

---

## Test Infrastructure and Fixtures

### Minimal synthetic TTF fixture

- [x] Create `ttf/testdata/generate/main.go` (build-tagged `//go:build ignore`)
- [x] Generator: write valid `head` table (magic number, unitsPerEm=1000, bbox)
- [x] Generator: write valid `hhea` table (ascent, descent, lineGap, numOfLongHorMetrics)
- [x] Generator: write valid `maxp` table (version, numGlyphs)
- [x] Generator: write valid `name` table (family="Minimal", subfamily="Regular", PostScript name)
- [x] Generator: write valid `OS/2` table (weightClass, fsType, ulCharRange, version)
- [x] Generator: write valid `post` table (format 3.0 — no glyph names, simplest form)
- [x] Generator: write cmap format 4 covering U+0020–U+007E (ASCII printable)
- [x] Generator: extend cmap to include U+00C0–U+00FF (Latin supplement)
- [x] Generator: extend cmap to include U+0391–U+03A9 (Greek capitals)
- [x] Generator: extend cmap to include two CJK codepoints (U+4E2D, U+6587)
- [x] Generator: write `glyf` table — identical square contour for each glyph
- [x] Generator: write `loca` table matching `glyf` offsets
- [x] Generator: write `hmtx` table — advanceWidth=512 for all glyphs
- [x] Generator: assemble tables in spec-required order and compute `head.checkSumAdjustment`
- [x] Run generator and commit `ttf/testdata/minimal.ttf`
- [x] Self-verification test: parse `minimal.ttf` with `LoadFont`, assert expected metric values
- [x] Add `ttf/testdata/README.md` documenting the fixture and how to regenerate

### Minimal TTC fixture

- [x] Extend generator to produce `ttf/testdata/minimal.ttc`
- [x] TTC contains two sub-fonts: "Minimal Regular" and "Minimal Bold"
- [x] "Minimal Bold" differs only in `name` table and OS/2 `usWeightClass`
- [x] Commit `ttf/testdata/minimal.ttc`
- [x] Test: `LoadFontInfosFromTTC("testdata/minimal.ttc")` returns exactly 2 entries with correct family/style

### CJK sample fixture

- [x] Choose a source font under SIL OFL or Apache 2 licence (e.g. Noto Sans CJK subset)
- [x] Use `pyftsubset` (fonttools) to extract ~50–100 CJK codepoints + U+0020–U+007E
- [x] Commit the subset to `ttf/testdata/cjk-sample.ttf`
- [x] Document source font, version, licence, and subsetting command in `ttf/testdata/README.md`
- [x] Test: parse `cjk-sample.ttf`, verify cmap returns valid glyph IDs for selected CJK codepoints

### Golden file infrastructure

- [x] Add `-update` flag to `pdf` package tests (`var update = flag.Bool("update", false, "regenerate golden files")`)
- [x] Add `compareGolden(t, got []byte, path string)` helper to `pdf/super_test.go`
- [x] Add `normaliseDate(data []byte) []byte` helper that replaces the `CreationDate` value with a fixed string
- [x] Create `pdf/testdata/golden/` directory (add a `.gitkeep` until first golden file is committed)
- [ ] Document golden file update procedure in `development-process.md` (already described — link from test helper)

### Shared test utilities

- [ ] Evaluate whether `expect*` helpers in `ttf/support_test.go` and `SuperTest` in `pdf/super_test.go` should be consolidated into `internal/testutil`
- [ ] If consolidated: migrate existing callers in `ttf/`, `pdf/`, `rich_text/`
- [x] Add `testFontSource(t, path string) font.FontSource` helper that loads a fixture TTF and skips on error

---

## Phase 1 — ToUnicode CMaps for existing simple fonts

- [x] Add `Map() []rune` accessor to `codepage.Codepage` (byte index → rune, for CMap generation)
- [x] Add `toUnicodeCMap(encoding []rune) []byte` helper in `pdf/` that writes a CMap stream
- [x] Add `ToUnicode` stream field to the simple `TrueTypeFont` PDF object
- [x] Add `ToUnicode` stream field to the `Type1Font` PDF object
- [x] Wire `toUnicodeCMap` into `DocWriter.fontKey()` for both font types
- [x] Write tests: generated PDF contains `/ToUnicode` entry for each embedded font
- [x] Write tests: ToUnicode CMap correctly round-trips all 256 positions of CP1252
- [ ] Write tests: ToUnicode CMap correctly round-trips ISO-8859-2 differences
- [ ] Manual verification: copy-paste text from a generated PDF works in a viewer

---

## Phase 2 — Composite font (Type 0) path for TTF

### PDF object model

- [x] Add `CIDSystemInfo` struct (Registry, Ordering, Supplement)
- [x] Add `CIDFont` PDF object (`/Subtype /CIDFontType2`, descendant of Type0)
- [x] Add `Type0Font` PDF object (`/Subtype /Type0`, `/Encoding /Identity-H`)
- [x] Add `/W` (glyph width array) builder — sparse format per PDF spec §5.6.3
- [x] Add `FontDescriptor` wiring for CIDFont (reuse existing descriptor fields where possible)
- [x] Add `CIDToGIDMap /Identity` entry to CIDFont object

### Glyph ID tracking

- [x] Add `GlyphRecorder` type: per-font map of `glyphID → rune` accumulated during rendering
- [x] Integrate `GlyphRecorder` into `DocWriter` (one recorder per font face)
- [x] Replace `CharForCodepoint` call in page writer with `cmapTable.glyphIndex(rune)` for TTF fonts
- [x] Write text as big-endian uint16 glyph ID pairs (update `tw.show()` for Unicode mode)

### ToUnicode CMap for composite fonts

- [x] Extend `toUnicodeCMap` to accept a `map[uint16]rune` (glyph ID → Unicode) input
- [x] Emit CMap using `beginbfchar` / `endbfchar` blocks (max 100 entries per block per spec)
- [x] Attach ToUnicode stream to `Type0Font` object at document close

### Width array

- [x] Build `/W` array from `hmtxTable` for all recorded glyph IDs at document close
- [x] Use the font's `unitsPerEm` to scale widths to 1000-unit PDF space
- [ ] Emit `/DW` (default width) as most-common glyph width to minimise `/W` entries

### DocWriter integration

- [x] Add `UnicodeMode bool` option (or constructor variant) to `DocWriter`
- [x] When `UnicodeMode` is set, select `Type0Font` path instead of `TrueTypeFont` path
- [x] `fontKey()` returns a key per font face (not per font+codepage) in Unicode mode
- [x] Accumulate all glyph recorders and flush CMap + width array at `DocWriter.Close()`

### Testing

- [x] Unit test: `Type0Font` object serialises correctly to PDF syntax
- [x] Unit test: `/W` sparse width array is correctly formed
- [x] Unit test: ToUnicode CMap stream is valid and parseable
- [x] Integration test: render ASCII text in Unicode mode, verify output matches simple-font path
- [ ] Integration test: render Greek + Cyrillic mixed string in a single font call
- [ ] Integration test: render a sample of CJK characters (requires a CJK font on disk)
- [ ] Integration test: render emoji (requires a color/emoji font — may be stretch goal)
- [x] Verify existing simple-font tests still pass unchanged

---

## Phase 3 — Remove EachCodepage from TTF rendering path

- [ ] Audit all call sites of `EachCodepage` — identify which are TTF vs AFM/Type1
- [ ] Add `RenderUnicode(text string, font *Font)` method to `PageWriter` (no codepage segmentation)
- [ ] Refactor `PageWriter` text rendering loop to use glyph ID accumulation for TTF fonts
- [ ] Retain `EachCodepage` loop for AFM/Type1 font path only
- [ ] Remove codepage-per-segment font-key switching from the TTF path
- [ ] Update `rich_text` / `ltml` rendering to call `RenderUnicode` when font is TTF
- [ ] Test: multi-script string (Latin + Greek + Cyrillic) renders as a single PDF text segment
- [ ] Test: AFM/Type1 path still segments correctly via `EachCodepage`
- [ ] Test: no regression in existing rich_text or ltml tests

---

## Phase 4 — Font subsetting

### TTF table parsing (prerequisites)

- [x] Parse `loca` table (glyph offset index) in `ttf/`
- [x] Parse `glyf` table headers (enough to find composite glyph component references)
- [x] Identify composite glyph component IDs recursively (closure expansion)

### Subset builder

- [x] Add `ttf.Subset(glyphIDs []uint16) ([]byte, error)` function
- [x] Rewrite `glyf` table: include only subset glyphs, pad/zero unused slots or use short loca
- [x] Rewrite `loca` table to match subset `glyf`
- [ ] Update `maxp.numGlyphs` to subset count
- [ ] Subset `hmtx` table to subset glyph IDs
- [ ] Subset `cmap` table: remove mappings for glyphs not in subset
- [ ] Subset `post` table glyph names (if format 2.0)
- [x] Preserve all non-glyph tables (`head`, `hhea`, `name`, `OS/2`, `kern`, etc.) unchanged
- [x] Recalculate `head.checkSumAdjustment` for the new file

### PDF integration

- [x] Generate a 6-character random uppercase tag for each subset font
- [x] Prefix PostScript name with tag (`ABCDEF+FontName`) per PDF spec §5.5.3
- [x] Replace full font stream with subset stream at `DocWriter.Close()`
- [x] Write tests: subset font file is a valid TTF (parse it back with `ttf.LoadFont`)
- [x] Write tests: subset contains exactly the expected glyph IDs
- [x] Write tests: composite glyph components are included in the closure
- [ ] Benchmark: measure embedded font stream size before and after subsetting for a Latin document
- [ ] Benchmark: measure embedded font stream size before and after subsetting for a CJK document

---

## Phase 5 — OpenType / CFF support (stretch goal)

- [ ] Parse `CFF ` table top-level structure (Header, Name Index, Top DICT Index)
- [ ] Parse CharStrings INDEX to enumerate glyph count and offsets
- [ ] Parse Charset to map glyph IDs to SID/glyph names
- [ ] Implement CFF subset: retain only used glyph charstrings + subroutines
- [ ] Emit `CIDFontType0` (instead of `CIDFontType2`) for CFF-based fonts
- [ ] Test with a common OTF font (e.g. a Google Fonts CFF font)

---

## Cross-cutting concerns

- [ ] Decide on public API surface for Unicode mode (option struct vs separate constructor)
- [x] Update `font.FontMetrics` interface if new methods are needed (e.g. `GlyphIndex(rune)`)
- [ ] Ensure `haru/` backend either gains equivalent composite font support or documents the gap
- [x] Add a sample program in `samples/` demonstrating multi-script Unicode output
- [ ] Update package-level doc comments to describe Unicode vs codepage paths
- [ ] Profile Phase 2 rendering path vs Phase 0 baseline for a Latin-only document
