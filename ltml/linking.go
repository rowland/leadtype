package ltml

import (
	"fmt"
	"sort"
	"strings"
)

const maxIndexLayoutPasses = 8

type destinationRegistrar interface {
	RegisterDestination(name string, x, y float64)
}

type documentLinkSummary struct {
	hasIndexes bool
	indexIDs   map[string]struct{}
	targetRefs map[string]struct{}
}

type documentDestination struct {
	Name           string
	PageNo         int
	PhysicalPageNo int
	X              float64
	Y              float64
}

type resolvedIndexEntry struct {
	Label  string
	Target string
	PageNo int
}

type collectedIndexEntry struct {
	IndexID string
	Target  string
	Label   string
}

type documentIndexSnapshot struct {
	destinations map[string]documentDestination
	indexes      map[string][]resolvedIndexEntry
}

// documentRenderContext collects the metadata we discover while a single render
// pass walks the document tree. Preflight passes populate it without producing
// durable PDF output; the final pass uses the stabilized snapshot for rendering.
type documentRenderContext struct {
	activeSnapshot   *documentIndexSnapshot
	activeIndexEntry *resolvedIndexEntry
	destinations     map[string]documentDestination
	indexEntries     []collectedIndexEntry
	seenIndexEntries map[Widget]bool
	preflight        bool
}

func newDocumentRenderContext(snapshot *documentIndexSnapshot, preflight bool) *documentRenderContext {
	return &documentRenderContext{
		activeSnapshot:   snapshot,
		destinations:     make(map[string]documentDestination),
		seenIndexEntries: make(map[Widget]bool),
		preflight:        preflight,
	}
}

func (ctx *documentRenderContext) registerDestination(name string, pageNo, physicalPageNo int, x, y float64) {
	if name == "" {
		return
	}
	if _, exists := ctx.destinations[name]; exists {
		return
	}
	ctx.destinations[name] = documentDestination{
		Name:           name,
		PageNo:         pageNo,
		PhysicalPageNo: physicalPageNo,
		X:              x,
		Y:              y,
	}
}

func (ctx *documentRenderContext) registerIndexEntry(widget Widget, indexID, target, label string) {
	if widget == nil || indexID == "" || target == "" {
		return
	}
	if ctx.seenIndexEntries[widget] {
		return
	}
	ctx.seenIndexEntries[widget] = true
	ctx.indexEntries = append(ctx.indexEntries, collectedIndexEntry{
		IndexID: indexID,
		Target:  target,
		Label:   label,
	})
}

func (ctx *documentRenderContext) snapshot(summary *documentLinkSummary) (*documentIndexSnapshot, error) {
	result := &documentIndexSnapshot{
		destinations: make(map[string]documentDestination, len(ctx.destinations)),
		indexes:      make(map[string][]resolvedIndexEntry, len(summary.indexIDs)),
	}
	for name, destination := range ctx.destinations {
		result.destinations[name] = destination
	}
	for id := range summary.indexIDs {
		result.indexes[id] = nil
	}
	var missingTargets []string
	missingSeen := make(map[string]struct{})
	for _, entry := range ctx.indexEntries {
		destination, ok := ctx.destinations[entry.Target]
		if !ok {
			if _, exists := missingSeen[entry.Target]; !exists {
				missingSeen[entry.Target] = struct{}{}
				missingTargets = append(missingTargets, entry.Target)
			}
			continue
		}
		result.indexes[entry.IndexID] = append(result.indexes[entry.IndexID], resolvedIndexEntry{
			Label:  entry.Label,
			Target: entry.Target,
			PageNo: destination.PageNo,
		})
	}
	for target := range summary.targetRefs {
		if _, ok := ctx.destinations[target]; ok {
			continue
		}
		if _, exists := missingSeen[target]; exists {
			continue
		}
		missingSeen[target] = struct{}{}
		missingTargets = append(missingTargets, target)
	}
	if len(missingTargets) > 0 {
		sort.Strings(missingTargets)
		return nil, fmt.Errorf("missing internal target(s): %s", strings.Join(missingTargets, ", "))
	}
	return result, nil
}

func (snapshot *documentIndexSnapshot) Equal(other *documentIndexSnapshot) bool {
	if snapshot == nil || other == nil {
		return snapshot == other
	}
	if len(snapshot.indexes) != len(other.indexes) {
		return false
	}
	for id, entries := range snapshot.indexes {
		otherEntries, ok := other.indexes[id]
		if !ok || len(entries) != len(otherEntries) {
			return false
		}
		for i := range entries {
			if entries[i] != otherEntries[i] {
				return false
			}
		}
	}
	return true
}

func (d *StdDocument) activeIndexEntries(id string) []resolvedIndexEntry {
	if d.renderContext == nil || d.renderContext.activeSnapshot == nil {
		return nil
	}
	return d.renderContext.activeSnapshot.indexes[id]
}

func (d *StdDocument) describeLinking() (*documentLinkSummary, error) {
	summary := &documentLinkSummary{
		indexIDs:   make(map[string]struct{}),
		targetRefs: make(map[string]struct{}),
	}
	indexRefs := make(map[string]struct{})
	var err error
	walkWidgets(d, func(widget Widget) bool {
		switch value := widget.(type) {
		case *StdA:
			switch {
			case value.uri == "" && value.target == "":
				err = fmt.Errorf("<a> requires exactly one of uri or target")
				return false
			case value.uri != "" && value.target != "":
				err = fmt.Errorf("<a> cannot specify both uri and target")
				return false
			}
			if value.target != "" {
				summary.targetRefs[value.target] = struct{}{}
			}
		case *StdTarget:
			if value.ID == "" {
				err = fmt.Errorf("<target> requires an id")
				return false
			}
		case *StdIndex:
			if value.ID == "" {
				err = fmt.Errorf("<index> requires an id")
				return false
			}
			if _, exists := summary.indexIDs[value.ID]; exists {
				err = fmt.Errorf("duplicate index id %q", value.ID)
				return false
			}
			summary.hasIndexes = true
			summary.indexIDs[value.ID] = struct{}{}
		case *StdIndexEntry:
			if value.indexID == "" {
				err = fmt.Errorf("<index-entry> requires an index attribute")
				return false
			}
			if value.target == "" {
				err = fmt.Errorf("<index-entry> requires a target attribute")
				return false
			}
			summary.hasIndexes = true
			indexRefs[value.indexID] = struct{}{}
			summary.targetRefs[value.target] = struct{}{}
		}
		return true
	})
	if err != nil {
		return nil, err
	}
	var missingIndexes []string
	for indexID := range indexRefs {
		if _, ok := summary.indexIDs[indexID]; ok {
			continue
		}
		missingIndexes = append(missingIndexes, indexID)
	}
	if len(missingIndexes) > 0 {
		sort.Strings(missingIndexes)
		return nil, fmt.Errorf("missing index definition(s): %s", strings.Join(missingIndexes, ", "))
	}
	return summary, nil
}

func (d *StdDocument) printWithIndexes(w Writer) error {
	summary, err := d.describeLinking()
	if err != nil {
		return err
	}
	probe := newLayoutProbeWriter(w)
	// The first probe pass discovers destinations and index entries. If indexes
	// exist, later probe passes keep reflowing until the resolved page numbers
	// stop changing.
	snapshot, err := d.renderPass(probe, summary, nil, true)
	if err != nil {
		return err
	}
	if summary.hasIndexes {
		stable := false
		for pass := 1; pass < maxIndexLayoutPasses; pass++ {
			next, err := d.renderPass(probe, summary, snapshot, true)
			if err != nil {
				return err
			}
			if snapshot.Equal(next) {
				snapshot = next
				stable = true
				break
			}
			snapshot = next
		}
		if !stable {
			return fmt.Errorf("index layout did not stabilize after %d passes", maxIndexLayoutPasses)
		}
	}
	_, err = d.renderPass(w, summary, snapshot, false)
	return err
}

func (d *StdDocument) renderPass(w Writer, summary *documentLinkSummary, snapshot *documentIndexSnapshot, preflight bool) (*documentIndexSnapshot, error) {
	d.resetRenderState()
	ctx := newDocumentRenderContext(snapshot, preflight)
	d.renderContext = ctx
	defer func() {
		d.renderContext = nil
	}()
	if err := d.DrawContent(w); err != nil {
		return nil, err
	}
	return ctx.snapshot(summary)
}

func (d *StdDocument) resetRenderState() {
	d.documentPageNo = 0
	d.physicalPageNo = 0
	d.pendingStart = nil
	d.canvasCaptureStack = nil
	walkWidgets(d, func(widget Widget) bool {
		widget.SetPrinted(false)
		widget.SetVisible(true)
		widget.SetDisabled(false)
		widget.ClearResolvedWidth()
		widget.ClearResolvedHeight()
		switch value := widget.(type) {
		case *StdPage:
			value.flowPageIndex = 0
			value.flowItems = nil
			value.activeChildren = nil
		case *StdContainer:
			value.activeChildren = nil
		case *StdIndex:
			value.clearExpandedState()
			value.clearMeasuredGeometry()
		}
		return true
	})
	d.eachCanvas(func(canvas *StdCanvas) {
		resetCanvasWidgetRenderState(canvas)
	})
}

func registerPrintedWidgetMetadata(widget Widget, writer Writer) error {
	containerWidget, ok := widget.(interface{ Container() Container })
	if !ok {
		return nil
	}
	doc := documentForContainer(containerWidget.Container())
	if doc == nil || doc.renderContext == nil {
		return nil
	}
	if documentVisualCaptureActive(doc) {
		return nil
	}
	switch value := widget.(type) {
	case *StdIndexEntry:
		// Index entries are hidden metadata widgets. They never render directly;
		// instead, the active index widget consumes the collected entry later.
		doc.renderContext.registerIndexEntry(widget, value.indexID, value.target, value.ResolveText(doc))
		return nil
	}
	if identifier, ok := widget.(interface{ GetID() string }); ok {
		name := identifier.GetID()
		if name != "" {
			// Destination registration happens after layout so page/widget ids map
			// to final coordinates rather than guessed positions from parse time.
			x, y := widgetDestinationPoint(widget)
			doc.renderContext.registerDestination(name, doc.CurrentPageNo(), doc.CurrentPhysicalPageNo(), x, y)
			if registrar, ok := writer.(destinationRegistrar); ok {
				registrar.RegisterDestination(name, x, y)
			}
		}
	}
	return nil
}

func widgetDestinationPoint(widget Widget) (x, y float64) {
	switch value := widget.(type) {
	case *StdPage:
		return ContentLeft(value), ContentTop(value)
	default:
		return widget.Left(), widget.Top()
	}
}
