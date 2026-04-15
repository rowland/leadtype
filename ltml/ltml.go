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
)

type Doc struct {
	root         *StdDocument
	stack        []any
	scopes       []HasScope
	rootScope    Scope // per-document root scope; parent = &defaultScope
	parseErr     error
	assetFS      fs.FS
	sourceDir    string
	assetSources *assetSourceManager
}

type ParseOption func(*Doc)

func WithAssetFS(assetFS fs.FS) ParseOption {
	return func(doc *Doc) {
		doc.SetAssetFS(assetFS)
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
	doc.applyPseudoRules()
	return nil
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
	trueTag := elem.Name.Local
	var defaultAttrs map[string]string
	if elem.Name.Space == DefaultSpace {
		if alias, ok := doc.scope().AliasFor(trueTag); ok {
			trueTag, defaultAttrs = alias.Tag, alias.Attrs
			debugf("Alias %s=%s %v\n", elem.Name.Local, trueTag, defaultAttrs)
		}
	}
	e := makeElement(elem.Name.Space, trueTag)
	if e == nil {
		debugf("Unknown tag: %s:%s\n", elem.Name.Space, elem.Name.Local)
	}
	capturesBody := isComponentTag(elem.Name.Space, trueTag)
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
		parent = parentCurrent
		if widget, ok := e.(Widget); ok {
			if isRadialLayoutStyle(parent.LayoutStyle()) {
				if _, isSector := widget.(*StdSector); !isSector {
					wrapper = &StdSector{}
					if ws, ok := any(wrapper).(WantsScope); ok {
						ws.SetScope(doc.scope())
					}
					parentCurrent.AddChild(wrapper)
					if wc, ok := any(wrapper).(WantsContainer); ok {
						if err = wc.SetContainer(parentCurrent); err != nil {
							debugf("Setting radial wrapper container: %s\n", err)
						}
					}
					wrapper.AddChild(widget)
					if wc, ok := e.(WantsContainer); ok {
						if err = wc.SetContainer(wrapper); err != nil {
							debugf("Setting wrapped container: %s\n", err)
						}
					}
				} else {
					parentCurrent.AddChild(widget)
					if wc, ok := e.(WantsContainer); ok {
						if err = wc.SetContainer(parentCurrent); err != nil {
							debugf("Setting container: %s\n", err)
						}
					}
				}
			} else {
				parentCurrent.AddChild(widget)
				if wc, ok := e.(WantsContainer); ok {
					if err = wc.SetContainer(parentCurrent); err != nil {
						debugf("Setting container: %s\n", err)
					}
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
	applyElementAttrs(doc.scope(), e, defaultAttrs, attrs, sourcePath)
	if wrapper != nil {
		applyElementAttrs(doc.scope(), wrapper, defaultAttrs, attrs, sourcePath)
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
				ps.SetAttrs("", attrs)
				style = ps
			case *PenStyle:
				ps := st.Clone()
				ps.SetAttrs(attrs)
				style = ps
			}
		}
		if err := doc.scope().AddStyle(style); err != nil {
			debugf("Adding style: %s\n", err)
		}
	}
	if layout, ok := e.(*LayoutStyle); ok {
		if layout0, ok := doc.scope().LayoutFor(layout.ID()); ok {
			layout = layout0.Clone()
			layout.SetAttrs(attrs)
		}
		if err := doc.scope().AddLayout(layout); err != nil {
			debugf("Adding layout: %s\n", err)
		}
	}
	if pageStyle, ok := e.(*PageStyle); ok {
		if err := doc.scope().AddPageStyle(pageStyle); err != nil {
			debugf("Adding page style: %s\n", err)
		}
	}
	if alias, ok := e.(*Alias); ok {
		if err := doc.scope().AddAlias(alias); err != nil {
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

// applyPseudoRules performs a second pass over the parsed widget tree so rules
// that rely on structural pseudo-classes can be matched with layout context.
func (doc *Doc) applyPseudoRules() {
	resolver := newSelectorStructureResolver()
	if doc.root != nil {
		doc.applyPseudoRulesToWidget(doc.root, resolver)
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
func applyElementAttrs(scope HasScope, target any, defaultAttrs, attrs map[string]string, pathOverride string) {
	if target == nil {
		return
	}
	if e, ok := target.(HasAttrs); ok {
		e.SetAttrs(defaultAttrs)
		path := pathOverride
		if path == "" {
			if p, ok := target.(HasPath); ok {
				path = p.Path()
			}
		}
		if path != "" {
			scope.EachRuleFor(path, func(rule *Rule) {
				e.SetAttrs(rule.Attrs)
			})
		}
		e.SetAttrs(attrs)
	}
	if e, ok := target.(HasAttrsPrefix); ok {
		e.SetAttrs("", defaultAttrs)
		path := pathOverride
		if path == "" {
			if p, ok := target.(HasPath); ok {
				path = p.Path()
			}
		}
		if path != "" {
			scope.EachRuleFor(path, func(rule *Rule) {
				e.SetAttrs("", rule.Attrs)
			})
		}
		e.SetAttrs("", attrs)
	}
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
			e.SetAttrs(rule.Attrs)
		})
		if matched {
			e.SetAttrs(rawAttrs)
		}
	}
	if e, ok := any(target).(HasAttrsPrefix); ok {
		prefixMatched := false
		pseudoScope.EachPseudoRuleForWidget(target, resolver, func(rule *Rule) {
			prefixMatched = true
			e.SetAttrs("", rule.Attrs)
		})
		if prefixMatched {
			e.SetAttrs("", rawAttrs)
		}
	}
}

func (doc *Doc) endElement(_ xml.EndElement) {
	doc.pop()
}

func (doc *Doc) charData(data xml.CharData) {
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
		doc.scopes = append(doc.scopes, scope)
	}
}

func (doc *Doc) pop() (value any) {
	if len(doc.stack) > 0 {
		value, doc.stack = doc.stack[len(doc.stack)-1], doc.stack[:len(doc.stack)-1]
		if _, ok := value.(HasScope); ok {
			doc.scopes = doc.scopes[:len(doc.scopes)-1]
		}
	}
	return
}

func (doc *Doc) current() (value any) {
	if len(doc.stack) > 0 {
		value = doc.stack[len(doc.stack)-1]
	}
	return
}

func (doc *Doc) scope() HasScope {
	if len(doc.scopes) > 0 {
		return doc.scopes[len(doc.scopes)-1]
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
