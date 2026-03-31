# Curved Text Roadmap

This note captures the staged plan for curved text support in Leadtype's PDF
API. The implementation is intentionally PDF-first. LTML syntax is planned here
but deferred until the lower-level API has stabilized.

## Terminology

- `path distance`: Distance measured along the path from its configured start.
- `tangent`: The unit direction vector of the path at a sampled point.
- `normal`: The unit vector perpendicular to the tangent, used to offset the
  text baseline away from the path.
- `inside` / `outside`: The side of a closed curve pointed to by the chosen
  normal policy.
- `start angle`: The angular position where path measurement begins for circle
  and ellipse helpers.
- `sweep`: The direction and extent in which text advances from the start
  angle.
- `horizontal anchor`: How the text span is aligned along the path:
  `left`, `center`, `right`.
- `vertical anchor`: How the text baseline is offset relative to the path:
  `top`, `above`, `middle`, `baseline`, `below`.
- `readable orientation`: A policy that may flip glyph orientation to keep text
  comfortable to read in common business-graphics layouts.

## API Family

The public API should grow in layers:

1. Stable circle helper on `DocWriter` and `PageWriter`
2. Stable ellipse helper on `DocWriter` and `PageWriter`
3. Deferred, more general path-following API for experimental or internal use

The first stable surface is a pair of dedicated helpers rather than a single
generic path API. Simple cases should not require callers to construct a path
object or reason about Bezier control points.

Expected public additions:

- `DrawTextOnCircle(...)`
- `DrawTextOnEllipse(...)`

Expected deferred surface:

- `DrawTextOnPath(...)`

## Default Semantics

- The `origin` for circle and ellipse text is the anchor point on the curve
  after horizontal and vertical anchoring are applied.
- Horizontal anchors default to `left`.
- Vertical anchors default to `baseline`.
- Readable orientation defaults to an upright, reader-friendly mode rather than
  strict geometric orientation.
- Glyphs are rotated and positioned along the path; glyph outlines are not
  warped to match curvature.

## Delivery Sequence

1. Design and API foundations ([#42](https://github.com/rowland/leadtype/issues/42))
2. Internal glyph-on-path placement model ([#43](https://github.com/rowland/leadtype/issues/43))
3. Circle text public API ([#44](https://github.com/rowland/leadtype/issues/44))
4. Readability and orientation policy cleanup ([#41](https://github.com/rowland/leadtype/issues/41))
5. Measurable path abstraction ([#46](https://github.com/rowland/leadtype/issues/46))
6. Ellipse text support ([#47](https://github.com/rowland/leadtype/issues/47))
7. Experimental Bezier and piecewise-path support ([#45](https://github.com/rowland/leadtype/issues/45))
8. LTML syntax design after the PDF API has been validated ([#48](https://github.com/rowland/leadtype/issues/48))

## Non-Goals For V1

- No glyph warping or envelope deformation
- No LTML curved-text implementation yet
- No commitment to full SVG `textPath` semantics
- No guarantee of support for cusps, self-intersections, or highly irregular
  arbitrary paths in the first implementation
