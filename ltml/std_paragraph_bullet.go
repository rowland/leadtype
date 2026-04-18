package ltml

import (
	"fmt"
	"strings"

	"github.com/rowland/leadtype/colors"
	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/pdf"
	"github.com/rowland/leadtype/rich_text"
)

type paragraphBulletLayout struct {
	slotX        float64
	slotY        float64
	slotWidth    float64
	baselineY    float64
	renderX      float64
	renderY      float64
	renderWidth  float64
	renderHeight float64
	shape        *pdf.ClosedShape
	shapeBounds  pdf.Bounds
}

func (p *StdParagraph) drawBullet(w Writer, bullet *BulletStyle, line *rich_text.RichText, slotX, baselineY, textHeight float64) error {
	layout := p.bulletLayout(w, bullet, line, slotX, baselineY, textHeight)
	switch {
	case bullet.IsImage():
		return p.drawImageBullet(w, bullet, layout)
	case bullet.IsShape():
		return p.drawShapeBullet(w, bullet, layout)
	default:
		return p.drawTextBullet(w, bullet, layout)
	}
}

func (p *StdParagraph) drawTextBullet(w Writer, bullet *BulletStyle, layout paragraphBulletLayout) error {
	if bullet.font != nil {
		applyExplicitFontForContainer(w, p, bullet.font)
	} else if p.Font() != nil {
		applyContainerFont(w, p)
	}
	w.MoveTo(p.textBulletX(w, bullet, layout), layout.baselineY)
	return w.Print(bullet.Text())
}

func (p *StdParagraph) drawImageBullet(w Writer, bullet *BulletStyle, layout paragraphBulletLayout) error {
	ref, err := p.resolveBulletAsset(bullet.Source())
	if err != nil {
		return err
	}
	if ref.identifier == "" {
		return fmt.Errorf("bullet src must be specified")
	}
	width := layout.renderWidth
	height := layout.renderHeight
	_, _, err = w.PrintImageFile(ref.identifier, layout.renderX, layout.renderY, &width, &height)
	return err
}

func (p *StdParagraph) drawShapeBullet(w Writer, bullet *BulletStyle, layout paragraphBulletLayout) error {
	brush := bullet.brush
	pen := bullet.pen
	if brush == nil && pen == nil {
		brush = &BrushStyle{id: "bullet_default", color: colors.Black}
	}

	if brush != nil && brush.Kind() != BrushKindSolid {
		if err := p.paintBulletShapeBrush(w, layout, brush); err != nil {
			return err
		}
	} else {
		if pen != nil {
			pen.Apply(w)
		}
		if brush != nil {
			brush.Apply(w)
		}
		if err := p.drawBulletShape(w, layout, pen != nil, brush != nil); err != nil {
			return err
		}
	}

	if pen != nil && brush != nil && brush.Kind() != BrushKindSolid {
		pen.Apply(w)
		if err := p.drawBulletShape(w, layout, true, false); err != nil {
			return err
		}
	}
	return nil
}

func (p *StdParagraph) paintBulletShapeBrush(w Writer, layout paragraphBulletLayout, brush *BrushStyle) error {
	if layout.shape == nil {
		return fmt.Errorf("shape bullet geometry is not available")
	}
	var paintErr error
	if err := w.ClipClosedShape(*layout.shape, func() {
		paintErr = p.paintBrushInRect(w, brush, layout.shapeBounds.MinX, layout.shapeBounds.MinY, layout.shapeBounds.Width(), layout.shapeBounds.Height())
	}); err != nil {
		return err
	}
	return paintErr
}

func (p *StdParagraph) drawBulletShape(w Writer, layout paragraphBulletLayout, border, fill bool) error {
	if layout.shape == nil {
		return fmt.Errorf("shape bullet geometry is not available")
	}
	return w.DrawClosedShape(*layout.shape, border, fill)
}

func (p *StdParagraph) bulletLayout(w Writer, bullet *BulletStyle, line *rich_text.RichText, slotX, baselineY, textHeight float64) paragraphBulletLayout {
	slotWidth := bullet.Width()
	lineHeight, lineAscent := bulletLineMetrics(w, line)
	renderWidth, renderHeight := p.bulletRenderSize(w, bullet, line, lineHeight)
	if renderWidth <= 0 {
		renderWidth = min(slotWidth, renderHeight)
	}
	if renderWidth > slotWidth && slotWidth > 0 {
		scale := slotWidth / renderWidth
		renderWidth = slotWidth
		renderHeight *= scale
	}
	slotY := baselineY - lineAscent
	renderX := p.bulletRenderX(bullet, slotX, slotWidth, renderWidth)
	renderY := p.bulletRenderY(bullet, slotY, textHeight, baselineY, renderHeight)
	var shape *pdf.ClosedShape
	var shapeBounds pdf.Bounds
	if bullet.IsShape() {
		shapeValue := closedShapeForBullet(bullet, renderX+renderWidth/2, renderY+renderHeight/2, lineHeight)
		if bounds, err := w.ClosedShapeBounds(shapeValue); err == nil {
			shapeValue.Center.X += renderX - bounds.MinX
			shapeValue.Center.Y += renderY + max((renderHeight-bounds.Height())/2, 0) - bounds.MinY
			if shiftedBounds, err := w.ClosedShapeBounds(shapeValue); err == nil {
				shapeBounds = shiftedBounds
				shape = &shapeValue
			}
		}
	}
	return paragraphBulletLayout{
		slotX:        slotX,
		slotY:        slotY,
		slotWidth:    slotWidth,
		baselineY:    baselineY,
		renderX:      renderX,
		renderY:      renderY,
		renderWidth:  renderWidth,
		renderHeight: renderHeight,
		shape:        shape,
		shapeBounds:  shapeBounds,
	}
}

func bulletLineMetrics(w Writer, line *rich_text.RichText) (lineHeight, lineAscent float64) {
	if line != nil {
		lineHeight = line.Height()
		lineAscent = line.Ascent()
	}
	if lineHeight <= 0 {
		lineHeight = w.FontSize()
		if lineHeight <= 0 {
			lineHeight = defaultFontSize
		}
		lineAscent = lineHeight
	}
	return lineHeight, lineAscent
}

func (p *StdParagraph) bulletRenderSize(w Writer, bullet *BulletStyle, line *rich_text.RichText, lineHeight float64) (renderWidth, renderHeight float64) {
	if bullet.IsImage() {
		renderHeight = bullet.Height()
		if renderHeight <= 0 {
			renderHeight = lineHeight
		}
		return p.resolveImageBulletSize(w, bullet, bullet.Width(), renderHeight)
	}
	if !bullet.IsShape() {
		renderHeight = bullet.Height()
		if renderHeight <= 0 {
			renderHeight = lineHeight
		}
		return bullet.Width(), renderHeight
	}
	shape := closedShapeForBullet(bullet, 0, 0, lineHeight)
	bounds, err := w.ClosedShapeBounds(shape)
	if err == nil && bounds.Width() > 0 && bounds.Height() > 0 {
		return min(bullet.Width(), bounds.Width()), bounds.Height()
	}
	fallback := 2 * defaultBulletOuterRadius(bullet, lineHeight, bullet.Width())
	return min(bullet.Width(), fallback), fallback
}

func (p *StdParagraph) textBulletX(w Writer, bullet *BulletStyle, layout paragraphBulletLayout) float64 {
	if !IsRTL(p) {
		return layout.slotX
	}
	return layout.slotX + max(layout.slotWidth-p.bulletTextWidth(w, bullet), 0)
}

func (p *StdParagraph) bulletTextWidth(w Writer, bullet *BulletStyle) float64 {
	if bullet.font != nil {
		applyExplicitFontForContainer(w, p, bullet.font)
	} else {
		applyContainerFont(w, p)
	}
	rt, err := rich_text.New(bullet.Text(), w.Fonts(), w.FontSize(), options.Options{})
	if err != nil || rt == nil {
		return 0
	}
	return rt.Width()
}

func (p *StdParagraph) resolveImageBulletSize(w Writer, bullet *BulletStyle, slotWidth, targetHeight float64) (float64, float64) {
	ref, err := p.resolveBulletAsset(bullet.Source())
	if err != nil || ref.identifier == "" {
		return min(slotWidth, targetHeight), targetHeight
	}
	imageWidth, imageHeight, err := w.ImageDimensionsFromFile(ref.identifier)
	if err != nil || imageWidth <= 0 || imageHeight <= 0 {
		return min(slotWidth, targetHeight), targetHeight
	}
	width := targetHeight * float64(imageWidth) / float64(imageHeight)
	height := targetHeight
	if width > slotWidth && slotWidth > 0 {
		scale := slotWidth / width
		width = slotWidth
		height *= scale
	}
	return width, height
}

func (p *StdParagraph) bulletBoxHeightForLines(w Writer, lines []*rich_text.RichText, textHeight float64) float64 {
	bullet := p.Bullet()
	if bullet == nil || p.suppressBullet || len(lines) == 0 {
		return 0
	}
	return p.bulletBoxHeight(w, bullet, lines[0], textHeight)
}

func (p *StdParagraph) bulletBoxHeight(w Writer, bullet *BulletStyle, line *rich_text.RichText, textHeight float64) float64 {
	lineHeight, _ := bulletLineMetrics(w, line)
	_, renderHeight := p.bulletRenderSize(w, bullet, line, lineHeight)
	boxHeight := renderHeight
	if bullet.Height() > boxHeight {
		boxHeight = bullet.Height()
	}
	if boxHeight < textHeight {
		boxHeight = textHeight
	}
	return boxHeight
}

func (p *StdParagraph) bulletRenderX(bullet *BulletStyle, slotX, slotWidth, renderWidth float64) float64 {
	rtl := IsRTL(p)
	switch bullet.AlignX() {
	case "center":
		return slotX + max((slotWidth-renderWidth)/2, 0)
	case "end":
		if rtl {
			return slotX
		}
		return slotX + max(slotWidth-renderWidth, 0)
	default:
		if rtl {
			return slotX + max(slotWidth-renderWidth, 0)
		}
		return slotX
	}
}

func (p *StdParagraph) bulletRenderY(bullet *BulletStyle, slotY, textHeight, baselineY, renderHeight float64) float64 {
	switch bullet.AlignY() {
	case "middle":
		return slotY + (textHeight-renderHeight)/2
	case "baseline":
		return baselineY - renderHeight
	}
	return slotY
}

func (p *StdParagraph) resolveBulletAsset(src string) (assetSourceRef, error) {
	if strings.TrimSpace(src) == "" {
		return assetSourceRef{}, nil
	}
	if p.doc == nil {
		return assetSourceRef{}, fmt.Errorf("bullet document is not set")
	}
	return p.doc.resolveAssetSource(p.container, src)
}

func bulletSides(bullet *BulletStyle) int {
	switch bullet.Shape() {
	case "triangle":
		return 3
	case "square":
		return 4
	}
	if bullet.sides >= 3 {
		return bullet.sides
	}
	return 3
}

func bulletPoints(bullet *BulletStyle) int {
	if bullet.points >= 2 {
		return bullet.points
	}
	return 5
}

func closedShapeForBullet(bullet *BulletStyle, centerX, centerY, lineHeight float64) pdf.ClosedShape {
	outerRadius := defaultBulletOuterRadius(bullet, lineHeight, bullet.Width())
	rotation := defaultBulletRotation(bullet) + bullet.rotation
	switch bullet.Shape() {
	case "circle":
		return pdf.ClosedShape{
			Kind:   pdf.ClosedShapeCircle,
			Center: pdf.Location{X: centerX, Y: centerY},
			Radius: outerRadius,
		}
	case "ellipse":
		rx := bullet.RadiusX()
		if rx <= 0 {
			rx = bullet.Radius()
		}
		if rx <= 0 {
			rx = outerRadius
		}
		ry := bullet.RadiusY()
		if ry <= 0 {
			ry = bullet.Radius()
		}
		if ry <= 0 {
			ry = rx
		}
		return pdf.ClosedShape{
			Kind:     pdf.ClosedShapeEllipse,
			Center:   pdf.Location{X: centerX, Y: centerY},
			RadiusX:  rx,
			RadiusY:  ry,
			Rotation: rotation,
		}
	case "polygon", "triangle", "square":
		return pdf.ClosedShape{
			Kind:     pdf.ClosedShapePolygon,
			Center:   pdf.Location{X: centerX, Y: centerY},
			Radius:   outerRadius,
			Sides:    bulletSides(bullet),
			Rotation: rotation,
		}
	case "star":
		innerRadius := outerRadius * 0.5
		if bullet.r0 > 0 {
			innerRadius = min(bullet.r0, outerRadius)
		}
		return pdf.ClosedShape{
			Kind:        pdf.ClosedShapeStar,
			Center:      pdf.Location{X: centerX, Y: centerY},
			Radius:      outerRadius,
			InnerRadius: innerRadius,
			Points:      bulletPoints(bullet),
			Rotation:    rotation,
		}
	default:
		return pdf.ClosedShape{}
	}
}

func defaultBulletOuterRadius(bullet *BulletStyle, lineHeight, slotWidth float64) float64 {
	if bullet.Radius() > 0 {
		return bullet.Radius()
	}
	if slotWidth > 0 {
		return slotWidth / 2
	}
	if lineHeight > 0 {
		return lineHeight / 2
	}
	return defaultFontSize / 2
}

func defaultBulletRotation(bullet *BulletStyle) float64 {
	switch bullet.Shape() {
	case "triangle":
		return 180
	case "polygon":
		switch bulletSides(bullet) {
		case 5:
			return 36
		case 6:
			return 30
		}
	case "star":
		switch bulletPoints(bullet) {
		case 2:
			return 90
		case 4:
			return 45
		case 5:
			return 36
		case 6:
			return 30
		}
	}
	return 0
}
