package pdf

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/rowland/leadtype/options"
)

func TestDocWriter_WriteTo_EmitsTaggedPDFStructure(t *testing.T) {
	dw := newAFMDocWriter(t)
	dw.EnableTaggedPDF(true)
	pw := dw.NewPage()
	if _, err := pw.SetFont("Helvetica", 12, options.Options{}); err != nil {
		t.Fatal(err)
	}

	if err := pw.WithAccessibilityTag("P", AccessibilityOptions{
		ActualText: "Accessible greeting",
		ID:         "paragraph-1",
	}, func() {
		pw.MoveTo(72, 720)
		_ = pw.Print("Hello")
	}); err != nil {
		t.Fatal(err)
	}

	if err := pw.WithAccessibilityTag("Link", AccessibilityOptions{ID: "link-1"}, func() {
		_ = pw.AddURILink(1, 1, 1, 0.25, "https://example.com")
	}); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if _, err := dw.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	pdfText := buf.String()

	for _, fragment := range []string{
		"%PDF-1.7\n",
		"/MarkInfo <<",
		"/Marked true",
		"/StructTreeRoot",
		"/Type /StructTreeRoot",
		"/ParentTree",
		"/Type /StructElem",
		"/S /P",
		"/ActualText (Accessible greeting)",
		"/Type /MCR",
		"/MCID 0",
		"/StructParents 0",
		"/Type /OBJR",
		"/StructParent 1",
	} {
		if !strings.Contains(pdfText, fragment) {
			t.Fatalf("tagged PDF output missing %q:\n%s", fragment, pdfText)
		}
	}
}

func TestDocWriter_WriteTo_EncodesNonASCIIDocumentActualTextAsUTF16BE(t *testing.T) {
	dw := newAFMDocWriter(t)
	dw.EnableTaggedPDF(true)
	pw := dw.NewPage()
	if _, err := pw.SetFont("Helvetica", 12, options.Options{}); err != nil {
		t.Fatal(err)
	}

	if err := pw.WithAccessibilityTag("P", AccessibilityOptions{
		ActualText: "cafe noir \u2615",
		ID:         "paragraph-1",
	}, func() {
		pw.MoveTo(72, 720)
		_ = pw.Print("Hello")
	}); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if _, err := dw.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	pdfText := buf.String()

	if !strings.Contains(pdfText, "/ActualText <feff00630061006600650020006e006f0069007200202615>") {
		t.Fatalf("expected UTF-16BE /ActualText in output:\n%s", pdfText)
	}
}

func TestPageWriter_WithAccessibilityArtifact_WrapsContent(t *testing.T) {
	dw := NewDocWriter()
	dw.EnableTaggedPDF(true)
	pw := dw.NewPage()

	if err := pw.WithAccessibilityArtifact(func() {
		pw.Rectangle(1, 1, 1, 0.5, true, false)
	}); err != nil {
		t.Fatal(err)
	}

	stream := pw.stream.String()
	if !strings.Contains(stream, "/Artifact BMC\n") || !strings.Contains(stream, "EMC\n") {
		t.Fatalf("expected artifact marked-content wrapper, got:\n%s", stream)
	}
}

func TestPageWriter_PrintImage_UsesCurrentAccessibilityTag(t *testing.T) {
	dw := NewDocWriter()
	dw.EnableTaggedPDF(true)
	pw := dw.NewPage()
	data, err := os.ReadFile("testdata/eidetic.png")
	if err != nil {
		t.Fatal(err)
	}

	if err := pw.WithAccessibilityTag("Figure", AccessibilityOptions{ID: "figure-1"}, func() {
		_, _, err = pw.PrintImage(data, 1, 1, nil, nil)
	}); err != nil {
		t.Fatal(err)
	}
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if _, err := dw.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	pdfText := buf.String()
	for _, fragment := range []string{
		"/S /Figure",
		"/Type /MCR",
		"/MCID 0",
		"/StructParents 0",
		"/Image",
	} {
		if !strings.Contains(pdfText, fragment) {
			t.Fatalf("expected figure tagging fragment %q in output:\n%s", fragment, pdfText)
		}
	}
}
