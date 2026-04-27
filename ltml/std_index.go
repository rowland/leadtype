package ltml

import "fmt"

type StdIndex struct {
	StdContainer
	expandedTargets []string
	explicitWidth   bool
	explicitHeight  bool
	expandErr       error
}

// targetLinkWriter lets StdIndex attach one row-sized internal link annotation
// without changing the LTML Writer interface for all callers.
type targetLinkWriter interface {
	AddTargetLink(x, y, width, height float64, target string) error
}

func (i *StdIndex) BeforePrint(w Writer) error {
	if err := i.ensureExpanded(); err != nil {
		return err
	}
	return i.StdContainer.BeforePrint(w)
}

func (i *StdIndex) LayoutWidget(w Writer) {
	if err := i.ensureExpanded(); err != nil {
		debugf("StdIndex.LayoutWidget: %v", err)
		return
	}
	i.StdContainer.LayoutWidget(w)
}

func (i *StdIndex) PreferredHeight(w Writer) float64 {
	if i.HeightIsSet() {
		return i.Height()
	}
	if err := i.ensureExpanded(); err != nil {
		debugf("StdIndex.PreferredHeight: %v", err)
		return NonContentHeight(i)
	}
	return i.StdContainer.PreferredHeight(w)
}

func (i *StdIndex) PreferredWidth(Writer) float64 {
	if i.WidthIsSet() {
		return i.Width()
	}
	return 0
}

func (i *StdIndex) DrawContent(w Writer) error {
	if err := i.ensureExpanded(); err != nil {
		return err
	}
	if err := i.StdContainer.DrawContent(w); err != nil {
		return err
	}
	adder, ok := w.(targetLinkWriter)
	if !ok {
		return nil
	}
	children := i.Widgets()
	for idx, child := range children {
		if idx >= len(i.expandedTargets) || !child.Visible() || child.Disabled() {
			continue
		}
		target := i.expandedTargets[idx]
		if target == "" || child.Width() <= 0 || child.Height() <= 0 {
			continue
		}
		if err := adder.AddTargetLink(child.Left(), child.Top(), child.Width(), child.Height(), target); err != nil {
			return err
		}
	}
	return nil
}

func (i *StdIndex) SplitForHeight(avail float64, w Writer) (*SplitResult, error) {
	if err := i.ensureExpanded(); err != nil {
		return nil, err
	}
	if len(i.Widgets()) < 2 {
		return nil, nil
	}
	result, err := i.StdContainer.SplitForHeight(avail, w)
	if err != nil || result == nil {
		return result, err
	}
	head := i.wrapSplitFragment(result.Head)
	tail := i.wrapSplitFragment(result.Tail)
	if head == nil {
		return nil, nil
	}
	return &SplitResult{Head: head, Tail: tail}, nil
}

func (i *StdIndex) SetAttrs(attrs map[string]string) {
	i.StdContainer.SetAttrs(attrs)
	i.explicitWidth = MapHasAnyKey(attrs, "width") || (i.sides[leftSide].IsSet && i.sides[rightSide].IsSet)
	i.explicitHeight = MapHasAnyKey(attrs, "height") || (i.sides[topSide].IsSet && i.sides[bottomSide].IsSet)
}

func (i *StdIndex) clearExpandedState() {
	i.activeChildren = nil
	i.expandedTargets = nil
	i.expandErr = nil
}

func (i *StdIndex) clearMeasuredGeometry() {
	if !i.explicitWidth {
		i.width = 0
		i.widthPct = 0
		i.widthRel = 0
		i.widthSet = false
	}
	if !i.explicitHeight {
		i.height = 0
		i.heightPct = 0
		i.heightRel = 0
		i.heightSet = false
	}
}

func (i *StdIndex) currentEntries() []resolvedIndexEntry {
	doc := documentForContainer(i.container)
	if doc == nil || documentVisualCaptureActive(doc) {
		return nil
	}
	return doc.activeIndexEntries(i.ID)
}

func (i *StdIndex) ensureExpanded() error {
	if i.expandErr != nil {
		return i.expandErr
	}
	if i.activeChildren != nil {
		return nil
	}
	template, err := i.templateWidget()
	if err != nil {
		i.expandErr = err
		return err
	}
	entries := i.currentEntries()
	i.activeChildren = make([]Widget, 0, len(entries))
	i.expandedTargets = make([]string, 0, len(entries))
	for _, entry := range entries {
		child, err := i.cloneTemplateWidget(template, entry)
		if err != nil {
			i.activeChildren = nil
			i.expandedTargets = nil
			i.expandErr = err
			return err
		}
		i.activeChildren = append(i.activeChildren, child)
		i.expandedTargets = append(i.expandedTargets, entry.Target)
	}
	return nil
}

func (i *StdIndex) templateWidget() (Widget, error) {
	children := i.children
	if len(children) == 0 {
		return i.defaultTemplateWidget(), nil
	}
	if len(children) != 1 {
		return nil, fmt.Errorf("<index> supports exactly one block template child")
	}
	switch children[0].(type) {
	case *StdParagraph, *StdLabel, *StdContainer:
		return children[0], nil
	default:
		return nil, fmt.Errorf("<index> template child must be a block widget, got %T", children[0])
	}
}

func (i *StdIndex) defaultTemplateWidget() Widget {
	row := &StdParagraph{}
	row.SetScope(i.scope)
	row.SetDoc(i.doc)
	baseFont := i.Font()
	row.textPieces = []textPiece{
		{content: &StdIndexTitle{}, font: baseFont},
		{content: &StdLeader{}, font: baseFont},
		{content: &StdIndexPage{}, font: baseFont},
	}
	row.SetWidthPct(100)
	return row
}

func (i *StdIndex) cloneTemplateWidget(widget Widget, entry resolvedIndexEntry) (Widget, error) {
	clone := cloneWidgetShallow(widget)
	clone.SetPrinted(false)
	clone.SetVisible(true)
	clone.SetDisabled(false)
	if pathSetter, ok := clone.(interface{ SetPath(string) }); ok {
		pathSetter.SetPath("")
	}

	switch value := clone.(type) {
	case *StdParagraph:
		value.children = nil
		value.activeChildren = nil
		value.textPieces = rewriteIndexTextPieces(value.textPieces, entry)
		value.richText = nil
		value.splitLines = nil
	case *StdLabel:
		value.children = nil
		value.activeChildren = nil
		value.textPieces = rewriteIndexTextPieces(value.textPieces, entry)
		value.richText = nil
	case *StdContainer:
		value.children = nil
		value.activeChildren = nil
		source := widget.(Container)
		for _, child := range source.Widgets() {
			next, err := i.cloneTemplateWidget(child, entry)
			if err != nil {
				return nil, err
			}
			if wc, ok := next.(WantsContainer); ok {
				if err := wc.SetContainer(value); err != nil {
					return nil, err
				}
			}
			value.activeChildren = append(value.activeChildren, next)
		}
	}
	if wc, ok := clone.(WantsContainer); ok {
		if err := wc.SetContainer(i); err != nil {
			return nil, err
		}
	}
	return clone, nil
}

func rewriteIndexTextPieces(pieces []textPiece, entry resolvedIndexEntry) []textPiece {
	if len(pieces) == 0 {
		return nil
	}
	result := make([]textPiece, 0, len(pieces))
	for _, piece := range pieces {
		switch piece.content.(type) {
		case *StdIndexTitle:
			result = append(result, newStaticTextPiece(entry.Label, indexPieceFont(piece)))
		case *StdIndexPage:
			result = append(result, newStaticTextPiece(formatPageNo(entry.PageNo), indexPieceFont(piece)))
		default:
			result = append(result, piece)
		}
	}
	return result
}

func indexPieceFont(piece textPiece) *FontStyle {
	if piece.font != nil {
		return piece.font
	}
	content, ok := piece.content.(inlineTextWithFont)
	if !ok {
		return nil
	}
	return content.Font()
}

func (i *StdIndex) wrapSplitFragment(fragment Widget) *StdIndex {
	if fragment == nil {
		return nil
	}
	fragmentContainer, ok := fragment.(*StdContainer)
	if !ok {
		return nil
	}
	targets := i.targetsForChildren(fragmentContainer.Widgets())
	clone := *i
	clone.activeChildren = append([]Widget(nil), fragmentContainer.Widgets()...)
	clone.expandedTargets = targets
	clone.expandErr = nil
	clone.clearMeasuredGeometry()
	clone.printed = false
	clone.invisible = false
	clone.disabled = false
	clone.path = ""
	for _, child := range clone.activeChildren {
		if wc, ok := child.(WantsContainer); ok {
			_ = wc.SetContainer(&clone)
		}
	}
	return &clone
}

func (i *StdIndex) targetsForChildren(children []Widget) []string {
	if len(children) == 0 {
		return nil
	}
	byLogicalID := make(map[string]string, len(i.activeChildren))
	for idx, child := range i.activeChildren {
		if idx >= len(i.expandedTargets) {
			break
		}
		logicalID := widgetLogicalID(child)
		if logicalID == "" {
			continue
		}
		byLogicalID[logicalID] = i.expandedTargets[idx]
	}
	targets := make([]string, 0, len(children))
	for _, child := range children {
		targets = append(targets, byLogicalID[widgetLogicalID(child)])
	}
	return targets
}

func widgetLogicalID(widget Widget) string {
	identified, ok := widget.(interface{ AccessibilityLogicalID() string })
	if !ok {
		return ""
	}
	return identified.AccessibilityLogicalID()
}

func init() {
	registerTag(DefaultSpace, "index", func() any { return &StdIndex{} })
}

var _ Container = (*StdIndex)(nil)
var _ HasAttrs = (*StdIndex)(nil)
var _ Identifier = (*StdIndex)(nil)
var _ Splittable = (*StdIndex)(nil)
var _ WantsContainer = (*StdIndex)(nil)
