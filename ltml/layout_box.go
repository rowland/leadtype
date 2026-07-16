package ltml

import "math"

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
func LayoutHBox(container Container, style *LayoutStyle, writer Writer) (err error) {
	defer func() { err = wrapLayoutError("hbox", containerPath(container), err) }()
	if err := validateLayoutInputs(container, style); err != nil {
		return err
	}
	containerFull := false

	// Static children participate in the box algorithm. Everything else is
	// handled later by layoutPositionedChildren and should be hidden here so the
	// static run is easy to reason about.
	static, remaining := printableWidgets(container, Static)
	for _, widget := range remaining {
		widget.SetVisible(false)
	}
	// An hbox does not propagate continuation, so pgbr is inert here. Still lay
	// the marker out at zero size so normal printed-state bookkeeping consumes
	// it; leaving it in the horizontal runs would also distort alignment counts.
	participating := static[:0]
	for _, widget := range static {
		if !isPageBreak(widget) {
			participating = append(participating, widget)
			continue
		}
		widget.SetVisible(true)
		widget.ResolveWidth(0)
		widget.ResolveHeight(0)
		widget.SetLeft(ContentLeft(container))
		widget.SetTop(ContentTop(container))
		if err := widget.LayoutWidget(writer); err != nil {
			return err
		}
	}
	static = participating

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
		containerFull = widthAvail+layoutFitEpsilon < 0
		widget.SetDisabled(containerFull)
		widthAvail -= style.HPadding()
	}

	// Percent widths are resolved against whatever width remains after the hard
	// commitments above. If the requested percents over-allocate, they are scaled
	// down proportionally instead of being treated as independent hard failures.
	// If even the padding gaps cannot fit, the whole percent group is disabled.
	if len(percents) > 0 {
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
		// A non-empty percent group reserves one trailing gap before the flexible
		// groups. Empty groups must not run the internal-gap calculation above:
		// len(percents)-1 would otherwise add one padding unit to widthAvail.
		widthAvail -= style.HPadding()
	}

	// Edge-aligned flexible panels are placed outside the ordinary center run,
	// so reserve their preferred widths before omitted/auto center children are
	// sized. Otherwise a wide center child can consume the space that a
	// right-aligned image or panel will later claim during placement.
	widthAvail, err = reserveHBoxEdgeFlexibleWidths(omitted, style, writer, widthAvail, &containerFull)
	if err != nil {
		return err
	}
	widthAvail, err = reserveHBoxEdgeFlexibleWidths(auto, style, writer, widthAvail, &containerFull)
	if err != nil {
		return err
	}
	omitted = filterHBoxCenterFlexible(omitted)
	auto = filterHBoxCenterFlexible(auto)

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
		omittedPreferredTotal := 0.0
		for _, widget := range omitted {
			width, err := widget.PreferredWidth(writer)
			if err != nil {
				return err
			}
			omittedPreferredTotal += width
		}
		autoPreferredTotal := 0.0
		for _, widget := range auto {
			width, err := widget.PreferredWidth(writer)
			if err != nil {
				return err
			}
			autoPreferredTotal += width
		}
		preferredTotal := omittedPreferredTotal + autoPreferredTotal
		if widthAvail+layoutFitEpsilon >= preferredTotal+paddingCost {
			widthAvail -= paddingCost
			for _, widget := range omitted {
				pw, err := widget.PreferredWidth(writer)
				if err != nil {
					return err
				}
				widget.ResolveWidth(pw)
				widthAvail -= pw
			}
			surplus := widthAvail - autoPreferredTotal
			autoExtra := surplus / float64(len(auto))
			for _, widget := range auto {
				width, err := widget.PreferredWidth(writer)
				if err != nil {
					return err
				}
				widget.ResolveWidth(width + autoExtra)
			}
		} else if widthAvail+layoutFitEpsilon >= paddingCost {
			widthAvail -= paddingCost
			if preferredTotal <= 0 {
				autoWidth := 0.0
				if len(auto) > 0 {
					autoWidth = widthAvail / float64(len(auto))
				}
				for _, widget := range omitted {
					widget.ResolveWidth(0)
				}
				for _, widget := range auto {
					widget.ResolveWidth(autoWidth)
				}
			} else {
				ratio := min(widthAvail/preferredTotal, 1)
				for _, widget := range omitted {
					width, err := widget.PreferredWidth(writer)
					if err != nil {
						return err
					}
					widget.ResolveWidth(width * ratio)
				}
				for _, widget := range auto {
					width, err := widget.PreferredWidth(writer)
					if err != nil {
						return err
					}
					widget.ResolveWidth(width * ratio)
				}
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
			height, err := widget.PreferredHeight(writer)
			if err != nil {
				return err
			}
			widget.ResolveHeight(height)
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
			if err := widget.LayoutWidget(writer); err != nil {
				return err
			}
		}
	}
	return layoutPositionedChildren(container, writer)
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
func LayoutVBox(container Container, style *LayoutStyle, writer Writer) (err error) {
	defer func() { err = wrapLayoutError("vbox", containerPath(container), err) }()
	if err := validateLayoutInputs(container, style); err != nil {
		return err
	}
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

	fragment, err := measureVBoxFragment(container, style, writer, headers, unaligned, footers)
	if err != nil {
		return err
	}
	if err := layoutVBoxFragment(container, style, writer, fragment); err != nil {
		return err
	}

	// Positioned children are intentionally outside the static stacking
	// algorithm, so they are laid out after the vbox flow has settled.
	return layoutPositionedChildren(container, writer)
}

type vboxMeasuredWidget struct {
	widget Widget
	height float64
}

type vboxFragment struct {
	headers []vboxMeasuredWidget
	body    []vboxMeasuredWidget
	footers []vboxMeasuredWidget
}

func measureVBoxFragment(container Container, style *LayoutStyle, writer Writer, headers, body, footers []Widget) (vboxFragment, error) {
	constrained := container.Height() != 0
	continues := containerHasEffectiveContinuation(container)
	bottom := math.Inf(1)
	if constrained {
		bottom = ContentTop(container) + MaxContentHeight(container)
	}

	fragment := vboxFragment{}
	top := ContentTop(container)
	for _, widget := range headers {
		entry, err := measureVBoxChild(container, writer, widget)
		if err != nil {
			return vboxFragment{}, err
		}
		widget.SetTop(top)
		if widgetZeroFootprint(widget) {
			widget.SetVisible(!continues || widget.Top() <= bottom)
			if widget.Visible() {
				fragment.headers = append(fragment.headers, entry)
			}
			continue
		}
		top += entry.height + style.VPadding()
		widget.SetVisible(!continues || widget.Bottom() <= bottom)
		if widget.Visible() {
			fragment.headers = append(fragment.headers, entry)
		}
	}

	footerTop := bottom
	if len(footers) > 0 {
		footerBottom := bottom
		for i := len(footers) - 1; i >= 0; i-- {
			widget := footers[i]
			entry, err := measureVBoxChild(container, writer, widget)
			if err != nil {
				return vboxFragment{}, err
			}
			widget.SetBottom(footerBottom)
			if widgetZeroFootprint(widget) {
				widget.SetVisible(!continues || widget.Top() >= top)
			} else {
				footerBottom = widget.Top() - style.VPadding()
				widget.SetVisible(!continues || widget.Top() >= top)
			}
			if widget.Visible() {
				fragment.footers = append([]vboxMeasuredWidget{entry}, fragment.footers...)
				footerTop = min(footerTop, widget.Top())
			}
		}
	}

	bodyBottom := bottom
	if len(fragment.footers) > 0 {
		bodyBottom = footerTop - style.VPadding()
	}
	containerFull := false
	bodyPlaced := false
	for i, widget := range body {
		widget.SetVisible(!containerFull)
		if containerFull {
			continue
		}
		entry, err := measureVBoxChild(container, writer, widget)
		if err != nil {
			return vboxFragment{}, err
		}
		widget.SetTop(top)
		if isPageBreak(widget) {
			// The marker belongs to this vbox fragment but consumes no stack
			// height. Once real body content precedes it, hide later content so
			// the page retry/split machinery advances to another fragment.
			widget.SetVisible(true)
			fragment.body = append(fragment.body, entry)
			if continues && bodyPlaced && hasOrdinaryVBoxBodyAfter(body, i+1) {
				containerFull = true
			}
			continue
		}
		if widgetZeroFootprint(widget) {
			widget.SetVisible(!continues || widget.Top() <= bodyBottom)
			if widget.Visible() {
				fragment.body = append(fragment.body, entry)
			}
			continue
		}
		// A continuing child with an internal pgbr must be split even when its
		// measured height fits in the remaining space.
		if constrained && continues && (widgetHasActionablePageBreak(widget) || top+entry.height > bodyBottom+layoutFitEpsilon) {
			containerFull = true
			widget.SetVisible(false)
			continue
		}
		fragment.body = append(fragment.body, entry)
		if widget.Display() == DisplayOnce {
			bodyPlaced = true
		}
		top += entry.height + style.VPadding()
	}

	if !constrained {
		container.ResolveHeight(vboxFragmentStackHeight(style, fragment.headers, fragment.body, fragment.footers) + NonContentHeight(container))
		return fragment, nil
	}
	distributeVBoxAutoHeight(container, style, &fragment)
	return fragment, nil
}

func measureVBoxChild(container Container, writer Writer, widget Widget) (vboxMeasuredWidget, error) {
	if err := resolveVBoxChildWidth(container, writer, widget); err != nil {
		return vboxMeasuredWidget{}, err
	}
	height := widget.Height()
	if !widgetHeightSpecified(widget) {
		var err error
		height, err = widget.PreferredHeight(writer)
		if err != nil {
			return vboxMeasuredWidget{}, err
		}
	}
	if !widgetHeightSpecified(widget) {
		widget.ResolveHeight(height)
		height = widget.Height()
	}
	return vboxMeasuredWidget{widget: widget, height: height}, nil
}

func resolveVBoxChildWidth(container Container, writer Writer, widget Widget) error {
	if widgetAutoWidth(widget) || !widgetWidthSpecified(widget) {
		cw := ContentWidth(container)
		pw := 0.0
		if _, ok := widget.(*StdParagraph); ok {
			pw = cw
		} else {
			var err error
			pw, err = widget.PreferredWidth(writer)
			if err != nil {
				return err
			}
		}
		if pw == 0 {
			pw = cw
		}
		widget.ResolveWidth(min(pw, cw))
	}
	widget.SetLeft(vboxCrossAxisLeft(container, widget, vboxChildCrossAxisRTL(container, widget)))
	return nil
}

func distributeVBoxAutoHeight(container Container, style *LayoutStyle, fragment *vboxFragment) {
	var auto []*vboxMeasuredWidget
	for _, group := range []*[]vboxMeasuredWidget{&fragment.headers, &fragment.body, &fragment.footers} {
		for i := range *group {
			entry := &(*group)[i]
			if widgetAutoHeight(entry.widget) && !widgetZeroFootprint(entry.widget) {
				auto = append(auto, entry)
			}
		}
	}
	if len(auto) == 0 {
		return
	}
	baseline := vboxFragmentStackHeight(style, fragment.headers, fragment.body, fragment.footers)
	if surplus := ContentHeight(container) - baseline; surplus > 0 {
		extra := surplus / float64(len(auto))
		for _, entry := range auto {
			entry.height += extra
			entry.widget.ResolveHeight(entry.height)
		}
	}
}

func layoutVBoxFragment(container Container, style *LayoutStyle, writer Writer, fragment vboxFragment) error {
	top := ContentTop(container)
	bottom := ContentTop(container) + MaxContentHeight(container)
	for _, entry := range fragment.headers {
		widget := entry.widget
		widget.SetTop(top)
		widget.SetVisible(true)
		if !widgetZeroFootprint(widget) {
			top += entry.height + style.VPadding()
		}
		if err := widget.LayoutWidget(writer); err != nil {
			return err
		}
	}
	footerBottom := bottom
	for i := len(fragment.footers) - 1; i >= 0; i-- {
		entry := fragment.footers[i]
		widget := entry.widget
		widget.SetBottom(footerBottom)
		widget.SetVisible(true)
		if !widgetZeroFootprint(widget) {
			footerBottom = widget.Top() - style.VPadding()
		}
		if err := widget.LayoutWidget(writer); err != nil {
			return err
		}
	}
	for _, entry := range fragment.body {
		widget := entry.widget
		widget.SetTop(top)
		widget.SetVisible(true)
		if !widgetZeroFootprint(widget) {
			top += entry.height + style.VPadding()
		}
		if err := widget.LayoutWidget(writer); err != nil {
			return err
		}
	}
	return nil
}

func vboxFragmentStackHeight(style *LayoutStyle, groups ...[]vboxMeasuredWidget) float64 {
	height := 0.0
	seen := 0
	for _, group := range groups {
		for _, entry := range group {
			if widgetZeroFootprint(entry.widget) {
				continue
			}
			if seen > 0 {
				height += style.VPadding()
			}
			height += entry.height
			seen++
		}
	}
	return height
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
	return widgetWidthAuthored(widget) || widgetAspectWidthInferred(widget)
}

func widgetWidthAuthored(widget Widget) bool {
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
	return widgetHeightAuthored(widget) || widgetAspectHeightInferred(widget)
}

func widgetHeightAuthored(widget Widget) bool {
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

func reserveHBoxEdgeFlexibleWidths(widgets []Widget, style *LayoutStyle, writer Writer, widthAvail float64, containerFull *bool) (float64, error) {
	for _, widget := range widgets {
		if widget.Align() != AlignLeft && widget.Align() != AlignRight {
			continue
		}
		width, err := widget.PreferredWidth(writer)
		if err != nil {
			return 0, err
		}
		widget.ResolveWidth(width)
		widthAvail -= width
		*containerFull = widthAvail+layoutFitEpsilon < 0
		widget.SetDisabled(*containerFull)
		widthAvail -= style.HPadding()
	}
	return widthAvail, nil
}

func filterHBoxCenterFlexible(widgets []Widget) []Widget {
	if len(widgets) == 0 {
		return widgets
	}
	center := widgets[:0]
	for _, widget := range widgets {
		if widget.Align() == AlignLeft || widget.Align() == AlignRight {
			continue
		}
		center = append(center, widget)
	}
	return center
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

func vboxChildCrossAxisRTL(container Container, widget Widget) bool {
	if child, ok := widget.(Container); ok {
		return IsRTL(child)
	}
	return IsRTL(container)
}
