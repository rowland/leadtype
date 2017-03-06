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

var NamedColors = map[string]Color{
	"aliceblue":            AliceBlue,
	"antiquewhite":         AntiqueWhite,
	"aqua":                 Aqua,
	"aquamarine":           Aquamarine,
	"azure":                Azure,
	"beige":                Beige,
	"bisque":               Bisque,
	"black":                Black,
	"blanchedalmond":       BlanchedAlmond,
	"blue":                 Blue,
	"blueviolet":           BlueViolet,
	"brown":                Brown,
	"burlywood":            BurlyWood,
	"cadetblue":            CadetBlue,
	"chartreuse":           Chartreuse,
	"chocolate":            Chocolate,
	"coral":                Coral,
	"cornflowerblue":       CornflowerBlue,
	"cornsilk":             Cornsilk,
	"crimson":              Crimson,
	"cyan":                 Cyan,
	"darkblue":             DarkBlue,
	"darkcyan":             DarkCyan,
	"darkgoldenrod":        DarkGoldenRod,
	"darkgray":             DarkGray,
	"darkgrey":             DarkGrey,
	"darkgreen":            DarkGreen,
	"darkkhaki":            DarkKhaki,
	"darkmagenta":          DarkMagenta,
	"darkolivegreen":       DarkOliveGreen,
	"darkorange":           DarkOrange,
	"darkorchid":           DarkOrchid,
	"darkred":              DarkRed,
	"darksalmon":           DarkSalmon,
	"darkseagreen":         DarkSeaGreen,
	"darkslateblue":        DarkSlateBlue,
	"darkslategray":        DarkSlateGray,
	"darkslategrey":        DarkSlateGrey,
	"darkturquoise":        DarkTurquoise,
	"darkviolet":           DarkViolet,
	"deeppink":             DeepPink,
	"deepskyblue":          DeepSkyBlue,
	"dimgray":              DimGray,
	"dimgrey":              DimGrey,
	"dodgerblue":           DodgerBlue,
	"firebrick":            FireBrick,
	"floralwhite":          FloralWhite,
	"forestgreen":          ForestGreen,
	"fuchsia":              Fuchsia,
	"gainsboro":            Gainsboro,
	"ghostwhite":           GhostWhite,
	"gold":                 Gold,
	"goldenrod":            GoldenRod,
	"gray":                 Gray,
	"grey":                 Grey,
	"green":                Green,
	"greenyellow":          GreenYellow,
	"honeydew":             HoneyDew,
	"hotpink":              HotPink,
	"indianred":            IndianRed,
	"indigo":               Indigo,
	"ivory":                Ivory,
	"khaki":                Khaki,
	"lavender":             Lavender,
	"lavenderblush":        LavenderBlush,
	"lawngreen":            LawnGreen,
	"lemonchiffon":         LemonChiffon,
	"lightblue":            LightBlue,
	"lightcoral":           LightCoral,
	"lightcyan":            LightCyan,
	"lightgoldenrodyellow": LightGoldenRodYellow,
	"lightgray":            LightGray,
	"lightgrey":            LightGrey,
	"lightgreen":           LightGreen,
	"lightpink":            LightPink,
	"lightsalmon":          LightSalmon,
	"lightseagreen":        LightSeaGreen,
	"lightskyblue":         LightSkyBlue,
	"lightslategray":       LightSlateGray,
	"lightslategrey":       LightSlateGrey,
	"lightsteelblue":       LightSteelBlue,
	"lightyellow":          LightYellow,
	"lime":                 Lime,
	"limegreen":            LimeGreen,
	"linen":                Linen,
	"magenta":              Magenta,
	"maroon":               Maroon,
	"mediumaquamarine":     MediumAquaMarine,
	"mediumblue":           MediumBlue,
	"mediumorchid":         MediumOrchid,
	"mediumpurple":         MediumPurple,
	"mediumseagreen":       MediumSeaGreen,
	"mediumslateblue":      MediumSlateBlue,
	"mediumspringgreen":    MediumSpringGreen,
	"mediumturquoise":      MediumTurquoise,
	"mediumvioletred":      MediumVioletRed,
	"midnightblue":         MidnightBlue,
	"mintcream":            MintCream,
	"mistyrose":            MistyRose,
	"moccasin":             Moccasin,
	"navajowhite":          NavajoWhite,
	"navy":                 Navy,
	"oldlace":              OldLace,
	"olive":                Olive,
	"olivedrab":            OliveDrab,
	"orange":               Orange,
	"orangered":            OrangeRed,
	"orchid":               Orchid,
	"palegoldenrod":        PaleGoldenRod,
	"palegreen":            PaleGreen,
	"paleturquoise":        PaleTurquoise,
	"palevioletred":        PaleVioletRed,
	"papayawhip":           PapayaWhip,
	"peachpuff":            PeachPuff,
	"peru":                 Peru,
	"pink":                 Pink,
	"plum":                 Plum,
	"powderblue":           PowderBlue,
	"purple":               Purple,
	"red":                  Red,
	"rosybrown":            RosyBrown,
	"royalblue":            RoyalBlue,
	"saddlebrown":          SaddleBrown,
	"salmon":               Salmon,
	"sandybrown":           SandyBrown,
	"seagreen":             SeaGreen,
	"seashell":             SeaShell,
	"sienna":               Sienna,
	"silver":               Silver,
	"skyblue":              SkyBlue,
	"slateblue":            SlateBlue,
	"slategray":            SlateGray,
	"slategrey":            SlateGrey,
	"snow":                 Snow,
	"springgreen":          SpringGreen,
	"steelblue":            SteelBlue,
	"tan":                  Tan,
	"teal":                 Teal,
	"thistle":              Thistle,
	"tomato":               Tomato,
	"turquoise":            Turquoise,
	"violet":               Violet,
	"wheat":                Wheat,
	"white":                White,
	"whitesmoke":           WhiteSmoke,
	"yellow":               Yellow,
	"yellowgreen":          YellowGreen,
}

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
	ColorNames = make(map[Color]string, len(NamedColors))
	for s, c := range NamedColors {
		ColorNames[c] = s
	}
}
