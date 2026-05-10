package svg

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/rowland/leadtype/colors"
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
		Gradients: map[string]*Gradient{},
		Masks:     map[string]*Mask{},
		Filters:   map[string]*Filter{},
		Refs:      map[string]Node{},
		Classes:   map[string]map[string]string{},
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
		node := &Group{Common: common, Children: children}
		registerNodeRef(doc, node)
		return node, nil
	case "defs":
		children, err := parseChildren(decoder, start.Name.Local, doc, viewport)
		if err != nil {
			return nil, err
		}
		node := &Defs{Common: common, Children: children}
		registerNodeRef(doc, node)
		return node, nil
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
	case "linearGradient":
		return parseGradientElement(decoder, start, doc, GradientLinear)
	case "radialGradient":
		return parseGradientElement(decoder, start, doc, GradientRadial)
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
		registerNodeRef(doc, node)
		return node, nil
	case "mask":
		return parseMaskElement(decoder, start, doc, viewport, common)
	case "filter":
		return parseFilterElement(decoder, start, doc, viewport)
	case "line":
		if err := decoder.Skip(); err != nil {
			return nil, err
		}
		node := &Line{
			Common: common,
			X1:     mustLength(attrs["x1"], viewport.Width),
			Y1:     mustLength(attrs["y1"], viewport.Height),
			X2:     mustLength(attrs["x2"], viewport.Width),
			Y2:     mustLength(attrs["y2"], viewport.Height),
		}
		registerNodeRef(doc, node)
		return node, nil
	case "rect":
		if err := decoder.Skip(); err != nil {
			return nil, err
		}
		node := &Rect{
			Common: common,
			X:      mustLength(attrs["x"], viewport.Width),
			Y:      mustLength(attrs["y"], viewport.Height),
			Width:  mustLength(attrs["width"], viewport.Width),
			Height: mustLength(attrs["height"], viewport.Height),
			RX:     mustLength(attrs["rx"], viewport.Width),
			RY:     mustLength(attrs["ry"], viewport.Height),
		}
		registerNodeRef(doc, node)
		return node, nil
	case "circle":
		if err := decoder.Skip(); err != nil {
			return nil, err
		}
		node := &Circle{
			Common: common,
			CX:     mustLength(attrs["cx"], viewport.Width),
			CY:     mustLength(attrs["cy"], viewport.Height),
			R:      mustLength(attrs["r"], min(viewport.Width, viewport.Height)),
		}
		registerNodeRef(doc, node)
		return node, nil
	case "ellipse":
		if err := decoder.Skip(); err != nil {
			return nil, err
		}
		node := &Ellipse{
			Common: common,
			CX:     mustLength(attrs["cx"], viewport.Width),
			CY:     mustLength(attrs["cy"], viewport.Height),
			RX:     mustLength(attrs["rx"], viewport.Width),
			RY:     mustLength(attrs["ry"], viewport.Height),
		}
		registerNodeRef(doc, node)
		return node, nil
	case "polyline":
		if err := decoder.Skip(); err != nil {
			return nil, err
		}
		points, err := parsePoints(attrs["points"])
		if err != nil {
			return nil, err
		}
		node := &Polyline{Common: common, Points: points}
		registerNodeRef(doc, node)
		return node, nil
	case "polygon":
		if err := decoder.Skip(); err != nil {
			return nil, err
		}
		points, err := parsePoints(attrs["points"])
		if err != nil {
			return nil, err
		}
		node := &Polygon{Common: common, Points: points}
		registerNodeRef(doc, node)
		return node, nil
	case "path":
		if err := decoder.Skip(); err != nil {
			return nil, err
		}
		segments, err := parsePathData(attrs["d"])
		if err != nil {
			return nil, err
		}
		node := &Path{Common: common, Segments: segments}
		registerNodeRef(doc, node)
		return node, nil
	case "text":
		body, err := collectText(decoder, start.Name.Local, doc)
		if err != nil {
			return nil, err
		}
		node := &Text{
			Common: common,
			X:      mustLength(attrs["x"], viewport.Width),
			Y:      mustLength(attrs["y"], viewport.Height),
			Body:   strings.TrimSpace(body),
		}
		registerNodeRef(doc, node)
		return node, nil
	case "use":
		if err := decoder.Skip(); err != nil {
			return nil, err
		}
		href := strings.TrimSpace(attrs["href"])
		if href == "" {
			href = strings.TrimSpace(attrs["xlink:href"])
		}
		href = strings.TrimPrefix(strings.TrimSpace(href), "#")
		node := &Use{
			Common: common,
			Href:   href,
			X:      mustLength(attrs["x"], viewport.Width),
			Y:      mustLength(attrs["y"], viewport.Height),
			Width:  mustLength(attrs["width"], viewport.Width),
			Height: mustLength(attrs["height"], viewport.Height),
		}
		registerNodeRef(doc, node)
		return node, nil
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
	props := parseStyleProperties(doc.Classes, attrs)
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
	common.Style = parseStyleSpecFromProperties(props, viewport)
	if warnings := fontFamilyWarnings(element, props["font-family"], common.Style.FontFamilies); len(warnings) > 0 {
		doc.Warnings = append(doc.Warnings, warnings...)
	}
	if clipPath := props["clip-path"]; clipPath != "" {
		common.ClipPathRef = parseURLReference(clipPath)
	}
	if mask := props["mask"]; mask != "" {
		common.MaskRef = parseURLReference(mask)
	}
	if filter := props["filter"]; filter != "" {
		common.FilterRef = parseURLReference(filter)
	}
	return common, nil
}

func parseStyleSpecFromProperties(merged map[string]string, viewport ViewBox) StyleSpec {
	spec := StyleSpec{}
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
		families := parseFontFamilies(value)
		if len(families) > 0 {
			spec.FontFamilies = families
			spec.FontFamily = &families[0]
		}
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
	if value := merged["mix-blend-mode"]; value != "" {
		spec.BlendMode = &value
	}
	if value := merged["display"]; value != "" {
		spec.Display = &value
	}
	return spec
}

func parseFontFamilies(value string) []string {
	var families []string
	for _, entry := range splitFontFamilyList(value) {
		family := normalizeFontFamily(entry)
		if family == "" {
			continue
		}
		families = append(families, family)
	}
	return families
}

func splitFontFamilyList(value string) []string {
	var entries []string
	var b strings.Builder
	var quote rune
	escaped := false
	for _, r := range value {
		switch {
		case escaped:
			b.WriteRune(r)
			escaped = false
		case r == '\\' && quote != 0:
			escaped = true
		case quote != 0:
			b.WriteRune(r)
			if r == quote {
				quote = 0
			}
		case r == '\'' || r == '"':
			quote = r
			b.WriteRune(r)
		case r == ',':
			entries = append(entries, b.String())
			b.Reset()
		default:
			b.WriteRune(r)
		}
	}
	entries = append(entries, b.String())
	return entries
}

func normalizeFontFamily(value string) string {
	value = strings.TrimSpace(value)
	for {
		if len(value) < 2 {
			return value
		}
		first := value[0]
		last := value[len(value)-1]
		if (first != '\'' && first != '"') || first != last {
			return value
		}
		value = strings.TrimSpace(value[1 : len(value)-1])
	}
}

func fontFamilyWarnings(element, raw string, families []string) []Warning {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	entries := splitFontFamilyList(raw)
	warnings := []Warning{}
	if hasUnterminatedFontFamilyQuote(raw) {
		warnings = append(warnings, Warning{Element: element, Attribute: "font-family", Message: fmt.Sprintf("malformed family list %q: unterminated quote", raw)})
	}
	for _, entry := range entries {
		if strings.TrimSpace(entry) == "" || normalizeFontFamily(entry) == "" {
			warnings = append(warnings, Warning{Element: element, Attribute: "font-family", Message: fmt.Sprintf("ignored empty family in %q", raw)})
		}
	}
	if len(families) == 0 {
		warnings = append(warnings, Warning{Element: element, Attribute: "font-family", Message: fmt.Sprintf("no usable families in %q", raw)})
	}
	return warnings
}

func hasUnterminatedFontFamilyQuote(value string) bool {
	var quote rune
	escaped := false
	for _, r := range value {
		switch {
		case escaped:
			escaped = false
		case r == '\\' && quote != 0:
			escaped = true
		case quote != 0:
			if r == quote {
				quote = 0
			}
		case r == '\'' || r == '"':
			quote = r
		}
	}
	return quote != 0
}

var cssRuleRE = regexp.MustCompile(`(?s)\.([A-Za-z0-9_-]+)\s*\{([^}]*)\}`)

func parseStyleSheet(doc *Document, css string, viewport ViewBox) {
	_ = viewport
	for _, match := range cssRuleRE.FindAllStringSubmatch(css, -1) {
		if len(match) != 3 {
			continue
		}
		className := strings.TrimSpace(match[1])
		if className == "" {
			continue
		}
		props := parseStyleProperties(nil, map[string]string{"style": match[2]})
		doc.Classes[className] = mergeProperties(doc.Classes[className], props)
	}
}

func registerNodeRef(doc *Document, node Node) {
	if doc == nil || node == nil {
		return
	}
	if id := strings.TrimSpace(node.NodeCommon().ID); id != "" {
		doc.Refs[id] = node
	}
}

func parseStyleProperties(classes map[string]map[string]string, attrs map[string]string) map[string]string {
	merged := map[string]string{}
	if classes != nil {
		for _, className := range strings.Fields(attrs["class"]) {
			if classProps, ok := classes[className]; ok {
				merged = mergeProperties(merged, classProps)
			}
		}
	}
	for key, value := range attrs {
		merged[key] = strings.TrimSpace(value)
	}
	if inline := attrs["style"]; inline != "" {
		merged = mergeProperties(merged, parseInlineStyle(inline))
	}
	return merged
}

func parseInlineStyle(text string) map[string]string {
	props := map[string]string{}
	for _, entry := range strings.Split(text, ";") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		parts := strings.SplitN(entry, ":", 2)
		if len(parts) != 2 {
			continue
		}
		props[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}
	return props
}

func mergeProperties(base, extra map[string]string) map[string]string {
	if base == nil {
		base = map[string]string{}
	}
	for key, value := range extra {
		base[key] = value
	}
	return base
}

func parseURLReference(value string) string {
	value = strings.TrimSpace(value)
	if !strings.HasPrefix(value, "url(") || !strings.HasSuffix(value, ")") {
		return ""
	}
	value = strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(value, "url("), ")"))
	value = strings.Trim(value, `"'`)
	return strings.TrimPrefix(value, "#")
}

func parseGradientElement(decoder *xml.Decoder, start xml.StartElement, doc *Document, kind GradientKind) (Node, error) {
	attrs := attrMap(start.Attr)
	gradient := &Gradient{
		ID:           strings.TrimSpace(attrs["id"]),
		Kind:         kind,
		Units:        strings.TrimSpace(attrs["gradientUnits"]),
		SpreadMethod: strings.TrimSpace(attrs["spreadMethod"]),
		Transform:    IdentityTransform(),
		Href:         strings.TrimSpace(attrs["href"]),
	}
	if gradient.Href == "" {
		gradient.Href = strings.TrimSpace(attrs["xlink:href"])
	}
	if gradient.Href != "" {
		gradient.Href = strings.TrimPrefix(gradient.Href, "#")
		doc.Warnings = append(doc.Warnings, Warning{Element: start.Name.Local, Attribute: "href", Message: "gradient inheritance is not yet implemented"})
	}
	if gradient.Units == "" {
		gradient.Units = "objectBoundingBox"
	}
	if gradient.SpreadMethod == "" {
		gradient.SpreadMethod = "pad"
	}
	if gradient.SpreadMethod != "pad" {
		doc.Warnings = append(doc.Warnings, Warning{Element: start.Name.Local, Attribute: "spreadMethod", Message: "only spreadMethod=pad is supported"})
	}
	if transformText := attrs["gradientTransform"]; transformText != "" {
		transform, err := parseTransform(transformText)
		if err != nil {
			doc.Warnings = append(doc.Warnings, Warning{Element: start.Name.Local, Attribute: "gradientTransform", Message: err.Error()})
		} else {
			gradient.Transform = transform
		}
	}
	switch kind {
	case GradientLinear:
		gradient.Linear = LinearGradient{
			X1: mustLengthValue(attrs["x1"], Length{Value: 0, Unit: LengthPercent}),
			Y1: mustLengthValue(attrs["y1"], Length{Value: 0, Unit: LengthPercent}),
			X2: mustLengthValue(attrs["x2"], Length{Value: 100, Unit: LengthPercent}),
			Y2: mustLengthValue(attrs["y2"], Length{Value: 0, Unit: LengthPercent}),
		}
	case GradientRadial:
		cx := mustLengthValue(attrs["cx"], Length{Value: 50, Unit: LengthPercent})
		cy := mustLengthValue(attrs["cy"], Length{Value: 50, Unit: LengthPercent})
		gradient.Radial = RadialGradient{
			CX: cx,
			CY: cy,
			R:  mustLengthValue(attrs["r"], Length{Value: 50, Unit: LengthPercent}),
			FX: mustLengthValue(attrs["fx"], cx),
			FY: mustLengthValue(attrs["fy"], cy),
			FR: mustLengthValue(attrs["fr"], Length{Value: 0}),
		}
	}
	stops, err := parseGradientStops(decoder, start.Name.Local, doc)
	if err != nil {
		return nil, err
	}
	if len(stops) >= 2 {
		gradient.Stops = stops
		if gradient.ID != "" {
			doc.Gradients[gradient.ID] = gradient
		}
	} else {
		doc.Warnings = append(doc.Warnings, Warning{Element: start.Name.Local, Attribute: "id", Message: "gradient skipped because it has fewer than two valid stops"})
	}
	return nil, nil
}

func parseGradientStops(decoder *xml.Decoder, parent string, doc *Document) ([]GradientStop, error) {
	stops := make([]GradientStop, 0, 4)
	lastPos := 0.0
	for {
		token, err := decoder.Token()
		if err != nil {
			return nil, err
		}
		switch tok := token.(type) {
		case xml.StartElement:
			if tok.Name.Local != "stop" {
				doc.Warnings = append(doc.Warnings, Warning{Element: parent, Message: fmt.Sprintf("unsupported child <%s> skipped", tok.Name.Local)})
				if err := decoder.Skip(); err != nil {
					return nil, err
				}
				continue
			}
			stop, ok := parseGradientStop(tok, doc)
			if err := decoder.Skip(); err != nil {
				return nil, err
			}
			if !ok {
				continue
			}
			if stop.Position < lastPos {
				stop.Position = lastPos
			}
			lastPos = stop.Position
			stops = append(stops, stop)
		case xml.EndElement:
			if tok.Name.Local == parent {
				return normalizeGradientStops(stops), nil
			}
		}
	}
}

func parseGradientStop(start xml.StartElement, doc *Document) (GradientStop, bool) {
	attrs := attrMap(start.Attr)
	props := parseStyleProperties(doc.Classes, attrs)
	position, err := parseUnitInterval(props["offset"])
	if err != nil {
		doc.Warnings = append(doc.Warnings, Warning{Element: "stop", Attribute: "offset", Message: err.Error()})
		return GradientStop{}, false
	}
	paint, err := parseColor(props["stop-color"])
	if err != nil {
		if props["stop-color"] == "" {
			paint = Paint{Set: true, Color: colors.Black}
		} else {
			doc.Warnings = append(doc.Warnings, Warning{Element: "stop", Attribute: "stop-color", Message: err.Error()})
			return GradientStop{}, false
		}
	}
	opacity := 1.0
	if value := props["stop-opacity"]; value != "" {
		if parsed, err := parseNumber(value); err == nil {
			opacity = clamp01(parsed)
		} else {
			doc.Warnings = append(doc.Warnings, Warning{Element: "stop", Attribute: "stop-opacity", Message: err.Error()})
		}
	}
	return GradientStop{Position: position, Color: paint.Color, Opacity: opacity}, true
}

func normalizeGradientStops(stops []GradientStop) []GradientStop {
	if len(stops) == 0 {
		return nil
	}
	out := append([]GradientStop(nil), stops...)
	out[0].Position = 0
	out[len(out)-1].Position = 1
	for i := 1; i < len(out); i++ {
		if out[i].Position < out[i-1].Position {
			out[i].Position = out[i-1].Position
		}
	}
	return out
}

func parseMaskElement(decoder *xml.Decoder, start xml.StartElement, doc *Document, viewport ViewBox, common Common) (Node, error) {
	attrs := attrMap(start.Attr)
	children, err := parseChildren(decoder, start.Name.Local, doc, viewport)
	if err != nil {
		return nil, err
	}
	mask := &Mask{
		Common:       common,
		X:            mustLength(attrs["x"], viewport.Width),
		Y:            mustLength(attrs["y"], viewport.Height),
		Width:        mustLength(attrs["width"], viewport.Width),
		Height:       mustLength(attrs["height"], viewport.Height),
		Units:        attrs["maskUnits"],
		ContentUnits: attrs["maskContentUnits"],
		Children:     children,
	}
	if mask.Units == "" {
		mask.Units = "objectBoundingBox"
	}
	if mask.ContentUnits == "" {
		mask.ContentUnits = "userSpaceOnUse"
	}
	if mask.ID != "" {
		doc.Masks[mask.ID] = mask
	}
	registerNodeRef(doc, mask)
	return mask, nil
}

func parseFilterElement(decoder *xml.Decoder, start xml.StartElement, doc *Document, viewport ViewBox) (Node, error) {
	attrs := attrMap(start.Attr)
	filter := &Filter{
		ID:     strings.TrimSpace(attrs["id"]),
		X:      mustLength(attrs["x"], viewport.Width),
		Y:      mustLength(attrs["y"], viewport.Height),
		Width:  mustLength(attrs["width"], viewport.Width),
		Height: mustLength(attrs["height"], viewport.Height),
		Units:  attrs["filterUnits"],
	}
	for {
		token, err := decoder.Token()
		if err != nil {
			return nil, err
		}
		switch tok := token.(type) {
		case xml.StartElement:
			switch tok.Name.Local {
			case "feFlood":
				primitive := FilterPrimitive{Kind: "feFlood"}
				props := parseStyleProperties(doc.Classes, attrMap(tok.Attr))
				if primitive.Result = props["result"]; primitive.Result == "" {
					primitive.Result = props["id"]
				}
				if colorText := props["flood-color"]; colorText != "" {
					if paint, err := parseColor(colorText); err == nil {
						primitive.FloodColor = paint.Color
					}
				} else {
					primitive.FloodColor = colors.White
				}
				filter.Primitives = append(filter.Primitives, primitive)
				if err := decoder.Skip(); err != nil {
					return nil, err
				}
			case "feBlend":
				props := parseStyleProperties(doc.Classes, attrMap(tok.Attr))
				filter.Primitives = append(filter.Primitives, FilterPrimitive{
					Kind:   "feBlend",
					In:     props["in"],
					In2:    props["in2"],
					Result: props["result"],
				})
				if err := decoder.Skip(); err != nil {
					return nil, err
				}
			default:
				doc.Warnings = append(doc.Warnings, Warning{Element: "filter", Message: fmt.Sprintf("unsupported child <%s> skipped", tok.Name.Local)})
				if err := decoder.Skip(); err != nil {
					return nil, err
				}
			}
		case xml.EndElement:
			if tok.Name.Local == start.Name.Local {
				if filter.ID != "" {
					doc.Filters[filter.ID] = filter
				}
				return nil, nil
			}
		}
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
