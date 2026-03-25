# Unicode PDF Rendering — Project Plan

## Background

Leadtype currently uses a **single-byte encoding pipeline** inherited from the Type1/AFM era of PDF. Every character in a rendered string must be mapped through a codepage to a byte value in the range 0–255, and the PDF font object is declared with an encoding dictionary (typically WinAnsiEncoding plus a differences array). This approach was standard for PDF 1.2 (1996) but has been superseded by composite fonts since PDF 1.4 (2001).

### Current pipeline

```
Unicode source text
  → EachCodepage()          (rich_text: splits string into single-codepage runs)
  → CharForCodepoint()      (codepage: maps rune → byte 0–255)
  → PDF text string (bytes) (pdf: written with /Encoding WinAnsiEncoding+diffs)
  → font metrics via cmap   (ttf: glyph index used only for width lookup)
```

### Problems

1. **Hard ceiling of 256 glyphs per encoding.** CJK, Arabic, Devanagari, emoji, and anything outside the 24 supported codepages cannot be rendered at all.
2. **Multi-script text requires font-encoding switches.** A string mixing Greek and Cyrillic produces two separate PDF text segments with different font keys. Ligature and shaping boundaries are broken.
3. **The codepage layer is unnecessary for TTF/OTF.** TTF fonts already address all glyphs via a cmap (Unicode → glyph index). The current code ignores that for output and routes everything through a byte-encoding layer, applying the Type1 workflow to fonts that do not need it.
4. **No ToUnicode CMap in output.** Copy-paste, search, and accessibility all fail because the PDF contains no reverse mapping from encoded bytes back to Unicode.
5. **No font subsetting.** Entire font files are embedded. A document using a single CJK character embeds a 20–50 MB font.

---

## Goal

Produce a PDF rendering path for TTF/OTF fonts that:

- Addresses glyphs by their TTF glyph index, not by a codepage byte.
- Emits **Type 0 / CIDFont** composite font objects so any glyph in the font can be used.
- Emits a **ToUnicode CMap** for every font so text extraction and search work correctly.
- Subsets font files to include only the glyphs actually used in a document.
- Leaves the existing codepage path intact for Type1/AFM output.

---

## Architecture

### Composite font (Type 0) in PDF

The modern font declaration in the PDF file changes from:

```
% Simple font — single byte encoding
/Type /Font
/Subtype /TrueType
/BaseFont /ArialMT
/Encoding /WinAnsiEncoding
/Widths [...]
```

to:

```
% Composite font — glyph index addressing
/Type /Font
/Subtype /Type0
/BaseFont /ArialMT
/Encoding /Identity-H
/DescendantFonts [
  << /Type /Font
     /Subtype /CIDFontType2
     /BaseFont /ArialMT
     /CIDSystemInfo << /Registry (Adobe) /Ordering (Identity) /Supplement 0 >>
     /DW 1000
     /W [...]           % sparse glyph-index width array
     /CIDToGIDMap /Identity
     /FontDescriptor << ... >>
  >>
]
/ToUnicode <stream>    % CMap mapping glyph IDs back to Unicode
```

Text strings are written as big-endian uint16 glyph IDs rather than single bytes.

### Glyph ID pipeline

```
Unicode source text
  → font.cmapTable.glyphIndex(rune)   (ttf: Unicode → glyph ID, uint16)
  → PDF text string (uint16 BE pairs) (pdf: written under /Identity-H encoding)
  → ToUnicode CMap stream             (pdf: emitted alongside font object)
  → font subset stream                (pdf: only used glyphs embedded)
```

### ToUnicode CMap

A ToUnicode CMap stream maps each glyph ID used in the document back to its Unicode codepoint, enabling text extraction and copy-paste. Format:

```
/CIDInit /ProcSet findresource begin
12 dict begin
begincmap
/CIDSystemInfo << /Registry (Adobe) /Ordering (UCS) /Supplement 0 >> def
/CMapName /Adobe-Identity-UCS def
/CMapType 2 def
1 begincodespacerange
<0000> <FFFF>
endcodespacerange
N beginbfchar
<GLYPHID> <UNICODE>
...
endbfchar
endcmap
CMap end
end
```

### Font subsetting

Track which glyph IDs are referenced during document construction. Before embedding, strip the font to a subset:

- Rewrite the `glyf` table (for TrueType outlines) to include only used glyphs plus their composite dependencies.
- Rewrite the `loca` table to match.
- Update `maxp`, `hmtx`, `cmap`, `post` to reflect the subset.
- Rename the font (prefix the PostScript name with a 6-char tag, e.g. `ABCDEF+ArialMT`) per the PDF spec requirement for embedded subsets.

For CFF-based OpenType fonts, subset the `CFF ` table similarly.

---

## Phased Implementation Plan

### Phase 1 — ToUnicode CMaps for existing simple fonts (low risk)

Emit ToUnicode CMap streams for all currently produced simple-font (Type1 / TrueType) encodings. This does not change rendered output; it only adds reverse-mapping metadata so viewers can extract text correctly. The codepage's `Map()` method already provides the byte → rune mapping needed to construct the CMap.

**Deliverable:** Every font object in generated PDFs includes a `/ToUnicode` stream.

---

### Phase 2 — Composite font (Type 0) path for TTF

Add a parallel rendering path for TTF fonts that bypasses the codepage layer entirely:

1. New `Type0Font` PDF object (alongside existing `TrueTypeFont`).
2. New `CIDFont` PDF object (descendant of Type0).
3. `IdentityHEncoding` constant — no encoding dictionary needed beyond `/Identity-H`.
4. Text rendered as big-endian uint16 glyph ID pairs.
5. Glyph ID width array (`/W`) constructed directly from `hmtxTable`.
6. ToUnicode CMap generated from the glyph ID → Unicode map accumulated during rendering.
7. Opt-in at the `DocWriter` level: a flag or constructor option selects Unicode mode.

**Deliverable:** A document created with the Unicode path renders all glyphs in a TTF font, including CJK and emoji, with correct text extraction.

---

### Phase 3 — Remove EachCodepage from TTF rendering path

Once Phase 2 is stable, `EachCodepage` is no longer needed for TTF output. The `rich_text` rendering loop changes from iterating codepage segments to iterating characters directly, collecting glyph IDs per font face. `EachCodepage` is retained for the AFM/Type1 path only.

**Deliverable:** Multi-script TTF text renders in a single font-key segment. No mid-string font-encoding switches.

---

### Phase 4 — Font subsetting

Track glyph IDs referenced per font during document construction. Before `Close()`:

1. Open the original font file.
2. Build a closure of required glyph IDs (add composite glyph components).
3. Write a subset font stream containing only those glyphs.
4. Update all font objects to reference the subset stream.
5. Prefix the PostScript name with a 6-char random tag per PDF spec.

**Deliverable:** Embedded font streams contain only glyphs used in the document. File sizes for CJK-heavy documents reduce dramatically.

---

### Phase 5 — OpenType/CFF support (stretch goal)

Extend Phase 2 and Phase 4 to handle OpenType fonts with CFF outlines (`CIDFontType0` instead of `CIDFontType2`). Requires parsing the `CFF ` table for subsetting.

---

## What Stays the Same

- The `codepage/` package and its generated tables remain for Type1/AFM output.
- `EachCodepage` remains available for callers that still target simple fonts.
- All existing tests and the existing simple-font PDF output path continue to work unchanged.
- The TTF parsing infrastructure (`cmapTable`, `hmtxTable`, etc.) is already present and usable without modification.

---

## References

- [PDF Reference 1.7 §5.6 — Composite Fonts](https://opensource.adobe.com/dc-acrobat-sdk-docs/pdfstandards/pdfreference1.7old.pdf)
- [PDF Reference 1.7 §5.9 — ToUnicode CMaps](https://opensource.adobe.com/dc-acrobat-sdk-docs/pdfstandards/pdfreference1.7old.pdf)
- [OpenType spec — glyf table](https://learn.microsoft.com/en-us/typography/opentype/spec/glyf)
- [OpenType spec — CFF table](https://learn.microsoft.com/en-us/typography/opentype/spec/cff)
