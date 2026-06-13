// Copyright 2026 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package profile

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestProfilerNestedSpans(t *testing.T) {
	now := time.Unix(0, 0)
	p := New()
	p.clock = func() time.Time { return now }

	outer := p.Begin("outer")
	now = now.Add(10 * time.Millisecond)
	inner := p.Begin("inner")
	now = now.Add(3 * time.Millisecond)
	inner.End()
	now = now.Add(2 * time.Millisecond)
	outer.End()

	entries := p.Entries()
	if len(entries) != 2 {
		t.Fatalf("entries = %d, want 2", len(entries))
	}
	if entries[0].Label != "outer" || entries[0].Count != 1 || entries[0].Total != 15*time.Millisecond || entries[0].Self != 12*time.Millisecond || entries[0].Max != 15*time.Millisecond {
		t.Fatalf("outer entry = %#v", entries[0])
	}
	if entries[1].Label != "inner" || entries[1].Count != 1 || entries[1].Total != 3*time.Millisecond || entries[1].Self != 3*time.Millisecond || entries[1].Max != 3*time.Millisecond {
		t.Fatalf("inner entry = %#v", entries[1])
	}
}

func TestProfilerEntriesSortByTotalThenLabel(t *testing.T) {
	now := time.Unix(0, 0)
	p := New()
	p.clock = func() time.Time { return now }

	a := p.Begin("beta")
	now = now.Add(time.Millisecond)
	a.End()
	b := p.Begin("alpha")
	now = now.Add(time.Millisecond)
	b.End()

	entries := p.Entries()
	if got := []string{entries[0].Label, entries[1].Label}; got[0] != "alpha" || got[1] != "beta" {
		t.Fatalf("entry order = %v, want alpha before beta for equal totals", got)
	}
}

func TestProfilerWriteText(t *testing.T) {
	now := time.Unix(0, 0)
	p := New()
	p.clock = func() time.Time { return now }
	span := p.Begin("render")
	now = now.Add(2 * time.Millisecond)
	span.End()

	var out bytes.Buffer
	if err := p.WriteText(&out); err != nil {
		t.Fatal(err)
	}
	text := out.String()
	for _, want := range []string{"leadtype profile:", "label", "render", "2ms"} {
		if !strings.Contains(text, want) {
			t.Fatalf("profile output %q missing %q", text, want)
		}
	}
}

func TestNilProfilerIsNoop(t *testing.T) {
	var p *Profiler
	span := p.Begin("ignored")
	span.End()
	if entries := p.Entries(); len(entries) != 0 {
		t.Fatalf("nil profiler entries = %#v", entries)
	}
}
