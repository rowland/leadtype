package ltml

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	reExtraSpaces              = regexp.MustCompile(`\s{2,}`)
	reSpacesAroundAngleBracket = regexp.MustCompile(`\s*>\s*`)
)

type selectorCombinator uint8

const (
	selectorCombinatorNone selectorCombinator = iota
	selectorCombinatorDescendant
	selectorCombinatorChild
)

type compiledSelectorList struct {
	selectors []*compiledSelector
}

type compiledSelector struct {
	parts       []selectorPart
	specificity Specificity
	hasPseudo   bool
}

type selectorPart struct {
	item       selectorItem
	combinator selectorCombinator
}

type selectorItem struct {
	tag     string
	id      string
	classes []string
	pseudos []selectorPseudo
}

type selectorPseudo struct {
	name  string
	value int
}

type selectorMatchNode interface {
	ParentSelectorNode() selectorMatchNode
	SelectorIdentity() selectorIdentity
	SelectorPseudoContext() (selectorPseudoContext, bool)
}

type selectorIdentity struct {
	Tag     string
	ID      string
	Classes []string
}

type selectorPseudoContext struct {
	widget   Widget
	resolver *selectorStructureResolver
}

type selectorStructureResolver struct {
	parentByWidget   map[Widget]Container
	childrenByParent map[Container][]Widget
	tableInfoByBox   map[Container]*tableStructureInfo
}

type tableStructureInfo struct {
	anchorByWidget map[Widget]tableAnchor
	rows           int
	cols           int
}

type tableAnchor struct {
	row int
	col int
}

type widgetSelectorNode struct {
	widget   Widget
	resolver *selectorStructureResolver
}

type pathSelectorNode struct {
	identity selectorIdentity
	parent   *pathSelectorNode
}

type pseudoMatcher func(ctx selectorPseudoContext, pseudo selectorPseudo) bool

var pseudoMatchers = map[string]pseudoMatcher{
	"dir(ltr)": func(ctx selectorPseudoContext, pseudo selectorPseudo) bool {
		return selectorDirection(ctx) == DirLTR
	},
	"dir(rtl)": func(ctx selectorPseudoContext, pseudo selectorPseudo) bool {
		return selectorDirection(ctx) == DirRTL
	},
	"first-child": func(ctx selectorPseudoContext, pseudo selectorPseudo) bool {
		index, count, ok := ctx.resolver.childPosition(ctx.widget)
		return ok && index == 0 && count > 0
	},
	"last-child": func(ctx selectorPseudoContext, pseudo selectorPseudo) bool {
		index, count, ok := ctx.resolver.childPosition(ctx.widget)
		return ok && count > 0 && index == count-1
	},
	"first-row": func(ctx selectorPseudoContext, pseudo selectorPseudo) bool {
		row, _, _, _, ok := ctx.resolver.tablePosition(ctx.widget)
		return ok && row == 0
	},
	"last-row": func(ctx selectorPseudoContext, pseudo selectorPseudo) bool {
		row, _, rows, _, ok := ctx.resolver.tablePosition(ctx.widget)
		return ok && rows > 0 && row == rows-1
	},
	"first-col": func(ctx selectorPseudoContext, pseudo selectorPseudo) bool {
		_, col, _, _, ok := ctx.resolver.tablePosition(ctx.widget)
		return ok && col == 0
	},
	"last-col": func(ctx selectorPseudoContext, pseudo selectorPseudo) bool {
		_, col, _, cols, ok := ctx.resolver.tablePosition(ctx.widget)
		return ok && cols > 0 && col == cols-1
	},
	"row-even": func(ctx selectorPseudoContext, pseudo selectorPseudo) bool {
		row, _, _, _, ok := ctx.resolver.tablePosition(ctx.widget)
		return ok && row%2 == 0
	},
	"row-odd": func(ctx selectorPseudoContext, pseudo selectorPseudo) bool {
		row, _, _, _, ok := ctx.resolver.tablePosition(ctx.widget)
		return ok && row%2 == 1
	},
	"col-even": func(ctx selectorPseudoContext, pseudo selectorPseudo) bool {
		_, col, _, _, ok := ctx.resolver.tablePosition(ctx.widget)
		return ok && col%2 == 0
	},
	"col-odd": func(ctx selectorPseudoContext, pseudo selectorPseudo) bool {
		_, col, _, _, ok := ctx.resolver.tablePosition(ctx.widget)
		return ok && col%2 == 1
	},
	"row-n": func(ctx selectorPseudoContext, pseudo selectorPseudo) bool {
		row, _, _, _, ok := ctx.resolver.tablePosition(ctx.widget)
		return ok && row == pseudo.value
	},
	"col-n": func(ctx selectorPseudoContext, pseudo selectorPseudo) bool {
		_, col, _, _, ok := ctx.resolver.tablePosition(ctx.widget)
		return ok && col == pseudo.value
	},
}

// selectorDirection returns the effective direction of the matched element.
// Containers may override inherited direction themselves; leaf widgets inherit
// it from their structural parent. An unattached widget follows LTML's LTR
// default.
func selectorDirection(ctx selectorPseudoContext) Dir {
	if container, ok := ctx.widget.(Container); ok {
		return container.Dir()
	}
	if ctx.resolver != nil {
		if parent := ctx.resolver.structuralParent(ctx.widget); parent != nil {
			return parent.Dir()
		}
	}
	return DirLTR
}

// compileSelectorList parses a possibly grouped selector string into compiled
// selectors that can be matched repeatedly during rule evaluation.
func compileSelectorList(selector string) (*compiledSelectorList, error) {
	rawSelectors := splitRuleSelectors(selector)
	list := &compiledSelectorList{selectors: make([]*compiledSelector, 0, len(rawSelectors))}
	for _, raw := range rawSelectors {
		compiled, err := compileSingleSelector(raw)
		if err != nil {
			return nil, err
		}
		list.selectors = append(list.selectors, compiled)
	}
	return list, nil
}

// splitRuleSelectors breaks a grouped selector list ("p, .note") into
// individual selectors and removes empty items.
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

// compileSingleSelector parses one selector into ordered parts and
// combinators, and computes its specificity metadata.
func compileSingleSelector(selector string) (*compiledSelector, error) {
	normalized := normalizeSelector(selector)
	if normalized == "" {
		return &compiledSelector{}, nil
	}
	tokens := tokenizeSelector(normalized)
	if len(tokens) == 0 {
		return &compiledSelector{}, nil
	}
	parts := make([]selectorPart, 0, len(tokens))
	nextCombinator := selectorCombinatorNone
	for _, token := range tokens {
		switch token {
		case ">":
			nextCombinator = selectorCombinatorChild
		case " ":
			if nextCombinator == selectorCombinatorNone {
				nextCombinator = selectorCombinatorDescendant
			}
		default:
			item, err := parseSelectorItem(token)
			if err != nil {
				return nil, err
			}
			parts = append(parts, selectorPart{
				item:       item,
				combinator: nextCombinator,
			})
			nextCombinator = selectorCombinatorNone
		}
	}
	var spec Specificity
	hasPseudo := false
	for _, part := range parts {
		spec = addSpecificity(spec, specificityForSelectorItem(part.item))
		if len(part.item.pseudos) > 0 {
			hasPseudo = true
		}
	}
	return &compiledSelector{parts: parts, specificity: spec, hasPseudo: hasPseudo}, nil
}

// normalizeSelector trims and canonicalizes spacing so tokenization is stable.
func normalizeSelector(selector string) string {
	selector = strings.TrimSpace(selector)
	selector = reExtraSpaces.ReplaceAllLiteralString(selector, " ")
	selector = reSpacesAroundAngleBracket.ReplaceAllLiteralString(selector, ">")
	return selector
}

// tokenizeSelector converts a normalized selector into item and combinator
// tokens while preserving descendant-space combinators.
func tokenizeSelector(selector string) []string {
	var tokens []string
	var current strings.Builder
	flush := func() {
		if current.Len() > 0 {
			tokens = append(tokens, current.String())
			current.Reset()
		}
	}
	for _, r := range selector {
		switch r {
		case '>':
			flush()
			tokens = append(tokens, ">")
		case ' ':
			flush()
			if len(tokens) == 0 || tokens[len(tokens)-1] == " " || tokens[len(tokens)-1] == ">" {
				continue
			}
			tokens = append(tokens, " ")
		default:
			current.WriteRune(r)
		}
	}
	flush()
	if len(tokens) > 0 && tokens[len(tokens)-1] == " " {
		tokens = tokens[:len(tokens)-1]
	}
	return tokens
}

// parseSelectorItem parses a single selector token such as "table.row:first-row"
// into tag, id, classes, and pseudo-class components.
func parseSelectorItem(token string) (selectorItem, error) {
	var item selectorItem
	base, pseudoText, _ := strings.Cut(token, ":")
	if err := parseSelectorBase(base, &item); err != nil {
		return selectorItem{}, err
	}
	if pseudoText == "" && strings.Contains(token, ":") {
		return selectorItem{}, fmt.Errorf("invalid selector %q", token)
	}
	if pseudoText != "" {
		parts := strings.Split(pseudoText, ":")
		for _, part := range parts {
			pseudo, err := parseSelectorPseudo(part)
			if err != nil {
				return selectorItem{}, err
			}
			item.pseudos = append(item.pseudos, pseudo)
		}
	}
	return item, nil
}

// parseSelectorBase parses the non-pseudo selector component and assigns its
// tag, id, and class values to item.
func parseSelectorBase(base string, item *selectorItem) error {
	if base == "" {
		return nil
	}
	i := 0
	if isWordByte(base[0]) {
		start := i
		for i < len(base) && isWordByte(base[i]) {
			i++
		}
		item.tag = base[start:i]
	}
	for i < len(base) {
		switch base[i] {
		case '#':
			i++
			start := i
			for i < len(base) && isWordByte(base[i]) {
				i++
			}
			if start == i {
				return fmt.Errorf("invalid selector %q", base)
			}
			item.id = base[start:i]
		case '.':
			i++
			start := i
			for i < len(base) && isWordByte(base[i]) {
				i++
			}
			if start == i {
				return fmt.Errorf("invalid selector %q", base)
			}
			item.classes = append(item.classes, base[start:i])
		default:
			return fmt.Errorf("invalid selector %q", base)
		}
	}
	return nil
}

// parseSelectorPseudo validates and parses a supported pseudo-class expression.
func parseSelectorPseudo(part string) (selectorPseudo, error) {
	if part == "" {
		return selectorPseudo{}, fmt.Errorf("invalid empty pseudo-class")
	}
	if strings.HasPrefix(part, "row-") {
		n, err := strconv.Atoi(part[len("row-"):])
		if err == nil {
			return selectorPseudo{name: "row-n", value: n}, nil
		}
	}
	if strings.HasPrefix(part, "col-") {
		n, err := strconv.Atoi(part[len("col-"):])
		if err == nil {
			return selectorPseudo{name: "col-n", value: n}, nil
		}
	}
	if _, ok := pseudoMatchers[part]; ok {
		return selectorPseudo{name: part}, nil
	}
	return selectorPseudo{}, fmt.Errorf("unknown pseudo-class %q", part)
}

func isWordByte(b byte) bool {
	return b == '_' || b == '-' || b >= '0' && b <= '9' || b >= 'A' && b <= 'Z' || b >= 'a' && b <= 'z'
}

func (list *compiledSelectorList) MatchesPath(path string) bool {
	node := selectorNodeForPath(path)
	if node == nil {
		return false
	}
	for _, selector := range list.selectors {
		if selector.hasPseudo {
			continue
		}
		if selector.matches(node) {
			return true
		}
	}
	return false
}

func (list *compiledSelectorList) MatchesWidget(widget Widget, resolver *selectorStructureResolver) bool {
	node := &widgetSelectorNode{widget: widget, resolver: resolver}
	for _, selector := range list.selectors {
		if selector.matches(node) {
			return true
		}
	}
	return false
}

func (selector *compiledSelector) matches(node selectorMatchNode) bool {
	if node == nil || len(selector.parts) == 0 {
		return false
	}
	return selector.matchPart(len(selector.parts)-1, node)
}

func (selector *compiledSelector) matchPart(index int, node selectorMatchNode) bool {
	if selectorNodeIsNil(node) {
		return false
	}
	if !selector.parts[index].item.matches(node) {
		return false
	}
	if index == 0 {
		return true
	}
	switch selector.parts[index].combinator {
	case selectorCombinatorChild:
		return selector.matchPart(index-1, node.ParentSelectorNode())
	case selectorCombinatorDescendant:
		for parent := node.ParentSelectorNode(); parent != nil; parent = parent.ParentSelectorNode() {
			if selector.matchPart(index-1, parent) {
				return true
			}
		}
		return false
	default:
		return selector.matchPart(index-1, node.ParentSelectorNode())
	}
}

func selectorNodeIsNil(node selectorMatchNode) bool {
	switch n := node.(type) {
	case nil:
		return true
	case *pathSelectorNode:
		return n == nil
	case *widgetSelectorNode:
		return n == nil
	default:
		return false
	}
}

func (item selectorItem) matches(node selectorMatchNode) bool {
	identity := node.SelectorIdentity()
	if item.tag != "" && identity.Tag != item.tag {
		return false
	}
	if item.id != "" && identity.ID != item.id {
		return false
	}
	for _, className := range item.classes {
		if !containsString(identity.Classes, className) {
			return false
		}
	}
	if len(item.pseudos) == 0 {
		return true
	}
	ctx, ok := node.SelectorPseudoContext()
	if !ok {
		return false
	}
	for _, pseudo := range item.pseudos {
		matcher, ok := pseudoMatchers[pseudo.name]
		if !ok || !matcher(ctx, pseudo) {
			return false
		}
	}
	return true
}

func (node *widgetSelectorNode) ParentSelectorNode() selectorMatchNode {
	if node == nil {
		return nil
	}
	parent := node.resolver.structuralParent(node.widget)
	if parent == nil {
		return nil
	}
	return &widgetSelectorNode{widget: parent, resolver: node.resolver}
}

func (node *widgetSelectorNode) SelectorIdentity() selectorIdentity {
	if node == nil {
		return selectorIdentity{}
	}
	return selectorIdentityForPath(node.widget.Path())
}

func (node *widgetSelectorNode) SelectorPseudoContext() (selectorPseudoContext, bool) {
	if node == nil {
		return selectorPseudoContext{}, false
	}
	return selectorPseudoContext{widget: node.widget, resolver: node.resolver}, true
}

func (node *pathSelectorNode) ParentSelectorNode() selectorMatchNode {
	if node == nil {
		return nil
	}
	return node.parent
}

func (node *pathSelectorNode) SelectorIdentity() selectorIdentity {
	if node == nil {
		return selectorIdentity{}
	}
	return node.identity
}

func (node *pathSelectorNode) SelectorPseudoContext() (selectorPseudoContext, bool) {
	if node == nil {
		return selectorPseudoContext{}, false
	}
	return selectorPseudoContext{}, false
}

func selectorNodeForPath(path string) selectorMatchNode {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil
	}
	var root *pathSelectorNode
	for _, segment := range strings.Split(path, "/") {
		if segment == "" {
			continue
		}
		root = &pathSelectorNode{
			identity: selectorIdentityForSegment(segment),
			parent:   root,
		}
	}
	return root
}

func selectorIdentityForPath(path string) selectorIdentity {
	parts := strings.Split(path, "/")
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] != "" {
			return selectorIdentityForSegment(parts[i])
		}
	}
	return selectorIdentity{}
}

func selectorIdentityForSegment(segment string) selectorIdentity {
	var id selectorIdentity
	head, classesText, _ := strings.Cut(segment, ".")
	if classesText != "" {
		id.Classes = strings.Split(classesText, ".")
	}
	if head == "" {
		return id
	}
	tag, ident, hasID := strings.Cut(head, "#")
	if hasID {
		id.Tag = tag
		id.ID = ident
		return id
	}
	if strings.HasPrefix(head, "#") {
		id.ID = strings.TrimPrefix(head, "#")
		return id
	}
	id.Tag = head
	return id
}

func specificityForSelector(selector string) Specificity {
	list, err := compileSelectorList(selector)
	if err != nil {
		return Specificity{}
	}
	var spec Specificity
	for _, current := range list.selectors {
		if current.specificity.Compare(spec) > 0 {
			spec = current.specificity
		}
	}
	return spec
}

func specificityForSelectorItem(item selectorItem) Specificity {
	spec := Specificity{
		Classes: len(item.classes) + len(item.pseudos),
	}
	if item.id != "" {
		spec.IDs++
	}
	if item.tag != "" {
		spec.Tags++
	}
	return spec
}

func addSpecificity(left, right Specificity) Specificity {
	return Specificity{
		IDs:     left.IDs + right.IDs,
		Classes: left.Classes + right.Classes,
		Tags:    left.Tags + right.Tags,
	}
}

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

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func newSelectorStructureResolver() *selectorStructureResolver {
	return &selectorStructureResolver{
		parentByWidget:   make(map[Widget]Container),
		childrenByParent: make(map[Container][]Widget),
		tableInfoByBox:   make(map[Container]*tableStructureInfo),
	}
}

func (r *selectorStructureResolver) structuralParent(widget Widget) Container {
	if widget == nil {
		return nil
	}
	if parent, ok := r.parentByWidget[widget]; ok {
		return parent
	}
	containerWidget, ok := widget.(interface{ Container() Container })
	if !ok {
		return nil
	}
	parent := containerWidget.Container()
	if sector, ok := parent.(*StdSector); ok && sector.Tag == "" {
		parent = sector.Container()
	}
	r.parentByWidget[widget] = parent
	return parent
}

func (r *selectorStructureResolver) structuralChildren(container Container) []Widget {
	if container == nil {
		return nil
	}
	if children, ok := r.childrenByParent[container]; ok {
		return children
	}
	raw := container.Widgets()
	children := make([]Widget, 0, len(raw))
	for _, child := range raw {
		if sector, ok := child.(*StdSector); ok && sector.Tag == "" {
			children = append(children, sector.Widgets()...)
			continue
		}
		children = append(children, child)
	}
	r.childrenByParent[container] = children
	return children
}

func (r *selectorStructureResolver) childPosition(widget Widget) (index int, count int, ok bool) {
	parent := r.structuralParent(widget)
	if parent == nil {
		return 0, 0, false
	}
	children := r.structuralChildren(parent)
	for i, child := range children {
		if child == widget {
			return i, len(children), true
		}
	}
	return 0, len(children), false
}

func (r *selectorStructureResolver) tablePosition(widget Widget) (row int, col int, rows int, cols int, ok bool) {
	parent := r.structuralParent(widget)
	if parent == nil || parent.LayoutStyle() == nil || parent.LayoutStyle().manager != "table" {
		return 0, 0, 0, 0, false
	}
	info, ok := r.tableInfoByBox[parent]
	if !ok {
		info = r.buildTableInfo(parent)
		r.tableInfoByBox[parent] = info
	}
	anchor, ok := info.anchorByWidget[widget]
	if !ok {
		return 0, 0, info.rows, info.cols, false
	}
	return anchor.row, anchor.col, info.rows, info.cols, true
}

func (r *selectorStructureResolver) buildTableInfo(container Container) *tableStructureInfo {
	info := &tableStructureInfo{anchorByWidget: make(map[Widget]tableAnchor)}
	var (
		grid *WidgetGrid
		err  error
	)
	if container.Order() == TableOrderRows {
		grid, err = rowGrid(container)
	} else {
		grid, err = colGrid(container)
	}
	if err != nil || grid == nil {
		return info
	}
	info.rows = grid.Rows()
	info.cols = grid.Cols()
	for row := 0; row < grid.Rows(); row++ {
		for col := 0; col < grid.Cols(); col++ {
			widget := grid.Cell(col, row)
			if widget == nil {
				continue
			}
			if _, exists := info.anchorByWidget[widget]; !exists {
				info.anchorByWidget[widget] = tableAnchor{row: row, col: col}
			}
		}
	}
	return info
}
