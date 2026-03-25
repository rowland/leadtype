# TODO

## Review Follow-Ups

- [x] Preserve shaping offsets during PDF text emission
  The shaped TrueType path in [pdf/page_writer.go](pdf/page_writer.go) currently serializes only glyph IDs and emits them with a plain `Tj` operator. That drops `XOffset` and `YOffset` from `shaping.GlyphPosition`, so any shaped glyph that depends on per-glyph positioning will render incorrectly even though shaping itself succeeded.

  Why this matters:
  Complex-script shaping does not only choose glyph IDs. It also computes glyph placement. Combining marks and other positioned glyphs can require non-zero offsets relative to the base glyph. Ignoring those offsets means the output PDF can be visibly wrong.

  Suggested direction:
  Update the shaped emission path so glyph positioning data is honored. That may require switching from a simple `Tj` stream to a more expressive text operator sequence, or otherwise incorporating the shaping offsets into the emitted text layout. Add focused coverage for Arabic text with combining marks or another fixture that produces non-zero offsets.

- [x] Record full Unicode clusters for ligatures in ToUnicode generation
  The current glyph recorder in [pdf/glyph_recorder.go](pdf/glyph_recorder.go) stores a single `rune` per glyph ID, and the shaped-text path records `runes[gp.ClusterIndex]` from [pdf/page_writer.go](pdf/page_writer.go). For ligatures, `ClusterIndex` identifies only the first rune in the source cluster, not the whole cluster.

  Why this matters:
  A shaped ligature glyph can represent multiple Unicode code points. When ToUnicode maps that glyph back to only the first rune, PDF copy/paste, search, and text extraction become lossy. Arabic ligatures are the clearest example, but the issue is more general than Arabic.

  Suggested direction:
  Change the recording model so a glyph can map to a full Unicode sequence, not just a single rune. Then update composite ToUnicode generation to emit the full destination sequence for ligature glyphs. Add a test that shapes a known ligature and verifies extraction-oriented data preserves every code point in the original cluster.

- [x] Encode non-BMP Unicode correctly in composite ToUnicode CMaps
  The composite ToUnicode writer in [pdf/to_unicode.go](pdf/to_unicode.go) emits destination values with `<%04X>`, which only represents a single 16-bit unit. That is not sufficient for runes above `U+FFFF`, which must be written as UTF-16BE surrogate pairs in a ToUnicode CMap.

  Why this matters:
  Any supplementary-plane character supported by an embedded TrueType font will be truncated or mis-encoded in extraction/search metadata. Rendering may still appear correct if the glyph is present, but round-tripping text from the PDF will fail.

  Suggested direction:
  Convert destination Unicode values to UTF-16BE before writing the `bfchar` entries. For BMP characters, keep the existing one-unit form. For supplementary-plane characters, emit both surrogate code units in sequence. Add a test using a rune above `U+FFFF` to confirm the generated CMap contains the expected surrogate-pair encoding.

## Sample 005 Follow-Ups

- [x] Make cmap loading resilient to modern Unicode subtables and broaden preferred-subtable selection
  Running `go run ./samples -all` surfaced multiple `test_005_ttf_fonts` failures that look like `cmap` support gaps rather than fundamentally unreadable fonts. The clearest symptoms were:
  `Unsupported mapping table format: 14`
  `Unable to select preferred cmap glyph indexer.`

  Current analysis:
  In [ttf/cmap_table.go](ttf/cmap_table.go), `readMapping` aborts the entire font load on any unsupported subtable format. That is especially problematic for format 14, which carries Unicode variation-selector data and often coexists with an ordinary format 4 or format 12 Unicode mapping that would be sufficient for our renderer. Rejecting the whole font because it also contains format 14 is stricter than necessary.

  The preferred-subtable selection logic is also narrow. It recognizes Unicode platform encodings and Microsoft UCS-2 (`platform 3, specific 1`), but it does not appear to account for modern Microsoft full-Unicode encodings that pair with format 12. That likely explains the `Unable to select preferred cmap glyph indexer.` failures on otherwise mainstream fonts such as Helvetica Neue.

  Why this matters:
  Sample 5 is intended to enumerate system TrueType fonts. Right now a meaningful chunk of modern system fonts get dropped because the loader insists on a small subset of `cmap` layouts. That reduces practical compatibility for both sample code and real document generation.

  Suggested direction:
  Change `cmap` loading so unsupported auxiliary subtables, especially format 14, are skipped rather than treated as fatal when another supported Unicode mapping exists. Expand preferred-indexer selection to include modern Microsoft full-repertoire encodings and, more generally, prefer any successfully parsed Unicode-capable subtable over failing the font outright. The corresponding subset-selection logic in [ttf/subset.go](ttf/subset.go) should stay in sync.

- [ ] Relax post format 2 parsing so malformed glyph-name arrays do not reject otherwise usable fonts
  Sample 5 also surfaced many failures of the form:
  `invalid post format 2 glyph name index 311 (newNames len=53)`

  Current analysis:
  In [ttf/post_table.go](ttf/post_table.go), `readFormat2Names` returns an error as soon as any custom glyph-name index points past the decoded `newNames` slice. That is a reasonable validation rule for a pristine parser, but in this codebase it rejects the entire font even though the rest of the font may be perfectly usable for rendering. The library does not appear to rely on those glyph names for normal text layout; the main uses of `post` today are header metrics like italic angle, underline metrics, fixed-pitch detection, and subset serialization.

  Why this matters:
  Rejecting a font over broken glyph-name metadata is harsher than needed, especially in sample 5 where the goal is to show what the system can render. It also blocks some fonts that could likely be measured, embedded, and subsetted successfully if we degraded gracefully.

  Suggested direction:
  Make format 2 glyph-name decoding non-fatal when the table is internally inconsistent. Reasonable fallbacks include leaving unknown names blank, synthesizing placeholder names, or treating the subsetted output more like post format 3 when names cannot be trusted. The important part is to preserve the rest of the font's usable metrics instead of failing the entire load. Add focused tests around malformed format 2 tables to lock in the relaxed behavior.
