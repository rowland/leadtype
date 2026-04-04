# PDF Gradients — Implementation Plan

See [development-process.md](development-process.md) for testing strategy and fixture details.

---

## Overview

PDF supports two gradient types via **shading dictionaries** (PDF spec 8.7):

- **Type 2 (Axial)**: linear gradient along a line between two points.
- **Type 3 (Radial)**: radial gradient between two circles.

Both are driven by a **function** (Type 2 exponential interpolation or Type 3 stitching) that maps a parameter `t ∈ [0,1]` to colour values.

This plan adds gradient fill and stroke support to `PageWriter`, following the same architecture as existing colour and image handling: public setters on `PageWriter`, deferred state emission via `checkSet*` methods, shared resources registered on `DocWriter`, and low-level PDF operators emitted through the specialised writer types.

### Two rendering strategies

The PDF spec offers two ways to paint a gradient:

1. **`sh` operator** — paints the shading directly into the current clipping region. Simple, but the gradient always fills the entire clip; it cannot be used as a path fill colour.
2. **Pattern colour space** — wraps a shading in a Type 2 (shading) pattern, selects the `/Pattern` colour space with `cs`/`CS`, then references the pattern via `scn`/`SCN`. The gradient becomes the fill or stroke colour for any subsequent path operation.

Strategy 2 is more flexible and integrates cleanly with the existing `checkSetFillColor` / `checkSetLineColor` state model, so it is the primary approach. Strategy 1 (`sh`) is exposed as a convenience for full-clip gradient paints.

---

## Branching

```
master
└── pdf-gradients                       ← integration branch
    ├── gradients/phase1-types          ← gradient types, shading objects
    ├── gradients/phase2-resources      ← resource dict extensions
    ├── gradients/phase3-operators      ← scn/SCN, sh, gs operators
    ├── gradients/phase4-page-writer    ← PageWriter API, state management
    ├── gradients/phase5-tests-samples  ← integration tests, samples
    └── gradients/phase6-docs           ← user documentation
```

---

## Phase 1 — Gradient types and shading objects

### New file: `pdf/gradient.go`

Define the public-facing gradient types and the internal PDF shading objects.

```go
// GradientStop defines a colour at a position along the gradient axis.
// Position is in the range [0,1].
type GradientStop struct {
    Position float64
    Color    colors.Color
}

// LinearGradient describes an axial (linear) gradient between two points
// in user-space coordinates.
type LinearGradient struct {
    X0, Y0 float64        // start point
    X1, Y1 float64        // end point
    Stops  []GradientStop // at least two stops required
}

// RadialGradient describes a radial gradient between two circles
// in user-space coordinates.
type RadialGradient struct {
    X0, Y0, R0 float64    // start circle centre and radius
    X1, Y1, R1 float64    // end circle centre and radius
    Stops      []GradientStop
}
```

Validation: `Stops` must have at least two entries, positions must be sorted in `[0,1]`. Return errors from a `validate()` method.

### Shading dictionary (internal)

A `shadingDict` wraps the shading parameters as a PDF indirect object:

```go
// shadingDict is a PDF shading dictionary (Type 2 or Type 3).
type shadingDict struct {
    dictionaryObject
}
```

Build it from `LinearGradient` or `RadialGradient`:

- `/ShadingType` 2 (axial) or 3 (radial)
- `/ColorSpace /DeviceRGB`
- `/Coords [x0 y0 x1 y1]` (axial) or `[x0 y0 r0 x1 y1 r1]` (radial)
- `/Function` — Type 2 exponential interpolation for two-stop gradients, Type 3 stitching function for multi-stop
- `/Extend [true true]` — extend gradient beyond endpoints

### Function objects (internal)

```go
// exponentialFunction is a PDF Type 2 (exponential interpolation) function.
type exponentialFunction struct {
    dictionaryObject
}

// stitchingFunction is a PDF Type 3 (stitching) function combining
// multiple sub-functions across domains.
type stitchingFunction struct {
    dictionaryObject
}
```

For a two-stop gradient, a single Type 2 function with `N=1` maps `t` linearly from `C0` to `C1`. For multi-stop, a Type 3 stitching function chains one Type 2 function per adjacent stop pair.

### Checklist

- [ ] Define `GradientStop`, `LinearGradient`, `RadialGradient` in `pdf/gradient.go`
- [ ] Add `validate()` methods that return descriptive errors
- [ ] Implement `exponentialFunction` (Type 2 PDF function)
- [ ] Implement `stitchingFunction` (Type 3 PDF function) for multi-stop gradients
- [ ] Implement `shadingDict` construction for axial (Type 2) and radial (Type 3)
- [ ] Unit tests in `pdf/gradient_test.go`: serialise each object type and compare expected PDF syntax

---

## Phase 2 — Resource dictionary extensions

### Modify: `pdf/objects.go`

Extend `resources` to hold shading, pattern, and extended graphics state dictionaries, following the same pattern as `fonts` and `xObjects`:

```go
type resources struct {
    dictionaryObject
    fonts     dictionary
    xObjects  dictionary
    shadings  dictionary  // new: /Shading resource entries
    patterns  dictionary  // new: /Pattern resource entries
    extGStates dictionary // new: /ExtGState resource entries
}

func (r *resources) setShading(name string, ref *indirectObjectRef) {
    if r.shadings == nil {
        r.shadings = dictionary{}
        r.dict["Shading"] = r.shadings
    }
    r.shadings[name] = ref
}

func (r *resources) setPattern(name string, ref *indirectObjectRef) {
    if r.patterns == nil {
        r.patterns = dictionary{}
        r.dict["Pattern"] = r.patterns
    }
    r.patterns[name] = ref
}

func (r *resources) setExtGState(name string, ref *indirectObjectRef) {
    if r.extGStates == nil {
        r.extGStates = dictionary{}
        r.dict["ExtGState"] = r.extGStates
    }
    r.extGStates[name] = ref
}
```

### Shading pattern (internal)

A Type 2 shading pattern wraps a shading dictionary so it can be used as a pattern colour space fill:

```go
// shadingPattern is a PDF Type 2 pattern that references a shading dictionary.
type shadingPattern struct {
    dictionaryObject
}
```

Fields: `/PatternType 2`, `/Shading <ref>`, optional `/Matrix`.

### Checklist

- [ ] Add `shadings`, `patterns`, `extGStates` fields to `resources`
- [ ] Add `setShading`, `setPattern`, `setExtGState` methods
- [ ] Implement `shadingPattern` (Type 2 pattern) in `pdf/gradient.go`
- [ ] Unit tests: verify resource dictionary serialisation includes `/Shading`, `/Pattern`, `/ExtGState` entries only when populated

---

## Phase 3 — Content stream operators

### Modify: `pdf/misc_writer.go`

Implement the `scn`/`SCN` operators (resolving the existing TODO at line 43) and the `gs` operator:

```go
// setColorFillPattern sets the fill colour to a named pattern (scn operator
// with Pattern colour space).
func (mw *miscWriter) setColorFillPattern(name string) {
    fmt.Fprintf(mw.wr, "/%s scn\n", name)
}

// setColorStrokePattern sets the stroke colour to a named pattern (SCN operator
// with Pattern colour space).
func (mw *miscWriter) setColorStrokePattern(name string) {
    fmt.Fprintf(mw.wr, "/%s SCN\n", name)
}

// setExtGState activates a named extended graphics state (gs operator).
func (mw *miscWriter) setExtGState(name string) {
    fmt.Fprintf(mw.wr, "/%s gs\n", name)
}
```

### Modify: `pdf/graph_writer.go`

Add the `sh` operator for direct shading paint:

```go
// paintShading paints the named shading, filling the current clipping region.
func (gw *graphWriter) paintShading(name string) {
    fmt.Fprintf(gw.wr, "/%s sh\n", name)
}
```

### Checklist

- [ ] Add `setColorFillPattern` and `setColorStrokePattern` to `miscWriter`
- [ ] Add `setExtGState` to `miscWriter`
- [ ] Add `paintShading` to `graphWriter`
- [ ] Update TODO comment at `misc_writer.go:43` (remove or mark done)
- [ ] Unit tests in `pdf/misc_writer_test.go` and `pdf/graph_writer_test.go` for each new operator

---

## Phase 4 — PageWriter integration

### Modify: `pdf/draw_state.go`

```go
type drawState struct {
    // ... existing fields ...
    fillGradient *gradientState // new: active fill gradient (nil = solid colour)
    lineGradient *gradientState // new: active stroke gradient (nil = solid colour)
}
```

Where `gradientState` holds the resolved pattern name so `checkSet*` can compare by identity:

```go
// gradientState tracks a resolved gradient resource for state comparison.
type gradientState struct {
    patternName string
}
```

### Modify: `pdf/page_writer.go`

#### Public API

```go
// SetFillGradient sets the fill to a linear gradient. Pass nil to revert to
// solid colour fill. Coordinates are in the current unit system.
func (pw *PageWriter) SetFillLinearGradient(g *LinearGradient) error

// SetFillRadialGradient sets the fill to a radial gradient.
func (pw *PageWriter) SetFillRadialGradient(g *RadialGradient) error

// SetLineLinearGradient sets the stroke to a linear gradient.
func (pw *PageWriter) SetLineLinearGradient(g *LinearGradient) error

// SetLineRadialGradient sets the stroke to a radial gradient.
func (pw *PageWriter) SetLineRadialGradient(g *RadialGradient) error

// ClearFillGradient reverts to solid colour fill.
func (pw *PageWriter) ClearFillGradient()

// ClearLineGradient reverts to solid colour stroke.
func (pw *PageWriter) ClearLineGradient()

// PaintLinearGradient paints a linear gradient directly into the current
// clipping region using the sh operator. This is a convenience for full-clip
// gradient fills without constructing a path.
func (pw *PageWriter) PaintLinearGradient(g *LinearGradient) error

// PaintRadialGradient paints a radial gradient into the current clipping region.
func (pw *PageWriter) PaintRadialGradient(g *RadialGradient) error
```

#### State management

Add `checkSetFillGradient()` and `checkSetLineGradient()` following the `checkSetFillColor` / `checkSetLineColor` pattern:

```go
func (pw *PageWriter) checkSetFillGradient() {
    // If gradient is active and differs from last, emit:
    //   /Pattern cs      (set fill colour space)
    //   /P<n> scn        (select gradient pattern)
    // If gradient was cleared, revert to DeviceRGB:
    //   /DeviceRGB cs
    //   r g b rg
}
```

Integrate the gradient check into the existing call sites that invoke `checkSetFillColor` — the gradient takes precedence when set.

#### Gradient registration on DocWriter

```go
// Modify: pdf/doc_writer.go

type DocWriter struct {
    // ... existing fields ...
    gradients map[string]*cachedGradient // cache by content hash
}

type cachedGradient struct {
    shadingName string
    patternName string
}
```

Use a content hash (SHA-1 of gradient parameters, matching the image caching pattern) to deduplicate identical gradients across pages.

Registration flow:
1. `PageWriter.SetFillLinearGradient` calls `pw.dw.registerGradient(g)`.
2. `registerGradient` checks the cache; if miss, creates the shading dict, function objects, and shading pattern, adds them to the PDF body, registers names in resources.
3. Returns the pattern name for use in the content stream.

### Modify: `pdf/doc_writer.go`

Forward gradient methods from `DocWriter` to `CurPage()`, matching the pattern of `SetFillColor`, `SetLineColor`, etc.:

```go
func (dw *DocWriter) SetFillLinearGradient(g *LinearGradient) error {
    return dw.CurPage().SetFillLinearGradient(g)
}
```

### Coordinate handling

Gradient coordinates must be converted from the caller's unit system to PDF points (the internal coordinate system), using `pw.units.toPts()`. The page-height flip (`pw.pageHeight - y`) used elsewhere for user-space to PDF-space conversion also applies to gradient endpoints.

### Checklist

- [ ] Add `fillGradient`, `lineGradient` to `drawState`
- [ ] Add `gradientState` type
- [ ] Add `SetFillLinearGradient`, `SetFillRadialGradient` to `PageWriter`
- [ ] Add `SetLineLinearGradient`, `SetLineRadialGradient` to `PageWriter`
- [ ] Add `ClearFillGradient`, `ClearLineGradient` to `PageWriter`
- [ ] Add `PaintLinearGradient`, `PaintRadialGradient` to `PageWriter`
- [ ] Implement `checkSetFillGradient` and `checkSetLineGradient`
- [ ] Integrate gradient checks into existing draw paths (`Fill`, `FillAndStroke`, `Path`, etc.)
- [ ] Add `cachedGradient` map and `registerGradient` to `DocWriter`
- [ ] Add forwarding methods to `DocWriter`
- [ ] Handle unit conversion and Y-axis flip for gradient coordinates
- [ ] Unit tests: state transitions, operator emission, gradient-to-solid revert
- [ ] Unit tests: gradient caching and deduplication in `DocWriter`

---

## Phase 5 — Tests and samples

### Unit tests (`pdf/gradient_test.go`)

Test gradient types, validation, and PDF object serialisation in isolation:

```go
func TestLinearGradient_Validate(t *testing.T) {
    // valid: two stops
    // invalid: zero stops, one stop, positions out of range, unsorted
}

func TestShadingDict_Axial_TwoStop(t *testing.T) {
    // Verify serialised PDF syntax for a simple two-colour axial shading
}

func TestShadingDict_Radial_MultiStop(t *testing.T) {
    // Verify stitching function + radial shading
}

func TestExponentialFunction_Serialisation(t *testing.T) {
    // Verify Type 2 function with C0, C1, N=1
}

func TestStitchingFunction_Serialisation(t *testing.T) {
    // Verify Type 3 stitching of multiple sub-functions
}
```

### State and operator tests (`pdf/page_writer_test.go`)

Follow existing patterns (e.g. `TestPageWriter_checkSetFillColor`):

```go
func TestPageWriter_checkSetFillGradient(t *testing.T) {
    dw := NewDocWriter()
    pw := newPageWriter(dw, options.Options{})
    // Set gradient, verify stream contains "/Pattern cs" and "/P1 scn"
}

func TestPageWriter_FillGradientRevertToSolid(t *testing.T) {
    // Set gradient, then clear and set solid colour, verify stream
}
```

### Operator tests (`pdf/misc_writer_test.go`, `pdf/graph_writer_test.go`)

```go
func TestMiscWriter_setColorFillPattern(t *testing.T) {
    var buf bytes.Buffer
    mw := newMiscWriter(&buf)
    mw.setColorFillPattern("P1")
    expectS(t, "/P1 scn\n", buf.String())
}

func TestGraphWriter_paintShading(t *testing.T) {
    var buf bytes.Buffer
    gw := newGraphWriter(&buf)
    gw.paintShading("Sh1")
    expectS(t, "/Sh1 sh\n", buf.String())
}
```

### Resource tests (`pdf/objects_test.go`)

```go
func TestResources_Shading(t *testing.T) {
    // Verify /Shading appears in serialised resource dict when set
}

func TestResources_Pattern(t *testing.T) {
    // Verify /Pattern appears in serialised resource dict when set
}

func TestResources_ExtGState(t *testing.T) {
    // Verify /ExtGState appears in serialised resource dict when set
}

func TestResources_EmptyGradient(t *testing.T) {
    // Verify /Shading, /Pattern, /ExtGState are absent when not used
}
```

### Golden file tests

Add golden files for gradient output verification:

```
pdf/testdata/golden/
    gradient-linear-two-stop.pdf
    gradient-linear-multi-stop.pdf
    gradient-radial-two-stop.pdf
    gradient-radial-multi-stop.pdf
```

### Integration sample: `samples/test_021_gradients.go`

```go
func runTest021Gradients() (string, error) {
    return writeDoc("test_021_gradients.pdf", func(doc *pdf.DocWriter) error {
        doc.NewPage()
        doc.SetUnits("in")

        // Linear gradient fill
        doc.SetFillLinearGradient(&pdf.LinearGradient{
            X0: 1, Y0: 1, X1: 4, Y1: 1,
            Stops: []pdf.GradientStop{
                {Position: 0, Color: colors.Red},
                {Position: 1, Color: colors.Blue},
            },
        })
        doc.Rectangle(1, 1, 3, 2, false, true)

        // Multi-stop linear gradient
        doc.SetFillLinearGradient(&pdf.LinearGradient{
            X0: 1, Y0: 4, X1: 4, Y1: 4,
            Stops: []pdf.GradientStop{
                {Position: 0, Color: colors.Red},
                {Position: 0.5, Color: colors.White},
                {Position: 1, Color: colors.Blue},
            },
        })
        doc.Rectangle(1, 4, 3, 2, false, true)

        // Radial gradient
        doc.SetFillRadialGradient(&pdf.RadialGradient{
            X0: 6, Y0: 2, R0: 0,
            X1: 6, Y1: 2, R1: 1.5,
            Stops: []pdf.GradientStop{
                {Position: 0, Color: colors.Yellow},
                {Position: 1, Color: colors.Green},
            },
        })
        doc.Circle(6, 2, 1.5)

        // Revert to solid fill
        doc.ClearFillGradient()
        doc.SetFillColor(colors.Gray)
        doc.Rectangle(5, 5, 2, 1, false, true)

        return nil
    })
}
```

### Checklist

- [ ] Unit tests for gradient type validation
- [ ] Unit tests for PDF object serialisation (shading, function, pattern)
- [ ] Unit tests for new `miscWriter` and `graphWriter` operators
- [ ] Unit tests for resource dictionary extensions
- [ ] State management tests in `page_writer_test.go`
- [ ] Golden file tests for linear and radial gradients
- [ ] Sample `test_021_gradients.go` demonstrating all gradient types
- [ ] Verify all existing tests pass (`go test ./...`)

---

## Phase 6 — Code comments and user documentation

### Code comments

Add comments following the existing codebase style (brief, above the declaration, godoc-compatible):

- `gradient.go`: document each exported type and its fields
- `page_writer.go`: document each new public method with usage notes
- `misc_writer.go`: document `scn`/`SCN` operators with PDF spec references
- `graph_writer.go`: document `sh` operator with PDF spec reference

### User documentation

Add a section to any existing user guide or README covering gradient usage. Minimal example:

```
## Gradients

Leadtype supports linear (axial) and radial gradient fills for shapes.

### Linear gradient

    doc.SetFillLinearGradient(&pdf.LinearGradient{
        X0: 1, Y0: 1, X1: 4, Y1: 1,
        Stops: []pdf.GradientStop{
            {Position: 0, Color: colors.Red},
            {Position: 1, Color: colors.Blue},
        },
    })
    doc.Rectangle(1, 1, 3, 2, false, true)

### Radial gradient

    doc.SetFillRadialGradient(&pdf.RadialGradient{
        X0: 3, Y0: 3, R0: 0,
        X1: 3, Y1: 3, R1: 2,
        Stops: []pdf.GradientStop{
            {Position: 0, Color: colors.White},
            {Position: 1, Color: colors.Black},
        },
    })
    doc.Circle(3, 3, 2)

### Multi-stop gradients

Supply more than two stops to create multi-colour transitions:

    Stops: []pdf.GradientStop{
        {Position: 0, Color: colors.Red},
        {Position: 0.33, Color: colors.Yellow},
        {Position: 0.66, Color: colors.Green},
        {Position: 1, Color: colors.Blue},
    }

### Clearing a gradient

    doc.ClearFillGradient()
    doc.SetFillColor(colors.Gray) // back to solid colour
```

### Checklist

- [ ] Godoc comments on all exported types and methods in `gradient.go`
- [ ] Godoc comments on new `PageWriter` methods
- [ ] Godoc comments on new `miscWriter` and `graphWriter` methods
- [ ] User documentation with examples for linear, radial, multi-stop, and clearing

---

## Dependencies and ordering

The phases have the following dependency graph:

```
Phase 1 (types, shading objects)
  └── Phase 2 (resource dict extensions)
        └── Phase 3 (content stream operators)
              └── Phase 4 (PageWriter integration)
                    └── Phase 5 (tests, samples)
                          └── Phase 6 (documentation)
```

Phases 1 and 2 can be developed and tested independently of each other, then combined. Phase 3 depends on both. Phase 4 ties everything together. Phases 5 and 6 run in parallel once Phase 4 is complete.

### External dependencies

None. The implementation uses only the standard library, consistent with the project's zero-dependency policy.

### Internal dependencies

| Dependency | Status | Notes |
|---|---|---|
| `colors.Color` type | Exists | Gradient stops use the existing `Color` type |
| `resources` struct | Exists | Extended with new dictionaries |
| `dictionaryObject` / `indirectObjectRef` | Exists | Base types for new shading/pattern objects |
| `miscWriter` / `graphWriter` | Exists | Extended with new operator methods |
| `drawState` | Exists | Extended with gradient fields |
| `PageWriter.checkSet*` pattern | Exists | Gradient checks follow the same pattern |
| `DocWriter` caching pattern | Exists | Gradient caching mirrors image caching |
| `units` conversion | Exists | Gradient coordinates pass through `toPts()` |
| `g()` number formatter | Exists | Used for PDF number formatting in operators |
| `samples/` harness (`writeDoc`) | Exists | Sample follows existing pattern |
| `expectS` / `check` test helpers | Exists | Used in all unit tests |
| `compareGolden` test helper | Exists | Used for golden file comparison |

---

## Files modified or created

| File | Action | Phase |
|---|---|---|
| `pdf/gradient.go` | Create | 1 |
| `pdf/gradient_test.go` | Create | 1, 5 |
| `pdf/objects.go` | Modify | 2 |
| `pdf/objects_test.go` | Modify | 2, 5 |
| `pdf/misc_writer.go` | Modify | 3 |
| `pdf/misc_writer_test.go` | Modify | 3, 5 |
| `pdf/graph_writer.go` | Modify | 3 |
| `pdf/graph_writer_test.go` | Modify | 3, 5 |
| `pdf/draw_state.go` | Modify | 4 |
| `pdf/page_writer.go` | Modify | 4 |
| `pdf/page_writer_test.go` | Modify | 4, 5 |
| `pdf/doc_writer.go` | Modify | 4 |
| `pdf/doc_writer_test.go` | Modify | 4, 5 |
| `samples/test_021_gradients.go` | Create | 5 |
| `pdf/testdata/golden/gradient-*.pdf` | Create | 5 |
| `docs/gradients-plan.md` | Create (this file) | 6 |

---

## Master checklist

### Phase 1 — Gradient types and shading objects
- [ ] `GradientStop`, `LinearGradient`, `RadialGradient` types
- [ ] `validate()` methods with descriptive errors
- [ ] `exponentialFunction` (PDF Type 2 function)
- [ ] `stitchingFunction` (PDF Type 3 function)
- [ ] `shadingDict` for axial and radial shadings
- [ ] Unit tests for all new types and serialisation

### Phase 2 — Resource dictionary extensions
- [ ] `shadings`, `patterns`, `extGStates` on `resources`
- [ ] `setShading`, `setPattern`, `setExtGState` methods
- [ ] `shadingPattern` (PDF Type 2 pattern)
- [ ] Unit tests for resource serialisation

### Phase 3 — Content stream operators
- [ ] `setColorFillPattern` (`scn`) on `miscWriter`
- [ ] `setColorStrokePattern` (`SCN`) on `miscWriter`
- [ ] `setExtGState` (`gs`) on `miscWriter`
- [ ] `paintShading` (`sh`) on `graphWriter`
- [ ] Remove/resolve TODO comment at `misc_writer.go:43`
- [ ] Unit tests for each new operator

### Phase 4 — PageWriter integration
- [ ] `fillGradient`, `lineGradient` on `drawState`
- [ ] `SetFillLinearGradient`, `SetFillRadialGradient` on `PageWriter`
- [ ] `SetLineLinearGradient`, `SetLineRadialGradient` on `PageWriter`
- [ ] `ClearFillGradient`, `ClearLineGradient` on `PageWriter`
- [ ] `PaintLinearGradient`, `PaintRadialGradient` on `PageWriter`
- [ ] `checkSetFillGradient`, `checkSetLineGradient`
- [ ] Integrate gradient checks into existing draw paths
- [ ] `cachedGradient` map and `registerGradient` on `DocWriter`
- [ ] Forwarding methods on `DocWriter`
- [ ] Unit conversion and Y-axis flip for gradient coordinates
- [ ] Unit tests for state transitions and caching

### Phase 5 — Tests and samples
- [ ] Golden file tests for linear and radial gradients
- [ ] Integration sample `test_021_gradients.go`
- [ ] Verify all existing tests pass (`go test ./...`)

### Phase 6 — Code comments and user documentation
- [ ] Godoc comments on all exported gradient types and methods
- [ ] Godoc comments on new operator methods
- [ ] User-facing documentation with examples
