// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"bytes"
	"path/filepath"
	"runtime"
	"testing"
)

func TestParse(t *testing.T) {
	doc, err := Parse([]byte("<ltml></ltml>"))
	if err != nil {
		t.Fatal(err)
	}
	if doc == nil {
		t.Error("doc is nil")
	}
}

func TestParseFile(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	dir, _ := filepath.Abs(filepath.Dir(file))
	sample := filepath.Join(dir, "samples", "test_001_empty_doc.ltml")
	doc, err := ParseFile(sample)
	if err != nil {
		t.Fatal(err)
	}
	if doc == nil {
		t.Error("doc is nil")
	}
}

func TestParseReader(t *testing.T) {
	r := bytes.NewReader([]byte("<ltml></ltml>"))
	doc, err := ParseReader(r)
	if err != nil {
		t.Fatal(err)
	}
	if doc == nil {
		t.Error("doc is nil")
	}
}

func sampleFile(filename string) string {
	_, file, _, _ := runtime.Caller(0)
	dir, _ := filepath.Abs(filepath.Dir(file))
	sample := filepath.Join(dir, "samples", filename)
	return sample
}

func TestSample001(t *testing.T) {
	doc, err := ParseFile(sampleFile("test_001_empty_doc.ltml"))
	if err != nil {
		t.Fatal(err)
	}
	doc.Print(nil)
}

func TestSample002(t *testing.T) {
	doc, err := ParseFile(sampleFile("test_002_empty_page.ltml"))
	if err != nil {
		t.Fatal(err)
	}
	doc.Print(nil)
}

func TestSample003(t *testing.T) {
	doc, err := ParseFile(sampleFile("test_003_hello_world.ltml"))
	if err != nil {
		t.Fatal(err)
	}
	doc.Print(nil)
}

func TestSample004(t *testing.T) {
	doc, err := ParseFile(sampleFile("test_004_two_pages.ltml"))
	if err != nil {
		t.Fatal(err)
	}
	doc.Print(nil)
}

func TestSample005(t *testing.T) {
	doc, err := ParseFile(sampleFile("test_005_rounded_rect.ltml"))
	if err != nil {
		t.Fatal(err)
	}
	doc.Print(nil)
}

func TestSample006(t *testing.T) {
	doc, err := ParseFile(sampleFile("test_006_bullets.ltml"))
	if err != nil {
		t.Fatal(err)
	}
	doc.Print(nil)
}

func TestSample007(t *testing.T) {
	doc, err := ParseFile(sampleFile("test_007_flow_layout.ltml"))
	if err != nil {
		t.Fatal(err)
	}
	doc.Print(nil)
}

func TestSample008(t *testing.T) {
	doc, err := ParseFile(sampleFile("test_008_vbox_layout.ltml"))
	if err != nil {
		t.Fatal(err)
	}
	doc.Print(nil)
}

func TestSample009(t *testing.T) {
	doc, err := ParseFile(sampleFile("test_009_hbox_layout.ltml"))
	if err != nil {
		t.Fatal(err)
	}
	doc.Print(nil)
}

func TestSample010(t *testing.T) {
	doc, err := ParseFile(sampleFile("test_010_rich_text.ltml"))
	if err != nil {
		t.Fatal(err)
	}
	doc.Print(nil)
}
