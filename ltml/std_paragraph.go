// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/rowland/leadtype/rich_text"
	"github.com/rowland/leadtype/wordbreaking"
)

type StdParagraph struct {
	StdContainer
	textPieces         []textPiece
	richText           *rich_text.RichText
	textFill           *BrushStyle
	bullets            []*BulletStyle
	overflowEnabled    bool
	overflowExplicit   bool
	orphans            int
	widows             int
	splitLines         []*rich_text.RichText
	suppressBullet     bool
	continuationIndent float64
	angle              float64
	angleSet           bool
	facing             sectorFacing
	facingSet          bool
}

func (p *StdParagraph) AddText(text string) {
	p.AddTextWithFont(text, p.explicitFont())
}

func (p *StdParagraph) AddTextWithFont(text string, font *FontStyle) {
	addNormalizedTextPiece(&p.textPieces, &p.richText, text, font, normalizeXMLText)
}

func (p *StdParagraph) AddInlineWithFont(content inlineText, font *FontStyle) {
	addInlineTextPiece(&p.textPieces, &p.richText, content, font)
}

func normalizeXMLText(text string) string {
	var b strings.Builder
	b.Grow(len(text))
	inSpace := false
	for _, r := range text {
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

func (p *StdParagraph) BeforePrint(w Writer) error {
	// fmt.Printf("Printing %s\n", p)
	return nil
}

func (p *StdParagraph) LayoutWidget(Writer) error {
	// Paragraphs are text leaf widgets. Inline descendants contribute rich text
	// rather than participating in container child layout.
	return nil
}

func (p *StdParagraph) curvedInSector() bool {
	_, inSector := p.Container().(*StdSector)
	return inSector && (!p.angleSet || p.angle != 0)
}

func (p *StdParagraph) horizontalInSector() bool {
	_, inSector := p.Container().(*StdSector)
	return inSector && p.angleSet && p.angle == 0
}

func (p *StdParagraph) sectorTextFacing() sectorFacing {
	if p.facingSet {
		return p.facing
	}
	return sectorFacingAuto
}

func (p *StdParagraph) OriginX() OriginX {
	if p.curvedInSector() && p.StdWidget.originX == OriginXUnspecified {
		switch p.ParagraphStyle().ResolvedTextAlign(p) {
		case HAlignCenter:
			return OriginXCenter
		case HAlignRight:
			return OriginXEnd
		default:
			return OriginXStart
		}
	}
	return p.StdWidget.originX
}

func (p *StdParagraph) OriginY() OriginY {
	if p.curvedInSector() && p.StdWidget.originY == OriginYUnspecified {
		return OriginYMiddle
	}
	return p.StdWidget.originY
}

func (p *StdParagraph) paintWithTransform(w Writer, fn func() error) error {
	if p.curvedInSector() {
		return fn()
	}
	return p.StdWidget.paintWithTransform(w, fn)
}

func (p *StdParagraph) PaintBackground(w Writer) error {
	if p.curvedInSector() {
		return nil
	}
	return p.StdWidget.PaintBackground(w)
}

func (p *StdParagraph) DrawBorder(w Writer) error {
	if p.curvedInSector() {
		return nil
	}
	return p.StdWidget.DrawBorder(w)
}

func (p *StdParagraph) Bullet() *BulletStyle {
	if bullets := p.Bullets(); len(bullets) > 0 {
		return bullets[0]
	}
	return nil
}

func (p *StdParagraph) Bullets() []*BulletStyle {
	if len(p.bullets) > 0 {
		return p.bullets
	}
	if ps := p.ParagraphStyle(); ps != nil {
		return ps.Bullets()
	}
	return nil
}

func (p *StdParagraph) bulletWidth() float64 {
	width := 0.0
	for _, b := range p.Bullets() {
		width += b.Width()
	}
	return width
}

func (p *StdParagraph) DrawContent(w Writer) error {
	return withWidgetRoleAccessibility(w, &p.StdWidget, "P", p.AccessibilityText(), func() error {
		if provider, ok := p.container.(sectorParagraphLayoutProvider); ok {
			para := p.Lines(w, p.lineWidth())
			if len(para) == 0 {
				return nil
			}
			return provider.drawSectorParagraph(p, w, provider.sectorParagraphLayoutFor(p, w))
		}
		para := p.Lines(w, p.lineWidth())
		if len(para) == 0 {
			return nil
		}
		indent := p.textIndent()
		textX := p.textStartX(indent)
		textHeight := p.textContentHeightForLines(para, w)
		baselineY := ContentTop(p) + para[0].Ascent()
		w.MoveTo(textX, baselineY)
		if bullets := p.Bullets(); len(bullets) > 0 && !p.suppressBullet {
			y := baselineY
			if err := withAccessibilityArtifact(w, func() error {
				return p.drawBullets(w, bullets, para[0], y, textHeight)
			}); err != nil {
				return err
			}
			w.MoveTo(textX, y)
		}
		if p.textFill != nil {
			return p.paintTextFill(w, para, textX, baselineY, ContentWidth(p)-indent)
		}
		w.PrintParagraph(para, paragraphTextFillOptions(p))
		return nil
	})
}

func (p *StdParagraph) Lines(w Writer, width float64) []*rich_text.RichText {
	if p.splitLines != nil {
		return p.splitLines
	}
	if provider, ok := p.container.(sectorParagraphLayoutProvider); ok {
		return provider.sectorParagraphLayoutFor(p, w).lines
	}
	// A leader only changes how the final visible line is composed. If no
	// leader placeholder is present, this falls straight through to ordinary
	// rich-text wrapping.
	if lines, ok := prepareLeaderLines(w, p, p.textPieces, width, true); ok {
		return lines
	}
	rt := p.RichText(w)
	flags := make([]wordbreaking.Flags, rt.Len())
	wordbreaking.MarkRuneAttributes(rt.String(), flags)
	return rt.WrapToWidth(width, flags, false)
}

func (p *StdParagraph) PreferredHeight(w Writer) (float64, error) {
	if profiler := profilerForWidget(w, p); profiler != nil {
		defer beginWidgetProfileSpan(profiler, "preferred_height", p).End()
	}
	if provider, ok := p.container.(sectorParagraphLayoutProvider); ok {
		layout := provider.sectorParagraphLayoutFor(p, w)
		if layout.err != nil {
			return 0, layout.err
		}
		if p.curvedInSector() {
			return layout.total, nil
		}
		if p.height != 0 {
			return float64(p.height), nil
		}
		return NonContentHeight(p) + layout.total, nil
	}
	if p.height != 0 {
		return float64(p.height), nil
	}
	return p.heightForLines(p.Lines(w, p.lineWidth()), w), nil
}

func (p *StdParagraph) AccessibilityText() string {
	return resolvedTextPieces(p.textPieces, documentForContainer(p))
}

func (p *StdParagraph) PreferredWidth(w Writer) (float64, error) {
	if profiler := profilerForWidget(w, p); profiler != nil {
		defer beginWidgetProfileSpan(profiler, "preferred_width", p).End()
	}
	if p.width != 0 {
		return float64(p.width), nil
	}
	if lines, ok := prepareLeaderLines(w, p, p.textPieces, ContentWidth(p.container), true); ok {
		return lineMaxWidth(lines) + p.bulletWidth() + NonContentWidth(p) + 1, nil
	}
	return p.RichText(w).Width() + p.bulletWidth() + NonContentWidth(p) + 1, nil
}

func (p *StdParagraph) RichText(w Writer) *rich_text.RichText {
	return richTextForTextPieces(w, p, p.textPieces, &p.richText, p.Font())
}

func (p *StdParagraph) SetAttrs(attrs map[string]string) {
	p.StdContainer.SetAttrs(attrs)
	SetBrushStyle(&p.textFill, "text-fill", attrs, p.scope, p.Units())
	p.orphans = 2
	p.widows = 2
	SetParagraphStyle(&p.paragraphStyle, "style", attrs, p.scope, p)
	if bullet, ok := attrs["bullet"]; ok {
		p.bullets = bulletStylesFor(bullet, p.scope)
	}
	if overflow, ok := attrs["overflow"]; ok {
		p.overflowExplicit = true
		p.overflowEnabled = strings.TrimSpace(overflow) == "true"
	}
	if orphans, ok := attrs["orphans"]; ok {
		if value, err := strconv.Atoi(orphans); err == nil {
			p.orphans = value
		}
	}
	if widows, ok := attrs["widows"]; ok {
		if value, err := strconv.Atoi(widows); err == nil {
			p.widows = value
		}
	}
	if angle, ok := attrs["angle"]; ok {
		if value, err := strconv.ParseFloat(strings.TrimSpace(angle), 64); err == nil {
			p.angle = value
			p.angleSet = true
		}
	}
	if facing, ok := attrs["facing"]; ok {
		p.facingSet = true
		p.facing = sectorFacingAuto
		switch facing {
		case "upright":
			p.facing = sectorFacingUpright
		case "upside-down":
			p.facing = sectorFacingUpsideDown
		}
	}
	if sector, ok := p.Container().(*StdSector); ok {
		sector.invalidateParagraphLayout(p)
	}
}

func (p *StdParagraph) paintTextFill(w Writer, para []*rich_text.RichText, startX, startY, width float64) error {
	x, y, fillWidth, fillHeight := p.backgroundRect()
	align := p.ParagraphStyle().ResolvedTextAlign(p)
	currentY := startY
	for i, line := range para {
		xOffset, clipped := paragraphAlignedLine(line, width, align)
		w.MoveTo(startX+xOffset, currentY)
		var paintErr error
		if err := w.ClipRichText(clipped, func() {
			paintErr = p.PaintBrushInRect(w, p.textFill, x, y, fillWidth, fillHeight)
		}); err != nil {
			return err
		}
		if paintErr != nil {
			return paintErr
		}
		if i+1 < len(para) {
			currentY += para[i+1].Leading() * w.LineSpacing()
		}
	}
	return nil
}

func (p *StdParagraph) SplitForHeight(avail float64, w Writer) (*SplitResult, error) {
	if profiler := profilerForWidget(w, p); profiler != nil {
		defer beginWidgetProfileSpan(profiler, "split", p).End()
	}
	if !p.overflowAllowed() {
		return nil, nil
	}
	if _, ok := detectLeaderPieces(p.textPieces); ok {
		return nil, nil
	}
	lines := p.Lines(w, p.lineWidth())
	if len(lines) < 2 {
		return nil, nil
	}
	avail -= NonContentHeight(p)
	if avail <= 0 {
		return nil, nil
	}
	fit := 0
	for i := 1; i <= len(lines); i++ {
		if p.contentHeightForLines(lines[:i], w) <= avail {
			fit = i
			continue
		}
		break
	}
	if fit == len(lines) {
		return nil, nil
	}
	fit = min(fit, len(lines)-p.widowCount())
	if fit < p.orphanCount() {
		return nil, nil
	}
	head := p.cloneForSplit(lines[:fit], p.suppressBullet, p.continuationIndent)
	tail := p.cloneForSplit(lines[fit:], true, p.textIndent())
	return &SplitResult{Head: head, Tail: tail}, nil
}

func (p *StdParagraph) SplitEnabled() bool {
	return p.overflowAllowed()
}

func (p *StdParagraph) overflowAllowed() bool {
	if p.overflowExplicit {
		return p.overflowEnabled
	}
	return true
}

func (p *StdParagraph) cloneForSplit(lines []*rich_text.RichText, suppressBullet bool, continuationIndent float64) *StdParagraph {
	p.AccessibilityLogicalID()
	clone := *p
	clone.splitLines = append([]*rich_text.RichText(nil), lines...)
	clone.suppressBullet = suppressBullet
	clone.continuationIndent = continuationIndent
	clone.ClearResolvedWidth()
	clone.ClearResolvedHeight()
	clone.richText = nil
	clone.printed = false
	clone.invisible = false
	clone.disabled = false
	clone.path = ""
	return &clone
}

func (p *StdParagraph) textIndent() float64 {
	if p.suppressBullet {
		return p.continuationIndent
	}
	return p.bulletWidth()
}

func (p *StdParagraph) textStartX(indent float64) float64 {
	if IsRTL(p) {
		return ContentLeft(p)
	}
	return ContentLeft(p) + indent
}

func (p *StdParagraph) lineWidth() float64 {
	if p.Width() == 0 {
		return ContentWidth(p.container) - p.textIndent() - NonContentWidth(p)
	}
	return ContentWidth(p) - p.textIndent()
}

func (p *StdParagraph) heightForLines(lines []*rich_text.RichText, w Writer) float64 {
	return NonContentHeight(p) + p.contentHeightForLines(lines, w)
}

func (p *StdParagraph) contentHeightForLines(lines []*rich_text.RichText, w Writer) float64 {
	textHeight := p.textContentHeightForLines(lines, w)
	if len(lines) == 0 {
		return textHeight
	}
	if bulletHeight := p.bulletBoxHeightForLines(w, lines, textHeight); bulletHeight > textHeight {
		return bulletHeight
	}
	return textHeight
}

func (p *StdParagraph) textContentHeightForLines(lines []*rich_text.RichText, w Writer) float64 {
	height := 0.0
	for _, line := range lines {
		height += line.Leading() * w.LineSpacing()
	}
	if len(lines) > 0 {
		height -= lines[len(lines)-1].Height() * (w.LineSpacing() - 1)
		height -= lines[len(lines)-1].LineGap()
	}
	return height
}

func (p *StdParagraph) orphanCount() int {
	if p.orphans < 1 {
		return 1
	}
	return p.orphans
}

func (p *StdParagraph) widowCount() int {
	if p.widows < 1 {
		return 1
	}
	return p.widows
}

func (p *StdParagraph) String() string {
	return fmt.Sprintf("StdParagraph %s", &p.StdContainer)
}

func init() {
	registerTag(DefaultSpace, "p", func() any { return &StdParagraph{} })
}

var _ Container = (*StdParagraph)(nil)
var _ HasAttrs = (*StdParagraph)(nil)
var _ Identifier = (*StdParagraph)(nil)
var _ Printer = (*StdParagraph)(nil)
var _ WantsContainer = (*StdParagraph)(nil)
