package ltml

import (
	"strings"
	"testing"

	"github.com/rowland/leadtype/ltml/ltpdf"
)

const (
	benchmarkParagraphPages = 20
	benchmarkParagraphRows  = 12
)

func BenchmarkParagraphLayout_NestedVBoxHBox(b *testing.B) {
	input := benchmarkNestedParagraphLayout()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		doc, err := Parse([]byte(input))
		if err != nil {
			b.Fatal(err)
		}
		writer := ltpdf.NewDocWriter()
		b.StartTimer()
		if err := doc.Print(writer); err != nil {
			b.Fatal(err)
		}
	}
}

func benchmarkNestedParagraphLayout() string {
	var input strings.Builder
	input.WriteString(`<ltml units="pt"><font id="body" name="Helvetica" size="11pt"/>`)
	for range benchmarkParagraphPages {
		input.WriteString(`<page layout="vbox" margin="36" font="body" layout.vpadding="4">`)
		input.WriteString(`<vbox width="100%" layout.vpadding="3">`)
		for range benchmarkParagraphRows {
			input.WriteString(`<vbox width="100%" layout="vbox" layout.vpadding="2">`)
			input.WriteString(`<p width="100%">A representative paragraph with enough words to exercise wrapping and repeated preferred-height probes.</p>`)
			input.WriteString(`<hbox width="100%" layout.hpadding="8">`)
			input.WriteString(`<vbox width="auto"><p>Nested left-side explanatory text that wraps naturally.</p></vbox>`)
			input.WriteString(`<vbox width="120"><p>Nested right-side detail text.</p></vbox>`)
			input.WriteString(`</hbox></vbox>`)
		}
		input.WriteString(`</vbox></page>`)
	}
	input.WriteString(`</ltml>`)
	return input.String()
}
