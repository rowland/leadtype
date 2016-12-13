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

	PreferredHeight(Writer) float64
	PreferredWidth(Writer) float64

	SetTop(float64)
	SetRight(float64)
	SetBottom(float64)
	SetLeft(float64)

	SetHeight(float64)
	SetWidth(float64)

	Height() float64
	Width() float64

	LayoutWidget(Writer)
}

func ContentHeight(widget Widget) float64 {
	return widget.Height() - NonContentHeight(widget)
}

func ContentWidth(widget Widget) float64 {
	return widget.Width() - NonContentWidth(widget)
}

func ContentTop(widget Widget) float64 {
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
