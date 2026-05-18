# TTF Test Fixtures

This directory contains synthetic and extracted font fixtures used by the `ttf`
package tests.  All fixtures are either generated programmatically or sourced
from fonts with permissive licences.

---

## `minimal.ttf`, `minimal-cff.otf`, and `minimal.ttc`

**Source:** generated
**Licence:** public domain (programmatically constructed)
**Generator:** `ttf/testdata/generate/main.go`

Minimal valid TrueType and OpenType/CFF fonts with a square glyph for every
mapped codepoint. Designed to be small and stable — metrics and embedding
behavior are what matter, not outline design.

### Codepoint coverage

| Range | Description |
|---|---|
| U+0020–U+007E | ASCII printable |
| U+00C0–U+00FF | Latin supplement |
| U+0391–U+03A9 | Greek capitals |
| U+4E2D, U+6587 | Two CJK ideographs (中, 文) |

### Key metrics

| Parameter | Value |
|---|---|
| unitsPerEm | 1000 |
| advanceWidth | 512 (all glyphs) |
| ascent | 800 |
| descent | −200 |
| numGlyphs | 187 (186 mapped + .notdef) |

### Regenerating

```
go run ttf/testdata/generate/main.go
```

Run from the module root.  The generator writes `minimal.ttf`,
`minimal-cff.otf`, and `minimal.ttc` into this directory. Regenerate only when
the fixture specification changes, then re-run `go test ./ttf/...` to verify.

### `minimal-cff.otf`

`minimal-cff.otf` uses an OpenType `OTTO` sfnt wrapper with a CFF table, a CFF
`maxp` version 0.5 table, and the same cmap/metrics coverage as `minimal.ttf`.
It exists to exercise CFF-backed OpenType loading, subsetting, and PDF
embedding without depending on workstation fonts or redistributing third-party
font files.

### `minimal.ttc` layout

Two sub-fonts sharing the same glyph data:

| Index | Family | Style | Weight |
|---|---|---|---|
| 0 | Minimal | Regular | 400 |
| 1 | Minimal | Bold | 700 |

---

## `cjk-sample.ttf`

**Source:** Noto Sans CJK SC, version 2.004 (© 2014–2021 Adobe)
**Licence:** [SIL Open Font License 1.1](https://scripts.sil.org/OFL)
**Original download:** Noto CJK fonts are available at https://github.com/googlefonts/noto-cjk

Subset extracted from the Noto Sans CJK SC Regular font. Contains:

- Latin fallback glyphs U+0020–U+007E (95 codepoints)
- CJK Unified Ideographs U+4E00–U+4E63 (100 codepoints)
- Total: 263 glyphs (including .notdef), 30 KB

Subsetting command used (fonttools `pyftsubset`):

```
pyftsubset wqy-microhei.ttc \
  --font-number=0 \
  --unicodes="U+0020-007E,U+4E00-4E63" \
  --output-file=ttf/testdata/cjk-sample.ttf \
  --no-hinting
```

The source TTC was downloaded as `wqy-microhei.ttc` but identified as Noto Sans
CJK SC Regular upon inspection of its name table (a common repackaging of the
Noto CJK fonts circulates under the WenQuanYi name).
