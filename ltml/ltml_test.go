// Copyright 2016, 2017 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"bytes"
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/rowland/leadtype/ltml/ltpdf"
	"github.com/rowland/leadtype/ttf_fonts"
)

var openSamplePDFs = flag.Bool("open-sample-pdfs", false, "open generated LTML sample PDFs after tests run")

func TestParse(t *testing.T) {
	t.Run("Bytes", func(t *testing.T) {
		doc, err := Parse([]byte("<ltml></ltml>"))
		if err != nil {
			t.Fatal(err)
		}
		if doc == nil {
			t.Error("doc is nil")
		}
	})

	t.Run("File", func(t *testing.T) {
		doc, err := ParseFile(sampleFile("test_001_empty_doc.ltml"))
		if err != nil {
			t.Fatal(err)
		}
		if doc == nil {
			t.Error("doc is nil")
		}
	})

	t.Run("Reader", func(t *testing.T) {
		r := bytes.NewReader([]byte("<ltml></ltml>"))
		doc, err := ParseReader(r)
		if err != nil {
			t.Fatal(err)
		}
		if doc == nil {
			t.Error("doc is nil")
		}
	})
}

func sampleFile(filename string) string {
	_, file, _, _ := runtime.Caller(0)
	dir, _ := filepath.Abs(filepath.Dir(file))
	sample := filepath.Join(dir, "samples", filename)
	return sample
}

func writeSamplePDF(name string, t *testing.T) {
	t.Helper()

	doc, err := ParseFile(sampleFile(name + ".ltml"))
	if err != nil {
		t.Fatal(err)
	}

	f, err := os.Create(sampleFile(name + ".pdf"))
	if err != nil {
		t.Fatal(err)
	}

	w := ltpdf.NewDocWriter()
	ttFonts, err := ttf_fonts.NewFromSystemFonts()
	if err != nil {
		panic(err)
	}
	w.AddFontSource(ttFonts)

	if err := doc.Print(w); err != nil {
		t.Errorf("Printing sample: %v", err)
	}

	w.WriteTo(f)
	f.Close()
	if *openSamplePDFs {
		exec.Command("open", sampleFile(name+".pdf")).Start()
	}
}

func TestSamples(t *testing.T) {
	samples := []string{
		"test_001_empty_doc",
		"test_002_empty_page",
		"test_003_hello_world",
		"test_004_two_pages",
		"test_005_rounded_rect",
		"test_006_bullets",
		"test_007_flow_layout",
		"test_008_vbox_layout",
		"test_009_hbox_layout",
		"test_010_rich_text",
		"test_011_table_layout",
		"test_012_cjk_thai_grid",
		"test_013_font_fallback",
		"test_014_label_br",
		"test_015_pre",
		"test_016_image",
		"test_017_line",
		"test_018_shapes",
		"test_019_pageno",
		"test_020_positioning",
		"test_021_compression",
		"test_030_encodings",
	}

	for _, sample := range samples {
		sample := sample
		t.Run(sample, func(t *testing.T) {
			writeSamplePDF(sample, t)
		})
	}
}
