// Copyright 2011-2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package colors

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Color int32

type NamedColorEntry struct {
	Name  string
	Color Color
}

func (this Color) RGB() (r, g, b uint8) {
	b = uint8(this & 0xFF)
	g = uint8((this >> 8) & 0xFF)
	r = uint8((this >> 16) & 0xFF)
	return
}

func (this Color) RGB32() (r, g, b float32) {
	ri, gi, bi := this.RGB()
	r, g, b = float32(ri)/255, float32(gi)/255, float32(bi)/255
	return
}

func (this Color) RGB64() (r, g, b float64) {
	ri, gi, bi := this.RGB()
	r, g, b = float64(ri)/255, float64(gi)/255, float64(bi)/255
	return
}

func (this Color) String() string {
	if s, ok := ColorNames[this]; ok {
		return s
	}
	return fmt.Sprintf("%06X", int32(this))
}

const (
	AliceBlue            = Color(0xF0F8FF)
	AntiqueWhite         = Color(0xFAEBD7)
	Aqua                 = Color(0x00FFFF)
	Aquamarine           = Color(0x7FFFD4)
	Azure                = Color(0xF0FFFF)
	Beige                = Color(0xF5F5DC)
	Bisque               = Color(0xFFE4C4)
	Black                = Color(0x000000)
	BlanchedAlmond       = Color(0xFFEBCD)
	Blue                 = Color(0x0000FF)
	BlueViolet           = Color(0x8A2BE2)
	Brown                = Color(0xA52A2A)
	BurlyWood            = Color(0xDEB887)
	CadetBlue            = Color(0x5F9EA0)
	Chartreuse           = Color(0x7FFF00)
	Chocolate            = Color(0xD2691E)
	Coral                = Color(0xFF7F50)
	CornflowerBlue       = Color(0x6495ED)
	Cornsilk             = Color(0xFFF8DC)
	Crimson              = Color(0xDC143C)
	Cyan                 = Color(0x00FFFF)
	DarkBlue             = Color(0x00008B)
	DarkCyan             = Color(0x008B8B)
	DarkGoldenRod        = Color(0xB8860B)
	DarkGray             = Color(0xA9A9A9)
	DarkGrey             = Color(0xA9A9A9)
	DarkGreen            = Color(0x006400)
	DarkKhaki            = Color(0xBDB76B)
	DarkMagenta          = Color(0x8B008B)
	DarkOliveGreen       = Color(0x556B2F)
	DarkOrange           = Color(0xFF8C00)
	DarkOrchid           = Color(0x9932CC)
	DarkRed              = Color(0x8B0000)
	DarkSalmon           = Color(0xE9967A)
	DarkSeaGreen         = Color(0x8FBC8F)
	DarkSlateBlue        = Color(0x483D8B)
	DarkSlateGray        = Color(0x2F4F4F)
	DarkSlateGrey        = Color(0x2F4F4F)
	DarkTurquoise        = Color(0x00CED1)
	DarkViolet           = Color(0x9400D3)
	DeepPink             = Color(0xFF1493)
	DeepSkyBlue          = Color(0x00BFFF)
	DimGray              = Color(0x696969)
	DimGrey              = Color(0x696969)
	DodgerBlue           = Color(0x1E90FF)
	FireBrick            = Color(0xB22222)
	FloralWhite          = Color(0xFFFAF0)
	ForestGreen          = Color(0x228B22)
	Fuchsia              = Color(0xFF00FF)
	Gainsboro            = Color(0xDCDCDC)
	GhostWhite           = Color(0xF8F8FF)
	Gold                 = Color(0xFFD700)
	GoldenRod            = Color(0xDAA520)
	Gray                 = Color(0x808080)
	Grey                 = Color(0x808080)
	Green                = Color(0x008000)
	GreenYellow          = Color(0xADFF2F)
	HoneyDew             = Color(0xF0FFF0)
	HotPink              = Color(0xFF69B4)
	IndianRed            = Color(0xCD5C5C)
	Indigo               = Color(0x4B0082)
	Ivory                = Color(0xFFFFF0)
	Khaki                = Color(0xF0E68C)
	Lavender             = Color(0xE6E6FA)
	LavenderBlush        = Color(0xFFF0F5)
	LawnGreen            = Color(0x7CFC00)
	LemonChiffon         = Color(0xFFFACD)
	LightBlue            = Color(0xADD8E6)
	LightCoral           = Color(0xF08080)
	LightCyan            = Color(0xE0FFFF)
	LightGoldenRodYellow = Color(0xFAFAD2)
	LightGray            = Color(0xD3D3D3)
	LightGrey            = Color(0xD3D3D3)
	LightGreen           = Color(0x90EE90)
	LightPink            = Color(0xFFB6C1)
	LightSalmon          = Color(0xFFA07A)
	LightSeaGreen        = Color(0x20B2AA)
	LightSkyBlue         = Color(0x87CEFA)
	LightSlateGray       = Color(0x778899)
	LightSlateGrey       = Color(0x778899)
	LightSteelBlue       = Color(0xB0C4DE)
	LightYellow          = Color(0xFFFFE0)
	Lime                 = Color(0x00FF00)
	LimeGreen            = Color(0x32CD32)
	Linen                = Color(0xFAF0E6)
	Magenta              = Color(0xFF00FF)
	Maroon               = Color(0x800000)
	MediumAquaMarine     = Color(0x66CDAA)
	MediumBlue           = Color(0x0000CD)
	MediumOrchid         = Color(0xBA55D3)
	MediumPurple         = Color(0x9370D8)
	MediumSeaGreen       = Color(0x3CB371)
	MediumSlateBlue      = Color(0x7B68EE)
	MediumSpringGreen    = Color(0x00FA9A)
	MediumTurquoise      = Color(0x48D1CC)
	MediumVioletRed      = Color(0xC71585)
	MidnightBlue         = Color(0x191970)
	MintCream            = Color(0xF5FFFA)
	MistyRose            = Color(0xFFE4E1)
	Moccasin             = Color(0xFFE4B5)
	NavajoWhite          = Color(0xFFDEAD)
	Navy                 = Color(0x000080)
	OldLace              = Color(0xFDF5E6)
	Olive                = Color(0x808000)
	OliveDrab            = Color(0x6B8E23)
	Orange               = Color(0xFFA500)
	OrangeRed            = Color(0xFF4500)
	Orchid               = Color(0xDA70D6)
	PaleGoldenRod        = Color(0xEEE8AA)
	PaleGreen            = Color(0x98FB98)
	PaleTurquoise        = Color(0xAFEEEE)
	PaleVioletRed        = Color(0xD87093)
	PapayaWhip           = Color(0xFFEFD5)
	PeachPuff            = Color(0xFFDAB9)
	Peru                 = Color(0xCD853F)
	Pink                 = Color(0xFFC0CB)
	Plum                 = Color(0xDDA0DD)
	PowderBlue           = Color(0xB0E0E6)
	Purple               = Color(0x800080)
	Red                  = Color(0xFF0000)
	RosyBrown            = Color(0xBC8F8F)
	RoyalBlue            = Color(0x4169E1)
	SaddleBrown          = Color(0x8B4513)
	Salmon               = Color(0xFA8072)
	SandyBrown           = Color(0xF4A460)
	SeaGreen             = Color(0x2E8B57)
	SeaShell             = Color(0xFFF5EE)
	Sienna               = Color(0xA0522D)
	Silver               = Color(0xC0C0C0)
	SkyBlue              = Color(0x87CEEB)
	SlateBlue            = Color(0x6A5ACD)
	SlateGray            = Color(0x708090)
	SlateGrey            = Color(0x708090)
	Snow                 = Color(0xFFFAFA)
	SpringGreen          = Color(0x00FF7F)
	SteelBlue            = Color(0x4682B4)
	Tan                  = Color(0xD2B48C)
	Teal                 = Color(0x008080)
	Thistle              = Color(0xD8BFD8)
	Tomato               = Color(0xFF6347)
	Turquoise            = Color(0x40E0D0)
	Violet               = Color(0xEE82EE)
	Wheat                = Color(0xF5DEB3)
	White                = Color(0xFFFFFF)
	WhiteSmoke           = Color(0xF5F5F5)
	Yellow               = Color(0xFFFF00)
	YellowGreen          = Color(0x9ACD32)
)

var NamedColorEntries = []NamedColorEntry{
	{Name: "AliceBlue", Color: AliceBlue},
	{Name: "AntiqueWhite", Color: AntiqueWhite},
	{Name: "Aqua", Color: Aqua},
	{Name: "Aquamarine", Color: Aquamarine},
	{Name: "Azure", Color: Azure},
	{Name: "Beige", Color: Beige},
	{Name: "Bisque", Color: Bisque},
	{Name: "Black", Color: Black},
	{Name: "BlanchedAlmond", Color: BlanchedAlmond},
	{Name: "Blue", Color: Blue},
	{Name: "BlueViolet", Color: BlueViolet},
	{Name: "Brown", Color: Brown},
	{Name: "BurlyWood", Color: BurlyWood},
	{Name: "CadetBlue", Color: CadetBlue},
	{Name: "Chartreuse", Color: Chartreuse},
	{Name: "Chocolate", Color: Chocolate},
	{Name: "Coral", Color: Coral},
	{Name: "CornflowerBlue", Color: CornflowerBlue},
	{Name: "Cornsilk", Color: Cornsilk},
	{Name: "Crimson", Color: Crimson},
	{Name: "Cyan", Color: Cyan},
	{Name: "DarkBlue", Color: DarkBlue},
	{Name: "DarkCyan", Color: DarkCyan},
	{Name: "DarkGoldenRod", Color: DarkGoldenRod},
	{Name: "DarkGray", Color: DarkGray},
	{Name: "DarkGrey", Color: DarkGrey},
	{Name: "DarkGreen", Color: DarkGreen},
	{Name: "DarkKhaki", Color: DarkKhaki},
	{Name: "DarkMagenta", Color: DarkMagenta},
	{Name: "DarkOliveGreen", Color: DarkOliveGreen},
	{Name: "DarkOrange", Color: DarkOrange},
	{Name: "DarkOrchid", Color: DarkOrchid},
	{Name: "DarkRed", Color: DarkRed},
	{Name: "DarkSalmon", Color: DarkSalmon},
	{Name: "DarkSeaGreen", Color: DarkSeaGreen},
	{Name: "DarkSlateBlue", Color: DarkSlateBlue},
	{Name: "DarkSlateGray", Color: DarkSlateGray},
	{Name: "DarkSlateGrey", Color: DarkSlateGrey},
	{Name: "DarkTurquoise", Color: DarkTurquoise},
	{Name: "DarkViolet", Color: DarkViolet},
	{Name: "DeepPink", Color: DeepPink},
	{Name: "DeepSkyBlue", Color: DeepSkyBlue},
	{Name: "DimGray", Color: DimGray},
	{Name: "DimGrey", Color: DimGrey},
	{Name: "DodgerBlue", Color: DodgerBlue},
	{Name: "FireBrick", Color: FireBrick},
	{Name: "FloralWhite", Color: FloralWhite},
	{Name: "ForestGreen", Color: ForestGreen},
	{Name: "Fuchsia", Color: Fuchsia},
	{Name: "Gainsboro", Color: Gainsboro},
	{Name: "GhostWhite", Color: GhostWhite},
	{Name: "Gold", Color: Gold},
	{Name: "GoldenRod", Color: GoldenRod},
	{Name: "Gray", Color: Gray},
	{Name: "Grey", Color: Grey},
	{Name: "Green", Color: Green},
	{Name: "GreenYellow", Color: GreenYellow},
	{Name: "HoneyDew", Color: HoneyDew},
	{Name: "HotPink", Color: HotPink},
	{Name: "IndianRed", Color: IndianRed},
	{Name: "Indigo", Color: Indigo},
	{Name: "Ivory", Color: Ivory},
	{Name: "Khaki", Color: Khaki},
	{Name: "Lavender", Color: Lavender},
	{Name: "LavenderBlush", Color: LavenderBlush},
	{Name: "LawnGreen", Color: LawnGreen},
	{Name: "LemonChiffon", Color: LemonChiffon},
	{Name: "LightBlue", Color: LightBlue},
	{Name: "LightCoral", Color: LightCoral},
	{Name: "LightCyan", Color: LightCyan},
	{Name: "LightGoldenRodYellow", Color: LightGoldenRodYellow},
	{Name: "LightGray", Color: LightGray},
	{Name: "LightGrey", Color: LightGrey},
	{Name: "LightGreen", Color: LightGreen},
	{Name: "LightPink", Color: LightPink},
	{Name: "LightSalmon", Color: LightSalmon},
	{Name: "LightSeaGreen", Color: LightSeaGreen},
	{Name: "LightSkyBlue", Color: LightSkyBlue},
	{Name: "LightSlateGray", Color: LightSlateGray},
	{Name: "LightSlateGrey", Color: LightSlateGrey},
	{Name: "LightSteelBlue", Color: LightSteelBlue},
	{Name: "LightYellow", Color: LightYellow},
	{Name: "Lime", Color: Lime},
	{Name: "LimeGreen", Color: LimeGreen},
	{Name: "Linen", Color: Linen},
	{Name: "Magenta", Color: Magenta},
	{Name: "Maroon", Color: Maroon},
	{Name: "MediumAquaMarine", Color: MediumAquaMarine},
	{Name: "MediumBlue", Color: MediumBlue},
	{Name: "MediumOrchid", Color: MediumOrchid},
	{Name: "MediumPurple", Color: MediumPurple},
	{Name: "MediumSeaGreen", Color: MediumSeaGreen},
	{Name: "MediumSlateBlue", Color: MediumSlateBlue},
	{Name: "MediumSpringGreen", Color: MediumSpringGreen},
	{Name: "MediumTurquoise", Color: MediumTurquoise},
	{Name: "MediumVioletRed", Color: MediumVioletRed},
	{Name: "MidnightBlue", Color: MidnightBlue},
	{Name: "MintCream", Color: MintCream},
	{Name: "MistyRose", Color: MistyRose},
	{Name: "Moccasin", Color: Moccasin},
	{Name: "NavajoWhite", Color: NavajoWhite},
	{Name: "Navy", Color: Navy},
	{Name: "OldLace", Color: OldLace},
	{Name: "Olive", Color: Olive},
	{Name: "OliveDrab", Color: OliveDrab},
	{Name: "Orange", Color: Orange},
	{Name: "OrangeRed", Color: OrangeRed},
	{Name: "Orchid", Color: Orchid},
	{Name: "PaleGoldenRod", Color: PaleGoldenRod},
	{Name: "PaleGreen", Color: PaleGreen},
	{Name: "PaleTurquoise", Color: PaleTurquoise},
	{Name: "PaleVioletRed", Color: PaleVioletRed},
	{Name: "PapayaWhip", Color: PapayaWhip},
	{Name: "PeachPuff", Color: PeachPuff},
	{Name: "Peru", Color: Peru},
	{Name: "Pink", Color: Pink},
	{Name: "Plum", Color: Plum},
	{Name: "PowderBlue", Color: PowderBlue},
	{Name: "Purple", Color: Purple},
	{Name: "Red", Color: Red},
	{Name: "RosyBrown", Color: RosyBrown},
	{Name: "RoyalBlue", Color: RoyalBlue},
	{Name: "SaddleBrown", Color: SaddleBrown},
	{Name: "Salmon", Color: Salmon},
	{Name: "SandyBrown", Color: SandyBrown},
	{Name: "SeaGreen", Color: SeaGreen},
	{Name: "SeaShell", Color: SeaShell},
	{Name: "Sienna", Color: Sienna},
	{Name: "Silver", Color: Silver},
	{Name: "SkyBlue", Color: SkyBlue},
	{Name: "SlateBlue", Color: SlateBlue},
	{Name: "SlateGray", Color: SlateGray},
	{Name: "SlateGrey", Color: SlateGrey},
	{Name: "Snow", Color: Snow},
	{Name: "SpringGreen", Color: SpringGreen},
	{Name: "SteelBlue", Color: SteelBlue},
	{Name: "Tan", Color: Tan},
	{Name: "Teal", Color: Teal},
	{Name: "Thistle", Color: Thistle},
	{Name: "Tomato", Color: Tomato},
	{Name: "Turquoise", Color: Turquoise},
	{Name: "Violet", Color: Violet},
	{Name: "Wheat", Color: Wheat},
	{Name: "White", Color: White},
	{Name: "WhiteSmoke", Color: WhiteSmoke},
	{Name: "Yellow", Color: Yellow},
	{Name: "YellowGreen", Color: YellowGreen},
}

var NamedColors map[string]Color

func NamedColor(name string) (Color, error) {
	if color, ok := NamedColors[strings.ToLower(name)]; ok {
		return color, nil
	}
	if name != "" {
		if value, err := strconv.ParseInt(name, 16, 32); err == nil {
			return Color(value), nil
		}
	}
	return Black, errors.New("Expected name of color or numeric value in hex format.")
}

var ColorNames map[Color]string

func init() {
	NamedColors = make(map[string]Color, len(NamedColorEntries))
	ColorNames = make(map[Color]string, len(NamedColorEntries))
	for _, entry := range NamedColorEntries {
		key := strings.ToLower(entry.Name)
		NamedColors[key] = entry.Color
		if _, exists := ColorNames[entry.Color]; !exists {
			ColorNames[entry.Color] = key
		}
	}
}
