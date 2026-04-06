# Arabic ToUnicode Bug Investigation

## Problem Statement

Arabic text renders visually correctly in generated PDFs, but text extraction
via `pdftotext` or copy-paste produces broken ligatures and other anomalies.
The goal is to identify and fix the root cause in the ToUnicode CMap generation
pipeline.

Known symptoms from prior investigation (`arabic-extraction-investigation.md`):

- `الخبرات` extracts as `ال خبرات`
- `الجيران` extracts as `ال جيران`
- `الحِرَف` extracts as `ال حِ رَف`
- `اليدوية` extracts as `اليدو ية`

## Pipeline Summary

1. Input runes (logical Unicode order) → shaper → `[]GlyphPosition` (visual order)
2. `shapedGlyphRuneAssignments()` maps each glyph index back to source runes
3. `glyphRecorder.recordRunes()` assigns CIDs and tracks CID→Unicode mapping
4. Content stream emits CIDs; CIDToGIDMap maps CID→glyphID; ToUnicode maps CID→Unicode
5. At doc close, `toUnicodeCMapDataComposite()` builds the ToUnicode CMap from the mapping

---

## Hypothesis 1: `reverseRunes` corrupts ToUnicode sequences

### Rationale

In `shapedGlyphRuneAssignments()` (page_writer.go:1327), the rune subsequence
for each cluster is reversed via `reverseRunes()` before being assigned to
glyphs. The ToUnicode CMap should contain runes in **logical** (Unicode)
order, not visual order. Reversing them would produce incorrect Unicode
sequences for any multi-rune cluster (e.g., ligatures like lam-alef).

### Experiment

Create a unit test that shapes known Arabic words, calls
`shapedGlyphRuneAssignments()`, and verifies that concatenating all assigned
rune sequences (in cluster order) reconstructs the original input.

### Results

**HYPOTHESIS FALSE** for simple words. All tested Arabic words (without
diacritics) produce only single-rune clusters after shaping, so `reverseRunes`
has no effect. However, `reverseRunes` was later confirmed harmful for words
with diacritics (see Hypothesis 3 follow-up).

Key observations:
- Each shaped glyph maps to exactly one source rune for simple words
- Some clusters produce multiple glyphs (e.g., خ → base + dot component)
- Secondary/unmapped glyphs sit at exactly the positions where spaces appear

---

## Hypothesis 2: /W width mismatch with shaper advances

### Rationale

The /W array stores raw advance widths from the font's `hmtx` table. The
shaper may produce different advances via GPOS. If `pdftotext` compares
expected positions (from /W) with actual positions (from Tm), a mismatch
could cause spurious spaces.

### Results

**HYPOTHESIS FALSE.** Font metric widths and shaper advances match closely
(within 0.01pt) for all glyphs, including secondary/unmapped glyphs.

---

## Hypothesis 3: Unmapped CIDs cause text extraction breaks

### Rationale

When a secondary glyph gets a CID via `use()` (no ToUnicode mapping),
`pdftotext` cannot determine the Unicode character and breaks the word.

### Experiment

Generate minimal single-word PDFs and run `pdftotext` on each.

### Results

**HYPOTHESIS CONFIRMED.** With `pdftotext` available, the correlation is
definitive:

| Word         | pdftotext output    | Secondary glyphs? |
|--------------|--------------------|--------------------|
| الخبرات      | ال خبرات            | YES (glyph[5])     |
| الجيران      | ال جيران            | YES (glyph[5])     |
| الحِرَف      | ال حِ رَف           | YES (multiple)     |
| اليدوية      | اليدو ية           | YES (XOffset gap)  |
| الإنجليزية   | الإنجليز ية         | YES (XOffset gap)  |
| مرحبا        | مرحبا (correct)     | NO                 |
| بسم          | بسم (correct)       | NO                 |

Every word with unmapped secondary glyphs or XOffset position gaps has
spurious spaces. Words without these issues extract perfectly.

---

## Root Cause Analysis (Three Contributing Factors)

### Factor A: Unmapped secondary glyphs

Secondary glyphs in shaped Arabic clusters (e.g., dot components of خ, ج)
are recorded via `glyphRecorder.use()`, which creates CIDs with **no
ToUnicode mapping**. When these unmapped CIDs appear in the PDF content
stream, text extractors break the Arabic character sequence.

### Factor B: Per-glyph Tm positioning with XOffset

Each shaped glyph was emitted as a separate `Tm` + `Tj` pair. Glyphs with
non-zero XOffset (GPOS positioning adjustments) created Tm position jumps
that `pdftotext` interpreted as word boundaries. For example, in "اليدوية",
the و glyph has XOffset=1.44pt, creating a 1.44pt gap that exceeds
pdftotext's space-detection threshold (~1.13pt).

### Factor C: `reverseRunes` corrupts multi-rune cluster sequences

For clusters containing base + diacritic (e.g., حِ = ha + kasra),
`reverseRunes` reversed the rune sequence before assigning it to the
ToUnicode CMap. This produced incorrect Unicode order (kasra before ha
instead of ha before kasra), causing text extractors to output diacritics
in the wrong position.

Additionally, the old code assigned the ENTIRE multi-rune sequence to
the first glyph in a cluster, leaving subsequent glyphs unmapped. This
created additional unmapped CIDs for mark glyphs.

---

## Fix Applied

### Fix A: Map secondary glyphs to CGJ (U+034F)

For secondary glyphs that have no corresponding rune (cluster has more
glyphs than runes), map them to U+034F COMBINING GRAPHEME JOINER in the
ToUnicode CMap. CGJ is invisible, has NSM bidi class (inherits
directionality), and keeps the Arabic run together.

**Files changed:** `pdf/page_writer.go`

### Fix B: Use TJ arrays with effective widths

Changed shaped text emission from per-glyph `Tm` + `Tj` to a single
`TJ` array with inline positioning adjustments. This keeps all glyphs
in one text-showing operation, preventing `pdftotext` from splitting
them into separate words.

Additionally, compute "effective widths" for each CID that account for
GPOS XOffset adjustments. These effective widths are stored via
`glyphRecorder.setEffectiveWidth()` and used in the /W array at doc
close, making TJ adjustments near-zero so pdftotext doesn't interpret
them as spaces.

When glyphs have varying YOffset (requiring vertical positioning), the
code falls back to per-glyph `Tm` + `Tj`.

**Files changed:** `pdf/page_writer.go`, `pdf/glyph_recorder.go`,
`pdf/doc_writer.go`, `pdf/text_writer.go`

### Fix C: Distribute runes per-glyph, remove reverseRunes

Rewrote `shapedGlyphRuneAssignments()` to distribute individual runes
across cluster glyphs instead of assigning the whole sequence to the
first glyph. For clusters where glyph count equals rune count, each
glyph gets one rune in reversed order (visual L-to-R ↔ logical RTL
reversal). Removed the incorrect `reverseRunes` call.

**Files changed:** `pdf/page_writer.go`

---

## Final Verification

After all three fixes, `pdftotext` extraction results:

| Word         | Before          | After           | Status         |
|--------------|----------------|-----------------|----------------|
| الخبرات      | ال خبرات        | ال͏خبرات         | Word intact    |
| الجيران      | ال جيران        | ال͏جيران         | Word intact    |
| الحِرَف      | ال حِ رَف       | الحِرَف          | **Perfect**    |
| اليدوية      | اليدو ية       | اليدوية          | **Perfect**    |
| الإنجليزية   | الإنجليز ية     | الإنجليزية       | **Perfect**    |
| مرحبا        | مرحبا           | مرحبا            | Already OK     |
| بسم          | بسم             | بسم              | Already OK     |

Note: الخبرات and الجيران contain an invisible U+034F (CGJ) for secondary
glyphs where the cluster has more glyphs than input runes. This is
functionally correct — the word is intact, and CGJ is invisible in all
rendering contexts. All other words extract with perfect fidelity.

All existing tests pass with no regressions.
