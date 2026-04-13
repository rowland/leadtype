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
