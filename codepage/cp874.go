// Copyright 2011-2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package codepage

var CP874 = Codepage{
	{0x0000, 0x007F, 128, 0},
	{0x00A0, 0x00A0, 1, 0},
	{0x0E01, 0x0E3A, 58, -3424},
	{0x0E3F, 0x0E5B, 29, -3424},
	{0x2013, 0x2014, 2, -8061},
	{0x2018, 0x2019, 2, -8071},
	{0x201C, 0x201D, 2, -8073},
	{0x2022, 0x2022, 1, -8077},
	{0x2026, 0x2026, 1, -8097},
	{0x20AC, 0x20AC, 1, -8236},
}
