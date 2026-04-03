// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package main

import (
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/rowland/leadtype/ttf"
	"github.com/rowland/leadtype/ttf_fonts"
)

type capability struct {
	name        string
	description string
	matches     func(*ttf.FontInfo) bool
}

var capabilities = map[string]capability{
	"arabic": {
		name:        "arabic",
		description: "fonts whose OS/2 Unicode ranges indicate Arabic support",
		matches: func(fi *ttf.FontInfo) bool {
			return supportsArabic(fi.CharRanges())
		},
	},
}

type fontMatch struct {
	family    string
	fullName  string
	style     string
	filename  string
	ttcOffset int64
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr *os.File) int {
	if len(args) == 0 {
		printUsage(stderr)
		return 2
	}

	switch args[0] {
	case "capabilities", "list":
		printCapabilities(stdout)
		return 0
	}

	capabilityName := args[0]
	cap, ok := capabilities[capabilityName]
	if !ok {
		fmt.Fprintf(stderr, "Unknown capability %q.\n\n", capabilityName)
		printUsage(stderr)
		return 2
	}

	matches, err := queryCapability(cap, args[1:])
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	for _, match := range matches {
		if match.ttcOffset != 0 {
			fmt.Fprintf(stdout, "%s\t%s\t%s\t%s\t(offset=%d)\n", match.family, match.style, match.fullName, match.filename, match.ttcOffset)
			continue
		}
		fmt.Fprintf(stdout, "%s\t%s\t%s\t%s\n", match.family, match.style, match.fullName, match.filename)
	}
	return 0
}

func queryCapability(cap capability, patterns []string) ([]fontMatch, error) {
	fc, err := loadFontInfos(patterns)
	if err != nil {
		return nil, err
	}

	matches := make([]fontMatch, 0, len(fc.FontInfos))
	for _, fi := range fc.FontInfos {
		if fi == nil || !cap.matches(fi) {
			continue
		}
		matches = append(matches, fontMatch{
			family:    fi.Family(),
			fullName:  fi.FullName(),
			style:     fi.Style(),
			filename:  fi.Filename(),
			ttcOffset: fi.TTCOffset(),
		})
	}

	sort.Slice(matches, func(i, j int) bool {
		if matches[i].family != matches[j].family {
			return matches[i].family < matches[j].family
		}
		if matches[i].style != matches[j].style {
			return matches[i].style < matches[j].style
		}
		if matches[i].fullName != matches[j].fullName {
			return matches[i].fullName < matches[j].fullName
		}
		if matches[i].filename != matches[j].filename {
			return matches[i].filename < matches[j].filename
		}
		return matches[i].ttcOffset < matches[j].ttcOffset
	})
	return matches, nil
}

func loadFontInfos(patterns []string) (*ttf_fonts.TtfFonts, error) {
	if len(patterns) == 0 {
		return ttf_fonts.NewFromSystemFonts()
	}

	fc := &ttf_fonts.TtfFonts{}
	for _, pattern := range patterns {
		if err := fc.Add(pattern); err != nil {
			return nil, err
		}
	}
	return fc, nil
}

func supportsArabic(ranges *ttf.CharRanges) bool {
	if ranges == nil {
		return false
	}
	return ranges.IsSet(13) || ranges.IsSet(63) || ranges.IsSet(67)
}

func printCapabilities(w io.Writer) {
	names := make([]string, 0, len(capabilities))
	for name := range capabilities {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		cap := capabilities[name]
		fmt.Fprintf(w, "%s\t%s\n", cap.name, cap.description)
	}
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  ttquery capabilities")
	fmt.Fprintln(w, "  ttquery <capability> [font-glob ...]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "If no font globs are supplied, ttquery scans the system font directories.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Capabilities:")
	printCapabilities(w)
}
