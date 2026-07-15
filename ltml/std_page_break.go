// Copyright 2026 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

// StdPageBreak is a zero-footprint layout marker. Overflow-capable layout
// managers consume it and continue subsequent content in the next fragment.
type StdPageBreak struct {
	StdWidget
}

func (b *StdPageBreak) ZeroFootprint() bool {
	return true
}

func isPageBreak(widget Widget) bool {
	_, ok := widget.(*StdPageBreak)
	return ok
}

// widgetHasActionablePageBreak lets an enclosing continuing layout notice a
// break hidden inside a direct child container. The child may fit by height,
// but it still has to enter the normal split path so content after the marker
// can be carried into the next fragment. This does not make pgbr bubble through
// arbitrary wrappers: effective continuation only crosses propagating vboxes.
func widgetHasActionablePageBreak(widget Widget) bool {
	container, ok := widget.(Container)
	if !ok || !containerHasEffectiveContinuation(container) || container.LayoutStyle() == nil {
		return false
	}
	switch container.LayoutStyle().manager {
	case "vbox":
		static, _ := printableWidgets(container, Static)
		seenBody := false
		for i, child := range static {
			if child.Align() == AlignTop || child.Align() == AlignBottom {
				continue
			}
			if isPageBreak(child) {
				if seenBody && hasOrdinaryVBoxBodyAfter(static, i+1) {
					return true
				}
				continue
			}
			if !widgetZeroFootprint(child) && child.Display() == DisplayOnce {
				seenBody = true
			}
		}
	case "table":
		info, err := tableGridFor(container)
		if err != nil {
			return false
		}
		bodyStart, bodyEnd, err := tableBodyRange(container, info.grid.Rows())
		if err != nil {
			return false
		}
		return firstActionableTableBreak(info.breaks, bodyStart, bodyEnd) >= 0
	}
	return false
}

// Edge markers are deliberately not actionable. Requiring ordinary one-time
// body content on the far side collapses leading, consecutive, and trailing
// pgbr markers instead of manufacturing empty fragments.
func hasOrdinaryVBoxBodyAfter(widgets []Widget, start int) bool {
	for _, child := range widgets[start:] {
		if child.Align() == AlignTop || child.Align() == AlignBottom || isPageBreak(child) || widgetZeroFootprint(child) {
			continue
		}
		if child.Display() == DisplayOnce {
			return true
		}
	}
	return false
}

func init() {
	registerTag(DefaultSpace, "pgbr", func() any { return &StdPageBreak{} })
}

var _ HasAttrs = (*StdPageBreak)(nil)
var _ Printer = (*StdPageBreak)(nil)
var _ WantsContainer = (*StdPageBreak)(nil)
var _ ZeroFootprint = (*StdPageBreak)(nil)
var _ Widget = (*StdPageBreak)(nil)
