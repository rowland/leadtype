package ltml

import "testing"

func TestCompileSelectorList_matches_basic_tag_id_and_class_paths(t *testing.T) {
	compiled, err := compileSelectorList("foo, .bar, #hero")
	if err != nil {
		t.Fatalf("compileSelectorList: %v", err)
	}
	if !compiled.MatchesPath("foo") {
		t.Fatal("expected tag selector to match")
	}
	if !compiled.MatchesPath("foo.bar") {
		t.Fatal("expected class selector to match")
	}
	if !compiled.MatchesPath("foo#hero") {
		t.Fatal("expected id selector to match")
	}
	if compiled.MatchesPath("baz") {
		t.Fatal("unexpected match for unrelated tag")
	}
}

func TestCompileSelectorList_matches_child_and_descendant_paths(t *testing.T) {
	direct, err := compileSelectorList("foo#bar>foo.baz")
	if err != nil {
		t.Fatalf("compileSelectorList direct: %v", err)
	}
	if !direct.MatchesPath("foo#bar/foo.baz") {
		t.Fatal("expected direct child match")
	}
	if direct.MatchesPath("foo#bar/a/foo.baz") {
		t.Fatal("direct child selector should not match descendant")
	}

	desc, err := compileSelectorList("foo#bar foo.baz")
	if err != nil {
		t.Fatalf("compileSelectorList descendant: %v", err)
	}
	if !desc.MatchesPath("foo#bar/a/foo.baz") {
		t.Fatal("expected descendant match")
	}
}

func TestCompileSelectorList_matches_universal_selectors(t *testing.T) {
	for _, tt := range []struct {
		selector string
		path     string
		want     bool
	}{
		{selector: "*", path: "p", want: true},
		{selector: "*", path: "div/label", want: true},
		{selector: "div>*", path: "div/label", want: true},
		{selector: "div>*", path: "div/vbox/label", want: false},
		{selector: "div *", path: "div/vbox/label", want: true},
		{selector: "*.notice", path: "p.notice", want: true},
		{selector: "*.notice", path: "p", want: false},
		{selector: "*#hero", path: "label#hero", want: true},
		{selector: "p, *#hero", path: "label#hero", want: true},
	} {
		compiled, err := compileSelectorList(tt.selector)
		if err != nil {
			t.Fatalf("compileSelectorList(%q): %v", tt.selector, err)
		}
		if got := compiled.MatchesPath(tt.path); got != tt.want {
			t.Errorf("%q match for %q = %v, want %v", tt.selector, tt.path, got, tt.want)
		}
	}
}

func TestCompileSelectorList_rejects_malformed_universal_selectors(t *testing.T) {
	for _, selector := range []string{"**", "*tag", "tag*"} {
		if _, err := compileSelectorList(selector); err == nil {
			t.Errorf("compileSelectorList(%q) succeeded, want an error", selector)
		}
	}
}

func TestCompileSelectorList_matches_universal_selector_with_pseudo_class(t *testing.T) {
	parent := &StdContainer{}
	first := &StdLabel{}
	second := &StdParagraph{}
	if err := first.SetContainer(parent); err != nil {
		t.Fatal(err)
	}
	if err := second.SetContainer(parent); err != nil {
		t.Fatal(err)
	}
	parent.AddChild(first)
	parent.AddChild(second)

	compiled, err := compileSelectorList("*:first-child")
	if err != nil {
		t.Fatalf("compileSelectorList: %v", err)
	}
	resolver := newSelectorStructureResolver()
	if !compiled.MatchesWidget(first, resolver) {
		t.Fatal("universal first-child selector did not match the first child")
	}
	if compiled.MatchesWidget(second, resolver) {
		t.Fatal("universal first-child selector matched the second child")
	}
}

func TestCompileSelectorList_rejects_unknown_pseudo_class(t *testing.T) {
	if _, err := compileSelectorList("p:wat"); err == nil {
		t.Fatal("expected compile error for unknown pseudo class")
	}
}

func TestCompileSelectorList_accepts_opaque_direction_pseudos(t *testing.T) {
	resolver := newSelectorStructureResolver()
	for _, tt := range []struct {
		selector string
		widget   Widget
		want     bool
	}{
		{selector: ":dir(ltr)", widget: &StdLabel{}, want: true},
		{selector: ":dir(rtl)", widget: &StdLabel{}, want: false},
		{selector: ":dir(rtl)", widget: &StdContainer{dir: DirRTL, dirExplicit: true}, want: true},
	} {
		compiled, err := compileSelectorList(tt.selector)
		if err != nil {
			t.Fatalf("compileSelectorList(%q): %v", tt.selector, err)
		}
		if got := compiled.MatchesWidget(tt.widget, resolver); got != tt.want {
			t.Errorf("%s match = %v, want %v", tt.selector, got, tt.want)
		}
	}
}

func TestCompileSelectorList_rejects_noncanonical_direction_pseudos(t *testing.T) {
	for _, selector := range []string{"p:dir", "p:dir()", "p:dir(auto)", "p:dir( rtl )"} {
		if _, err := compileSelectorList(selector); err == nil {
			t.Errorf("compileSelectorList(%q) succeeded, want an error", selector)
		}
	}
}

func TestCompileSelectorList_accepts_hyphenated_class_and_id_names(t *testing.T) {
	compiled, err := compileSelectorList("div#hero-panel.demo-card")
	if err != nil {
		t.Fatalf("compileSelectorList: %v", err)
	}
	if !compiled.MatchesPath("div#hero-panel.demo-card") {
		t.Fatal("expected hyphenated selector to match path")
	}
}

func TestCompileSelectorList_rejects_malformed_row_and_col_pseudos(t *testing.T) {
	for _, selector := range []string{"p:row-x", "p:col-z"} {
		if _, err := compileSelectorList(selector); err == nil {
			t.Fatalf("expected compile error for %q", selector)
		}
	}
}

func TestSpecificityForSelector_counts_pseudo_classes_as_classes(t *testing.T) {
	got := specificityForSelector("div.notice:first-child > p:row-0:dir(rtl)")
	want := Specificity{IDs: 0, Classes: 4, Tags: 2}
	if got != want {
		t.Fatalf("specificity = %+v, want %+v", got, want)
	}
}

func TestSpecificityForSelector_universal_selector_adds_no_specificity(t *testing.T) {
	for _, tt := range []struct {
		selector string
		want     Specificity
	}{
		{selector: "*", want: Specificity{}},
		{selector: "div > *", want: Specificity{Tags: 1}},
		{selector: "*#hero.notice:first-child", want: Specificity{IDs: 1, Classes: 2}},
	} {
		if got := specificityForSelector(tt.selector); got != tt.want {
			t.Errorf("specificityForSelector(%q) = %+v, want %+v", tt.selector, got, tt.want)
		}
	}
}
