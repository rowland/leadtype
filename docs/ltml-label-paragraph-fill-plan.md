# LTML plan: Label/Paragraph fill with image or gradient

## Current state (as of 2026-04-14)

### What works today

- `Label` and `Paragraph` both inherit `StdWidget` via `StdContainer`, and `Print(...)` always calls `PaintBackground(...)` before text drawing.
- `StdWidget.PaintBackground(...)` draws a rectangle behind the widget when `fill` is set.
- This means both `<label fill="...">` and `<p fill="...">` currently support **background fill rectangles**.

### What does **not** work today

- LTML `fill` resolves to `BrushStyle`, and `BrushStyle` currently stores only a single `color` and applies only `SetFillColor(...)`.
- The LTML `Writer` interface does not expose gradient APIs (`SetFillLinearGradient`, `PaintLinearGradient`, etc.), even though the PDF layer supports them.
- There is no LTML brush/image-fill API for painting image patterns inside widget bounds.
- Label/paragraph text painting uses `PrintRichText` / `PrintParagraph`; there is no LTML-level text-fill mode that applies gradient/image paint to glyph interiors.

## Goal

Enable LTML-first authoring for `Label` and `Paragraph` to support:

1. background fills with solid color (existing), gradient, or image;
2. text fill with solid color (existing), gradient, or image;
3. text-fill behavior in two modes:
   - **without clipping** (approximate, fast path),
   - **with text clipping** (true glyph-interior fill path).

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
- `text-fill.clip="auto|true|false"`
  - `false` => no clipping mode
  - `true` => force clip mode
  - `auto` => clip for gradients/images, direct font color for solid

For paragraphs, define default geometry scope for gradient/image coordinates:

- `text-fill.units="paragraph|line"`
  - `paragraph`: single brush across full paragraph block
  - `line`: re-resolve brush per line (often better visual continuity for narrow columns)

## Implementation plan

### Phase 1: Brush model and parser foundation

1. Extend `BrushStyle` to represent multiple paint kinds:
   - solid color
   - linear/radial gradient
   - image
2. Add robust attr parsing for stops/coordinates/image options.
3. Keep backward compatibility:
   - existing `fill="ColorName"` remains valid.

Deliverables:
- Updated `BrushStyle` structure and parsing.
- Unit tests for style parsing and clone behavior.

### Phase 2: Writer capabilities needed by LTML

1. Extend LTML `Writer` with gradient methods needed for widget/background use:
   - set/clear fill gradient and/or direct paint in clip.
2. Add image paint helper(s) for clipped region painting, likely one of:
   - a generic `PaintImageInClip(...)`, or
   - a pattern-like image fill abstraction.
3. Implement these in the LTML writer-backed PDF writer.

Deliverables:
- Interface and concrete writer support.
- Tests ensuring state restoration after gradient/image painting.

### Phase 3: Background fill for Label/Paragraph

1. Replace `BrushStyle.Apply(...)` + `Rectangle2(...fill=true...)` assumption with brush-aware painting path:
   - solid: current behavior.
   - gradient: build rect path, clip, paint gradient.
   - image: build rect path, clip, paint image according to fit/repeat.
2. Ensure border drawing remains unchanged and layered above background.

Deliverables:
- New background rendering behavior for all widgets that use `StdWidget.PaintBackground`.
- Regression tests for existing solid-fill outputs.

### Phase 4: Text fill without clipping

1. Add text-fill style resolution for `Label` and `Paragraph`.
2. Non-clipped rendering strategy:
   - solid: existing font color path.
   - gradient/image: fallback strategy (e.g., average color sampling or first stop color), with explicit warning in docs that this is approximate.

Deliverables:
- Fast path behavior available even where text clipping is undesirable.
- Deterministic fallback tests.

### Phase 5: Text fill with clipping (true fill)

1. For `Label`:
   - build/print rich text via `ClipRichText(...)`, then paint gradient/image within clip.
2. For `Paragraph`:
   - layout lines first; clip line-by-line or paragraph-wide depending on `text-fill.units`.
3. Preserve accessibility tagging semantics:
   - clipping pass should avoid duplicate semantic text artifacts.

Deliverables:
- Correct glyph-interior gradient/image fill for labels and paragraphs.
- Tests using PDF stream assertions for clipping and shading/image operators.

### Phase 6: LTML docs and samples

1. Add/refresh LTML samples for:
   - label gradient background
   - paragraph image background
   - label text gradient clip on/off
   - paragraph text gradient with `paragraph` vs `line` units
2. Document limitations and performance implications.

## Significant challenges and impacts

1. **Writer interface expansion impact**
   - Adding gradient/image-paint APIs to LTML `Writer` affects test doubles and helper writers across LTML tests.

2. **Image-as-fill semantics**
   - PDF image drawing is rectangle-based; emulating true image patterns (especially tiled/rotated) may require repeated draws and careful clipping.

3. **State management complexity**
   - Mixed operations (text clipping + gradient paint + existing fill/stroke/font state) are sensitive to graphics-state leakage.

4. **Paragraph line handling tradeoffs**
   - Paragraph-wide gradients look continuous but can be less readable on wrapped lines; line-scoped gradients are clearer but visually discontinuous.

5. **Performance and file size**
   - Clip-heavy text fill (especially long paragraphs) may increase stream size and rendering cost.

6. **Accessibility interactions**
   - Need to preserve tagged-PDF behavior while introducing clip-only paint phases.

## Suggested sequencing for low risk

1. Ship background gradient/image fill first (no text clipping required).
2. Ship text-fill API with solid + non-clipped fallback.
3. Ship clipped text-fill for labels.
4. Ship clipped text-fill for paragraphs.
5. Harden with sample-driven and golden PDF tests.
