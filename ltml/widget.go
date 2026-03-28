package ltml

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
	SetHeightPct(value float64)
	SetHeightRel(value float64)
	SetWidth(value float64)
	SetWidthPct(value float64)
	SetWidthRel(value float64)

	Height() float64
	HeightIsSet() bool
	Width() float64
	WidthPctIsSet() bool
	WidthRelIsSet() bool
	WidthIsSet() bool

	LayoutWidget(writer Writer)

	BeforePrint(writer Writer) error
	DrawBorder(writer Writer) error
	DrawContent(writer Writer) error
	PaintBackground(writer Writer) error
	Position() Position

	Align() Align
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

type DisplayMode string

const (
	DisplayOnce       DisplayMode = "once"
	DisplayAlways     DisplayMode = "always"
	DisplayFirst      DisplayMode = "first"
	DisplaySucceeding DisplayMode = "succeeding"
	DisplayEven       DisplayMode = "even"
	DisplayOdd        DisplayMode = "odd"
)

func ParseDisplayMode(value string) DisplayMode {
	switch DisplayMode(value) {
	case DisplayAlways, DisplayFirst, DisplaySucceeding, DisplayEven, DisplayOdd:
		return DisplayMode(value)
	default:
		return DisplayOnce
	}
}

func widgetDisplayForRender(widget Widget, parentRepeats bool, flowPageIndex int, physicalPageNo int) bool {
	if !parentRepeats {
		return !widget.Printed()
	}
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
		fallthrough
	default:
		return !widget.Printed()
	}
}

func ContentHeight(widget Widget) float64 {
	return widget.Height() - NonContentHeight(widget)
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
	if err := widget.BeforePrint(writer); err != nil {
		return err
	}
	render := func() error {
		if err := widget.PaintBackground(writer); err != nil {
			return err
		}
		if err := widget.DrawBorder(writer); err != nil {
			return err
		}
		if err := widget.DrawContent(writer); err != nil {
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
