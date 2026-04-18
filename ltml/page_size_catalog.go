package ltml

import "strings"

type BuiltInPageSize struct {
	ID     string
	Family string
	Width  float64
	Height float64
}

var builtInPageSizes = []BuiltInPageSize{
	{ID: "A0", Family: "ISO A", Width: mmToPoints(841), Height: mmToPoints(1189)},
	{ID: "A1", Family: "ISO A", Width: mmToPoints(594), Height: mmToPoints(841)},
	{ID: "A2", Family: "ISO A", Width: mmToPoints(420), Height: mmToPoints(594)},
	{ID: "A3", Family: "ISO A", Width: mmToPoints(297), Height: mmToPoints(420)},
	{ID: "A4", Family: "ISO A", Width: mmToPoints(210), Height: mmToPoints(297)},
	{ID: "A5", Family: "ISO A", Width: mmToPoints(148), Height: mmToPoints(210)},
	{ID: "A6", Family: "ISO A", Width: mmToPoints(105), Height: mmToPoints(148)},
	{ID: "A7", Family: "ISO A", Width: mmToPoints(74), Height: mmToPoints(105)},
	{ID: "A8", Family: "ISO A", Width: mmToPoints(52), Height: mmToPoints(74)},
	{ID: "A9", Family: "ISO A", Width: mmToPoints(37), Height: mmToPoints(52)},
	{ID: "A10", Family: "ISO A", Width: mmToPoints(26), Height: mmToPoints(37)},
	{ID: "B0", Family: "ISO B", Width: mmToPoints(1000), Height: mmToPoints(1414)},
	{ID: "B1", Family: "ISO B", Width: mmToPoints(707), Height: mmToPoints(1000)},
	{ID: "B2", Family: "ISO B", Width: mmToPoints(500), Height: mmToPoints(707)},
	{ID: "B3", Family: "ISO B", Width: mmToPoints(353), Height: mmToPoints(500)},
	{ID: "B4", Family: "ISO B", Width: mmToPoints(250), Height: mmToPoints(353)},
	{ID: "B5", Family: "ISO B", Width: mmToPoints(176), Height: mmToPoints(250)},
	{ID: "B6", Family: "ISO B", Width: mmToPoints(125), Height: mmToPoints(176)},
	{ID: "B7", Family: "ISO B", Width: mmToPoints(88), Height: mmToPoints(125)},
	{ID: "B8", Family: "ISO B", Width: mmToPoints(62), Height: mmToPoints(88)},
	{ID: "B9", Family: "ISO B", Width: mmToPoints(44), Height: mmToPoints(62)},
	{ID: "B10", Family: "ISO B", Width: mmToPoints(31), Height: mmToPoints(44)},
	{ID: "C0", Family: "ISO C", Width: mmToPoints(917), Height: mmToPoints(1297)},
	{ID: "C1", Family: "ISO C", Width: mmToPoints(648), Height: mmToPoints(917)},
	{ID: "C2", Family: "ISO C", Width: mmToPoints(458), Height: mmToPoints(648)},
	{ID: "C3", Family: "ISO C", Width: mmToPoints(324), Height: mmToPoints(458)},
	{ID: "C4", Family: "ISO C", Width: mmToPoints(229), Height: mmToPoints(324)},
	{ID: "C5", Family: "ISO C", Width: mmToPoints(162), Height: mmToPoints(229)},
	{ID: "C6", Family: "ISO C", Width: mmToPoints(114), Height: mmToPoints(162)},
	{ID: "C7", Family: "ISO C", Width: mmToPoints(81), Height: mmToPoints(114)},
	{ID: "C8", Family: "ISO C", Width: mmToPoints(57), Height: mmToPoints(81)},
	{ID: "C9", Family: "ISO C", Width: mmToPoints(40), Height: mmToPoints(57)},
	{ID: "C10", Family: "ISO C", Width: mmToPoints(28), Height: mmToPoints(40)},
	{ID: "RA0", Family: "ISO RA", Width: mmToPoints(860), Height: mmToPoints(1220)},
	{ID: "RA1", Family: "ISO RA", Width: mmToPoints(610), Height: mmToPoints(860)},
	{ID: "RA2", Family: "ISO RA", Width: mmToPoints(430), Height: mmToPoints(610)},
	{ID: "RA3", Family: "ISO RA", Width: mmToPoints(305), Height: mmToPoints(430)},
	{ID: "RA4", Family: "ISO RA", Width: mmToPoints(215), Height: mmToPoints(305)},
	{ID: "SRA0", Family: "ISO SRA", Width: mmToPoints(900), Height: mmToPoints(1280)},
	{ID: "SRA1", Family: "ISO SRA", Width: mmToPoints(640), Height: mmToPoints(900)},
	{ID: "SRA2", Family: "ISO SRA", Width: mmToPoints(450), Height: mmToPoints(640)},
	{ID: "SRA3", Family: "ISO SRA", Width: mmToPoints(320), Height: mmToPoints(450)},
	{ID: "SRA4", Family: "ISO SRA", Width: mmToPoints(225), Height: mmToPoints(320)},
	{ID: "halfletter", Family: "North American", Width: inchesToPoints(5.5), Height: inchesToPoints(8.5)},
	{ID: "statement", Family: "North American", Width: inchesToPoints(5.5), Height: inchesToPoints(8.5)},
	{ID: "letter", Family: "North American", Width: inchesToPoints(8.5), Height: inchesToPoints(11)},
	{ID: "legal", Family: "North American", Width: inchesToPoints(8.5), Height: inchesToPoints(14)},
	{ID: "juniorlegal", Family: "North American", Width: inchesToPoints(5), Height: inchesToPoints(8)},
	{ID: "tabloid", Family: "North American", Width: inchesToPoints(11), Height: inchesToPoints(17)},
	{ID: "ledger", Family: "North American", Width: inchesToPoints(11), Height: inchesToPoints(17)},
	{ID: "governmentletter", Family: "North American", Width: inchesToPoints(8), Height: inchesToPoints(10.5)},
	{ID: "governmentlegal", Family: "North American", Width: inchesToPoints(8.5), Height: inchesToPoints(13)},
	{ID: "executive", Family: "North American", Width: inchesToPoints(7.25), Height: inchesToPoints(10.5)},
	{ID: "folio", Family: "North American", Width: inchesToPoints(8.5), Height: inchesToPoints(13)},
	{ID: "quarto", Family: "North American", Width: inchesToPoints(8), Height: inchesToPoints(10)},
	{ID: "ansia", Family: "ANSI", Width: inchesToPoints(8.5), Height: inchesToPoints(11)},
	{ID: "ansib", Family: "ANSI", Width: inchesToPoints(11), Height: inchesToPoints(17)},
	{ID: "ansic", Family: "ANSI", Width: inchesToPoints(17), Height: inchesToPoints(22)},
	{ID: "ansid", Family: "ANSI", Width: inchesToPoints(22), Height: inchesToPoints(34)},
	{ID: "ansie", Family: "ANSI", Width: inchesToPoints(34), Height: inchesToPoints(44)},
	{ID: "archa", Family: "ARCH", Width: inchesToPoints(9), Height: inchesToPoints(12)},
	{ID: "archb", Family: "ARCH", Width: inchesToPoints(12), Height: inchesToPoints(18)},
	{ID: "archc", Family: "ARCH", Width: inchesToPoints(18), Height: inchesToPoints(24)},
	{ID: "archd", Family: "ARCH", Width: inchesToPoints(24), Height: inchesToPoints(36)},
	{ID: "arche", Family: "ARCH", Width: inchesToPoints(36), Height: inchesToPoints(48)},
	{ID: "arche1", Family: "ARCH", Width: inchesToPoints(30), Height: inchesToPoints(42)},
	{ID: "arche2", Family: "ARCH", Width: inchesToPoints(26), Height: inchesToPoints(38)},
	{ID: "arche3", Family: "ARCH", Width: inchesToPoints(27), Height: inchesToPoints(39)},
}

var builtInPageSizesByID = initBuiltInPageSizes()
var defaultPageStyles = initDefaultPageStyles()

func inchesToPoints(inches float64) float64 {
	return inches * 72.0
}

func mmToPoints(mm float64) float64 {
	return mm * 72.0 / 25.4
}

func initBuiltInPageSizes() map[string]BuiltInPageSize {
	byID := make(map[string]BuiltInPageSize, len(builtInPageSizes))
	for _, spec := range builtInPageSizes {
		byID[normalizeBuiltInPageSizeID(spec.ID)] = spec
	}
	return byID
}

func initDefaultPageStyles() map[string]*PageStyle {
	styles := make(map[string]*PageStyle, len(builtInPageSizes))
	for _, spec := range builtInPageSizes {
		styles[spec.ID] = &PageStyle{id: spec.ID, width: spec.Width, height: spec.Height}
	}
	return styles
}

func normalizeBuiltInPageSizeID(id string) string {
	return strings.ToLower(strings.TrimSpace(id))
}

func lookupBuiltInPageSize(id string) (BuiltInPageSize, bool) {
	spec, ok := builtInPageSizesByID[normalizeBuiltInPageSizeID(id)]
	return spec, ok
}
