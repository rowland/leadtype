package ltml

import "testing"

type transformWidget struct {
	StdWidget
	steps []string
}

func (w *transformWidget) PaintBackground(Writer) error {
	w.steps = append(w.steps, "background")
	return nil
}

func (w *transformWidget) DrawBorder(Writer) error {
	w.steps = append(w.steps, "border")
	return nil
}

func (w *transformWidget) DrawContent(Writer) error {
	w.steps = append(w.steps, "content")
	return nil
}

func TestStdWidget_ShiftOffsetsResolvedCoordinates(t *testing.T) {
	parent := positionedContainer(72, 144, 200, 200)

	w := &StdWidget{}
	if err := w.SetContainer(parent); err != nil {
		t.Fatal(err)
	}
	w.SetAttrs(map[string]string{
		"position": "relative",
		"left":     "10",
		"top":      "15",
		"right":    "-20",
		"bottom":   "-25",
		"shift":    "5,-3",
	})

	if got := w.Left(); got != 87 {
		t.Fatalf("Left() = %v, want 87", got)
	}
	if got := w.Top(); got != 156 {
		t.Fatalf("Top() = %v, want 156", got)
	}
	if got := w.Right(); got != 257 {
		t.Fatalf("Right() = %v, want 257", got)
	}
	if got := w.Bottom(); got != 316 {
		t.Fatalf("Bottom() = %v, want 316", got)
	}
}

func TestStdWidget_OriginHelpers(t *testing.T) {
	w := &StdWidget{}
	w.SetLeft(10)
	w.SetTop(20)
	w.SetWidth(30)
	w.SetHeight(40)

	if got := w.OriginX(); got != 10 {
		t.Fatalf("default OriginX() = %v, want 10", got)
	}
	if got := w.OriginY(); got != 20 {
		t.Fatalf("default OriginY() = %v, want 20", got)
	}

	w.originX = "center"
	w.originY = "middle"
	if got := w.OriginX(); got != 25 {
		t.Fatalf("center OriginX() = %v, want 25", got)
	}
	if got := w.OriginY(); got != 40 {
		t.Fatalf("middle OriginY() = %v, want 40", got)
	}

	w.originX = "right"
	w.originY = "bottom"
	if got := w.OriginX(); got != 40 {
		t.Fatalf("right OriginX() = %v, want 40", got)
	}
	if got := w.OriginY(); got != 60 {
		t.Fatalf("bottom OriginY() = %v, want 60", got)
	}
}

func TestPrint_RotateWrapsWidgetRendering(t *testing.T) {
	angle := 30.0
	widget := &transformWidget{}
	widget.SetLeft(10)
	widget.SetTop(20)
	widget.SetWidth(30)
	widget.SetHeight(40)
	widget.rotate = &angle
	widget.originX = "center"
	widget.originY = "middle"

	writer := &labelTestWriter{t: t}
	if err := Print(widget, writer); err != nil {
		t.Fatal(err)
	}

	if len(writer.rotations) != 1 {
		t.Fatalf("rotation count = %d, want 1", len(writer.rotations))
	}
	call := writer.rotations[0]
	if call.angle != 30 || call.x != 25 || call.y != 40 {
		t.Fatalf("rotation = %+v, want angle=30 x=25 y=40", call)
	}
	if got := widget.steps; len(got) != 3 || got[0] != "background" || got[1] != "border" || got[2] != "content" {
		t.Fatalf("steps = %v, want [background border content]", got)
	}
	if !widget.Printed() {
		t.Fatal("widget should be marked printed")
	}
}
