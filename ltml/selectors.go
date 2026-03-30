package ltml

import (
	"fmt"
	"regexp"
	"strings"
)

// regexpForSelector compiles an LTML selector into a regexp that can be matched
// against an element path such as "div/p" or "div/p#hero.notice".
func regexpForSelector(selector string) *regexp.Regexp {
	return regexp.MustCompile(regexpStringForSelector(selector))
}

var (
	reExtraSpaces              = regexp.MustCompile(`\s{2,}`)
	reSpacesAroundAngleBracket = regexp.MustCompile(`\s*>\s*`)
)

const (
	resGT        = `/`
	resSpace     = resGT + `([^/]+/)*`
	resTag       = `\w+`
	resID        = `(#\w+)?`
	resMiscClass = `(\.\w+)*`
	resSpecClass = `(\.\w+)*\.%s(\.\w+)*`
)

// regexpStringForSelector normalizes a selector string and converts it into the
// regexp source used for matching LTML element paths.
func regexpStringForSelector(selector string) string {
	selector = strings.TrimSpace(selector)
	selector = reExtraSpaces.ReplaceAllLiteralString(selector, " ")
	selector = reSpacesAroundAngleBracket.ReplaceAllLiteralString(selector, ">")

	selectors := splitRuleSelectors(selector)
	var result []string
	for _, sel := range selectors {
		groups := strings.Split(sel, " ")
		var reGroups []string
		for _, group := range groups {
			reGroups = append(reGroups, regexpStringForSelectorGroup(group))
		}
		result = append(result, strings.Join(reGroups, resSpace)+"$")
	}
	if len(result) > 1 {
		return "(" + strings.Join(result, "|") + ")"
	}
	if len(result) == 1 {
		return result[0]
	}
	return ""
}

// splitRuleSelectors splits a comma-separated selector list into individual,
// trimmed selectors and discards empty entries.
func splitRuleSelectors(selector string) []string {
	rawSelectors := strings.Split(selector, ",")
	selectors := make([]string, 0, len(rawSelectors))
	for _, s := range rawSelectors {
		s = strings.TrimSpace(s)
		if s != "" {
			selectors = append(selectors, s)
		}
	}
	return selectors
}

// regexpStringForSelectorGroup converts one direct-child chain, such as
// "div>p.notice", into the corresponding regexp fragment.
func regexpStringForSelectorGroup(group string) string {
	items := strings.Split(group, ">")
	var reItems []string
	for _, item := range items {
		reItems = append(reItems, regexpStringForSelectorItem(item))
	}
	return strings.Join(reItems, resGT)
}

// regexpStringForSelectorItem converts one selector item, such as "p.notice"
// or "#hero", into the regexp fragment that matches a single path segment.
func regexpStringForSelectorItem(item string) string {
	t, k := split2(item, ".")
	if t == "" {
		t = resTag + resID
	} else if t[0] == '#' {
		t = resTag + t
	} else if !strings.Contains(t, "#") {
		t += resID
	}
	if k == "" {
		k = resMiscClass
	} else {
		k = fmt.Sprintf(resSpecClass, k)
	}
	return t + k
}

// specificityForSelector returns the highest specificity among the selectors in
// a comma-separated selector list.
func specificityForSelector(selector string) Specificity {
	var spec Specificity
	for _, sel := range splitRuleSelectors(selector) {
		current := specificityForSingleSelector(sel)
		if current.Compare(spec) > 0 {
			spec = current
		}
	}
	return spec
}

// specificityForSingleSelector computes CSS-like specificity for one selector.
// IDs outrank classes, which outrank tag names; combinators contribute no
// weight.
func specificityForSingleSelector(selector string) Specificity {
	selector = strings.TrimSpace(selector)
	selector = reExtraSpaces.ReplaceAllLiteralString(selector, " ")
	selector = reSpacesAroundAngleBracket.ReplaceAllLiteralString(selector, ">")

	var spec Specificity
	for _, group := range strings.Split(selector, " ") {
		for _, item := range strings.Split(group, ">") {
			item = strings.TrimSpace(item)
			if item == "" {
				continue
			}
			spec = addSpecificity(spec, specificityForSelectorItem(item))
		}
	}
	return spec
}

// specificityForSelectorItem computes specificity for a single selector item
// such as "p.notice", ".notice", or "p#hero".
func specificityForSelectorItem(item string) Specificity {
	var spec Specificity
	parts := strings.Split(item, ".")
	head := parts[0]
	if len(parts) > 1 {
		spec.Classes += len(parts) - 1
	}
	if head == "" {
		return spec
	}
	if strings.HasPrefix(head, "#") {
		spec.IDs++
		return spec
	}
	if strings.Contains(head, "#") {
		spec.Tags++
		spec.IDs++
		return spec
	}
	spec.Tags++
	return spec
}

// addSpecificity adds two specificity tuples component-wise.
func addSpecificity(left, right Specificity) Specificity {
	return Specificity{
		IDs:     left.IDs + right.IDs,
		Classes: left.Classes + right.Classes,
		Tags:    left.Tags + right.Tags,
	}
}

// cmpInt compares two integers and returns -1, 0, or 1.
func cmpInt(left, right int) int {
	switch {
	case left < right:
		return -1
	case left > right:
		return 1
	default:
		return 0
	}
}
