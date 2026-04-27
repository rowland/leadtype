package ltml

import (
	"fmt"
	"math"
	"strings"

	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/rich_text"
	"github.com/rowland/leadtype/wordbreaking"
)

const defaultLeaderText = "."

type StdLeader struct {
	StdSpan
	text string
}

type leaderInline interface {
	inlineText
	LeaderText() string
}

type leaderPieceSpec struct {
	leftPieces  []textPiece
	rightPieces []textPiece
	leaderPiece textPiece
	leader      leaderInline
}

func (l *StdLeader) AddText(string) {
	// Attribute-only inline placeholder.
}

func (l *StdLeader) Dynamic() bool {
	return true
}

func (l *StdLeader) Resolve(*StdDocument) string {
	return ""
}

func (l *StdLeader) LeaderText() string {
	if strings.TrimSpace(l.text) == "" {
		return defaultLeaderText
	}
	return l.text
}

func (l *StdLeader) SetAttrs(attrs map[string]string) {
	l.StdSpan.SetAttrs(attrs)
	l.text = attrs["text"]
}

func (l *StdLeader) SetContainer(container Container) error {
	if err := l.StdSpan.SetContainer(container); err != nil {
		return err
	}
	inlineContainer, ok := container.(InlineContainer)
	if !ok {
		return fmt.Errorf("leader must be child of an inline text container")
	}
	inlineContainer.AddInlineWithFont(l, l.explicitFont())
	return nil
}

func detectLeaderPieces(pieces []textPiece) (*leaderPieceSpec, bool) {
	splitIndex := -1
	var leaderPiece textPiece
	var leader leaderInline
	for idx, piece := range pieces {
		value, ok := piece.content.(leaderInline)
		if !ok {
			continue
		}
		if splitIndex >= 0 {
			return nil, false
		}
		splitIndex = idx
		leaderPiece = piece
		leader = value
	}
	if splitIndex < 0 {
		return nil, false
	}
	return &leaderPieceSpec{
		leftPieces:  append([]textPiece(nil), pieces[:splitIndex]...),
		rightPieces: append([]textPiece(nil), pieces[splitIndex+1:]...),
		leaderPiece: leaderPiece,
		leader:      leader,
	}, true
}

func prepareLeaderLines(w Writer, container Container, pieces []textPiece, width float64, wrap bool) ([]*rich_text.RichText, bool) {
	spec, ok := detectLeaderPieces(pieces)
	if !ok {
		return nil, false
	}
	leftText := richTextForTextPieces(w, container, spec.leftPieces, nil, container.Font())
	tailText := richTextForTextPieces(w, container, spec.rightPieces, nil, container.Font())

	if !wrap {
		// Labels stay single-line. When a leader is present we synthesize one
		// composed rich-text run, then let the host widget apply its usual
		// alignment, rotation, and optional shrink-to-fit rules.
		return []*rich_text.RichText{
			buildLeaderLine(w, container, spec.leaderPiece, spec.leader, leftText, tailText, max(width, 0)),
		}, true
	}

	if leftText == nil || leftText.Len() == 0 {
		return []*rich_text.RichText{
			buildLeaderLine(w, container, spec.leaderPiece, spec.leader, nil, tailText, max(width, 0)),
		}, true
	}

	flags := make([]wordbreaking.Flags, leftText.Len())
	wordbreaking.MarkRuneAttributes(leftText.String(), flags)
	lineWidths := []float64{max(width, 1)}
	lines := wrapRichTextToWidths(leftText, flags, lineWidths)

	// Leaders belong on the final rendered line. We iteratively wrap the left
	// side until the last line stabilizes with enough reserved space for the
	// leader fill and trailing tail text.
	finalLeftWidth := width - tailWidthWithLeaderGap(w, container, spec.leaderPiece, spec.leader, tailText)
	if finalLeftWidth <= 0 {
		finalLeftWidth = 1
	}
	for range 8 {
		if len(lines) == 0 {
			lines = []*rich_text.RichText{{}}
		}
		widths := make([]float64, len(lines))
		for idx := range widths {
			widths[idx] = max(width, 1)
		}
		widths[len(widths)-1] = finalLeftWidth
		next := wrapRichTextToWidths(leftText, flags, widths)
		if richTextLinesEqual(lines, next) {
			lines = next
			break
		}
		lines = next
	}
	if len(lines) == 0 {
		lines = []*rich_text.RichText{nil}
	}
	lines[len(lines)-1] = buildLeaderLine(w, container, spec.leaderPiece, spec.leader, lines[len(lines)-1], tailText, max(width, 0))
	return lines, true
}

func buildLeaderLine(w Writer, container Container, piece textPiece, leader leaderInline, leftLine *rich_text.RichText, tailText *rich_text.RichText, width float64) *rich_text.RichText {
	leaderText := buildLeaderFillText(w, container, piece, leader, leftLine, tailText, width)
	return combineRichText(leftLine, leaderText, tailText)
}

func tailWidthWithLeaderGap(w Writer, container Container, piece textPiece, leader leaderInline, tailText *rich_text.RichText) float64 {
	width := leaderGapWidth(w, container, piece, leader)
	if tailText != nil {
		width += tailText.Width()
	}
	return width
}

func richTextForPieceText(w Writer, container Container, piece textPiece, text string) *rich_text.RichText {
	if text == "" {
		return &rich_text.RichText{}
	}
	font, explicit := piece.Font(container.Font())
	if explicit {
		applyExplicitFontForContainer(w, container, font)
	} else {
		applyContainerFont(w, container)
	}
	rt, err := rich_text.New(text, w.Fonts(), w.FontSize(), piece.RichTextOptions(font.RichTextOptions()))
	if err != nil {
		debugf("richTextForPieceText: %v", err)
		return &rich_text.RichText{}
	}
	return rt
}

func leaderPatternText(leader leaderInline) string {
	text := leader.LeaderText()
	return text
}

func isDefaultDotLeader(leader leaderInline) bool {
	return strings.TrimSpace(leader.LeaderText()) == defaultLeaderText
}

func defaultDotLeaderMetrics(w Writer, container Container, piece textPiece) (dotWidth, preferredGap float64) {
	single := richTextForPieceText(w, container, piece, defaultLeaderText)
	if single.Len() == 0 || single.Width() <= 0 {
		return 0, 0
	}
	dotWidth = single.Width()
	space := richTextForPieceText(w, container, piece, " ")
	if space.Len() == 0 || space.Width() <= 0 {
		return dotWidth, 0
	}
	return dotWidth, space.Width()
}

func richTextForPieceTextWithOptions(w Writer, container Container, piece textPiece, text string, extra options.Options) *rich_text.RichText {
	if text == "" {
		return &rich_text.RichText{}
	}
	font, explicit := piece.Font(container.Font())
	if explicit {
		applyExplicitFontForContainer(w, container, font)
	} else {
		applyContainerFont(w, container)
	}
	opts := piece.RichTextOptions(font.RichTextOptions())
	for k, v := range extra {
		opts[k] = v
	}
	rt, err := rich_text.New(text, w.Fonts(), w.FontSize(), opts)
	if err != nil {
		debugf("richTextForPieceTextWithOptions: %v", err)
		return &rich_text.RichText{}
	}
	return rt
}

func leaderGapWidth(w Writer, container Container, piece textPiece, leader leaderInline) float64 {
	if isDefaultDotLeader(leader) {
		dotWidth, preferredGap := defaultDotLeaderMetrics(w, container, piece)
		if dotWidth > 0 {
			return (dotWidth * 2) + (preferredGap * 2)
		}
		return 8
	}
	pattern := leaderPatternText(leader)
	unit := richTextForPieceText(w, container, piece, pattern+pattern)
	if unit.Len() == 0 || unit.Width() == 0 {
		return 8
	}
	return unit.Width()
}

func buildLeaderFillText(w Writer, container Container, piece textPiece, leader leaderInline, leftLine *rich_text.RichText, tailText *rich_text.RichText, width float64) *rich_text.RichText {
	leftWidth := 0.0
	if leftLine != nil {
		leftWidth = leftLine.Width()
	}
	tailWidth := 0.0
	if tailText != nil {
		tailWidth = tailText.Width()
	}
	gapWidth := width - leftWidth - tailWidth
	if isDefaultDotLeader(leader) {
		dotWidth, preferredGap := defaultDotLeaderMetrics(w, container, piece)
		if dotWidth <= 0 {
			return nil
		}
		startGap := 0.0
		endGap := 0.0
		if preferredGap > 0 && gapWidth >= dotWidth+(2*preferredGap) {
			startGap = preferredGap
			endGap = preferredGap
		}
		usableGapWidth := gapWidth - startGap - endGap
		if usableGapWidth <= 0 {
			return nil
		}
		stride := math.Max(dotWidth+preferredGap, dotWidth)
		count := int(usableGapWidth / stride)
		if count < 1 && usableGapWidth >= dotWidth {
			count = 1
		}
		if count < 1 {
			return nil
		}
		charSpacing := (usableGapWidth - (float64(count) * dotWidth)) / float64(count)
		if charSpacing < 0 {
			charSpacing = 0
		}
		dots := richTextForPieceTextWithOptions(
			w,
			container,
			piece,
			strings.Repeat(defaultLeaderText, count),
			options.Options{"char_spacing": charSpacing},
		)
		var parts []*rich_text.RichText
		if startGap > 0 {
			parts = append(parts, richTextForPieceText(w, container, piece, " "))
		}
		parts = append(parts, dots)
		if endGap > 0 {
			parts = append(parts, richTextForPieceText(w, container, piece, " "))
		}
		return combineRichText(parts...)
	}
	unitText := leaderPatternText(leader)
	unit := richTextForPieceText(w, container, piece, unitText)
	if unit.Len() == 0 || unit.Width() <= 0 {
		return nil
	}
	if gapWidth <= unit.Width() {
		return nil
	}
	count := int(gapWidth / unit.Width())
	if count < 1 {
		return nil
	}
	return richTextForPieceText(w, container, piece, strings.Repeat(unitText, count))
}

func init() {
	registerTag(DefaultSpace, "leader", func() any { return &StdLeader{} })
}
