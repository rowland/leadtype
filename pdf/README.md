# PDF Package

The `pdf` package provides Leadtype's low-level PDF drawing API: document and page writers, text and image output, path drawing, font registration, and related serialization support.

## Gradients

Leadtype supports linear (axial) and radial gradients for PDF drawing.

Use `SetFillLinearGradient` or `SetFillRadialGradient` to make a gradient the active fill color for subsequent shapes:

```go
if err := doc.SetFillLinearGradient(&pdf.LinearGradient{
	X0: 1, Y0: 1, X1: 4, Y1: 1,
	Stops: []pdf.GradientStop{
		{Position: 0, Color: colors.Red},
		{Position: 1, Color: colors.Blue},
	},
}); err != nil {
	panic(err)
}
doc.Rectangle(1, 1, 3, 1.5, false, true)
```

Use `SetLineLinearGradient` or `SetLineRadialGradient` for stroked outlines, and call `ClearFillGradient` or `ClearLineGradient` to return to solid colors.

Gradient stops must:

- start at `0`
- end at `1`
- increase strictly between stops

Gradient coordinates use the current document unit system, just like other drawing operations.

### Painting Into A Clip

Use `PaintLinearGradient` or `PaintRadialGradient` when you want to paint a shading directly into the current clipping region instead of using it as the current fill or stroke color:

```go
if err := doc.Path(func() {
	doc.MoveTo(5, 5)
	doc.LineTo(7, 5)
	doc.LineTo(7, 7)
	doc.LineTo(5, 7)
	doc.LineTo(5, 5)
	_ = doc.Clip(func() {
		_ = doc.PaintLinearGradient(&pdf.LinearGradient{
			X0: 5, Y0: 5, X1: 7, Y1: 7,
			Stops: []pdf.GradientStop{
				{Position: 0, Color: colors.White},
				{Position: 1, Color: colors.DarkBlue},
			},
		})
	})
}); err != nil {
	panic(err)
}
```

Use fill/stroke gradients when the gradient should behave like the current drawing color for later path operations. Use `Paint*Gradient` when the gradient should fill the current clip immediately.

## Pseudo-Sweep Bands

Leadtype also supports sweep-like annular bands built from explicit angular
segments and clipped linear paints.

Use `PaintSweepBand` when you want to compose a ring from solid wedges and
gradient wedges without introducing a new PDF shading type:

```go
err := doc.PaintSweepBand(&pdf.SweepBand{
	X: 6.25, Y: 8.5,
	InnerRadius: 0.58,
	OuterRadius: 1.02,
	Segments: []pdf.SweepBandSegment{
		{StartAngle: -22.5, EndAngle: 22.5, StartColor: colors.Red, EndColor: colors.Red},
		{StartAngle: 22.5, EndAngle: 67.5, StartColor: colors.Red, EndColor: colors.Yellow, Steps: 12},
	},
})
if err != nil {
	panic(err)
}
```

Each segment is clipped to its annular slice. Gradient segments use a linear
gradient whose axis follows the chord between the segment's start-angle and
end-angle points at the midpoint radius, which gives horizontal, vertical, or
diagonal blends depending on the slice. If `StartColor` and `EndColor` are the
same, the segment renders as a visually solid band while still using the same
gradient paint path.

Set `Steps` on a segment to subdivide it into that many equal-angle painted
subsegments. This makes it practical to approximate the older “many thin arches”
look without manually authoring every slice. `Steps <= 0` defaults to one
painted segment.

`PaintSweepArc` provides the same behavior through a convenience parameter list.

For a runnable end-to-end example, see [`../samples/test_021_gradients.go`](../samples/test_021_gradients.go).
