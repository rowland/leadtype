// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/rowland/leadtype/rich_text"
)

type StdLabel struct {
	StdContainer
	textPieces    []textPiece
	richText      *rich_text.RichText
	textFill      *BrushStyle
	shrinkToFit   bool
	angle         float64
	angleSet      bool
	facing        sectorFacing
	facingSet     bool
	textAlign     HAlign
	textAlignSet  bool
	textVAlign    VAlign
	textVAlignSet bool
}

func (l *StdLabel) AddText(text string) {
	l.AddTextWithFont(text, l.explicitFont())
}

func (l *StdLabel) AddTextWithFont(text string, font *FontStyle) {
	addNormalizedTextPiece(&l.textPieces, &l.richText, text, font, normalizeLabelXMLText)
}

func (l *StdLabel) AddInlineWithFont(content inlineText, font *FontStyle) {
	addInlineTextPiece(&l.textPieces, &l.richText, content, font)
}

func normalizeLabelXMLText(text string) string {
	var b strings.Builder
	b.Grow(len(text))
	inSpace := false
	for _, r := range text {
		if r == '\n' || r == '\r' {
			continue
		}
		if r == '\t' {
			r = ' '
		}
		if unicode.IsSpace(r) {
			if !inSpace {
				b.WriteByte(' ')
				inSpace = true
			}
			continue
		}
		b.WriteRune(r)
		inSpace = false
	}
	return b.String()
}

func (l *StdLabel) LayoutWidget(w Writer) error {
	// Labels are leaf widgets even though they embed StdContainer to collect
	// inline children like <span> and <pageno>.
	if sector, ok := l.Container().(*StdSector); ok {
		return sector.layoutSectorLabel(l, w)
	}
	return nil
}

func (l *StdLabel) DrawContent(w Writer) error {
	if sector, ok := l.Container().(*StdSector); ok {
		return sector.drawSectorLabel(l, w)
	}
	return l.drawBoxLabelContent(w, l.angle)
}

func (l *StdLabel) drawBoxLabelContent(w Writer, angle float64) error {
	return withWidgetRoleAccessibility(w, &l.StdWidget, "P", l.AccessibilityText(), func() error {
		rt := l.layoutRichText(w)
		if rt.Len() == 0 {
			return nil
		}
		applyContainerFont(w, l)
		anchorX, anchorY := l.textAnchor(rt)
		startX := anchorX - l.textAnchorOffset(rt)
		if l.textFill != nil {
			x, y, width, height := l.backgroundRect()
			paintFill := func() error {
				w.MoveTo(startX, anchorY)
				var paintErr error
				if err := w.ClipRichText(rt, func() {
					paintErr = l.PaintBrushInRect(w, l.textFill, x, y, width, height)
				}); err != nil {
					return err
				}
				if paintErr != nil {
					return paintErr
				}
				return nil
			}
			if angle == 0 {
				return paintFill()
			}
			var paintErr error
			if err := w.Rotate(angle, anchorX, anchorY, func() {
				paintErr = paintFill()
			}); err != nil {
				return err
			}
			return paintErr
		}
		if angle == 0 {
			w.MoveTo(startX, anchorY)
			w.PrintRichText(rt)
			return nil
		}
		if err := w.Rotate(angle, anchorX, anchorY, func() {
			w.MoveTo(startX, anchorY)
			w.PrintRichText(rt)
		}); err != nil {
			return err
		}
		return nil
	})
}

func (l *StdLabel) PreferredHeight(w Writer) (float64, error) {
	if l.height != 0 {
		return float64(l.height), nil
	}
	rt := l.layoutRichText(w)
	if rt.Len() == 0 {
		return effectiveFontSizeForContainer(l)*w.LineSpacing() + NonContentHeight(l), nil
	}
	return rt.Leading()*w.LineSpacing() + NonContentHeight(l), nil
}

func (l *StdLabel) PreferredWidth(w Writer) (float64, error) {
	if l.width != 0 {
		return float64(l.width), nil
	}
	rt := l.layoutRichText(w)
	return rt.Width() + NonContentWidth(l), nil
}

func (l *StdLabel) AccessibilityText() string {
	return resolvedTextPieces(l.textPieces, documentForContainer(l))
}

func (l *StdLabel) sectorTextAngle() (float64, bool) {
	return l.angle, l.angleSet
}

func (l *StdLabel) sectorTextAlign() HAlign {
	if l.textAlignSet {
		return resolveTextAlign(l.textAlign, l)
	}
	return HAlignCenter
}

func (l *StdLabel) sectorTextVAlign() VAlign {
	if l.textVAlignSet {
		return l.textVAlign
	}
	return VAlignMiddle
}

func (l *StdLabel) sectorTextFacing() sectorFacing {
	if l.facingSet {
		return l.facing
	}
	return sectorFacingAuto
}

func (l *StdLabel) Left() float64 {
	if layout := l.cachedSectorLayout(); layout != nil {
		return layout.boxLeft
	}
	return l.StdWidget.Left()
}

func (l *StdLabel) Right() float64 {
	if layout := l.cachedSectorLayout(); layout != nil {
		return layout.boxLeft + layout.boxWidth
	}
	return l.StdWidget.Right()
}

func (l *StdLabel) Top() float64 {
	if layout := l.cachedSectorLayout(); layout != nil {
		return layout.boxTop
	}
	return l.StdWidget.Top()
}

func (l *StdLabel) Bottom() float64 {
	if layout := l.cachedSectorLayout(); layout != nil {
		return layout.boxTop + layout.boxHeight
	}
	return l.StdWidget.Bottom()
}

func (l *StdLabel) OriginXValue() float64 {
	if l.StdWidget.originX == OriginXCustom {
		return l.StdWidget.OriginXValue()
	}
	if layout := l.cachedSectorLayout(); layout != nil {
		return layout.anchorX
	}
	return l.StdWidget.OriginXValue()
}

func (l *StdLabel) OriginYValue() float64 {
	if l.StdWidget.originY == OriginYCustom {
		return l.StdWidget.OriginYValue()
	}
	if layout := l.cachedSectorLayout(); layout != nil {
		return layout.anchorY
	}
	return l.StdWidget.OriginYValue()
}

func (l *StdLabel) paintWithTransform(w Writer, fn func() error) error {
	layout := l.cachedSectorLayout()
	if sector, ok := l.Container().(*StdSector); ok && layout == nil {
		if err := sector.layoutSectorLabel(l, w); err != nil {
			return err
		}
		layout = sector.cachedLabelLayout(l)
	}
	if layout == nil {
		return l.StdWidget.paintWithTransform(w, fn)
	}
	if !layout.straight || l.rotate == 0 {
		return fn()
	}
	var renderErr error
	if err := w.Rotate(float64(l.rotate), l.OriginXValue(), l.OriginYValue(), func() {
		renderErr = fn()
	}); err != nil {
		return err
	}
	return renderErr
}

func (l *StdLabel) PaintBackground(w Writer) error {
	layout := l.cachedSectorLayout()
	if layout == nil {
		return l.StdWidget.PaintBackground(w)
	}
	if !layout.straight || l.fill == nil {
		return nil
	}
	x, y, width, height := l.sectorBackgroundRect()
	if width <= 0 || height <= 0 {
		return nil
	}
	return l.PaintBrushInRect(w, l.fill, x, y, width, height)
}

func (l *StdLabel) DrawBorder(w Writer) error {
	layout := l.cachedSectorLayout()
	if layout == nil {
		return l.StdWidget.DrawBorder(w)
	}
	if !layout.straight {
		return nil
	}
	x1, y1, width, height := l.sectorBackgroundRect()
	return drawRectBorders(w, x1, y1, width, height, l.corners.Float64sFor(width, height),
		l.border, l.borders, l.borderSideSet)
}

func (l *StdLabel) sectorBackgroundRect() (x, y, width, height float64) {
	return l.Left() + l.MarginLeft(),
		l.Top() + l.MarginTop(),
		l.Width() - l.MarginLeft() - l.MarginRight(),
		l.Height() - l.MarginTop() - l.MarginBottom()
}

func (l *StdLabel) RichText(w Writer) *rich_text.RichText {
	return richTextForTextPieces(w, l, l.textPieces, &l.richText, l.Font())
}

func (l *StdLabel) SetAttrs(attrs map[string]string) {
	l.StdContainer.SetAttrs(attrs)
	SetBrushStyle(&l.textFill, "text-fill", attrs, l.scope, l.Units())
	if fit, ok := attrs["fit"]; ok {
		l.shrinkToFit = fit == "shrink"
	}
	if angle, ok := attrs["angle"]; ok {
		if value, err := strconv.ParseFloat(strings.TrimSpace(angle), 64); err == nil {
			l.angle = value
			l.angleSet = true
		}
	}
	if facing, ok := attrs["facing"]; ok {
		l.facingSet = true
		l.facing = sectorFacingAuto
		switch facing {
		case "upright":
			l.facing = sectorFacingUpright
		case "upside-down":
			l.facing = sectorFacingUpsideDown
		}
	}
	if textAlign, ok := attrs["text-align"]; ok {
		l.textAlignSet = true
		l.textAlign = parseTextAlign(textAlign, false)
	}
	if textVAlign, ok := attrs["text-valign"]; ok {
		l.textVAlignSet = true
		l.textVAlign = parseLabelTextVAlign(textVAlign)
	}
	if sector, ok := l.Container().(*StdSector); ok {
		sector.invalidateLabelLayout(l)
	}
}

func (l *StdLabel) cachedSectorLayout() *sectorLabelLayout {
	if sector, ok := l.Container().(*StdSector); ok {
		return sector.cachedLabelLayout(l)
	}
	return nil
}

func (l *StdLabel) fittedRichText(w Writer) *rich_text.RichText {
	rt := l.RichText(w)
	if rt == nil || rt.Len() == 0 || !l.shrinkToFit || l.width == 0 {
		return rt
	}
	availableWidth := ContentWidth(l)
	if availableWidth <= 0 {
		return rt
	}
	measuredWidth := rt.Width()
	if measuredWidth <= 0 || measuredWidth <= availableWidth {
		return rt
	}
	return rt.Scale(availableWidth/measuredWidth, 6.0)
}

func (l *StdLabel) layoutRichText(w Writer) *rich_text.RichText {
	if width := ContentWidth(l); width > 0 {
		// Leader-bearing labels are composed as a single rich-text line first.
		// The shrink-to-fit logic below is intentionally the same policy used by
		// ordinary labels; the leader path just supplies a different source line.
		if lines, ok := prepareLeaderLines(w, l, l.textPieces, width, false); ok && len(lines) > 0 {
			rt := lines[0]
			if rt == nil || rt.Len() == 0 || !l.shrinkToFit || l.width == 0 {
				return rt
			}
			availableWidth := ContentWidth(l)
			measuredWidth := rt.Width()
			if measuredWidth <= 0 || measuredWidth <= availableWidth {
				return rt
			}
			return rt.Scale(availableWidth/measuredWidth, 6.0)
		}
	}
	return l.fittedRichText(w)
}

func (l *StdLabel) textAnchorX() float64 {
	switch l.resolvedTextAlign() {
	case HAlignCenter:
		return (ContentLeft(l) + ContentRight(l)) / 2
	case HAlignRight:
		return ContentRight(l)
	default:
		return ContentLeft(l)
	}
}

func (l *StdLabel) textAnchor(rt *rich_text.RichText) (x, y float64) {
	ascent := rt.Ascent()
	descent := rt.Descent()
	textHeight := ascent - descent
	contentTop := ContentTop(l)
	contentHeight := ContentHeight(l)
	textVAlign := l.textVAlign
	if _, ok := l.Container().(*StdSector); ok {
		textVAlign = l.sectorTextVAlign()
	}
	switch textVAlign {
	case VAlignMiddle:
		y = contentTop + max((contentHeight-textHeight)/2, 0) + ascent
	case VAlignCapMiddle:
		capHeight := rt.CapHeight()
		y = contentTop + max((contentHeight-capHeight)/2, 0) + capHeight
	case VAlignBottom:
		y = ContentBottom(l) + descent
	default:
		y = contentTop + ascent
	}
	return l.textAnchorX(), y
}

func (l *StdLabel) textAnchorOffset(rt *rich_text.RichText) float64 {
	switch l.resolvedTextAlign() {
	case HAlignCenter:
		return rt.Width() / 2
	case HAlignRight:
		return rt.Width()
	default:
		return 0
	}
}

func parseLabelTextVAlign(value string) VAlign {
	switch strings.TrimSpace(value) {
	case "middle":
		return VAlignMiddle
	case "cap-middle":
		return VAlignCapMiddle
	case "bottom":
		return VAlignBottom
	case "baseline":
		return VAlignBaseline
	default:
		return VAlignTop
	}
}

func (l *StdLabel) resolvedTextAlign() HAlign {
	if _, ok := l.Container().(*StdSector); ok {
		return l.sectorTextAlign()
	}
	if l.textAlignSet {
		return resolveTextAlign(l.textAlign, l)
	}
	return resolveTextAlign(textAlignStart, l)
}

func (l *StdLabel) String() string {
	return fmt.Sprintf("StdLabel %s", &l.StdContainer)
}

func init() {
	registerTag(DefaultSpace, "label", func() any { return &StdLabel{} })
}

var _ Container = (*StdLabel)(nil)
var _ HasAttrs = (*StdLabel)(nil)
var _ HasText = (*StdLabel)(nil)
var _ Identifier = (*StdLabel)(nil)
var _ Printer = (*StdLabel)(nil)
var _ WantsContainer = (*StdLabel)(nil)
