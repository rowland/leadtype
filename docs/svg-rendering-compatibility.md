# SVG Rendering Compatibility

LeadType's SVG renderer defaults to the highest-fidelity PDF output it can
produce. For SVG fills that use varying `stop-opacity`, that means emitting a
soft-mask-based PDF representation.

Some PDF viewers handle that path better than others. Chrome and Firefox tend
to match the intended SVG appearance more closely. Preview and PDF.js can
sometimes render the same PDF with noticeably flatter or more desaturated
colors.

## Gradient Stop-Opacity Compatibility Mode

When broader viewer compatibility matters more than preserving the exact
intra-gradient transparency ramp, switch SVG gradient stop-opacity rendering to
`compatibility` mode.

What it does:

- keeps the gradient colors
- replaces varying stop alpha with a single flat object opacity
- avoids the PDF soft-mask path for that SVG gradient fill

Tradeoff:

- colors usually survive better in Preview and PDF.js
- subtle transparency variation inside the gradient is flattened

## Go API

Document-wide:

```go
doc := pdf.NewDocWriter()
doc.SetSVGGradientStopOpacityMode(pdf.SVGGradientStopOpacityModeCompatibility)
```

Page-specific:

```go
page := doc.NewPage()
page.SetSVGGradientStopOpacityMode(pdf.SVGGradientStopOpacityModeCompatibility)
```

Accepted LTML values:

- `soft-mask` (default)
- `compatibility`

## LTML

Use the root document attribute:

```xml
<ltml svg-gradient-stop-opacity-mode="compatibility">
  <page>
    <image src="hero.svg" width="4in" />
  </page>
</ltml>
```

## When To Use It

Use `compatibility` when:

- the SVG artwork uses varying `stop-opacity`
- the PDF must look acceptable in Preview or PDF.js
- a slightly flatter transparency result is preferable to washed-out color

Leave the default `soft-mask` mode when:

- Chrome/PDFium-style rendering is the priority
- preserving the original SVG transparency ramp matters more than cross-viewer consistency

## Blend Mode Compatibility

By default, LeadType honors SVG `mix-blend-mode` declarations (e.g.
`mix-blend-mode: hard-light`) and emits the corresponding PDF blend-mode entry.
This produces output that matches WebKit/Chrome rendering.

Some legacy SVG→PDF pipelines (notably older PDFlib releases) silently drop
`mix-blend-mode`. Designer artwork that was tuned against that flatter output
can look surprisingly bright or saturated when blend modes are honored,
because operators like `hard-light` typically amplify contrast against the
backdrop instead of muting it.

When matching legacy PDFlib-style output matters more than spec-faithful
blending, switch SVG blend-mode handling to `ignore`.

What it does:

- maps every SVG `mix-blend-mode` to "normal" (no blend mode is emitted)
- treats blended elements as plain alpha-composited overlays

Tradeoff:

- artwork that *relied* on `hard-light`, `multiply`, etc. to brighten or
  darken the backdrop will lose that effect

### Go API

Document-wide:

```go
doc := pdf.NewDocWriter()
doc.SetSVGBlendMode(pdf.SVGBlendModeIgnore)
```

Page-specific:

```go
page := doc.NewPage()
page.SetSVGBlendMode(pdf.SVGBlendModeIgnore)
```

Accepted LTML values:

- `respect` (default)
- `ignore`

### LTML

Use the root document attribute:

```xml
<ltml svg-blend-mode="ignore">
  <page>
    <image src="hero.svg" width="4in" />
  </page>
</ltml>
```

`svg-blend-mode` is independent of `svg-gradient-stop-opacity-mode` and the
two can be combined when reproducing legacy PDFlib output:

```xml
<ltml svg-gradient-stop-opacity-mode="compatibility" svg-blend-mode="ignore">
  ...
</ltml>
```
