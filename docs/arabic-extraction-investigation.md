# Arabic Extraction Investigation

This note summarizes the recent Arabic PDF extraction experiments after the
visual shaping and bidi layout work had largely stabilized.

## Current Baseline

Committed baseline before the latest uncommitted experiments:

- `808c0f5` `pdf: separate cids for distinct unicode mappings`

That change improved `ToUnicode` identity mapping by assigning distinct PDF
CIDs to distinct Unicode mappings, even when multiple mappings reused the same
source glyph ID.

Observed result at that baseline:

- Visual rendering: good
- Subtitle/header extraction: improved
- Major wrong-letter / wrong-codepoint corruption: improved
- Remaining extraction problem: Arabic words still split internally with spaces

Examples from `pdftotext` at that stage:

- `ال خبرات`
- `ال جيران`
- `ال حِ رَف`
- `اليدو ية`
- `الإنجليز ية`

## Hypotheses Tested

### 1. Per-glyph `Tm`/`Tj` caused the splits

Experiment:

- Changed shaped-run emission to use a single `TJ` run when possible, instead
  of one `Tm` + `Tj` per glyph.

Result:

- The sample extraction did not materially improve.

Conclusion:

- Per-glyph positioning alone is not the full cause of the remaining splits.

### 2. Multiple `TJ` elements inside one run caused the splits

Experiment:

- Coalesced adjacent glyph codes into a single hex string inside `TJ`, only
  breaking the `TJ` array when a numeric adjustment was needed.

Result:

- Minimal single-word PDFs still extracted with internal spaces.

Conclusion:

- `TJ` element fragmentation alone is not the full cause.

### 3. Unmapped secondary glyphs in a cluster caused the splits

Experiment:

- For single-rune clusters that shaped to multiple glyphs, temporarily mapped
  the same Unicode rune onto every glyph in the cluster instead of leaving
  secondary glyphs unmapped.

Result:

- Extraction did not change for the test word `الخبرات`.

Conclusion:

- Unmapped secondary glyphs alone are not the full cause.

## Low-level Findings

Minimal one-word PDFs reproduced the same extraction problem as the full LTML
sample. This is good: the issue can be studied without the full sample.

Example minimal-word outputs:

- `الخبرات` => `ال خبرات`
- `الجيران` => `ال جيران`
- `الحِرَف` => `ال حِ رَف`
- `اليدوية` => `اليدو ية`
- `الإنجليزية` => `الإنجليز ية`

This means the remaining defect is not primarily about paragraph layout,
tables, or multi-line LTML behavior.

Shaping inspection on `Amiri-Regular.ttf` showed:

- Many problematic words have `YOffset == 0` for all glyphs.
- Some words have non-zero `XOffset` values, but not all.
- Some problematic words include multi-glyph clusters for a single source rune.

This suggests the remaining split may be caused by how extractors interpret the
overall shaped glyph stream, not by one single missing map entry.

## What Was Learned

The previous committed `ToUnicode` CID fix was worthwhile:

- it reduced wrong-letter / wrong-codepoint corruption
- it preserved Unicode identity better for reused glyph IDs

The recent uncommitted emission experiments were informative but did not yet
improve the real sample enough to warrant keeping them as-is.

## Recommended Next Step

Compare the same minimal Arabic word across different shapers and/or content
forms from a clean baseline:

1. Revert the current uncommitted emission experiments.
2. Keep baseline `808c0f5`.
3. Compare go-text shaping vs HarfBuzz shaping for the same words.
4. For each shaper, compare extraction of a tiny one-word PDF.

The key question is now:

> Is the remaining split caused by our PDF text emission, or by the semantic
> structure of the shaped glyph stream itself?

## Uncommitted Experiment Files At Time Of Writing

These were the active experiment files when this note was written:

- `pdf/page_writer.go`
- `pdf/objects.go`
- `pdf/composite_font_test.go`
- `ltml/samples/test_033_arabic_program.pdf`

These changes were exploratory and should be evaluated before keeping.
