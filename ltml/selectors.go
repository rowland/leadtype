package ltml

import (
	"fmt"
	"regexp"
	"strings"
)

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

func regexpStringForSelector(selector string) string {
	selector = strings.TrimSpace(selector)
	selector = reExtraSpaces.ReplaceAllLiteralString(selector, " ")
	selector = reSpacesAroundAngleBracket.ReplaceAllLiteralString(selector, ">")

	selectors := strings.Split(selector, ",")
	for i, s := range selectors {
		selectors[i] = strings.TrimSpace(s)
	}
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

func regexpStringForSelectorGroup(group string) string {
	items := strings.Split(group, ">")
	var reItems []string
	for _, item := range items {
		reItems = append(reItems, regexpStringForSelectorItem(item))
	}
	return strings.Join(reItems, resGT)
}

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
