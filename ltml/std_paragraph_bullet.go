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

func (p *StdParagraph) drawBullet(w Writer, bullet *BulletStyle, line *rich_text.RichText, slotX, baselineY float64) error {
	layout := p.bulletLayout(w, bullet, line, slotX, baselineY)
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
		bullet.font.Apply(w)
	} else if p.Font() != nil {
		p.Font().Apply(w)
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

func (p *StdParagraph) bulletLayout(w Writer, bullet *BulletStyle, line *rich_text.RichText, slotX, baselineY float64) paragraphBulletLayout {
	slotWidth := bullet.Width()
	lineHeight := 0.0
	lineAscent := 0.0
	if line != nil {
		lineHeight = line.Height()
		lineAscent = line.Ascent()
	}
	if lineHeight <= 0 {
		lineHeight = w.FontSize()
		if lineHeight <= 0 {
			lineHeight = defaultFontSize
		}
	}
	renderHeight := bullet.Height()
	explicitHeight := renderHeight > 0
	if !explicitHeight {
		renderHeight = lineHeight
	}
	if !explicitHeight && renderHeight > lineHeight {
		renderHeight = lineHeight
	}
	renderWidth := slotWidth
	if bullet.IsImage() {
		renderWidth, renderHeight = p.resolveImageBulletSize(w, bullet, slotWidth, renderHeight)
	} else if bullet.IsShape() {
		renderWidth = min(slotWidth, renderHeight)
	}
	if renderWidth <= 0 {
		renderWidth = min(slotWidth, renderHeight)
	}
	if renderWidth > slotWidth && slotWidth > 0 {
		scale := slotWidth / renderWidth
		renderWidth = slotWidth
		renderHeight *= scale
	}
	slotY := baselineY - lineAscent
	renderY := slotY + max((lineHeight-renderHeight)/2, 0)
	renderX := slotX
	if IsRTL(p) {
		renderX += max(slotWidth-renderWidth, 0)
	}
	var shape *pdf.ClosedShape
	var shapeBounds pdf.Bounds
	if bullet.IsShape() {
		shapeValue := closedShapeForBullet(bullet, renderX+renderWidth/2, renderY+renderHeight/2, renderWidth, renderHeight)
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

func (p *StdParagraph) textBulletX(w Writer, bullet *BulletStyle, layout paragraphBulletLayout) float64 {
	if !IsRTL(p) {
		return layout.slotX
	}
	return layout.slotX + max(layout.slotWidth-p.bulletTextWidth(w, bullet), 0)
}

func (p *StdParagraph) bulletTextWidth(w Writer, bullet *BulletStyle) float64 {
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
	if bullet.points >= 5 {
		return bullet.points
	}
	return 5
}

func closedShapeForBullet(bullet *BulletStyle, centerX, centerY, renderWidth, renderHeight float64) pdf.ClosedShape {
	outerRadius := min(renderWidth, renderHeight) / 2
	rotation := defaultBulletRotation(bullet) + bullet.rotation
	switch bullet.Shape() {
	case "circle":
		return pdf.ClosedShape{
			Kind:   pdf.ClosedShapeCircle,
			Center: pdf.Location{X: centerX, Y: centerY},
			Radius: outerRadius,
		}
	case "ellipse":
		return pdf.ClosedShape{
			Kind:    pdf.ClosedShapeEllipse,
			Center:  pdf.Location{X: centerX, Y: centerY},
			RadiusX: renderWidth / 2,
			RadiusY: renderHeight / 2,
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
		case 5:
			return 36
		case 6:
			return 30
		}
	}
	return 0
}
