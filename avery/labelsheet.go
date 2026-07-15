package avery

import (
	"fmt"
	"slices"
	"strings"

	"github.com/rowland/leadtype/colors"
	"github.com/rowland/leadtype/ltml"
	"github.com/rowland/leadtype/options"
)

const Namespace = "avery"

type Stock struct {
	ID              string
	SheetWidth      float64
	SheetHeight     float64
	LabelWidth      float64
	LabelHeight     float64
	Columns         int
	Rows            int
	LeftMargin      float64
	TopMargin       float64
	HorizontalPitch float64
	VerticalPitch   float64
	Aliases         []string
}

type LabelSheet struct {
	ltml.StdContainer
	stockID     string
	showMetrics bool
	showOutline bool
}

type Label struct {
	ltml.StdContainer
	text *ltml.StdParagraph
}

var builtInStocks = []Stock{
	{
		ID:              "avery5160",
		SheetWidth:      inchesToPoints(8.5),
		SheetHeight:     inchesToPoints(11),
		LabelWidth:      inchesToPoints(2.625),
		LabelHeight:     inchesToPoints(1.0),
		Columns:         3,
		Rows:            10,
		LeftMargin:      inchesToPoints(0.1875),
		TopMargin:       inchesToPoints(0.5),
		HorizontalPitch: inchesToPoints(2.75),
		VerticalPitch:   inchesToPoints(1.0),
		Aliases:         []string{"5160", "8160"},
	},
	{
		ID:              "avery5161",
		SheetWidth:      inchesToPoints(8.5),
		SheetHeight:     inchesToPoints(11),
		LabelWidth:      inchesToPoints(4.0),
		LabelHeight:     inchesToPoints(1.0),
		Columns:         2,
		Rows:            10,
		LeftMargin:      inchesToPoints(0.1875),
		TopMargin:       inchesToPoints(0.5),
		HorizontalPitch: inchesToPoints(4.125),
		VerticalPitch:   inchesToPoints(1.0),
		Aliases:         []string{"5161", "8161"},
	},
	{
		ID:              "avery5162",
		SheetWidth:      inchesToPoints(8.5),
		SheetHeight:     inchesToPoints(11),
		LabelWidth:      inchesToPoints(4.0),
		LabelHeight:     inchesToPoints(4.0 / 3.0),
		Columns:         2,
		Rows:            7,
		LeftMargin:      inchesToPoints(0.1875),
		TopMargin:       inchesToPoints(0.5),
		HorizontalPitch: inchesToPoints(4.125),
		VerticalPitch:   inchesToPoints(4.0 / 3.0),
		Aliases:         []string{"5162", "8162"},
	},
	{
		ID:              "avery5163",
		SheetWidth:      inchesToPoints(8.5),
		SheetHeight:     inchesToPoints(11),
		LabelWidth:      inchesToPoints(4.0),
		LabelHeight:     inchesToPoints(2.0),
		Columns:         2,
		Rows:            5,
		LeftMargin:      inchesToPoints(0.1875),
		TopMargin:       inchesToPoints(0.5),
		HorizontalPitch: inchesToPoints(4.125),
		VerticalPitch:   inchesToPoints(2.0),
		Aliases:         []string{"5163", "8163"},
	},
	{
		ID:              "avery5164",
		SheetWidth:      inchesToPoints(8.5),
		SheetHeight:     inchesToPoints(11),
		LabelWidth:      inchesToPoints(10.0 / 3.0),
		LabelHeight:     inchesToPoints(4.0),
		Columns:         2,
		Rows:            3,
		LeftMargin:      inchesToPoints(0.15625),
		TopMargin:       inchesToPoints(0.5),
		HorizontalPitch: inchesToPoints(4.1875),
		VerticalPitch:   inchesToPoints(3.5),
		Aliases:         []string{"5164", "8164"},
	},
	{
		ID:              "avery5167",
		SheetWidth:      inchesToPoints(8.5),
		SheetHeight:     inchesToPoints(11),
		LabelWidth:      inchesToPoints(1.75),
		LabelHeight:     inchesToPoints(0.5),
		Columns:         4,
		Rows:            20,
		LeftMargin:      inchesToPoints(0.25),
		TopMargin:       inchesToPoints(0.5),
		HorizontalPitch: inchesToPoints(2.0),
		VerticalPitch:   inchesToPoints(0.5),
		Aliases:         []string{"5167", "8167"},
	},
	{
		ID:              "avery5366",
		SheetWidth:      inchesToPoints(8.5),
		SheetHeight:     inchesToPoints(11),
		LabelWidth:      inchesToPoints(3.4375),
		LabelHeight:     inchesToPoints(2.0 / 3.0),
		Columns:         2,
		Rows:            15,
		LeftMargin:      inchesToPoints(0.5625),
		TopMargin:       inchesToPoints(0.5),
		HorizontalPitch: inchesToPoints(3.875),
		VerticalPitch:   inchesToPoints(2.0 / 3.0),
		Aliases:         []string{"5366"},
	},
	{
		ID:              "avery5395",
		SheetWidth:      inchesToPoints(8.5),
		SheetHeight:     inchesToPoints(11),
		LabelWidth:      inchesToPoints(3.375),
		LabelHeight:     inchesToPoints(7.0 / 3.0),
		Columns:         2,
		Rows:            4,
		LeftMargin:      inchesToPoints(0.6875),
		TopMargin:       inchesToPoints(0.55),
		HorizontalPitch: inchesToPoints(3.75),
		VerticalPitch:   inchesToPoints(2.5233333333333334),
		Aliases:         []string{"5395"},
	},
}

var builtInStocksByID = initBuiltInStocks()

func init() {
	if err := ltml.RegisterTag(Namespace, "labelsheet", func() any { return &LabelSheet{} }); err != nil {
		panic(err)
	}
	if err := ltml.RegisterTag(Namespace, "label", func() any { return &Label{} }); err != nil {
		panic(err)
	}
}

func LookupStock(id string) (Stock, bool) {
	stock, ok := builtInStocksByID[normalizeStockID(id)]
	return stock, ok
}

func Stocks() []Stock {
	out := make([]Stock, len(builtInStocks))
	copy(out, builtInStocks)
	return out
}

func (ls *LabelSheet) BeforePrint(w ltml.Writer) error {
	stock, ok := LookupStock(ls.stockID)
	if !ok {
		return fmt.Errorf("unknown avery labelsheet stock %q", ls.stockID)
	}
	if len(ls.Widgets()) > stock.Columns*stock.Rows {
		return fmt.Errorf("avery labelsheet stock %q holds %d labels, got %d", stock.ID, stock.Columns*stock.Rows, len(ls.Widgets()))
	}
	return ls.StdContainer.BeforePrint(w)
}

func (ls *LabelSheet) DrawContent(w ltml.Writer) error {
	stock, ok := LookupStock(ls.stockID)
	if !ok {
		return fmt.Errorf("unknown avery labelsheet stock %q", ls.stockID)
	}
	if ls.showOutline {
		w.SetLineColor(colors.DimGray)
		w.SetLineWidth(0.4)
		w.SetLineDashPattern("solid")
		w.SetLineCapStyle("butt_cap")
		for row := 0; row < stock.Rows; row++ {
			for col := 0; col < stock.Columns; col++ {
				x := ls.Left() + stock.LeftMargin + float64(col)*stock.HorizontalPitch
				y := ls.Top() + stock.TopMargin + float64(row)*stock.VerticalPitch
				w.Rectangle2(x, y, stock.LabelWidth, stock.LabelHeight, true, false, nil, false, false)
			}
		}
	}
	if ls.showMetrics {
		if _, err := w.SetFont("Helvetica", 10, options.Options{"color": colors.DimGray}); err != nil {
			return err
		}
		w.MoveTo(ls.Left()+6, ls.Top()+14)
		if err := w.Print(fmt.Sprintf("Avery stock: %s  %.3fin x %.3fin  %dx%d  order=%s", stock.ID, stock.LabelWidth/72.0, stock.LabelHeight/72.0, stock.Columns, stock.Rows, orderName(ls.Order()))); err != nil {
			return err
		}
	}
	children := slices.Clone(ls.Widgets())
	slices.SortStableFunc(children, func(a, b ltml.Widget) int {
		return a.ZIndex() - b.ZIndex()
	})
	for _, child := range children {
		if !child.Visible() || child.Disabled() {
			continue
		}
		if err := ltml.Print(child, w); err != nil {
			return err
		}
	}
	return nil
}

func (ls *LabelSheet) LayoutWidget(w ltml.Writer) error {
	stock, ok := LookupStock(ls.stockID)
	if !ok {
		return nil
	}
	ls.applySheetGeometry(stock)
	for i, child := range ls.Widgets() {
		row, col := slotPosition(i, stock, ls.Order())
		child.SetPosition(ltml.Static)
		child.SetVisible(true)
		child.SetLeft(ls.Left() + stock.LeftMargin + float64(col)*stock.HorizontalPitch)
		child.SetTop(ls.Top() + stock.TopMargin + float64(row)*stock.VerticalPitch)
		child.SetWidth(stock.LabelWidth)
		child.SetHeight(stock.LabelHeight)
		if err := child.LayoutWidget(w); err != nil {
			return err
		}
	}
	return nil
}

func (ls *LabelSheet) PreferredHeight(ltml.Writer) (float64, error) {
	if ls.HeightIsSet() {
		return ls.Height(), nil
	}
	stock, ok := LookupStock(ls.stockID)
	if !ok {
		return 0, nil
	}
	return stock.SheetHeight, nil
}

func (ls *LabelSheet) PreferredWidth(ltml.Writer) (float64, error) {
	if ls.WidthIsSet() {
		return ls.Width(), nil
	}
	stock, ok := LookupStock(ls.stockID)
	if !ok {
		return 0, nil
	}
	return stock.SheetWidth, nil
}

func (ls *LabelSheet) SetAttrs(attrs map[string]string) {
	ls.StdContainer.SetAttrs(attrs)
	ls.stockID = attrs["stock"]
	ls.showMetrics = attrs["show-metrics"] == "true"
	ls.showOutline = attrs["show-outline"] == "true"
}

func (ls *LabelSheet) applySheetGeometry(stock Stock) {
	if !ls.WidthIsSet() {
		ls.SetWidth(stock.SheetWidth)
	}
	if !ls.HeightIsSet() {
		ls.SetHeight(stock.SheetHeight)
	}
}

func (l *Label) AddText(text string) {
	if strings.TrimSpace(text) == "" {
		return
	}
	l.ensureTextParagraph().AddText(text)
}

func (l *Label) SetAttrs(attrs map[string]string) {
	l.StdContainer.SetAttrs(attrs)
}

func (l *Label) ensureTextParagraph() *ltml.StdParagraph {
	if l.text != nil {
		return l.text
	}
	l.text = &ltml.StdParagraph{}
	l.text.SetScope(l.Scope())
	l.text.SetContainer(l)
	l.text.SetAttrs(map[string]string{"width": "100%", "height": "100%"})
	l.AddChild(l.text)
	return l.text
}

func inchesToPoints(inches float64) float64 {
	return inches * 72.0
}

func initBuiltInStocks() map[string]Stock {
	byID := make(map[string]Stock, len(builtInStocks)*2)
	for _, stock := range builtInStocks {
		byID[normalizeStockID(stock.ID)] = stock
		for _, alias := range stock.Aliases {
			byID[normalizeStockID(alias)] = stock
		}
	}
	return byID
}

func normalizeStockID(id string) string {
	return strings.ToLower(strings.TrimSpace(id))
}

func slotPosition(index int, stock Stock, order ltml.TableOrder) (row, col int) {
	switch order {
	case ltml.TableOrderCols:
		col = index / stock.Rows
		row = index % stock.Rows
	default:
		row = index / stock.Columns
		col = index % stock.Columns
	}
	return row, col
}

func orderName(order ltml.TableOrder) string {
	switch order {
	case ltml.TableOrderCols:
		return "cols"
	default:
		return "rows"
	}
}
