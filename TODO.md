# TODO

## Review Follow-Ups

- [x] Preserve shaping offsets during PDF text emission
  The shaped TrueType path in [pdf/page_writer.go](/Users/brent/src/leadtype/pdf/page_writer.go) currently serializes only glyph IDs and emits them with a plain `Tj` operator. That drops `XOffset` and `YOffset` from `shaping.GlyphPosition`, so any shaped glyph that depends on per-glyph positioning will render incorrectly even though shaping itself succeeded.

  Why this matters:
  Complex-script shaping does not only choose glyph IDs. It also computes glyph placement. Combining marks and other positioned glyphs can require non-zero offsets relative to the base glyph. Ignoring those offsets means the output PDF can be visibly wrong.

  Suggested direction:
  Update the shaped emission path so glyph positioning data is honored. That may require switching from a simple `Tj` stream to a more expressive text operator sequence, or otherwise incorporating the shaping offsets into the emitted text layout. Add focused coverage for Arabic text with combining marks or another fixture that produces non-zero offsets.

- [ ] Record full Unicode clusters for ligatures in ToUnicode generation
  The current glyph recorder in [pdf/glyph_recorder.go](/Users/brent/src/leadtype/pdf/glyph_recorder.go) stores a single `rune` per glyph ID, and the shaped-text path records `runes[gp.ClusterIndex]` from [pdf/page_writer.go](/Users/brent/src/leadtype/pdf/page_writer.go). For ligatures, `ClusterIndex` identifies only the first rune in the source cluster, not the whole cluster.

  Why this matters:
  A shaped ligature glyph can represent multiple Unicode code points. When ToUnicode maps that glyph back to only the first rune, PDF copy/paste, search, and text extraction become lossy. Arabic ligatures are the clearest example, but the issue is more general than Arabic.

  Suggested direction:
  Change the recording model so a glyph can map to a full Unicode sequence, not just a single rune. Then update composite ToUnicode generation to emit the full destination sequence for ligature glyphs. Add a test that shapes a known ligature and verifies extraction-oriented data preserves every code point in the original cluster.

- [ ] Encode non-BMP Unicode correctly in composite ToUnicode CMaps
  The composite ToUnicode writer in [pdf/to_unicode.go](/Users/brent/src/leadtype/pdf/to_unicode.go) emits destination values with `<%04X>`, which only represents a single 16-bit unit. That is not sufficient for runes above `U+FFFF`, which must be written as UTF-16BE surrogate pairs in a ToUnicode CMap.

  Why this matters:
  Any supplementary-plane character supported by an embedded TrueType font will be truncated or mis-encoded in extraction/search metadata. Rendering may still appear correct if the glyph is present, but round-tripping text from the PDF will fail.

  Suggested direction:
  Convert destination Unicode values to UTF-16BE before writing the `bfchar` entries. For BMP characters, keep the existing one-unit form. For supplementary-plane characters, emit both surrogate code units in sequence. Add a test using a rune above `U+FFFF` to confirm the generated CMap contains the expected surrogate-pair encoding.
