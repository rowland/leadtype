// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import (
	"errors"
	"math"

	"github.com/rowland/leadtype/codepage"
	"github.com/rowland/leadtype/colors"
	"github.com/rowland/leadtype/font"
	"github.com/rowland/leadtype/rich_text"
	"github.com/rowland/leadtype/shaping"
)

var errCurvedTextOutsidePath = errors.New("curved text extends beyond path bounds")

type CurvedTextHAlign uint8

const (
	CurvedTextAlignLeft CurvedTextHAlign = iota
	CurvedTextAlignCenter
	CurvedTextAlignRight
)

type CurvedTextDirection uint8

const (
	CurvedTextDirectionAuto CurvedTextDirection = iota
	CurvedTextClockwise
	CurvedTextCounterClockwise
)

type CurvedTextOrientation uint8

const (
	CurvedTextOrientationOutside CurvedTextOrientation = iota
	CurvedTextOrientationInside
)

type CurvedTextFacing uint8

const (
	CurvedTextFacingAuto CurvedTextFacing = iota
	CurvedTextFacingUpright
	CurvedTextFacingUpsideDown
)

type CurvedTextOptions struct {
	Align       CurvedTextHAlign
	VAlign      VerticalTextAlign
	Direction   CurvedTextDirection
	Orientation CurvedTextOrientation
	Facing      CurvedTextFacing
}

type curvedTextAlign uint8

const (
	curvedTextAlignLeft curvedTextAlign = iota
	curvedTextAlignCenter
	curvedTextAlignRight
)

// curvedTextGlyph is the minimal internal description of one glyph or shaped
// cluster to be placed on a path.
type curvedTextGlyph struct {
	GlyphID uint16
	Text    string
	Advance float64
}

type curvedTextRenderGlyph struct {
	curvedTextGlyph
	font       *font.Font
	fontSize   float64
	color      colors.Color
	codepage   codepage.CodepageIndex
	mappedText []rune
}

// curvedTextGlyphPlacement records one glyph or shaped cluster positioned along
// a path. Point, Tangent, and Normal are sampled at the glyph midpoint.
type curvedTextGlyphPlacement struct {
	GlyphID        uint16
	Text           string
	Advance        float64
	PathDistance   float64
	SampleDistance float64
	Point          Location
	Tangent        Location
	Normal         Location
	TangentAngle   float64
	BaselineOffset float64
}

func (p curvedTextGlyphPlacement) baselinePoint() Location {
	return Location{
		X: p.Point.X + (p.Normal.X * p.BaselineOffset),
		Y: p.Point.Y + (p.Normal.Y * p.BaselineOffset),
	}
}

// originPoint returns the baseline origin for PDF text emission, treating the
// sampled point as the midpoint of the glyph advance.
func (p curvedTextGlyphPlacement) originPoint() Location {
	base := p.baselinePoint()
	halfAdvance := p.Advance / 2.0
	return Location{
		X: base.X - (p.Tangent.X * halfAdvance),
		Y: base.Y - (p.Tangent.Y * halfAdvance),
	}
}

func (p curvedTextGlyphPlacement) originPointWithAxes(xAxis, offsetAxis Location, offset float64) Location {
	base := Location{
		X: p.Point.X + (offsetAxis.X * offset),
		Y: p.Point.Y + (offsetAxis.Y * offset),
	}
	halfAdvance := p.Advance / 2.0
	return Location{
		X: base.X - (xAxis.X * halfAdvance),
		Y: base.Y - (xAxis.Y * halfAdvance),
	}
}

type curvedTextPath interface {
	closed() bool
	length() float64
	pointAt(distance float64) Location
	tangentAt(distance float64) Location
	normalAt(distance float64) Location
}

type curvedTextCirclePath struct {
	center     Location
	radius     float64
	startAngle float64
	clockwise  bool
}

func (p curvedTextCirclePath) closed() bool {
	return true
}

func (p curvedTextCirclePath) length() float64 {
	return 2 * math.Pi * p.radius
}

func (p curvedTextCirclePath) pointAt(distance float64) Location {
	theta := p.theta(distance)
	return Location{
		X: p.center.X + (p.radius * math.Cos(theta)),
		Y: p.center.Y + (p.radius * math.Sin(theta)),
	}
}

func (p curvedTextCirclePath) tangentAt(distance float64) Location {
	theta := p.theta(distance)
	if p.clockwise {
		return Location{X: math.Sin(theta), Y: -math.Cos(theta)}
	}
	return Location{X: -math.Sin(theta), Y: math.Cos(theta)}
}

func (p curvedTextCirclePath) normalAt(distance float64) Location {
	theta := p.theta(distance)
	return Location{X: math.Cos(theta), Y: math.Sin(theta)}
}

func (p curvedTextCirclePath) theta(distance float64) float64 {
	length := p.length()
	if p.radius == 0 || length == 0 {
		return p.startAngle * math.Pi / 180.0
	}
	distance = math.Mod(distance, length)
	if distance < 0 {
		distance += length
	}
	theta := (p.startAngle * math.Pi / 180.0) + (distance / p.radius)
	if p.clockwise {
		return (p.startAngle * math.Pi / 180.0) - (distance / p.radius)
	}
	return theta
}

func placeCurvedTextGlyphs(path curvedTextPath, glyphs []curvedTextGlyph, anchorDistance float64, align curvedTextAlign, baselineOffset float64) ([]curvedTextGlyphPlacement, error) {
	if len(glyphs) == 0 {
		return nil, nil
	}

	totalAdvance := 0.0
	for _, glyph := range glyphs {
		totalAdvance += glyph.Advance
	}

	startDistance := anchorDistance
	switch align {
	case curvedTextAlignCenter:
		startDistance -= totalAdvance / 2.0
	case curvedTextAlignRight:
		startDistance -= totalAdvance
	}

	pathLength := path.length()
	placements := make([]curvedTextGlyphPlacement, 0, len(glyphs))
	offset := 0.0
	for _, glyph := range glyphs {
		sampleDistance := startDistance + offset + (glyph.Advance / 2.0)
		if !path.closed() && (sampleDistance < 0 || sampleDistance > pathLength) {
			return nil, errCurvedTextOutsidePath
		}
		tangent := path.tangentAt(sampleDistance)
		placements = append(placements, curvedTextGlyphPlacement{
			GlyphID:        glyph.GlyphID,
			Text:           glyph.Text,
			Advance:        glyph.Advance,
			PathDistance:   startDistance + offset,
			SampleDistance: sampleDistance,
			Point:          path.pointAt(sampleDistance),
			Tangent:        tangent,
			Normal:         path.normalAt(sampleDistance),
			TangentAngle:   math.Atan2(tangent.Y, tangent.X) * 180.0 / math.Pi,
			BaselineOffset: baselineOffset,
		})
		offset += glyph.Advance
	}

	return placements, nil
}

func (opts CurvedTextOptions) curvedAlign() curvedTextAlign {
	switch opts.Align {
	case CurvedTextAlignCenter:
		return curvedTextAlignCenter
	case CurvedTextAlignRight:
		return curvedTextAlignRight
	default:
		return curvedTextAlignLeft
	}
}

func (opts CurvedTextOptions) clockwise() bool {
	return opts.Direction != CurvedTextCounterClockwise
}

func (opts CurvedTextOptions) resolveCircleAuto(glyphs []curvedTextRenderGlyph, center Location, radius, startAngle float64) (CurvedTextDirection, CurvedTextFacing) {
	if opts.Direction != CurvedTextDirectionAuto && opts.Facing != CurvedTextFacingAuto {
		return opts.Direction, opts.Facing
	}

	inputs := make([]curvedTextGlyph, 0, len(glyphs))
	for _, glyph := range glyphs {
		inputs = append(inputs, glyph.curvedTextGlyph)
	}
	placements, err := placeCurvedTextGlyphs(curvedTextCirclePath{
		center:     center,
		radius:     radius,
		startAngle: startAngle,
		clockwise:  true,
	}, inputs, 0, opts.curvedAlign(), 0)
	if err != nil {
		return opts.Direction, opts.Facing
	}

	totalAdvance := 0.0
	weightedY := 0.0
	for _, placement := range placements {
		totalAdvance += placement.Advance
		weightedY += placement.Point.Y * placement.Advance
	}
	if totalAdvance == 0 {
		return opts.Direction, opts.Facing
	}
	// Keep top/left/right spans upright by default and flip only when the text
	// span's weighted center is clearly below the circle midline.
	centroidY := weightedY / totalAdvance
	mostlyBelow := centroidY < (center.Y - (radius * 0.05))

	direction := opts.Direction
	if direction == CurvedTextDirectionAuto {
		if mostlyBelow {
			direction = CurvedTextCounterClockwise
		} else {
			direction = CurvedTextClockwise
		}
	}

	facing := opts.Facing
	if facing == CurvedTextFacingAuto {
		if mostlyBelow {
			facing = CurvedTextFacingUpsideDown
		} else {
			facing = CurvedTextFacingUpright
		}
	}

	return direction, facing
}

func orientCurvedTextPlacement(p curvedTextGlyphPlacement, orientation CurvedTextOrientation) curvedTextGlyphPlacement {
	if orientation == CurvedTextOrientationInside {
		p.Normal.X = -p.Normal.X
		p.Normal.Y = -p.Normal.Y
	}
	p.TangentAngle = math.Atan2(p.Tangent.Y, p.Tangent.X) * 180.0 / math.Pi
	return p
}

func curvedTextFrame(tangent, desiredNormal Location, facing CurvedTextFacing) (xAxis, yAxis Location) {
	xAxis = tangent
	yAxis = Location{X: -tangent.Y, Y: tangent.X}
	if (yAxis.X*desiredNormal.X)+(yAxis.Y*desiredNormal.Y) < 0 {
		xAxis.X = -xAxis.X
		xAxis.Y = -xAxis.Y
		yAxis.X = -yAxis.X
		yAxis.Y = -yAxis.Y
	}
	if facing == CurvedTextFacingUpsideDown {
		xAxis.X = -xAxis.X
		xAxis.Y = -xAxis.Y
		yAxis.X = -yAxis.X
		yAxis.Y = -yAxis.Y
	}
	return xAxis, yAxis
}

func (pw *PageWriter) DrawTextOnCircle(text string, x, y, r, startAngle float64, opts CurvedTextOptions) error {
	piece, err := pw.richTextForString(text)
	if err != nil {
		return err
	}
	return pw.DrawRichTextOnCircle(piece, x, y, r, startAngle, opts)
}

func (pw *PageWriter) DrawRichTextOnCircle(text *rich_text.RichText, x, y, r, startAngle float64, opts CurvedTextOptions) error {
	pw.flushText()
	if text == nil {
		return nil
	}

	glyphs, err := pw.curvedTextGlyphsForRichText(text)
	if err != nil {
		return err
	}
	if len(glyphs) == 0 {
		return nil
	}

	xpts := pw.units.toPts(x)
	ypts := pw.translate(pw.units.toPts(y))
	rpts := pw.units.toPts(r)
	resolvedDirection, resolvedFacing := opts.resolveCircleAuto(glyphs, Location{xpts, ypts}, rpts, startAngle)
	path := curvedTextCirclePath{
		center:     Location{xpts, ypts},
		radius:     rpts,
		startAngle: startAngle,
		clockwise:  resolvedDirection != CurvedTextCounterClockwise,
	}

	inputs := make([]curvedTextGlyph, 0, len(glyphs))
	for _, glyph := range glyphs {
		inputs = append(inputs, glyph.curvedTextGlyph)
	}
	placements, err := placeCurvedTextGlyphs(path, inputs, 0, opts.curvedAlign(), 0)
	if err != nil {
		return err
	}
	sharedBaselineOffset, useSharedBaseline := curvedTextSharedBaselineOffset(text, opts.VAlign)

	savedFontColor := pw.fontColor
	savedFontKey := pw.fontKey
	savedFontSize := pw.fontSize
	savedCharSpacing := pw.charSpacing
	savedWordSpacing := pw.wordSpacing
	savedVTextAlign := pw.vTextAlign
	savedVTextAlignPts := pw.vTextAlignPts

	defer func() {
		pw.fontColor = savedFontColor
		pw.fontKey = savedFontKey
		pw.fontSize = savedFontSize
		pw.charSpacing = savedCharSpacing
		pw.wordSpacing = savedWordSpacing
		pw.vTextAlign = savedVTextAlign
		pw.vTextAlignPts = savedVTextAlignPts
	}()

	pw.startText()
	defer func() {
		if pw.inText {
			pw.endText()
		}
	}()
	for i, placement := range placements {
		glyph := glyphs[i]
		if useSharedBaseline {
			placement.BaselineOffset = sharedBaselineOffset
		} else {
			placement.BaselineOffset = textRiseForFont(glyph.font, glyph.fontSize, opts.VAlign)
		}
		placement = orientCurvedTextPlacement(placement, opts.Orientation)
		if err := pw.showCurvedTextGlyph(glyph, placement, resolvedFacing); err != nil {
			return err
		}
	}
	return nil
}

func curvedTextSharedBaselineOffset(text *rich_text.RichText, align VerticalTextAlign) (float64, bool) {
	return sharedRichTextRise(text, align)
}

func (dw *DocWriter) DrawTextOnCircle(text string, x, y, r, startAngle float64, opts CurvedTextOptions) error {
	return dw.CurPage().DrawTextOnCircle(text, x, y, r, startAngle, opts)
}

func (dw *DocWriter) DrawRichTextOnCircle(text *rich_text.RichText, x, y, r, startAngle float64, opts CurvedTextOptions) error {
	return dw.CurPage().DrawRichTextOnCircle(text, x, y, r, startAngle, opts)
}

func (pw *PageWriter) curvedTextGlyphsForRichText(text *rich_text.RichText) ([]curvedTextRenderGlyph, error) {
	if text == nil {
		return nil, nil
	}

	merged := text.Merge()
	displayPieces := bidiDisplayPieces(merged, pw.bidiBaseDirection(merged.String()))
	glyphs := []curvedTextRenderGlyph{}
	for _, p := range displayPieces {
		if p == nil || p.Font == nil {
			continue
		}
		if pw.supportsArabicShaping &&
			p.Font.SupportsArabic() &&
			shaping.ContainsArabic(p.Text) {
			runes := []rune(p.Text)
			shaped, err := p.Font.Shaper.Shape(runes, p.Font, float32(p.FontSize))
			if err == nil && shaped != nil {
				assignments := shapedGlyphRuneAssignments(shaped, runes)
				for i, gp := range shaped {
					seq := assignments[i]
					glyphs = append(glyphs, curvedTextRenderGlyph{
						curvedTextGlyph: curvedTextGlyph{
							GlyphID: gp.GlyphID,
							Text:    string(seq),
							Advance: float64(gp.XAdvance) / 64.0,
						},
						font:       p.Font,
						fontSize:   p.FontSize,
						color:      p.Color,
						mappedText: append([]rune(nil), seq...),
					})
				}
				continue
			}
		}

		scale := p.FontSize * 0.001
		if upm := p.Font.UnitsPerEm(); upm > 0 {
			scale = p.FontSize / float64(upm)
		}
		for _, r := range p.Text {
			advanceWidth, _ := p.Font.AdvanceWidth(r)
			advance := (scale * float64(advanceWidth)) + p.CharSpacing
			if r == ' ' {
				advance += p.WordSpacing
			}
			cpi, _ := codepage.IndexForCodepoint(r)
			glyphs = append(glyphs, curvedTextRenderGlyph{
				curvedTextGlyph: curvedTextGlyph{
					GlyphID: p.Font.GlyphIndex(r),
					Text:    string(r),
					Advance: advance,
				},
				font:       p.Font,
				fontSize:   p.FontSize,
				color:      p.Color,
				codepage:   cpi,
				mappedText: []rune{r},
			})
		}
	}

	return glyphs, nil
}

func (pw *PageWriter) showCurvedTextGlyph(glyph curvedTextRenderGlyph, placement curvedTextGlyphPlacement, facing CurvedTextFacing) error {
	pw.SetFontColor(glyph.color)
	pw.checkSetFontColor()
	pw.SetFontSize(glyph.fontSize)

	origin := placement.originPoint()
	if glyph.font.SubType() == "TrueType" {
		pw.fontKey = pw.dw.fontKeyUnicode(glyph.font)
		pw.checkSetFont()
		pw.charSpacing = 0
		pw.wordSpacing = 0
		pw.checkSetSpacing()
		pw.checkSetVTextAlign(false)

		code := glyph.GlyphID
		var gr *glyphRecorder
		gr = pw.dw.glyphRecorders[glyph.font.PostScriptName()]
		if gr != nil {
			if len(glyph.mappedText) > 0 {
				code = gr.recordRunes(glyph.GlyphID, glyph.mappedText)
			} else {
				code = gr.recordEmpty(glyph.GlyphID)
			}
		}
		buf := []byte{byte(code >> 8), byte(code & 0xFF)}
		xAxis, yAxis := curvedTextFrame(placement.Tangent, placement.Normal, facing)
		origin = placement.originPointWithAxes(xAxis, yAxis, placement.BaselineOffset)
		pw.tw.setMatrix(xAxis.X, xAxis.Y, yAxis.X, yAxis.Y, origin.X, origin.Y)
		closeAliasText := gr != nil && gr.requiresActualText(code, glyph.mappedText) && pw.beginActualTextContent(string(glyph.mappedText))
		pw.tw.showHex(buf)
		if closeAliasText {
			pw.mw.endMarkedContent()
		}
		return nil
	}

	pw.fontKey = pw.dw.fontKey(glyph.font, glyph.codepage)
	pw.checkSetFont()
	pw.charSpacing = 0
	pw.wordSpacing = 0
	pw.checkSetSpacing()
	pw.checkSetVTextAlign(false)

	if len(glyph.Text) == 0 {
		return nil
	}
	runes := []rune(glyph.Text)
	ch, _ := glyph.codepage.Codepage().CharForCodepoint(runes[0])
	xAxis, yAxis := curvedTextFrame(placement.Tangent, placement.Normal, facing)
	origin = placement.originPointWithAxes(xAxis, yAxis, placement.BaselineOffset)
	pw.tw.setMatrix(xAxis.X, xAxis.Y, yAxis.X, yAxis.Y, origin.X, origin.Y)
	pw.tw.show([]byte{byte(ch)})
	return nil
}
