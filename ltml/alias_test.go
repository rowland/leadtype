// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"strings"
	"sync"
	"testing"
)

type aliasCascadeWidget struct {
	StdWidget
	value string
}

func (w *aliasCascadeWidget) DefaultAttrs(HasScope) map[string]string {
	return map[string]string{"value": "default"}
}

func (w *aliasCascadeWidget) SetAttrs(attrs map[string]string) {
	w.StdWidget.SetAttrs(attrs)
	if value, ok := attrs["value"]; ok {
		w.value = value
	}
}

var registerAliasCascadeWidgetOnce sync.Once

func registerAliasCascadeWidget(t *testing.T) {
	t.Helper()
	registerAliasCascadeWidgetOnce.Do(func() {
		if err := RegisterTag("alias-cascade", "card", func() any { return &aliasCascadeWidget{} }); err != nil {
			t.Fatalf("register alias cascade widget: %v", err)
		}
	})
}

func TestParse_QualifiedDefineTargetPreservesAttributePrecedence(t *testing.T) {
	registerAliasCascadeWidget(t)

	doc, err := Parse([]byte(`
<ltml xmlns:ac="alias-cascade">
  <define id="defined-card" tag="alias-cascade:card" value="define" />
  <define id="styled-card" tag="alias-cascade:card" value="define" />
  <define id="instance-card" tag="alias-cascade:card" value="define" />
  <style>
    styled-card { value: style; }
    instance-card { value: style; }
  </style>
  <page>
    <ac:card />
    <defined-card />
    <styled-card />
    <instance-card value="instance" />
  </page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}

	widgets := doc.Page(0).Widgets()
	if len(widgets) != 4 {
		t.Fatalf("widgets = %d, want 4", len(widgets))
	}
	want := []string{"default", "define", "style", "instance"}
	for i, widget := range widgets {
		card, ok := widget.(*aliasCascadeWidget)
		if !ok {
			t.Fatalf("widget %d = %T, want *aliasCascadeWidget", i, widget)
		}
		if card.value != want[i] {
			t.Fatalf("widget %d value = %q, want %q", i, card.value, want[i])
		}
	}
}

func TestParse_RejectsInvalidQualifiedDefineTargets(t *testing.T) {
	tests := []string{
		":card",
		"alias-cascade:",
		"bar:baz:boom",
		"bad namespace:card",
		"alias-cascade:bad.tag",
	}
	for _, target := range tests {
		t.Run(target, func(t *testing.T) {
			_, err := Parse([]byte(`<ltml><define id="bad" tag="` + target + `" /></ltml>`))
			if err == nil {
				t.Fatalf("Parse() error = nil, want invalid define target error")
			}
			if !strings.Contains(err.Error(), `expected "tag" or "namespace:tag"`) {
				t.Fatalf("Parse() error = %q", err)
			}
		})
	}
}

var _ HasDefaultAttrs = (*aliasCascadeWidget)(nil)
var _ HasAttrs = (*aliasCascadeWidget)(nil)
