package ltml

import (
	"errors"
	"strings"
	"testing"
)

func TestLayoutTable_ReturnsContextualLayoutError(t *testing.T) {
	container := &StdContainer{}
	container.SetPath("/ltml/page/table")
	container.SetWidth(100)
	style := defaultLayouts["table"].Clone()
	container.layout = style

	err := LayoutTable(container, style, nil)
	var layoutErr *LayoutError
	if !errors.As(err, &layoutErr) {
		t.Fatalf("LayoutTable() error = %v, want LayoutError", err)
	}
	if layoutErr.Manager != "table" || layoutErr.Path != "/ltml/page/table" {
		t.Fatalf("LayoutError = %#v, want table manager and container path", layoutErr)
	}
	if !strings.Contains(layoutErr.Error(), "cols must be specified") {
		t.Fatalf("LayoutError = %q, want missing-cols cause", layoutErr)
	}
}

func TestLayoutRadialTable_PreservesCause(t *testing.T) {
	container := &StdContainer{}
	container.SetPath("/ltml/page/radial")
	container.SetWidth(100)
	container.SetHeight(100)
	container.layout = defaultLayouts["radial"].Clone()
	sector := &StdSector{}
	if err := sector.SetContainer(container); err != nil {
		t.Fatal(err)
	}
	container.AddChild(sector)

	err := LayoutRadialTable(container, container.layout, nil)
	if !errors.Is(err, errRadialNeedsDimension) {
		t.Fatalf("LayoutRadialTable() error = %v, want errRadialNeedsDimension", err)
	}
	var layoutErr *LayoutError
	if !errors.As(err, &layoutErr) || layoutErr.Manager != "radial" || layoutErr.Path != "/ltml/page/radial" {
		t.Fatalf("LayoutRadialTable() error = %#v, want contextual LayoutError", err)
	}
}

func TestLayoutRadialTable_RejectsNonSectorChild(t *testing.T) {
	container := &StdContainer{}
	container.SetPath("/ltml/page/radial")
	container.SetWidth(100)
	container.SetHeight(100)
	container.cols = 1
	container.layout = defaultLayouts["radial"].Clone()
	child := &StdLabel{}
	if err := child.SetContainer(container); err != nil {
		t.Fatal(err)
	}
	container.AddChild(child)

	err := LayoutRadialTable(container, container.layout, nil)
	var layoutErr *LayoutError
	if !errors.As(err, &layoutErr) || !strings.Contains(layoutErr.Error(), "is not a sector") {
		t.Fatalf("LayoutRadialTable() error = %v, want non-sector LayoutError", err)
	}
}

func TestPreferredHeight_PropagatesNestedLayoutError(t *testing.T) {
	outer := &StdContainer{}
	outer.SetLeft(0)
	outer.SetTop(0)
	outer.SetWidth(200)
	outer.layout = defaultLayouts["vbox"].Clone()
	child := &StdContainer{}
	child.SetPath("/ltml/page/outer/table")
	child.layout = defaultLayouts["table"].Clone()
	if err := child.SetContainer(outer); err != nil {
		t.Fatal(err)
	}
	outer.AddChild(child)

	_, err := outer.PreferredHeight(nil)
	var layoutErr *LayoutError
	if !errors.As(err, &layoutErr) {
		t.Fatalf("PreferredHeight() error = %v, want LayoutError", err)
	}
	if layoutErr.Manager != "table" || layoutErr.Path != child.Path() {
		t.Fatalf("LayoutError = %#v, want nested table context", layoutErr)
	}
}

func TestParse_ValidatesDeterministicLayoutErrors(t *testing.T) {
	_, err := Parse([]byte(`<ltml><page><div layout="table"><label>cell</label></div></page></ltml>`))
	var layoutErr *LayoutError
	if !errors.As(err, &layoutErr) {
		t.Fatalf("Parse() error = %v, want LayoutError", err)
	}
	if layoutErr.Manager != "table" || !strings.Contains(layoutErr.Error(), "cols must be specified") {
		t.Fatalf("Parse() error = %v, want table missing-cols error", err)
	}
}

func TestParse_RejectsUnknownLayoutInsteadOfFallingBackToVBox(t *testing.T) {
	_, err := Parse([]byte(`<ltml><page><div layout="unknown"/></page></ltml>`))
	var layoutErr *LayoutError
	if !errors.As(err, &layoutErr) || layoutErr.Manager != "unknown" {
		t.Fatalf("Parse() error = %v, want unknown-manager LayoutError", err)
	}
}

func TestDocPrint_ValidatesProgrammaticLayoutChanges(t *testing.T) {
	doc, err := Parse([]byte(`<ltml><page><div/></page></ltml>`))
	if err != nil {
		t.Fatal(err)
	}
	container := doc.Root().Page(0).Widgets()[0].(*StdContainer)
	container.layout = &LayoutStyle{manager: "unknown"}

	err = doc.Print(nil)
	var layoutErr *LayoutError
	if !errors.As(err, &layoutErr) || layoutErr.Manager != "unknown" {
		t.Fatalf("Print() error = %v, want unknown-manager LayoutError", err)
	}
}

func TestLayoutFunctions_RejectNilInputsWithoutPanicking(t *testing.T) {
	err := LayoutTable(nil, defaultLayouts["table"], nil)
	var layoutErr *LayoutError
	if !errors.As(err, &layoutErr) || !strings.Contains(layoutErr.Error(), "container is nil") {
		t.Fatalf("LayoutTable(nil) error = %v, want nil-container LayoutError", err)
	}

	err = LayoutVBox(&StdContainer{}, nil, nil)
	if !errors.As(err, &layoutErr) || !strings.Contains(layoutErr.Error(), "layout style is nil") {
		t.Fatalf("LayoutVBox(nil style) error = %v, want nil-style LayoutError", err)
	}
}

func TestCloneWidgetShallow_ReturnsErrorForNilWidget(t *testing.T) {
	if _, err := cloneWidgetShallow(nil); err == nil {
		t.Fatal("cloneWidgetShallow(nil) error = nil")
	}
}
