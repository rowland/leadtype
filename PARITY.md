# Parity Plan

This document tracks feature parity between the current Go-based `leadtype`
project and the older Ruby codebases:

- `~/src/eideticpdf` for PDF-generation features
- `~/src/eideticrml` for markup/layout features

The goal is capability parity where it still makes sense for `leadtype`, not
byte-for-byte or API-for-API compatibility. Existing `leadtype` features that
already exceed the legacy implementation, especially the newer Unicode/TTF
pipeline, stay in place.

## Source Basis

The feature inventory below was derived from:

- legacy docs, samples, specs, and library code in `eideticpdf`
- legacy docs, samples, specs, and library code in `eideticrml`
- current `leadtype` packages:
  - `/Users/brent/src/leadtype/pdf`
  - `/Users/brent/src/leadtype/ltml`

## Legacy Feature Inventory

### `eideticpdf`

Major feature areas discovered in the older project:

- document lifecycle and page setup
  - document open/close
  - per-page open/close/new page
  - units, margins, page sizing, rotation
  - virtual page imposition (`pages_up`, layout direction, unscaled mode)
- movement and text cursor controls
  - move to / move by / current pen position
  - tabs, vertical tabs, indentation
- drawing and paths
  - raw line and Bezier support
  - path lifecycle helpers
  - clipping
  - line width, cap style, dash pattern
- shape primitives
  - rectangle
  - angled line
  - circle, ellipse
  - arc, pie, arch
  - polygon, star
- colors and fonts
  - named colors
  - fill, line, and font colors
  - Type1 font selection and styling
  - text fill/stroke behavior
  - vertical text alignment
- text layout
  - print, puts, print-at-position
  - wrapping and paragraph helpers
  - bullets
- images
  - JPEG detection
  - JPEG dimension extraction
  - image placement from file or in-memory data
- transforms
  - rotate around an origin
  - scale around an origin

### `eideticrml`

Major feature areas discovered in the older project:

- document/page/style/rule system
  - styles for page, font, pen, brush, paragraph, bullet, layout
  - aliases / shorthand tags
  - CSS-like rule matching
- layout managers
  - vbox, hbox, table, flow
  - absolute and relative layout behavior
  - overflow-aware continuation behavior
- core widgets
  - document, page, container, paragraph, span, rectangle
- additional widgets
  - label
  - preformatted text
  - image
  - line
  - page number
- shape widgets
  - circle, ellipse
  - polygon, star
  - arc, pie, arch
- page-flow and repeated-content behavior
  - `display` modes such as `always`, `first`, `succeeding`, `even`, `odd`
  - header/footer style repeated content
  - overflow continuation across pages
- transform and positioning features
  - absolute / relative positioning attrs
  - `shift`
  - rotation around configurable origins
  - z-index-like stacking hints

## Current `leadtype` Assessment

### Already Implemented Or Substantially Present

These areas already exist in `leadtype` and are not parity gaps:

- document/page creation and PDF serialization
- page size, orientation, margins, and units
- text output, paragraph wrapping, rich text, bullets, underline, strikeout
- line color, fill color, line width, dash pattern, and line cap style
- rectangles and raw Bezier curve support
- LTML styles, aliases, rules, and standard layouts `vbox`, `hbox`, `table`,
  and `flow`
- LTML core tags:
  - `<ltml>`
  - `<page>`
  - `<p>`
  - `<span>`
  - `<div>`
  - `<rect>`
- Unicode/composite-font support in the PDF layer, which is newer than the
  legacy implementation

### Missing And Approved For Parity

These gaps are intentionally in scope for parity work.

#### PDF

- explicit path API
  - `path`
  - `fill`
  - `stroke`
  - `fill_and_stroke`
  - `clip`
- higher-level shape primitives
  - angled `line`
  - `circle`
  - `ellipse`
  - `arc`
  - `pie`
  - `arch`
  - `polygon`
  - `star`
- image embedding and placement
- transform helpers
  - `rotate`
  - `scale`
- vertical text alignment control

#### LTML

- `<label>`
- `<pre>`
- `<image>`
- `<line>`
- shape widgets beyond `<rect>`
  - `<circle>`
  - `<ellipse>`
  - `<polygon>`
  - `<star>`
  - `<arc>`
  - `<pie>`
  - `<arch>`
- `<pageno>`
- working absolute and relative positioning behavior
- page-flow and display behavior
  - repeated header/footer style content
  - overflow continuation
  - `display` modes
- transform-related widget attributes
  - `rotate`
  - `origin_x`
  - `origin_y`
  - `shift`
  - `z_index`

### Missing And Deferred

- PDF virtual page imposition support
  - `pages_up`
  - `pages_up_layout`
  - `unscaled`

### Missing And Explicitly Excluded

- PDF tab stop and indent controls
  - `tabs`
  - `tab`
  - `vtabs`
  - `vtab`
  - `indent`

## Decision Table

| Feature | Legacy Source | Current Status | Decision | Notes |
|---|---|---|---|---|
| PDF tabs / vertical tabs / indent | `eideticpdf` | Not implemented as supported public behavior | Exclude | Keep out of parity scope |
| PDF explicit path API | `eideticpdf` | Low-level PDF operators exist, no public parity API | Include | Needed for methodical shape and clip work |
| PDF higher-level shape primitives | `eideticpdf` | Only rectangles and raw curve helpers are present | Include | Add user-facing geometry helpers |
| PDF images | `eideticpdf` | Not implemented | Include | Legacy parity requires image placement support |
| PDF transforms | `eideticpdf` | Not implemented as public block helpers | Include | Needed directly and for LTML transforms |
| PDF vertical text alignment | `eideticpdf` | Mentioned in legacy, missing in current writer | Include | Separate from existing underline/strikeout |
| PDF pages-up / imposition | `eideticpdf` | Not implemented | Defer | Useful but not needed for first parity pass |
| LTML `<label>` | `eideticrml` | Alias `<br>` points at missing label tag | Include | Unblocks line-break helper and label widget |
| LTML `<pre>` | `eideticrml` | Not implemented | Include | Needed for preformatted/codeblock parity |
| LTML `<image>` | `eideticrml` | Not implemented | Include | Depends on PDF image support |
| LTML `<line>` | `eideticrml` | Not implemented | Include | Thin wrapper over PDF line drawing |
| LTML shape widgets | `eideticrml` | Not implemented beyond `<rect>` | Include | Depends on PDF shape support |
| LTML `<pageno>` | `eideticrml` | Not implemented | Include | Needed for headers/footers parity |
| LTML absolute/relative positioning | `eideticrml` | Layout names exist, behavior is placeholder/incomplete | Include | Must make current documented behavior real |
| LTML display / overflow page flow | `eideticrml` | Not implemented end-to-end | Include | Needed for repeated content and continuation |
| LTML transforms / stacking attrs | `eideticrml` | Not implemented | Include | Depends on PDF transforms and widget placement |

## Delivery Strategy

The implementation trackers for approved parity work live in:

- `/Users/brent/src/leadtype/pdf/TODO.md`
- `/Users/brent/src/leadtype/ltml/TODO.md`

Those files break the approved parity scope into checklists so the work can be
completed incrementally without losing the parity boundary.

## Suggested Execution Order

A practical order for implementation is:

1. Add core PDF drawing APIs:
   path lifecycle, shapes, transforms, vertical text alignment, images.
2. Add LTML leaf widgets that map directly to those PDF capabilities:
   label, pre, line, image, shapes, page number.
3. Finish LTML positioning and transform semantics:
   absolute/relative layout, shift, rotate, origin handling, stacking hints.
4. Finish LTML page-flow behavior:
   display modes, overflow continuation, repeated content handling.
5. Reassess deferred PDF pages-up support after the first parity milestone.
