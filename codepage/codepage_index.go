// Copyright 2011-2014 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package codepage

type CodepageIndex int

const (
	Idx_ISO_8859_1 = CodepageIndex(iota)
	Idx_ISO_8859_2
	Idx_ISO_8859_3
	Idx_ISO_8859_4
	Idx_ISO_8859_5
	Idx_ISO_8859_6
	Idx_ISO_8859_7
	Idx_ISO_8859_8
	Idx_ISO_8859_9
	Idx_ISO_8859_10
	Idx_ISO_8859_11
	Idx_ISO_8859_13
	Idx_ISO_8859_14
	Idx_ISO_8859_15
	Idx_ISO_8859_16
	Idx_CP1252
	Idx_CP1250
	Idx_CP1251
	Idx_CP1253
	Idx_CP1254
	Idx_CP1256
	Idx_CP1257
	Idx_CP1258
	Idx_CP874
)

var codepointCodepages = CodepageRanges{
	{0x0000, 0x00FF, 256, Idx_ISO_8859_1},
	{0x0100, 0x0101, 2, Idx_ISO_8859_4},
	{0x0102, 0x0107, 6, Idx_ISO_8859_2},
	{0x0108, 0x010B, 4, Idx_ISO_8859_3},
	{0x010C, 0x0111, 6, Idx_ISO_8859_2},
	{0x0112, 0x0115, 4, Idx_ISO_8859_4},
	{0x0118, 0x011B, 4, Idx_ISO_8859_2},
	{0x011C, 0x0121, 6, Idx_ISO_8859_3},
	{0x0122, 0x0123, 2, Idx_ISO_8859_4},
	{0x0124, 0x0127, 4, Idx_ISO_8859_3},
	{0x0128, 0x012D, 6, Idx_ISO_8859_4},
	{0x0130, 0x0133, 4, Idx_ISO_8859_3},
	{0x0136, 0x0138, 3, Idx_ISO_8859_4},
	{0x0139, 0x013A, 2, Idx_ISO_8859_2},
	{0x013B, 0x013C, 2, Idx_ISO_8859_4},
	{0x013D, 0x0142, 6, Idx_ISO_8859_2},
	{0x0145, 0x0146, 2, Idx_ISO_8859_4},
	{0x0147, 0x0148, 2, Idx_ISO_8859_2},
	{0x014A, 0x014D, 4, Idx_ISO_8859_4},
	{0x0150, 0x0151, 2, Idx_ISO_8859_2},
	{0x0152, 0x0153, 2, Idx_ISO_8859_15},
	{0x0154, 0x0155, 2, Idx_ISO_8859_2},
	{0x0156, 0x0157, 2, Idx_ISO_8859_4},
	{0x0158, 0x015B, 4, Idx_ISO_8859_2},
	{0x015C, 0x015D, 2, Idx_ISO_8859_3},
	{0x015E, 0x0165, 8, Idx_ISO_8859_2},
	{0x0166, 0x016B, 6, Idx_ISO_8859_4},
	{0x016C, 0x016D, 2, Idx_ISO_8859_3},
	{0x016E, 0x0171, 4, Idx_ISO_8859_2},
	{0x0172, 0x0173, 2, Idx_ISO_8859_4},
	{0x0174, 0x0178, 5, Idx_ISO_8859_14},
	{0x0179, 0x017E, 6, Idx_ISO_8859_2},
	{0x0192, 0x0192, 1, Idx_CP1252},
	{0x01A0, 0x01A2, 3, Idx_CP1258},
	{0x0218, 0x021B, 4, Idx_ISO_8859_16},
	{0x02C6, 0x02C6, 1, Idx_CP1252},
	{0x02C7, 0x02CA, 4, Idx_ISO_8859_2},
	{0x02DC, 0x02DC, 1, Idx_CP1252},
	{0x02DD, 0x02DD, 1, Idx_ISO_8859_2},
	{0x0300, 0x0303, 4, Idx_CP1258},
	{0x037A, 0x03C1, 72, Idx_ISO_8859_7},
	{0x0401, 0x045C, 92, Idx_ISO_8859_5},
	{0x0490, 0x0491, 2, Idx_CP1251},
	{0x05D0, 0x05EA, 27, Idx_ISO_8859_8},
	{0x060C, 0x063B, 48, Idx_ISO_8859_6},
	{0x0679, 0x0684, 12, Idx_CP1256},
	{0x0E01, 0x0E57, 87, Idx_ISO_8859_11},
	{0x1E02, 0x1E15, 20, Idx_ISO_8859_14},
	{0x1EEE, 0x1EEE, 1, Idx_CP1258},
	{0x1EF2, 0x1EF3, 2, Idx_ISO_8859_14},
	{0x200C, 0x200D, 2, Idx_CP1256},
	{0x200E, 0x200F, 2, Idx_ISO_8859_8},
	{0x2013, 0x2014, 2, Idx_CP1252},
	{0x2015, 0x2015, 1, Idx_ISO_8859_7},
	{0x2017, 0x2017, 1, Idx_ISO_8859_8},
	{0x2018, 0x2019, 2, Idx_ISO_8859_7},
	{0x201A, 0x201A, 1, Idx_CP1252},
	{0x201C, 0x201E, 3, Idx_ISO_8859_13},
	{0x2020, 0x2026, 7, Idx_CP1252},
	{0x20AC, 0x20AD, 2, Idx_ISO_8859_7},
	{0x2116, 0x2116, 1, Idx_ISO_8859_5},
	{0x2122, 0x2122, 1, Idx_CP1252},
}

func ForCodepoint(rune rune) (codepage Codepage, found bool) {
	cpi, found := codepointCodepages.CodepageIndexForCodepoint(rune)
	if found {
		codepage = cpi.Codepage()
	}
	return
}

// IndexForCodepoint returns the index into Codepages for which a codepage containing the rune can be found.
// Returns -1 if no codepage can be found containing the rune.
func IndexForCodepoint(rune rune) (index CodepageIndex, found bool) {
	return codepointCodepages.CodepageIndexForCodepoint(rune)
}

var Codepages = []Codepage{
	ISO_8859_1,
	ISO_8859_2,
	ISO_8859_3,
	ISO_8859_4,
	ISO_8859_5,
	ISO_8859_6,
	ISO_8859_7,
	ISO_8859_8,
	ISO_8859_9,
	ISO_8859_10,
	ISO_8859_11,
	ISO_8859_13,
	ISO_8859_14,
	ISO_8859_15,
	ISO_8859_16,
	CP1252,
	CP1250,
	CP1251,
	CP1253,
	CP1254,
	CP1256,
	CP1257,
	CP1258,
	CP874,
}

var CodepageMaps = [][]rune{
	ISO_8859_1_Map,
	ISO_8859_2_Map,
	ISO_8859_3_Map,
	ISO_8859_4_Map,
	ISO_8859_5_Map,
	ISO_8859_6_Map,
	ISO_8859_7_Map,
	ISO_8859_8_Map,
	ISO_8859_9_Map,
	ISO_8859_10_Map,
	ISO_8859_11_Map,
	ISO_8859_13_Map,
	ISO_8859_14_Map,
	ISO_8859_15_Map,
	ISO_8859_16_Map,
	CP1252_Map,
	CP1250_Map,
	CP1251_Map,
	CP1253_Map,
	CP1254_Map,
	CP1256_Map,
	CP1257_Map,
	CP1258_Map,
	CP874_Map,
}

var codepageNames = []string{
	"ISO-8859-1",
	"ISO-8859-2",
	"ISO-8859-3",
	"ISO-8859-4",
	"ISO-8859-5",
	"ISO-8859-6",
	"ISO-8859-7",
	"ISO-8859-8",
	"ISO-8859-9",
	"ISO-8859-10",
	"ISO-8859-11",
	"ISO-8859-13",
	"ISO-8859-14",
	"ISO-8859-15",
	"ISO-8859-16",
	"CP1252",
	"CP1250",
	"CP1251",
	"CP1253",
	"CP1254",
	"CP1256",
	"CP1257",
	"CP1258",
	"CP874",
}

func (idx CodepageIndex) Codepage() Codepage {
	return Codepages[idx]
}

// Map returns a slice of runes for a codepage mapping chars 0-255 to Unicode runes.
// A map for codepage ISO-8859-1 is returned if the CodepageIndex is invalid.
func (idx CodepageIndex) Map() []rune {
	if idx < 0 {
		return CodepageMaps[0]
	}
	return CodepageMaps[idx]
}

func (idx CodepageIndex) String() string {
	if idx < 0 {
		return "Unicode"
	}
	return codepageNames[idx]
}
