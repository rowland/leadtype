# LTML plan: Label/Paragraph fill with image or gradient

## Current state (as of 2026-04-14)

### What works today

- `Label` and `Paragraph` both inherit `StdWidget` via `StdContainer`, and `Print(...)` always calls `PaintBackground(...)` before text drawing.
- `StdWidget.PaintBackground(...)` draws a rectangle behind the widget when `fill` is set.
- This means both `<label fill="...">` and `<p fill="...">` currently support **background fill rectangles**.

### Phase 3 status (audited 2026-04-16)

Phase 3 is now implemented for the standard widget background path:

- `StdWidget.PaintBackground(...)` routes solid, gradient, and image brushes through one shared brush-aware painter.
- Standard widgets that rely on `StdWidget.PaintBackground(...)`, including `Label` and `Paragraph`, inherit the new background behavior automatically.
- Image brushes support uniform opacity, clipped painting, and tile sizing overrides including percentage-driven tile dimensions for high-resolution source art.
- Inline style override call sites now own prefix filtering via `filterMapAttrs(...)`; style objects receive already-normalized attribute maps.
- The sample `ltml/samples/test_045_widget_brush_backgrounds.ltml` and the widget-fill tests exercise the shared path.

### What does **not** work today

- Paragraph brush geometry scoping such as `text-fill.units="paragraph|line"` is intentionally out of scope.

## Goal

Enable LTML-first authoring for `Label` and `Paragraph` to support:

1. background fills with solid color, gradient, or image;
2. text fill with solid color, gradient, or image via true glyph clipping.

## Proposed LTML authoring model

### Background fill (widget box)

Allow existing `fill` and `fill.*` attrs to describe any brush kind:

- Solid color
  - `fill="Gold"`
  - `fill.kind="solid" fill.color="Gold"`
- Gradient
  - `fill.kind="linear-gradient"`
  - `fill.x0`, `fill.y0`, `fill.x1`, `fill.y1`
  - `fill.stops="0:#3B82F6,1:#8B5CF6"`
- Image
  - `fill.kind="image"`
  - `fill.src="..."`
  - `fill.fit="cover|contain|stretch|tile"`
  - optional `fill.anchor`, `fill.repeat`, `fill.opacity`

### Text fill

Add optional attrs on `<label>` and `<p>`:

- `text-fill="..."` or `text-fill.kind="solid|linear-gradient|radial-gradient|image"`
- `text-fill.*` uses the same brush attributes as widget `fill.*`

Semantics:
- `Label`: clip the label text and paint the widget rectangle's text-fill brush through that clip.
- `Paragraph`: treat the paragraph like a borderless widget rectangle; clip each rendered text fragment and reveal the same rectangle-wide brush through the text stencil.
- No extra text-fill units or line-scope modes are needed.

### Image brush attribute semantics

`BrushImageStyle` is the LTML-side parsed representation of an image brush. It preserves author intent for later painting phases.

- `src`
  Selects the image asset or path to render into the brush.

- `fit`
  Controls how a single image instance is sized relative to the destination area.
  - `cover`: preserve aspect ratio and scale until the destination is fully covered; cropping is allowed.
  - `contain`: preserve aspect ratio and scale until the whole image fits inside the destination; empty space is allowed.
  - `stretch`: force the image to exactly match destination width and height; aspect ratio may distort.
  - `tile`: treat the image as a repeating tile instead of a single fitted placement.

- `anchor`
  Controls alignment of the image within the destination when placement leaves either cropped overflow or unused space.
  - `center`
  - `top`
  - `bottom`
  - `left`
  - `right`
  - `top-left`
  - `top-right`
  - `bottom-left`
  - `bottom-right`

  Interpretation notes:
  - with `cover`, anchor determines which portion of the oversized image remains visible after cropping;
  - with `contain`, anchor determines where the fully visible image sits inside remaining empty space;
  - with `stretch`, anchor usually has little or no effect;
  - with `tile`, anchor can act as the tile-grid origin.

- `repeat`
  Controls whether the image repeats across the destination.
  Recommended normalized values:
  - `no-repeat`
  - `repeat`
  - `repeat-x`
  - `repeat-y`

  Current behavior:
  - `fit="tile"` selects tile-sized image placement.
  - `repeat` controls which axes repeat.
  - when `fit="tile"` is used without an explicit `repeat`, rendering promotes `no-repeat` to `repeat` so tile mode behaves like a tiled brush by default.

- `opacity`
  Controls image brush transparency as a floating-point value in the range `0..1`.
  - `0`: fully transparent
  - `1`: fully opaque
  - intermediate values such as `0.25`, `0.5`, and `0.8` represent partial transparency

  Implementation note:
  - the intended semantic range is `0..1`;
  - parser behavior should remain forgiving, with later phases deciding whether to clamp invalid values or fall back to `1`.

## Implementation plan

### Phase 5: Text fill with clipping (true fill)

1. For `Label`:
   - build/print rich text via `ClipRichText(...)`, then paint gradient/image within clip.
2. For `Paragraph`:
   - layout lines first;
   - clip each rendered text fragment;
   - paint the same paragraph rectangle-wide brush through that stencil.
3. Preserve accessibility tagging semantics:
   - clipping pass should avoid duplicate semantic text artifacts.

Deliverables:
- Correct glyph-interior gradient/image fill for labels and paragraphs.
- Tests using PDF stream assertions for clipping and shading/image operators.

Status:
- Complete, using the simplest rectangle-stencil paragraph model.

### Phase 6: LTML docs and samples

1. Add/refresh LTML samples for:
   - label text gradient clip
   - paragraph image text fill using the same paragraph rectangle-wide brush
2. Document limitations and performance implications.

Status:
- Complete.

## Significant challenges and impacts

1. **Image-as-fill semantics**
   - PDF image drawing is rectangle-based; emulating true image patterns (especially tiled/rotated) may require repeated draws and careful clipping.

2. **State management complexity**
   - Mixed operations (text clipping + gradient paint + existing fill/stroke/font state) are sensitive to graphics-state leakage.

3. **Paragraph line handling tradeoffs**
   - Keeping one rectangle-wide brush behind the paragraph is simpler to explain and keeps authoring free of extra geometry parameters.

4. **Performance and file size**
   - Clip-heavy text fill (especially long paragraphs) may increase stream size and rendering cost.

5. **Accessibility interactions**
   - Need to preserve tagged-PDF behavior while introducing clip-only paint phases.
