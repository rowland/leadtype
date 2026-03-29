package svg

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
	"strings"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

const (
	defaultSVGWidth  = 300.0
	defaultSVGHeight = 150.0
)

func Parse(data []byte) (*Document, []Warning, error) {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	decoder.CharsetReader = svgCharsetReader
	doc := &Document{
		ClipPaths: map[string]*ClipPath{},
		Classes:   map[string]StyleSpec{},
	}
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, err
		}
		start, ok := token.(xml.StartElement)
		if !ok {
			continue
		}
		if start.Name.Local != "svg" {
			return nil, nil, fmt.Errorf("root element must be <svg>")
		}
		if err := parseRoot(doc, decoder, start); err != nil {
			return nil, nil, err
		}
		break
	}
	if doc.Width <= 0 {
		doc.Width = defaultSVGWidth
	}
	if doc.Height <= 0 {
		doc.Height = defaultSVGHeight
	}
	if doc.ViewBox.Width <= 0 || doc.ViewBox.Height <= 0 {
		doc.ViewBox = ViewBox{Width: doc.Width, Height: doc.Height}
	}
	return doc, doc.Warnings, nil
}

func LooksLikeSVG(data []byte) bool {
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" {
		return false
	}
	if strings.HasPrefix(trimmed, "<?xml") {
		if idx := strings.Index(trimmed, "?>"); idx >= 0 {
			trimmed = strings.TrimSpace(trimmed[idx+2:])
		}
	}
	return strings.HasPrefix(trimmed, "<svg") || strings.Contains(trimmed, "<svg ")
}

func svgCharsetReader(charset string, input io.Reader) (io.Reader, error) {
	switch strings.ToLower(strings.TrimSpace(charset)) {
	case "iso-8859-1", "iso8859-1", "latin-1", "latin1":
		return transform.NewReader(input, charmap.ISO8859_1.NewDecoder()), nil
	default:
		return nil, fmt.Errorf("unsupported xml charset %q", charset)
	}
}

func parseRoot(doc *Document, decoder *xml.Decoder, start xml.StartElement) error {
	attrs := attrMap(start.Attr)
	viewBox, hasViewBox, err := parseViewBox(attrs["viewBox"])
	if err != nil {
		return err
	}
	doc.ViewBox = viewBox
	width, widthErr := parseLength(attrs["width"], viewBox.Width)
	height, heightErr := parseLength(attrs["height"], viewBox.Height)
	switch {
	case attrs["width"] != "" && attrs["height"] != "" && widthErr == nil && heightErr == nil:
		doc.Width = width
		doc.Height = height
	case hasViewBox:
		doc.Width = viewBox.Width
		doc.Height = viewBox.Height
	default:
		if widthErr == nil {
			doc.Width = width
		}
		if heightErr == nil {
			doc.Height = height
		}
	}
	if !hasViewBox {
		doc.ViewBox = ViewBox{Width: doc.Width, Height: doc.Height}
	}
	children, err := parseChildren(decoder, start.Name.Local, doc, viewportFromDoc(doc))
	if err != nil {
		return err
	}
	doc.Children = children
	return nil
}

func viewportFromDoc(doc *Document) ViewBox {
	if doc.ViewBox.Width > 0 && doc.ViewBox.Height > 0 {
		return doc.ViewBox
	}
	return ViewBox{Width: doc.Width, Height: doc.Height}
}

func parseChildren(decoder *xml.Decoder, parent string, doc *Document, viewport ViewBox) ([]Node, error) {
	children := []Node{}
	for {
		token, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				return children, nil
			}
			return nil, err
		}
		switch tok := token.(type) {
		case xml.StartElement:
			node, err := parseElement(decoder, tok, doc, viewport)
			if err != nil {
				return nil, err
			}
			if node != nil {
				children = append(children, node)
			}
		case xml.EndElement:
			if tok.Name.Local == parent {
				return children, nil
			}
		}
	}
}

func parseElement(decoder *xml.Decoder, start xml.StartElement, doc *Document, viewport ViewBox) (Node, error) {
	attrs := attrMap(start.Attr)
	common, err := parseCommon(doc, attrs, viewport, start.Name.Local)
	if err != nil {
		doc.Warnings = append(doc.Warnings, Warning{Element: start.Name.Local, Message: err.Error()})
		if skipErr := decoder.Skip(); skipErr != nil {
			return nil, skipErr
		}
		return nil, nil
	}
	switch start.Name.Local {
	case "g":
		children, err := parseChildren(decoder, start.Name.Local, doc, viewport)
		if err != nil {
			return nil, err
		}
		return &Group{Common: common, Children: children}, nil
	case "defs":
		children, err := parseChildren(decoder, start.Name.Local, doc, viewport)
		if err != nil {
			return nil, err
		}
		return &Defs{Common: common, Children: children}, nil
	case "style":
		css, err := collectText(decoder, start.Name.Local, doc)
		if err != nil {
			return nil, err
		}
		parseStyleSheet(doc, css, viewport)
		return nil, nil
	case "title", "desc":
		if _, err := collectText(decoder, start.Name.Local, doc); err != nil {
			return nil, err
		}
		return nil, nil
	case "clipPath":
		if attrs["clipPathUnits"] != "" && attrs["clipPathUnits"] != "userSpaceOnUse" {
			doc.Warnings = append(doc.Warnings, Warning{Element: "clipPath", Attribute: "clipPathUnits", Message: "only userSpaceOnUse is supported"})
		}
		children, err := parseChildren(decoder, start.Name.Local, doc, viewport)
		if err != nil {
			return nil, err
		}
		node := &ClipPath{Common: common, Units: attrs["clipPathUnits"], Children: children}
		if node.ID != "" {
			doc.ClipPaths[node.ID] = node
		}
		return node, nil
	case "line":
		if err := decoder.Skip(); err != nil {
			return nil, err
		}
		return &Line{
			Common: common,
			X1:     mustLength(attrs["x1"], viewport.Width),
			Y1:     mustLength(attrs["y1"], viewport.Height),
			X2:     mustLength(attrs["x2"], viewport.Width),
			Y2:     mustLength(attrs["y2"], viewport.Height),
		}, nil
	case "rect":
		if err := decoder.Skip(); err != nil {
			return nil, err
		}
		return &Rect{
			Common: common,
			X:      mustLength(attrs["x"], viewport.Width),
			Y:      mustLength(attrs["y"], viewport.Height),
			Width:  mustLength(attrs["width"], viewport.Width),
			Height: mustLength(attrs["height"], viewport.Height),
			RX:     mustLength(attrs["rx"], viewport.Width),
			RY:     mustLength(attrs["ry"], viewport.Height),
		}, nil
	case "circle":
		if err := decoder.Skip(); err != nil {
			return nil, err
		}
		return &Circle{
			Common: common,
			CX:     mustLength(attrs["cx"], viewport.Width),
			CY:     mustLength(attrs["cy"], viewport.Height),
			R:      mustLength(attrs["r"], min(viewport.Width, viewport.Height)),
		}, nil
	case "ellipse":
		if err := decoder.Skip(); err != nil {
			return nil, err
		}
		return &Ellipse{
			Common: common,
			CX:     mustLength(attrs["cx"], viewport.Width),
			CY:     mustLength(attrs["cy"], viewport.Height),
			RX:     mustLength(attrs["rx"], viewport.Width),
			RY:     mustLength(attrs["ry"], viewport.Height),
		}, nil
	case "polyline":
		if err := decoder.Skip(); err != nil {
			return nil, err
		}
		points, err := parsePoints(attrs["points"])
		if err != nil {
			return nil, err
		}
		return &Polyline{Common: common, Points: points}, nil
	case "polygon":
		if err := decoder.Skip(); err != nil {
			return nil, err
		}
		points, err := parsePoints(attrs["points"])
		if err != nil {
			return nil, err
		}
		return &Polygon{Common: common, Points: points}, nil
	case "path":
		if err := decoder.Skip(); err != nil {
			return nil, err
		}
		segments, err := parsePathData(attrs["d"])
		if err != nil {
			return nil, err
		}
		return &Path{Common: common, Segments: segments}, nil
	case "text":
		body, err := collectText(decoder, start.Name.Local, doc)
		if err != nil {
			return nil, err
		}
		return &Text{
			Common: common,
			X:      mustLength(attrs["x"], viewport.Width),
			Y:      mustLength(attrs["y"], viewport.Height),
			Body:   strings.TrimSpace(body),
		}, nil
	default:
		doc.Warnings = append(doc.Warnings, Warning{Element: start.Name.Local, Message: "unsupported element skipped"})
		if err := decoder.Skip(); err != nil {
			return nil, err
		}
		return nil, nil
	}
}

func collectText(decoder *xml.Decoder, parent string, doc *Document) (string, error) {
	var b strings.Builder
	for {
		token, err := decoder.Token()
		if err != nil {
			return "", err
		}
		switch tok := token.(type) {
		case xml.CharData:
			b.Write(tok)
		case xml.StartElement:
			doc.Warnings = append(doc.Warnings, Warning{Element: parent, Message: "nested text elements are unsupported and were skipped"})
			if err := decoder.Skip(); err != nil {
				return "", err
			}
		case xml.EndElement:
			if tok.Name.Local == parent {
				return b.String(), nil
			}
		}
	}
}

func parseCommon(doc *Document, attrs map[string]string, viewport ViewBox, element string) (Common, error) {
	common := Common{
		ID:        attrs["id"],
		Transform: IdentityTransform(),
	}
	if transformText := attrs["transform"]; transformText != "" {
		transform, err := parseTransform(transformText)
		if err != nil {
			return Common{}, err
		}
		common.Transform = transform
	}
	common.Style = parseStyleSpecWithClasses(doc.Classes, attrs, viewport)
	if clipPath := attrs["clip-path"]; clipPath != "" {
		common.ClipPathRef = strings.TrimSpace(strings.TrimPrefix(strings.TrimSuffix(clipPath, ")"), "url(#"))
	}
	return common, nil
}

func parseStyleSpec(attrs map[string]string, viewport ViewBox) StyleSpec {
	return parseStyleSpecWithClasses(nil, attrs, viewport)
}

func parseStyleSpecWithClasses(classes map[string]StyleSpec, attrs map[string]string, viewport ViewBox) StyleSpec {
	merged := map[string]string{}
	spec := StyleSpec{}
	if classes != nil {
		for _, className := range strings.Fields(attrs["class"]) {
			if classSpec, ok := classes[className]; ok {
				spec = mergeStyleSpec(spec, classSpec)
			}
		}
	}
	for key, value := range attrs {
		merged[key] = value
	}
	if inline := attrs["style"]; inline != "" {
		for _, entry := range strings.Split(inline, ";") {
			entry = strings.TrimSpace(entry)
			if entry == "" {
				continue
			}
			parts := strings.SplitN(entry, ":", 2)
			if len(parts) != 2 {
				continue
			}
			merged[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	if value := merged["fill"]; value != "" {
		if paint, err := parseColor(value); err == nil {
			spec.Fill = &paint
		}
	}
	if value := merged["stroke"]; value != "" {
		if paint, err := parseColor(value); err == nil {
			spec.Stroke = &paint
		}
	}
	if value := merged["stroke-width"]; value != "" {
		if f, err := parseLength(value, min(viewport.Width, viewport.Height)); err == nil {
			spec.StrokeWidth = &f
		}
	}
	if value := merged["fill-opacity"]; value != "" {
		if f, err := parseNumber(value); err == nil {
			spec.FillOpacity = &f
		}
	}
	if value := merged["stroke-opacity"]; value != "" {
		if f, err := parseNumber(value); err == nil {
			spec.StrokeOpacity = &f
		}
	}
	if value := merged["opacity"]; value != "" {
		if f, err := parseNumber(value); err == nil {
			spec.Opacity = &f
		}
	}
	if value := merged["stroke-linecap"]; value != "" {
		spec.LineCap = &value
	}
	if value := merged["stroke-linejoin"]; value != "" {
		spec.LineJoin = &value
	}
	if value := merged["stroke-miterlimit"]; value != "" {
		if f, err := parseNumber(value); err == nil {
			spec.MiterLimit = &f
		}
	}
	if value := merged["stroke-dasharray"]; value != "" {
		if dash, err := parseDashArray(value); err == nil {
			spec.DashArray = &dash
		}
	}
	if value := merged["stroke-dashoffset"]; value != "" {
		if f, err := parseNumber(value); err == nil {
			spec.DashOffset = &f
		}
	}
	if value := merged["fill-rule"]; value != "" {
		spec.FillRule = &value
	}
	if value := merged["font-family"]; value != "" {
		spec.FontFamily = &value
	}
	if value := merged["font-size"]; value != "" {
		if f, err := parseLength(value, viewport.Height); err == nil {
			spec.FontSize = &f
		}
	}
	if value := merged["font-style"]; value != "" {
		spec.FontStyle = &value
	}
	if value := merged["font-weight"]; value != "" {
		spec.FontWeight = &value
	}
	if value := merged["text-anchor"]; value != "" {
		spec.TextAnchor = &value
	}
	if value := merged["display"]; value != "" {
		spec.Display = &value
	}
	return spec
}

func mergeStyleSpec(base, extra StyleSpec) StyleSpec {
	out := base
	if extra.Fill != nil {
		out.Fill = extra.Fill
	}
	if extra.Stroke != nil {
		out.Stroke = extra.Stroke
	}
	if extra.StrokeWidth != nil {
		out.StrokeWidth = extra.StrokeWidth
	}
	if extra.FillOpacity != nil {
		out.FillOpacity = extra.FillOpacity
	}
	if extra.StrokeOpacity != nil {
		out.StrokeOpacity = extra.StrokeOpacity
	}
	if extra.Opacity != nil {
		out.Opacity = extra.Opacity
	}
	if extra.LineCap != nil {
		out.LineCap = extra.LineCap
	}
	if extra.LineJoin != nil {
		out.LineJoin = extra.LineJoin
	}
	if extra.MiterLimit != nil {
		out.MiterLimit = extra.MiterLimit
	}
	if extra.DashArray != nil {
		out.DashArray = extra.DashArray
	}
	if extra.DashOffset != nil {
		out.DashOffset = extra.DashOffset
	}
	if extra.FillRule != nil {
		out.FillRule = extra.FillRule
	}
	if extra.FontFamily != nil {
		out.FontFamily = extra.FontFamily
	}
	if extra.FontSize != nil {
		out.FontSize = extra.FontSize
	}
	if extra.FontStyle != nil {
		out.FontStyle = extra.FontStyle
	}
	if extra.FontWeight != nil {
		out.FontWeight = extra.FontWeight
	}
	if extra.TextAnchor != nil {
		out.TextAnchor = extra.TextAnchor
	}
	if extra.Display != nil {
		out.Display = extra.Display
	}
	return out
}

var cssRuleRE = regexp.MustCompile(`(?s)\.([A-Za-z0-9_-]+)\s*\{([^}]*)\}`)

func parseStyleSheet(doc *Document, css string, viewport ViewBox) {
	for _, match := range cssRuleRE.FindAllStringSubmatch(css, -1) {
		if len(match) != 3 {
			continue
		}
		className := strings.TrimSpace(match[1])
		if className == "" {
			continue
		}
		spec := parseStyleSpec(map[string]string{"style": match[2]}, viewport)
		doc.Classes[className] = mergeStyleSpec(doc.Classes[className], spec)
	}
}

func parseViewBox(text string) (ViewBox, bool, error) {
	if strings.TrimSpace(text) == "" {
		return ViewBox{}, false, nil
	}
	values, err := parseFloatList(text)
	if err != nil {
		return ViewBox{}, false, err
	}
	if len(values) != 4 {
		return ViewBox{}, false, fmt.Errorf("viewBox requires four numbers")
	}
	return ViewBox{MinX: values[0], MinY: values[1], Width: values[2], Height: values[3]}, true, nil
}

func attrMap(attrs []xml.Attr) map[string]string {
	out := make(map[string]string, len(attrs))
	for _, attr := range attrs {
		out[attr.Name.Local] = attr.Value
	}
	return out
}

func mustLength(text string, relative float64) float64 {
	if text == "" {
		return 0
	}
	value, err := parseLength(text, relative)
	if err != nil {
		return 0
	}
	return value
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
