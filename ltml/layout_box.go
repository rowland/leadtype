package ltml

const layoutFitEpsilon = 0.001

// LayoutHBox performs a single-pass horizontal box layout with a few important
// twists:
//
//   - child widths are resolved in priority order so fixed commitments consume
//     space before flexible ones
//   - percent widths are scaled down proportionally when they collectively ask
//     for more than the remaining width
//   - omitted widths and width="auto" are similar in tight layouts, but in a
//     roomy hbox only auto children absorb surplus while omitted children stay
//     at preferred width
//   - left/right-aligned panels are placed at the edges first, then the
//     remaining unaligned children fill the center run
//
// The algorithm deliberately separates "resolve size" from "place children".
// Later layout managers and split logic depend on being able to understand
// exactly when a child became fixed to a concrete width or height.
func LayoutHBox(container Container, style *LayoutStyle, writer Writer) {
	containerFull := false

	// Static children participate in the box algorithm. Everything else is
	// handled later by layoutPositionedChildren and should be hidden here so the
	// static run is easy to reason about.
	static, remaining := printableWidgets(container, Static)
	for _, widget := range remaining {
		widget.SetVisible(false)
	}

	// Alignment affects placement order, not sizing. We first split the static
	// widgets into the three horizontal runs the hbox knows how to place:
	// edge-pinned left panels, edge-pinned right panels, and the ordinary
	// unaligned run between them.
	var lpanels, rpanels, unaligned []Widget
	for _, widget := range static {
		switch widget.Align() {
		case AlignLeft:
			lpanels = append(lpanels, widget)
		case AlignRight:
			rpanels = append(rpanels, widget)
		default:
			unaligned = append(unaligned, widget)
		}
	}

	// Width modes drive the allocation algorithm. We intentionally classify
	// children by "how width was authored" rather than by current resolved width,
	// because prior probe/layout passes may already have stamped temporary
	// resolved geometry onto the widget.
	//
	// The allocation priority is:
	//   1. explicitly specified widths, including relative widths and widths
	//      implied by opposing horizontal sides
	//   2. percent widths
	//   3. omitted widths and auto widths
	//
	// The last two are separated because width="auto" has special roomy-layout
	// behavior in hbox.
	var percents, specified, omitted, auto []Widget
	for _, widget := range static {
		if widget.WidthPctIsSet() {
			percents = append(percents, widget)
		} else if widgetAutoWidth(widget) {
			auto = append(auto, widget)
		} else if widgetWidthSpecified(widget) {
			specified = append(specified, widget)
		} else {
			omitted = append(omitted, widget)
		}
	}

	// widthAvail tracks the content width that remains for the still-unresolved
	// width groups. We subtract both child widths and the inter-child padding
	// that must exist between committed children.
	widthAvail := ContentWidth(container)

	// Specified widths are hard commitments: they consume space first and are not
	// renegotiated by hbox. If they alone overfill the row, later widgets are
	// disabled so they do not participate in placement or rendering.
	for _, widget := range specified {
		widthAvail -= widget.Width()
		containerFull = widthAvail < 0
		widget.SetDisabled(containerFull)
		widthAvail -= style.HPadding()
	}

	// Percent widths are resolved against whatever width remains after the hard
	// commitments above. If the requested percents over-allocate, they are scaled
	// down proportionally instead of being treated as independent hard failures.
	// If even the padding gaps cannot fit, the whole percent group is disabled.
	if widthAvail-float64(len(percents)-1)*style.HPadding() >= float64(len(percents)) {
		widthAvail -= float64(len(percents)-1) * style.HPadding()
		totalPercents := 0.0
		for _, widget := range percents {
			totalPercents += widget.Width()
		}
		ratio := widthAvail / totalPercents
		for _, widget := range percents {
			if ratio < 1.0 {
				widget.ResolveWidth(widget.Width() * ratio)
			}
			widthAvail -= widget.Width()
		}
	} else {
		containerFull = true
		for _, widget := range percents {
			widget.SetDisabled(true)
		}
	}
	widthAvail -= style.HPadding()

	// The final allocation step handles children whose width was omitted or set
	// to auto.
	//
	// In a traditional hbox, omitted widths simply split the leftover width
	// evenly. The new auto mode only changes the roomy case: when there is more
	// than enough space for everyone's preferred widths, omitted children keep
	// their preferred width and only auto children absorb the slack.
	if len(auto) > 0 {
		remaining := make([]Widget, 0, len(omitted)+len(auto))
		remaining = append(remaining, omitted...)
		remaining = append(remaining, auto...)
		paddingCost := float64(len(remaining)-1) * style.HPadding()
		preferredTotal := 0.0
		for _, widget := range remaining {
			preferredTotal += widget.PreferredWidth(writer)
		}
		if widthAvail > preferredTotal+paddingCost {
			widthAvail -= paddingCost
			for _, widget := range omitted {
				pw := widget.PreferredWidth(writer)
				widget.ResolveWidth(pw)
				widthAvail -= pw
			}
			autoWidth := widthAvail / float64(len(auto))
			for _, widget := range auto {
				widget.ResolveWidth(autoWidth)
			}
		} else if widthAvail-float64(len(remaining)-1)*style.HPadding() >= float64(len(remaining)) {
			widthAvail -= float64(len(remaining)-1) * style.HPadding()
			remainingWidth := widthAvail / float64(len(remaining))
			for _, widget := range remaining {
				widget.ResolveWidth(remainingWidth)
			}
		} else {
			containerFull = true
			for _, widget := range remaining {
				widget.SetDisabled(true)
			}
		}
		// With no auto widths present, omitted children keep the historical hbox
		// behavior: split whatever width remains evenly.
	} else if len(omitted) > 0 && widthAvail-float64(len(omitted)-1)*style.HPadding() >= float64(len(omitted)) {
		widthAvail -= float64(len(omitted)-1) * style.HPadding()
		omittedWidth := widthAvail / float64(len(omitted))
		for _, widget := range omitted {
			widget.ResolveWidth(omittedWidth)
		}
	} else if len(omitted) > 0 {
		containerFull = true
		for _, widget := range omitted {
			widget.SetDisabled(true)
		}
	}

	// HBox does not negotiate heights. Unspecified or auto heights simply fall
	// back to preferred height, then each child is vertically aligned within the
	// container's cross axis.
	for _, widget := range static {
		if widgetAutoHeight(widget) || !widgetHeightSpecified(widget) {
			widget.ResolveHeight(widget.PreferredHeight(writer))
		}
		widget.SetTop(hboxCrossAxisTop(container, widget))
	}

	// Placement is performed after every width is known. We maintain two moving
	// edges so left/right-aligned panels can claim the outer slots first and the
	// unaligned run naturally fills the space between them.
	left := ContentLeft(container)
	right := ContentRight(container)

	// RTL reverses the interpretation of the logical left/right runs while
	// keeping the same sizing decisions. The groups are still processed in a
	// stable order so authored child order remains meaningful.
	if IsRTL(container) {
		for _, widget := range lpanels {
			if widget.Disabled() {
				continue
			}
			widget.SetLeft(right - widget.Width())
			right -= (widget.Width() + style.HPadding())
		}
		for i := len(rpanels) - 1; i >= 0; i-- {
			widget := rpanels[i]
			if widget.Disabled() {
				continue
			}
			widget.SetLeft(left)
			left += (widget.Width() + style.HPadding())
		}
		for _, widget := range unaligned {
			if widget.Disabled() {
				continue
			}
			widget.SetLeft(right - widget.Width())
			right -= (widget.Width() + style.HPadding())
		}
	} else {
		for _, widget := range lpanels {
			if widget.Disabled() {
				continue
			}
			widget.SetLeft(left)
			left += (widget.Width() + style.HPadding())
		}
		for i := len(rpanels) - 1; i >= 0; i-- {
			widget := rpanels[i]
			if widget.Disabled() {
				continue
			}
			widget.SetRight(right)
			right -= (widget.Width() + style.HPadding())
		}
		for _, widget := range unaligned {
			if widget.Disabled() {
				continue
			}
			widget.SetLeft(left)
			left += (widget.Width() + style.HPadding())
		}
	}

	// An unsized hbox takes the tallest participating child as its content
	// height. Child layout runs after this so nested widgets see the final box
	// dimensions chosen above.
	if !container.HeightIsSet() {
		contentHeight := 0.0
		for _, widget := range static {
			if widget.Height() > contentHeight {
				contentHeight = widget.Height()
			}
		}
		container.ResolveHeight(contentHeight + NonContentHeight(container))
	}
	for _, widget := range static {
		if widget.Visible() && !widget.Disabled() {
			widget.LayoutWidget(writer)
		}
	}
	layoutPositionedChildren(container, writer)
}

// LayoutVBox performs vertical stacking with separate treatment for headers,
// body children, and footers.
//
// The vertical axis is more subtle than hbox because vbox must reconcile three
// concerns at once:
//
//   - natural-height containers need to discover their own height from children
//   - constrained containers must stop or split when they run out of room
//   - height="auto" can absorb only true surplus height, and only after the
//     baseline stack of specified, percent, and preferred heights has been
//     computed
//
// As in hbox, the function first resolves the child sizes it needs, then
// performs placement with explicit overflow checks.
func LayoutVBox(container Container, style *LayoutStyle, writer Writer) {
	containerFull := false

	// Only static children are part of the vertical flow. Positioned children are
	// laid out afterward in their own coordinate system.
	static, remaining := printableWidgets(container, Static)
	for _, widget := range remaining {
		widget.SetVisible(false)
	}

	// Top-aligned children are treated as headers, bottom-aligned children as
	// footers, and everything else as the body run. This lets vbox reserve footer
	// space from the bottom while still laying out the main body top-to-bottom.
	var headers, footers, unaligned []Widget
	for _, widget := range static {
		switch widget.Align() {
		case AlignTop:
			headers = append(headers, widget)
		case AlignBottom:
			footers = append(footers, widget)
		default:
			unaligned = append(unaligned, widget)
		}
	}
	rtl := IsRTL(container)

	// Width resolution in vbox is simpler than hbox: children are stacked, so
	// each child can independently take its preferred width up to the content
	// width. Paragraphs are a deliberate special case because their natural width
	// is effectively the full available measure rather than the unwrapped line
	// width of their text.
	for _, widget := range static {
		if widgetAutoWidth(widget) || !widgetWidthSpecified(widget) {
			cw := ContentWidth(container)
			pw := 0.0
			if _, ok := widget.(*StdParagraph); ok {
				pw = cw
			} else {
				pw = widget.PreferredWidth(writer)
			}
			if pw == 0 {
				pw = cw
			}
			w := min(pw, cw)
			widget.ResolveWidth(w)
		}
		widget.SetLeft(vboxCrossAxisLeft(container, widget, rtl))
	}

	// Before we place any child, compute the baseline vertical stack that this
	// fragment would occupy with no surplus distribution. We keep the per-widget
	// resolved height in a side map so we can later add auto-height surplus
	// without losing track of the original preferred/specifed result.
	resolvedHeights := make(map[Widget]float64, len(static))
	autoWidgets := make([]Widget, 0, len(static))
	baselineHeight := 0.0
	seen := 0
	for _, group := range [][]Widget{headers, unaligned, footers} {
		for _, widget := range group {
			height := widget.Height()
			if !widgetHeightSpecified(widget) {
				height = widget.PreferredHeight(writer)
				if widgetAutoHeight(widget) && !widgetZeroFootprint(widget) {
					autoWidgets = append(autoWidgets, widget)
				}
			}
			resolvedHeights[widget] = height
			if widgetZeroFootprint(widget) {
				continue
			}
			if seen > 0 {
				baselineHeight += style.VPadding()
			}
			baselineHeight += height
			seen++
		}
	}

	// For a natural-height vbox, the baseline stack determines the container's
	// own resolved height.
	//
	// For a constrained vbox, the baseline is instead used to decide whether
	// there is true surplus height that auto children should absorb. Omitted
	// heights never absorb this slack; they stay at preferred height.
	if !container.HeightIsSet() {
		container.ResolveHeight(baselineHeight + NonContentHeight(container))
	} else if len(autoWidgets) > 0 {
		if surplus := ContentHeight(container) - baselineHeight; surplus > 0 {
			extra := surplus / float64(len(autoWidgets))
			for _, widget := range autoWidgets {
				resolvedHeights[widget] += extra
			}
		}
	}

	// Commit the resolved heights for any child that was not already height-
	// specified by the author. This keeps the later placement and nested layout
	// passes working from a stable page-local height.
	for _, widget := range static {
		if !widgetHeightSpecified(widget) {
			widget.ResolveHeight(resolvedHeights[widget])
		}
	}

	// top advances downward through the content band. bottom is the maximum
	// usable bottom edge for this fragment, which matters both for overflow and
	// for footer placement.
	top := ContentTop(container)
	bottom := ContentTop(container) + MaxContentHeight(container)

	// Headers consume space from the top in source order.
	for _, widget := range headers {
		widget.SetTop(top)
		widget.LayoutWidget(writer)
		if widgetZeroFootprint(widget) {
			widget.SetVisible(widget.Top() <= bottom)
			continue
		}
		top += widget.Height() + style.VPadding()
		widget.SetVisible(widget.Bottom() <= bottom)
	}

	// Footers are placed from the bottom upward so they reserve their space
	// before the body run is checked for overflow.
	if len(footers) > 0 {
		footerBottom := bottom
		for i := len(footers) - 1; i >= 0; i-- {
			widget := footers[i]
			widget.SetBottom(footerBottom)
			widget.LayoutWidget(writer)
			if widgetZeroFootprint(widget) {
				widget.SetVisible(widget.Top() >= top)
				continue
			}
			footerBottom = widget.Top() - style.VPadding()
			widget.SetVisible(widget.Top() >= top)
		}
	}

	// The unaligned body run consumes whatever vertical band remains between the
	// headers and footers. Once a non-zero-footprint child would cross the bottom
	// edge, that child is hidden and later siblings are skipped for this fragment.
	for _, widget := range unaligned {
		widget.SetVisible(!containerFull)
		if containerFull {
			continue
		}
		widget.SetTop(top)
		widget.LayoutWidget(writer)
		if widgetZeroFootprint(widget) {
			widget.SetVisible(widget.Top() <= bottom)
			continue
		}
		top += widget.Height()
		if top > bottom+layoutFitEpsilon {
			containerFull = true
			widget.SetVisible(false)
		}
		top += style.VPadding()
	}

	// Positioned children are intentionally outside the static stacking
	// algorithm, so they are laid out after the vbox flow has settled.
	layoutPositionedChildren(container, writer)
}

func hboxCrossAxisTop(container Container, widget Widget) float64 {
	switch widget.SelfAlign() {
	case SelfAlignEnd:
		return ContentBottom(container) - widget.Height()
	case SelfAlignCenter:
		return ContentTop(container) + max(ContentHeight(container)-widget.Height(), 0)/2
	default:
		if container.Align() == AlignBottom {
			return ContentBottom(container) - widget.Height()
		}
		return ContentTop(container)
	}
}

func widgetWidthSpecified(widget Widget) bool {
	if widget.LeftIsSet() && widget.RightIsSet() {
		return true
	}
	switch widget.WidthMode() {
	case DimLiteral, DimPct, DimRel:
		return true
	default:
		return false
	}
}

func widgetHeightSpecified(widget Widget) bool {
	if widget.TopIsSet() && widget.BottomIsSet() {
		return true
	}
	switch widget.HeightMode() {
	case DimLiteral, DimPct, DimRel:
		return true
	default:
		return false
	}
}

func widgetAutoWidth(widget Widget) bool {
	return widget.WidthMode() == DimAuto && !(widget.LeftIsSet() && widget.RightIsSet())
}

func widgetAutoHeight(widget Widget) bool {
	return widget.HeightMode() == DimAuto && !(widget.TopIsSet() && widget.BottomIsSet())
}

func vboxCrossAxisLeft(container Container, widget Widget, rtl bool) float64 {
	switch widget.SelfAlign() {
	case SelfAlignCenter:
		return ContentLeft(container) + max(ContentWidth(container)-widget.Width(), 0)/2
	case SelfAlignEnd:
		if rtl {
			return ContentLeft(container)
		}
		return ContentRight(container) - widget.Width()
	default:
		if rtl {
			return ContentRight(container) - widget.Width()
		}
		return ContentLeft(container)
	}
}
