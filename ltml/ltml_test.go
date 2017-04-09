// Copyright 2016, 2017 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/rowland/leadtype/ltml/harupdf"
	"github.com/rowland/leadtype/ltml/ltpdf"
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

func sampleFile(filename string) string {
	_, file, _, _ := runtime.Caller(0)
	dir, _ := filepath.Abs(filepath.Dir(file))
	sample := filepath.Join(dir, "samples", filename)
	return sample
}

func writeSamplePDF(name string, t *testing.T) {
	doc, err := ParseFile(sampleFile(name + ".ltml"))
	if err != nil {
		t.Fatal(err)
	}

	f, err := os.Create(sampleFile(name + ".pdf"))
	if err != nil {
		t.Fatal(err)
	}

	w := ltpdf.NewDocWriter()
	ttFonts, err := ttf_fonts.New("/Library/Fonts/*.ttf")
	if err != nil {
		panic(err)
	}
	w.AddFontSource(ttFonts)

	if err := doc.Print(w); err != nil {
		t.Errorf("Printing sample: %v", err)
	}

	w.WriteTo(f)
	f.Close()
	exec.Command("open", sampleFile(name+".pdf")).Start()
}

func writeSampleHaru(name string, t *testing.T) {
	const suffix = "-haru"
	doc, err := ParseFile(sampleFile(name + ".ltml"))
	if err != nil {
		t.Fatal(err)
	}

	f, err := os.Create(sampleFile(name + suffix + ".pdf"))
	if err != nil {
		t.Fatal(err)
	}

	w := harupdf.NewDocWriter()
	ttFonts, err := ttf_fonts.New("/Library/Fonts/*.ttf")
	if err != nil {
		panic(err)
	}
	w.AddFontSource(ttFonts)

	if err := doc.Print(w); err != nil {
		t.Errorf("Printing sample: %v", err)
	}

	w.WriteTo(f)
	f.Close()
	exec.Command("open", sampleFile(name+suffix+".pdf")).Start()
}

func TestSample001(t *testing.T) {
	writeSamplePDF("test_001_empty_doc", t)
}

func TestSample002(t *testing.T) {
	writeSamplePDF("test_002_empty_page", t)
}

func TestSample003(t *testing.T) {
	writeSamplePDF("test_003_hello_world", t)
}

func TestSample004(t *testing.T) {
	writeSamplePDF("test_004_two_pages", t)
}

func TestSample005(t *testing.T) {
	writeSamplePDF("test_005_rounded_rect", t)
}

func TestSample006(t *testing.T) {
	writeSamplePDF("test_006_bullets", t)
}

func TestSample007(t *testing.T) {
	writeSamplePDF("test_007_flow_layout", t)
}

func TestSample008(t *testing.T) {
	writeSamplePDF("test_008_vbox_layout", t)
}

func TestSample009(t *testing.T) {
	writeSamplePDF("test_009_hbox_layout", t)
}

func TestSample010(t *testing.T) {
	writeSamplePDF("test_010_rich_text", t)
}

func TestSample011(t *testing.T) {
	writeSamplePDF("test_011_table_layout", t)
}

func TestSample030(t *testing.T) {
	writeSamplePDF("test_030_encodings", t)
	writeSampleHaru("test_030_encodings", t)
}
