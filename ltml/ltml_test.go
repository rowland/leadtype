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
	"sync"
	"testing"

	"github.com/rowland/leadtype/afm_fonts"
	"github.com/rowland/leadtype/ltml/ltpdf"
	"github.com/rowland/leadtype/pdf"
	"github.com/rowland/leadtype/ttf_fonts"
)

var openSamplePDFs = flag.Bool("open-sample-pdfs", false, "open generated LTML sample PDFs after tests run")

var (
	sampleFontSourcesOnce sync.Once
	sampleTTFonts         *ttf_fonts.TtfFonts
	sampleAFMFonts        *afm_fonts.AfmFonts
	sampleFontSourcesErr  error
)

func loadSampleFontSources() (*ttf_fonts.TtfFonts, *afm_fonts.AfmFonts, error) {
	sampleFontSourcesOnce.Do(func() {
		sampleTTFonts, sampleFontSourcesErr = ttf_fonts.NewFromSystemFonts()
		if sampleFontSourcesErr != nil {
			return
		}
		sampleAFMFonts, sampleFontSourcesErr = afm_fonts.Default()
	})
	return sampleTTFonts, sampleAFMFonts, sampleFontSourcesErr
}

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

func writeSamplePDF(name string, outputFile string, t *testing.T) {
	t.Helper()

	doc, err := ParseFile(sampleFile(name + ".ltml"))
	if err != nil {
		t.Fatal(err)
	}

	f, err := os.Create(outputFile)
	if err != nil {
		t.Fatal(err)
	}

	ttFonts, afmFonts, err := loadSampleFontSources()
	if err != nil {
		t.Fatal(err)
	}
	w := &ltpdf.DocWriter{DocWriter: pdf.NewDocWriter()}
	w.AddFontSource(ttFonts)
	w.AddFontSource(afmFonts)

	if err := doc.Print(w); err != nil {
		t.Errorf("Printing sample: %v", err)
	}

	w.WriteTo(f)
	f.Close()
	if *openSamplePDFs {
		exec.Command("open", outputFile).Start()
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
		"test_022_transforms",
		"test_023_zindex",
		"test_024_vbox_overflow",
		"test_025_flow_overflow",
		"test_026_table_overflow",
		"test_027_paragraph_split",
		"test_028_table_split_headers",
		"test_029_table_split_headers_footers",
		"test_030_encodings",
		"test_032_label_shrink_to_fit",
	}

	for _, sample := range samples {
		sample := sample
		t.Run(sample, func(t *testing.T) {
			outputFile := filepath.Join(t.TempDir(), sample+".pdf")
			writeSamplePDF(sample, outputFile, t)
		})
	}
}
