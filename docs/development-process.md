# Development Process

This document covers the development workflow, testing strategy, and test fixture management for the Unicode PDF rendering project. It should be read alongside [unicode-pdf-rendering.md](unicode-pdf-rendering.md) (design) and [unicode-pdf-rendering-todo.md](unicode-pdf-rendering-todo.md) (task checklist).

---

## Workflow

### Branching

Work in feature branches off `master`. Branch names should reflect the phase or component:

```
unicode/phase1-tounicode
unicode/phase2-type0-font
unicode/phase2-glyph-recorder
unicode/phase4-subsetting
```

Each branch should be focused enough that its PR is reviewable in a single sitting. Phases 2 and 4 are large enough to warrant multiple PRs.

### Commits

- One logical change per commit. Prefer small, buildable commits over large all-at-once ones.
- `go build ./...` and `go test ./...` must pass at every commit.
- Commit messages: imperative mood, present tense. Lead with the package or subsystem (`ttf: `, `pdf: `, `codepage: `).

### Pull Requests

- Reference the relevant phase from the design doc in the PR description.
- Include a brief test plan: what was manually verified beyond the automated tests.
- The CI workflow (`.github/workflows/ci.yml`) must be green before merging.

### CI

The existing CI workflow runs `go build ./...` and `go test ./...`. For the new Unicode work, tests that depend on system fonts must either skip gracefully when fonts are absent or use in-repo test fixtures (see below). CI should pass on a clean Linux runner with no fonts installed.

---

## Testing Philosophy

The codebase has two distinct kinds of tests today:

1. **Pure serialisation tests** (`pdf/objects_test.go`, `codepage/codepage_test.go`) — no external dependencies, fast, always pass. These are the gold standard and should be the primary form for new code.
2. **System-font integration tests** (`ttf/font_test.go`, `ttf_fonts/ttf_fonts_test.go`) — depend on specific files at `/Library/Fonts/`. These skip when the fonts are absent, which means they never run in CI. They also encode exact glyph counts and metric values for a specific font version.

The new Unicode work should avoid the second pattern. All font-dependent tests should use **in-repo test fixtures** — small, version-controlled font files with known, stable contents.

### Test categories

| Category | Location convention | Dependencies | Run in CI |
|---|---|---|---|
| Unit — pure logic | `*_test.go` alongside source | None | Yes |
| Unit — fixture font | `testdata/` subfolder per package | In-repo font files | Yes |
| Integration — system font | `*_test.go` with `t.Skip` guard | System fonts | No (skip) |
| Golden file | `testdata/golden/` | In-repo reference output | Yes (compare) |
| Benchmark | `Benchmark*` functions | May use system fonts | No (manual) |

---

## Test Fixtures

### Directory layout

```
ttf/testdata/
    minimal.ttf          # Smallest valid TTF with a handful of glyphs
    minimal.ttc          # Two-font TTC wrapping two copies of minimal.ttf
    cjk-sample.ttf       # Subset of a CJK font (a few dozen glyphs)
    cmap-format4.ttf     # Font whose cmap uses format 4 (segment mapping)
    cmap-format12.ttf    # Font whose cmap uses format 12 (full Unicode)

pdf/testdata/
    golden/
        simple-latin.pdf         # Expected output: Latin text, simple font
        unicode-latin.pdf        # Expected output: Latin text, Type0 font
        unicode-multiscript.pdf  # Expected output: Greek + Cyrillic, Type0
        unicode-cjk.pdf          # Expected output: CJK glyphs, Type0
        subset-latin.pdf         # Expected output: Latin text, subsetted font
```

### Minimal synthetic TTF fixture

Many tests need only a handful of glyphs and stable metric values. Rather than shipping a real font (which carries a licence and may change), we generate a minimal TTF programmatically and commit the result.

The minimal fixture should contain:

- **cmap** (format 4): mappings for ASCII printable range U+0020–U+007E, plus a small selection of non-ASCII glyphs (e.g. U+00C0–U+00FF Latin supplement, U+0391–U+03A9 Greek capitals, U+4E2D U+6587 two CJK characters).
- **glyf**: simple square glyphs (one contour, four points) for every mapped codepoint. All glyphs identical — metrics are what matter, not outlines.
- **hmtx**: fixed advance width of 512 for all glyphs; units-per-em = 1000.
- **head**, **hhea**, **maxp**, **name**, **post**, **OS/2**: minimal valid values.

A generator program at `ttf/testdata/generate/main.go` (build-tagged `//go:build ignore`) produces `ttf/testdata/minimal.ttf`. Run it once and commit the output. Regenerate only when the fixture specification changes.

```
go run ttf/testdata/generate/main.go
```

The generator itself is tested by parsing its output with `ttf.LoadFont` and asserting expected values — this is a self-verifying fixture.

### Minimal TTC fixture

The TTC fixture wraps two copies of `minimal.ttf` at different offsets, with different `name` table values (`Minimal Regular` and `Minimal Bold`). The generator produces this alongside `minimal.ttf`.

### CJK sample fixture

For Phase 2 and Phase 4 tests we need a font with CJK cmap entries. Use a small CFF or TrueType subset extracted from a freely-licensed font (e.g. a subset of Noto Sans CJK or Source Han Sans under the SIL OFL). The subset should contain:

- Roughly 50–100 CJK Unified Ideograph codepoints.
- Latin fallback glyphs U+0020–U+007E so mixed-script tests work in a single font.

Use a subsetting tool (e.g. `pyftsubset` from fonttools) to extract the subset, then commit the result. Document the source font and the subsetting command in a `ttf/testdata/README.md`.

### Golden files

Golden files are reference PDF outputs committed to the repository. Tests render a document and compare the output byte-for-byte (or structurally, for fields like timestamps) against the golden file.

**Generating a golden file:**

```go
// In test:
const update = false // set to true to regenerate

if update {
    os.WriteFile("testdata/golden/simple-latin.pdf", got, 0644)
}
golden, _ := os.ReadFile("testdata/golden/simple-latin.pdf")
if !bytes.Equal(got, golden) {
    t.Errorf("output differs from golden; run with update=true to regenerate")
}
```

Or use a flag:

```
go test ./pdf/... -update
```

**What to include in golden PDFs:**

- Keep them small: one or two lines of text, one font, one page.
- Strip or normalise the `CreationDate` entry before comparison (it changes on every run). Either set it to a fixed value in test mode or compare everything except the timestamp line.
- Commit golden files uncompressed where possible so diffs are readable.

**When to regenerate:**

Golden files must be regenerated (and the diff reviewed) whenever:
- PDF object serialisation changes.
- Font embedding logic changes.
- A new PDF feature intentionally alters output structure.

Never regenerate silently. Always review the diff before committing updated golden files.

---

## Writing Tests for New Code

### Phase 1 — ToUnicode CMap

CMap generation is pure string/byte logic with no font file dependency. Test it as a unit:

```go
func TestToUnicodeCMap_CP1252(t *testing.T) {
    cmap := toUnicodeCMap(codepage.Idx_CP1252.Map())
    // Parse the CMap stream and verify round-trip for a sample of entries
    assertCMapEntry(t, cmap, 0x80, 0x20AC) // byte 128 → EURO SIGN
    assertCMapEntry(t, cmap, 0x41, 0x0041) // byte 65  → LATIN CAPITAL A
}
```

Test the PDF object that wraps the stream:

```go
func TestSimpleFont_ToUnicodeEntry(t *testing.T) {
    f := newTrueTypeFont(...)
    s := stringFromWriter(f)
    if !strings.Contains(s, "/ToUnicode") {
        t.Error("font object missing /ToUnicode entry")
    }
}
```

### Phase 2 — Type 0 font objects

Test PDF object serialisation in isolation (no font file needed):

```go
func TestType0Font_Serialisation(t *testing.T) {
    f := newType0Font(42, 0, "ArialMT", widths, descriptor)
    got := stringFromWriter(f)
    want := `42 0 obj
<<
/BaseFont /ArialMT
/DescendantFonts [43 0 R ]
/Encoding /Identity-H
/Subtype /Type0
/ToUnicode 44 0 R
/Type /Font
>>
endobj
`
    if got != want {
        t.Errorf("unexpected serialisation:\ngot:  %s\nwant: %s", got, want)
    }
}
```

Test the glyph width array builder:

```go
func TestCIDWidthArray(t *testing.T) {
    // widths: glyph 1 → 500, glyph 2 → 500, glyph 5 → 750
    w := buildCIDWidthArray(map[uint16]int{1: 500, 2: 500, 5: 750})
    // Expect sparse format: [1 [500 500] 5 [750]]
    expectS(t, "widths", "[1 [500 500 ] 5 [750 ] ] ", stringFromWriter(w))
}
```

Test glyph rendering using the minimal fixture font:

```go
func TestGlyphRecorder(t *testing.T) {
    f, err := ttf.LoadFont("testdata/minimal.ttf")
    require.NoError(t, err)
    rec := newGlyphRecorder()
    rec.record(f.GlyphIndex('A'), 'A')
    rec.record(f.GlyphIndex('B'), 'B')
    ids, runes := rec.glyphs()
    assert.Equal(t, []uint16{glyphID_A, glyphID_B}, ids)
    assert.Equal(t, []rune{'A', 'B'}, runes)
}
```

### Phase 4 — Font subsetting

Test the subset builder against the minimal fixture:

```go
func TestSubset_ReducesGlyphs(t *testing.T) {
    data, _ := os.ReadFile("testdata/minimal.ttf")
    keep := []uint16{1, 5, 12} // glyph IDs to keep
    subset, err := ttf.Subset(data, keep)
    require.NoError(t, err)

    // Parse the subset back and verify
    f, err := ttf.LoadFontFromBytes(subset)
    require.NoError(t, err)
    assert.Equal(t, len(keep)+1, f.NumGlyphs()) // +1 for .notdef
}

func TestSubset_CompositeGlyphClosure(t *testing.T) {
    // Verify that composite glyph components are included even if not in the keep set
    ...
}
```

### Integration tests

Integration tests render a complete document and compare against a golden file:

```go
func TestUnicodeRender_Latin(t *testing.T) {
    var buf bytes.Buffer
    dw := NewDocWriter(WithUnicodeMode())
    dw.AddFontSource(testFontSource(t)) // loads testdata/minimal.ttf
    pw := dw.NewPage()
    pw.SetFont("Minimal", "", 12, nil)
    pw.MoveTo(72, 720)
    pw.Print("Hello, World!")
    dw.WriteTo(&buf)

    compareGolden(t, buf.Bytes(), "testdata/golden/unicode-latin.pdf")
}
```

`compareGolden` normalises the `CreationDate` field before comparing:

```go
func compareGolden(t *testing.T, got []byte, path string) {
    t.Helper()
    got = normaliseDate(got)
    if *update {
        os.WriteFile(path, got, 0644)
        return
    }
    want, err := os.ReadFile(path)
    require.NoError(t, err)
    if !bytes.Equal(got, normaliseDate(want)) {
        t.Errorf("output differs from %s\nrun `go test -update` to regenerate", path)
    }
}
```

---

## Running Tests

```bash
# All tests (will skip system-font tests if fonts are absent)
go test ./...

# Only fixture-based tests (no system fonts required)
go test ./ttf/... ./pdf/... ./codepage/...

# Regenerate golden files
go test ./pdf/... -update

# Regenerate minimal TTF fixture
go run ttf/testdata/generate/main.go

# Run benchmarks
go test ./ttf/... -bench=. -benchmem -run=^$

# Race detector (run before each PR)
go test -race ./...
```

---

## Test Coverage Goals

| Package | Minimum target | Notes |
|---|---|---|
| `codepage` | 90% | Pure table logic, easy to cover fully |
| `ttf` | 75% | Parser paths depend on fixture font variety |
| `pdf` | 80% | Object serialisation is cheap to test exhaustively |
| `ttf_fonts` | 60% | System-font paths inherently hard to cover in CI |
| `rich_text` | 75% | Segment logic is pure; rendering integration harder |

Check coverage with:

```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

---

## Dependency Management

The project currently has no external dependencies beyond the standard library. Keep it that way:

- Do not add a testing framework dependency (e.g. testify) without discussion. The existing `SuperTest` helper in `pdf/super_test.go` and the `expect*` helpers in `ttf/support_test.go` cover most needs.
- If the same assertion helper pattern is duplicated across three or more packages, factor it into a shared internal test-helper package (`internal/testutil`).
- Font files committed to `testdata/` must be under a permissive licence (SIL OFL, Apache 2, public domain). Document the source and licence in `testdata/README.md` in each package that has fixture fonts.
