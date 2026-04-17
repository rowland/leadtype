// Copyright 2016, 2017 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"bytes"
	"flag"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"testing/fstest"

	"github.com/rowland/leadtype/afm_fonts"
	"github.com/rowland/leadtype/internal/overlayfs"
	"github.com/rowland/leadtype/ltml/ltpdf"
	"github.com/rowland/leadtype/pdf"
	"github.com/rowland/leadtype/ttf_fonts"
)

var openSamplePDFs = flag.Bool("open-sample-pdfs", false, "open generated LTML sample PDFs after tests run")
var writeSamplePDFs = flag.Bool("write-sample-pdfs", false, "write LTML sample PDFs beside their .ltml files instead of to temporary test output")

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
		// Include the in-repo Arabic fixture font so LTML sample regeneration
		// does not depend on system Arabic font quality or TTC shaping support.
		if err := sampleTTFonts.Add(filepath.Join(filepath.Dir(sampleFile("placeholder")), "..", "shaping", "testdata", "Amiri-Regular.ttf")); err != nil {
			sampleFontSourcesErr = err
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

	t.Run("WithAssetFS", func(t *testing.T) {
		doc, err := Parse([]byte(`
<ltml xmlns:xt="componenttest">
  <page>
    <xt:card src="snippet.xml" />
  </page>
</ltml>`), WithAssetFS(fstest.MapFS{
			"snippet.xml": &fstest.MapFile{Data: []byte("<p>from parse option</p>")},
		}))
		if err != nil {
			t.Fatal(err)
		}
		if doc == nil {
			t.Fatal("doc is nil")
		}
		page := doc.Page(0)
		component, ok := page.Widgets()[0].(*testComponent)
		if !ok {
			t.Fatalf("first child = %T, want *testComponent", page.Widgets()[0])
		}
		if got, want := component.Body(), "<p>from parse option</p>"; got != want {
			t.Fatalf("body = %q, want %q", got, want)
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

	if assetFS, err := sampleAssetFS(); err != nil {
		t.Fatal(err)
	} else {
		doc, err := ParseFile(sampleFile(name + ".ltml"))
		if err != nil && assetFS != nil {
			doc, err = ParseFile(sampleFile(name+".ltml"), WithAssetFS(assetFS))
		}
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
		if assetFS != nil {
			w.SetAssetFS(assetFS)
		}
		// Tagged PDF accessibility is document-driven for LTML samples. Add
		// ua="true" to the root <ltml> element to opt a sample in, or call
		// w.EnableTaggedPDF(true) here to force tagged output for the whole suite.
		if tagged, ok := os.LookupEnv("LTML_UA"); ok && tagged == "true" {
			w.EnableTaggedPDF(tagged == "true")
		}

		if err := doc.Print(w); err != nil {
			t.Errorf("Printing sample: %v", err)
		}

		w.WriteTo(f)
		f.Close()
		if *openSamplePDFs {
			exec.Command("open", outputFile).Start()
		}
		return
	}
}

func sampleOutputFile(name string, t *testing.T) string {
	t.Helper()
	if *writeSamplePDFs {
		return sampleFile(name + ".pdf")
	}
	if !*openSamplePDFs {
		return filepath.Join(t.TempDir(), name+".pdf")
	}
	dir, err := os.MkdirTemp("", "leadtype-ltml-samples-")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("opened sample PDFs are being written to %s", dir)
	return filepath.Join(dir, name+".pdf")
}

func sampleAssetFS() (fs.FS, error) {
	root := filepath.Dir(sampleFile("placeholder"))
	repoRoot := filepath.Join(root, "..", "..")
	absRepoRoot, err := filepath.Abs(repoRoot)
	if err != nil {
		return nil, err
	}
	assetDir := filepath.Join(absRepoRoot, "pdf", "testdata")
	return overlayfs.New(os.DirFS(assetDir), os.DirFS(absRepoRoot)), nil
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
		"test_031_render_ltml_images",
		"test_032_label_shrink_to_fit",
		"test_034_svg_image",
		"test_035_links",
		"test_036_indexes",
		"test_037_relative_layout",
		"test_038_radial_layout",
		"test_039_radial_out",
		"test_040_radial_sweep_star_wars",
		"test_041_pseudo_classes",
		"test_042_svg_tag",
		"test_043_svg_advanced",
		"test_044_decoration_overrides",
		"test_045_widget_brush_backgrounds",
		"test_046_text_fill_clipping",
		"test_047_style_src",
		"test_048_bullet_variants",
	}

	for _, sample := range samples {
		sample := sample
		t.Run(sample, func(t *testing.T) {
			outputFile := sampleOutputFile(sample, t)
			writeSamplePDF(sample, outputFile, t)
		})
	}
}
