# SVG TODO

SVG v1 is treated like an image asset for placement and sizing, but it renders
through vector drawing rather than PDF image XObjects.

Non-goals for v1: no JavaScript, no interaction, no external CSS, no
animation, no filters, no gradients, no masks, no `<use>`. Internal
`<style>` blocks are supported only for a narrow class-based presentation-style
subset.

## API Surface

- [x] Add a new `svg` package that owns XML parsing, style parsing, intrinsic sizing, path parsing, warnings, and PDF-agnostic rendering structures.
- [x] Add explicit PDF APIs:
  - [x] `SVGDimensions(data []byte) (width, height int, err error)`
  - [x] `SVGDimensionsFromFile(filename string) (width, height int, err error)`
  - [x] `PrintSVG(data []byte, x, y float64, width, height *float64) (actualWidth, actualHeight float64, err error)`
  - [x] `PrintSVGFile(filename string, x, y float64, width, height *float64) (actualWidth, actualHeight float64, err error)`
- [x] Extend existing image APIs so SVG is auto-detected by `ImageDimensions*` and `PrintImage*`.
- [x] Keep LTML `<image>` unchanged so `<image src="foo.svg">` works through the updated image path.
- [x] Update LTML docs to describe JPEG, PNG, and SVG support through the same tag.

## Parsing And Scene Model

- [x] Parse root `<svg>` sizing from `width`/`height`.
- [x] If `width`/`height` are omitted, derive intrinsic size from `viewBox`.
- [x] Match current raster-image sizing semantics: intrinsic numeric width/height behave like point-sized dimensions unless explicit placement size is supplied.
- [x] Build a small scene graph for supported nodes:
  - [x] `<svg>`
  - [x] `<g>`
  - [x] `<line>`
  - [x] `<rect>`
  - [x] `<circle>`
  - [x] `<ellipse>`
  - [x] `<polyline>`
  - [x] `<polygon>`
  - [x] `<path>`
  - [x] `<text>`
  - [x] `<defs>`
  - [x] `<clipPath>`
- [x] Support presentation attributes plus inline `style="..."`.
- [x] Support basic color parsing:
  - [x] named colors
  - [x] `#rgb`
  - [x] `#rrggbb`
  - [x] `rgb(...)`
  - [x] `none`
- [x] Support length parsing for bare numbers, `px`, `pt`, `in`, `cm`, `mm`.
- [x] Support percentages only where they map directly to the current viewport; warn and skip cases that need object-bounding-box semantics.

## Geometry And Paint

- [ ] Support line/fill style properties:
  - [x] `fill`
  - [x] `stroke`
  - [x] `stroke-width`
  - [ ] `fill-opacity`
  - [ ] `stroke-opacity`
  - [ ] `opacity`
  - [x] `stroke-linecap`
  - [x] `stroke-linejoin`
  - [x] `stroke-miterlimit`
  - [x] `stroke-dasharray`
  - [x] `stroke-dashoffset`
  - [x] `fill-rule`
- [x] Implement SVG path-data parsing for:
  - [x] `M/m`
  - [x] `L/l`
  - [x] `H/h`
  - [x] `V/v`
  - [x] `C/c`
  - [x] `S/s`
  - [x] `Q/q`
  - [x] `T/t`
  - [x] `A/a`
  - [x] `Z/z`
- [x] Include elliptical arcs in scope by converting SVG arc commands to cubic Bezier segments before rendering.
- [x] Reuse existing circle/ellipse/arc math where helpful, but keep SVG arc conversion logic in the `svg` package so it remains format-owned.
- [x] Support `fill-rule="nonzero"` in the first pass.
- [ ] Add even-odd fill/clip handling if needed for realistic SVG compatibility; otherwise log a warning when `evenodd` is encountered and track it as an immediate follow-on.

## Transforms And Clipping

- [x] Support transform parsing for:
  - [x] `matrix(...)`
  - [x] `translate(...)`
  - [x] `scale(...)`
  - [x] `rotate(...)`
  - [x] `skewX(...)`
  - [x] `skewY(...)`
- [x] Add or expose the PDF-side transform primitive needed to apply arbitrary affine transforms in a scoped way.
- [x] Support simple `<clipPath>` with `clipPathUnits="userSpaceOnUse"`.
- [x] Support clip-path children that are basic shapes and paths.
- [x] Defer object-bounding-box clip units and mask behavior with warnings.

## Text

- [x] Support basic `<text>` rendering for a single text node payload.
- [x] Support `x`, `y`, `fill`, `font-family`, `font-size`, `font-style`, `font-weight`, and `text-anchor`.
- [x] Use existing PDF font lookup/fallback behavior; warn when the requested SVG family is unavailable.
- [x] Defer `<tspan>`, text-on-path, advanced baseline handling, and SVG text layout details.

## PDF Backend Work

- [ ] Define a narrow renderer-facing canvas interface over the existing PDF writer.
- [x] Ensure the canvas supports:
  - [x] manual path begin/end
  - [x] move/line/cubic curve/close
  - [x] fill/stroke/fill+stroke
  - [x] clip
  - [x] scoped transforms
  - [x] fill and stroke styling
  - [x] text output
- [x] Expose PDF line-join and miter-limit controls publicly if they are not already reachable through the current public API.
- [ ] Add any missing public helper needed for even-odd fill/clip if that becomes part of the v1 implementation.
- [x] Keep raster image caching behavior unchanged; do not force SVG into image XObject caching in v1.

## Unsupported Features Policy

- [x] Treat unsupported elements, attributes, and values as warn-and-skip, not fatal, whenever the rest of the document can still render.
- [x] Return hard errors only for malformed XML, malformed path data, unreadable assets, or unusable root sizing.
- [x] Include element/attribute context in warnings.
- [x] Send warnings to standard error by default.

## Tests

- [x] Add `svg` package unit tests for:
  - [x] root sizing
  - [x] color parsing
  - [x] length parsing
  - [x] style parsing
  - [x] transform parsing
  - [x] path parsing including elliptical arcs
  - [x] clipPath parsing
  - [x] warning collection
- [x] Add `pdf` tests for:
  - [x] SVG detection in `ImageDimensions*`
  - [x] SVG detection in `PrintImage*`
  - [x] explicit `PrintSVG*` APIs
  - [x] stream-level rendering of shapes, paths, transforms, clipping, and text
- [ ] Add `ltml` tests for:
  - [x] `<image src="logo.svg">` through asset FS
  - [x] inferred dimensions from `viewBox`
  - [x] width-only and height-only aspect-ratio inference
  - [ ] unsupported SVG content warning while rendering continues
- [x] Add at least one PDF golden for a mixed SVG asset using shapes, path, arc, clipPath, and text.

## Samples And Self-Verification

- [x] Add a sample Go program that renders one or more SVG files directly through the PDF API.
- [x] Add an LTML sample that embeds SVG through `<image>`.
- [x] Add a deliberately mixed SVG fixture that exercises supported and unsupported features together.
- [x] Rasterize generated PDFs during development as a normal verification step.
- [ ] Prefer local visual verification with:
  - [x] `/opt/homebrew/bin/pdftoppm`
  - [ ] `/opt/homebrew/bin/pdftocairo`
  - [ ] `/usr/bin/sips` as needed for quick inspection/conversion
- [x] Compare rasterized sample output frequently while implementing, not only at the end.
- [x] Use rasterized review to catch transform, winding, clipping, and stroke-style mistakes that stream-level assertions may miss.
- [x] Keep implementation moving autonomously: when the PDF writer lacks a needed primitive, treat that as the next concrete sub-project and continue until blocked by a real design decision rather than speculation.

## Acceptance Criteria

- [x] Existing JPEG and PNG behavior remains unchanged.
- [x] Existing LTML `<image>` documents continue to work unchanged.
- [x] SVG files with the supported subset render through both direct PDF APIs and LTML `<image>`.
- [x] Unsupported SVG features are reported to stderr and skipped without aborting the whole document when safe.
- [x] Verification includes `go build ./...`, focused tests for touched packages, and repeated rasterized sample review during development.
