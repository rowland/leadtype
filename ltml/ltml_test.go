// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/rowland/leadtype/afm_fonts"
	"github.com/rowland/leadtype/pdf"
	"github.com/rowland/leadtype/ttf_fonts"
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

type DocWriter struct {
	*pdf.DocWriter
}

func (dw *DocWriter) NewPage() {
	dw.DocWriter.NewPage()
}

func (dw *DocWriter) SetLineDashPattern(pattern string) {
	dw.DocWriter.SetLineDashPattern(pattern)
}

func (dw *DocWriter) SetLineWidth(width float64) {
	dw.DocWriter.SetLineWidth(width, "pt")
}

func NewDocWriter() *DocWriter {
	dw := pdf.NewDocWriter()
	ttFonts, err := ttf_fonts.New("/Library/Fonts/*.ttf")
	if err != nil {
		panic(err)
	}
	dw.AddFontSource(ttFonts)

	afmFonts, err := afm_fonts.New("../afm/data/fonts/*.afm")
	if err != nil {
		panic(err)
	}
	dw.AddFontSource(afmFonts)

	return &DocWriter{dw}
}

func sampleFile(filename string) string {
	_, file, _, _ := runtime.Caller(0)
	dir, _ := filepath.Abs(filepath.Dir(file))
	sample := filepath.Join(dir, "samples", filename)
	return sample
}

func writeSample(name string, t *testing.T) {
	doc, err := ParseFile(sampleFile(name + ".ltml"))
	if err != nil {
		t.Fatal(err)
	}

	f, err := os.Create(sampleFile(name + ".pdf"))
	if err != nil {
		t.Fatal(err)
	}

	w := NewDocWriter()

	if err := doc.Print(w); err != nil {
		t.Errorf("Printing sample: %v", err)
	}

	w.WriteTo(f)
	f.Close()
	exec.Command("open", sampleFile(name+".pdf")).Start()
}

func TestSample001(t *testing.T) {
	writeSample("test_001_empty_doc", t)
}

func TestSample002(t *testing.T) {
	writeSample("test_002_empty_page", t)
}

func TestSample003(t *testing.T) {
	writeSample("test_003_hello_world", t)
}

func TestSample004(t *testing.T) {
	writeSample("test_004_two_pages", t)
}

func TestSample005(t *testing.T) {
	writeSample("test_005_rounded_rect", t)
}

func TestSample006(t *testing.T) {
	writeSample("test_006_bullets", t)
}

func TestSample007(t *testing.T) {
	writeSample("test_007_flow_layout", t)
}

func TestSample008(t *testing.T) {
	writeSample("test_008_vbox_layout", t)
}

func TestSample009(t *testing.T) {
	writeSample("test_009_hbox_layout", t)
}

func TestSample010(t *testing.T) {
	writeSample("test_010_rich_text", t)
}
