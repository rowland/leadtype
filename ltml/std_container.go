// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

type StdContainer struct {
	StdWidget
	Children
	cols            int
	dir             Dir
	dirExplicit     bool
	layout          *LayoutStyle
	listBulletIDs   string
	listPrepared    bool
	order           TableOrder
	paragraphStyle  *ParagraphStyle
	preferredHeight float64
	preferredWidth  float64
	rows            int
	activeChildren  []Widget
	splitEnabled    bool
	splitExplicit   bool
	headerRows      int
	footerRows      int
	baseAngle       float64
	angles          []float64
	radialSweep     radialSweep
	centerX         float64
	centerXSet      bool
	centerY         float64
	centerYSet      bool
	outerRadius     float64
	innerRadius     float64
}

func (c *StdContainer) Cols() int {
	return c.cols
}

func (c *StdContainer) Container() Container {
	return c.container
}

func (c *StdContainer) Dir() Dir {
	if !c.dirExplicit && c.container != nil {
		return c.container.Dir()
	}
	return c.dir
}

func (c *StdContainer) DrawContent(w Writer) error {
	return withWidgetRoleAccessibility(w, &c.StdWidget, "", "", func() error {
		return c.drawChildren(w)
	})
}

func (c *StdContainer) BeforePrint(w Writer) error {
	c.prepareForLayout(w)
	return nil
}

func (c *StdContainer) drawChildren(w Writer) error {
	// fmt.Printf("DrawContent %s\n", c)
	children := slices.Clone(c.Widgets())
	slices.SortStableFunc(children, func(a, b Widget) int {
		return a.ZIndex() - b.ZIndex()
	})
	for _, child := range children {
		if !child.Visible() || child.Disabled() {
			continue
		}
		if err := Print(child, w); err != nil {
			return err
		}
	}
	return nil
}

func (c *StdContainer) LayoutStyle() *LayoutStyle {
	if c.layout == nil {
		return LayoutStyleFor("vbox", c.scope)
	}
	return c.layout
}

func (c *StdContainer) LayoutWidget(w Writer) {
	c.prepareForLayout(w)
	LayoutContainer(c, w)
}

func (c *StdContainer) Widgets() []Widget {
	if c.activeChildren != nil {
		return c.activeChildren
	}
	return c.children
}

func (c *StdContainer) PreferredHeight(w Writer) float64 {
	if c.HeightIsSet() {
		return c.Height()
	}
	c.prepareForLayout(w)
	if isRadialLayoutStyle(c.layout) {
		if height, ok := c.radialInferredHeight(); ok {
			return height
		}
	}
	saved := c.SaveState()
	LayoutContainer(c, newLayoutProbeWriter(w))
	height := c.Height()
	walkWidgets(c, func(widget Widget) bool {
		if widget != c {
			widget.ClearResolvedWidth()
			widget.ClearResolvedHeight()
		}
		return true
	})
	c.RestoreState(saved)
	return height
}

func (c *StdContainer) PreferredWidth(w Writer) float64 {
	if c.WidthIsSet() {
		return c.Width()
	}
	if isRadialLayoutStyle(c.layout) {
		if width, ok := c.radialInferredWidth(); ok {
			return width
		}
	}
	c.prepareForLayout(w)
	static, _ := printableWidgets(c, Static)
	switch c.LayoutStyle().manager {
	case "hbox", "flow":
		width := NonContentWidth(c)
		first := true
		for _, widget := range static {
			if widgetZeroFootprint(widget) {
				continue
			}
			if !first {
				width += c.LayoutStyle().HPadding()
			}
			width += widget.PreferredWidth(w)
			first = false
		}
		return width
	default:
		width := 0.0
		for _, widget := range static {
			if widgetZeroFootprint(widget) {
				continue
			}
			width = max(width, widget.PreferredWidth(w))
		}
		return width + NonContentWidth(c)
	}
}

func (c *StdContainer) Order() TableOrder {
	return c.order
}

func (c *StdContainer) ParagraphStyle() *ParagraphStyle {
	if c.paragraphStyle == nil {
		if c.container == nil {
			return defaultParagraphStyle
		}
		return c.container.ParagraphStyle()
	}
	return c.paragraphStyle
}

func (c *StdContainer) Rows() int {
	return c.rows
}

func (c *StdContainer) SetAttrs(attrs map[string]string) {
	c.StdWidget.SetAttrs(attrs)
	if dirVal, ok := attrs["dir"]; ok {
		c.dirExplicit = true
		c.dir = ParseDir(dirVal)
	}
	if layout, ok := attrs["layout"]; ok {
		c.layout = LayoutStyleFor(layout, c.scope)
	}
	if bullets, ok := attrs["bullets"]; ok {
		c.listBulletIDs = strings.TrimSpace(bullets)
		c.listPrepared = false
	}
	if MapHasKeyPrefix(attrs, "layout.") {
		c.layout = c.LayoutStyle().Clone()
		c.layout.SetAttrs(addUnits(filterMapAttrs("layout.", attrs), c.Units()))
	}
	if order, ok := attrs["order"]; ok {
		switch order {
		case "rows":
			c.order = TableOrderRows
		case "cols":
			c.order = TableOrderCols
		}
	}
	if rows, ok := attrs["rows"]; ok {
		if value, err := strconv.Atoi(rows); err == nil {
			c.rows = value
		}
	}
	if cols, ok := attrs["cols"]; ok {
		if value, err := strconv.Atoi(cols); err == nil {
			c.cols = value
		}
	}
	if baseAngle, ok := attrs["base-angle"]; ok {
		if value, err := strconv.ParseFloat(strings.TrimSpace(baseAngle), 64); err == nil {
			c.baseAngle = value
		}
	}
	c.radialSweep = radialSweepCCW
	if sweep, ok := attrs["sweep"]; ok && strings.EqualFold(strings.TrimSpace(sweep), "cw") {
		c.radialSweep = radialSweepCW
	}
	if angles, ok := attrs["angles"]; ok {
		c.angles = c.angles[:0]
		for _, part := range strings.Split(angles, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			if value, err := strconv.ParseFloat(part, 64); err == nil {
				c.angles = append(c.angles, value)
			}
		}
	}
	if centerX, ok := attrs["center-x"]; ok {
		c.centerX = ParseMeasurement(centerX, c.Units())
		c.centerXSet = true
	}
	if centerY, ok := attrs["center-y"]; ok {
		c.centerY = ParseMeasurement(centerY, c.Units())
		c.centerYSet = true
	}
	if radius, ok := attrs["r"]; ok {
		c.outerRadius = ParseMeasurement(radius, c.Units())
	}
	if radius0, ok := attrs["r0"]; ok {
		c.innerRadius = ParseMeasurement(radius0, c.Units())
	}
	if split, ok := attrs["split"]; ok {
		c.splitExplicit = true
		c.splitEnabled = split != "false"
	}
	if headerRows, ok := attrs["header-rows"]; ok {
		if value, err := strconv.Atoi(headerRows); err == nil {
			c.headerRows = value
		}
	}
	if footerRows, ok := attrs["footer-rows"]; ok {
		if value, err := strconv.Atoi(footerRows); err == nil {
			c.footerRows = value
		}
	}
	if ps, ok := attrs["paragraph-style"]; ok {
		c.paragraphStyle = ParagraphStyleFor(ps, c.scope)
	}
	if MapHasKeyPrefix(attrs, "paragraph-style.") {
		c.paragraphStyle = c.ParagraphStyle().Clone()
		c.paragraphStyle.SetAttrs(filterMapAttrs("paragraph-style.", attrs))
	}
}

func (c *StdContainer) BaseAngle() float64 {
	return c.baseAngle
}

func (c *StdContainer) Angles() []float64 {
	return c.angles
}

func (c *StdContainer) RadialSweep() radialSweep {
	return c.radialSweep
}

func (c *StdContainer) radialInferredHeight() (float64, bool) {
	if !isRadialLayoutStyle(c.layout) {
		return 0, false
	}
	if c.outerRadius > 0 {
		return (c.outerRadius * 2) + NonContentHeight(c), true
	}
	if c.WidthIsSet() {
		diameter := max(ContentWidth(c), c.innerRadius*2)
		if diameter > 0 {
			return diameter + NonContentHeight(c), true
		}
	}
	return 0, false
}

func (c *StdContainer) radialInferredWidth() (float64, bool) {
	if !isRadialLayoutStyle(c.layout) {
		return 0, false
	}
	if c.outerRadius > 0 {
		return (c.outerRadius * 2) + NonContentWidth(c), true
	}
	if c.HeightIsSet() {
		diameter := max(ContentHeight(c), c.innerRadius*2)
		if diameter > 0 {
			return diameter + NonContentWidth(c), true
		}
	}
	return 0, false
}

func (c *StdContainer) CenterX() (float64, bool) {
	return c.centerX, c.centerXSet
}

func (c *StdContainer) CenterY() (float64, bool) {
	return c.centerY, c.centerYSet
}

func (c *StdContainer) OuterRadius() float64 {
	return c.outerRadius
}

func (c *StdContainer) RadiusValue() float64 {
	return c.OuterRadius()
}

func (c *StdContainer) InnerRadius() float64 {
	return c.innerRadius
}

func (c *StdContainer) String() string {
	return fmt.Sprintf("StdContainer layout=%v paragraphStyle=%v %s", c.layout, c.paragraphStyle, &c.StdWidget)
}

var errTableSplitUnsupportedRowSpan = errors.New("table splitting does not support rowspan > 1")

func (c *StdContainer) SplitForHeight(avail float64, w Writer) (*SplitResult, error) {
	c.prepareForLayout(w)
	if c.LayoutStyle() == nil {
		return nil, nil
	}
	switch c.LayoutStyle().manager {
	case "table":
		if !c.tableSplitEnabled() {
			return nil, nil
		}
		metrics, err := c.tableSplitMetrics(w)
		if err != nil {
			return nil, err
		}
		bodyCount := metrics.bodyEnd - metrics.bodyStart
		if c.tableFragmentHeight(metrics, metrics.bodyStart, metrics.bodyEnd) <= avail {
			rows := append([]int{}, metrics.headerRows...)
			for r := metrics.bodyStart; r < metrics.bodyEnd; r++ {
				rows = append(rows, r)
			}
			rows = append(rows, metrics.footerRows...)
			return &SplitResult{Head: c.cloneTableFragment(metrics, rows), Tail: nil}, nil
		}
		if bodyCount < 2 {
			return nil, nil
		}
		fitBodies := 0
		for n := 1; n < bodyCount; n++ {
			if c.tableFragmentHeight(metrics, metrics.bodyStart, metrics.bodyStart+n) <= avail {
				fitBodies = n
				continue
			}
			break
		}
		if fitBodies == 0 {
			return nil, nil
		}
		headRows := append([]int{}, metrics.headerRows...)
		for r := metrics.bodyStart; r < metrics.bodyStart+fitBodies; r++ {
			headRows = append(headRows, r)
		}
		headRows = append(headRows, metrics.footerRows...)

		tailRows := append([]int{}, metrics.headerRows...)
		for r := metrics.bodyStart + fitBodies; r < metrics.bodyEnd; r++ {
			tailRows = append(tailRows, r)
		}
		tailRows = append(tailRows, metrics.footerRows...)

		head := c.cloneTableFragment(metrics, headRows)
		tail := c.cloneTableFragment(metrics, tailRows)
		return &SplitResult{Head: head, Tail: tail}, nil
	case "vbox":
		if !c.vboxSplitEnabled() {
			return nil, nil
		}
		return c.splitVBoxForHeight(avail, w)
	default:
		return nil, nil
	}
}

func (c *StdContainer) SplitEnabled() bool {
	if c.LayoutStyle() == nil {
		return false
	}
	switch c.LayoutStyle().manager {
	case "table":
		return c.tableSplitEnabled()
	case "vbox":
		return c.vboxSplitEnabled()
	default:
		return false
	}
}

func (c *StdContainer) tableSplitEnabled() bool {
	if c.splitExplicit {
		return c.splitEnabled
	}
	return c.LayoutStyle() != nil && c.LayoutStyle().manager == "table"
}

func (c *StdContainer) vboxSplitEnabled() bool {
	if c.splitExplicit {
		return c.splitEnabled
	}
	return c.LayoutStyle() != nil && c.LayoutStyle().manager == "vbox"
}

type tableSplitMetrics struct {
	grid       *WidgetGrid
	rowHeights []float64
	headerRows []int
	footerRows []int
	bodyStart  int
	bodyEnd    int
}

func (c *StdContainer) tableSplitMetrics(w Writer) (*tableSplitMetrics, error) {
	var grid *WidgetGrid
	var err error
	if c.Order() == TableOrderRows {
		grid, err = rowGrid(c)
	} else {
		grid, err = colGrid(c)
	}
	if err != nil {
		return nil, err
	}
	for _, widget := range c.Widgets() {
		if widget.RowSpan() > 1 {
			return nil, errTableSplitUnsupportedRowSpan
		}
	}
	widths := planTableColumnWidths(grid, c, c.LayoutStyle(), w).resolvedSizes()

	rowHeights := make([]float64, grid.Rows())
	for r := 0; r < grid.Rows(); r++ {
		maxHeight := 0.0
		for col := 0; col < grid.Cols(); col++ {
			widget := grid.Cell(col, r)
			if widget == nil {
				continue
			}
			if widths[col] <= 0 {
				continue
			}
			widget.ResolveWidth(tableCellWidth(widths, col, widget.ColSpan(), c.LayoutStyle().HPadding()))
			height := widget.Height()
			if !widgetHeightSpecified(widget) {
				height = widget.PreferredHeight(w)
			}
			if height > maxHeight {
				maxHeight = height
			}
		}
		rowHeights[r] = maxHeight
	}
	headerCount := min(c.headerRows, grid.Rows())
	footerCount := min(c.footerRows, max(0, grid.Rows()-headerCount))
	headerRows := make([]int, 0, headerCount)
	for i := 0; i < headerCount; i++ {
		headerRows = append(headerRows, i)
	}
	footerRows := make([]int, 0, footerCount)
	for i := grid.Rows() - footerCount; i < grid.Rows(); i++ {
		if i >= headerCount {
			footerRows = append(footerRows, i)
		}
	}
	return &tableSplitMetrics{
		grid:       grid,
		rowHeights: rowHeights,
		headerRows: headerRows,
		footerRows: footerRows,
		bodyStart:  headerCount,
		bodyEnd:    grid.Rows() - footerCount,
	}, nil
}

func (c *StdContainer) tableFragmentHeight(metrics *tableSplitMetrics, bodyStart, bodyEnd int) float64 {
	rows := make([]int, 0, len(metrics.headerRows)+(bodyEnd-bodyStart)+len(metrics.footerRows))
	rows = append(rows, metrics.headerRows...)
	for r := bodyStart; r < bodyEnd; r++ {
		rows = append(rows, r)
	}
	rows = append(rows, metrics.footerRows...)
	height := NonContentHeight(c)
	for i, row := range rows {
		height += metrics.rowHeights[row]
		if i > 0 {
			height += c.LayoutStyle().VPadding()
		}
	}
	return height
}

func (c *StdContainer) cloneTableFragment(metrics *tableSplitMetrics, rows []int) *StdContainer {
	clone := *c
	clone.activeChildren = c.cloneTableWidgetsForRows(metrics.grid, rows, &clone)
	clone.ClearResolvedWidth()
	clone.ClearResolvedHeight()
	clone.printed = false
	clone.invisible = false
	clone.disabled = false
	clone.path = ""
	for _, child := range clone.activeChildren {
		child.SetPrinted(false)
		child.SetVisible(true)
		child.SetDisabled(false)
	}
	return &clone
}

func (c *StdContainer) cloneTableWidgetsForRows(grid *WidgetGrid, rows []int, parent Container) []Widget {
	var widgets []Widget
	seen := map[Widget]bool{}
	for _, r := range rows {
		for col := 0; col < grid.Cols(); col++ {
			widget := grid.Cell(col, r)
			if widget == nil || seen[widget] {
				continue
			}
			seen[widget] = true
			clone := cloneWidgetShallow(widget)
			if wc, ok := clone.(WantsContainer); ok {
				_ = wc.SetContainer(parent)
			}
			widgets = append(widgets, clone)
		}
	}
	return widgets
}

const (
	defaultListBulletGap  = 6.0
	defaultListBulletSize = 18.0
)

type listBulletTemplate struct {
	bullet    *BulletStyle
	autoWidth bool
	width     float64
}

type vboxSplitMetrics struct {
	headers    []Widget
	body       []Widget
	footers    []Widget
	positioned map[Widget]bool
	heights    map[Widget]float64
}

func (c *StdContainer) prepareForLayout(w Writer) {
	c.prepareListBullets(w)
}

func (c *StdContainer) prepareListBullets(w Writer) {
	if c.listPrepared || c.LayoutStyle() == nil || c.LayoutStyle().manager != "vbox" {
		return
	}
	templates := c.listBulletTemplates()
	if len(templates) == 0 {
		return
	}
	c.measureFormattedListBulletWidths(w, templates)
	itemNo := 0
	for _, child := range c.children {
		para, ok := child.(*StdParagraph)
		if !ok {
			continue
		}
		itemNo++
		if len(para.Bullets()) > 0 {
			continue
		}
		para.bullets = materializeListBullets(templates, itemNo)
	}
	c.listPrepared = true
}

func (c *StdContainer) listBulletTemplates() []listBulletTemplate {
	var templates []listBulletTemplate
	for _, id := range strings.Fields(c.listBulletIDs) {
		bullet := BulletStyleFor(id, c.scope)
		if bullet == nil {
			continue
		}
		clone := bullet.Clone()
		width := clone.Width()
		autoWidth := clone.IsFormatted() && !clone.WidthIsSet()
		if autoWidth {
			width = 0
		}
		templates = append(templates, listBulletTemplate{
			bullet:    clone,
			autoWidth: autoWidth,
			width:     width,
		})
	}
	return templates
}

func (c *StdContainer) measureFormattedListBulletWidths(w Writer, templates []listBulletTemplate) {
	itemNo := 0
	for _, child := range c.children {
		para, ok := child.(*StdParagraph)
		if !ok {
			continue
		}
		itemNo++
		for i := range templates {
			if !templates[i].autoWidth {
				continue
			}
			templates[i].width = max(templates[i].width, c.formattedListMarkerWidth(w, para, templates[i].bullet, itemNo))
		}
	}
}

func materializeListBullets(templates []listBulletTemplate, itemNo int) []*BulletStyle {
	bullets := make([]*BulletStyle, 0, len(templates))
	for _, template := range templates {
		bullet := template.bullet.Clone()
		if bullet.IsFormatted() {
			bullet.text = fmt.Sprintf(bullet.Format(), itemNo)
			bullet.src = ""
			bullet.shape = ""
			if template.autoWidth {
				bullet.width = template.width
				bullet.widthSet = true
			}
		}
		bullets = append(bullets, bullet)
	}
	return bullets
}

func (c *StdContainer) formattedListMarkerWidth(w Writer, para *StdParagraph, template *BulletStyle, itemNo int) float64 {
	marker := template.Clone()
	marker.text = fmt.Sprintf(template.Format(), itemNo)
	marker.src = ""
	marker.shape = ""
	width := para.bulletTextWidth(w, marker)
	if width <= 0 {
		width = marker.Width()
	}
	return width + defaultListBulletGap
}

func (c *StdContainer) splitVBoxForHeight(avail float64, w Writer) (*SplitResult, error) {
	metrics := c.vboxSplitMetrics(w)
	if len(metrics.body) == 0 {
		return nil, nil
	}
	if c.vboxFragmentHeight(metrics, nil) > avail {
		return nil, nil
	}

	headWhole := make(map[Widget]bool)
	tailWhole := make(map[Widget]bool)
	for _, child := range metrics.headers {
		headWhole[child] = true
		tailWhole[child] = true
	}
	for _, child := range metrics.footers {
		headWhole[child] = true
		tailWhole[child] = true
	}
	for child := range metrics.positioned {
		headWhole[child] = true
		tailWhole[child] = true
	}

	var splitSource Widget
	var splitHead Widget
	var splitTail Widget
	bodyIncluded := 0
	for i, child := range metrics.body {
		c.measureVBoxSplitChild(metrics, child, w)
		candidate := append([]Widget(nil), metrics.body[:i+1]...)
		if c.vboxFragmentHeight(metrics, candidate) <= avail {
			headWhole[child] = true
			bodyIncluded++
			continue
		}
		availForChild := c.vboxSplitAvailableForChild(metrics, metrics.body[:i], child, avail)
		if splittable, ok := child.(Splittable); ok && availForChild > 0 {
			result, err := splittable.SplitForHeight(availForChild, w)
			if err != nil {
				return nil, err
			}
			if result != nil && result.Head != nil {
				splitSource = child
				splitHead = result.Head
				splitTail = result.Tail
				bodyIncluded++
			}
		}
		break
	}
	if bodyIncluded == 0 {
		return nil, nil
	}

	for i, child := range metrics.body {
		if splitSource != nil && child == splitSource {
			if splitTail != nil {
				tailWhole[child] = false
			}
			for _, remaining := range metrics.body[i+1:] {
				tailWhole[remaining] = true
			}
			break
		}
		if headWhole[child] {
			continue
		}
		tailWhole[child] = true
	}
	if splitSource == nil {
		for _, child := range metrics.body {
			if !headWhole[child] {
				tailWhole[child] = true
			}
		}
	} else if splitTail == nil {
		delete(tailWhole, splitSource)
	}

	headReplacements := map[Widget]Widget{}
	if splitSource != nil && splitHead != nil {
		headReplacements[splitSource] = splitHead
	}
	head := c.cloneVBoxFragment(headWhole, headReplacements)
	tailReplacements := map[Widget]Widget{}
	if splitSource != nil && splitTail != nil {
		tailReplacements[splitSource] = splitTail
	}
	tail := c.cloneVBoxFragment(tailWhole, tailReplacements)
	if len(tail.Widgets()) == 0 {
		return &SplitResult{Head: head, Tail: nil}, nil
	}
	return &SplitResult{Head: head, Tail: tail}, nil
}

func (c *StdContainer) vboxSplitMetrics(w Writer) *vboxSplitMetrics {
	static, _ := printableWidgets(c, Static)
	absolute, _ := printableWidgets(c, Absolute)
	relative, _ := printableWidgets(c, Relative)
	metrics := &vboxSplitMetrics{
		positioned: make(map[Widget]bool, len(absolute)+len(relative)),
		heights:    make(map[Widget]float64, len(static)),
	}
	for _, child := range absolute {
		metrics.positioned[child] = true
	}
	for _, child := range relative {
		metrics.positioned[child] = true
	}
	for _, child := range static {
		switch child.Align() {
		case AlignTop:
			c.measureVBoxSplitChild(metrics, child, w)
			metrics.headers = append(metrics.headers, child)
		case AlignBottom:
			c.measureVBoxSplitChild(metrics, child, w)
			metrics.footers = append(metrics.footers, child)
		default:
			metrics.body = append(metrics.body, child)
		}
	}
	return metrics
}

func (c *StdContainer) measureVBoxSplitChild(metrics *vboxSplitMetrics, child Widget, w Writer) {
	if _, ok := metrics.heights[child]; ok {
		return
	}
	entry := measureVBoxChild(c, w, child)
	metrics.heights[child] = entry.height
}

func (c *StdContainer) vboxFragmentHeight(metrics *vboxSplitMetrics, body []Widget) float64 {
	return NonContentHeight(c) + c.vboxStackHeight(metrics, metrics.headers, body, metrics.footers)
}

func (c *StdContainer) vboxStackHeight(metrics *vboxSplitMetrics, groups ...[]Widget) float64 {
	height := 0.0
	seen := 0
	for _, group := range groups {
		for _, child := range group {
			if widgetZeroFootprint(child) {
				continue
			}
			if seen > 0 {
				height += c.LayoutStyle().VPadding()
			}
			height += metrics.heights[child]
			seen++
		}
	}
	return height
}

func (c *StdContainer) vboxSplitAvailableForChild(metrics *vboxSplitMetrics, bodyBefore []Widget, child Widget, avail float64) float64 {
	contentAvail := avail - NonContentHeight(c)
	base := c.vboxStackHeight(metrics, metrics.headers, bodyBefore, metrics.footers)
	if contentAvail <= base {
		return 0
	}
	childVisible := !widgetZeroFootprint(child)
	prevVisible := c.vboxHasVisibleWidget(metrics.headers, bodyBefore)
	footerVisible := c.vboxHasVisibleWidget(metrics.footers)
	extraGap := 0.0
	if childVisible {
		if prevVisible {
			extraGap += c.LayoutStyle().VPadding()
		}
		if footerVisible {
			extraGap += c.LayoutStyle().VPadding()
		}
		if prevVisible && footerVisible {
			extraGap -= c.LayoutStyle().VPadding()
		}
	}
	availForChild := contentAvail - base - extraGap
	if availForChild < 0 {
		return 0
	}
	return availForChild
}

func (c *StdContainer) vboxHasVisibleWidget(groups ...[]Widget) bool {
	for _, group := range groups {
		for _, child := range group {
			if !widgetZeroFootprint(child) {
				return true
			}
		}
	}
	return false
}

func (c *StdContainer) cloneVBoxFragment(included map[Widget]bool, replacements map[Widget]Widget) *StdContainer {
	clone := *c
	clone.activeChildren = make([]Widget, 0, len(included)+len(replacements))
	clone.ClearResolvedWidth()
	clone.ClearResolvedHeight()
	clone.printed = false
	clone.invisible = false
	clone.disabled = false
	clone.path = ""
	for _, child := range c.Widgets() {
		replacement, replaced := replacements[child]
		if replaced && replacement == nil {
			continue
		}
		if !replaced && !included[child] {
			continue
		}
		next := replacement
		if !replaced {
			next = cloneWidgetShallow(child)
		}
		if wc, ok := next.(WantsContainer); ok {
			_ = wc.SetContainer(&clone)
		}
		next.SetPrinted(false)
		next.SetVisible(true)
		next.SetDisabled(false)
		clone.activeChildren = append(clone.activeChildren, next)
	}
	return &clone
}

func cloneWidgetShallow(widget Widget) Widget {
	value := reflect.ValueOf(widget)
	if value.Kind() != reflect.Pointer || value.IsNil() {
		panic("cloneWidgetShallow expects non-nil pointer widget")
	}
	if accessible, ok := widget.(interface{ AccessibilityLogicalID() string }); ok {
		accessible.AccessibilityLogicalID()
	}
	clone := reflect.New(value.Elem().Type())
	clone.Elem().Set(value.Elem())
	w, ok := clone.Interface().(Widget)
	if !ok {
		panic("cloneWidgetShallow produced non-widget clone")
	}
	w.ClearResolvedWidth()
	w.ClearResolvedHeight()
	w.SetPrinted(false)
	w.SetVisible(true)
	w.SetDisabled(false)
	return w
}

func init() {
	registerTag(DefaultSpace, "div", func() any { return &StdContainer{} })
}

var _ Container = (*StdContainer)(nil)
var _ HasAttrs = (*StdContainer)(nil)
var _ Identifier = (*StdContainer)(nil)
var _ Printer = (*StdContainer)(nil)
var _ WantsContainer = (*StdContainer)(nil)
