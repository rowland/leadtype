package ltml

import (
	"math"
	"testing"
)

func TestPageStyle_SetAttrs_ID(t *testing.T) {
	ps := &PageStyle{}
	ps.SetAttrs(map[string]string{"id": "mypage", "width": "432", "height": "648"})
	if ps.ID() != "mypage" {
		t.Errorf("expected id %q, got %q", "mypage", ps.ID())
	}
	if ps.Width() != 432 {
		t.Errorf("expected width 432, got %v", ps.Width())
	}
	if ps.Height() != 648 {
		t.Errorf("expected height 648, got %v", ps.Height())
	}
}

func TestPageStyle_SetAttrs_Units(t *testing.T) {
	ps := &PageStyle{}
	ps.SetAttrs(map[string]string{"id": "book", "units": "in", "width": "6", "height": "9"})
	if ps.Width() != 6*72 {
		t.Errorf("expected width %v, got %v", 6*72.0, ps.Width())
	}
	if ps.Height() != 9*72 {
		t.Errorf("expected height %v, got %v", 9*72.0, ps.Height())
	}
}

func TestParsePageStyleTag_RegistersInDocumentScope(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml>
  <pagestyle id="book" units="in" width="6" height="9"/>
  <page style="book"></page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}
	ltml := doc.Root()
	ps, ok := ltml.PageStyleFor("book")
	if !ok {
		t.Fatal("page style 'book' not found in document scope")
	}
	if ps.Width() != 6*72 {
		t.Errorf("expected width %v, got %v", 6*72.0, ps.Width())
	}
	if ps.Height() != 9*72 {
		t.Errorf("expected height %v, got %v", 9*72.0, ps.Height())
	}
}

func TestParsePageStyleTag_PageUsesCustomStyle(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml>
  <pagestyle id="book" units="in" width="6" height="9"/>
  <page style="book"></page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}
	page := doc.Root().Page(0)
	if page == nil {
		t.Fatal("expected a page, got nil")
	}
	if page.Width() != 6*72 {
		t.Errorf("expected page width %v, got %v", 6*72.0, page.Width())
	}
	if page.Height() != 9*72 {
		t.Errorf("expected page height %v, got %v", 9*72.0, page.Height())
	}
}

func TestParsePageStyleTag_BuiltinStylesUnaffected(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml>
  <pagestyle id="book" units="in" width="6" height="9"/>
  <page style="letter"></page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}
	page := doc.Root().Page(0)
	if page == nil {
		t.Fatal("expected a page, got nil")
	}
	// letter = 612 x 792 pt
	if page.Width() != 612 {
		t.Errorf("expected letter width 612, got %v", page.Width())
	}
	if page.Height() != 792 {
		t.Errorf("expected letter height 792, got %v", page.Height())
	}
}

func TestParsePageStyleTag_PageStyleOverridesBuiltinOrientation(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml>
  <page style="tabloid" style.orientation="landscape"></page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}
	page := doc.Root().Page(0)
	if page == nil {
		t.Fatal("expected a page, got nil")
	}
	if page.Width() != inchesToPoints(17) {
		t.Errorf("expected landscape tabloid width %v, got %v", inchesToPoints(17), page.Width())
	}
	if page.Height() != inchesToPoints(11) {
		t.Errorf("expected landscape tabloid height %v, got %v", inchesToPoints(11), page.Height())
	}
}

func TestPageStyleFor_BuiltinsAreCaseInsensitive(t *testing.T) {
	ps := PageStyleFor("a4", &defaultScope)
	if ps == nil {
		t.Fatal("a4 page style not found")
	}
	wantWidth := mmToPoints(210)
	wantHeight := mmToPoints(297)
	if math.Abs(ps.Width()-wantWidth) > 0.001 {
		t.Fatalf("a4 width = %v, want %v", ps.Width(), wantWidth)
	}
	if math.Abs(ps.Height()-wantHeight) > 0.001 {
		t.Fatalf("a4 height = %v, want %v", ps.Height(), wantHeight)
	}
}

func TestPageStyleFor_BuiltinsCoverRepresentativeFamilies(t *testing.T) {
	tests := []struct {
		id         string
		wantWidth  float64
		wantHeight float64
	}{
		{id: "A0", wantWidth: mmToPoints(841), wantHeight: mmToPoints(1189)},
		{id: "governmentlegal", wantWidth: inchesToPoints(8.5), wantHeight: inchesToPoints(13)},
		{id: "ansic", wantWidth: inchesToPoints(17), wantHeight: inchesToPoints(22)},
		{id: "archd", wantWidth: inchesToPoints(24), wantHeight: inchesToPoints(36)},
	}
	for _, tt := range tests {
		ps := PageStyleFor(tt.id, &defaultScope)
		if ps == nil {
			t.Fatalf("%s page style not found", tt.id)
		}
		if math.Abs(ps.Width()-tt.wantWidth) > 0.001 {
			t.Fatalf("%s width = %v, want %v", tt.id, ps.Width(), tt.wantWidth)
		}
		if math.Abs(ps.Height()-tt.wantHeight) > 0.001 {
			t.Fatalf("%s height = %v, want %v", tt.id, ps.Height(), tt.wantHeight)
		}
	}
}

func TestPageStyle_SetAttrs_OrientationSwapsBuiltInDimensions(t *testing.T) {
	ps := &PageStyle{}
	ps.SetAttrs(map[string]string{"size": "a4", "orientation": "landscape"})
	if math.Abs(ps.Width()-mmToPoints(297)) > 0.001 {
		t.Fatalf("landscape a4 width = %v, want %v", ps.Width(), mmToPoints(297))
	}
	if math.Abs(ps.Height()-mmToPoints(210)) > 0.001 {
		t.Fatalf("landscape a4 height = %v, want %v", ps.Height(), mmToPoints(210))
	}
}
