// Copyright 2024 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ttf

import (
	"fmt"
	"io"
	"os"
)

const ttcTag = "ttcf"

type ttcHeader struct {
	version  uint32
	numFonts uint32
	offsets  []uint32
}

func readTTCHeader(file io.ReadSeeker) (*ttcHeader, error) {
	tag := make([]byte, 4)
	if _, err := io.ReadFull(file, tag); err != nil {
		return nil, err
	}
	if string(tag) != ttcTag {
		return nil, fmt.Errorf("not a TTC file: tag %q", string(tag))
	}
	var h ttcHeader
	if err := readValues(file, &h.version, &h.numFonts); err != nil {
		return nil, err
	}
	h.offsets = make([]uint32, h.numFonts)
	for i := uint32(0); i < h.numFonts; i++ {
		if err := readValues(file, &h.offsets[i]); err != nil {
			return nil, err
		}
	}
	return &h, nil
}

// LoadFontInfosFromTTC loads FontInfo for every sub-font in a TrueType Collection file.
func LoadFontInfosFromTTC(filename string) ([]*FontInfo, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	h, err := readTTCHeader(file)
	if err != nil {
		return nil, err
	}

	infos := make([]*FontInfo, 0, h.numFonts)
	for _, offset := range h.offsets {
		fi := &FontInfo{
			filename:  filename,
			ttcOffset: int64(offset),
		}
		if err := fi.init(file, int64(offset)); err != nil {
			continue
		}
		infos = append(infos, fi)
	}
	return infos, nil
}
