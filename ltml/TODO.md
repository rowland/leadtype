# LTML Parity TODO

This checklist tracks the `ltml`-layer parity work approved from the legacy
`eideticrml` project.

## Approved Parity Work

### Foundational Widget Tags ([#11](https://github.com/rowland/leadtype/issues/11))

- [x] Add a standard `<label>` widget
  - [x] support simple text rendering
  - [x] make the built-in `<br>` alias resolve to working behavior
  - [x] decide whether label is inline-only, block-only, or dual-purpose and document it
- [x] Add a standard `<pre>` widget
  - [x] preserve preformatted whitespace/newlines
  - [x] support inline code-block parity use cases from legacy samples
- [x] Add a standard `<image>` widget ([#10](https://github.com/rowland/leadtype/issues/10))
  - [x] map widget attrs to the PDF image API
  - [x] support sizing with natural aspect ratio rules
  - [x] avoid JPEG-specific assumptions in widget design so PNG can be added later without LTML API churn
- [x] Add a standard `<line>` widget
  - [x] define horizontal/vertical/general line semantics
  - [x] map border/pen attrs cleanly to the PDF layer

### Shape Widgets ([#10](https://github.com/rowland/leadtype/issues/10))

- [x] Add standard shape widgets beyond `<rect>`
  - [x] `<circle>`
  - [x] `<ellipse>`
  - [x] `<polygon>`
  - [x] `<star>`
  - [x] `<arc>`
  - [x] `<pie>`
  - [x] `<arch>`
- [x] register all new tags in the standard tag factory
- [x] ensure widget attrs map predictably to the corresponding PDF primitives
- [x] add parser and rendering tests for each tag

### Page Numbering ([#15](https://github.com/rowland/leadtype/issues/15))

- [x] Add a standard `<pageno>` widget
  - [x] support current-page output
  - [x] support embedded usage inside paragraph and label content
  - [x] support `start`, `reset`, and `hidden` page-number control attrs
- [x] add tests covering multipage documents and repeated-content-compatible page counter scenarios

### Absolute / Relative Positioning ([#14](https://github.com/rowland/leadtype/issues/14))

- [x] Make documented `absolute` and `relative` layout modes functionally real
- [x] implement widget positioning semantics for `position="absolute"` and `position="relative"`
- [x] verify `top`, `right`, `bottom`, and `left` behave consistently with existing dimensions logic
- [x] ensure width/height percentages and relative values work under positioned layouts
- [x] add rendering tests for positioned widgets inside containers and pages

### Display Modes And Overflow Flow ([#13](https://github.com/rowland/leadtype/issues/13))

- [ ] Add page-flow behavior from legacy ERML
  - [ ] `display="always"`
  - [ ] `display="first"`
  - [ ] `display="succeeding"`
  - [ ] `display="even"`
  - [ ] `display="odd"`
- [ ] support repeated header/footer style content across pages
- [ ] support overflow continuation for multi-page containers where intended
- [ ] define interaction between overflow, repeated content, and page numbering
- [ ] add tests for first-page-only, repeating, even/odd, and overflow cases

### Transforms And Placement Extras ([#14](https://github.com/rowland/leadtype/issues/14))

- [x] Add widget transform/placement attrs
  - [x] `rotate`
  - [x] `origin_x`
  - [x] `origin_y`
  - [x] `shift`
  - [x] `z_index`
- [x] define how transforms interact with layout bounds and printing order
- [x] implement graphics-state handling through the PDF writer
- [x] add tests for rotated and shifted widgets
- [x] define whether `z_index` is true stacking order or best-effort ordering and document it

### StdDocument Writer Configuration ([#18](https://github.com/rowland/leadtype/issues/18))

- [x] Add LTML document-level attributes for PDF stream compression ([#18](https://github.com/rowland/leadtype/issues/18))
  - [x] add `compress-pages`
  - [x] add `compress-to-unicode`
  - [x] add `compress-embedded-fonts`
- [x] Map those attributes from `StdDocument` into the corresponding fluent `DocWriter` methods ([#18](https://github.com/rowland/leadtype/issues/18))
  - [x] `CompressPages(bool)`
  - [x] `CompressToUnicode(bool)`
  - [x] `CompressEmbeddedFonts(bool)`
- [x] default the LTML attributes to the same compatibility-preserving `false` defaults as the PDF layer
- [x] ensure LTML document parsing does not require constructor churn in the PDF package
- [x] add LTML tests proving document attributes reach the writer and affect generated stream dictionaries ([#18](https://github.com/rowland/leadtype/issues/18))

## Cross-Cutting Integration

- [x] Keep LTML syntax docs in sync with the supported tag set and attributes
- [x] add or refresh sample documents for the newly supported widgets and behaviors
- [x] confirm new LTML features do not regress existing paragraph, table, hbox, vbox, and flow behavior

## Acceptance Checklist

- [x] `go test ./ltml` passes
- [x] any dependent package tests still pass
- [x] each new standard tag has parser coverage
- [x] each behavior that affects layout or pagination has rendering-oriented coverage
