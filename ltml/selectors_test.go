package ltml

import "testing"

func TestRegexpForSelector_match_tag(t *testing.T) {
	if regexpStringForSelector("foo") != `foo(#\w+)?(\.\w+)*$` {
		t.Error("should generate an expression matching a tag")
	}
}

func TestRegexpForSelector_match_tag_with_id(t *testing.T) {
	if regexpStringForSelector("foo#bar") != `foo#bar(\.\w+)*$` {
		t.Error("should generate an expression matching a tag with an id")
	}
}

func TestRegexpForSelector_match_tag_with_class(t *testing.T) {
	if regexpStringForSelector("foo.bar") != `foo(#\w+)?(\.\w+)*\.bar(\.\w+)*$` {
		t.Error("should generate an expression matching a tag with a class")
	}
}

func TestRegexpForSelector_match_id(t *testing.T) {
	if regexpStringForSelector("#bar") != `\w+#bar(\.\w+)*$` {
		t.Error("should generate an expression matching an id")
	}
}

func TestRegexpForSelector_match_class(t *testing.T) {
	if re := regexpStringForSelector(".bar"); re != `\w+(#\w+)?(\.\w+)*\.bar(\.\w+)*$` {
		t.Errorf("should generate an expression matching a class, got %s", re)
	}
}

func TestRegexpForSelector_match_tag_with_class_child_of_tag_with_id(t *testing.T) {
	if re := regexpStringForSelector("foo#bar>foo.bar"); re != `foo#bar(\.\w+)*/foo(#\w+)?(\.\w+)*\.bar(\.\w+)*$` {
		t.Errorf("should generate an expression matching a tag with a class that is the direct child of a tag with an id, got %s", re)
	}
}

func TestRegexpForSelector_match_tag_with_class_descendent_of_tag_with_id(t *testing.T) {
	if re := regexpStringForSelector("foo#bar foo.bar"); re != `foo#bar(\.\w+)*/([^/]+/)*foo(#\w+)?(\.\w+)*\.bar(\.\w+)*$` {
		t.Errorf("should generate an expression matching a tag with a class that is a descendent of a tag with an id, got %s", re)
	}
}

func TestRegexpForSelector_match_list_of_tags(t *testing.T) {
	if re := regexpStringForSelector("foo, bar"); re != `(foo(#\w+)?(\.\w+)*$|bar(#\w+)?(\.\w+)*$)` {
		t.Errorf("should generate an expression matching any of a comma-delimited list of tags, got %s", re)
	}
}

func TestRegexpForSelector_tag_matches(t *testing.T) {
	re := regexpForSelector("foo")

	if !re.MatchString("foo") {
		t.Error("should match a tag")
	}
	if !re.MatchString("foo#bar") {
		t.Error("should match a tag with an id")
	}
	if !re.MatchString("foo.bar") {
		t.Error("should match a tag with class")
	}
	if !re.MatchString("foo#bar.baz") {
		t.Error("should match a tag with an id and a class")
	}
}

func TestRegexpForSelector_comma_list_matches(t *testing.T) {
	re := regexpForSelector("foo, bar")

	if !re.MatchString("foo") {
		t.Error("should match the first tag")
	}
	if !re.MatchString("foo#bar") {
		t.Error("should match the first tag with an id")
	}
	if !re.MatchString("bar") {
		t.Error("should match the second tag")
	}
	if !re.MatchString("bar.baz") {
		t.Error("should match the second tag with a class")
	}
	if re.MatchString("baz") {
		t.Error("should not match an unspecified tag")
	}
}

func TestRegexpForSelector_id_matches(t *testing.T) {
	re := regexpForSelector("#bar")

	if re.MatchString("foo") {
		t.Error("should not match just a simple tag")
	}
	if !re.MatchString("foo#bar") {
		t.Error("should match a tag with the specified id")
	}
	if re.MatchString("foo.bar") {
		t.Error("should not match a tag with a class but no id")
	}
	if !re.MatchString("foo#bar.baz") {
		t.Error("should match a tag with the specified id and a class")
	}
}

func TestRegexpForSelector_class_matches(t *testing.T) {
	re := regexpForSelector(".bar")

	if re.MatchString("foo") {
		t.Error("should not match just a simple tag")
	}
	if re.MatchString("foo#bar") {
		t.Error("should not match a tag with an id but no class")
	}
	if !re.MatchString("foo.bar") {
		t.Error("should match a tag with the specified class")
	}
	if re.MatchString("foo#bar.baz") {
		t.Error("should not match a tag with an id named the same as the specified class")
	}
}

func TestRegexpForSelector_another_class_matches(t *testing.T) {
	re := regexpForSelector(".baz")

	if re.MatchString("foo") {
		t.Error("should not match just a simple tag")
	}
	if re.MatchString("foo#bar") {
		t.Error("should not match a tag with an id but no class")
	}
	if re.MatchString("foo.bar") {
		t.Error("should match a tag with the wrong class")
	}
	if !re.MatchString("foo#bar.baz") {
		t.Error("should match a tag with an id and the specified class")
	}
}

func TestRegexpForSelector_matching_a_direct_child(t *testing.T) {
	re := regexpForSelector("foo#bar>foo.bar")

	if !re.MatchString("foo#bar/foo.bar") {
		t.Error("should match an element exactly as specified")
	}

	if re.MatchString("foo.bar/foo#bar") {
		t.Error("should not match an element where the id and the class have been switched")
	}

	if re.MatchString("foo#bar/foo#bar") {
		t.Error("should not match an element where an id has been traded for a class")
	}

	if re.MatchString("foo.bar/foo.bar") {
		t.Error("should not match an element where a class has been traded for an id")
	}

	if !re.MatchString("foo#bar.baz/foo.bar") {
		t.Error("should match an element as specified with a class added to the parent")
	}

	if !re.MatchString("foo#bar/foo.bar.baz") {
		t.Error("should match an element as specified with a class added to the child")
	}

	if !re.MatchString("foo#bar.baz/foo.bar.baz") {
		t.Error("should match an element as specified with classes added to both parent and child")
	}
}

func TestRegexpForSelector_matching_an_indirect_child(t *testing.T) {
	re := regexpForSelector("foo#bar foo.bar")

	if !re.MatchString("foo#bar/foo.bar") {
		t.Error("should also match a direct child")
	}

	if re.MatchString("foo.bar/foo#bar") {
		t.Error("should not match a direct child where the id and the class have been switched")
	}

	if re.MatchString("foo#bar/foo#bar") {
		t.Error("should not match a direct child where an id has been traded for a class")
	}

	if re.MatchString("foo.bar/foo.bar") {
		t.Error("should not match an element where a class has been traded for an id")
	}

	if !re.MatchString("foo#bar.baz/foo.bar") {
		t.Error("should match a direct child as specified with a class added to the parent")
	}

	if !re.MatchString("foo#bar/foo.bar.baz") {
		t.Error("should match a direct child as specified with a class added to the child")
	}

	if !re.MatchString("foo#bar.baz/foo.bar.baz") {
		t.Error("should match a direct child as specified with classes added to both parent and child")
	}

	if !re.MatchString("foo#bar.baz/a/foo.bar.baz") {
		t.Error("should match a descendant separated by a simple tag")
	}

	if !re.MatchString("foo#bar.baz/a#b/foo.bar.baz") {
		t.Error("should match a descendant separated by a tag with an id")
	}

	if !re.MatchString("foo#bar.baz/a.c/foo.bar.baz") {
		t.Error("should match a descendant separated by a tag with a class")
	}

	if !re.MatchString("foo#bar.baz/a#b.c/foo.bar.baz") {
		t.Error("should match a descendant separated by a tag with an id and class")
	}

	if !re.MatchString("foo#bar.baz/#b/foo.bar.baz") {
		t.Error("should match a descendant separated by an id")
	}

	if !re.MatchString("foo#bar.baz/.c/foo.bar.baz") {
		t.Error("should match a descendant separated by a class")
	}
}
