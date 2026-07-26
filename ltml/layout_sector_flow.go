package ltml

import (
	"math"
	"slices"
	"strconv"
	"strings"
)

type sectorFlowItem struct {
	widget   Widget
	width    float64
	height   float64
	fullBand bool
}

type sectorFlowRow struct {
	start  int
	end    int
	width  float64
	height float64
}

type sectorFlowScore struct {
	contained int
	overflow  float64
	distance  float64
	height    float64
	unused    float64
	signature string
}

type sectorFlowLayout struct {
	rows         []sectorFlowRow
	slots        map[Widget]radialBounds
	labelAnchors map[*StdLabel]sectorFlowLabelAnchor
	score        sectorFlowScore
}

func (s *StdSector) layoutStaticFlow(w Writer) error {
	for pass := 0; pass < 5; pass++ {
		stable, err := s.layoutStaticFlowPass(w)
		if err != nil || stable {
			return err
		}
	}
	return nil
}

func (s *StdSector) layoutStaticFlowPass(w Writer) (bool, error) {
	static, _ := printableWidgets(s, Static)
	visible := make(map[Widget]bool, len(static))
	for _, child := range static {
		visible[child] = true
		child.SetVisible(true)
	}
	for _, child := range s.Widgets() {
		if child.Position() == Static && !visible[child] {
			child.SetVisible(false)
		}
	}
	if len(static) == 0 {
		s.flowSlots = nil
		s.flowLabelAnchors = nil
		return true, nil
	}

	items, err := s.sectorFlowItems(static, w)
	if err != nil {
		return false, err
	}
	hgap, vgap := 0.0, 0.0
	if style := s.LayoutStyle(); style != nil {
		hgap, vgap = style.HPadding(), style.VPadding()
	}
	partitions := sectorFlowPartitions(items, hgap, s.contentBounds.MaxX-s.contentBounds.MinX)
	var best *sectorFlowLayout
	for _, rows := range partitions {
		candidate := s.placeSectorFlowRows(items, rows, hgap, vgap)
		if candidate != nil && (best == nil || betterSectorFlowScore(candidate.score, best.score)) {
			best = candidate
		}
	}
	if best == nil {
		return true, nil
	}

	s.flowSlots = best.slots
	s.flowLabelAnchors = best.labelAnchors
	s.invalidateTextLayouts()
	stable := true
	for _, item := range items {
		slot := best.slots[item.widget]
		if label, ok := item.widget.(*StdLabel); ok {
			if err := label.LayoutWidget(w); err != nil {
				return false, err
			}
			continue
		}
		paragraph, isParagraph := item.widget.(*StdParagraph)
		if isParagraph {
			item.widget.SetLeft(s.geometry.AnchorX + slot.MinX)
			item.widget.SetTop(s.geometry.AnchorY + slot.MinY)
		} else {
			centerX, centerY := rotatePagePoint(
				s.geometry.AnchorX+(slot.MinX+slot.MaxX)/2,
				s.geometry.AnchorY+(slot.MinY+slot.MaxY)/2,
				s.geometry.AnchorX, s.geometry.AnchorY, s.flowRotation,
			)
			item.widget.SetLeft(centerX - item.width/2)
			item.widget.SetTop(centerY - item.height/2)
		}
		if isParagraph &&
			(paragraph.curvedInSector() || paragraph.HeightMode() == DimUnspecified || paragraph.HeightMode() == DimAuto) {
			paragraph.ClearResolvedHeight()
			s.paragraphLayouts = nil
			height, err := paragraph.PreferredHeight(w)
			if err != nil {
				return false, err
			}
			paragraph.ResolveHeight(height)
			if math.Abs(height-item.height) > 0.25 {
				stable = false
			}
		}
		if err := item.widget.LayoutWidget(w); err != nil {
			return false, err
		}
	}
	return stable, nil
}

func (s *StdSector) sectorFlowItems(widgets []Widget, w Writer) ([]sectorFlowItem, error) {
	items := make([]sectorFlowItem, 0, len(widgets))
	seedWidth := max(s.seedContentWidth(), 1)
	_, centerY := s.contentLocalCenter()
	for _, child := range widgets {
		item := sectorFlowItem{widget: child}
		if label, ok := child.(*StdLabel); ok {
			if _, straight := label.sectorTextAngle(); !straight {
				rt := label.RichText(w)
				if rt != nil {
					item.width = rt.Width()
					item.height = rt.Leading() * w.LineSpacing()
				}
				if item.height <= 0 {
					item.height = effectiveFontSizeForContainer(label) * w.LineSpacing()
				}
				items = append(items, item)
				continue
			}
		}
		if paragraph, ok := child.(*StdParagraph); ok {
			item.fullBand = true
			if paragraph.curvedInSector() {
				s.paragraphLayouts = nil
				layout := s.sectorParagraphLayoutFor(paragraph, w)
				if layout.err != nil {
					return nil, layout.err
				}
				item.height = layout.total
				paragraph.ResolveWidth(0)
				paragraph.ResolveHeight(item.height)
				items = append(items, item)
				continue
			}
			if child.WidthMode() == DimUnspecified || child.WidthMode() == DimAuto {
				child.ResolveWidth(seedWidth)
			}
			if slot, ok := s.flowSlots[child]; ok {
				child.SetTop(s.geometry.AnchorY + slot.MinY)
			} else {
				child.SetTop(s.geometry.AnchorY + centerY)
			}
			s.paragraphLayouts = nil
			if child.HeightMode() == DimUnspecified || child.HeightMode() == DimAuto {
				child.ClearResolvedHeight()
			}
		}
		if child.WidthIsSet() {
			item.width = child.Width()
		} else {
			width, err := child.PreferredWidth(w)
			if err != nil {
				return nil, err
			}
			if width <= 0 {
				width = seedWidth
			}
			item.width = min(width, seedWidth)
			child.ResolveWidth(item.width)
		}
		if child.HeightIsSet() {
			item.height = child.Height()
		} else {
			height, err := child.PreferredHeight(w)
			if err != nil {
				return nil, err
			}
			item.height = height
			child.ResolveHeight(height)
		}
		// Paragraphs consume a complete radial band. Their horizontal footprint
		// is the sequence of sector-shaped line intervals, not a rectangle with
		// the paragraph's seed width. Keeping a rectangular width here makes the
		// row scorer move paragraphs toward whichever chord happens to contain
		// that rectangle, even though drawing never uses that rectangle.
		if item.fullBand {
			item.width = 0
		}
		items = append(items, item)
	}
	return items, nil
}

// sectorFlowPartitions uses dynamic programming at every meaningful row-width
// threshold. This preserves source order while considering non-greedy breaks.
func sectorFlowPartitions(items []sectorFlowItem, gap, maxWidth float64) [][]sectorFlowRow {
	thresholds := []float64{maxWidth}
	for start := range items {
		width := 0.0
		for end := start; end < len(items); end++ {
			if end > start {
				width += gap
			}
			width += items[end].width
			thresholds = append(thresholds, width)
			if items[start].fullBand || items[end].fullBand {
				break
			}
		}
	}
	slices.Sort(thresholds)
	thresholds = slices.CompactFunc(thresholds, func(a, b float64) bool { return math.Abs(a-b) < 0.001 })
	seen := make(map[string]bool)
	result := make([][]sectorFlowRow, 0, len(thresholds)+1)
	for _, threshold := range thresholds {
		rows := sectorFlowPartitionForWidth(items, gap, max(threshold, 0))
		signature := sectorFlowRowsSignature(rows)
		if len(rows) > 0 && !seen[signature] {
			seen[signature] = true
			result = append(result, rows)
		}
	}
	return result
}

type sectorFlowDPState struct {
	valid    bool
	overflow float64
	height   float64
	unused   float64
	rows     []sectorFlowRow
}

func sectorFlowPartitionForWidth(items []sectorFlowItem, gap, target float64) []sectorFlowRow {
	dp := make([]sectorFlowDPState, len(items)+1)
	dp[len(items)] = sectorFlowDPState{valid: true}
	for start := len(items) - 1; start >= 0; start-- {
		width, height := 0.0, 0.0
		for end := start + 1; end <= len(items); end++ {
			item := items[end-1]
			if end-start > 1 {
				if item.fullBand || items[start].fullBand {
					break
				}
				width += gap
			}
			width += item.width
			height = max(height, item.height)
			if !dp[end].valid {
				continue
			}
			rowOverflow := max(width-target, 0)
			state := sectorFlowDPState{
				valid:    true,
				overflow: rowOverflow + dp[end].overflow,
				height:   height + dp[end].height,
				unused:   max(target-width, 0) + dp[end].unused,
				rows:     append([]sectorFlowRow{{start: start, end: end, width: width, height: height}}, dp[end].rows...),
			}
			current := dp[start]
			if !current.valid || state.overflow < current.overflow-0.001 ||
				(math.Abs(state.overflow-current.overflow) < 0.001 && (state.height < current.height-0.001 ||
					(math.Abs(state.height-current.height) < 0.001 && state.unused < current.unused))) {
				dp[start] = state
			}
			if item.fullBand {
				break
			}
		}
	}
	return dp[0].rows
}

func (s *StdSector) placeSectorFlowRows(items []sectorFlowItem, rows []sectorFlowRow, hgap, vgap float64) *sectorFlowLayout {
	if len(rows) == 0 {
		return nil
	}
	totalHeight := -vgap
	rowOffsets := make([]float64, len(rows))
	maxWidth := 0.0
	for i, row := range rows {
		rowOffsets[i] = totalHeight + vgap
		totalHeight += row.height + vgap
		maxWidth = max(maxWidth, row.width)
	}
	centerX, centerY := s.contentLocalCenter()
	allCurved := true
	for _, item := range items {
		if !sectorFlowItemIsCurved(item) {
			allCurved = false
			break
		}
	}
	if allCurved {
		// Curved flow is centered in the sector's polar band, whose local
		// origin is the midpoint angle and midpoint radius. An area centroid
		// moves toward the diameter of wide annular sectors and can put text on
		// the inner boundary instead of in the middle of its track.
		centerX, centerY = 0, s.contentPolarCenterLocalY()
	}
	tops := []float64{
		centerY - totalHeight/2,
		s.contentBounds.MinY,
		s.contentBounds.MaxY - totalHeight,
	}
	for _, point := range s.contentPolygon {
		for i, row := range rows {
			tops = append(tops, point.Y-rowOffsets[i], point.Y-rowOffsets[i]-row.height)
		}
	}
	var best *sectorFlowLayout
	for _, top := range tops {
		slots := make(map[Widget]radialBounds, len(items))
		labelAnchors := make(map[*StdLabel]sectorFlowLabelAnchor)
		contained, overflow, unused := 0, 0.0, 0.0
		centroidX, centroidY, centroidWeight := 0.0, 0.0, 0.0
		for rowIndex, row := range rows {
			rowTop := top + rowOffsets[rowIndex]
			curvedRow := sectorFlowRowIsCurved(items, row)
			fullBandRow := row.end == row.start+1 && items[row.start].fullBand
			band := s.contentBandForHeight(rowTop, row.height)
			if fullBandRow {
				// Paragraph lines resolve their own width at each baseline. For flow
				// placement we need only a representative horizontal center and the
				// vertical extent; intersecting every chord across the complete
				// paragraph height can collapse a perfectly usable shaped band.
				band = s.contentLineIntervalAt(rowTop + row.height/2)
			}
			if curvedRow {
				arcWidth := s.flowArcWidthAtLocalY(rowTop + row.height/2)
				band = radialInterval{MinX: -arcWidth / 2, MaxX: arcWidth / 2}
			}
			bandWidth := max(band.MaxX-band.MinX, 0)
			rowCenterX := centerX
			if curvedRow {
				rowCenterX = 0
			}
			envelopeLeft := rowCenterX - maxWidth/2
			if maxWidth <= bandWidth {
				envelopeLeft = clampFloat(envelopeLeft, band.MinX, band.MaxX-maxWidth)
			} else {
				envelopeLeft = band.MinX
			}
			rowLeft := envelopeLeft
			if IsRTL(s) {
				rowLeft += maxWidth - row.width
			}
			x := rowLeft
			rowAxisX := envelopeLeft + maxWidth/2
			for itemIndex := row.start; itemIndex < row.end; itemIndex++ {
				item := items[itemIndex]
				if IsRTL(s) {
					x = rowLeft + row.width - item.width
					for preceding := row.start; preceding < itemIndex; preceding++ {
						x -= items[preceding].width + hgap
					}
				}
				slot := radialBounds{MinX: x, MinY: rowTop, MaxX: x + item.width, MaxY: rowTop + item.height}
				if item.fullBand {
					slot.MinX, slot.MaxX = band.MinX, band.MaxX
				}
				slots[item.widget] = slot
				if label, ok := item.widget.(*StdLabel); ok {
					if _, straight := label.sectorTextAngle(); !straight {
						factor := 0.5
						switch label.sectorTextAlign() {
						case HAlignLeft:
							factor = 0
						case HAlignRight:
							factor = 1
						}
						yFactor := 0.5
						switch label.sectorTextVAlign() {
						case VAlignTop:
							yFactor = 0
						case VAlignBottom:
							yFactor = 1
						}
						arcOffset := slot.MinX + item.width*factor - rowAxisX
						anchor := s.flowLabelAnchorAt(rowTop+item.height*yFactor, arcOffset)
						anchor.arcWidth = max(min(slot.MaxX, band.MaxX)-max(slot.MinX, band.MinX), 0)
						labelAnchors[label] = anchor
					}
				}
				itemOverflow := max(s.contentBounds.MinY-slot.MinY, 0) + max(slot.MaxY-s.contentBounds.MaxY, 0)
				if !item.fullBand {
					itemOverflow += max(band.MinX-slot.MinX, 0) + max(slot.MaxX-band.MaxX, 0)
				}
				if itemOverflow <= 0.001 {
					contained++
				}
				overflow += itemOverflow
				weight := max((slot.MaxX-slot.MinX)*item.height, 1)
				centroidX += ((slot.MinX + slot.MaxX) / 2) * weight
				centroidY += ((slot.MinY + slot.MaxY) / 2) * weight
				centroidWeight += weight
				if !IsRTL(s) {
					x += item.width + hgap
				}
			}
			if !fullBandRow {
				unused += max(bandWidth-row.width, 0)
			}
		}
		if centroidWeight > 0 {
			centroidX /= centroidWeight
			centroidY /= centroidWeight
		}
		score := sectorFlowScore{
			contained: contained,
			overflow:  overflow,
			distance:  math.Hypot(centroidX-centerX, centroidY-centerY),
			height:    totalHeight,
			unused:    unused,
			signature: sectorFlowRowsSignature(rows),
		}
		candidate := &sectorFlowLayout{rows: rows, slots: slots, labelAnchors: labelAnchors, score: score}
		if best == nil || betterSectorFlowScore(score, best.score) {
			best = candidate
		}
	}
	return best
}

func sectorFlowRowIsCurved(items []sectorFlowItem, row sectorFlowRow) bool {
	if row.start == row.end {
		return false
	}
	for i := row.start; i < row.end; i++ {
		if !sectorFlowItemIsCurved(items[i]) {
			return false
		}
	}
	return true
}

func sectorFlowItemIsCurved(item sectorFlowItem) bool {
	switch widget := item.widget.(type) {
	case *StdLabel:
		_, straight := widget.sectorTextAngle()
		return !straight
	case *StdParagraph:
		return widget.curvedInSector()
	default:
		return false
	}
}

func (s *StdSector) flowArcWidthAtLocalY(localY float64) float64 {
	anchor := s.flowLabelAnchorAt(localY, 0)
	if anchor.x == s.geometry.CenterX && anchor.y == s.geometry.CenterY {
		return 0
	}
	radius := math.Hypot(anchor.x-s.geometry.CenterX, anchor.y-s.geometry.CenterY)
	return s.contentArcWidth(radius)
}

func (s *StdSector) flowLabelAnchorAt(localY, arcOffset float64) sectorFlowLabelAnchor {
	baseX, baseY := rotatePagePoint(s.geometry.AnchorX, s.geometry.AnchorY+localY,
		s.geometry.AnchorX, s.geometry.AnchorY, s.flowRotation)
	radius := math.Hypot(baseX-s.geometry.CenterX, baseY-s.geometry.CenterY)
	if radius <= radialAngleEpsilon {
		return sectorFlowLabelAnchor{x: baseX, y: baseY}
	}
	baseAngle := math.Atan2(s.geometry.CenterY-baseY, baseX-s.geometry.CenterX) * 180 / math.Pi
	testX, testY := rotatePagePoint(s.geometry.AnchorX+1, s.geometry.AnchorY+localY,
		s.geometry.AnchorX, s.geometry.AnchorY, s.flowRotation)
	testAngle := math.Atan2(s.geometry.CenterY-testY, testX-s.geometry.CenterX) * 180 / math.Pi
	delta := testAngle - baseAngle
	for delta > 180 {
		delta -= 360
	}
	for delta < -180 {
		delta += 360
	}
	direction := 1.0
	if delta < 0 {
		direction = -1
	}
	angle := baseAngle + direction*arcOffset/radius*180/math.Pi
	x, y := radialPointAt(s.geometry.CenterX, s.geometry.CenterY, radius, angle)
	return sectorFlowLabelAnchor{x: x, y: y}
}

func (s *StdSector) contentBandForHeight(top, height float64) radialInterval {
	return polygonBandForHeight(s.contentPolygon, s.contentBounds, top, height)
}

func polygonBandForHeight(polygon []radialPoint, bounds radialBounds, top, height float64) radialInterval {
	yValues := []float64{top, top + height/2, top + height}
	for _, point := range polygon {
		if point.Y > top && point.Y < top+height {
			yValues = append(yValues, point.Y)
		}
	}
	centerX, _, ok := polygonCentroid(polygon)
	if !ok {
		centerX = (bounds.MinX + bounds.MaxX) / 2
	}
	result := radialInterval{MinX: math.Inf(-1), MaxX: math.Inf(1)}
	for _, y := range yValues {
		interval := polygonLineIntervalAt(polygon, bounds, y, centerX)
		result.MinX = max(result.MinX, interval.MinX)
		result.MaxX = min(result.MaxX, interval.MaxX)
	}
	if math.IsInf(result.MinX, 0) || math.IsInf(result.MaxX, 0) || result.MaxX < result.MinX {
		return radialInterval{MinX: centerX, MaxX: centerX}
	}
	return result
}

func betterSectorFlowScore(a, b sectorFlowScore) bool {
	if a.contained != b.contained {
		return a.contained > b.contained
	}
	if math.Abs(a.overflow-b.overflow) > 0.001 {
		return a.overflow < b.overflow
	}
	if math.Abs(a.distance-b.distance) > 0.001 {
		return a.distance < b.distance
	}
	if math.Abs(a.height-b.height) > 0.001 {
		return a.height < b.height
	}
	if math.Abs(a.unused-b.unused) > 0.001 {
		return a.unused < b.unused
	}
	return a.signature < b.signature
}

func sectorFlowRowsSignature(rows []sectorFlowRow) string {
	var result strings.Builder
	for i, row := range rows {
		if i > 0 {
			result.WriteByte(',')
		}
		result.WriteString(strconv.Itoa(row.end))
	}
	return result.String()
}
