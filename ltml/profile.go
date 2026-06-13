// Copyright 2026 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
	"reflect"

	"github.com/rowland/leadtype/profile"
)

type profilerProvider interface {
	Profiler() *profile.Profiler
}

type profilerSetter interface {
	SetProfiler(*profile.Profiler)
}

func profilerForWriter(w Writer) *profile.Profiler {
	if provider, ok := any(w).(profilerProvider); ok {
		return provider.Profiler()
	}
	return nil
}

func setWriterProfiler(w Writer, profiler *profile.Profiler) {
	if setter, ok := any(w).(profilerSetter); ok {
		setter.SetProfiler(profiler)
	}
}

func profilerForWidget(w Writer, widget Widget) *profile.Profiler {
	if profiler := profilerForWriter(w); profiler != nil {
		return profiler
	}
	if doc, ok := widget.(*StdDocument); ok {
		return doc.renderProfiler
	}
	containerWidget, ok := widget.(interface{ Container() Container })
	if !ok {
		return nil
	}
	if doc := documentForContainer(containerWidget.Container()); doc != nil {
		return doc.renderProfiler
	}
	return nil
}

func profilerForContainer(w Writer, container Container) *profile.Profiler {
	if profiler := profilerForWriter(w); profiler != nil {
		return profiler
	}
	if doc, ok := container.(*StdDocument); ok {
		return doc.renderProfiler
	}
	if doc := documentForContainer(container); doc != nil {
		return doc.renderProfiler
	}
	return nil
}

func beginWidgetProfileSpan(profiler *profile.Profiler, prefix string, widget Widget) profile.Span {
	if profiler == nil {
		return profile.Span{}
	}
	return profiler.Begin(widgetProfileLabel(prefix, widget))
}

func widgetProfileLabel(prefix string, widget Widget) string {
	return fmt.Sprintf("ltml.widget.%s.%s", prefix, widgetTypeName(widget))
}

func widgetTypeName(widget Widget) string {
	if widget == nil {
		return "<nil>"
	}
	t := reflect.TypeOf(widget)
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t.Name()
}
