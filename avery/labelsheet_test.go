package avery

import (
	"strings"
	"testing"

	"github.com/rowland/leadtype/ltml"
)

type testWriter struct {
	ltml.NoopWriter
	rectCount int
	printed   []string
}

func (w *testWriter) Print(text string) error {
	w.printed = append(w.printed, text)
	return nil
}

func (w *testWriter) Rectangle2(float64, float64, float64, float64, bool, bool, []float64, bool, bool) {
	w.rectCount++
}
func TestLookupStock_CanonicalAndAlias(t *testing.T) {
	canonical, ok := LookupStock("avery5160")
	if !ok {
		t.Fatal("avery5160 stock not found")
	}
	alias, ok := LookupStock("8160")
	if !ok {
		t.Fatal("8160 alias not found")
	}
	if canonical.ID != alias.ID {
		t.Fatalf("alias resolved to %q, want %q", alias.ID, canonical.ID)
	}
}

func TestLookupStock_Geometry(t *testing.T) {
	tests := []struct {
		id          string
		labelWidth  float64
		labelHeight float64
		columns     int
		rows        int
	}{
		{id: "5160", labelWidth: 2.625 * 72, labelHeight: 1.0 * 72, columns: 3, rows: 10},
		{id: "5163", labelWidth: 4.0 * 72, labelHeight: 2.0 * 72, columns: 2, rows: 5},
		{id: "5167", labelWidth: 1.75 * 72, labelHeight: 0.5 * 72, columns: 4, rows: 20},
		{id: "5395", labelWidth: 3.375 * 72, labelHeight: (7.0 / 3.0) * 72, columns: 2, rows: 4},
	}
	for _, tt := range tests {
		stock, ok := LookupStock(tt.id)
		if !ok {
			t.Fatalf("%s stock not found", tt.id)
		}
		if got := stock.Columns; got != tt.columns {
			t.Fatalf("%s columns = %d, want %d", tt.id, got, tt.columns)
		}
		if got := stock.Rows; got != tt.rows {
			t.Fatalf("%s rows = %d, want %d", tt.id, got, tt.rows)
		}
		if got := stock.LabelWidth; got != tt.labelWidth {
			t.Fatalf("%s label width = %v, want %v", tt.id, got, tt.labelWidth)
		}
		if got := stock.LabelHeight; got != tt.labelHeight {
			t.Fatalf("%s label height = %v, want %v", tt.id, got, tt.labelHeight)
		}
	}
}

func TestLabelSheet_PrintDrawsExpectedCellCount(t *testing.T) {
	doc, err := ltml.Parse([]byte(`
		<ltml xmlns:avery="avery">
			<page style="letter" margin="0">
				<avery:labelsheet stock="5160" show-metrics="true" show-outline="true"/>
			</page>
		</ltml>`))
	if err != nil {
		t.Fatal(err)
	}

	w := &testWriter{}
	if err := doc.Print(w); err != nil {
		t.Fatal(err)
	}
	if got := w.rectCount; got != 30 {
		t.Fatalf("rect count = %d, want 30", got)
	}
	if len(w.printed) == 0 || !strings.Contains(w.printed[0], "Avery stock: avery5160") {
		t.Fatalf("first printed text = %q, want avery metrics", firstPrinted(w.printed))
	}
}

func TestLabelSheet_LaysOutLabelsByRows(t *testing.T) {
	doc, err := ltml.Parse([]byte(`
		<ltml xmlns:avery="avery">
			<page style="letter" margin="0">
				<avery:labelsheet stock="5160" order="rows">
					<avery:label>One</avery:label>
					<avery:label>Two</avery:label>
					<avery:label>Three</avery:label>
					<avery:label>Four</avery:label>
				</avery:labelsheet>
			</page>
		</ltml>`))
	if err != nil {
		t.Fatal(err)
	}

	w := &testWriter{}
	if err := doc.Print(w); err != nil {
		t.Fatal(err)
	}

	page := doc.Page(0)
	sheet, ok := page.Widgets()[0].(*LabelSheet)
	if !ok {
		t.Fatalf("page child type = %T, want *LabelSheet", page.Widgets()[0])
	}
	if len(sheet.Widgets()) != 4 {
		t.Fatalf("label count = %d, want 4", len(sheet.Widgets()))
	}
	first := sheet.Widgets()[0]
	second := sheet.Widgets()[1]
	third := sheet.Widgets()[2]
	fourth := sheet.Widgets()[3]

	if first.Left() != 13.5 || first.Top() != 36 {
		t.Fatalf("first label at (%v,%v), want (13.5,36)", first.Left(), first.Top())
	}
	if second.Left() != 211.5 || second.Top() != 36 {
		t.Fatalf("second label at (%v,%v), want (211.5,36)", second.Left(), second.Top())
	}
	if third.Left() != 409.5 || third.Top() != 36 {
		t.Fatalf("third label at (%v,%v), want (409.5,36)", third.Left(), third.Top())
	}
	if fourth.Left() != 13.5 || fourth.Top() != 108 {
		t.Fatalf("fourth label at (%v,%v), want (13.5,108)", fourth.Left(), fourth.Top())
	}
}

func TestLabelSheet_LaysOutLabelsByCols(t *testing.T) {
	doc, err := ltml.Parse([]byte(`
		<ltml xmlns:avery="avery">
			<page style="letter" margin="0">
				<avery:labelsheet stock="5163" order="cols">
					<avery:label>One</avery:label>
					<avery:label>Two</avery:label>
					<avery:label>Three</avery:label>
				</avery:labelsheet>
			</page>
		</ltml>`))
	if err != nil {
		t.Fatal(err)
	}

	w := &testWriter{}
	if err := doc.Print(w); err != nil {
		t.Fatal(err)
	}

	page := doc.Page(0)
	sheet := page.Widgets()[0].(*LabelSheet)
	first := sheet.Widgets()[0]
	second := sheet.Widgets()[1]
	third := sheet.Widgets()[2]

	if first.Left() != 13.5 || first.Top() != 36 {
		t.Fatalf("first label at (%v,%v), want (13.5,36)", first.Left(), first.Top())
	}
	if second.Left() != 13.5 || second.Top() != 180 {
		t.Fatalf("second label at (%v,%v), want (13.5,180)", second.Left(), second.Top())
	}
	if third.Left() != 13.5 || third.Top() != 324 {
		t.Fatalf("third label at (%v,%v), want (13.5,324)", third.Left(), third.Top())
	}
}

func TestLabelSheet_PrintUnknownStockFailsClearly(t *testing.T) {
	doc, err := ltml.Parse([]byte(`
		<ltml xmlns:avery="avery">
			<page style="letter" margin="0">
				<avery:labelsheet stock="bogus"/>
			</page>
		</ltml>`))
	if err != nil {
		t.Fatal(err)
	}
	w := &testWriter{}
	err = doc.Print(w)
	if err == nil || !strings.Contains(err.Error(), `unknown avery labelsheet stock "bogus"`) {
		t.Fatalf("print err = %v, want unknown stock error", err)
	}
}

func TestLabelSheet_RejectsTooManyLabels(t *testing.T) {
	doc, err := ltml.Parse([]byte(`
		<ltml xmlns:avery="avery">
			<page style="letter" margin="0">
				<avery:labelsheet stock="5395">
					<avery:label>1</avery:label>
					<avery:label>2</avery:label>
					<avery:label>3</avery:label>
					<avery:label>4</avery:label>
					<avery:label>5</avery:label>
					<avery:label>6</avery:label>
					<avery:label>7</avery:label>
					<avery:label>8</avery:label>
					<avery:label>9</avery:label>
				</avery:labelsheet>
			</page>
		</ltml>`))
	if err != nil {
		t.Fatal(err)
	}
	w := &testWriter{}
	err = doc.Print(w)
	if err == nil || !strings.Contains(err.Error(), `avery labelsheet stock "avery5395" holds 8 labels, got 9`) {
		t.Fatalf("print err = %v, want overflow error", err)
	}
}

func firstPrinted(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}
