# LTML plan: optional underline/strikeout pen settings + position overrides

## Goal

Allow LTML `FontStyle` to optionally carry:

1. A style reference for underline rendering.
2. A style reference for strikeout rendering.
3. Numeric position overrides for underline and strikeout lines.

Implementation must remain layered: **no LTML types (for example `ltml.PenStyle`) should be pushed into `rich_text` or `pdf`**.

## Current state (high level)

- LTML `FontStyle` supports `underline` / `strikeout` booleans.
- `ltml.PenStyle` already exists for line styling in LTML space.
- Decoration lines are rendered in lower layers (`rich_text`/`pdf`) using font metrics and writer state.

## Proposed public LTML behavior

### New attributes

Add optional font attributes:

- `font.underline-pen="<pen-id>"`
- `font.strikeout-pen="<pen-id>"`
- `font.underline-pos="<number><unit?>"`
- `font.strikeout-pos="<number><unit?>"`

Inline aliases on label-like elements:

- `underline-pen`, `strikeout-pen`, `underline-pos`, `strikeout-pos`

### Semantics

- `font.underline="true"` / `font.strikeout="true"` continue to control whether lines are drawn.
- If a `*-pen` attr is absent, keep existing default decoration appearance.
- If a `*-pos` attr is absent, keep existing font-metric/default placement.
- Position overrides apply only when the corresponding decoration is enabled.
- `*-pos` values in LTML should be treated as normal LTML measurements and resolved to points, not as raw font units.
- This differs intentionally from the lower-level `font.Font` override API, which may continue to expose font-metric overrides in font units.

### Backward compatibility

- Existing LTML remains unchanged unless the new attributes are set.
- Existing AFM/simple and TTF paths must continue to work.

### Units and layering decision

- LTML is an author-facing styling layer, so decoration overrides should use author-facing units (`pt`, `cm`, `in`, etc.), not raw font units.
- The lower-level `font.Font` override API can remain font-metric-oriented, because that layer already operates in font units and is not the primary authoring surface.
- These two layers should therefore use different semantics without sharing the same storage field or interpretation rule.
- Do not overload one option or field so that sometimes it means font units and sometimes it means points.
- If a point-space LTML override is present, lower layers should use it directly; otherwise they should continue to derive decoration placement/thickness from font metrics and any low-level font overrides.

## Layering strategy (key design constraint)

To preserve architecture boundaries:

1. LTML resolves `underline-pen` / `strikeout-pen` IDs to `ltml.PenStyle`.
2. LTML **copies only primitive/lower-level fields** from `PenStyle` into text/font data used by `rich_text`.
3. `rich_text` and `pdf` consume those copied fields using their own types (already-available lower-level types like `colors.Color`, numeric width, string pattern/cap, and numeric offsets), never `ltml.PenStyle`.

## Implementation plan

### Phase 1: Extend LTML `FontStyle`

1. Add optional fields on `ltml.FontStyle`:
   - `underlinePenID string`
   - `strikeoutPenID string`
   - `underlinePos *float64`
   - `strikeoutPos *float64`
2. Parse the new attrs in `SetAttrs(...)` using LTML measurement parsing, resolving them to points at LTML parse/application time.
3. Preserve existing `Clone()`, `Attrs()`, and string/debug output behavior with the new fields included.
4. Treat invalid measurements as unset rather than silently coercing them to zero.

### Phase 2: Define lower-layer decoration payload (no LTML types)

1. Introduce/extend decoration style fields in the structure that `rich_text` uses for rendering decisions (rich text piece/font payload), e.g.:
   - underline/strikeout line color
   - line width
   - dash/pattern
   - cap style
   - optional position overrides expressed in points
2. Use only lower-level package types already available in those layers.
3. Keep payload optional so nil/zero values mean “use existing defaults”.
4. Keep this payload semantically separate from `font.Font`'s own metric overrides, which remain font-unit-oriented.

### Phase 3: Thread LTML pen values into rich text construction

1. In LTML text construction path, resolve pen IDs with `PenStyleFor(...)`.
2. Copy pen fields into the lower-layer decoration payload (not as `PenStyle`, but as scalar/color values).
3. Thread `underlinePos` / `strikeoutPos` point-space overrides into the same payload.
4. Ensure inheritance and inline override precedence are explicit and tested.
5. Do not translate LTML point overrides back into font units; keep them as points once resolved.

### Phase 4: Apply payload in PDF text-decoration rendering

1. Update decoration drawing path to read optional payload fields from rich text/font structures.
2. If style payload exists, override decoration paint/stroke properties.
3. If a point-space position override exists, use it directly in page-space; else keep current metric-driven/default Y calculations.
4. Preserve existing behavior when payload fields are unset.
5. Keep the metric-derived path intact for callers that rely on font metrics or lower-level font-unit overrides.

### Phase 5: Validation and fallback

1. Unknown pen ID: ignore style override and fall back to existing defaults.
2. Invalid LTML position measurement: ignore override.
3. Keep rendering resilient (no hard failures for malformed decoration attrs).
4. Avoid introducing mixed-unit ambiguity in the same field or option name.

### Phase 6: Tests

#### LTML tests

- Parse + cloning coverage for new attrs on `FontStyle`.
- Measurement parsing coverage verifying `underline-pos` / `strikeout-pos` resolve to points.
- Rich text construction tests that confirm LTML pen attrs are converted into lower-layer payload values.
- Precedence tests for inherited style vs inline attrs.

#### Lower-layer tests (`rich_text` / `pdf`)

- Decoration rendering uses payload line style overrides when set.
- Decoration rendering uses point-space position overrides when set.
- Regression tests verifying identical output when no new attrs are provided.
- Regression tests verifying the metric-derived path still works when no LTML point override is present.

#### Full checks

- `go build ./...`
- `go test ./...`

## Documentation updates

1. Update `ltml/SYNTAX.md` with the four new font attributes.
2. Update `ltml/EXTENDING.md` to explain decoration styling flow and layering rule (LTML style IDs resolved in LTML, scalar values propagated downward).
3. Add example snippet showing custom underline/strikeout styling and point-based position overrides.

## Suggested rollout

1. Land attr parsing + payload plumbing first (no behavior change).
2. Land rendering behavior and tests next.
3. Land docs and samples last.

This keeps the change reviewable while enforcing package boundaries.

## Appendix: Alternative Architecture Without Reusing `font.Font`

The current low-level override work on `font.Font` is useful, but if the main
motivation is LTML pen styling, it may not be the cleanest long-term carrier.
This appendix outlines an alternative design where LTML decoration overrides do
not travel through `font.Font` at all.

### Motivation

- `font.Font` is fundamentally a font/metric abstraction.
- LTML pen styling is an author-facing text-decoration concern, not a font
  metric concern.
- Reusing `font.Font` for LTML decoration overrides can force awkward unit
  choices and, if LTML wants point-space values, may require duplicating
  fields/semantics between font-space and point-space representations.

### Alternative boundary

Instead of storing LTML-derived decoration overrides on `font.Font`:

1. Keep `font.Font` focused on intrinsic font metrics plus any true low-level
   font-metric overrides.
2. Introduce a dedicated lower-layer decoration payload on the rendered text
   side (`rich_text` leaf payload, or a small nested struct there).
3. Have LTML resolve pen IDs and measurements into that payload directly.
4. Have `pdf` consume that payload when drawing underline/strikeout lines.

### Shape of the payload

The payload would contain only lower-layer, already-resolved values such as:

- underline/strikeout line color
- line width in points
- cap style
- dash/pattern
- optional position override in points

This keeps the payload purely render-oriented and avoids storing LTML types or
font-unit semantics in the wrong layer.

### How it would work

#### LTML layer

- Parse `underline-pen` / `strikeout-pen`.
- Parse `underline-pos` / `strikeout-pos` as LTML measurements and resolve them
  to points.
- Copy the resolved scalar values into the decoration payload attached to the
  text/rich-text construction path.

#### `rich_text` layer

- Carry the optional decoration payload on each relevant leaf piece.
- Preserve it through cloning, splitting, merging, wrapping, and paragraph
  layout.
- Keep it independent from font metric computation.

#### `pdf` layer

- Prefer payload overrides when present.
- Otherwise fall back to the existing metric-derived decoration path.

### Pros

- Cleaner separation of concerns: font metrics stay in `font`, author styling
  stays in LTML, resolved rendering data lives with rendered text.
- No need to duplicate font-unit and point-unit meanings for the same override.
- Easier to explain: LTML resolves to points once, then lower layers just draw.
- Better fit if LTML is the primary driver for customizable decoration styling.

### Cons

- Requires more plumbing through `rich_text` and possibly more care in merge /
  split / clone behavior.
- Direct non-LTML callers who want decoration overrides may need a new API on
  `rich_text` or `pdf` rather than piggybacking on `font.Font`.
- Slightly larger rendered-text payloads.

### When this alternative is preferable

Prefer this design if most or all of the following are true:

- LTML is the main reason decoration overrides exist.
- Author-facing units should be points/measurements, not font units.
- You want to avoid mixed semantics on `font.Font`.
- You are willing to do somewhat more plumbing in exchange for a cleaner model.

### When the current `font.Font`-based approach is preferable

Prefer the current approach if most or all of the following are true:

- Direct library callers, outside LTML, need decoration overrides frequently.
- Font-unit overrides are acceptable or desirable at the API boundary.
- Minimizing new payload plumbing matters more than keeping the layers pure.

### Recommendation

If the main goal is LTML pen styling, the cleaner long-term architecture is:

- keep `font.Font` overrides for true font-metric concerns
- carry LTML-derived decoration styling and point-space position overrides in a
  dedicated lower-layer decoration payload

That said, the current `font.Font` approach may still be a practical stepping
stone if it helps land behavior incrementally. If used as a stepping stone, the
docs and code should explicitly treat it as an interim convenience rather than
the ideal long-term boundary.

## Appendix: Incremental Migration From `font.Font` Overrides To A Decoration Payload

If the current `font.Font`-based work is already underway, it does not have to
be thrown away immediately. The cleaner payload-based architecture can still be
reached incrementally.

### Migration goal

Move from:

- LTML-derived decoration overrides being stored on `font.Font`

to:

- LTML-derived decoration overrides being stored on a dedicated lower-layer
  decoration payload carried with rendered text

while keeping:

- existing metric-derived decoration behavior working throughout
- changes reviewable and low-risk
- direct non-LTML callers functional during the transition

### Proposed migration phases

### Phase A: Treat current `font.Font` support as a compatibility layer

- Keep the existing `font.Font` decoration override support working.
- Document clearly that this is acceptable for low-level callers, but is not the
  preferred long-term LTML transport.
- Avoid adding more LTML-specific semantics to `font.Font` than necessary.

This phase prevents churn while preserving already-landed behavior.

### Phase B: Introduce a dedicated lower-layer decoration payload

- Add a small optional decoration payload type in `rich_text` or immediately
  adjacent lower-layer code.
- Include only resolved render-time values:
  - line color
  - line width in points
  - cap style
  - dash/pattern
  - position override in points
- Keep the payload optional and leaf-oriented.

At this point, no behavior needs to change yet; the payload can exist in
parallel with the current `font.Font` override path.

### Phase C: Thread the payload through `rich_text`

- Ensure the payload survives:
  - `Clone()`
  - `DeepClone()`
  - `Split()`
  - merge behavior
  - wrapping / paragraph layout
- Add tests specifically for payload preservation through these operations.

This is the most important plumbing step, because once the payload reliably
travels with text fragments, LTML no longer needs `font.Font` as its carrier.

### Phase D: Have LTML populate the payload directly

- Resolve `underline-pen` / `strikeout-pen` to scalar lower-layer fields.
- Resolve `underline-pos` / `strikeout-pos` to points.
- Attach those values to the decoration payload on rich-text construction.
- Stop relying on `font.Font` as the primary LTML transport path for these
  values.

At the end of this phase, LTML uses the cleaner boundary, but the old
`font.Font` override path may still remain available for direct callers.

### Phase E: Make PDF prefer the payload over `font.Font`

- In decoration drawing:
  - first consult payload overrides
  - then consult any direct low-level `font.Font` overrides if they still
    matter
  - otherwise fall back to metric-derived values
- Add precedence tests so the intended override order is explicit.

This lets the new architecture take control without breaking the lower-level
compatibility path immediately.

### Phase F: Re-evaluate whether `font.Font` should keep all decoration overrides

Once LTML is fully on the payload path, decide case by case whether each
`font.Font` override should remain:

- Keep low-level font-unit overrides that are genuinely useful to direct callers.
- Remove or de-emphasize overrides that existed only to shuttle LTML values
  across layers.

This phase should be a deliberate API decision, not an automatic cleanup.

### Precedence during the transition

During migration, the cleanest precedence model is:

1. payload override from LTML or other higher-level text construction
2. explicit low-level `font.Font` override, if still supported
3. intrinsic font metric / existing default behavior

This keeps the author-facing path in control while preserving backward
compatibility for direct library users.

### Why this migration path is attractive

- It avoids a disruptive rewrite.
- It lets already-landed `font.Font` work keep providing value.
- It gives LTML a path to cleaner point-based semantics without waiting for a
  large refactor.
- It keeps the final architectural decision reversible until more usage is
  understood.

### Main risk to watch

The biggest implementation risk is not the PDF drawing code; it is ensuring the
new decoration payload survives all rich-text transformations correctly. If that
plumbing is incomplete, decoration styling can disappear or become inconsistent
after wrapping, splitting, or merging.

### Recommendation

If the current codebase is already partway down the `font.Font` path, the most
practical migration strategy is:

1. keep the current support working
2. add the dedicated payload
3. move LTML onto the payload
4. make PDF prefer payload values
5. only then decide what should remain on `font.Font`

That gives you a concrete route from the current implementation to the cleaner
long-term boundary without forcing an all-at-once change.
