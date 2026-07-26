// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import (
	"strings"
	"testing"

	"github.com/rowland/leadtype/afm_fonts"
	"github.com/rowland/leadtype/font"
	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/rich_text"
)

type curvedTextTestLinePath struct {
	total float64
}

func (p curvedTextTestLinePath) closed() bool {
	return false
}

func (p curvedTextTestLinePath) length() float64 {
	return p.total
}

func (p curvedTextTestLinePath) pointAt(distance float64) Location {
	return Location{X: distance, Y: 0}
}

func (p curvedTextTestLinePath) tangentAt(distance float64) Location {
	return Location{X: 1, Y: 0}
}

func (p curvedTextTestLinePath) normalAt(distance float64) Location {
	return Location{X: 0, Y: 1}
}

func TestPlaceCurvedTextGlyphs_LeftAlign(t *testing.T) {
	path := curvedTextTestLinePath{total: 200}
	glyphs := []curvedTextGlyph{
		{GlyphID: 1, Text: "A", Advance: 10},
		{GlyphID: 2, Text: "B", Advance: 20},
		{GlyphID: 3, Text: "C", Advance: 30},
	}

	placements, err := placeCurvedTextGlyphs(path, glyphs, 100, curvedTextAlignLeft, 6)
	if err != nil {
		t.Fatalf("placeCurvedTextGlyphs returned error: %v", err)
	}

	expectI(t, 3, len(placements))

	expectF(t, 100, placements[0].PathDistance)
	expectF(t, 110, placements[1].PathDistance)
	expectF(t, 130, placements[2].PathDistance)

	expectF(t, 105, placements[0].Point.X)
	expectF(t, 120, placements[1].Point.X)
	expectF(t, 145, placements[2].Point.X)

	expectF(t, 6, placements[0].BaselineOffset)
	expectF(t, 0, placements[0].TangentAngle)
}

func TestPlaceCurvedTextGlyphs_CenterAndRightAlign(t *testing.T) {
	path := curvedTextTestLinePath{total: 200}
	glyphs := []curvedTextGlyph{
		{GlyphID: 1, Text: "A", Advance: 10},
		{GlyphID: 2, Text: "B", Advance: 20},
		{GlyphID: 3, Text: "C", Advance: 30},
	}

	center, err := placeCurvedTextGlyphs(path, glyphs, 100, curvedTextAlignCenter, 0)
	if err != nil {
		t.Fatalf("center align returned error: %v", err)
	}
	expectF(t, 70, center[0].PathDistance)
	expectF(t, 80, center[1].PathDistance)
	expectF(t, 100, center[2].PathDistance)

	right, err := placeCurvedTextGlyphs(path, glyphs, 100, curvedTextAlignRight, 0)
	if err != nil {
		t.Fatalf("right align returned error: %v", err)
	}
	expectF(t, 40, right[0].PathDistance)
	expectF(t, 50, right[1].PathDistance)
	expectF(t, 70, right[2].PathDistance)
}

func TestCurvedTextGlyphPlacement_BaselineAndOriginPoint(t *testing.T) {
	placement := curvedTextGlyphPlacement{
		Advance:        20,
		Point:          Location{X: 50, Y: 10},
		Tangent:        Location{X: 1, Y: 0},
		Normal:         Location{X: 0, Y: 1},
		BaselineOffset: 6,
	}

	base := placement.baselinePoint()
	expectV(t, Location{50, 16}, base)

	origin := placement.originPoint()
	expectV(t, Location{40, 16}, origin)
}

func TestCurvedTextSharedBaselineOffset_CapMiddleUsesWholeLine(t *testing.T) {
	afm, err := afm_fonts.Default()
	if err != nil {
		t.Fatal(err)
	}
	helvetica, err := font.New("Helvetica", options.Options{}, font.FontSources{afm})
	if err != nil {
		t.Fatal(err)
	}
	courier, err := font.New("Courier", options.Options{}, font.FontSources{afm})
	if err != nil {
		t.Fatal(err)
	}
	text := (&rich_text.RichText{
		Text: "A", Font: helvetica, FontSize: 10,
	}).AddPiece(&rich_text.RichText{
		Text: "B", Font: courier, FontSize: 20,
	})

	offset, shared := curvedTextSharedBaselineOffset(text, VTextAlignCapMiddle)
	if !shared {
		t.Fatal("cap-middle did not select a shared baseline")
	}
	expectF(t, -(text.CapHeight() / 2), offset)

	if _, shared := curvedTextSharedBaselineOffset(text, VTextAlignMiddle); shared {
		t.Fatal("ordinary middle unexpectedly selected a shared baseline")
	}
}

func TestCurvedTextCapMiddleCentersVisibleCapsForEitherFacing(t *testing.T) {
	const capHeight = 8.0
	placement := curvedTextGlyphPlacement{
		Advance:        12,
		Point:          Location{X: 40, Y: 50},
		Tangent:        Location{X: 1, Y: 0},
		Normal:         Location{X: 0, Y: 1},
		BaselineOffset: -capHeight / 2,
	}
	for _, facing := range []CurvedTextFacing{CurvedTextFacingUpright, CurvedTextFacingUpsideDown} {
		xAxis, yAxis := curvedTextFrame(placement.Tangent, placement.Normal, facing)
		origin := placement.originPointWithAxes(xAxis, yAxis, placement.BaselineOffset)
		visibleCenter := Location{
			X: origin.X + xAxis.X*placement.Advance/2 + yAxis.X*capHeight/2,
			Y: origin.Y + xAxis.Y*placement.Advance/2 + yAxis.Y*capHeight/2,
		}
		expectV(t, placement.Point, visibleCenter)
	}
}

func TestPlaceCurvedTextGlyphs_OutsidePath(t *testing.T) {
	path := curvedTextTestLinePath{total: 20}
	glyphs := []curvedTextGlyph{
		{GlyphID: 1, Text: "A", Advance: 15},
		{GlyphID: 2, Text: "B", Advance: 15},
	}

	_, err := placeCurvedTextGlyphs(path, glyphs, 0, curvedTextAlignLeft, 0)
	if err != errCurvedTextOutsidePath {
		t.Fatalf("expected errCurvedTextOutsidePath, got %v", err)
	}
}

func TestPageWriter_DrawTextOnCircle_CardinalAnchors(t *testing.T) {
	dw := NewDocWriter()
	afm, err := afm_fonts.Default()
	if err != nil {
		t.Fatalf("Default AFM fonts returned error: %v", err)
	}
	dw.AddFontSource(afm)
	pw := newPageWriter(dw, options.Options{})
	if _, err := pw.SetFont("Courier", 12, options.Options{}); err != nil {
		t.Fatalf("SetFont returned error: %v", err)
	}

	font := pw.Fonts()[0]
	scale := pw.FontSize() * 0.001
	if upm := font.UnitsPerEm(); upm > 0 {
		scale = pw.FontSize() / float64(upm)
	}
	advanceWidth, _ := font.AdvanceWidth('A')
	halfAdvance := scale * float64(advanceWidth) / 2.0
	centerY := pw.translate(50)

	tests := []struct {
		name       string
		startAngle float64
		direction  CurvedTextDirection
		facing     CurvedTextFacing
	}{
		{name: "top", startAngle: 90, direction: CurvedTextClockwise, facing: CurvedTextFacingUpright},
		{name: "right", startAngle: 0, direction: CurvedTextClockwise, facing: CurvedTextFacingUpright},
		{name: "bottom", startAngle: -90, direction: CurvedTextCounterClockwise, facing: CurvedTextFacingUpright},
		{name: "left", startAngle: 180, direction: CurvedTextCounterClockwise, facing: CurvedTextFacingUpright},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pw.stream.Reset()
			pw.inText = false
			opts := CurvedTextOptions{Align: CurvedTextAlignCenter, Direction: tt.direction, Facing: tt.facing}
			if err := pw.DrawTextOnCircle("A", 50, 50, 20, tt.startAngle, opts); err != nil {
				t.Fatalf("DrawTextOnCircle returned error: %v", err)
			}
			path := curvedTextCirclePath{
				center:     Location{50, centerY},
				radius:     20,
				startAngle: tt.startAngle,
				clockwise:  tt.direction == CurvedTextClockwise,
			}
			placements, err := placeCurvedTextGlyphs(path, []curvedTextGlyph{{GlyphID: 1, Text: "A", Advance: halfAdvance * 2}}, 0, curvedTextAlignCenter, 0)
			if err != nil {
				t.Fatalf("placeCurvedTextGlyphs returned error: %v", err)
			}
			placement := orientCurvedTextPlacement(placements[0], CurvedTextOrientationOutside)
			xAxis, yAxis := curvedTextFrame(placement.Tangent, placement.Normal, tt.facing)
			origin := placement.originPointWithAxes(xAxis, yAxis, placement.BaselineOffset)
			matrix := g(xAxis.X) + " " + g(xAxis.Y) + " " + g(yAxis.X) + " " + g(yAxis.Y) + " " + g(origin.X) + " " + g(origin.Y) + " Tm\n"
			if !strings.Contains(pw.stream.String(), matrix) {
				t.Fatalf("expected matrix %q in output:\n%s", matrix, pw.stream.String())
			}
			if !strings.Contains(pw.stream.String(), "(A) Tj\n") {
				t.Fatalf("expected glyph show operator, got:\n%s", pw.stream.String())
			}
		})
	}
}

func TestOrientCurvedTextPlacement_KeepsBaselineOutside(t *testing.T) {
	placement := curvedTextGlyphPlacement{
		Advance:        20,
		Point:          Location{X: 50, Y: 100},
		Tangent:        Location{X: -1, Y: 0},
		Normal:         Location{X: 0, Y: -1},
		BaselineOffset: 12,
	}

	oriented := orientCurvedTextPlacement(placement, CurvedTextOrientationOutside)
	base := oriented.baselinePoint()
	expectV(t, Location{50, 88}, base)
	expectV(t, Location{0, -1}, oriented.Normal)
	expectV(t, Location{-1, 0}, oriented.Tangent)
}

func TestOrientCurvedTextPlacement_InsideFlipsNormal(t *testing.T) {
	placement := curvedTextGlyphPlacement{
		Advance:        20,
		Point:          Location{X: 50, Y: 100},
		Tangent:        Location{X: 1, Y: 0},
		Normal:         Location{X: 0, Y: -1},
		BaselineOffset: 12,
	}

	oriented := orientCurvedTextPlacement(placement, CurvedTextOrientationInside)
	base := oriented.baselinePoint()
	expectV(t, Location{50, 112}, base)
	expectV(t, Location{0, 1}, oriented.Normal)
}

func TestCurvedTextFrame_UprightAvoidsMirroring(t *testing.T) {
	xAxis, yAxis := curvedTextFrame(Location{X: 0.7071, Y: -0.7071}, Location{X: -0.7071, Y: -0.7071}, CurvedTextFacingUpright)
	det := (xAxis.X * yAxis.Y) - (xAxis.Y * yAxis.X)
	if det <= 0 {
		t.Fatalf("expected positive determinant, got %f", det)
	}
}

func TestOriginPointWithAxes_UsesRenderedYAxisForVerticalAnchors(t *testing.T) {
	placement := curvedTextGlyphPlacement{
		Advance:        20,
		Point:          Location{X: 50, Y: 100},
		Tangent:        Location{X: 1, Y: 0},
		Normal:         Location{X: 0, Y: -1},
		BaselineOffset: 12,
	}

	xAxis, yAxis := curvedTextFrame(placement.Tangent, placement.Normal, CurvedTextFacingUpsideDown)
	origin := placement.originPointWithAxes(xAxis, yAxis, placement.BaselineOffset)

	expectV(t, Location{1, 0}, xAxis)
	expectV(t, Location{0, 1}, yAxis)
	expectV(t, Location{40, 112}, origin)
}

func TestCurvedTextFrame_FacingUpsideDownChangesFrameButNotSide(t *testing.T) {
	tangent := Location{X: 1, Y: 0}
	normal := Location{X: 0, Y: -1}

	upX, upY := curvedTextFrame(tangent, normal, CurvedTextFacingUpright)
	downX, downY := curvedTextFrame(tangent, normal, CurvedTextFacingUpsideDown)

	expectV(t, Location{-1, 0}, upX)
	expectV(t, Location{0, -1}, upY)
	expectV(t, Location{1, 0}, downX)
	expectV(t, Location{0, 1}, downY)

	upPlacement := curvedTextGlyphPlacement{
		Advance:        20,
		Point:          Location{X: 50, Y: 100},
		Normal:         normal,
		BaselineOffset: 12,
	}
	upOrigin := upPlacement.originPointWithAxes(upX, upY, upPlacement.BaselineOffset)
	downOrigin := upPlacement.originPointWithAxes(downX, downY, upPlacement.BaselineOffset)
	expectV(t, Location{60, 88}, upOrigin)
	expectV(t, Location{40, 112}, downOrigin)
}

func TestCurvedTextFrame_BottomOutsideTopAlignMovesTextAwayFromCircle(t *testing.T) {
	placement := curvedTextGlyphPlacement{
		Advance: 20,
		Point:   Location{X: 50, Y: 100},
		Tangent: Location{X: 1, Y: 0},
		Normal:  Location{X: 0, Y: -1},
	}

	xAxis, yAxis := curvedTextFrame(placement.Tangent, placement.Normal, CurvedTextFacingUpsideDown)

	baseOrigin := placement.originPointWithAxes(xAxis, yAxis, 0)
	topOrigin := placement.originPointWithAxes(xAxis, yAxis, -12)

	expectV(t, Location{40, 100}, baseOrigin)
	expectV(t, Location{40, 88}, topOrigin)
}

func TestPageWriter_DrawTextOnCircle_BottomOutsideTopAlign(t *testing.T) {
	dw := NewDocWriter()
	afm, err := afm_fonts.Default()
	if err != nil {
		t.Fatalf("Default AFM fonts returned error: %v", err)
	}
	dw.AddFontSource(afm)
	pw := newPageWriter(dw, options.Options{})
	if _, err := pw.SetFont("Courier", 12, options.Options{}); err != nil {
		t.Fatalf("SetFont returned error: %v", err)
	}

	pw.stream.Reset()
	pw.inText = false
	opts := CurvedTextOptions{
		Align:       CurvedTextAlignCenter,
		VAlign:      VTextAlignTop,
		Direction:   CurvedTextCounterClockwise,
		Orientation: CurvedTextOrientationOutside,
		Facing:      CurvedTextFacingUpsideDown,
	}
	if err := pw.DrawTextOnCircle("A", 50, 50, 20, -90, opts); err != nil {
		t.Fatalf("DrawTextOnCircle returned error: %v", err)
	}

	font := pw.Fonts()[0]
	scale := pw.FontSize() * 0.001
	if upm := font.UnitsPerEm(); upm > 0 {
		scale = pw.FontSize() / float64(upm)
	}
	advanceWidth, _ := font.AdvanceWidth('A')
	advance := scale * float64(advanceWidth)
	path := curvedTextCirclePath{
		center:     Location{50, pw.translate(50)},
		radius:     20,
		startAngle: -90,
		clockwise:  false,
	}
	placements, err := placeCurvedTextGlyphs(path, []curvedTextGlyph{{GlyphID: 1, Text: "A", Advance: advance}}, 0, curvedTextAlignCenter, 0)
	if err != nil {
		t.Fatalf("placeCurvedTextGlyphs returned error: %v", err)
	}
	placement := orientCurvedTextPlacement(placements[0], CurvedTextOrientationOutside)
	placement.BaselineOffset = textRiseForFont(font, pw.FontSize(), VTextAlignTop)
	xAxis, yAxis := curvedTextFrame(placement.Tangent, placement.Normal, CurvedTextFacingUpsideDown)
	origin := placement.originPointWithAxes(xAxis, yAxis, placement.BaselineOffset)
	matrix := g(xAxis.X) + " " + g(xAxis.Y) + " " + g(yAxis.X) + " " + g(yAxis.Y) + " " + g(origin.X) + " " + g(origin.Y) + " Tm\n"
	if !strings.Contains(pw.stream.String(), matrix) {
		t.Fatalf("expected matrix %q in output:\n%s", matrix, pw.stream.String())
	}
}

func TestPageWriter_DrawRichTextOnCircle_MatchesStringPath(t *testing.T) {
	dw := NewDocWriter()
	afm, err := afm_fonts.Default()
	if err != nil {
		t.Fatalf("Default AFM fonts returned error: %v", err)
	}
	dw.AddFontSource(afm)

	pwText := newPageWriter(dw, options.Options{})
	if _, err := pwText.SetFont("Courier", 12, options.Options{}); err != nil {
		t.Fatalf("SetFont returned error: %v", err)
	}
	opts := CurvedTextOptions{
		Align:       CurvedTextAlignCenter,
		Direction:   CurvedTextClockwise,
		Orientation: CurvedTextOrientationOutside,
		Facing:      CurvedTextFacingUpright,
	}
	if err := pwText.DrawTextOnCircle("Hello", 50, 50, 20, 90, opts); err != nil {
		t.Fatalf("DrawTextOnCircle returned error: %v", err)
	}

	pwRich := newPageWriter(dw, options.Options{})
	if _, err := pwRich.SetFont("Courier", 12, options.Options{}); err != nil {
		t.Fatalf("SetFont returned error: %v", err)
	}
	rt, err := rich_text.New("Hello", pwRich.Fonts(), pwRich.FontSize(), options.Options{
		"color":     pwRich.FontColor(),
		"strikeout": pwRich.strikeout,
		"underline": pwRich.underline,
	})
	if err != nil {
		t.Fatalf("rich_text.New returned error: %v", err)
	}
	if err := pwRich.DrawRichTextOnCircle(rt, 50, 50, 20, 90, opts); err != nil {
		t.Fatalf("DrawRichTextOnCircle returned error: %v", err)
	}

	expectS(t, pwText.stream.String(), pwRich.stream.String())
}

func TestCurvedTextOptions_ResolveCircleAuto_TopPrefersClockwiseUpright(t *testing.T) {
	opts := CurvedTextOptions{
		Align:     CurvedTextAlignCenter,
		Direction: CurvedTextDirectionAuto,
		Facing:    CurvedTextFacingAuto,
	}
	glyphs := []curvedTextRenderGlyph{
		{curvedTextGlyph: curvedTextGlyph{Text: "A", Advance: 12}},
		{curvedTextGlyph: curvedTextGlyph{Text: "B", Advance: 12}},
		{curvedTextGlyph: curvedTextGlyph{Text: "C", Advance: 12}},
	}

	direction, facing := opts.resolveCircleAuto(glyphs, Location{X: 0, Y: 0}, 40, 90)

	if direction != CurvedTextClockwise {
		t.Fatalf("direction = %v, want %v", direction, CurvedTextClockwise)
	}
	if facing != CurvedTextFacingUpright {
		t.Fatalf("facing = %v, want %v", facing, CurvedTextFacingUpright)
	}
}

func TestCurvedTextOptions_ResolveCircleAuto_BottomPrefersCounterClockwiseUpsideDown(t *testing.T) {
	opts := CurvedTextOptions{
		Align:     CurvedTextAlignCenter,
		Direction: CurvedTextDirectionAuto,
		Facing:    CurvedTextFacingAuto,
	}
	glyphs := []curvedTextRenderGlyph{
		{curvedTextGlyph: curvedTextGlyph{Text: "A", Advance: 18}},
		{curvedTextGlyph: curvedTextGlyph{Text: "B", Advance: 18}},
		{curvedTextGlyph: curvedTextGlyph{Text: "C", Advance: 18}},
		{curvedTextGlyph: curvedTextGlyph{Text: "D", Advance: 18}},
	}

	direction, facing := opts.resolveCircleAuto(glyphs, Location{X: 0, Y: 0}, 40, -90)

	if direction != CurvedTextCounterClockwise {
		t.Fatalf("direction = %v, want %v", direction, CurvedTextCounterClockwise)
	}
	if facing != CurvedTextFacingUpsideDown {
		t.Fatalf("facing = %v, want %v", facing, CurvedTextFacingUpsideDown)
	}
}

func TestCurvedTextOptions_ResolveCircleAuto_RightPrefersClockwiseUpright(t *testing.T) {
	opts := CurvedTextOptions{
		Align:     CurvedTextAlignCenter,
		Direction: CurvedTextDirectionAuto,
		Facing:    CurvedTextFacingAuto,
	}
	glyphs := []curvedTextRenderGlyph{
		{curvedTextGlyph: curvedTextGlyph{Text: "A", Advance: 18}},
		{curvedTextGlyph: curvedTextGlyph{Text: "B", Advance: 18}},
		{curvedTextGlyph: curvedTextGlyph{Text: "C", Advance: 18}},
		{curvedTextGlyph: curvedTextGlyph{Text: "D", Advance: 18}},
	}

	direction, facing := opts.resolveCircleAuto(glyphs, Location{X: 0, Y: 0}, 40, 0)

	if direction != CurvedTextClockwise {
		t.Fatalf("direction = %v, want %v", direction, CurvedTextClockwise)
	}
	if facing != CurvedTextFacingUpright {
		t.Fatalf("facing = %v, want %v", facing, CurvedTextFacingUpright)
	}
}

func TestCurvedTextOptions_ResolveCircleAuto_ExplicitDirectionStillResolvesFacing(t *testing.T) {
	opts := CurvedTextOptions{
		Align:     CurvedTextAlignCenter,
		Direction: CurvedTextClockwise,
		Facing:    CurvedTextFacingAuto,
	}
	glyphs := []curvedTextRenderGlyph{
		{curvedTextGlyph: curvedTextGlyph{Text: "A", Advance: 18}},
		{curvedTextGlyph: curvedTextGlyph{Text: "B", Advance: 18}},
		{curvedTextGlyph: curvedTextGlyph{Text: "C", Advance: 18}},
		{curvedTextGlyph: curvedTextGlyph{Text: "D", Advance: 18}},
	}

	direction, facing := opts.resolveCircleAuto(glyphs, Location{X: 0, Y: 0}, 40, -90)

	if direction != CurvedTextClockwise {
		t.Fatalf("direction = %v, want %v", direction, CurvedTextClockwise)
	}
	if facing != CurvedTextFacingUpsideDown {
		t.Fatalf("facing = %v, want %v", facing, CurvedTextFacingUpsideDown)
	}
}

func TestCurvedTextOptions_ZeroValueUsesAutoForDirectionAndFacing(t *testing.T) {
	opts := CurvedTextOptions{Align: CurvedTextAlignCenter}
	glyphs := []curvedTextRenderGlyph{
		{curvedTextGlyph: curvedTextGlyph{Text: "A", Advance: 18}},
		{curvedTextGlyph: curvedTextGlyph{Text: "B", Advance: 18}},
		{curvedTextGlyph: curvedTextGlyph{Text: "C", Advance: 18}},
		{curvedTextGlyph: curvedTextGlyph{Text: "D", Advance: 18}},
	}

	topDirection, topFacing := opts.resolveCircleAuto(glyphs, Location{X: 0, Y: 0}, 40, 90)
	if topDirection != CurvedTextClockwise {
		t.Fatalf("top direction = %v, want %v", topDirection, CurvedTextClockwise)
	}
	if topFacing != CurvedTextFacingUpright {
		t.Fatalf("top facing = %v, want %v", topFacing, CurvedTextFacingUpright)
	}

	bottomDirection, bottomFacing := opts.resolveCircleAuto(glyphs, Location{X: 0, Y: 0}, 40, -90)
	if bottomDirection != CurvedTextCounterClockwise {
		t.Fatalf("bottom direction = %v, want %v", bottomDirection, CurvedTextCounterClockwise)
	}
	if bottomFacing != CurvedTextFacingUpsideDown {
		t.Fatalf("bottom facing = %v, want %v", bottomFacing, CurvedTextFacingUpsideDown)
	}
}
