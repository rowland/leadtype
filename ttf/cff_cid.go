// Copyright 2026 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ttf

import (
	"fmt"
)

const cffStandardStringCount = 391

// cffCIDMetadata is the small, immutable portion of a CID-keyed CFF program
// needed outside the subsetter. glyphCIDs is indexed by source GID: unlike a
// name-keyed CFF font, a CID-keyed charset stores a CID for every glyph after
// .notdef. fdSelect identifies the Font DICT supplying that glyph's private
// dictionary, local subroutines, and possibly FontMatrix.
type cffCIDMetadata struct {
	registry   string
	ordering   string
	supplement int
	glyphCIDs  []uint16
	fdSelect   []uint8
}

// parseCFFCIDMetadata recognizes CID-keyed CFF1 data and extracts the ROS,
// GID-to-CID charset, and GID-to-FD assignment. A valid name-keyed CFF program
// returns (nil, nil); malformed or unsupported CFF returns an error so callers
// do not quietly treat a broken CID font as an ordinary name-keyed font.
func parseCFFCIDMetadata(data []byte, numGlyphs int) (*cffCIDMetadata, error) {
	if len(data) < 4 || data[0] != 1 || data[1] != 0 {
		return nil, fmt.Errorf("cff: unsupported header")
	}
	off := int(data[2])
	_, next, err := parseCFFIndex(data, off)
	if err != nil {
		return nil, err
	}
	topDicts, next, err := parseCFFIndex(data, next)
	if err != nil {
		return nil, err
	}
	if len(topDicts) != 1 {
		return nil, fmt.Errorf("cff: expected one top dict, got %d", len(topDicts))
	}
	stringsIndex, _, err := parseCFFIndex(data, next)
	if err != nil {
		return nil, err
	}
	top, err := parseCFFTopDict(topDicts[0])
	if err != nil {
		return nil, err
	}
	if !top.isCID {
		return nil, nil
	}
	registry, err := resolveCFFSID(top.rosRegistrySID, stringsIndex)
	if err != nil {
		return nil, err
	}
	ordering, err := resolveCFFSID(top.rosOrderingSID, stringsIndex)
	if err != nil {
		return nil, err
	}
	cids, err := parseCFFCharset(data, top.charsetOffset, numGlyphs)
	if err != nil {
		return nil, err
	}
	fds, err := parseCFFFDSelect(data, top.fdSelectOffset, numGlyphs)
	if err != nil {
		return nil, err
	}
	return &cffCIDMetadata{
		registry:   registry,
		ordering:   ordering,
		supplement: top.rosSupplement,
		glyphCIDs:  cids,
		fdSelect:   fds,
	}, nil
}

// resolveCFFSID resolves an SID used by ROS. Registry and Ordering normally
// live in the font's custom String INDEX. Standard SIDs are rejected here
// because none of the standard CFF strings are valid ROS names in our bundled
// fonts, and accepting one without a complete standard-string table would
// produce a misleading PDF CIDSystemInfo dictionary.
func resolveCFFSID(sid int, stringsIndex [][]byte) (string, error) {
	index := sid - cffStandardStringCount
	if index < 0 {
		return "", fmt.Errorf("cff: ROS uses unsupported standard SID %d", sid)
	}
	if index >= len(stringsIndex) {
		return "", fmt.Errorf("cff: ROS SID %d is outside String INDEX", sid)
	}
	return string(stringsIndex[index]), nil
}

// parseCFFFDSelect expands the compact on-disk FDSelect into one Font DICT
// index per source GID. Subsetting works on this expanded form, then emits a
// canonical format 3 table after unused Font DICTs have been removed.
func parseCFFFDSelect(data []byte, off, numGlyphs int) ([]uint8, error) {
	if off <= 0 || off >= len(data) {
		return nil, fmt.Errorf("cff: invalid FDSelect offset")
	}
	fds := make([]uint8, numGlyphs)
	switch format := data[off]; format {
	case 0:
		if off+1+numGlyphs > len(data) {
			return nil, fmt.Errorf("cff: truncated FDSelect format 0")
		}
		copy(fds, data[off+1:off+1+numGlyphs])
	case 3:
		if off+3 > len(data) {
			return nil, fmt.Errorf("cff: truncated FDSelect format 3")
		}
		nRanges := int(data[off+1])<<8 | int(data[off+2])
		pos := off + 3
		if pos+nRanges*3+2 > len(data) {
			return nil, fmt.Errorf("cff: truncated FDSelect ranges")
		}
		firsts := make([]int, nRanges+1)
		values := make([]uint8, nRanges)
		for i := 0; i < nRanges; i++ {
			firsts[i] = int(data[pos])<<8 | int(data[pos+1])
			values[i] = data[pos+2]
			pos += 3
		}
		firsts[nRanges] = int(data[pos])<<8 | int(data[pos+1])
		if nRanges == 0 || firsts[0] != 0 || firsts[nRanges] != numGlyphs {
			return nil, fmt.Errorf("cff: invalid FDSelect range coverage")
		}
		for i := 0; i < nRanges; i++ {
			if firsts[i] > firsts[i+1] {
				return nil, fmt.Errorf("cff: unordered FDSelect ranges")
			}
			for gid := firsts[i]; gid < firsts[i+1]; gid++ {
				fds[gid] = values[i]
			}
		}
	default:
		return nil, fmt.Errorf("cff: unsupported FDSelect format %d", format)
	}
	return fds, nil
}
