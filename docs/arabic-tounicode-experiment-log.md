# Arabic ToUnicode Experiment Log

This log records the experiments that led from commit
`1f7435129ae3c7381dd69f64db8751c63a5760cc` to the current Arabic
round-trip recovery. The older
[`arabic-tounicode-bug-investigation.md`](arabic-tounicode-bug-investigation.md)
note remains the historical snapshot of the first investigation; this file is
the running lab notebook for subsequent iterations.

## 2026-04-06 — Baseline From `1f74351`

- Baseline commit: `1f7435129ae3c7381dd69f64db8751c63a5760cc`
- Hypothesis: The committed patch fixes the practical `pdftotext` failures, but
  it does so by introducing semantic shortcuts that should be removed later.
- Fixture words: `الخبرات`, `الجيران`, `الحِرَف`, `اليدوية`, `الإنجليزية`,
  `مرحبا`, `بسم`
- Command:
  `go test ./pdf -run TestPdftotextArabicWords -v`
- Raw extracted output with codepoints:
  `الخبرات` => `"\u202bال͏خبرات\u202c"` (`U+202B 0627 0644 034F 062E 0628 0631 0627 062A 202C`)
  `الجيران` => `"\u202bال͏جيران\u202c"` (`U+202B 0627 0644 034F 062C 064A 0631 0627 0646 202C`)
  `اليدوية` => `"\u202bاليدوية\u202c"` (`U+202B 0627 0644 064A 062F 0648 064A 0629 202C`)
- PDF content-stream / CMap observations:
  secondary glyphs were kept in the content stream by mapping them to
  `U+034F`, and `/W` widths were replaced with per-occurrence “effective”
  widths so the `TJ` adjustments stayed near zero.
- Conclusion:
  the patch did eliminate the visible extraction failures, but it did not
  preserve authored Unicode exactly and it made `/W` context-sensitive.

## 2026-04-06 — Empty Destinations + Raw `/W`

- Baseline commit: `1f7435129ae3c7381dd69f64db8751c63a5760cc`
- Hypothesis: Replacing CGJ with empty ToUnicode destinations and restoring raw
  `/W` widths would keep extraction correct while removing the semantic hacks.
- Fixture words: `الخبرات`, `الجيران`, `الحِرَف`, `اليدوية`, `الإنجليزية`
- Command:
  `go test ./pdf -run TestPdftotextArabicWords -v`
- Raw extracted output with codepoints:
  `الخبرات` => `"\u202bال خبرات\u202c"` (`U+202B 0627 0644 0020 062E 0628 0631 0627 062A 202C`)
  `الحِرَف` => `"\u202bالح ِر َف\u202c"` (`U+202B 0627 0644 062D 0020 0650 0631 0020 064E 0641 202C`)
  `اليدوية` => `"\u202bاليدو ية\u202c"` (`U+202B 0627 0644 064A 062F 0648 0020 064A 0629 202C`)
- PDF content-stream / CMap observations:
  the ToUnicode CMap did contain entries for all CIDs, including `<>` for
  zero-text glyphs, and the content stream used raw-width `TJ` adjustments such
  as `-119.8021` and `0.9375`.
- Conclusion:
  empty destinations plus raw `/W` were more principled, but Poppler still
  interpreted zero-text glyphs and large raw-width kerning adjustments as word
  boundaries.

## 2026-04-06 — Marked-Content `/ActualText` Probe

- Baseline commit: `1f7435129ae3c7381dd69f64db8751c63a5760cc`
- Hypothesis: A marked-content `/ActualText` override could preserve exact
  extraction without reintroducing CGJ or synthetic `/W` values.
- Fixture words: `الخبرات`, `اليدوية`
- Commands:
  `go test ./pdf -run TestPdftotextArabicWords/الخبرات -v`
  `go test ./pdf -run TestPdftotextArabicWords/اليدوية -v`
- Raw extracted output with codepoints:
  logical `/ActualText` produced reversed output without spaces:
  `الخبرات` => `"\u202bتاربخلا\u202c"` (`U+202B 062A 0627 0631 0628 062E 0644 0627 202C`)
  `اليدوية` => `"\u202bةيوديلا\u202c"` (`U+202B 0629 064A 0648 062F 064A 0644 0627 202C`)
- PDF content-stream / CMap observations:
  the `/Span << /ActualText ... >> BDC` wrapper changed Poppler’s extraction
  behavior enough to remove the spurious spaces, but supplying logical Arabic in
  `/ActualText` yielded reversed codepoint order in the extracted text.
- Conclusion:
  marked-content `/ActualText` was necessary for extractor stability, but for
  Arabic it had to be supplied in visual order rather than logical order to make
  Poppler return the authored logical text after stripping its bidi wrappers.

## 2026-04-06 — Final Accepted Behavior

- Baseline commit: `1f7435129ae3c7381dd69f64db8751c63a5760cc`
- Hypothesis: The correct combination is:
  raw `/W`, complete CID-to-ToUnicode coverage with `<>` for zero-text glyphs,
  logical-order CID emission for shaped Arabic runs, and marked-content
  `/ActualText` supplied in visual order for Poppler compatibility.
- Fixture words: `الخبرات`, `الجيران`, `الحِرَف`, `اليدوية`, `الإنجليزية`,
  `صُممت`, `مرحبا`, `بسم`
- Commands:
  `go test ./pdf -run TestPdftotextArabicWords -v`
  `go test ./pdf -run TestInspectContentStream_XOffsetWord -v`
  `go test ./pdf -run 'TestUnicodeMode_ArabicExtractionRoundTripsExactly|TestUnicodeMode_ArabicPageTextHasNoUnmappedCIDs|TestUnicodeMode_ArabicCurvedTextHasNoUnmappedCIDs|TestUnicodeMode_ShapedGlyphOffsetsWithZeroYOffsetUseTJAndRawWidths'`
- Raw extracted output with codepoints:
  `الخبرات` => `"\u202bالخبرات\u202c"` (`U+202B 0627 0644 062E 0628 0631 0627 062A 202C`)
  `اليدوية` => `"\u202bاليدوية\u202c"` (`U+202B 0627 0644 064A 062F 0648 064A 0629 202C`)
  `الإنجليزية` => `"\u202bالإنجليزية\u202c"` (`U+202B 0627 0644 0625 0646 062C 0644 064A 0632 064A 0629 202C`)
- PDF content-stream / CMap observations:
  the current `اليدوية` content stream is:
  `[<0007>391.4792 <0006>418.4896 <0005>715.3542 <0004>870.7396 <0003>709.8958 <0002>568.9063 <0001>] TJ`
  with `/ActualText <feff0629064a0648062f064a06440627>` in the enclosing marked
  content. The ToUnicode CMap remains CID-based and raw-metric based:
  CID `0001` → `U+0629`, ..., CID `0007` → `U+0627`. Words with decorative
  secondary glyphs still emit `<>` for the zero-text CID rather than omitting
  the CMap entry.
- Conclusion:
  this combination removed both improper strategies from the original patch
  while keeping the real Arabic extraction bugs fixed. `/W` now records real
  font metrics, every emitted CID has a ToUnicode entry, and the regression
  tests round-trip authored Arabic exactly after stripping only Poppler’s bidi
  wrappers.

# ChatGPT Verdict

1. Character Integrity
	•	All Arabic letters render correctly (no replacement glyphs like �).
	•	No evidence of mojibake (e.g., Latin-1 misinterpretation of UTF-8).
	•	Diacritics (e.g., الحِرَف) display properly where used.
	•	Ligatures (e.g., لا) are preserved.

Conclusion: Encoding is consistent with proper Unicode Arabic text.

⸻

2. Word and Sentence Continuity
	•	Words are not split incorrectly.
	•	No missing characters within words.
	•	Sentence flow is natural and grammatically intact.

Example (clean):

تدعو مكتبة المدينة سكان الحي إلى برنامج أسبوعي يجمع بين القراءة الإبداعية…

⸻

3. Layout and Line Wrapping
	•	Line breaks occur in expected positions for paragraph wrapping.
	•	No mid-word breaks due to extraction issues.
	•	Table-like section (“جدول مختصر”) is slightly flattened (expected in text extraction), but:
	•	Content remains readable
	•	No structural corruption—just loss of visual tabular formatting

⸻

4. Punctuation and Symbols
	•	Arabic punctuation (،) is consistent.
	•	Hyphen and colon usage appear normal.
	•	Email address is intact:
	•	example.citylibrary@community

⸻

5. Directionality (RTL)
	•	Text content is logically ordered.
	•	No indication of bidirectional (BiDi) corruption (a common issue in Arabic PDFs).
	•	Mixed elements (Arabic + email/Latin text) are handled correctly.

⸻

6. Potential Minor Observations (Non-critical)

These are not corruption issues, just typical PDF extraction artifacts:
	•	Table formatting is linearized.
	•	Some long lines may reflect layout rather than semantic grouping.

⸻

Final Verdict
	•	No text corruption detected
	•	The document appears to be:
	•	Properly encoded (Unicode)
	•	Linguistically intact
	•	Structurally coherent (within limits of plain-text extraction)
