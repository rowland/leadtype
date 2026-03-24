// Copyright 2016, 2017 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

//go:build haru

package ltml

import (
	"os"
	"os/exec"

	"testing"

	"github.com/rowland/leadtype/ltml/harupdf"
	"github.com/rowland/leadtype/ttf_fonts"
)

func writeSampleHaru(name string, t *testing.T) {
	const suffix = "-haru"
	doc, err := ParseFile(sampleFile(name + ".ltml"))
	if err != nil {
		t.Fatal(err)
	}

	f, err := os.Create(sampleFile(name + suffix + ".pdf"))
	if err != nil {
		t.Fatal(err)
	}

	w := harupdf.NewDocWriter()
	ttFonts, err := ttf_fonts.New("/Library/Fonts/*.ttf")
	if err != nil {
		panic(err)
	}
	w.AddFontSource(ttFonts)

	if err := doc.Print(w); err != nil {
		t.Errorf("Printing sample: %v", err)
	}

	w.WriteTo(f)
	f.Close()
	exec.Command("open", sampleFile(name+suffix+".pdf")).Start()
}

func TestSample030Haru(t *testing.T) {
	writeSampleHaru("test_030_encodings", t)
}
