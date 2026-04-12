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
doc.SetSVGGradientStopOpacityMode("compatibility")
```

Page-specific:

```go
page := doc.NewPage()
page.SetSVGGradientStopOpacityMode("compatibility")
```

Accepted values:

- `soft-mask`
- `compatibility`

Default: `soft-mask`

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
