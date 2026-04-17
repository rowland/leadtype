package ltml

import (
	"fmt"
	"math"
	"strings"

	"github.com/rowland/leadtype/colors"
	"github.com/rowland/leadtype/rich_text"
)

type paragraphBulletLayout struct {
	slotX        float64
	slotY        float64
	slotWidth    float64
	renderX      float64
	renderY      float64
	renderWidth  float64
	renderHeight float64
	centerX      float64
	centerY      float64
}

type bulletShapeBounds struct {
	minX float64
	maxX float64
	minY float64
	maxY float64
}

func (p *StdParagraph) drawBullet(w Writer, bullet *BulletStyle, line *rich_text.RichText, slotX, baselineY float64) error {
	layout := p.bulletLayout(w, bullet, line, slotX, baselineY)
	switch {
	case bullet.IsImage():
		return p.drawImageBullet(w, bullet, layout)
	case bullet.IsShape():
		return p.drawShapeBullet(w, bullet, layout)
	default:
		return p.drawTextBullet(w, bullet)
	}
}

func (p *StdParagraph) drawTextBullet(w Writer, bullet *BulletStyle) error {
	if bullet.font != nil {
		bullet.font.Apply(w)
	} else if p.Font() != nil {
		p.Font().Apply(w)
	}
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
		if err := p.paintBulletShapeBrush(w, bullet, layout, brush); err != nil {
			return err
		}
	} else {
		if pen != nil {
			pen.Apply(w)
		}
		if brush != nil {
			brush.Apply(w)
		}
		if err := p.drawBulletShape(w, bullet, layout, pen != nil, brush != nil); err != nil {
			return err
		}
	}

	if pen != nil && brush != nil && brush.Kind() != BrushKindSolid {
		pen.Apply(w)
		if err := p.drawBulletShape(w, bullet, layout, true, false); err != nil {
			return err
		}
	}
	return nil
}

func (p *StdParagraph) paintBulletShapeBrush(w Writer, bullet *BulletStyle, layout paragraphBulletLayout, brush *BrushStyle) error {
	var paintErr error
	if err := w.Path(func() {
		if e := p.buildBulletShapePath(w, bullet, layout); e != nil {
			paintErr = e
			return
		}
		if e := w.Clip(func() {
			paintErr = p.paintBrushInRect(w, brush, layout.renderX, layout.renderY, layout.renderWidth, layout.renderHeight)
		}); e != nil && paintErr == nil {
			paintErr = e
		}
	}); err != nil {
		return err
	}
	return paintErr
}

func (p *StdParagraph) buildBulletShapePath(w Writer, bullet *BulletStyle, layout paragraphBulletLayout) error {
	switch strings.TrimSpace(bullet.Shape()) {
	case "circle":
		return w.CirclePath(layout.centerX, layout.centerY, min(layout.renderWidth, layout.renderHeight)/2, false)
	case "ellipse":
		return w.EllipsePath(layout.centerX, layout.centerY, layout.renderWidth/2, layout.renderHeight/2, false)
	case "polygon":
		return w.PolygonPath(layout.centerX, layout.centerY, min(layout.renderWidth, layout.renderHeight)/2, bulletSides(bullet), false, bullet.rotation)
	case "star":
		return w.StarPath(layout.centerX, layout.centerY, min(layout.renderWidth, layout.renderHeight)/2, bulletStarInnerRadius(bullet, min(layout.renderWidth, layout.renderHeight)/2), bulletPoints(bullet), false, bullet.rotation)
	default:
		return fmt.Errorf("unsupported bullet shape %q", bullet.Shape())
	}
}

func (p *StdParagraph) drawBulletShape(w Writer, bullet *BulletStyle, layout paragraphBulletLayout, border, fill bool) error {
	switch strings.TrimSpace(bullet.Shape()) {
	case "circle":
		return w.Circle(layout.centerX, layout.centerY, min(layout.renderWidth, layout.renderHeight)/2, border, fill, false)
	case "ellipse":
		return w.Ellipse(layout.centerX, layout.centerY, layout.renderWidth/2, layout.renderHeight/2, border, fill, false)
	case "polygon":
		return w.Polygon(layout.centerX, layout.centerY, min(layout.renderWidth, layout.renderHeight)/2, bulletSides(bullet), border, fill, false, bullet.rotation)
	case "star":
		return w.Star(layout.centerX, layout.centerY, min(layout.renderWidth, layout.renderHeight)/2, bulletStarInnerRadius(bullet, min(layout.renderWidth, layout.renderHeight)/2), bulletPoints(bullet), border, fill, false, bullet.rotation)
	default:
		return fmt.Errorf("unsupported bullet shape %q", bullet.Shape())
	}
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
	if renderHeight <= 0 {
		renderHeight = lineHeight
	}
	if renderHeight > lineHeight {
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
	centerX := renderX + renderWidth/2
	centerY := renderY + renderHeight/2
	if bullet.IsShape() {
		bounds := bulletShapeGeometryBounds(bullet, renderWidth, renderHeight)
		shapeHeight := bounds.maxY - bounds.minY
		centerX = renderX - bounds.minX
		centerY = renderY + max((renderHeight-shapeHeight)/2, 0) - bounds.minY
	}
	return paragraphBulletLayout{
		slotX:        slotX,
		slotY:        slotY,
		slotWidth:    slotWidth,
		renderX:      renderX,
		renderY:      renderY,
		renderWidth:  renderWidth,
		renderHeight: renderHeight,
		centerX:      centerX,
		centerY:      centerY,
	}
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

func bulletStarInnerRadius(bullet *BulletStyle, outerRadius float64) float64 {
	if bullet.r0 > 0 {
		return math.Min(bullet.r0, outerRadius)
	}
	return outerRadius * 0.5
}

func bulletShapeGeometryBounds(bullet *BulletStyle, renderWidth, renderHeight float64) bulletShapeBounds {
	switch strings.TrimSpace(bullet.Shape()) {
	case "ellipse":
		return bulletShapeBounds{
			minX: -renderWidth / 2,
			maxX: renderWidth / 2,
			minY: -renderHeight / 2,
			maxY: renderHeight / 2,
		}
	case "polygon":
		return pointBounds(bulletPolygonPoints(0, 0, min(renderWidth, renderHeight)/2, bulletSides(bullet), bullet.rotation))
	case "star":
		outerRadius := min(renderWidth, renderHeight) / 2
		return pointBounds(bulletStarPoints(0, 0, outerRadius, bulletStarInnerRadius(bullet, outerRadius), bulletPoints(bullet), bullet.rotation))
	default:
		radius := min(renderWidth, renderHeight) / 2
		return bulletShapeBounds{
			minX: -radius,
			maxX: radius,
			minY: -radius,
			maxY: radius,
		}
	}
}

func bulletPolygonPoints(centerX, centerY, radius float64, sides int, rotation float64) [][2]float64 {
	if sides < 3 {
		return nil
	}
	step := 360.0 / float64(sides)
	angle := step/2.0 + 90.0
	points := make([][2]float64, 0, sides+1)
	for i := 0; i <= sides; i++ {
		dx, dy := bulletRotateXY(1, 0, angle)
		x := centerX + dx*radius
		y := centerY - dy*radius
		if rotation != 0 {
			x, y = bulletRotatePoint(centerX, centerY, x, y, -rotation)
		}
		points = append(points, [2]float64{x, y})
		angle += step
	}
	return points
}

func bulletStarPoints(centerX, centerY, outerRadius, innerRadius float64, points int, rotation float64) [][2]float64 {
	if points < 5 {
		return nil
	}
	outer := bulletPolygonPoints(centerX, centerY, outerRadius, points, rotation)
	inner := bulletPolygonPoints(centerX, centerY, innerRadius, points, rotation+(360.0/float64(points)/2.0))
	if len(outer) == 0 || len(inner) == 0 {
		return nil
	}
	path := make([][2]float64, 0, points*2+1)
	path = append(path, inner[0])
	for i := 0; i < points; i++ {
		path = append(path, outer[i], inner[i+1])
	}
	return path
}

func pointBounds(points [][2]float64) bulletShapeBounds {
	if len(points) == 0 {
		return bulletShapeBounds{}
	}
	bounds := bulletShapeBounds{
		minX: points[0][0],
		maxX: points[0][0],
		minY: points[0][1],
		maxY: points[0][1],
	}
	for _, point := range points[1:] {
		bounds.minX = min(bounds.minX, point[0])
		bounds.maxX = max(bounds.maxX, point[0])
		bounds.minY = min(bounds.minY, point[1])
		bounds.maxY = max(bounds.maxY, point[1])
	}
	return bounds
}

func bulletRotateXY(x, y, angle float64) (float64, float64) {
	theta := angle * math.Pi / 180.0
	return (x * math.Cos(theta)) - (y * math.Sin(theta)), (x * math.Sin(theta)) + (y * math.Cos(theta))
}

func bulletRotatePoint(centerX, centerY, x, y, angle float64) (float64, float64) {
	rotX, rotY := bulletRotateXY(x-centerX, centerY-y, angle)
	return centerX + rotX, centerY - rotY
}
