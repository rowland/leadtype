package ltml

import "testing"

type positionedTestWidget struct {
	StdWidget
	preferredWidth  float64
	preferredHeight float64
	layoutCalls     int
}

func (w *positionedTestWidget) PreferredWidth(Writer) float64 {
	return w.preferredWidth
}

func (w *positionedTestWidget) PreferredHeight(Writer) float64 {
	return w.preferredHeight
}

func (w *positionedTestWidget) LayoutWidget(Writer) {
	w.layoutCalls++
}

func positionedContainer(left, top, width, height float64) *StdContainer {
	c := &StdContainer{}
	c.SetLeft(left)
	c.SetTop(top)
	c.SetWidth(width)
	c.SetHeight(height)
	return c
}

func TestStdWidget_SetAttrs_Position(t *testing.T) {
	var absolute StdWidget
	absolute.SetAttrs(map[string]string{"position": "absolute"})
	if got := absolute.Position(); got != Absolute {
		t.Fatalf("position = %v, want %v", got, Absolute)
	}

	var autoRelative StdWidget
	autoRelative.SetAttrs(map[string]string{"left": "1in"})
	if got := autoRelative.Position(); got != Relative {
		t.Fatalf("position with left attr = %v, want %v", got, Relative)
	}

	var explicitStatic StdWidget
	explicitStatic.SetAttrs(map[string]string{"position": "static", "left": "1in"})
	if got := explicitStatic.Position(); got != Static {
		t.Fatalf("explicit static position = %v, want %v", got, Static)
	}
}

func TestStdWidget_RelativeCoordinatesAndNegativeAnchors(t *testing.T) {
	parent := positionedContainer(72, 72, 4*72, 3*72)

	child := &StdWidget{}
	if err := child.SetContainer(parent); err != nil {
		t.Fatal(err)
	}
	child.SetAttrs(map[string]string{"left": "0.5in", "top": "0.25in"})
	if got := child.Left(); got != 108 {
		t.Fatalf("relative left = %v, want 108", got)
	}
	if got := child.Top(); got != 90 {
		t.Fatalf("relative top = %v, want 90", got)
	}

	rightAnchored := &StdWidget{}
	if err := rightAnchored.SetContainer(parent); err != nil {
		t.Fatal(err)
	}
	rightAnchored.SetAttrs(map[string]string{"position": "relative", "right": "-0.5in", "width": "0.5in"})
	if got := rightAnchored.Right(); got != 324 {
		t.Fatalf("relative right = %v, want 324", got)
	}
	if got := rightAnchored.Left(); got != 288 {
		t.Fatalf("relative left from right anchor = %v, want 288", got)
	}

	sizedBySides := &StdWidget{}
	if err := sizedBySides.SetContainer(parent); err != nil {
		t.Fatal(err)
	}
	sizedBySides.SetAttrs(map[string]string{"position": "relative", "left": "0.5in", "right": "-0.5in", "top": "0.25in", "bottom": "-0.25in"})
	if got := sizedBySides.Width(); got != 216 {
		t.Fatalf("width from left/right = %v, want 216", got)
	}
	if got := sizedBySides.Height(); got != 180 {
		t.Fatalf("height from top/bottom = %v, want 180", got)
	}
}

func TestLayoutAbsolute_DefaultsToOriginAndPreferredSize(t *testing.T) {
	container := positionedContainer(72, 144, 200, 200)

	originChild := &positionedTestWidget{preferredWidth: 30, preferredHeight: 12}
	if err := originChild.SetContainer(container); err != nil {
		t.Fatal(err)
	}
	container.AddChild(originChild)

	positionedChild := &positionedTestWidget{preferredWidth: 25, preferredHeight: 10}
	if err := positionedChild.SetContainer(container); err != nil {
		t.Fatal(err)
	}
	positionedChild.SetAttrs(map[string]string{"left": "10", "top": "15"})
	container.AddChild(positionedChild)

	LayoutAbsolute(container, &LayoutStyle{}, &labelTestWriter{t: t})

	if got := originChild.Position(); got != Absolute {
		t.Fatalf("origin child position = %v, want %v", got, Absolute)
	}
	if got := originChild.Left(); got != 0 {
		t.Fatalf("origin child left = %v, want 0", got)
	}
	if got := originChild.Top(); got != 0 {
		t.Fatalf("origin child top = %v, want 0", got)
	}
	if got := originChild.Width(); got != 30 {
		t.Fatalf("origin child width = %v, want 30", got)
	}
	if got := originChild.Height(); got != 12 {
		t.Fatalf("origin child height = %v, want 12", got)
	}

	if got := positionedChild.Left(); got != 10 {
		t.Fatalf("absolute child left = %v, want 10", got)
	}
	if got := positionedChild.Top(); got != 15 {
		t.Fatalf("absolute child top = %v, want 15", got)
	}
}

func TestLayoutRelative_OffsetsFromContainer(t *testing.T) {
	container := positionedContainer(72, 144, 200, 200)

	child := &positionedTestWidget{preferredWidth: 20, preferredHeight: 8}
	if err := child.SetContainer(container); err != nil {
		t.Fatal(err)
	}
	child.SetAttrs(map[string]string{"left": "10", "top": "15", "width": "50%", "height": "50%"})
	container.AddChild(child)

	LayoutRelative(container, &LayoutStyle{}, &labelTestWriter{t: t})

	if got := child.Position(); got != Relative {
		t.Fatalf("child position = %v, want %v", got, Relative)
	}
	if got := child.Left(); got != 82 {
		t.Fatalf("relative child left = %v, want 82", got)
	}
	if got := child.Top(); got != 159 {
		t.Fatalf("relative child top = %v, want 159", got)
	}
	if got := child.Width(); got != 100 {
		t.Fatalf("relative child width = %v, want 100", got)
	}
	if got := child.Height(); got != 100 {
		t.Fatalf("relative child height = %v, want 100", got)
	}
}

func TestLayoutVBox_PositionedChildrenDoNotConsumeFlow(t *testing.T) {
	container := positionedContainer(10, 20, 200, 200)

	staticChild := &positionedTestWidget{preferredWidth: 60, preferredHeight: 24}
	if err := staticChild.SetContainer(container); err != nil {
		t.Fatal(err)
	}
	container.AddChild(staticChild)

	relativeChild := &positionedTestWidget{preferredWidth: 30, preferredHeight: 10}
	if err := relativeChild.SetContainer(container); err != nil {
		t.Fatal(err)
	}
	relativeChild.SetAttrs(map[string]string{"left": "70", "top": "5"})
	container.AddChild(relativeChild)

	LayoutVBox(container, &LayoutStyle{}, &labelTestWriter{t: t})

	if got := staticChild.Left(); got != 10 {
		t.Fatalf("static child left = %v, want 10", got)
	}
	if got := staticChild.Top(); got != 20 {
		t.Fatalf("static child top = %v, want 20", got)
	}
	if got := relativeChild.Position(); got != Relative {
		t.Fatalf("relative child position = %v, want %v", got, Relative)
	}
	if got := relativeChild.Left(); got != 80 {
		t.Fatalf("positioned child left = %v, want 80", got)
	}
	if got := relativeChild.Top(); got != 25 {
		t.Fatalf("positioned child top = %v, want 25", got)
	}
}

func TestStdPage_PaintBackground_DrawsDebugGrid(t *testing.T) {
	page := &StdPage{pageStyle: &PageStyle{width: 72, height: 72}}
	page.SetAttrs(map[string]string{"units": "in", "grid": "0.5in"})

	w := &labelTestWriter{t: t}
	if err := page.PaintBackground(w); err != nil {
		t.Fatal(err)
	}
	if len(w.moves) != 6 {
		t.Fatalf("grid move count = %d, want 6", len(w.moves))
	}
	want := [][2]float64{{0, 0}, {36, 0}, {72, 0}, {0, 0}, {0, 36}, {0, 72}}
	for i, move := range want {
		if w.moves[i] != move {
			t.Fatalf("move %d = %v, want %v", i, w.moves[i], move)
		}
	}
}
