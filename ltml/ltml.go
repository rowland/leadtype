// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/rowland/leadtype/profile"
)

type Doc struct {
	root         *StdDocument
	stack        []any
	scopeFrames  []scopeFrame
	rootScope    Scope // per-document root scope; parent = &defaultScope
	parseErr     error
	assetFS      fs.FS
	sourceDir    string
	assetSources *assetSourceManager
	profiler     *profile.Profiler
}

type scopeResourceOwner interface {
	resetResourceAttrs()
	setResourceAttrs(map[string]string, Units)
	Units() Units
}

type scopeAttrLayer struct {
	attrs map[string]string
	units Units
}

type scopeFrame struct {
	scope      HasScope
	owner      scopeResourceOwner
	attrLayers []scopeAttrLayer
}

type ParseOption func(*Doc)

func WithAssetFS(assetFS fs.FS) ParseOption {
	return func(doc *Doc) {
		doc.SetAssetFS(assetFS)
	}
}

func WithProfiler(profiler *profile.Profiler) ParseOption {
	return func(doc *Doc) {
		doc.profiler = profiler
	}
}

func newDocWithOptions(opts ...ParseOption) *Doc {
	doc := &Doc{
		assetSources: newAssetSourceManager(),
	}
	for _, opt := range opts {
		if opt != nil {
			opt(doc)
		}
	}
	return doc
}

func (doc *Doc) parseBytes(b []byte) error {
	r := bytes.NewReader(b)
	return doc.parseReader(r)
}

func (doc *Doc) parseFile(filename string) error {
	if abs, err := filepath.Abs(filename); err == nil {
		doc.sourceDir = filepath.Dir(abs)
	} else {
		doc.sourceDir = filepath.Dir(filename)
	}
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return doc.parseReader(f)
}

func (doc *Doc) parseReader(r io.Reader) error {
	span := doc.profiler.Begin("ltml.parse")
	defer span.End()

	dec := xml.NewDecoder(r)
	dec.DefaultSpace = DefaultSpace

	for {
		token, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		switch t := token.(type) {
		case xml.StartElement:
			traceStartElement(t)
			created, capturesBody := doc.startElement(t)
			if capturesBody && doc.parseErr == nil {
				var decoded struct {
					Body string `xml:",innerxml"`
				}
				if err := dec.DecodeElement(&decoded, &t); err != nil {
					return err
				}
				if raw, ok := created.(RawBody); ok {
					raw.SetBody(decoded.Body)
				}
				traceEndElement(xml.EndElement{Name: t.Name})
				doc.endElement(xml.EndElement{Name: t.Name})
			}
		case xml.EndElement:
			traceEndElement(t)
			doc.endElement(t)
		case xml.CharData:
			traceCharData(t)
			doc.charData(t)
		case xml.Comment:
			traceComment(t)
			doc.comment(t)
		}
		if doc.parseErr != nil {
			return doc.parseErr
		}
	}
	{
		span := doc.profiler.Begin("ltml.parse.pseudo_rules")
		doc.applyPseudoRules()
		span.End()
	}
	return validateDocumentLayouts(doc.root)
}

func (doc *Doc) SetAssetFS(assetFS fs.FS) {
	doc.assetFS = assetFS
}

func (doc *Doc) Root() *StdDocument {
	return doc.root
}

func (doc *Doc) Page(i int) *StdPage {
	if doc.root == nil {
		return nil
	}
	return doc.root.Page(i)
}

func (doc *Doc) Print(w Writer) (err error) {
	if doc.root == nil {
		return nil
	}
	if err := validateDocumentLayouts(doc.root); err != nil {
		return err
	}
	profiler := doc.profiler
	if profiler == nil {
		profiler = profilerForWriter(w)
	}
	if profiler != nil {
		doc.root.renderProfiler = profiler
		setWriterProfiler(w, profiler)
	}
	span := profiler.Begin("ltml.doc.print")
	defer span.End()
	defer func() {
		doc.root.renderProfiler = nil
	}()
	if doc.assetFS != nil {
		if assetSetter, ok := any(w).(interface{ SetAssetFS(fs.FS) }); ok {
			assetSetter.SetAssetFS(doc.assetFS)
		}
	}
	if err := doc.root.Print(w); err != nil {
		doc.cleanupAssetSources()
		return err
	}
	if registrar, ok := any(w).(interface{ RegisterWriteToCleanup(func()) }); ok {
		registrar.RegisterWriteToCleanup(doc.cleanupAssetSources)
		return nil
	}
	doc.cleanupAssetSources()
	return nil
}

func (doc *Doc) resolveAssetSource(container Container, src string) (assetSourceRef, error) {
	if doc == nil || doc.assetSources == nil {
		return assetSourceRef{}, fmt.Errorf("asset sources are not initialized")
	}
	return doc.assetSources.resolve(doc, container, src)
}

func (doc *Doc) cleanupAssetSources() {
	if doc == nil || doc.assetSources == nil {
		return
	}
	doc.assetSources.cleanup()
}

func (doc *Doc) startElement(elem xml.StartElement) (any, bool) {
	trueSpace := elem.Name.Space
	trueTag := elem.Name.Local
	var aliasDefaultAttrs map[string]string
	if elem.Name.Space == DefaultSpace {
		if alias, ok := doc.scope().AliasFor(trueTag); ok {
			var err error
			trueSpace, trueTag, err = aliasTarget(DefaultSpace, alias.Tag)
			if err != nil {
				doc.parseErr = err
				return nil, false
			}
			aliasDefaultAttrs = alias.Attrs
			debugf("Alias %s=%s %v\n", elem.Name.Local, trueTag, aliasDefaultAttrs)
		}
	}
	e := makeElement(trueSpace, trueTag)
	if e == nil {
		debugf("Unknown tag: %s:%s\n", elem.Name.Space, elem.Name.Local)
	}
	capturesBody := isComponentTag(trueSpace, trueTag)
	if ws, ok := e.(WantsScope); ok {
		ws.SetScope(doc.scope())
	}
	if wd, ok := e.(WantsDoc); ok {
		wd.SetDoc(doc)
	}
	var err error
	if child, ok := e.(HasParent); ok {
		if err = child.SetParent(doc.current()); err != nil {
			debugf("Setting parent: %s\n", err)
		}
	}
	var wrapper *StdSector
	var parent Container
	if parentCurrent, ok := doc.current().(Container); ok && err == nil {
		if sector, ok := parentCurrent.(*StdSector); ok {
			if widget, ok := e.(Widget); ok && isInlineOnlyWidget(widget) {
				doc.parseErr = fmt.Errorf("%s: <sector> cannot directly contain <%s>; wrap inline content in <label>", sector.Path(), elem.Name.Local)
				return e, capturesBody
			}
		}
		parent = parentCurrent
		if widget, ok := e.(Widget); ok {
			if canvas, ok := widget.(*StdCanvas); ok {
				if _, ok := doc.current().(*StdDocument); !ok {
					doc.parseErr = fmt.Errorf("<canvas> must be direct child of <ltml>")
					return e, capturesBody
				}
				if err = setWidgetContainer(parentCurrent, canvas); err != nil {
					doc.parseErr = err
					return e, capturesBody
				}
			} else if isRadialLayoutStyle(parent.LayoutStyle()) {
				if _, isSector := widget.(*StdSector); !isSector {
					wrapper = &StdSector{}
					if ws, ok := any(wrapper).(WantsScope); ok {
						ws.SetScope(doc.scope())
					}
					if wd, ok := any(wrapper).(WantsDoc); ok {
						wd.SetDoc(doc)
					}
					if err = attachWidgetToContainer(wrapper, widget); err != nil {
						doc.parseErr = err
						return e, capturesBody
					}
					if err = attachWidgetToContainer(parentCurrent, wrapper); err != nil {
						doc.parseErr = err
						return e, capturesBody
					}
				} else {
					if err = attachWidgetToContainer(parentCurrent, widget); err != nil {
						doc.parseErr = err
						return e, capturesBody
					}
				}
			} else {
				if err = attachWidgetToContainer(parentCurrent, widget); err != nil {
					doc.parseErr = err
					return e, capturesBody
				}
			}
		}
	}
	if d, ok := e.(*StdDocument); ok {
		if doc.root != nil {
			doc.parseErr = fmt.Errorf("multiple top-level ltml roots are not supported")
			return e, capturesBody
		}
		d.Scope.defaultRuleTier = 0
		doc.root = d
	} else if page, ok := e.(*StdPage); ok {
		page.Scope.defaultRuleTier = 1
	}
	doc.push(e)

	attrs := mapFromXmlAttrs(elem.Attr)
	if attrs["tag"] == "" {
		attrs["tag"] = elem.Name.Local
	}

	if ident, ok := e.(Identifier); ok {
		ident.SetIentifiers(attrs)
	}
	defaultAttrs := defaultElementAttrs(doc.scope(), e, aliasDefaultAttrs)
	sourcePath := ""
	if wrapper != nil {
		if setter, ok := e.(interface{ SetPath(string) }); ok {
			if parent != nil {
				if selector, ok := e.(interface{ SelectorTag() string }); ok {
					sourcePath = parent.Path() + "/" + selector.SelectorTag()
					setter.SetPath(sourcePath)
				}
			}
		}
	}
	attrLayers := applyElementAttrs(doc.scope(), e, defaultAttrs, attrs, sourcePath)
	doc.captureScopeOwnerAttrs(e, attrLayers)
	switch value := e.(type) {
	case *StdCanvas:
		if doc.root == nil {
			doc.parseErr = fmt.Errorf("<canvas> must be direct child of <ltml>")
			return e, capturesBody
		}
		if err := doc.root.registerCanvas(value); err != nil {
			doc.parseErr = err
			return e, capturesBody
		}
	case *StdDraw:
		if strings.TrimSpace(value.key) == "" {
			doc.parseErr = fmt.Errorf("<draw> requires a key")
			return e, capturesBody
		}
	}
	if widget, ok := e.(interface{ SetRawAttrs(map[string]string) }); ok {
		widget.SetRawAttrs(attrs)
	}
	if style, ok := e.(Styler); ok {
		if st, ok := doc.scope().StyleFor(style.ID()); ok {
			switch st := st.(type) {
			case *BrushStyle:
				bs := st.Clone()
				bs.SetAttrs(attrs)
				style = bs
			case *FontStyle:
				fs := st.Clone()
				fs.SetAttrs(attrs)
				style = fs
			case *ParagraphStyle:
				ps := st.Clone()
				ps.SetAttrs(attrs)
				style = ps
			case *PenStyle:
				ps := st.Clone()
				ps.SetAttrs(attrs)
				style = ps
			}
		}
		if err := doc.scope().AddStyle(style); err != nil {
			debugf("Adding style: %s\n", err)
		} else {
			doc.refreshScopeOwnerResources()
		}
	}
	if layout, ok := e.(*LayoutStyle); ok {
		if layout0, ok := doc.scope().LayoutFor(layout.ID()); ok {
			layout = layout0.Clone()
			layout.SetAttrs(attrs)
		}
		if err := doc.scope().AddLayout(layout); err != nil {
			debugf("Adding layout: %s\n", err)
		} else {
			doc.refreshScopeOwnerResources()
		}
	}
	if pageStyle, ok := e.(*PageStyle); ok {
		if err := doc.scope().AddPageStyle(pageStyle); err != nil {
			debugf("Adding page style: %s\n", err)
		} else {
			doc.refreshScopeOwnerResources()
		}
	}
	if alias, ok := e.(*Alias); ok {
		if _, _, err := aliasTarget(DefaultSpace, alias.Tag); err != nil {
			doc.parseErr = err
			debugf("Adding alias: %s\n", err)
		} else if err := doc.scope().AddAlias(alias); err != nil {
			debugf("Adding alias: %s\n", err)
		}
	}
	if rules, ok := e.(*Rules); ok {
		if err := doc.scope().AddRules(rules); err != nil {
			doc.parseErr = err
			debugf("Adding rules: %s\n", err)
		}
	}
	if _, ok := e.(RawBody); ok {
		capturesBody = true
	}
	return e, capturesBody
}

func attachWidgetToContainer(parent Container, widget Widget) error {
	if err := setWidgetContainer(parent, widget); err != nil {
		return err
	}
	parent.AddChild(widget)
	return nil
}

func setWidgetContainer(parent Container, widget Widget) error {
	if wc, ok := any(widget).(WantsContainer); ok {
		if err := wc.SetContainer(parent); err != nil {
			return err
		}
	}
	return nil
}

func defaultElementAttrs(scope HasScope, target any, aliasDefaults map[string]string) map[string]string {
	var merged map[string]string
	if provider, ok := target.(HasDefaultAttrs); ok {
		merged = cloneAttrs(provider.DefaultAttrs(scope))
	}
	if len(aliasDefaults) == 0 {
		return merged
	}
	if merged == nil {
		return cloneAttrs(aliasDefaults)
	}
	for k, v := range aliasDefaults {
		merged[k] = v
	}
	return merged
}

func cloneAttrs(attrs map[string]string) map[string]string {
	if len(attrs) == 0 {
		return nil
	}
	clone := make(map[string]string, len(attrs))
	for k, v := range attrs {
		clone[k] = v
	}
	return clone
}

// applyPseudoRules performs a second pass over the parsed widget tree so rules
// that rely on structural pseudo-classes can be matched with layout context.
func (doc *Doc) applyPseudoRules() {
	resolver := newSelectorStructureResolver()
	if doc.root != nil {
		doc.applyPseudoRulesToWidget(doc.root, resolver)
		doc.root.eachCanvas(func(canvas *StdCanvas) {
			doc.applyPseudoRulesToWidget(canvas, resolver)
		})
	}
}

// applyPseudoRulesToWidget walks the widget subtree, applying any pseudo-class
// rules that match each widget in scope order.
func (doc *Doc) applyPseudoRulesToWidget(widget Widget, resolver *selectorStructureResolver) {
	if widget == nil {
		return
	}
	scope := doc.scope()
	if scoped, ok := widget.(interface{ Scope() HasScope }); ok && scoped.Scope() != nil {
		scope = scoped.Scope()
	}
	applyPseudoRuleAttrs(scope, widget, resolver)
	if container, ok := widget.(Container); ok {
		for _, child := range container.Widgets() {
			doc.applyPseudoRulesToWidget(child, resolver)
		}
	}
}

// applyElementAttrs applies default attrs, matching selector rules, and direct
// element attrs to a target in that precedence order.
func applyElementAttrs(scope HasScope, target any, defaultAttrs, attrs map[string]string, pathOverride string) []scopeAttrLayer {
	if target == nil {
		return nil
	}
	var applied []scopeAttrLayer
	owner, captureAttrs := target.(scopeResourceOwner)
	if e, ok := target.(HasAttrs); ok {
		apply := func(values map[string]string) {
			applyAttrs(target, e, values)
			if captureAttrs {
				if resourceAttrs := scopeOwnerResourceAttrs(values); len(resourceAttrs) > 0 {
					applied = append(applied, scopeAttrLayer{
						attrs: resourceAttrs,
						units: owner.Units(),
					})
				}
			}
		}
		apply(defaultAttrs)
		path := pathOverride
		if path == "" {
			if p, ok := target.(HasPath); ok {
				path = p.Path()
			}
		}
		if path != "" {
			scope.EachRuleFor(path, func(rule *Rule) {
				apply(rule.Attrs)
			})
		}
		apply(attrs)
	}
	return applied
}

// applyAttrs routes cell presentation from a direct radial child to its
// transparent sector wrapper. Keeping this at the attribute-layer boundary
// makes defaults, selector rules, pseudo rules, and direct attrs obey the same
// ownership rules.
func applyAttrs(target any, receiver HasAttrs, attrs map[string]string) {
	widget, ok := target.(Widget)
	if !ok {
		receiver.SetAttrs(attrs)
		return
	}
	parented, ok := widget.(interface{ Container() Container })
	if !ok {
		receiver.SetAttrs(attrs)
		return
	}
	sector, ok := parented.Container().(*StdSector)
	if !ok || sector.Tag != "" {
		receiver.SetAttrs(attrs)
		return
	}
	cellAttrs, childAttrs := splitImplicitSectorAttrs(attrs)
	sector.SetAttrs(cellAttrs)
	receiver.SetAttrs(childAttrs)
}

func splitImplicitSectorAttrs(attrs map[string]string) (cell, child map[string]string) {
	for name, value := range attrs {
		cellOwned := name == "colspan" || name == "rowspan" || name == "display" || name == "z-index" ||
			name == "fill" || strings.HasPrefix(name, "fill.") ||
			name == "border" || strings.HasPrefix(name, "border.") || strings.HasPrefix(name, "border-") ||
			name == "padding" || strings.HasPrefix(name, "padding-")
		if cellOwned || name == "units" {
			if cell == nil {
				cell = make(map[string]string)
			}
			cell[name] = value
		}
		if !cellOwned || name == "units" {
			if child == nil {
				child = make(map[string]string)
			}
			child[name] = value
		}
	}
	return cell, child
}

func scopeOwnerResourceAttrs(attrs map[string]string) map[string]string {
	var resourceAttrs map[string]string
	for name, value := range attrs {
		switch {
		case name == "font", strings.HasPrefix(name, "font."),
			name == "fill", strings.HasPrefix(name, "fill."),
			name == "border", strings.HasPrefix(name, "border."),
			strings.HasPrefix(name, "border-"),
			name == "layout", strings.HasPrefix(name, "layout."),
			name == "paragraph-style", strings.HasPrefix(name, "paragraph-style."),
			name == "style", strings.HasPrefix(name, "style."):
			if resourceAttrs == nil {
				resourceAttrs = make(map[string]string)
			}
			resourceAttrs[name] = value
		}
	}
	return resourceAttrs
}

// applyPseudoRuleAttrs applies matching pseudo-class selector attrs to a
// widget, then reapplies its raw attrs so explicit element attrs remain final.
func applyPseudoRuleAttrs(scope HasScope, target Widget, resolver *selectorStructureResolver) {
	pseudoScope, ok := scope.(interface {
		EachPseudoRuleForWidget(Widget, *selectorStructureResolver, func(*Rule))
	})
	if !ok {
		return
	}
	rawCarrier, ok := target.(interface{ RawAttrs() map[string]string })
	if !ok {
		return
	}
	rawAttrs := rawCarrier.RawAttrs()
	if len(rawAttrs) == 0 {
		return
	}
	matched := false
	if e, ok := any(target).(HasAttrs); ok {
		pseudoScope.EachPseudoRuleForWidget(target, resolver, func(rule *Rule) {
			matched = true
			applyAttrs(target, e, rule.Attrs)
		})
		if matched {
			applyAttrs(target, e, rawAttrs)
		}
	}
}

func (doc *Doc) endElement(_ xml.EndElement) {
	if rules, ok := doc.current().(*Rules); ok && rules.parseErr != nil {
		doc.parseErr = rules.parseErr
	}
	doc.pop()
}

func (doc *Doc) charData(data xml.CharData) {
	if sector, ok := doc.current().(*StdSector); ok {
		if strings.TrimSpace(string(data)) != "" && doc.parseErr == nil {
			doc.parseErr = fmt.Errorf("%s: <sector> cannot directly contain text; wrap it in <label>", sector.Path())
		}
		return
	}
	if widget, ok := doc.current().(HasText); ok {
		widget.AddText(string(data))
	}
}

func (doc *Doc) comment(comment xml.Comment) {
	if widget, ok := doc.current().(HasComment); ok {
		widget.AddComment(string(comment))
	}
}

func (doc *Doc) push(value any) {
	doc.stack = append(doc.stack, value)
	if scope, ok := value.(HasScope); ok {
		scope.SetParentScope(doc.scope())
		frame := scopeFrame{scope: scope}
		if owner, ok := value.(scopeResourceOwner); ok {
			frame.owner = owner
			if scoped, ok := value.(WantsScope); ok {
				scoped.SetScope(scope)
			}
		}
		doc.scopeFrames = append(doc.scopeFrames, frame)
	}
}

func (doc *Doc) pop() (value any) {
	if len(doc.stack) > 0 {
		value, doc.stack = doc.stack[len(doc.stack)-1], doc.stack[:len(doc.stack)-1]
		if _, ok := value.(HasScope); ok {
			doc.scopeFrames = doc.scopeFrames[:len(doc.scopeFrames)-1]
		}
	}
	return
}

func (doc *Doc) captureScopeOwnerAttrs(value any, attrLayers []scopeAttrLayer) {
	if _, ok := value.(HasScope); !ok || len(doc.scopeFrames) == 0 {
		return
	}
	frame := &doc.scopeFrames[len(doc.scopeFrames)-1]
	if frame.owner == nil {
		return
	}
	frame.attrLayers = attrLayers
}

func (doc *Doc) refreshScopeOwnerResources() {
	if len(doc.scopeFrames) == 0 {
		return
	}
	frame := &doc.scopeFrames[len(doc.scopeFrames)-1]
	if frame.owner == nil || len(frame.attrLayers) == 0 {
		return
	}
	frame.owner.resetResourceAttrs()
	for _, layer := range frame.attrLayers {
		frame.owner.setResourceAttrs(layer.attrs, layer.units)
	}
}

func (doc *Doc) current() (value any) {
	if len(doc.stack) > 0 {
		value = doc.stack[len(doc.stack)-1]
	}
	return
}

func (doc *Doc) scope() HasScope {
	if len(doc.scopeFrames) > 0 {
		return doc.scopeFrames[len(doc.scopeFrames)-1].scope
	}
	// Return the per-document root scope, which inherits from defaultScope.
	// This ensures concurrent documents can carry different asset filesystems.
	if doc.rootScope.parent == nil {
		doc.rootScope.SetParentScope(&defaultScope)
	}
	return &doc.rootScope
}

func Parse(b []byte, opts ...ParseOption) (*Doc, error) {
	doc := newDocWithOptions(opts...)
	return doc, doc.parseBytes(b)
}

func ParseFile(filename string, opts ...ParseOption) (*Doc, error) {
	doc := newDocWithOptions(opts...)
	return doc, doc.parseFile(filename)
}

func ParseReader(r io.Reader, opts ...ParseOption) (*Doc, error) {
	doc := newDocWithOptions(opts...)
	return doc, doc.parseReader(r)
}

func traceStartElement(elem xml.StartElement) {
	debugf("StartElement %s:%s\n", elem.Name.Space, elem.Name.Local)
	for _, attr := range elem.Attr {
		debugf("%s=%s\n", attr.Name.Local, attr.Value)
	}
}

func traceEndElement(elem xml.EndElement) {
	debugf("EndElement %s:%s\n", elem.Name.Space, elem.Name.Local)
}

func traceCharData(data xml.CharData) {
	if text := strings.TrimSpace(string(data)); text != "" {
		debugf("%s\n", text)
	}
}

func traceComment(comment xml.Comment) {
	debugf("%s\n", string(comment))
}
