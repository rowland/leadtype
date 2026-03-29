# SVG TODO

SVG v1 is treated like an image asset for placement and sizing, but it renders
through vector drawing rather than PDF image XObjects.

Non-goals for v1: no JavaScript, no interaction, no external CSS, no
`<style>` blocks, no animation, no filters, no gradients, no masks, no
`<use>`.

## API Surface

- [ ] Add a new `svg` package that owns XML parsing, style parsing, intrinsic sizing, path parsing, warnings, and PDF-agnostic rendering structures.
- [ ] Add explicit PDF APIs:
  - [ ] `SVGDimensions(data []byte) (width, height int, err error)`
  - [ ] `SVGDimensionsFromFile(filename string) (width, height int, err error)`
  - [ ] `PrintSVG(data []byte, x, y float64, width, height *float64) (actualWidth, actualHeight float64, err error)`
  - [ ] `PrintSVGFile(filename string, x, y float64, width, height *float64) (actualWidth, actualHeight float64, err error)`
- [ ] Extend existing image APIs so SVG is auto-detected by `ImageDimensions*` and `PrintImage*`.
- [ ] Keep LTML `<image>` unchanged so `<image src="foo.svg">` works through the updated image path.
- [ ] Update LTML docs to describe JPEG, PNG, and SVG support through the same tag.

## Parsing And Scene Model

- [ ] Parse root `<svg>` sizing from `width`/`height`.
- [ ] If `width`/`height` are omitted, derive intrinsic size from `viewBox`.
- [ ] Match current raster-image sizing semantics: intrinsic numeric width/height behave like point-sized dimensions unless explicit placement size is supplied.
- [ ] Build a small scene graph for supported nodes:
  - [ ] `<svg>`
  - [ ] `<g>`
  - [ ] `<line>`
  - [ ] `<rect>`
  - [ ] `<circle>`
  - [ ] `<ellipse>`
  - [ ] `<polyline>`
  - [ ] `<polygon>`
  - [ ] `<path>`
  - [ ] `<text>`
  - [ ] `<defs>`
  - [ ] `<clipPath>`
- [ ] Support presentation attributes plus inline `style="..."`.
- [ ] Support basic color parsing:
  - [ ] named colors
  - [ ] `#rgb`
  - [ ] `#rrggbb`
  - [ ] `rgb(...)`
  - [ ] `none`
- [ ] Support length parsing for bare numbers, `px`, `pt`, `in`, `cm`, `mm`.
- [ ] Support percentages only where they map directly to the current viewport; warn and skip cases that need object-bounding-box semantics.

## Geometry And Paint

- [ ] Support line/fill style properties:
  - [ ] `fill`
  - [ ] `stroke`
  - [ ] `stroke-width`
  - [ ] `fill-opacity`
  - [ ] `stroke-opacity`
  - [ ] `opacity`
  - [ ] `stroke-linecap`
  - [ ] `stroke-linejoin`
  - [ ] `stroke-miterlimit`
  - [ ] `stroke-dasharray`
  - [ ] `stroke-dashoffset`
  - [ ] `fill-rule`
- [ ] Implement SVG path-data parsing for:
  - [ ] `M/m`
  - [ ] `L/l`
  - [ ] `H/h`
  - [ ] `V/v`
  - [ ] `C/c`
  - [ ] `S/s`
  - [ ] `Q/q`
  - [ ] `T/t`
  - [ ] `A/a`
  - [ ] `Z/z`
- [ ] Include elliptical arcs in scope by converting SVG arc commands to cubic Bezier segments before rendering.
- [ ] Reuse existing circle/ellipse/arc math where helpful, but keep SVG arc conversion logic in the `svg` package so it remains format-owned.
- [ ] Support `fill-rule="nonzero"` in the first pass.
- [ ] Add even-odd fill/clip handling if needed for realistic SVG compatibility; otherwise log a warning when `evenodd` is encountered and track it as an immediate follow-on.

## Transforms And Clipping

- [ ] Support transform parsing for:
  - [ ] `matrix(...)`
  - [ ] `translate(...)`
  - [ ] `scale(...)`
  - [ ] `rotate(...)`
  - [ ] `skewX(...)`
  - [ ] `skewY(...)`
- [ ] Add or expose the PDF-side transform primitive needed to apply arbitrary affine transforms in a scoped way.
- [ ] Support simple `<clipPath>` with `clipPathUnits="userSpaceOnUse"`.
- [ ] Support clip-path children that are basic shapes and paths.
- [ ] Defer object-bounding-box clip units and mask behavior with warnings.

## Text

- [ ] Support basic `<text>` rendering for a single text node payload.
- [ ] Support `x`, `y`, `fill`, `font-family`, `font-size`, `font-style`, `font-weight`, and `text-anchor`.
- [ ] Use existing PDF font lookup/fallback behavior; warn when the requested SVG family is unavailable.
- [ ] Defer `<tspan>`, text-on-path, advanced baseline handling, and SVG text layout details.

## PDF Backend Work

- [ ] Define a narrow renderer-facing canvas interface over the existing PDF writer.
- [ ] Ensure the canvas supports:
  - [ ] manual path begin/end
  - [ ] move/line/cubic curve/close
  - [ ] fill/stroke/fill+stroke
  - [ ] clip
  - [ ] scoped transforms
  - [ ] fill and stroke styling
  - [ ] text output
- [ ] Expose PDF line-join and miter-limit controls publicly if they are not already reachable through the current public API.
- [ ] Add any missing public helper needed for even-odd fill/clip if that becomes part of the v1 implementation.
- [ ] Keep raster image caching behavior unchanged; do not force SVG into image XObject caching in v1.

## Unsupported Features Policy

- [ ] Treat unsupported elements, attributes, and values as warn-and-skip, not fatal, whenever the rest of the document can still render.
- [ ] Return hard errors only for malformed XML, malformed path data, unreadable assets, or unusable root sizing.
- [ ] Include element/attribute context in warnings.
- [ ] Send warnings to standard error by default.

## Tests

- [ ] Add `svg` package unit tests for:
  - [ ] root sizing
  - [ ] color parsing
  - [ ] length parsing
  - [ ] style parsing
  - [ ] transform parsing
  - [ ] path parsing including elliptical arcs
  - [ ] clipPath parsing
  - [ ] warning collection
- [ ] Add `pdf` tests for:
  - [ ] SVG detection in `ImageDimensions*`
  - [ ] SVG detection in `PrintImage*`
  - [ ] explicit `PrintSVG*` APIs
  - [ ] stream-level rendering of shapes, paths, transforms, clipping, and text
- [ ] Add `ltml` tests for:
  - [ ] `<image src="logo.svg">` through asset FS
  - [ ] inferred dimensions from `viewBox`
  - [ ] width-only and height-only aspect-ratio inference
  - [ ] unsupported SVG content warning while rendering continues
- [ ] Add at least one PDF golden for a mixed SVG asset using shapes, path, arc, clipPath, and text.

## Samples And Self-Verification

- [ ] Add a sample Go program that renders one or more SVG files directly through the PDF API.
- [ ] Add an LTML sample that embeds SVG through `<image>`.
- [ ] Add a deliberately mixed SVG fixture that exercises supported and unsupported features together.
- [ ] Rasterize generated PDFs during development as a normal verification step.
- [ ] Prefer local visual verification with:
  - [ ] `/opt/homebrew/bin/pdftoppm`
  - [ ] `/opt/homebrew/bin/pdftocairo`
  - [ ] `/usr/bin/sips` as needed for quick inspection/conversion
- [ ] Compare rasterized sample output frequently while implementing, not only at the end.
- [ ] Use rasterized review to catch transform, winding, clipping, and stroke-style mistakes that stream-level assertions may miss.
- [ ] Keep implementation moving autonomously: when the PDF writer lacks a needed primitive, treat that as the next concrete sub-project and continue until blocked by a real design decision rather than speculation.

## Acceptance Criteria

- [ ] Existing JPEG and PNG behavior remains unchanged.
- [ ] Existing LTML `<image>` documents continue to work unchanged.
- [ ] SVG files with the supported subset render through both direct PDF APIs and LTML `<image>`.
- [ ] Unsupported SVG features are reported to stderr and skipped without aborting the whole document when safe.
- [ ] Verification includes `go build ./...`, focused tests for touched packages, and repeated rasterized sample review during development.
