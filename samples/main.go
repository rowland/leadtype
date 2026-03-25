// Copyright 2026 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
)

type sampleFunc func() (string, error)

type sample struct {
	name        string
	description string
	openArgs    []string
	run         sampleFunc
}

var registry = map[string]sample{}

func registerSample(name, description string, run sampleFunc, openArgs ...string) {
	if _, exists := registry[name]; exists {
		panic("duplicate sample registration: " + name)
	}
	registry[name] = sample{
		name:        name,
		description: description,
		openArgs:    append([]string(nil), openArgs...),
		run:         run,
	}
}

func main() {
	var openAfter bool
	var listOnly bool

	flag.BoolVar(&openAfter, "o", false, "open the generated PDF after writing it")
	flag.BoolVar(&openAfter, "open", false, "open the generated PDF after writing it")
	flag.BoolVar(&listOnly, "list", false, "list available samples")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: go run ./samples [flags] <sample>\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(flag.CommandLine.Output(), "\nAvailable samples:\n")
		for _, s := range sortedSamples() {
			fmt.Fprintf(flag.CommandLine.Output(), "  %s\t%s\n", s.name, s.description)
		}
	}
	flag.Parse()

	if listOnly {
		flag.Usage()
		return
	}
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}

	name := flag.Arg(0)
	s, ok := registry[name]
	if !ok {
		fmt.Fprintf(os.Stderr, "unknown sample %q\n\n", name)
		flag.Usage()
		os.Exit(2)
	}

	outputPath, err := s.run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("Wrote %s\n", outputPath)

	if openAfter {
		if err := openFile(outputPath, s.openArgs...); err != nil {
			fmt.Fprintf(os.Stderr, "open %s: %v\n", outputPath, err)
			os.Exit(1)
		}
	}
}

func sortedSamples() []sample {
	samples := make([]sample, 0, len(registry))
	for _, s := range registry {
		samples = append(samples, s)
	}
	sort.Slice(samples, func(i, j int) bool { return samples[i].name < samples[j].name })
	return samples
}
