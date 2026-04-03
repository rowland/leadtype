// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/rowland/leadtype/afm_fonts"
	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/pdf"
	"github.com/rowland/leadtype/shaping"
	"github.com/rowland/leadtype/ttf"
	"github.com/rowland/leadtype/ttf_fonts"
)

func init() {
	registerSample("test_018_arabic_fonts", "list Arabic-supporting system fonts with shaping tables and an Arabic specimen", runTest018ArabicFonts)
}

func runTest018ArabicFonts() (string, error) {
	return writeDoc("test_018_arabic_fonts.pdf", func(doc *pdf.DocWriter) error {
		afm, err := afm_fonts.Default()
		if err != nil {
			return err
		}
		doc.AddFontSource(afm)

		ttfc, err := ttf_fonts.NewFromSystemFonts()
		if err != nil {
			return err
		}
		ttfc.SetShaper(shaping.NewShaper())
		doc.AddFontSource(ttfc)

		doc.SetUnits("in")
		doc.SetLineSpacing(1.35)
		doc.MoveTo(0.75, 0.75)

		const specimen = "مثال على نص عربي  مرحبا بكم"
		infos := arabicFontInfos(ttfc.FontInfos)
		for _, info := range infos {
			if doc.Y() > 10 {
				doc.NewPage()
				doc.MoveTo(0.75, 0.75)
			}

			label := info.FullName()
			if label == "" {
				label = info.Family()
			}

			if _, err := doc.SetFont("Helvetica", 11, options.Options{}); err != nil {
				return err
			}
			doc.MoveTo(0.75, doc.Y())
			fmt.Fprintln(doc, label)

			fonts, err := doc.SetFont(info.Family(), 14, options.Options{"style": info.Style()})
			if err != nil {
				fmt.Fprintf(os.Stderr, "Setting font to %s/%s: %v\n", info.Family(), info.Style(), err)
				continue
			}
			if !canRender(specimen, fonts...) {
				continue
			}

			doc.MoveTo(3.6, doc.Y()-0.19)
			fmt.Fprintln(doc, specimen)
		}
		return nil
	})
}

func arabicFontInfos(infos []*ttf.FontInfo) []*ttf.FontInfo {
	filtered := make([]*ttf.FontInfo, 0, len(infos))
	for _, info := range infos {
		if info == nil || info.Family() == "" {
			continue
		}
		if strings.HasPrefix(info.Family(), ".") {
			continue
		}
		if !supportsArabicCharRanges(info.CharRanges()) || !info.HasOpenTypeShaping() {
			continue
		}
		filtered = append(filtered, info)
	}
	sort.Slice(filtered, func(i, j int) bool {
		if filtered[i].Family() != filtered[j].Family() {
			return filtered[i].Family() < filtered[j].Family()
		}
		if filtered[i].Style() != filtered[j].Style() {
			return filtered[i].Style() < filtered[j].Style()
		}
		return filtered[i].FullName() < filtered[j].FullName()
	})
	return filtered
}

func supportsArabicCharRanges(ranges *ttf.CharRanges) bool {
	if ranges == nil {
		return false
	}
	return ranges.IsSet(13) || ranges.IsSet(63) || ranges.IsSet(67)
}
