package ltml

import (
	"bytes"
	"strings"
	"testing"

	"github.com/rowland/leadtype/afm_fonts"
	"github.com/rowland/leadtype/font"
	"github.com/rowland/leadtype/ltml/ltpdf"
	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/pdf"
	"github.com/rowland/leadtype/rich_text"
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

func indexedTestFonts(t *testing.T) []*font.Font {
	t.Helper()

	afm, err := afm_fonts.Default()
	if err != nil {
		t.Fatal(err)
	}
	makeFont := func(opts options.Options) *font.Font {
		t.Helper()
		face, err := font.New("Helvetica", opts, font.FontSources{afm})
		if err != nil {
			t.Fatal(err)
		}
		return face
	}
	return []*font.Font{
		makeFont(options.Options{"size": 12.0}),
		makeFont(options.Options{"size": 12.0, "weight": "Bold"}),
		makeFont(options.Options{"size": 12.0, "style": "Italic"}),
	}
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

func printedRichTextByPage(w *labelTestWriter, pageNo int) []*rich_text.RichText {
	var result []*rich_text.RichText
	for i, rt := range w.printed {
		if w.printedPages[i] == pageNo {
			result = append(result, rt)
		}
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
    <index-entry index="toc" target="intro">Introduction</index-entry>
  </page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}

	page := doc.Root().Page(0)
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

func TestParse_IndexTemplatePlaceholderTags(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml>
  <page>
    <index id="toc">
      <p width="100%"><index-title font.weight="Bold" /><index-leader /><index-page /></p>
    </index>
  </page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}

	page := doc.Root().Page(0)
	index := page.children[0].(*StdIndex)
	if len(index.children) != 1 {
		t.Fatalf("index child count = %d, want 1", len(index.children))
	}
	row, ok := index.children[0].(*StdParagraph)
	if !ok {
		t.Fatalf("index row child = %T, want *StdParagraph", index.children[0])
	}
	if len(row.textPieces) != 3 {
		t.Fatalf("row text piece count = %d, want 3", len(row.textPieces))
	}
	title, ok := row.textPieces[0].content.(*StdIndexTitle)
	if !ok {
		t.Fatalf("first row piece = %T, want *StdIndexTitle", row.textPieces[0].content)
	}
	if title.Font().weight != "Bold" {
		t.Fatalf("index-title font weight = %q, want Bold", title.Font().weight)
	}
	leader, ok := row.textPieces[1].content.(*StdLeader)
	if !ok {
		t.Fatalf("second row piece = %T, want *StdLeader", row.textPieces[1].content)
	}
	if leader.LeaderText() != "." {
		t.Fatalf("leader text = %q, want .", leader.LeaderText())
	}
	if _, ok := row.textPieces[2].content.(*StdIndexPage); !ok {
		t.Fatalf("third row piece = %T, want *StdIndexPage", row.textPieces[2].content)
	}
}

func TestParse_LeaderTagInParagraph(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml>
  <page>
    <p>Left<leader text="-" />Right</p>
  </page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}

	page := doc.Root().Page(0)
	paragraph := page.children[0].(*StdParagraph)
	if len(paragraph.textPieces) != 3 {
		t.Fatalf("paragraph text piece count = %d, want 3", len(paragraph.textPieces))
	}
	leader, ok := paragraph.textPieces[1].content.(*StdLeader)
	if !ok {
		t.Fatalf("middle paragraph piece = %T, want *StdLeader", paragraph.textPieces[1].content)
	}
	if leader.LeaderText() != "-" {
		t.Fatalf("leader text = %q, want -", leader.LeaderText())
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
    <index-entry index="toc" target="intro">Intro</index-entry>
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
    <index-entry index="toc" target="intro">Introduction</index-entry>
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

func TestStdDocument_LegacyEmptyIndexUsesDottedLeader(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml units="pt" margin="36">
  <page layout="vbox">
    <index id="toc" width="140" />
  </page>
  <page layout="vbox">
    <label id="intro">Introduction</label>
    <index-entry index="toc" target="intro">Introduction</index-entry>
  </page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}

	w := &labelTestWriter{t: t, fonts: indexedTestFonts(t), lineSpacing: 1.0}
	if err := doc.Print(w); err != nil {
		t.Fatal(err)
	}
	page1 := printedRichTextByPage(w, 1)
	if len(page1) < 1 {
		t.Fatalf("page 1 printed count = %d, want at least 1 line", len(page1))
	}
	if page1[0].String() == "" || !strings.Contains(page1[0].String(), "Introduction") || strings.Trim(strings.TrimPrefix(page1[0].String(), "Introduction"), ".12 ") == page1[0].String() {
		t.Fatalf("page 1 line = %q, want introduction with dotted leader and page number", page1[0].String())
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
    <index-entry index="toc" target="chapter">Chapter</index-entry>
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
    <index-entry index="main" target="chapter">Main Entry</index-entry>
    <index-entry index="sub" target="chapter">Sub Entry</index-entry>
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
    <index-entry index="toc" target="intro">Introduction</index-entry>
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

func TestStdDocument_Print_RejectsMultipleIndexTemplateChildren(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml units="pt" margin="36">
  <page layout="vbox">
    <index id="toc">
      <p>One</p>
      <p>Two</p>
    </index>
  </page>
  <page layout="vbox">
    <label id="intro">Introduction</label>
    <index-entry index="toc" target="intro">Introduction</index-entry>
  </page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}
	if err := doc.Print(&labelTestWriter{t: t, fonts: defaultTestFonts(t), lineSpacing: 1.0}); err == nil || !strings.Contains(err.Error(), "exactly one block template child") {
		t.Fatalf("Print error = %v, want invalid <index> template child count error", err)
	}
}

func TestStdDocument_IndexTemplateRendersWrappedTitleWithPageOnFinalLine(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml units="pt" margin="36">
  <page layout="vbox">
    <index id="toc" width="90">
      <p width="100%"><index-title /><index-leader /><index-page /></p>
    </index>
  </page>
  <page layout="vbox">
    <label id="intro">A very long introduction heading for wrapping</label>
    <index-entry index="toc" target="intro">A very long introduction heading for wrapping</index-entry>
  </page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}

	w := &labelTestWriter{t: t, fonts: indexedTestFonts(t), lineSpacing: 1.0}
	if err := doc.Print(w); err != nil {
		t.Fatal(err)
	}
	page1 := printedRichTextByPage(w, 1)
	if len(page1) < 2 {
		t.Fatalf("page 1 printed count = %d, want at least 2 wrapped lines", len(page1))
	}
	if strings.Contains(page1[0].String(), "2") || strings.Contains(page1[0].String(), ".") {
		t.Fatalf("page 1 first line = %q, want plain wrapped title without leader/page", page1[0].String())
	}
	last := page1[len(page1)-1].String()
	if !strings.Contains(last, "2") {
		t.Fatalf("page 1 last line = %q, want page number on final line", last)
	}
	if !strings.Contains(last, ".") {
		t.Fatalf("page 1 last line = %q, want leader on final line", last)
	}
	pageTexts := pageTextsByNumber(w)
	if !strings.Contains(pageTexts[1], "A very long") || !strings.Contains(pageTexts[1], "introduction") || !strings.Contains(pageTexts[1], "wrapping") {
		t.Fatalf("page 1 text = %q, want wrapped title fragments", pageTexts[1])
	}
}

func TestStdDocument_IndexTemplateFontOverridesApplyPerPlaceholder(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml units="pt" margin="36">
  <page layout="vbox">
    <index id="toc">
      <p width="100%"><index-title font.weight="Bold" /><index-leader /><index-page font.style="Italic" /></p>
    </index>
  </page>
  <page layout="vbox">
    <label id="chapter">Chapter</label>
    <index-entry index="toc" target="chapter">Chapter</index-entry>
  </page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}

	w := &labelTestWriter{t: t, fonts: indexedTestFonts(t), lineSpacing: 1.0}
	if err := doc.Print(w); err != nil {
		t.Fatal(err)
	}
	index := doc.Root().Page(0).children[0].(*StdIndex)
	if got, want := len(index.Widgets()), 1; got != want {
		t.Fatalf("expanded index child count = %d, want %d", got, want)
	}
	row, ok := index.Widgets()[0].(*StdParagraph)
	if !ok {
		t.Fatalf("expanded index row = %T, want *StdParagraph", index.Widgets()[0])
	}
	titleFont, _ := row.textPieces[0].Font(row.Font())
	pageFont, _ := row.textPieces[2].Font(row.Font())
	if titleFont == nil || pageFont == nil {
		t.Fatalf("expanded placeholder fonts = title:%v page:%v", titleFont, pageFont)
	}
	if titleFont.weight != "Bold" {
		t.Fatalf("title font weight = %q, want Bold", titleFont.weight)
	}
	if pageFont.style != "Italic" {
		t.Fatalf("page font style = %q, want Italic", pageFont.style)
	}
}

func TestStdDocument_LeaderWorksOutsideIndexes(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml units="pt" margin="36">
  <page layout="vbox">
    <p width="120">Alpha<leader text="-" />Omega</p>
    <label width="120">Beta<leader text="*" />Gamma</label>
  </page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}

	w := &labelTestWriter{t: t, fonts: defaultTestFonts(t), lineSpacing: 1.0}
	if err := doc.Print(w); err != nil {
		t.Fatal(err)
	}
	page1 := printedRichTextByPage(w, 1)
	if len(page1) < 2 {
		t.Fatalf("page 1 printed count = %d, want at least 2 lines", len(page1))
	}
	if !strings.Contains(page1[0].String(), "Alpha") || !strings.Contains(page1[0].String(), "Omega") || !strings.Contains(page1[0].String(), "-") {
		t.Fatalf("paragraph leader line = %q, want combined Alpha/leader/Omega output", page1[0].String())
	}
	if !strings.Contains(page1[1].String(), "Beta") || !strings.Contains(page1[1].String(), "Gamma") || !strings.Contains(page1[1].String(), "*") {
		t.Fatalf("label leader line = %q, want combined Beta/leader/Gamma output", page1[1].String())
	}
}

func TestStdDocument_IndexHonorsLayoutVPaddingAndReservesHeight(t *testing.T) {
	docWithGap, err := Parse([]byte(`
<ltml units="pt" margin="36">
  <page layout="vbox">
    <index id="toc" layout.vpadding="12" width="140">
      <p width="100%"><index-title /><index-leader /><index-page /></p>
    </index>
    <p>After index</p>
  </page>
  <page layout="vbox">
    <label id="intro">Introduction</label>
    <index-entry index="toc" target="intro">Introduction</index-entry>
  </page>
  <page layout="vbox">
    <label id="appendix">Appendix A</label>
    <index-entry index="toc" target="appendix">Appendix A</index-entry>
  </page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}

	docWithoutGap, err := Parse([]byte(`
<ltml units="pt" margin="36">
  <page layout="vbox">
    <index id="toc" width="140">
      <p width="100%"><index-title /><index-leader /><index-page /></p>
    </index>
    <p>After index</p>
  </page>
  <page layout="vbox">
    <label id="intro">Introduction</label>
    <index-entry index="toc" target="intro">Introduction</index-entry>
  </page>
  <page layout="vbox">
    <label id="appendix">Appendix A</label>
    <index-entry index="toc" target="appendix">Appendix A</index-entry>
  </page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}

	w := &labelTestWriter{t: t, fonts: defaultTestFonts(t), lineSpacing: 1.0}
	if err := docWithGap.Print(w); err != nil {
		t.Fatal(err)
	}
	w2 := &labelTestWriter{t: t, fonts: defaultTestFonts(t), lineSpacing: 1.0}
	if err := docWithoutGap.Print(w2); err != nil {
		t.Fatal(err)
	}
	indexWithGap := docWithGap.Root().Page(0).children[0].(*StdIndex)
	indexWithoutGap := docWithoutGap.Root().Page(0).children[0].(*StdIndex)
	if got, want := len(indexWithGap.Widgets()), 2; got != want {
		t.Fatalf("index with gap child count = %d, want %d", got, want)
	}
	if got, want := len(indexWithoutGap.Widgets()), 2; got != want {
		t.Fatalf("index without gap child count = %d, want %d", got, want)
	}
	if got := indexWithGap.Widgets()[1].Top() - indexWithoutGap.Widgets()[1].Top(); got < 11.99 || got > 12.01 {
		t.Fatalf("second entry baseline delta = %v, want 12pt", got)
	}
	afterWithGap := docWithGap.Root().Page(0).children[1]
	afterWithoutGap := docWithoutGap.Root().Page(0).children[1]
	if got := afterWithGap.Top() - afterWithoutGap.Top(); got < 11.99 || got > 12.01 {
		t.Fatalf("post-index paragraph baseline delta = %v, want 12pt reserved below index", got)
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
