package ltml

import "strings"

type Widget interface {
	Printer

	Top() float64
	Right() float64
	Bottom() float64
	Left() float64

	MarginTop() float64
	MarginRight() float64
	MarginBottom() float64
	MarginLeft() float64

	PaddingTop() float64
	PaddingRight() float64
	PaddingBottom() float64
	PaddingLeft() float64

	PreferredHeight(writer Writer) float64
	PreferredWidth(writer Writer) float64

	SetTop(value float64)
	SetRight(value float64)
	SetBottom(value float64)
	SetLeft(value float64)
	SetPosition(value Position)

	TopIsSet() bool
	RightIsSet() bool
	BottomIsSet() bool
	LeftIsSet() bool

	SetHeight(value float64)
	SetHeightAuto()
	SetHeightPct(value float64)
	SetHeightRel(value float64)
	ResolveHeight(value float64)
	ClearResolvedHeight()
	SetWidth(value float64)
	SetWidthAuto()
	SetWidthPct(value float64)
	SetWidthRel(value float64)
	ResolveWidth(value float64)
	ClearResolvedWidth()
	SetMaxHeight(value float64)
	SetMaxWidth(value float64)
	ClearMaxHeight()
	ClearMaxWidth()

	Height() float64
	HeightIsSet() bool
	HeightMode() DimensionMode
	MaxHeight() float64
	MaxHeightIsSet() bool
	MaxWidth() float64
	MaxWidthIsSet() bool
	Width() float64
	WidthMode() DimensionMode
	WidthPctIsSet() bool
	WidthRelIsSet() bool
	WidthIsSet() bool

	LayoutWidget(writer Writer)

	BeforePrint(writer Writer) error
	DrawBorder(writer Writer) error
	DrawContent(writer Writer) error
	PaintBackground(writer Writer) error
	Position() Position
	OriginX() OriginX
	OriginY() OriginY
	OriginXValue() float64
	OriginYValue() float64

	Align() Align
	SelfAlign() SelfAlign
	Disabled() bool
	Printed() bool
	Visible() bool
	SetDisabled(value bool)
	SetPrinted(value bool)
	SetVisible(value bool)
	Display() DisplayMode
	ZIndex() int

	ColSpan() int
	RowSpan() int

	Path() string
}

type DisplayMode int8

const (
	DisplayOnce DisplayMode = iota
	DisplayAlways
	DisplayFirst
	DisplaySucceeding
	DisplayEven
	DisplayOdd
)

type ZeroFootprint interface {
	ZeroFootprint() bool
}

func ParseDisplayMode(value string) DisplayMode {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "always":
		return DisplayAlways
	case "first":
		return DisplayFirst
	case "succeeding":
		return DisplaySucceeding
	case "even":
		return DisplayEven
	case "odd":
		return DisplayOdd
	default:
		return DisplayOnce
	}
}

func (d DisplayMode) String() string {
	switch d {
	case DisplayAlways:
		return "always"
	case DisplayFirst:
		return "first"
	case DisplaySucceeding:
		return "succeeding"
	case DisplayEven:
		return "even"
	case DisplayOdd:
		return "odd"
	case DisplayOnce:
		fallthrough
	default:
		return "once"
	}
}

func widgetDisplayForRender(widget Widget, parentRepeats bool, flowPageIndex int, physicalPageNo int) bool {
	switch widget.Display() {
	case DisplayAlways:
		return true
	case DisplayFirst:
		return flowPageIndex == 1
	case DisplaySucceeding:
		return flowPageIndex > 1
	case DisplayEven:
		return physicalPageNo > 0 && physicalPageNo%2 == 0
	case DisplayOdd:
		return physicalPageNo%2 == 1
	case DisplayOnce:
		if parentRepeats && !widgetDisplayExplicit(widget) {
			return true
		}
		return !widget.Printed()
	default:
		return !widget.Printed()
	}
}

func widgetDisplayExplicit(widget Widget) bool {
	explicit, ok := widget.(interface{ DisplayExplicit() bool })
	return ok && explicit.DisplayExplicit()
}

func ContentHeight(widget Widget) float64 {
	return widget.Height() - NonContentHeight(widget)
}

func widgetZeroFootprint(widget Widget) bool {
	if zero, ok := widget.(ZeroFootprint); ok {
		return zero.ZeroFootprint()
	}
	return false
}

func ContentWidth(widget Widget) float64 {
	return widget.Width() - NonContentWidth(widget)
}

func ContentTop(widget Widget) float64 {
	if widget == nil {
		panic("ouch")
	}
	return widget.Top() + widget.MarginTop() + widget.PaddingTop()
}

func ContentRight(widget Widget) float64 {
	return widget.Right() - widget.MarginRight() - widget.PaddingRight()
}

func ContentBottom(widget Widget) float64 {
	return widget.Bottom() - widget.MarginBottom() - widget.PaddingBottom()
}

func ContentLeft(widget Widget) float64 {
	return widget.Left() + widget.MarginLeft() + widget.PaddingLeft()
}

func NonContentHeight(widget Widget) float64 {
	return widget.MarginTop() + widget.PaddingTop() + widget.PaddingBottom() + widget.MarginBottom()
}

func NonContentWidth(widget Widget) float64 {
	return widget.MarginLeft() + widget.PaddingLeft() + widget.PaddingRight() + widget.MarginRight()
}

func Print(widget Widget, writer Writer) error {
	profiler := profilerForWidget(writer, widget)
	if profiler != nil {
		defer beginWidgetProfileSpan(profiler, "print", widget).End()
	}
	{
		span := beginWidgetProfileSpan(profiler, "before", widget)
		err := widget.BeforePrint(writer)
		span.End()
		if err != nil {
			return err
		}
	}
	if err := registerPrintedWidgetMetadata(widget, writer); err != nil {
		return err
	}
	if shouldSkipRenderForPreflight(widget) {
		widget.SetPrinted(true)
		return nil
	}
	render := func() error {
		if err := withAccessibilityArtifact(writer, func() error {
			span := beginWidgetProfileSpan(profiler, "background", widget)
			err := widget.PaintBackground(writer)
			span.End()
			return err
		}); err != nil {
			return err
		}
		{
			span := beginWidgetProfileSpan(profiler, "content", widget)
			err := widget.DrawContent(writer)
			span.End()
			if err != nil {
				return err
			}
		}
		if err := withAccessibilityArtifact(writer, func() error {
			span := beginWidgetProfileSpan(profiler, "border", widget)
			err := widget.DrawBorder(writer)
			span.End()
			return err
		}); err != nil {
			return err
		}
		return nil
	}
	if tw, ok := widget.(interface {
		paintWithTransform(Writer, func() error) error
	}); ok {
		if err := tw.paintWithTransform(writer, render); err != nil {
			return err
		}
	} else if err := render(); err != nil {
		return err
	}
	widget.SetPrinted(true)
	return nil
}

func shouldSkipRenderForPreflight(widget Widget) bool {
	containerWidget, ok := widget.(interface{ Container() Container })
	if !ok {
		return false
	}
	doc := documentForContainer(containerWidget.Container())
	if doc == nil || doc.renderContext == nil || !doc.renderContext.preflight {
		return false
	}
	switch widget.(type) {
	case *StdA, *StdDraw, *StdIndex, *StdIndexEntry, *StdLabel, *StdPageNo, *StdParagraph, *StdPre, *StdSpan, *StdTarget:
		return true
	default:
		_, isContainer := widget.(Container)
		return !isContainer
	}
}
