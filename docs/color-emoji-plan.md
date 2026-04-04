# Color Emoji Support — Plan Of Attack

## Purpose

This document outlines a practical path to supporting Apple Color Emoji and
similar color emoji fonts in Leadtype's PDF pipeline.

The immediate motivation is that Leadtype can now:

- map emoji codepoints through the Unicode Type 0 / CIDFont path
- emit correct `ToUnicode` mappings for non-BMP codepoints
- embed and subset ordinary outline TrueType fonts

But it cannot yet render Apple Color Emoji correctly in generated PDFs. On the
current macOS system font, `Apple Color Emoji` includes `glyf`/`loca` tables
and also an `sbix` table. In practice, Leadtype currently emits tofu boxes for
that font because the rendering and embedding path does not understand or
preserve the color glyph data.

---

## Current State

### What already works

- Unicode text is emitted through Type 0 composite fonts.
- `ToUnicode` CMaps correctly encode surrogate pairs for non-BMP codepoints.
- Glyph IDs are tracked and subset into embedded TTF streams.
- Outline-capable monochrome symbol fonts can render emoji-style BMP symbols.

### What does not work yet

- Color emoji fonts such as Apple Color Emoji do not render as intended.
- The current TTF subset pipeline only meaningfully handles the classic outline
  path centered on `glyf`, `loca`, `hmtx`, `cmap`, and related tables.
- No code currently parses or preserves color-glyph tables such as:
  - `sbix`
  - `CBDT` / `CBLC`
  - `COLR` / `CPAL`
  - `SVG `
- No viewer-compatibility matrix exists yet for embedded color emoji output.
- No fallback strategy exists when the embedded font format is unsupported by
  the target PDF viewer.

---

## Key Unknowns

These are the main questions to answer before committing to a full
implementation:

1. Can common PDF viewers render embedded `sbix` glyphs in a subset
   `CIDFontType2` TrueType font?
2. If yes, what minimum set of tables must be preserved so the viewer finds the
   bitmap strikes correctly?
3. If no, is a font-embedding approach still viable for color emoji, or do we
   need a fallback that emits emoji as images instead of text glyphs?
4. How should multi-codepoint emoji sequences be represented in `ToUnicode`,
   shaping, and glyph recording?
5. What public API should control color-emoji behavior when viewer support is
   uncertain?

---

## Recommended Strategy

Treat this as a staged workstream with an early spike, then a narrow first
feature, then sequence support and broader format support.

Recommended order:

1. Research and proof-of-concept for Apple `sbix`
2. Single-codepoint `sbix` emoji rendering
3. Sequence and variation-selector support
4. Fallback image path for unsupported viewers or unsupported font formats
5. Optional support for other color font formats (`COLR`, `CBDT`, `SVG `)

---

## Phase 0 — Research Spike

### Goal

Determine whether embedded `sbix` emoji is viable in the existing PDF font
embedding model.

### Tasks

- Add a small inspection tool or temporary debug helper to enumerate color font
  tables from a TTF/TTC font.
- Confirm the exact tables present in Apple Color Emoji on a representative
  macOS version.
- Create a reduced experimental PDF using only one or two emoji glyphs from
  Apple Color Emoji.
- Compare behavior in several viewers:
  - macOS Preview
  - Adobe Acrobat Reader if available
  - browser PDF viewers if available
  - Poppler rendering tools (`pdftoppm`) as an additional sanity check
- Record whether:
  - tofu boxes appear
  - monochrome fallback appears
  - color bitmap glyphs appear
  - text extraction still works

### Deliverable

A short engineering note answering:

- whether embedded `sbix` is viewer-viable
- which tables appear necessary
- whether a fallback image path is required

### Exit Criteria

Do not start broad implementation until this question is answered:

"Can a minimally embedded Apple Color Emoji subset render in at least one major
viewer in a way worth supporting?"

---

## Phase 1 — Parse Color Glyph Metadata

### Goal

Teach `ttf/` enough about color font tables to inspect and subset them safely.

### Scope

Start with `sbix` because that is the format used by Apple Color Emoji on the
observed system font.

### Proposed parser work

- Add table presence tests for:
  - `sbix`
  - `CBDT`
  - `CBLC`
  - `COLR`
  - `CPAL`
  - `SVG `
- Add an `sbix` parser with enough structure to:
  - read strike headers
  - map glyph IDs to bitmap records
  - identify image format per glyph
  - identify glyphs that reuse another glyph's bitmap data
- Extend `FontInfo` or `Font` with lightweight capability helpers such as:
  - `HasColorGlyphs()`
  - `ColorGlyphFormat() string`
  - `HasSbix()`

### Suggested files

- new `ttf/sbix_table.go`
- optional `ttf/sbix_table_test.go`
- light-touch additions in `ttf/font_info.go` or `ttf/font.go`

### Tests

- Unit test: Apple Color Emoji `sbix` table loads without error
- Unit test: expected strikes are visible
- Unit test: a few known glyph IDs have `sbix` bitmap entries

### Deliverable

Leadtype can inspect a color emoji font and report that it is an `sbix` color
font rather than a plain outline-only font.

---

## Phase 2 — Preserve Color Tables In Subsets

### Goal

Make font subsetting preserve enough color-font data for embedded rendering to
remain possible.

### Current blocker

`ttf.Subset` is designed around outline fonts. It keeps a limited set of tables
and rewrites only the classic TrueType path. Even though Apple Color Emoji also
contains `glyf`/`loca`, simply preserving those tables is not enough.

### Proposed work for `sbix`

- Extend `subsetKeepTable` so the subsetter can retain `sbix` when present.
- Rewrite `sbix` so the subset contains only bitmap data needed by retained
  glyph IDs.
- Preserve or rewrite glyph references where one `sbix` record refers to
  another glyph's bitmap data.
- Ensure glyph IDs remain stable enough for `sbix` lookup to work after
  subsetting.
- Verify checksum and table directory assembly still produce a valid font file.

### Important design choice

Leadtype's current subsetter preserves original glyph IDs up to the highest used
glyph ID rather than renumbering densely. That is helpful here because `sbix`
records are indexed by glyph ID. Preserve this behavior unless there is a strong
reason to change it.

### Tests

- Unit test: subset of an `sbix` font re-parses successfully
- Unit test: subset retains `sbix`
- Unit test: subset glyph chosen for an emoji still has an `sbix` entry
- Size sanity test: subset removes unrelated bitmap glyph data

### Deliverable

A subsetted Apple Color Emoji font that still contains the bitmap data required
for selected glyphs.

---

## Phase 3 — PDF Embedding Viability

### Goal

Determine whether preserving `sbix` is sufficient for PDF viewers to render the
embedded color glyphs.

### Tasks

- Generate a tiny PDF using one emoji glyph in a font subset that preserves
  `sbix`.
- Verify rendering in the viewer matrix from Phase 0.
- Verify that text extraction still returns the original Unicode rune.
- Verify that file size remains reasonable compared with embedding the full
  font.

### Possible outcomes

#### Outcome A: Embedded `sbix` works

Proceed with a normal font-based implementation.

#### Outcome B: Embedded `sbix` does not work reliably

Use a fallback rendering path for color emoji:

- detect color emoji glyphs during text emission
- rasterize or extract glyph images from the font
- place those as PDF image content at glyph positions
- optionally keep an invisible text layer for extraction/search

This fallback may be the more robust cross-viewer path even if one viewer can
render embedded `sbix`.

### Deliverable

A go/no-go decision on the font-embedding approach.

---

## Phase 4A — Native Color Font Path

This phase applies if embedded `sbix` proves viable.

### Goal

Render single-codepoint Apple Color Emoji glyphs through the normal text path.

### Tasks

- Allow color-font subsets through the existing `DocWriter` close/flush logic.
- Ensure `FontDescriptor`, `Type0`, and `CIDFontType2` objects remain valid for
  the subset.
- Confirm that the current glyph recording path still maps glyph IDs correctly.
- Add integration tests using a guarded system font probe, similar to the
  current emoji integration test, but requiring `sbix`.

### Tests

- Integration test: render one Apple Color Emoji glyph and visually verify via
  PDF rasterization
- Integration test: rendered PDF still contains the correct `ToUnicode` mapping
- Integration test: subset tag and embedded stream remain valid

### Deliverable

Single-codepoint Apple Color Emoji renders correctly in at least one supported
viewer via embedded font data.

---

## Phase 4B — Fallback Image Path

This phase applies if embedded `sbix` is not viewer-reliable, or if support is
too inconsistent to expose as the default behavior.

### Goal

Render emoji visually by using bitmap/image placement while preserving text
semantics as much as practical.

### Proposed approach

- During text shaping/emission, detect glyphs that belong to a color font and
  require image fallback.
- Extract the glyph bitmap for a target strike size from `sbix`.
- Convert that bitmap into a PDF image XObject.
- Place the image at the glyph position using the same pen advance logic as the
  text path.
- Preserve a text layer for extraction/search/accessibility if possible.

### Design questions

- Should the text layer be invisible text underneath the image?
- Should emoji fallback be opt-in or automatic?
- How should this interact with char spacing, word spacing, transforms,
  clipping, and curved text?

### Tests

- Integration test: emoji glyph appears visually in rasterized PDF output
- Integration test: adjacent non-emoji text still uses the regular text path
- Integration test: mixed strings maintain spacing/alignment
- Integration test: extracted text still includes the emoji rune or sequence

### Deliverable

A viewer-robust rendering path for Apple Color Emoji even if embedded color
font support is weak.

---

## Phase 5 — Emoji Sequences And Shaping

### Goal

Handle real-world emoji text, not just isolated codepoints.

### Examples

- Variation selectors: `☺` vs `☺️`
- ZWJ sequences: family groupings, professions, kiss, etc.
- Skin tone modifiers
- Regional indicator flags
- Keycap sequences

### Requirements

- Determine whether the font uses GSUB/GPOS or cmap-level mapping for these
  sequences.
- Extend glyph recording to map a single glyph or glyph cluster to multiple
  runes in `ToUnicode`.
- Ensure the shaper path does not drop or duplicate emoji sequence text.
- Decide whether the existing shaping integration is enough or whether emoji
  fonts need their own shaping/fallback handling.

### Tests

- Unit test: cluster-to-rune mapping for emoji sequences
- Integration test: one visible emoji from a ZWJ sequence
- Integration test: `ToUnicode` maps back to the full rune sequence

### Deliverable

Emoji sequences render and extract correctly, not only single codepoints.

---

## Phase 6 — Broader Color Font Formats

### Goal

Extend support beyond Apple `sbix`.

### Candidate formats

- `COLR` / `CPAL`
  Useful for modern layered vector color fonts and likely a better long-term
  target than bitmap-only approaches.
- `CBDT` / `CBLC`
  Bitmap color glyph tables used by some emoji fonts.
- `SVG `
  Embedded SVG glyphs.

### Recommendation

Do not start here. Finish the Apple `sbix` decision path first. It is better to
prove one format end to end than to partially parse several formats at once.

---

## Public API Considerations

This feature will likely need explicit policy controls.

### Suggested options

- `DocWriter.ColorEmojiMode(...)`
- or an option struct on the Unicode pipeline

### Candidate modes

- `DisableColorEmoji`
  Treat emoji fonts as unsupported for visual color rendering.
- `PreferEmbeddedColorFont`
  Use preserved color font data if the format/viewer path is enabled.
- `PreferImageFallback`
  Render color emoji through image placement.
- `Auto`
  Pick the best available strategy for the detected font format.

### Reasoning

Color emoji support is much more viewer-sensitive than standard outline fonts.
An explicit mode will make behavior easier to test and easier to explain.

---

## Test Strategy

### Unit tests

- Table parsing for `sbix`
- Table-preservation and rewrite tests for subsets
- Capability detection tests for color font formats

### Integration tests

- System-font tests guarded by clean skips when Apple Color Emoji is unavailable
- PDF structure tests verifying embedded tables and `ToUnicode`
- Visual regression tests based on rasterized PDF output for a very small sample

### Manual verification

For each major milestone, manually verify in:

- Preview
- Acrobat Reader if available
- Poppler rasterization

### Fixtures

If licensing permits, consider a tiny committed `sbix` fixture or a small
subsampled color font fixture. If that is impractical, use guarded
system-font-based tests plus clear documentation.

---

## Risks

### Viewer support risk

PDF viewers may not support embedded `sbix` consistently, even if the PDF and
font are technically well-formed.

### Complexity risk

Color glyph support can easily sprawl into several font formats, shaping rules,
and fallback behaviors. Tight scoping is important.

### File size risk

Bitmap emoji data can inflate embedded font size or image streams quickly.

### Maintenance risk

A custom image fallback path may become a second text-rendering system unless
carefully isolated.

---

## Recommended First Milestone

The best next milestone is:

"Prove whether a subsetted Apple Color Emoji font with preserved `sbix` can
render a single-codepoint emoji in at least one major PDF viewer."

That milestone is small, answers the biggest uncertainty, and cleanly decides
between the native-font path and the image-fallback path.

---

## Concrete Next Tasks

1. Add `HasColorGlyphs`-style helpers and table detection in `ttf/`.
2. Implement minimal `sbix` parsing.
3. Add a temporary spike path to preserve `sbix` unchanged in subsets.
4. Generate a one-emoji experimental PDF from Apple Color Emoji.
5. Record viewer results in a short follow-up note.
6. Choose either:
   native embedded-color support, or image fallback as the primary path.

