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

func TestCompileSelectorList_rejects_unknown_pseudo_class(t *testing.T) {
	if _, err := compileSelectorList("p:wat"); err == nil {
		t.Fatal("expected compile error for unknown pseudo class")
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
	got := specificityForSelector("div.notice:first-child > p:row-0")
	want := Specificity{IDs: 0, Classes: 3, Tags: 2}
	if got != want {
		t.Fatalf("specificity = %+v, want %+v", got, want)
	}
}
