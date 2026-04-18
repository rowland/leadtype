package ltml

import (
	"strings"
	"testing"

	"github.com/rowland/leadtype/ltml/ltpdf"
)

func TestLayoutHBox_AutoHeightPanelsKeepAllNestedContentVisible(t *testing.T) {
	source := `
		<ltml units="in" margin="1">
			<layout id="vbox" padding="12pt" />
			<define id="panel" tag="div" border="thin" padding="10pt" layout="vbox" />
			<page layout="vbox" gap="12pt" font.name="Helvetica" font.size="12">
				<label font.weight="Bold" font.size="18">Built-in Heading Tags With rem Sizing</label>
				<p>
					This page uses a 12pt page font, so <b>1rem = 12pt</b>. The built-in
					heading aliases below expand to <b>&lt;label&gt;</b>, and each heading
					spells out its rem multiplier so the rendered scale is easy to verify.
				</p>
				<div layout="hbox" gap="12pt" width="100%">
					<panel width="50%" fill="AliceBlue">
						<label font.weight="Bold">Page-root scale</label>
						<p>Each alias below uses the same 12pt page root.</p>
						<h1>h1 — 2rem = 24pt</h1>
						<p>Largest level.</p>
						<h2>h2 — 1.75rem = 21pt</h2>
						<p>Strong section heading.</p>
						<h3>h3 — 1.5rem = 18pt</h3>
						<p>Common subheading size.</p>
						<h4>h4 — 1.25rem = 15pt</h4>
						<p>Just above body copy.</p>
						<h5>h5 — 1.125rem = 13.5pt</h5>
						<p>Matches the root size in bold.</p>
						<h6>h6 — 1rem = 12pt</h6>
						<p>Smallest custom heading.</p>
					</panel>
					<panel width="50%" fill="LemonChiffon" font.size="0.75rem">
						<label font.weight="Bold">Nested body at 0.75rem</label>
						<p>
							This panel reduces its inherited body text to <b>0.75rem = 9pt</b>.
							The headings below should still resolve against the 12pt page root.
						</p>
						<h2>Nested h2 stays 21pt</h2>
						<p>
							The body copy in this card is smaller, but the rem heading still
							resolves against the page root rather than the local panel font.
						</p>
						<h4>Nested h4 stays 15pt</h4>
						<p>
							This matches the left-column h4 even though the paragraph text is 9pt.
						</p>
						<h6>Nested h6 stays 12pt</h6>
						<p>
							The smallest heading still sits above the local body size, which
							makes the page-root behavior obvious at a glance.
						</p>
					</panel>
				</div>
			</page>
		</ltml>`
	doc := parseDoc(t, source)

	w := &labelTestWriter{t: t, lineSpacing: 1.0}
	if err := doc.Print(w); err != nil {
		t.Fatal(err)
	}

	var printed []string
	for _, rt := range w.printed {
		if rt == nil {
			continue
		}
		printed = append(printed, rt.String())
	}
	got := strings.Join(printed, "\n")
	for _, want := range []string{
		"h6 — 1rem = 12pt",
		"Nested h2 stays 21pt",
		"Nested h6 stays 12pt",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("printed output missing %q; got:\n%s", want, got)
		}
	}

	renderDoc := parseDoc(t, source)
	probeWriter := ltpdf.NewDocWriter()
	if err := renderDoc.Print(probeWriter); err != nil {
		t.Fatal(err)
	}
	page := firstPage(t, renderDoc)
	row, ok := page.children[2].(*StdContainer)
	if !ok {
		t.Fatalf("page child 2 is %T, want *StdContainer", page.children[2])
	}
	leftPanel, ok := row.children[0].(*StdContainer)
	if !ok {
		t.Fatalf("row child 0 is %T, want *StdContainer", row.children[0])
	}
	rightPanel, ok := row.children[1].(*StdContainer)
	if !ok {
		t.Fatalf("row child 1 is %T, want *StdContainer", row.children[1])
	}
	if !leftPanel.children[len(leftPanel.children)-1].Visible() {
		t.Fatalf("left panel last child should remain visible after real layout")
	}
	if !rightPanel.children[len(rightPanel.children)-1].Visible() {
		t.Fatalf("right panel last child should remain visible after real layout")
	}
}
