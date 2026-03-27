package ltml

import (
	"testing"

	"github.com/rowland/leadtype/pdf"
)

type compressionTestWriter struct {
	labelTestWriter
	compressPages         bool
	compressToUnicode     bool
	compressEmbeddedFonts bool
}

func (w *compressionTestWriter) CompressPages(value bool) *pdf.DocWriter {
	w.compressPages = value
	return nil
}

func (w *compressionTestWriter) CompressToUnicode(value bool) *pdf.DocWriter {
	w.compressToUnicode = value
	return nil
}

func (w *compressionTestWriter) CompressEmbeddedFonts(value bool) *pdf.DocWriter {
	w.compressEmbeddedFonts = value
	return nil
}

func TestStdDocument_Print_AppliesCompressionAttrs(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml compress-pages="true" compress-to-unicode="true" compress-embedded-fonts="true">
  <page><label>Hello</label></page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}

	w := &compressionTestWriter{labelTestWriter: labelTestWriter{t: t}}
	if err := doc.Print(w); err != nil {
		t.Fatal(err)
	}
	if !w.compressPages || !w.compressToUnicode || !w.compressEmbeddedFonts {
		t.Fatalf("compression flags = pages:%t toUnicode:%t embedded:%t, want all true",
			w.compressPages, w.compressToUnicode, w.compressEmbeddedFonts)
	}
}

func TestStdDocument_Print_DefaultCompressionAttrsFalse(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml>
  <page><label>Hello</label></page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}

	w := &compressionTestWriter{labelTestWriter: labelTestWriter{t: t}}
	if err := doc.Print(w); err != nil {
		t.Fatal(err)
	}
	if w.compressPages || w.compressToUnicode || w.compressEmbeddedFonts {
		t.Fatalf("compression flags = pages:%t toUnicode:%t embedded:%t, want all false",
			w.compressPages, w.compressToUnicode, w.compressEmbeddedFonts)
	}
}
