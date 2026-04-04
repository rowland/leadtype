package ltml

import (
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
	ltml := doc.ltmls[0]
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
	page := doc.ltmls[0].Page(0)
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
	page := doc.ltmls[0].Page(0)
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
