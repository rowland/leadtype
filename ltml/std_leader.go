package ltml

import (
	"fmt"
	"strings"

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

type preparedLeaderLayout struct {
	leftLines  []*rich_text.RichText
	tailText   *rich_text.RichText
	leaderText *rich_text.RichText
	tailX      float64
	height     float64
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

func prepareLeaderLayout(w Writer, container Container, pieces []textPiece, width float64, wrap bool) (*preparedLeaderLayout, bool) {
	spec, ok := detectLeaderPieces(pieces)
	if !ok {
		return nil, false
	}
	leftText := richTextForPieces(w, container, spec.leftPieces)
	tailText := richTextForPieces(w, container, spec.rightPieces)
	tailX := width
	if tailText != nil && tailText.Len() > 0 {
		tailX -= tailText.Width()
	}
	leftWidth := tailX - leaderGapWidth(w, container, spec.leaderPiece, spec.leader)
	if leftWidth < 0 {
		leftWidth = 0
	}
	var leftLines []*rich_text.RichText
	if wrap {
		leftLines = wrapRichText(leftText, leftWidth)
	} else if leftText != nil && leftText.Len() > 0 {
		leftLines = []*rich_text.RichText{leftText}
	}
	leaderText := buildLeaderFillText(w, container, spec.leaderPiece, spec.leader, leftLines, tailX)
	return &preparedLeaderLayout{
		leftLines:  leftLines,
		tailText:   tailText,
		leaderText: leaderText,
		tailX:      tailX,
		height:     leaderLinesHeight(w, leftLines, tailText),
	}, true
}

func richTextForPieces(w Writer, container Container, pieces []textPiece) *rich_text.RichText {
	if len(pieces) == 0 {
		return &rich_text.RichText{}
	}
	doc := documentForContainer(container)
	rt := &rich_text.RichText{}
	lastText := ""
	for _, piece := range pieces {
		font := applyTextPieceFontForContainer(w, container, piece, container.Font())
		text := piece.ResolvedText(doc)
		if text == "" {
			continue
		}
		if strings.HasSuffix(lastText, " ") && strings.HasPrefix(text, " ") {
			text = text[1:]
		}
		if text == "" {
			continue
		}
		next, err := rt.Add(text, w.Fonts(), w.FontSize(), piece.RichTextOptions(font.RichTextOptions()))
		if err != nil {
			debugf("richTextForPieces: %v", err)
			continue
		}
		rt = next
		lastText = text
	}
	return rt
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

func wrapRichText(rt *rich_text.RichText, width float64) []*rich_text.RichText {
	if rt == nil || rt.Len() == 0 {
		return nil
	}
	flags := make([]wordbreaking.Flags, rt.Len())
	wordbreaking.MarkRuneAttributes(rt.String(), flags)
	return rt.WrapToWidth(width, flags, false)
}

func leaderGapWidth(w Writer, container Container, piece textPiece, leader leaderInline) float64 {
	unit := leader.LeaderText()
	dot := richTextForPieceText(w, container, piece, unit+unit)
	if dot.Len() == 0 || dot.Width() == 0 {
		return 8
	}
	return dot.Width()
}

func buildLeaderFillText(w Writer, container Container, piece textPiece, leader leaderInline, leftLines []*rich_text.RichText, tailX float64) *rich_text.RichText {
	if len(leftLines) == 0 {
		return nil
	}
	unitText := leader.LeaderText()
	unit := richTextForPieceText(w, container, piece, unitText)
	if unit.Len() == 0 || unit.Width() <= 0 {
		return nil
	}
	gapWidth := tailX - leftLines[0].Width()
	if gapWidth <= unit.Width() {
		return nil
	}
	count := int(gapWidth / unit.Width())
	if count < 1 {
		return nil
	}
	return richTextForPieceText(w, container, piece, strings.Repeat(unitText, count))
}

func leaderLinesHeight(w Writer, leftLines []*rich_text.RichText, tailText *rich_text.RichText) float64 {
	if len(leftLines) == 0 {
		if tailText == nil || tailText.Len() == 0 {
			return 0
		}
		return tailText.Leading()*w.LineSpacing() - tailText.Height()*(w.LineSpacing()-1) - tailText.LineGap()
	}
	height := 0.0
	for _, line := range leftLines {
		height += line.Leading() * w.LineSpacing()
	}
	last := leftLines[len(leftLines)-1]
	height -= last.Height() * (w.LineSpacing() - 1)
	height -= last.LineGap()
	return height
}

func drawLeaderLayout(w Writer, left float64, top float64, layout *preparedLeaderLayout) {
	if layout == nil {
		return
	}
	if len(layout.leftLines) == 0 {
		if layout.tailText != nil && layout.tailText.Len() > 0 {
			baseline := top + layout.tailText.Ascent()
			w.MoveTo(left+layout.tailX, baseline)
			w.PrintRichText(layout.tailText)
		}
		return
	}
	currentTop := top
	for idx, line := range layout.leftLines {
		baseline := currentTop + line.Ascent()
		w.MoveTo(left, baseline)
		w.PrintRichText(line)
		if idx == 0 {
			if layout.leaderText != nil && layout.leaderText.Len() > 0 {
				w.MoveTo(left+line.Width(), baseline)
				w.PrintRichText(layout.leaderText)
			}
			if layout.tailText != nil && layout.tailText.Len() > 0 {
				w.MoveTo(left+layout.tailX, baseline)
				w.PrintRichText(layout.tailText)
			}
		}
		currentTop += line.Leading() * w.LineSpacing()
	}
}

func init() {
	registerTag(DefaultSpace, "leader", func() any { return &StdLeader{text: defaultLeaderText} })
}

var _ HasAttrs = (*StdLeader)(nil)
var _ HasText = (*StdLeader)(nil)
var _ HasFont = (*StdLeader)(nil)
