// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"
)

type Doc struct {
	ltmls  []*StdDocument
	stack  []interface{}
	scopes []HasScope
}

func (doc *Doc) Parse(b []byte) error {
	r := bytes.NewReader(b)
	return doc.ParseReader(r)
}

func (doc *Doc) ParseFile(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return doc.ParseReader(f)
}

func (doc *Doc) ParseReader(r io.Reader) error {
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
			doc.startElement(t)
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
	}
	return nil
}

func (doc *Doc) Print(w Writer) (err error) {
	fmt.Println("Printing Doc")
	for _, ltml := range doc.ltmls {
		if err = ltml.Print(w); err != nil {
			return
		}
	}
	return nil
}

func (doc *Doc) startElement(elem xml.StartElement) {
	trueTag := elem.Name.Local
	var defaultAttrs map[string]string
	if elem.Name.Space == DefaultSpace {
		if alias, ok := doc.scope().Alias(trueTag); ok {
			trueTag, defaultAttrs = alias.Tag, alias.Attrs
			fmt.Fprintf(os.Stderr, "Alias %s=%s\n", elem.Name.Local, trueTag)
		}
	}
	e := makeElement(elem.Name.Space, trueTag)
	if e == nil {
		fmt.Fprintf(os.Stderr, "Unknown tag: %s:%s\n", elem.Name.Space, elem.Name.Local)
	}
	if ws, ok := e.(WantsScope); ok {
		ws.SetScope(doc.scope())
	}
	var err error
	if child, ok := e.(HasParent); ok {
		if err = child.SetParent(doc.current()); err != nil {
			fmt.Fprintf(os.Stderr, "Setting parent: %s\n", err)
		}
	}
	if parent, ok := doc.current().(Container); ok && err == nil {
		if p, ok := e.(Printer); ok {
			parent.AddChild(p)
		}
		if wc, ok := e.(WantsContainer); ok {
			if err = wc.SetContainer(parent); err != nil {
				fmt.Fprintf(os.Stderr, "Setting container: %s\n", err)
			}
		}
	}
	if d, ok := e.(*StdDocument); ok {
		doc.ltmls = append(doc.ltmls, d)
	}
	doc.push(e)

	attrs := mapFromXmlAttrs(elem.Attr)
	if attrs["tag"] == "" {
		attrs["tag"] = elem.Name.Local
	}

	if ident, ok := e.(Identifier); ok {
		ident.SetIentifiers(attrs)
	}
	if e, ok := e.(HasAttrs); ok {
		e.SetAttrs(defaultAttrs)
		// apply rules
		e.SetAttrs(attrs)
	}
	if style, ok := e.(Styler); ok {
		if err := doc.scope().AddStyle(style); err != nil {
			fmt.Fprintf(os.Stderr, "Adding style: %s\n", err)
		}
	}
	if layout, ok := e.(*LayoutStyle); ok {
		if layout0, ok := doc.scope().Layout(layout.ID()); ok {
			layout = layout0.Clone()
			layout.SetAttrs(attrs)
		}
		if err := doc.scope().AddLayout(layout); err != nil {
			fmt.Fprintf(os.Stderr, "Adding layout: %s\n", err)
		}
	}
	if alias, ok := e.(*Alias); ok {
		if err := doc.scope().AddAlias(alias); err != nil {
			fmt.Fprintf(os.Stderr, "Adding alias: %s\n", err)
		}
	}
}

func (doc *Doc) endElement(elem xml.EndElement) {
	doc.pop()
}

func (doc *Doc) charData(data xml.CharData) {
	if widget, ok := doc.current().(HasText); ok {
		widget.AddText(string(data))
	}
}

func (doc *Doc) comment(comment xml.Comment) {
}

func (doc *Doc) push(value interface{}) {
	doc.stack = append(doc.stack, value)
	if scope, ok := value.(HasScope); ok {
		scope.SetParentScope(doc.scope())
		doc.scopes = append(doc.scopes, scope)
	}
}

func (doc *Doc) pop() (value interface{}) {
	if len(doc.stack) > 0 {
		value, doc.stack = doc.stack[len(doc.stack)-1], doc.stack[:len(doc.stack)-1]
		if _, ok := value.(HasScope); ok {
			doc.scopes = doc.scopes[:len(doc.scopes)-1]
		}
	}
	return
}

func (doc *Doc) current() (value interface{}) {
	if len(doc.stack) > 0 {
		value = doc.stack[len(doc.stack)-1]
	}
	return
}

func (doc *Doc) scope() HasScope {
	if len(doc.scopes) > 0 {
		return doc.scopes[len(doc.scopes)-1]
	}
	return &defaultScope
}

func Parse(b []byte) (*Doc, error) {
	var doc Doc
	return &doc, doc.Parse(b)
}

func ParseFile(filename string) (*Doc, error) {
	var doc Doc
	return &doc, doc.ParseFile(filename)
}

func ParseReader(r io.Reader) (*Doc, error) {
	var doc Doc
	return &doc, doc.ParseReader(r)
}

func traceStartElement(elem xml.StartElement) {
	fmt.Printf("StartElement %s:%s\n", elem.Name.Space, elem.Name.Local)
	for _, attr := range elem.Attr {
		fmt.Printf("%s=%s\n", attr.Name.Local, attr.Value)
	}
}

func traceEndElement(elem xml.EndElement) {
	fmt.Printf("EndElement %s:%s\n", elem.Name.Space, elem.Name.Local)
}

func traceCharData(data xml.CharData) {
	fmt.Println("CharData")
	fmt.Println(strings.TrimSpace(string(data)))
}

func traceComment(comment xml.Comment) {
	fmt.Println("Comment")
	fmt.Println(string(comment))
}
