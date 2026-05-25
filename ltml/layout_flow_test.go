package ltml

import (
	"math"
	"testing"

	"github.com/rowland/leadtype/ltml/ltpdf"
	"github.com/rowland/leadtype/pdf"
)

func TestLayoutFlow_UsesAspectInferredDimensions(t *testing.T) {
	flow := positionedContainer(0, 0, 50, 200)
	flow.layout = &LayoutStyle{manager: "flow"}

	widget := &aspectRatioTestWidget{
		positionedTestWidget: positionedTestWidget{preferredWidth: 100, preferredHeight: 100},
		aspectRatio:          4,
	}
	widget.SetWidth(80)
	if err := widget.SetContainer(flow); err != nil {
		t.Fatal(err)
	}
	flow.AddChild(widget)

	flow.prepareForLayout(&labelTestWriter{t: t})
	LayoutFlow(flow, flow.layout, &labelTestWriter{t: t})

	if got := widget.Width(); got != 80 {
		t.Fatalf("widget width = %v, want authored width 80", got)
	}
	if got := widget.Height(); got != 20 {
		t.Fatalf("widget height = %v, want aspect-inferred height 20", got)
	}
}

func TestLayoutFlow_RecomputesUnspecifiedCardHeightsAfterVBoxProbe(t *testing.T) {
	writer := &labelTestWriter{t: t}
	flowStyle := &LayoutStyle{manager: "flow", hpadding: 12, vpadding: 12}

	buildFlow := func() (*StdContainer, []*StdContainer) {
		flow := &StdContainer{}
		flow.layout = flowStyle
		flow.SetLeft(0)
		flow.SetTop(0)
		flow.SetWidth(550)
		flow.SetHeight(300)

		cards := make([]*StdContainer, 0, 5)
		for _, contentWidth := range []float64{76, 68, 60, 52, 44} {
			card := &StdContainer{}
			card.layout = &LayoutStyle{manager: "vbox", vpadding: 8}
			card.SetWidth(92)
			card.SetAttrs(map[string]string{"padding": "8pt"})
			if err := card.SetContainer(flow); err != nil {
				t.Fatal(err)
			}
			flow.AddChild(card)

			draw := &positionedTestWidget{preferredWidth: contentWidth, preferredHeight: contentWidth}
			draw.SetWidth(contentWidth)
			if err := draw.SetContainer(card); err != nil {
				t.Fatal(err)
			}
			card.AddChild(draw)

			caption := &positionedTestWidget{preferredWidth: 76, preferredHeight: 12}
			caption.SetWidth(76)
			if err := caption.SetContainer(card); err != nil {
				t.Fatal(err)
			}
			card.AddChild(caption)

			cards = append(cards, card)
		}

		return flow, cards
	}

	directFlow, directCards := buildFlow()
	LayoutFlow(directFlow, flowStyle, writer)
	directHeights := make([]float64, len(directCards))
	directTops := make([]float64, len(directCards))
	directVisible := make([]bool, len(directCards))
	for i, card := range directCards {
		directHeights[i] = card.Height()
		directTops[i] = card.Top()
		directVisible[i] = card.Visible()
	}

	probedFlow, probedCards := buildFlow()
	outer := positionedContainer(0, 0, 550, 300)
	outer.layout = &LayoutStyle{manager: "vbox"}
	if err := probedFlow.SetContainer(outer); err != nil {
		t.Fatal(err)
	}
	outer.AddChild(probedFlow)

	LayoutVBox(outer, outer.layout, writer)

	for i, card := range probedCards {
		if got := card.Visible(); got != directVisible[i] {
			t.Fatalf("card %d visible = %v, want %v", i, got, directVisible[i])
		}
		if got := card.Height(); math.Abs(got-directHeights[i]) > 0.001 {
			t.Fatalf("card %d height = %v, want %v", i, got, directHeights[i])
		}
		if got := card.Top(); math.Abs(got-directTops[i]) > 0.001 {
			t.Fatalf("card %d top = %v, want %v", i, got, directTops[i])
		}
	}
}

func TestCanvasSample_PostageStampFlowCardsRemainVisible(t *testing.T) {
	doc, err := ParseFile(sampleFile("test_054_canvas_draw.ltml"))
	if err != nil {
		t.Fatal(err)
	}
	ttFonts, afmFonts, err := loadSampleFontSources()
	if err != nil {
		t.Fatal(err)
	}
	writer := &ltpdf.DocWriter{DocWriter: pdf.NewDocWriter()}
	writer.AddFontSource(ttFonts)
	writer.AddFontSource(afmFonts)
	if err := doc.Print(writer); err != nil {
		t.Fatal(err)
	}

	var flow *StdContainer
	walkWidgets(doc.Root(), func(widget Widget) bool {
		container, ok := widget.(*StdContainer)
		if !ok || container.layout == nil || container.layout.manager != "flow" {
			return true
		}
		flow = container
		return false
	})
	if flow == nil {
		t.Fatal("flow container not found in canvas sample")
	}
	if len(flow.children) != 5 {
		t.Fatalf("flow child count = %d, want 5", len(flow.children))
	}
	for i, child := range flow.children {
		card, ok := child.(*StdContainer)
		if !ok {
			t.Fatalf("flow child %d type = %T, want *StdContainer", i, child)
		}
		if !card.Visible() {
			t.Fatalf("flow child %d visible = false (top=%v height=%v bottom=%v flowBottom=%v)", i, card.Top(), card.Height(), card.Bottom(), flow.Bottom())
		}
	}
}
