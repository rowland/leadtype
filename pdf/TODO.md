# PDF Parity TODO

This checklist tracks the `pdf`-layer parity work approved from the legacy
`eideticpdf` project.

## Approved Parity Work

### Path API ([#9](https://github.com/rowland/leadtype/issues/9))

- [x] Add a public path lifecycle API that mirrors the useful legacy behavior
  - [x] expose a `path` entry point for manual multi-segment path construction
  - [x] expose `fill`
  - [x] expose `stroke`
  - [x] expose `fill_and_stroke`
  - [x] expose `clip`
  - [x] define how path mode interacts with existing auto-path behavior
  - [x] add unit tests for path state transitions
  - [x] add integration tests for clipping and hollow-path scenarios

### Shape Primitives ([#7](https://github.com/rowland/leadtype/issues/7))

- [x] Add higher-level geometry helpers on top of the existing graph writer
  - [x] angled `line`
  - [x] `circle`
  - [x] `ellipse`
  - [x] `arc`
  - [x] `pie`
  - [x] `arch`
  - [x] `polygon`
  - [x] `star`
- [x] share curve/path internals where possible so shapes do not duplicate low-level logic
- [x] support fill, stroke, reverse-direction, and clip behaviors where legacy parity expects them
- [x] add focused unit tests for geometry generation
- [x] add PDF-level tests for rendered command sequences

### Images ([#8](https://github.com/rowland/leadtype/issues/8), future PNG follow-on [#16](https://github.com/rowland/leadtype/issues/16))

- [x] Add image support at the PDF layer
  - [x] support JPEG detection and metadata extraction
  - [x] support image placement from file input
  - [x] support image placement from in-memory data
  - [x] preserve aspect ratio when only one dimension is supplied
  - [x] define behavior when width and height are both supplied
- [x] add tests for JPEG dimension parsing
- [x] add PDF integration tests for image XObject output and placement
- [x] keep the public image API format-agnostic so PNG can be added later without API churn
- [x] keep the image object model open to future PNG import via decoded pixel data, `/FlateDecode`, and optional `/SMask`

### Future Image Extension ([#16](https://github.com/rowland/leadtype/issues/16))

- [ ] Add PNG support as a follow-on enhancement
  - [ ] support a first milestone of non-interlaced 8-bit grayscale, RGB, and RGBA PNGs
  - [ ] decode PNG scanlines and map them to standard PDF image XObjects
  - [ ] support alpha via `/SMask`
  - [ ] decide whether palette PNGs are expanded to RGB or mapped to Indexed color space in the first implementation
  - [ ] defer 16-bit PNG, Adam7 interlacing, and advanced color-management handling unless required

### Transforms ([#12](https://github.com/rowland/leadtype/issues/12))

- [x] Add block-scoped transform helpers
  - [x] `rotate(angle, x, y, ...)`
  - [x] `scale(x, y, scaleX, scaleY, ...)`
- [x] implement graphics-state save/restore around transforms
- [x] verify transforms compose correctly with text, images, and shapes
- [x] add tests for emitted transform matrices

### Vertical Text Alignment ([#12](https://github.com/rowland/leadtype/issues/12))

- [x] Add vertical text alignment control comparable to legacy `v_text_align`
- [x] define supported alignment values and their metric anchors
- [x] integrate with text placement without regressing underline/strikeout
- [x] add unit tests covering each alignment mode
- [x] add PDF-level tests proving baseline shifts in output

### Stream Compression

- [x] Add fluent `DocWriter` compression configuration ([#20](https://github.com/rowland/leadtype/issues/20))
  - [x] add `CompressPages(bool) *DocWriter`
  - [x] add `CompressToUnicode(bool) *DocWriter`
  - [x] add `CompressEmbeddedFonts(bool) *DocWriter`
  - [x] default all compression flags to `false` for compatibility
- [x] Add stream-level Flate compression support in the PDF object model ([#20](https://github.com/rowland/leadtype/issues/20))
  - [x] support emitting compressed stream data
  - [x] set `/Filter /FlateDecode` when compression is enabled
  - [x] preserve `/Length1` semantics for embedded font streams
- [x] Apply compression selectively by stream type
  - [x] compress page content streams when `CompressPages(true)` is set ([#19](https://github.com/rowland/leadtype/issues/19))
  - [x] compress ToUnicode streams when `CompressToUnicode(true)` is set ([#17](https://github.com/rowland/leadtype/issues/17))
  - [x] compress embedded font streams when `CompressEmbeddedFonts(true)` is set ([#17](https://github.com/rowland/leadtype/issues/17))
- [x] keep compression decisions at stream creation time rather than as a late global rewrite
- [x] add unit tests for compressed stream dictionaries and round-trippable payloads ([#20](https://github.com/rowland/leadtype/issues/20))
- [x] add document-level tests proving compressed and uncompressed output modes both work ([#19](https://github.com/rowland/leadtype/issues/19), [#17](https://github.com/rowland/leadtype/issues/17))
- [x] keep existing string-based PDF tests readable by leaving compression opt-in rather than default-on

## Deferred

- [ ] Virtual page imposition / pages-up support
  - [ ] `pages_up`
  - [ ] `pages_up_layout`
  - [ ] `unscaled`

This stays out of the first parity milestone.

## Excluded

The following legacy features were reviewed and intentionally excluded from the
parity effort:

- [ ] tabs / vertical tabs / indent controls
  - [ ] `tabs`
  - [ ] `tab`
  - [ ] `vtabs`
  - [ ] `vtab`
  - [ ] `indent`

These should not be pulled into parity work unless scope changes later.

## Acceptance Checklist

- [x] `go test ./pdf` passes
- [x] affected cross-package tests pass
- [x] any new serialized PDF behavior has stable test coverage
- [x] public APIs are documented in code comments or adjacent docs where needed
