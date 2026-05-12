package ltml

import (
	"os"
	"strings"
	"testing"

	"github.com/rowland/leadtype/ltml/ltpdf"
)

const benchmarkImageCount = 500

var benchmarkImageAssets = []string{
	"pdf/testdata/testimg.jpg",
	"pdf/testdata/eidetic.png",
	"pdf/testdata/test_scene.svg",
	"pdf/testdata/test_scene_svg_advanced_gradients.svg",
	"pdf/testdata/test_scene_svg_text_clip.svg",
	"pdf/testdata/test_scene_svg_text_gradient_clip.svg",
	"ltml/samples/test_034_scene.svg",
}

func BenchmarkImageLayout_FlowColumns(b *testing.B) {
	benchmarkImageLayout(b, benchmarkImageLayoutFlowColumns())
}

func BenchmarkImageLayout_FlowHBoxRows(b *testing.B) {
	benchmarkImageLayout(b, benchmarkImageLayoutHBoxRows("flow"))
}

func BenchmarkImageLayout_VBoxHBoxRows(b *testing.B) {
	benchmarkImageLayout(b, benchmarkImageLayoutHBoxRows("vbox"))
}

func benchmarkImageLayout(b *testing.B, input string) {
	b.Helper()
	assetFS := os.DirFS("..")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		doc, err := Parse([]byte(input), WithAssetFS(assetFS))
		if err != nil {
			b.Fatal(err)
		}
		w := ltpdf.NewDocWriter()
		b.StartTimer()
		if err := doc.Print(w); err != nil {
			b.Fatal(err)
		}
	}
}

func benchmarkImageLayoutFlowColumns() string {
	var b strings.Builder
	b.WriteString(`<ltml units="pt" margin="36"><page layout="flow" layout.hpadding="6" layout.vpadding="6">`)
	for i := 0; i < benchmarkImageCount; i++ {
		writeBenchmarkImage(&b, i)
	}
	b.WriteString(`</page></ltml>`)
	return b.String()
}

func benchmarkImageLayoutHBoxRows(pageLayout string) string {
	var b strings.Builder
	b.WriteString(`<ltml units="pt" margin="36"><page layout="`)
	b.WriteString(pageLayout)
	b.WriteString(`" layout.hpadding="6" layout.vpadding="6">`)
	for i := 0; i < benchmarkImageCount; i += 3 {
		b.WriteString(`<div layout="hbox" width="100%" layout.hpadding="6">`)
		for j := 0; j < 3 && i+j < benchmarkImageCount; j++ {
			writeBenchmarkImage(&b, i+j)
		}
		b.WriteString(`</div>`)
	}
	b.WriteString(`</page></ltml>`)
	return b.String()
}

func writeBenchmarkImage(b *strings.Builder, index int) {
	b.WriteString(`<image src="`)
	b.WriteString(benchmarkImageAssets[index%len(benchmarkImageAssets)])
	b.WriteString(`" width="160"/>`)
}
