package ltml

import (
	"strings"

	"github.com/rowland/leadtype/rich_text"
	"github.com/rowland/leadtype/wordbreaking"
)

type StdIndex struct {
	StdWidget
	source          *StdIndex
	entriesOverride []resolvedIndexEntry
}

type preparedIndexEntry struct {
	labelLines []*rich_text.RichText
	pageText   *rich_text.RichText
	dots       *rich_text.RichText
	height     float64
	pageX      float64
	target     string
}

// targetLinkWriter lets StdIndex attach one row-sized internal link annotation
// without changing the LTML Writer interface for all callers.
type targetLinkWriter interface {
	AddTargetLink(x, y, width, height float64, target string) error
}

func (i *StdIndex) PreferredHeight(w Writer) float64 {
	if i.height != 0 {
		return i.height
	}
	return NonContentHeight(i) + i.contentHeight(w)
}

func (i *StdIndex) PreferredWidth(Writer) float64 {
	if i.width != 0 {
		return i.width
	}
	return 0
}

func (i *StdIndex) SplitForHeight(avail float64, w Writer) (*SplitResult, error) {
	entries := i.currentEntries()
	if len(entries) < 2 {
		return nil, nil
	}
	avail -= NonContentHeight(i)
	if avail <= 0 {
		return nil, nil
	}
	prepared := i.preparedEntries(w)
	fit := 0
	height := 0.0
	for idx, entry := range prepared {
		if height+entry.height > avail {
			break
		}
		height += entry.height
		fit = idx + 1
	}
	if fit == 0 || fit >= len(entries) {
		return nil, nil
	}
	return &SplitResult{
		Head: i.cloneWithEntries(entries[:fit]),
		Tail: i.cloneWithEntries(entries[fit:]),
	}, nil
}

func (i *StdIndex) clearSplitOverride() {
	i.entriesOverride = nil
}

func (i *StdIndex) cloneWithEntries(entries []resolvedIndexEntry) *StdIndex {
	clone := *i
	clone.source = i.sourceKey()
	clone.entriesOverride = append([]resolvedIndexEntry(nil), entries...)
	clone.printed = false
	clone.invisible = false
	clone.disabled = false
	clone.path = ""
	return &clone
}

func (i *StdIndex) contentHeight(w Writer) float64 {
	height := 0.0
	for _, entry := range i.preparedEntries(w) {
		height += entry.height
	}
	return height
}

func (i *StdIndex) currentEntries() []resolvedIndexEntry {
	if i.entriesOverride != nil {
		return i.entriesOverride
	}
	doc := documentForContainer(i.container)
	if doc == nil {
		return nil
	}
	return doc.activeIndexEntries(i.ID)
}

func (i *StdIndex) DrawContent(w Writer) error {
	top := ContentTop(i)
	left := ContentLeft(i)
	for _, entry := range i.preparedEntries(w) {
		entryTop := top
		if len(entry.labelLines) == 0 {
			baseline := top + entry.pageText.Ascent()
			w.MoveTo(entry.pageX, baseline)
			w.PrintRichText(entry.pageText)
			top += entry.height
		} else {
			for lineIndex, line := range entry.labelLines {
				baseline := top + line.Ascent()
				w.MoveTo(left, baseline)
				w.PrintRichText(line)
				if lineIndex == 0 {
					if entry.dots != nil && entry.dots.Len() > 0 {
						w.MoveTo(left+line.Width(), baseline)
						w.PrintRichText(entry.dots)
					}
					w.MoveTo(entry.pageX, baseline)
					w.PrintRichText(entry.pageText)
				}
				top += line.Leading() * w.LineSpacing()
			}
			last := entry.labelLines[len(entry.labelLines)-1]
			top -= last.Height() * (w.LineSpacing() - 1)
			top -= last.LineGap()
		}
		if adder, ok := w.(targetLinkWriter); ok && entry.target != "" {
			// Viewers were more reliable with one link covering the whole row than
			// several tiny adjacent annotations for label, leaders, and page no.
			if err := adder.AddTargetLink(left, entryTop, ContentRight(i)-left, entry.height, entry.target); err != nil {
				return err
			}
		}
	}
	return nil
}

func (i *StdIndex) preparedEntries(w Writer) []preparedIndexEntry {
	entries := i.currentEntries()
	if len(entries) == 0 {
		return nil
	}
	result := make([]preparedIndexEntry, 0, len(entries))
	for _, entry := range entries {
		// Rebuild measurement data from the active snapshot on each pass so the
		// index can respond to page-number changes while convergence settles.
		pageText := i.richTextForString(w, formatPageNo(entry.PageNo))
		pageX := ContentRight(i) - pageText.Width()
		labelWidth := pageX - ContentLeft(i) - i.gapWidth(w)
		if labelWidth < 0 {
			labelWidth = 0
		}
		labelText := i.richTextForString(w, entry.Label)
		labelLines := i.wrapLines(labelText, labelWidth)
		dots := i.dotLeader(w, labelLines, pageX)
		height := i.linesHeight(w, labelLines, pageText)
		result = append(result, preparedIndexEntry{
			labelLines: labelLines,
			pageText:   pageText,
			dots:       dots,
			height:     height,
			pageX:      pageX,
			target:     entry.Target,
		})
	}
	return result
}

func (i *StdIndex) dotLeader(w Writer, labelLines []*rich_text.RichText, pageX float64) *rich_text.RichText {
	if len(labelLines) == 0 {
		return nil
	}
	dot := i.richTextForString(w, ".")
	if dot.Len() == 0 || dot.Width() <= 0 {
		return nil
	}
	gapWidth := pageX - ContentLeft(i) - labelLines[0].Width()
	if gapWidth <= dot.Width() {
		return nil
	}
	count := int(gapWidth / dot.Width())
	if count < 2 {
		return nil
	}
	return i.richTextForString(w, strings.Repeat(".", count))
}

func (i *StdIndex) gapWidth(w Writer) float64 {
	dot := i.richTextForString(w, "..")
	if dot.Len() == 0 || dot.Width() == 0 {
		return 8
	}
	return dot.Width()
}

func (i *StdIndex) linesHeight(w Writer, labelLines []*rich_text.RichText, pageText *rich_text.RichText) float64 {
	if len(labelLines) == 0 {
		if pageText == nil || pageText.Len() == 0 {
			return 0
		}
		return pageText.Leading()*w.LineSpacing() - pageText.Height()*(w.LineSpacing()-1) - pageText.LineGap()
	}
	height := 0.0
	for _, line := range labelLines {
		height += line.Leading() * w.LineSpacing()
	}
	last := labelLines[len(labelLines)-1]
	height -= last.Height() * (w.LineSpacing() - 1)
	height -= last.LineGap()
	return height
}

func (i *StdIndex) richTextForString(w Writer, text string) *rich_text.RichText {
	if text == "" {
		return &rich_text.RichText{}
	}
	if i.font != nil {
		i.font.applyWithSize(w, i.font.ResolveAgainstBase(rootFontSizeForContainer(i.container)))
	} else {
		i.Font().applyWithSize(w, effectiveFontSizeForContainer(i.container))
	}
	rt, err := rich_text.New(text, w.Fonts(), w.FontSize(), i.Font().RichTextOptions())
	if err != nil {
		debugf("StdIndex.richTextForString: %v", err)
		return &rich_text.RichText{}
	}
	return rt
}

func (i *StdIndex) sourceKey() *StdIndex {
	if i.source != nil {
		return i.source
	}
	return i
}

func (i *StdIndex) wrapLines(rt *rich_text.RichText, width float64) []*rich_text.RichText {
	if rt == nil || rt.Len() == 0 {
		return nil
	}
	flags := make([]wordbreaking.Flags, rt.Len())
	wordbreaking.MarkRuneAttributes(rt.String(), flags)
	return rt.WrapToWidth(width, flags, false)
}

func init() {
	registerTag(DefaultSpace, "index", func() any { return &StdIndex{} })
}

var _ HasAttrs = (*StdIndex)(nil)
var _ Identifier = (*StdIndex)(nil)
var _ Splittable = (*StdIndex)(nil)
var _ WantsContainer = (*StdIndex)(nil)
