# PDF Parity TODO

This checklist tracks the `pdf`-layer parity work approved from the legacy
`eideticpdf` project.

## Approved Parity Work

### Path API

- [ ] Add a public path lifecycle API that mirrors the useful legacy behavior
  - [ ] expose a `path` entry point for manual multi-segment path construction
  - [ ] expose `fill`
  - [ ] expose `stroke`
  - [ ] expose `fill_and_stroke`
  - [ ] expose `clip`
  - [ ] define how path mode interacts with existing auto-path behavior
  - [ ] add unit tests for path state transitions
  - [ ] add integration tests for clipping and hollow-path scenarios

### Shape Primitives

- [ ] Add higher-level geometry helpers on top of the existing graph writer
  - [ ] angled `line`
  - [ ] `circle`
  - [ ] `ellipse`
  - [ ] `arc`
  - [ ] `pie`
  - [ ] `arch`
  - [ ] `polygon`
  - [ ] `star`
- [ ] share curve/path internals where possible so shapes do not duplicate low-level logic
- [ ] support fill, stroke, reverse-direction, and clip behaviors where legacy parity expects them
- [ ] add focused unit tests for geometry generation
- [ ] add PDF-level tests for rendered command sequences

### Images

- [ ] Add image support at the PDF layer
  - [ ] support JPEG detection and metadata extraction
  - [ ] support image placement from file input
  - [ ] support image placement from in-memory data
  - [ ] preserve aspect ratio when only one dimension is supplied
  - [ ] define behavior when width and height are both supplied
- [ ] add tests for JPEG dimension parsing
- [ ] add PDF integration tests for image XObject output and placement

### Transforms

- [ ] Add block-scoped transform helpers
  - [ ] `rotate(angle, x, y, ...)`
  - [ ] `scale(x, y, scaleX, scaleY, ...)`
- [ ] implement graphics-state save/restore around transforms
- [ ] verify transforms compose correctly with text, images, and shapes
- [ ] add tests for emitted transform matrices

### Vertical Text Alignment

- [ ] Add vertical text alignment control comparable to legacy `v_text_align`
- [ ] define supported alignment values and their metric anchors
- [ ] integrate with text placement without regressing underline/strikeout
- [ ] add unit tests covering each alignment mode
- [ ] add PDF-level tests proving baseline shifts in output

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

- [ ] `go test ./pdf` passes
- [ ] affected cross-package tests pass
- [ ] any new serialized PDF behavior has stable test coverage
- [ ] public APIs are documented in code comments or adjacent docs where needed
