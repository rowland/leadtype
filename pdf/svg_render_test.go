// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import (
	"bytes"
	"strings"
	"testing"

	"github.com/rowland/leadtype/options"
)

func TestPageWriter_PrintSVG_TextFontFamilyQuotedName(t *testing.T) {
	msg := captureStderr(t, func() {
		dw := NewDocWriter()
		dw.AddFontSource(testFontSource(t, "../ttf/testdata/minimal.ttf"))
		pw := newPageWriter(dw, options.Options{"units": "pt"})
		data := []byte(`<svg width="80" height="20" xmlns="http://www.w3.org/2000/svg"><text x="10" y="12" font-family="&quot;'Minimal'&quot;" font-size="12">tiny</text></svg>`)
		width := 80.0
		if _, _, err := pw.PrintSVG(data, 0, 0, &width, nil); err != nil {
			t.Fatal(err)
		}
		pw.close()
		var buf bytes.Buffer
		if _, err := dw.WriteTo(&buf); err != nil {
			t.Fatal(err)
		}
	})
	if strings.Contains(msg, "font-family") {
		t.Fatalf("expected quoted font-family to resolve without warnings, got %q", msg)
	}
}

func TestPageWriter_PrintSVG_TextFontFamilyFallbackWarns(t *testing.T) {
	msg := captureStderr(t, func() {
		dw := NewDocWriter()
		dw.AddFontSource(testFontSource(t, "../ttf/testdata/minimal.ttf"))
		pw := newPageWriter(dw, options.Options{"units": "pt"})
		data := []byte(`<svg width="80" height="20" xmlns="http://www.w3.org/2000/svg"><text x="10" y="12" font-family="'Missing', Minimal" font-size="12">tiny</text></svg>`)
		width := 80.0
		if _, _, err := pw.PrintSVG(data, 0, 0, &width, nil); err != nil {
			t.Fatal(err)
		}
		pw.close()
		var buf bytes.Buffer
		if _, err := dw.WriteTo(&buf); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(msg, `font "Missing" unavailable; using fallback "Minimal"`) {
		t.Fatalf("expected fallback warning, got %q", msg)
	}
	if strings.Contains(msg, "fonts unavailable: tried") {
		t.Fatalf("expected successful fallback, got %q", msg)
	}
}

func TestPageWriter_PrintSVG_TextFontFamilyAllMissingWarnsOnce(t *testing.T) {
	msg := captureStderr(t, func() {
		dw := NewDocWriter()
		dw.AddFontSource(testFontSource(t, "../ttf/testdata/minimal.ttf"))
		pw := newPageWriter(dw, options.Options{"units": "pt"})
		data := []byte(`<svg width="80" height="20" xmlns="http://www.w3.org/2000/svg"><text x="10" y="12" font-family="'Missing', 'AlsoMissing'" font-size="12">tiny</text></svg>`)
		width := 80.0
		if _, _, err := pw.PrintSVG(data, 0, 0, &width, nil); err != nil {
			t.Fatal(err)
		}
		pw.close()
		var buf bytes.Buffer
		if _, err := dw.WriteTo(&buf); err != nil {
			t.Fatal(err)
		}
	})
	if count := strings.Count(msg, "svg: <text> font-family:"); count != 1 {
		t.Fatalf("warning count = %d, want 1; msg %q", count, msg)
	}
	if !strings.Contains(msg, `fonts unavailable: tried "Missing", "AlsoMissing"`) {
		t.Fatalf("expected consolidated missing-font warning, got %q", msg)
	}
}
