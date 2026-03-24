//go:build ignore

// generate.go reads the *_map.go files (byte-to-rune mappings) and regenerates
// the corresponding *.go Codepage range table files.
//
// Run with: go run generate.go
// Or via:  go generate ./codepage/...

package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type charRange struct {
	firstCode rune
	lastCode  rune
	count     int
	delta     int
}

var runeLineRE = regexp.MustCompile(`^\s*0x([0-9A-Fa-f]{4}),\s*//\s*(\d+)`)

func readMapFile(filename string) ([]rune, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	result := make([]rune, 256)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		m := runeLineRE.FindStringSubmatch(scanner.Text())
		if m == nil {
			continue
		}
		idx, _ := strconv.Atoi(m[2])
		val, _ := strconv.ParseInt(m[1], 16, 32)
		if idx < 256 {
			result[idx] = rune(val)
		}
	}
	return result, scanner.Err()
}

func computeRanges(codepoints []rune) []charRange {
	type entry struct {
		r   rune
		idx int
	}

	var entries []entry
	for i, r := range codepoints {
		// byte 0 = NUL (0x0000) is always valid; skip undefined entries for bytes 1-255
		if r == 0 && i != 0 {
			continue
		}
		entries = append(entries, entry{r, i})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].r < entries[j].r
	})

	var ranges []charRange
	for _, e := range entries {
		delta := e.idx - int(e.r)
		if len(ranges) > 0 {
			cur := &ranges[len(ranges)-1]
			if e.r == cur.lastCode+1 && delta == cur.delta {
				cur.lastCode = e.r
				cur.count++
				continue
			}
		}
		ranges = append(ranges, charRange{e.r, e.r, 1, delta})
	}
	return ranges
}

func goVarName(cp string) string {
	return strings.ReplaceAll(cp, "-", "_")
}

func mapFilename(cp string) string {
	return strings.ToLower(cp) + "_map.go"
}

func tableFilename(cp string) string {
	return strings.ToLower(cp) + ".go"
}

func generateTable(cp string, codepoints []rune) error {
	ranges := computeRanges(codepoints)
	filename := tableFilename(cp)
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	fmt.Fprintln(w, "// Copyright 2011-2012 Brent Rowland.")
	fmt.Fprintln(w, "// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "package codepage")
	fmt.Fprintln(w)
	fmt.Fprintf(w, "var %s = Codepage{\n", goVarName(cp))
	for _, r := range ranges {
		fmt.Fprintf(w, "\t{0x%04X, 0x%04X, %d, %d},\n", r.firstCode, r.lastCode, r.count, r.delta)
	}
	fmt.Fprintln(w, "}")
	return w.Flush()
}

func main() {
	codepages := []string{
		"ISO-8859-1", "ISO-8859-2", "ISO-8859-3", "ISO-8859-4",
		"ISO-8859-5", "ISO-8859-6", "ISO-8859-7", "ISO-8859-8",
		"ISO-8859-9", "ISO-8859-10", "ISO-8859-11", "ISO-8859-13",
		"ISO-8859-14", "ISO-8859-15", "ISO-8859-16",
		"CP1252", "CP1250", "CP1251", "CP1253", "CP1254",
		"CP1256", "CP1257", "CP1258", "CP874",
	}

	for _, cp := range codepages {
		mapFile := mapFilename(cp)
		codepoints, err := readMapFile(mapFile)
		if err != nil {
			log.Fatalf("reading %s: %v", mapFile, err)
		}
		if err := generateTable(cp, codepoints); err != nil {
			log.Fatalf("generating table for %s: %v", cp, err)
		}
		fmt.Printf("Generated %s\n", tableFilename(cp))
	}
}
