package ltml

import (
	"bytes"
	"strings"
	"testing"

	"github.com/rowland/leadtype/afm_fonts"
	"github.com/rowland/leadtype/ltml/ltpdf"
	"github.com/rowland/leadtype/pdf"
)

func newLTMLPDFWriter(t *testing.T) *ltpdf.DocWriter {
	t.Helper()
	afm, err := afm_fonts.Default()
	if err != nil {
		t.Fatal(err)
	}
	w := &ltpdf.DocWriter{DocWriter: pdf.NewDocWriter()}
	w.AddFontSource(afm)
	return w
}

func pageTextsByNumber(w *labelTestWriter) map[int]string {
	result := make(map[int]string)
	for i, rt := range w.printed {
		result[w.printedPages[i]] += rt.String()
	}
	for i, text := range w.plainPrinted {
		result[w.plainPages[i]] += text
	}
	return result
}

func TestParse_LinkTargetAndIndexTags(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml>
  <page>
    <p><a target="intro">Jump</a></p>
    <target id="intro" />
    <index id="toc" />
    <index_entry index="toc" target="intro">Introduction</index_entry>
  </page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}

	page := doc.ltmls[0].Page(0)
	paragraph := page.children[0].(*StdParagraph)
	if len(paragraph.textPieces) != 1 {
		t.Fatalf("paragraph text piece count = %d, want 1", len(paragraph.textPieces))
	}
	link, ok := paragraph.textPieces[0].content.(linkedInlineText)
	if !ok {
		t.Fatalf("paragraph content = %T, want linkedInlineText", paragraph.textPieces[0].content)
	}
	if link.LinkTarget() != "intro" {
		t.Fatalf("link target = %q, want intro", link.LinkTarget())
	}
	if _, ok := page.children[1].(*StdTarget); !ok {
		t.Fatalf("second child = %T, want *StdTarget", page.children[1])
	}
	if _, ok := page.children[2].(*StdIndex); !ok {
		t.Fatalf("third child = %T, want *StdIndex", page.children[2])
	}
	entry, ok := page.children[3].(*StdIndexEntry)
	if !ok {
		t.Fatalf("fourth child = %T, want *StdIndexEntry", page.children[3])
	}
	if entry.indexID != "toc" || entry.target != "intro" {
		t.Fatalf("index entry attrs = %#v", entry)
	}
}

func TestStdDocument_Print_RejectsInvalidLinkAttrs(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml>
  <page><p><a>broken</a></p></page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}
	if err := doc.Print(&labelTestWriter{t: t}); err == nil || !strings.Contains(err.Error(), "exactly one") {
		t.Fatalf("Print error = %v, want invalid <a> error", err)
	}
}

func TestStdDocument_Print_RejectsMissingIndexDefinition(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml>
  <page>
    <index_entry index="toc" target="intro">Intro</index_entry>
    <label id="intro">Introduction</label>
  </page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}
	if err := doc.Print(&labelTestWriter{t: t}); err == nil || !strings.Contains(err.Error(), "missing index") {
		t.Fatalf("Print error = %v, want missing index error", err)
	}
}

func TestStdDocument_Print_RejectsMissingTarget(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml>
  <page><p><a target="missing">Go</a></p></page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}
	if err := doc.Print(&labelTestWriter{t: t}); err == nil || !strings.Contains(err.Error(), "missing internal target") {
		t.Fatalf("Print error = %v, want missing target error", err)
	}
}

func TestStdDocument_IndexRendersResolvedPageNumbers(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml units="in" margin="0.5">
  <page layout="vbox">
    <index id="toc" />
  </page>
  <page layout="vbox">
    <label id="intro">Introduction</label>
    <index_entry index="toc" target="intro">Introduction</index_entry>
  </page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}

	w := &labelTestWriter{t: t, lineSpacing: 1.0}
	if err := doc.Print(w); err != nil {
		t.Fatal(err)
	}
	pageTexts := pageTextsByNumber(w)
	if !strings.Contains(pageTexts[1], "Introduction") || !strings.Contains(pageTexts[1], "2") {
		t.Fatalf("page 1 text = %q, want index label and page 2 reference", pageTexts[1])
	}
}

func TestStdDocument_IndexRespectsPageNumberStart(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml units="in" margin="0.5">
  <page layout="vbox">
    <index id="toc" />
  </page>
  <page layout="vbox">
    <p><pageno start="10" hidden="true" />Body</p>
    <label id="chapter">Chapter</label>
    <index_entry index="toc" target="chapter">Chapter</index_entry>
  </page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}

	w := &labelTestWriter{t: t, lineSpacing: 1.0}
	if err := doc.Print(w); err != nil {
		t.Fatal(err)
	}
	pageTexts := pageTextsByNumber(w)
	if !strings.Contains(pageTexts[1], "Chapter") || !strings.Contains(pageTexts[1], "10") {
		t.Fatalf("page 1 text = %q, want page number 10 in index", pageTexts[1])
	}
}

func TestStdDocument_MultipleIndexesRenderIndependently(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml units="in" margin="0.5">
  <page layout="vbox">
    <index id="main" />
    <index id="sub" />
  </page>
  <page layout="vbox">
    <label id="chapter">Chapter</label>
    <index_entry index="main" target="chapter">Main Entry</index_entry>
    <index_entry index="sub" target="chapter">Sub Entry</index_entry>
  </page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}

	w := &labelTestWriter{t: t, lineSpacing: 1.0}
	if err := doc.Print(w); err != nil {
		t.Fatal(err)
	}
	pageTexts := pageTextsByNumber(w)
	if !strings.Contains(pageTexts[1], "Main Entry") || !strings.Contains(pageTexts[1], "Sub Entry") {
		t.Fatalf("page 1 text = %q, want both indexes to render", pageTexts[1])
	}
}

func TestStdDocument_IndexEntriesCreateInternalLinks(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml units="in" margin="0.5">
  <page layout="vbox">
    <index id="toc" />
  </page>
  <page layout="vbox">
    <label id="intro">Introduction</label>
    <index_entry index="toc" target="intro">Introduction</index_entry>
  </page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}

	w := newLTMLPDFWriter(t)
	if err := doc.Print(w); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if _, err := w.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	pdfText := buf.String()
	if count := strings.Count(pdfText, "/Subtype /Link"); count < 1 {
		t.Fatalf("index link annotation count = %d, want at least 1", count)
	}
	if count := strings.Count(pdfText, "/S /GoTo"); count < 1 {
		t.Fatalf("index GoTo action count = %d, want at least 1", count)
	}
}

func TestLTML_LinkedParagraphWrapsIntoMultiplePDFAnnotations(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml units="pt" margin="36">
  <page layout="vbox">
    <p width="60"><a uri="https://example.com">One two three four five six</a></p>
  </page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}

	w := newLTMLPDFWriter(t)
	if err := doc.Print(w); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if _, err := w.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	if count := strings.Count(buf.String(), "/Subtype /Link"); count < 2 {
		t.Fatalf("link annotation count = %d, want at least 2", count)
	}
}
