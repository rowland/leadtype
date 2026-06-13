// Copyright 2026 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

// Package profile provides lightweight, opt-in render profiling helpers.
package profile

import (
	"fmt"
	"io"
	"sort"
	"time"
)

// Profiler records nested timing spans for one render.
//
// A nil *Profiler is the disabled state. Callers can keep hot paths cheap by
// checking for nil before creating labels, while Begin and Span.End are also
// nil-safe for convenience.
type Profiler struct {
	clock func() time.Time
	stats map[string]*Entry
	stack []activeSpan
}

// Entry is the aggregated timing data for a span label.
type Entry struct {
	Label string
	Count int
	Total time.Duration
	Self  time.Duration
	Max   time.Duration
}

type activeSpan struct {
	label string
	start time.Time
	child time.Duration
}

// Span is an active timing span returned by Profiler.Begin.
type Span struct {
	profiler *Profiler
	depth    int
}

// New returns an empty profiler.
func New() *Profiler {
	return &Profiler{
		clock: time.Now,
		stats: make(map[string]*Entry),
	}
}

// Begin starts a timing span with label. The returned Span must be ended by the
// caller. Begin is nil-safe: calling it on a nil profiler returns a no-op span.
func (p *Profiler) Begin(label string) Span {
	if p == nil || label == "" {
		return Span{}
	}
	if p.clock == nil {
		p.clock = time.Now
	}
	if p.stats == nil {
		p.stats = make(map[string]*Entry)
	}
	p.stack = append(p.stack, activeSpan{label: label, start: p.clock()})
	return Span{profiler: p, depth: len(p.stack)}
}

// End closes the span and records its inclusive, exclusive, and max duration.
// End is nil-safe and ignores spans that are no longer at the top of the stack.
func (s Span) End() {
	p := s.profiler
	if p == nil || s.depth == 0 || len(p.stack) != s.depth {
		return
	}
	now := p.clock()
	active := p.stack[len(p.stack)-1]
	p.stack = p.stack[:len(p.stack)-1]

	total := now.Sub(active.start)
	self := total - active.child
	if self < 0 {
		self = 0
	}
	entry := p.stats[active.label]
	if entry == nil {
		entry = &Entry{Label: active.label}
		p.stats[active.label] = entry
	}
	entry.Count++
	entry.Total += total
	entry.Self += self
	if total > entry.Max {
		entry.Max = total
	}
	if len(p.stack) > 0 {
		p.stack[len(p.stack)-1].child += total
	}
}

// Entries returns a stable snapshot sorted by descending total time, then label.
func (p *Profiler) Entries() []Entry {
	if p == nil || len(p.stats) == 0 {
		return nil
	}
	entries := make([]Entry, 0, len(p.stats))
	for _, entry := range p.stats {
		entries = append(entries, *entry)
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Total != entries[j].Total {
			return entries[i].Total > entries[j].Total
		}
		return entries[i].Label < entries[j].Label
	})
	return entries
}

// WriteText writes a human-readable summary of collected timing entries.
func (p *Profiler) WriteText(w io.Writer) error {
	if w == nil {
		return nil
	}
	if _, err := fmt.Fprintln(w, "leadtype profile:"); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "%-48s %8s %12s %12s %12s\n", "label", "count", "total", "self", "max"); err != nil {
		return err
	}
	for _, entry := range p.Entries() {
		if _, err := fmt.Fprintf(w, "%-48s %8d %12s %12s %12s\n",
			entry.Label, entry.Count, entry.Total, entry.Self, entry.Max); err != nil {
			return err
		}
	}
	return nil
}
