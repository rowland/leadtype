package ltml_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/rowland/leadtype/afm_fonts"
	_ "github.com/rowland/leadtype/avery"
	"github.com/rowland/leadtype/font"
	"github.com/rowland/leadtype/ltml"
	"github.com/rowland/leadtype/options"
)

func TestSample_PaperCatalogsRender(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	samples := []string{
		"test_051_paper_sizes.ltml",
		"test_052_large_paper_sizes.ltml",
		"test_053_avery_labels.ltml",
	}
	for _, name := range samples {
		name := name
		t.Run(name, func(t *testing.T) {
			sample := filepath.Join(filepath.Dir(file), "samples", name)
			doc, err := ltml.ParseFile(sample)
			if err != nil {
				t.Fatal(err)
			}
			w := &sampleTestWriter{t: t}
			if err := doc.Print(w); err != nil {
				t.Fatal(err)
			}
		})
	}
}

type sampleTestWriter struct {
	ltml.NoopWriter
	fonts    []*font.Font
	fontSize float64
	t        testing.TB
}

func (w *sampleTestWriter) ensureFonts() []*font.Font {
	if len(w.fonts) != 0 {
		return w.fonts
	}
	fontSource, err := afm_fonts.Default()
	if err != nil {
		w.t.Fatal(err)
	}
	face, err := font.New("Helvetica", options.Options{"size": 12.0}, font.FontSources{fontSource})
	if err != nil {
		w.t.Fatal(err)
	}
	w.fonts = []*font.Font{face}
	return w.fonts
}

func (w *sampleTestWriter) Fonts() []*font.Font { return w.ensureFonts() }

func (w *sampleTestWriter) FontSize() float64 {
	if w.fontSize == 0 {
		return 12
	}
	return w.fontSize
}

func (w *sampleTestWriter) AddFont(_ string, _ options.Options) ([]*font.Font, error) {
	return w.ensureFonts(), nil
}

func (w *sampleTestWriter) SetFont(_ string, size float64, _ options.Options) ([]*font.Font, error) {
	w.fontSize = size
	return w.ensureFonts(), nil
}
