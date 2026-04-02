package pdf

import (
	"bytes"
	"strings"
	"testing"

	"github.com/rowland/leadtype/afm_fonts"
	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/rich_text"
	"github.com/rowland/leadtype/wordbreaking"
)

func newAFMDocWriter(t *testing.T) *DocWriter {
	t.Helper()
	fonts, err := afm_fonts.Default()
	if err != nil {
		t.Fatal(err)
	}
	dw := NewDocWriter()
	dw.AddFontSource(fonts)
	return dw
}

func TestLinkAnnotation_Write_URI(t *testing.T) {
	annot := newLinkAnnotation(1, 0, rectangle{x1: 10, y1: 20, x2: 30, y2: 40})
	annot.setURI("https://example.com")

	got := stringFromWriter(annot)
	for _, fragment := range []string{
		"/Subtype /Link",
		"/Type /Annot",
		"/Rect [10 20 30 40 ]",
		"/S /URI",
		"/URI (https://example.com)",
	} {
		if !strings.Contains(got, fragment) {
			t.Fatalf("annotation output missing %q:\n%s", fragment, got)
		}
	}
}

func TestDocWriter_WriteTo_RejectsUnresolvedTargetLinks(t *testing.T) {
	dw := newAFMDocWriter(t)
	pw := dw.NewPage()

	if err := pw.AddTargetLink(1, 1, 1, 0.25, "missing"); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if _, err := dw.WriteTo(&buf); err == nil || !strings.Contains(err.Error(), "missing") {
		t.Fatalf("WriteTo error = %v, want unresolved target error", err)
	}
}

func TestDocWriter_TargetLinkResolvesToDestination(t *testing.T) {
	dw := newAFMDocWriter(t)
	pw := dw.NewPage()
	pw.RegisterDestination("intro", 1, 1)
	if err := pw.AddTargetLink(1, 1.2, 1, 0.25, "intro"); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if _, err := dw.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	pdfText := buf.String()
	if !strings.Contains(pdfText, "/Annots [") || !strings.Contains(pdfText, "/S /GoTo") || !strings.Contains(pdfText, "/D [") {
		t.Fatalf("expected internal link destination in output:\n%s", pdfText)
	}
}

func TestPageWriter_PrintRichText_AddsLinkAnnotation(t *testing.T) {
	dw := newAFMDocWriter(t)
	pw := dw.NewPage()
	if _, err := pw.SetFont("Helvetica", 12, options.Options{}); err != nil {
		t.Fatal(err)
	}
	pw.MoveTo(72, 720)

	rt, err := rich_text.New("Visit Leadtype", pw.Fonts(), pw.FontSize(), options.Options{
		"link_uri": "https://example.com",
	})
	if err != nil {
		t.Fatal(err)
	}
	pw.PrintRichText(rt)

	var buf bytes.Buffer
	if _, err := dw.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	pdfText := buf.String()
	if !strings.Contains(pdfText, "/Subtype /Link") || !strings.Contains(pdfText, "/URI (https://example.com)") {
		t.Fatalf("expected link annotation in output:\n%s", pdfText)
	}
}

func TestPageWriter_PrintParagraph_SplitsLinkedAnnotationsAcrossLines(t *testing.T) {
	dw := newAFMDocWriter(t)
	pw := dw.NewPage()
	if _, err := pw.SetFont("Helvetica", 12, options.Options{}); err != nil {
		t.Fatal(err)
	}
	rt, err := rich_text.New("One two three four five six", pw.Fonts(), pw.FontSize(), options.Options{
		"link_uri": "https://example.com",
	})
	if err != nil {
		t.Fatal(err)
	}
	flags := make([]wordbreaking.Flags, rt.Len())
	wordbreaking.MarkRuneAttributes(rt.String(), flags)
	lines := rt.WrapToWidth(60, flags, false)

	pw.MoveTo(72, 720)
	pw.PrintParagraph(lines, options.Options{"width": 60})

	var buf bytes.Buffer
	if _, err := dw.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	if count := strings.Count(buf.String(), "/Subtype /Link"); count < 2 {
		t.Fatalf("link annotation count = %d, want at least 2", count)
	}
}
