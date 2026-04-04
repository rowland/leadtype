# LTML Curved Text Proposal

LTML curved text should follow, not lead, the PDF API. This note records the
likely markup direction so the lower-level API can evolve without painting the
markup layer into a corner.

## Proposed Direction

The cleanest long-term model is a dedicated curved-text element instead of
overloading the existing `<label>` widget with many path-specific attributes.

Preferred direction:

- Add a future `<textpath>` widget for curved text

Deferred alternatives:

- adding curve attributes directly to `<label>`
- attaching text as children of shape widgets such as `<circle>` or `<ellipse>`

## Planned Concepts

The LTML surface should map directly onto the PDF helper concepts:

- curve kind: `circle`, `ellipse`, or later `path`
- geometry:
  - circle: center plus radius
  - ellipse: center plus `rx` / `ry`
  - future path: explicit control points or path data
- `start-angle`
- direction / sweep
- horizontal anchor: `left`, `center`, `right`
- vertical anchor: `top`, `above`, `middle`, `baseline`, `below`
- readable orientation policy

## SVG Relationship

SVG `textPath` is a useful reference for naming and expectations, especially
for start offset and anchoring vocabulary, but LTML should not promise SVG
compatibility in the first version.

## Deferred Until PDF API Validation

- Exact LTML syntax
- XML parser and widget implementation
- Arbitrary path-data support
- Compatibility promises with SVG text layout behavior
