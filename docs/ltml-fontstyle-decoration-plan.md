# LTML plan: optional underline/strikeout PenStyles + position overrides

## Goal

Allow LTML `FontStyle` to optionally carry:

1. A `PenStyle` for underline.
2. A `PenStyle` for strikeout.
3. Numeric position overrides for underline and strikeout lines.

The implementation should preserve existing behavior when these new fields are not provided.

## Current state (high level)

- LTML `FontStyle` currently supports `underline`/`strikeout` as booleans.
- Drawing of text decorations is ultimately controlled in the PDF writer path, with default line position/thickness behavior driven by font metrics and writer logic.
- LTML already has a reusable `PenStyle` type and style lookup via `PenStyleFor(...)`.

## Proposed public LTML behavior

### New font attributes in LTML

Add optional attributes on `FontStyle` and font-prefixed widget attributes:

- `font.underline-pen="<pen-id>"`
- `font.strikeout-pen="<pen-id>"`
- `font.underline-pos="<number><unit?>"` (line position override)
- `font.strikeout-pos="<number><unit?>"` (line position override)

Alias support for direct `label` font attributes:

- `underline-pen`, `strikeout-pen`, `underline-pos`, `strikeout-pos`

### Semantics

- `underline=true` enables underline drawing; if `underline-pen` is unset, existing pen behavior remains.
- `strikeout=true` enables strikeout drawing; if `strikeout-pen` is unset, existing pen behavior remains.
- Position overrides are only applied for the corresponding enabled decoration.
- Position values are offsets relative to the text baseline in the same coordinate convention used by existing decoration calculations.
- Unset position override means “use existing font-metric-driven/default placement.”

### Backward compatibility

- Existing LTML and rich text behavior remains unchanged unless new attrs are present.
- Existing boolean flags continue to work exactly as before.

## Implementation plan

## Phase 1: Extend LTML style model

1. Update `ltml.FontStyle` to include optional decoration styling fields:
   - `underlinePen *PenStyle`
   - `strikeoutPen *PenStyle`
   - `underlinePos *float64`
   - `strikeoutPos *float64`
2. Extend `SetAttrs(...)` parsing to read the four new attributes.
3. Resolve pen IDs via `PenStyleFor(...)` in current scope.
4. Ensure `Clone()`, `Attrs()`, and `String()` include the new fields.

## Phase 2: Writer contract for decoration overrides

1. Extend LTML/PDF writer interfaces with optional setters for decoration style/position, for example:
   - `SetUnderlinePenStyle(*PenStyle) (prev *PenStyle)`
   - `SetStrikeoutPenStyle(*PenStyle) (prev *PenStyle)`
   - `SetUnderlinePosition(*float64) (prev *float64)`
   - `SetStrikeoutPosition(*float64) (prev *float64)`
2. Propagate these through:
   - `ltml.Writer` interface
   - `layout_probe_writer`
   - `pdf.DocWriter`
   - `pdf.PageWriter`
   - test doubles used in `ltml/std_label_test.go` and related tests.
3. In `FontStyle.Apply(w)`, apply booleans first, then optional pen/position overrides.

## Phase 3: PDF drawing behavior

1. Add `PageWriter` draw-state fields for optional decoration pen + position overrides.
2. During text decoration drawing:
   - choose line color/width/pattern/cap from override pen when set;
   - otherwise keep current defaults.
3. Use explicit override position when set; otherwise keep current font-metric/default calculation.
4. Keep baseline alignment behavior intact with top/middle/bottom text alignment tests.

## Phase 4: Validation and fallbacks

1. Invalid pen ID: fallback to current default behavior (no hard error).
2. Invalid numeric position: ignore attribute and preserve default placement.
3. Document precedence rules:
   - inline font attrs override inherited style values;
   - unset override values do not clear defaults unless explicitly intended.

## Phase 5: Tests

### LTML unit tests

- `ltml/font_style_test.go` (or nearest existing test file):
  - parse/clone/attrs/string coverage for new fields.
- `ltml/std_label_test.go`:
  - verify new attrs flow into writer state for each rich text segment.
  - verify inheritance/override behavior when style + inline attrs are mixed.

### PDF tests

- `pdf/page_writer_test.go`:
  - underline with custom pen style writes expected line width/pattern/color ops.
  - strikeout with custom pen style writes expected line ops.
  - underline/strikeout position override changes Y coordinate deterministically.
  - no regressions when overrides are nil.

### Integration smoke tests

- Add/update LTML sample snippet with `font.underline-pen`, `font.strikeout-pen`, and position overrides.
- Run full build/test (`go build ./...`, `go test ./...`).

## Documentation updates

1. Update `ltml/SYNTAX.md` font attribute tables with the four new attrs and examples.
2. Update `ltml/EXTENDING.md` to describe how `PenStyle` can be reused for text decorations.
3. Add a short migration note: feature is additive and optional.

## Suggested rollout

1. Land API + state plumbing with no behavior change first (safe refactor).
2. Land drawing behavior + unit tests next.
3. Land syntax/docs/sample updates last.

This keeps review scope manageable and makes regressions easier to isolate.
