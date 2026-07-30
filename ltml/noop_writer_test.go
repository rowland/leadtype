package ltml_test

import (
	"errors"
	"testing"

	"github.com/rowland/leadtype/colors"
	"github.com/rowland/leadtype/ltml"
	"github.com/rowland/leadtype/pdf"
)

type downstreamWriter struct {
	ltml.NoopWriter
	printed []string
}

func (w *downstreamWriter) Print(text string) error {
	w.printed = append(w.printed, text)
	return nil
}

var _ ltml.Writer = (*downstreamWriter)(nil)

func TestNoopWriterDefaults(t *testing.T) {
	var w ltml.NoopWriter
	if w.FontColor() != colors.Black {
		t.Fatalf("FontColor() = %v, want black", w.FontColor())
	}
	if w.FontSize() != 12 {
		t.Fatalf("FontSize() = %v, want 12", w.FontSize())
	}
	if w.LineSpacing() != 1 {
		t.Fatalf("LineSpacing() = %v, want 1", w.LineSpacing())
	}
}

func TestNoopWriterSupportsDownstreamOverrides(t *testing.T) {
	w := &downstreamWriter{}
	if err := w.Print("hello"); err != nil {
		t.Fatal(err)
	}
	if len(w.printed) != 1 || w.printed[0] != "hello" {
		t.Fatalf("printed = %v, want [hello]", w.printed)
	}
}

func TestNoopWriterRunsCallbacks(t *testing.T) {
	var w ltml.NoopWriter
	calls := 0
	callback := func() { calls++ }

	if err := w.Clip(callback); err != nil {
		t.Fatal(err)
	}
	if err := w.Path(callback); err != nil {
		t.Fatal(err)
	}
	if err := w.Rotate(30, 1, 2, callback); err != nil {
		t.Fatal(err)
	}
	if err := w.WithAccessibilityArtifact(callback); err != nil {
		t.Fatal(err)
	}
	if err := w.WithAccessibilityTag("Figure", pdf.AccessibilityOptions{}, callback); err != nil {
		t.Fatal(err)
	}
	if calls != 5 {
		t.Fatalf("callback calls = %d, want 5", calls)
	}

	if err := w.Clip(nil); err != nil {
		t.Fatal(err)
	}
	if err := w.WithTextDirection(pdf.TextDirectionLTR, nil); err != nil {
		t.Fatal(err)
	}
}

func TestNoopWriterWithTextDirectionPropagatesError(t *testing.T) {
	var w ltml.NoopWriter
	want := errors.New("callback failed")
	if got := w.WithTextDirection(pdf.TextDirectionLTR, func() error { return want }); !errors.Is(got, want) {
		t.Fatalf("WithTextDirection() error = %v, want %v", got, want)
	}
}

func TestNoopWriterClosedShapeBounds(t *testing.T) {
	var w ltml.NoopWriter
	shape := pdf.ClosedShape{
		Kind:   pdf.ClosedShapeCircle,
		Center: pdf.Location{X: 10, Y: 20},
		Radius: 4,
	}
	got, err := w.ClosedShapeBounds(shape)
	if err != nil {
		t.Fatal(err)
	}
	want, err := shape.Bounds()
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("ClosedShapeBounds() = %#v, want %#v", got, want)
	}
}
