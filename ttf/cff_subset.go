// Copyright 2026 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ttf

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/rowland/leadtype/internal/pdfsubset"
)

const (
	cffOpCharset     = 15
	cffOpEncoding    = 16
	cffOpCharStrings = 17
	cffOpPrivate     = 18
	cffOpSubrs       = 19
)

var cffEndChar = []byte{14}
var cffReturn = []byte{11}

// cffSubsetData is the parsed CFF1 program in the form needed by both
// name-keyed and CID-keyed subsetters. INDEX objects are represented as slices
// of their object bytes so unused CharStrings and subroutines can be replaced
// or pruned without interpreting drawing operators.
type cffSubsetData struct {
	name       []byte
	topDict    []byte
	strings    [][]byte
	globalSubr [][]byte
	charsets   []uint16
	charstring [][]byte
	private    []byte
	localSubr  [][]byte
	isCID      bool
	fdSelect   []uint8
	fontDicts  []cffFontDictData
}

// cffFontDictData holds one entry from a CID font's FDArray together with the
// Private DICT and local Subrs reached from that entry. Local subroutine numbers
// are meaningful only within their owning Font DICT.
type cffFontDictData struct {
	dict      []byte
	private   []byte
	localSubr [][]byte
}

// cffTopDictInfo records offsets and CID operators from the Top DICT. The raw
// Top DICT is retained separately and rewritten after rebuilt INDEX sizes make
// all absolute offsets known.
type cffTopDictInfo struct {
	charsetOffset     int
	charStringsOffset int
	privateOffset     int
	privateLength     int
	isCID             bool
	rosRegistrySID    int
	rosOrderingSID    int
	rosSupplement     int
	cidCount          int
	fdArrayOffset     int
	fdSelectOffset    int
}

type cffPrivateInfo struct {
	subrsOffset int
}

type cffSubsetResourceKey struct {
	filename  string
	memory    *Font
	offset    uint32
	length    uint32
	numGlyphs uint16
}

type cffSubsetResource struct {
	data   []byte
	parsed *cffSubsetData
}

var readCFFSubsetTable = func(font *Font, entry *tableDirEntry) ([]byte, error) {
	if font.rawBytes != nil {
		start := uint64(entry.offset)
		end := start + uint64(entry.length)
		if end > uint64(len(font.rawBytes)) {
			return nil, fmt.Errorf("subset cff: CFF table exceeds in-memory font bounds")
		}
		return font.rawBytes[start:end], nil
	}
	file, err := os.Open(font.filename)
	if err != nil {
		return nil, fmt.Errorf("subset cff: opening %s: %w", font.filename, err)
	}
	defer file.Close()
	data := make([]byte, entry.length)
	reader := io.NewSectionReader(file, int64(entry.offset), int64(entry.length))
	if _, err := io.ReadFull(reader, data); err != nil {
		return nil, fmt.Errorf("subset cff: reading CFF table from %s: %w", font.filename, err)
	}
	return data, nil
}

func (font *Font) cffSubsetResource(session *pdfsubset.Session) (*cffSubsetResource, error) {
	entry := font.tableDir.table("CFF ")
	if entry == nil {
		return nil, fmt.Errorf("subset cff: font has no CFF table")
	}
	key := cffSubsetResourceKey{
		filename:  font.filename,
		offset:    entry.offset,
		length:    entry.length,
		numGlyphs: font.maxpTable.numGlyphs,
	}
	if font.rawBytes != nil {
		key.filename = ""
		key.memory = font
	}
	value, err := session.Load(key, func() (any, error) {
		data, err := readCFFSubsetTable(font, entry)
		if err != nil {
			return nil, err
		}
		parsed, err := parseCFFForSubset(data, int(font.maxpTable.numGlyphs))
		if err != nil {
			return nil, err
		}
		return &cffSubsetResource{data: data, parsed: parsed}, nil
	})
	if err != nil {
		return nil, err
	}
	resource, ok := value.(*cffSubsetResource)
	if !ok {
		return nil, fmt.Errorf("subset cff: invalid cached CFF resource %T", value)
	}
	return resource, nil
}

// PDFSubsetWithSession returns the exact font program required by PDF while
// sharing immutable CFF parsing within session. CID-keyed CFF fonts return raw
// CFF with CIDFontType0C; all other outlines preserve the existing subset path.
func (font *Font) PDFSubsetWithSession(session *pdfsubset.Session, glyphIDs []uint16) ([]byte, string, error) {
	if font.cffCID == nil {
		data, err := font.Subset(glyphIDs)
		return data, "", err
	}
	resource, err := font.cffSubsetResource(session)
	if err != nil {
		return nil, "", err
	}
	if !resource.parsed.isCID {
		return nil, "", fmt.Errorf("subset cff: loaded CID-keyed font has a name-keyed CFF table")
	}
	subset, err := buildCIDKeyedCFFSubset(resource.parsed, glyphIDs)
	if err != nil {
		return nil, "", err
	}
	return subset.cffData, "CIDFontType0C", nil
}

func (font *Font) subsetCFFOpenType(glyphIDs []uint16) ([]byte, error) {
	raw, err := os.ReadFile(font.filename)
	if err != nil {
		return nil, fmt.Errorf("subset cff: reading %s: %w", font.filename, err)
	}
	cffEntry := font.tableDir.table("CFF ")
	if cffEntry == nil {
		return nil, fmt.Errorf("subset cff: font has no CFF table")
	}
	cffRaw := make([]byte, cffEntry.length)
	copy(cffRaw, raw[cffEntry.offset:cffEntry.offset+cffEntry.length])

	parsed, err := parseCFFForSubset(cffRaw, int(font.maxpTable.numGlyphs))
	if err != nil {
		return nil, err
	}
	if parsed.isCID {
		return font.subsetCIDKeyedCFF(raw, parsed, glyphIDs)
	}

	closure := make(map[uint16]bool, len(glyphIDs)+1)
	closure[0] = true
	for _, gid := range glyphIDs {
		if int(gid) < len(parsed.charstring) {
			closure[gid] = true
		}
	}
	subsetGlyphCount := 1
	for gid := range closure {
		if int(gid)+1 > subsetGlyphCount {
			subsetGlyphCount = int(gid) + 1
		}
	}

	usedGlobals := make(map[int]bool)
	usedLocals := make(map[int]bool)
	for gid := range closure {
		if int(gid) >= len(parsed.charstring) {
			continue
		}
		collectCFFSubrs(parsed.charstring[gid], parsed.localSubr, parsed.globalSubr, usedLocals, usedGlobals)
	}

	charstrings := make([][]byte, subsetGlyphCount)
	for gid := 0; gid < subsetGlyphCount; gid++ {
		if closure[uint16(gid)] && gid < len(parsed.charstring) {
			charstrings[gid] = parsed.charstring[gid]
		} else {
			charstrings[gid] = cffEndChar
		}
	}
	globalSubrs := pruneCFFSubrs(parsed.globalSubr, usedGlobals)
	localSubrs := pruneCFFSubrs(parsed.localSubr, usedLocals)

	charsets := make([]uint16, subsetGlyphCount)
	copy(charsets, parsed.charsets)
	cffData, err := buildSubsetCFF(parsed.name, parsed.topDict, parsed.strings, globalSubrs, charstrings, charsets, parsed.private, localSubrs)
	if err != nil {
		return nil, err
	}

	maxpBuf := buildCFFMaxpTable(uint16(subsetGlyphCount))
	hheaBuf, err := subsetHheaTable(raw, font.tableDir.table("hhea"), uint16(subsetGlyphCount))
	if err != nil {
		return nil, err
	}
	hmtxBuf, err := font.subsetHmtxTable(uint16(subsetGlyphCount))
	if err != nil {
		return nil, err
	}
	cmapBuf, err := font.subsetCmapTable(closure)
	if err != nil {
		return nil, err
	}
	postBuf, err := font.subsetPostTable(raw, font.tableDir.table("post"), uint16(subsetGlyphCount))
	if err != nil {
		return nil, err
	}

	tables := make([]sfntTable, 0, len(font.tableDir.entries))
	for _, entry := range font.tableDir.entries {
		if !cffSubsetKeepTable(entry.tag) {
			continue
		}
		var data []byte
		switch entry.tag {
		case "CFF ":
			data = cffData
		case "maxp":
			data = maxpBuf
		case "hhea":
			data = hheaBuf
		case "hmtx":
			data = hmtxBuf
		case "cmap":
			data = cmapBuf
		case "post":
			data = postBuf
		default:
			data = make([]byte, entry.length)
			copy(data, raw[entry.offset:entry.offset+entry.length])
		}
		tables = append(tables, sfntTable{tag: entry.tag, data: data})
	}
	return assembleSFNT(font.scalar, tables)
}

type cidCFFSubsetResult struct {
	cffData   []byte
	closure   map[uint16]bool
	oldGlyphs []int
	oldToNew  map[uint16]uint16
}

// buildCIDKeyedCFFSubset builds a dense raw CFF subset. Retained source GIDs
// are sorted and assigned new contiguous GIDs while their original CIDs remain
// in the CFF charset. The parsed source is immutable and may be reused by
// multiple subsets.
func buildCIDKeyedCFFSubset(parsed *cffSubsetData, glyphIDs []uint16) (*cidCFFSubsetResult, error) {
	// Always retain .notdef, then establish the one source-GID to subset-GID
	// map shared by CharStrings, cmap, metrics, and maxp.
	closure := map[uint16]bool{0: true}
	for _, gid := range glyphIDs {
		if int(gid) >= len(parsed.charstring) {
			return nil, fmt.Errorf("subset cff: glyph %d is outside CharStrings INDEX", gid)
		}
		closure[gid] = true
	}
	oldGlyphs := make([]int, 0, len(closure))
	for gid := range closure {
		oldGlyphs = append(oldGlyphs, int(gid))
	}
	sort.Ints(oldGlyphs)
	oldToNew := make(map[uint16]uint16, len(oldGlyphs))
	charstrings := make([][]byte, len(oldGlyphs))
	charsets := make([]uint16, len(oldGlyphs))
	oldFDs := make([]uint8, len(oldGlyphs))
	for newGID, oldGID := range oldGlyphs {
		oldToNew[uint16(oldGID)] = uint16(newGID)
		charstrings[newGID] = parsed.charstring[oldGID]
		charsets[newGID] = parsed.charsets[oldGID]
		oldFDs[newGID] = parsed.fdSelect[oldGID]
	}

	// FDArray indexes are internal to this CFF program, so unused entries may
	// be removed as long as every retained glyph's FDSelect value is remapped.
	usedFDSet := make(map[uint8]bool)
	for _, fd := range oldFDs {
		usedFDSet[fd] = true
	}
	usedFDs := make([]int, 0, len(usedFDSet))
	for fd := range usedFDSet {
		usedFDs = append(usedFDs, int(fd))
	}
	sort.Ints(usedFDs)
	fdRemap := make(map[uint8]uint8, len(usedFDs))
	fontDicts := make([]cffFontDictData, len(usedFDs))
	usedLocals := make([]map[int]bool, len(usedFDs))
	for newFD, oldFD := range usedFDs {
		fdRemap[uint8(oldFD)] = uint8(newFD)
		fontDicts[newFD] = parsed.fontDicts[oldFD]
		usedLocals[newFD] = make(map[int]bool)
	}
	fdSelect := make([]uint8, len(oldFDs))
	for gid, oldFD := range oldFDs {
		fdSelect[gid] = fdRemap[oldFD]
	}

	// A global subroutine executes in the local-subroutine environment of its
	// caller's FD. Visit it once per (FD, global-index), not merely once per
	// global index, or local dependencies from later FDs can be missed.
	usedGlobals := make(map[int]bool)
	visitedGlobals := make(map[cffGlobalContext]bool)
	for newGID, oldGID := range oldGlyphs {
		fd := int(fdSelect[newGID])
		collectCFFCIDSubrs(parsed.charstring[oldGID], fd, fontDicts, parsed.globalSubr, usedLocals, usedGlobals, visitedGlobals, 0)
	}
	globalSubrs := pruneCFFSubrs(parsed.globalSubr, usedGlobals)
	for fd := range fontDicts {
		fontDicts[fd].localSubr = pruneCFFSubrs(fontDicts[fd].localSubr, usedLocals[fd])
	}

	cffData, err := buildSubsetCIDCFF(parsed.name, parsed.topDict, parsed.strings, globalSubrs, charstrings, charsets, fdSelect, fontDicts)
	if err != nil {
		return nil, err
	}
	return &cidCFFSubsetResult{
		cffData:   cffData,
		closure:   closure,
		oldGlyphs: oldGlyphs,
		oldToNew:  oldToNew,
	}, nil
}

// subsetCIDKeyedCFF wraps the dense CFF program in a standalone OpenType font.
// PDF embedding uses the raw program directly and deliberately skips this
// wrapper work.
func (font *Font) subsetCIDKeyedCFF(raw []byte, parsed *cffSubsetData, glyphIDs []uint16) ([]byte, error) {
	subset, err := buildCIDKeyedCFFSubset(parsed, glyphIDs)
	if err != nil {
		return nil, err
	}
	maxpBuf := buildCFFMaxpTable(uint16(len(subset.oldGlyphs)))
	hheaBuf, err := subsetHheaTable(raw, font.tableDir.table("hhea"), uint16(len(subset.oldGlyphs)))
	if err != nil {
		return nil, err
	}
	hmtxBuf := font.subsetHmtxTableRemapped(subset.oldGlyphs)
	cmapBuf, err := font.subsetCmapTableRemapped(subset.closure, subset.oldToNew)
	if err != nil {
		return nil, err
	}
	postBuf, err := subsetCFFPostTable(raw, font.tableDir.table("post"))
	if err != nil {
		return nil, err
	}

	tables := make([]sfntTable, 0, len(font.tableDir.entries))
	for _, entry := range font.tableDir.entries {
		if !cffSubsetKeepTable(entry.tag) {
			continue
		}
		var data []byte
		switch entry.tag {
		case "CFF ":
			data = subset.cffData
		case "maxp":
			data = maxpBuf
		case "hhea":
			data = hheaBuf
		case "hmtx":
			data = hmtxBuf
		case "cmap":
			data = cmapBuf
		case "post":
			data = postBuf
		default:
			data = append([]byte(nil), raw[entry.offset:entry.offset+entry.length]...)
		}
		tables = append(tables, sfntTable{tag: entry.tag, data: data})
	}
	return assembleSFNT(font.scalar, tables)
}

func (font *Font) subsetHmtxTableRemapped(oldGlyphs []int) []byte {
	var buf bytes.Buffer
	for _, oldGID := range oldGlyphs {
		metric := font.hmtxTable.lookup(oldGID)
		writeUint16(&buf, metric.advanceWidth)
		writeInt16(&buf, metric.leftSideBearing)
	}
	return buf.Bytes()
}

func (font *Font) subsetCmapTableRemapped(closure map[uint16]bool, oldToNew map[uint16]uint16) ([]byte, error) {
	mappings, err := font.subsetCodepointMappings(closure)
	if err != nil {
		return nil, err
	}
	for codepoint, oldGID := range mappings {
		newGID, ok := oldToNew[oldGID]
		if !ok {
			delete(mappings, codepoint)
			continue
		}
		mappings[codepoint] = newGID
	}
	for codepoint := range mappings {
		if codepoint > 0xffff {
			return buildSubsetFormat12Cmap(mappings), nil
		}
	}
	return buildSubsetFormat4Cmap(mappings), nil
}

func subsetCFFPostTable(raw []byte, entry *tableDirEntry) ([]byte, error) {
	if entry == nil || entry.length < 32 {
		return nil, fmt.Errorf("subset cff: malformed post table")
	}
	data := append([]byte(nil), raw[entry.offset:entry.offset+32]...)
	binary.BigEndian.PutUint32(data, 0x00030000)
	return data, nil
}

type cffGlobalContext struct {
	fd    int
	index int
}

// collectCFFCIDSubrs walks Type 2 callsubr/callgsubr instructions to compute
// subroutine closure for a CID font. It is not an outline interpreter; it only
// maintains enough operand and stem state to skip numbers and hint masks and
// to identify subroutine operands. The depth guard makes cyclic or malicious
// programs fail closed instead of recursing indefinitely.
func collectCFFCIDSubrs(charstring []byte, fd int, fontDicts []cffFontDictData, globalSubrs [][]byte, usedLocals []map[int]bool, usedGlobals map[int]bool, visitedGlobals map[cffGlobalContext]bool, depth int) {
	if depth > 32 || fd < 0 || fd >= len(fontDicts) {
		return
	}
	localSubrs := fontDicts[fd].localSubr
	stack := make([]int, 0, 16)
	stems := 0
	for pos := 0; pos < len(charstring); {
		b := charstring[pos]
		if n, ok := type2NumberLen(charstring[pos:]); ok {
			value, _ := parseType2Number(charstring[pos:])
			stack = append(stack, value)
			pos += n
			continue
		}
		if b == 10 || b == 29 {
			if len(stack) > 0 {
				operand := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				if b == 10 {
					idx := operand + cffSubrBias(len(localSubrs))
					if idx >= 0 && idx < len(localSubrs) && !usedLocals[fd][idx] {
						usedLocals[fd][idx] = true
						collectCFFCIDSubrs(localSubrs[idx], fd, fontDicts, globalSubrs, usedLocals, usedGlobals, visitedGlobals, depth+1)
					}
				} else {
					idx := operand + cffSubrBias(len(globalSubrs))
					ctx := cffGlobalContext{fd: fd, index: idx}
					if idx >= 0 && idx < len(globalSubrs) && !visitedGlobals[ctx] {
						visitedGlobals[ctx] = true
						usedGlobals[idx] = true
						collectCFFCIDSubrs(globalSubrs[idx], fd, fontDicts, globalSubrs, usedLocals, usedGlobals, visitedGlobals, depth+1)
					}
				}
			}
			pos++
			continue
		}
		switch b {
		case 1, 3, 18, 23:
			stems += len(stack) / 2
			stack = stack[:0]
			pos++
			continue
		case 19, 20:
			stems += len(stack) / 2
			pos++
			pos += (stems + 7) / 8
			if pos > len(charstring) {
				return
			}
			stack = stack[:0]
			continue
		}
		if b == 12 {
			pos += 2
		} else {
			pos++
		}
		stack = stack[:0]
	}
}

func cffSubsetKeepTable(tag string) bool {
	switch tag {
	case "CFF ", "cmap", "head", "hhea", "hmtx", "maxp", "name", "OS/2", "post":
		return true
	default:
		return false
	}
}

// parseCFFForSubset follows the CFF header and INDEX sequence, then resolves
// Top DICT offsets into the structures needed for rewriting. Both charset and
// FDSelect are expanded to one value per GID so dense remapping is explicit.
func parseCFFForSubset(data []byte, numGlyphs int) (*cffSubsetData, error) {
	if len(data) < 4 || data[0] != 1 || data[1] != 0 {
		return nil, fmt.Errorf("subset cff: unsupported CFF header")
	}
	off := int(data[2])
	names, next, err := parseCFFIndex(data, off)
	if err != nil {
		return nil, err
	}
	if len(names) != 1 {
		return nil, fmt.Errorf("subset cff: expected one font, got %d", len(names))
	}
	topDicts, next, err := parseCFFIndex(data, next)
	if err != nil {
		return nil, err
	}
	if len(topDicts) != 1 {
		return nil, fmt.Errorf("subset cff: expected one top dict, got %d", len(topDicts))
	}
	topInfo, err := parseCFFTopDict(topDicts[0])
	if err != nil {
		return nil, err
	}
	strings, next, err := parseCFFIndex(data, next)
	if err != nil {
		return nil, err
	}
	globalSubrs, _, err := parseCFFIndex(data, next)
	if err != nil {
		return nil, err
	}
	if topInfo.charStringsOffset <= 0 {
		return nil, fmt.Errorf("subset cff: missing CharStrings offset")
	}
	charstrings, _, err := parseCFFIndex(data, topInfo.charStringsOffset)
	if err != nil {
		return nil, err
	}
	if len(charstrings) != numGlyphs {
		return nil, fmt.Errorf("subset cff: CharStrings count %d != maxp glyphs %d", len(charstrings), numGlyphs)
	}
	charsets, err := parseCFFCharset(data, topInfo.charsetOffset, numGlyphs)
	if err != nil {
		return nil, err
	}

	var private []byte
	var localSubrs [][]byte
	var fdSelect []uint8
	var fontDicts []cffFontDictData
	if topInfo.isCID {
		fdSelect, err = parseCFFFDSelect(data, topInfo.fdSelectOffset, numGlyphs)
		if err != nil {
			return nil, err
		}
		if topInfo.fdArrayOffset <= 0 {
			return nil, fmt.Errorf("subset cff: missing FDArray offset")
		}
		fdDicts, _, err := parseCFFIndex(data, topInfo.fdArrayOffset)
		if err != nil {
			return nil, err
		}
		fontDicts = make([]cffFontDictData, len(fdDicts))
		for i, dict := range fdDicts {
			fdInfo, err := parseCFFTopDict(dict)
			if err != nil {
				return nil, err
			}
			fontDicts[i].dict = dict
			if fdInfo.privateLength > 0 {
				priv, subrs, err := parseCFFPrivateAndSubrs(data, fdInfo.privateOffset, fdInfo.privateLength)
				if err != nil {
					return nil, err
				}
				fontDicts[i].private = priv
				fontDicts[i].localSubr = subrs
			}
		}
		for gid, fd := range fdSelect {
			if int(fd) >= len(fontDicts) {
				return nil, fmt.Errorf("subset cff: glyph %d selects missing FD %d", gid, fd)
			}
		}
	} else if topInfo.privateLength > 0 {
		private, localSubrs, err = parseCFFPrivateAndSubrs(data, topInfo.privateOffset, topInfo.privateLength)
		if err != nil {
			return nil, err
		}
	}

	return &cffSubsetData{
		name:       names[0],
		topDict:    topDicts[0],
		strings:    strings,
		globalSubr: globalSubrs,
		charsets:   charsets,
		charstring: charstrings,
		private:    private,
		localSubr:  localSubrs,
		isCID:      topInfo.isCID,
		fdSelect:   fdSelect,
		fontDicts:  fontDicts,
	}, nil
}

func parseCFFPrivateAndSubrs(data []byte, offset, length int) ([]byte, [][]byte, error) {
	if offset < 0 || offset+length > len(data) {
		return nil, nil, fmt.Errorf("subset cff: invalid Private DICT bounds")
	}
	private := data[offset : offset+length]
	info, err := parseCFFPrivateDict(private)
	if err != nil {
		return nil, nil, err
	}
	var subrs [][]byte
	if info.subrsOffset > 0 {
		subrs, _, err = parseCFFIndex(data, offset+info.subrsOffset)
		if err != nil {
			return nil, nil, err
		}
	}
	return private, subrs, nil
}

func parseCFFTopDict(dict []byte) (cffTopDictInfo, error) {
	var info cffTopDictInfo
	return info, walkCFFDict(dict, func(op []byte, operands []int) {
		if len(op) == 1 {
			switch op[0] {
			case cffOpCharset:
				if len(operands) > 0 {
					info.charsetOffset = operands[len(operands)-1]
				}
			case cffOpCharStrings:
				if len(operands) > 0 {
					info.charStringsOffset = operands[len(operands)-1]
				}
			case cffOpPrivate:
				if len(operands) >= 2 {
					info.privateLength = operands[len(operands)-2]
					info.privateOffset = operands[len(operands)-1]
				}
			}
		} else if len(op) == 2 && op[0] == 12 {
			switch op[1] {
			case 30:
				info.isCID = true
				if len(operands) >= 3 {
					info.rosRegistrySID = operands[len(operands)-3]
					info.rosOrderingSID = operands[len(operands)-2]
					info.rosSupplement = operands[len(operands)-1]
				}
			case 34:
				if len(operands) > 0 {
					info.cidCount = operands[len(operands)-1]
				}
			case 36:
				if len(operands) > 0 {
					info.fdArrayOffset = operands[len(operands)-1]
				}
			case 37:
				if len(operands) > 0 {
					info.fdSelectOffset = operands[len(operands)-1]
				}
			}
		}
	})
}

func parseCFFPrivateDict(dict []byte) (cffPrivateInfo, error) {
	var info cffPrivateInfo
	return info, walkCFFDict(dict, func(op []byte, operands []int) {
		if len(op) == 1 && op[0] == cffOpSubrs && len(operands) > 0 {
			info.subrsOffset = operands[len(operands)-1]
		}
	})
}

func parseCFFIndex(data []byte, off int) ([][]byte, int, error) {
	if off < 0 || off+2 > len(data) {
		return nil, off, fmt.Errorf("subset cff: invalid INDEX offset")
	}
	count := int(binary.BigEndian.Uint16(data[off:]))
	off += 2
	if count == 0 {
		return nil, off, nil
	}
	if off >= len(data) {
		return nil, off, fmt.Errorf("subset cff: truncated INDEX")
	}
	offSize := int(data[off])
	off++
	if offSize < 1 || offSize > 4 {
		return nil, off, fmt.Errorf("subset cff: invalid INDEX offSize %d", offSize)
	}
	if off+(count+1)*offSize > len(data) {
		return nil, off, fmt.Errorf("subset cff: truncated INDEX offsets")
	}
	offsets := make([]int, count+1)
	for i := 0; i <= count; i++ {
		offsets[i] = readCFFOffset(data[off+i*offSize:off+(i+1)*offSize], offSize)
		if offsets[i] <= 0 {
			return nil, off, fmt.Errorf("subset cff: invalid INDEX object offset")
		}
	}
	objStart := off + (count+1)*offSize
	end := objStart + offsets[count] - 1
	if end > len(data) {
		return nil, off, fmt.Errorf("subset cff: INDEX data out of bounds")
	}
	items := make([][]byte, count)
	for i := 0; i < count; i++ {
		start := objStart + offsets[i] - 1
		stop := objStart + offsets[i+1] - 1
		if start > stop || stop > len(data) {
			return nil, off, fmt.Errorf("subset cff: invalid INDEX object bounds")
		}
		items[i] = data[start:stop]
	}
	return items, end, nil
}

func readCFFOffset(b []byte, size int) int {
	n := 0
	for i := 0; i < size; i++ {
		n = n<<8 | int(b[i])
	}
	return n
}

// parseCFFCharset expands formats 0, 1, and 2 into one SID or CID per GID.
// Element zero remains zero for .notdef, which is implicit in all CFF charset
// formats and therefore absent from the encoded ranges.
func parseCFFCharset(data []byte, off, numGlyphs int) ([]uint16, error) {
	charsets := make([]uint16, numGlyphs)
	if numGlyphs == 0 {
		return charsets, nil
	}
	if off <= 2 {
		return nil, fmt.Errorf("subset cff: predefined charsets are not supported")
	}
	if off < 0 || off >= len(data) {
		return nil, fmt.Errorf("subset cff: invalid charset offset")
	}
	format := data[off]
	pos := off + 1
	switch format {
	case 0:
		if pos+2*(numGlyphs-1) > len(data) {
			return nil, fmt.Errorf("subset cff: truncated charset format 0")
		}
		for gid := 1; gid < numGlyphs; gid++ {
			charsets[gid] = binary.BigEndian.Uint16(data[pos:])
			pos += 2
		}
	case 1:
		for gid := 1; gid < numGlyphs; {
			if pos+3 > len(data) {
				return nil, fmt.Errorf("subset cff: truncated charset format 1")
			}
			first := binary.BigEndian.Uint16(data[pos:])
			nLeft := int(data[pos+2])
			pos += 3
			for i := 0; i <= nLeft && gid < numGlyphs; i++ {
				charsets[gid] = first + uint16(i)
				gid++
			}
		}
	case 2:
		for gid := 1; gid < numGlyphs; {
			if pos+4 > len(data) {
				return nil, fmt.Errorf("subset cff: truncated charset format 2")
			}
			first := binary.BigEndian.Uint16(data[pos:])
			nLeft := int(binary.BigEndian.Uint16(data[pos+2:]))
			pos += 4
			for i := 0; i <= nLeft && gid < numGlyphs; i++ {
				charsets[gid] = first + uint16(i)
				gid++
			}
		}
	default:
		return nil, fmt.Errorf("subset cff: unsupported charset format %d", format)
	}
	return charsets, nil
}

// walkCFFDict tokenizes DICT data without assigning semantics to most
// operators. Offset parsers and rewriters share it so escaped operators and
// variable-width integer encodings are handled consistently.
func walkCFFDict(data []byte, handle func(op []byte, operands []int)) error {
	var operands []int
	for pos := 0; pos < len(data); {
		b := data[pos]
		if b <= 21 {
			if b == 12 {
				if pos+1 >= len(data) {
					return fmt.Errorf("subset cff: truncated escaped DICT operator")
				}
				handle(data[pos:pos+2], operands)
				pos += 2
			} else {
				handle(data[pos:pos+1], operands)
				pos++
			}
			operands = nil
			continue
		}
		value, n, err := parseCFFDictNumber(data[pos:])
		if err != nil {
			return err
		}
		operands = append(operands, value)
		pos += n
	}
	return nil
}

func parseCFFDictNumber(data []byte) (int, int, error) {
	if len(data) == 0 {
		return 0, 0, fmt.Errorf("subset cff: empty DICT number")
	}
	b := data[0]
	switch {
	case b >= 32 && b <= 246:
		return int(b) - 139, 1, nil
	case b >= 247 && b <= 250:
		if len(data) < 2 {
			return 0, 0, fmt.Errorf("subset cff: truncated positive DICT number")
		}
		return (int(b)-247)*256 + int(data[1]) + 108, 2, nil
	case b >= 251 && b <= 254:
		if len(data) < 2 {
			return 0, 0, fmt.Errorf("subset cff: truncated negative DICT number")
		}
		return -(int(b)-251)*256 - int(data[1]) - 108, 2, nil
	case b == 28:
		if len(data) < 3 {
			return 0, 0, fmt.Errorf("subset cff: truncated int16 DICT number")
		}
		return int(int16(binary.BigEndian.Uint16(data[1:]))), 3, nil
	case b == 29:
		if len(data) < 5 {
			return 0, 0, fmt.Errorf("subset cff: truncated int32 DICT number")
		}
		return int(int32(binary.BigEndian.Uint32(data[1:]))), 5, nil
	case b == 255:
		if len(data) < 5 {
			return 0, 0, fmt.Errorf("subset cff: truncated fixed DICT number")
		}
		return int(int32(binary.BigEndian.Uint32(data[1:])) >> 16), 5, nil
	case b == 30:
		n := 1
		for n < len(data) {
			hi, lo := data[n]>>4, data[n]&0x0f
			n++
			if hi == 0x0f || lo == 0x0f {
				return 0, n, nil
			}
		}
		return 0, 0, fmt.Errorf("subset cff: unterminated real DICT number")
	default:
		return 0, 0, fmt.Errorf("subset cff: invalid DICT number byte %d", b)
	}
}

func filterCFFDict(data []byte, skip func(op []byte) bool) ([]byte, error) {
	var out bytes.Buffer
	var operands [][]byte
	for pos := 0; pos < len(data); {
		b := data[pos]
		if b <= 21 {
			var op []byte
			if b == 12 {
				if pos+1 >= len(data) {
					return nil, fmt.Errorf("subset cff: truncated escaped DICT operator")
				}
				op = data[pos : pos+2]
				pos += 2
			} else {
				op = data[pos : pos+1]
				pos++
			}
			if !skip(op) {
				for _, operand := range operands {
					out.Write(operand)
				}
				out.Write(op)
			}
			operands = nil
			continue
		}
		_, n, err := parseCFFDictNumber(data[pos:])
		if err != nil {
			return nil, err
		}
		operands = append(operands, data[pos:pos+n])
		pos += n
	}
	return out.Bytes(), nil
}

func buildSubsetCFF(name []byte, topDict []byte, strings [][]byte, globalSubrs [][]byte, charstrings [][]byte, charsets []uint16, private []byte, localSubrs [][]byte) ([]byte, error) {
	nameIndex := buildCFFIndex([][]byte{name})
	stringIndex := buildCFFIndex(strings)
	globalIndex := buildCFFIndex(globalSubrs)
	charsetData := buildCFFCharsetFormat0(charsets)
	charStringsIndex := buildCFFIndex(charstrings)
	localIndex := buildCFFIndex(localSubrs)

	privateBase, err := filterCFFDict(private, func(op []byte) bool {
		return len(op) == 1 && op[0] == cffOpSubrs
	})
	if err != nil {
		return nil, err
	}
	privateDict := buildCFFPrivateDict(privateBase, len(localIndex) > 2)

	topBase, err := filterCFFDict(topDict, func(op []byte) bool {
		return len(op) == 1 && (op[0] == cffOpCharset || op[0] == cffOpEncoding || op[0] == cffOpCharStrings || op[0] == cffOpPrivate)
	})
	if err != nil {
		return nil, err
	}

	var out []byte
	var lastTop []byte
	for i := 0; i < 8; i++ {
		topIndex := buildCFFIndex([][]byte{lastTop})
		charsetOffset := 4 + len(nameIndex) + len(topIndex) + len(stringIndex) + len(globalIndex)
		charStringsOffset := charsetOffset + len(charsetData)
		privateOffset := charStringsOffset + len(charStringsIndex)
		top := append([]byte(nil), topBase...)
		top = appendCFFInt(top, charsetOffset)
		top = append(top, cffOpCharset)
		top = appendCFFInt(top, charStringsOffset)
		top = append(top, cffOpCharStrings)
		if len(privateDict) > 0 {
			top = appendCFFInt(top, len(privateDict))
			top = appendCFFInt(top, privateOffset)
			top = append(top, cffOpPrivate)
		}
		if bytes.Equal(top, lastTop) {
			var buf bytes.Buffer
			buf.Write([]byte{1, 0, 4, 4})
			buf.Write(nameIndex)
			buf.Write(buildCFFIndex([][]byte{top}))
			buf.Write(stringIndex)
			buf.Write(globalIndex)
			buf.Write(charsetData)
			buf.Write(charStringsIndex)
			buf.Write(privateDict)
			buf.Write(localIndex)
			out = buf.Bytes()
			break
		}
		lastTop = top
	}
	if out == nil {
		return nil, fmt.Errorf("subset cff: top dict offsets did not converge")
	}
	return out, nil
}

// buildSubsetCIDCFF serializes a CID-keyed CFF program after dense remapping.
// Absolute offsets in Top DICT, Font DICT, and Private DICT depend on the byte
// sizes of structures that contain those offsets. The builder therefore lays
// out provisional objects and iterates until encoded offsets and total sizes
// stabilize before producing the final byte stream.
func buildSubsetCIDCFF(name, topDict []byte, strings, globalSubrs, charstrings [][]byte, charsets []uint16, fdSelect []uint8, fontDicts []cffFontDictData) ([]byte, error) {
	nameIndex := buildCFFIndex([][]byte{name})
	stringIndex := buildCFFIndex(strings)
	globalIndex := buildCFFIndex(globalSubrs)
	charsetData := buildCFFCharsetFormat0(charsets)
	fdSelectData := buildCFFFDSelectFormat3(fdSelect)
	charStringsIndex := buildCFFIndex(charstrings)

	topBase, err := filterCFFDict(topDict, func(op []byte) bool {
		if len(op) == 1 {
			return op[0] == cffOpCharset || op[0] == cffOpEncoding || op[0] == cffOpCharStrings || op[0] == cffOpPrivate
		}
		return len(op) == 2 && op[0] == 12 && (op[1] == 36 || op[1] == 37)
	})
	if err != nil {
		return nil, err
	}

	privateBlocks := make([][]byte, len(fontDicts))
	fdBases := make([][]byte, len(fontDicts))
	for i, fd := range fontDicts {
		privateBase, err := filterCFFDict(fd.private, func(op []byte) bool {
			return len(op) == 1 && op[0] == cffOpSubrs
		})
		if err != nil {
			return nil, err
		}
		localIndex := buildCFFIndex(fd.localSubr)
		privateDict := buildCFFPrivateDict(privateBase, len(fd.localSubr) > 0)
		privateBlocks[i] = append(privateDict, localIndex...)
		fdBases[i], err = filterCFFDict(fd.dict, func(op []byte) bool {
			return len(op) == 1 && op[0] == cffOpPrivate
		})
		if err != nil {
			return nil, err
		}
	}

	var lastTop []byte
	lastFDDicts := make([][]byte, len(fontDicts))
	for iteration := 0; iteration < 16; iteration++ {
		topIndex := buildCFFIndex([][]byte{lastTop})
		fdArrayIndex := buildCFFIndex(lastFDDicts)
		charsetOffset := 4 + len(nameIndex) + len(topIndex) + len(stringIndex) + len(globalIndex)
		fdSelectOffset := charsetOffset + len(charsetData)
		charStringsOffset := fdSelectOffset + len(fdSelectData)
		fdArrayOffset := charStringsOffset + len(charStringsIndex)
		privateOffset := fdArrayOffset + len(fdArrayIndex)

		newFDDicts := make([][]byte, len(fontDicts))
		for i := range fontDicts {
			dict := append([]byte(nil), fdBases[i]...)
			privateDictLength := len(privateBlocks[i]) - len(buildCFFIndex(fontDicts[i].localSubr))
			if privateDictLength > 0 {
				dict = appendCFFInt(dict, privateDictLength)
				dict = appendCFFInt(dict, privateOffset)
				dict = append(dict, cffOpPrivate)
			}
			newFDDicts[i] = dict
			privateOffset += len(privateBlocks[i])
		}

		top := append([]byte(nil), topBase...)
		top = appendCFFInt(top, charsetOffset)
		top = append(top, cffOpCharset)
		top = appendCFFInt(top, charStringsOffset)
		top = append(top, cffOpCharStrings)
		top = appendCFFInt(top, fdArrayOffset)
		top = append(top, 12, 36)
		top = appendCFFInt(top, fdSelectOffset)
		top = append(top, 12, 37)

		if bytes.Equal(top, lastTop) && byteSlicesEqual(newFDDicts, lastFDDicts) {
			var buf bytes.Buffer
			buf.Write([]byte{1, 0, 4, 4})
			buf.Write(nameIndex)
			buf.Write(buildCFFIndex([][]byte{top}))
			buf.Write(stringIndex)
			buf.Write(globalIndex)
			buf.Write(charsetData)
			buf.Write(fdSelectData)
			buf.Write(charStringsIndex)
			buf.Write(buildCFFIndex(newFDDicts))
			for _, block := range privateBlocks {
				buf.Write(block)
			}
			return buf.Bytes(), nil
		}
		lastTop = top
		lastFDDicts = newFDDicts
	}
	return nil, fmt.Errorf("subset cff: CID offsets did not converge")
}

func byteSlicesEqual(a, b [][]byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !bytes.Equal(a[i], b[i]) {
			return false
		}
	}
	return true
}

// buildCFFFDSelectFormat3 run-length encodes the expanded FD assignment. Format
// 3 is deterministic and generally compact for fonts whose language regions
// use long contiguous runs of one Font DICT.
func buildCFFFDSelectFormat3(fds []uint8) []byte {
	var buf bytes.Buffer
	buf.WriteByte(3)
	if len(fds) == 0 {
		writeUint16(&buf, 0)
		writeUint16(&buf, 0)
		return buf.Bytes()
	}
	type fdRange struct {
		first int
		fd    uint8
	}
	ranges := []fdRange{{first: 0, fd: fds[0]}}
	for gid := 1; gid < len(fds); gid++ {
		if fds[gid] != fds[gid-1] {
			ranges = append(ranges, fdRange{first: gid, fd: fds[gid]})
		}
	}
	writeUint16(&buf, uint16(len(ranges)))
	for _, r := range ranges {
		writeUint16(&buf, uint16(r.first))
		buf.WriteByte(r.fd)
	}
	writeUint16(&buf, uint16(len(fds)))
	return buf.Bytes()
}

func buildCFFPrivateDict(base []byte, hasLocalSubrs bool) []byte {
	if !hasLocalSubrs {
		return base
	}
	out := append([]byte(nil), base...)
	for i := 0; i < 8; i++ {
		next := append([]byte(nil), base...)
		next = appendCFFInt(next, len(out))
		next = append(next, cffOpSubrs)
		if len(next) == len(out) {
			return next
		}
		out = next
	}
	return out
}

func buildCFFIndex(items [][]byte) []byte {
	var buf bytes.Buffer
	writeUint16(&buf, uint16(len(items)))
	if len(items) == 0 {
		return buf.Bytes()
	}
	total := 1
	for _, item := range items {
		total += len(item)
	}
	offSize := cffOffsetSize(total)
	buf.WriteByte(byte(offSize))
	offset := 1
	for _, item := range items {
		writeCFFOffset(&buf, offset, offSize)
		offset += len(item)
	}
	writeCFFOffset(&buf, offset, offSize)
	for _, item := range items {
		buf.Write(item)
	}
	return buf.Bytes()
}

func cffOffsetSize(maxOffset int) int {
	switch {
	case maxOffset <= 0xff:
		return 1
	case maxOffset <= 0xffff:
		return 2
	case maxOffset <= 0xffffff:
		return 3
	default:
		return 4
	}
}

func writeCFFOffset(buf *bytes.Buffer, value, size int) {
	for shift := (size - 1) * 8; shift >= 0; shift -= 8 {
		buf.WriteByte(byte(value >> shift))
	}
}

func buildCFFCharsetFormat0(charsets []uint16) []byte {
	var buf bytes.Buffer
	buf.WriteByte(0)
	for gid := 1; gid < len(charsets); gid++ {
		writeUint16(&buf, charsets[gid])
	}
	return buf.Bytes()
}

func appendCFFInt(out []byte, v int) []byte {
	switch {
	case v >= -107 && v <= 107:
		return append(out, byte(v+139))
	case v >= 108 && v <= 1131:
		v -= 108
		return append(out, byte(v/256+247), byte(v%256))
	case v <= -108 && v >= -1131:
		v = -v - 108
		return append(out, byte(v/256+251), byte(v%256))
	case v >= -32768 && v <= 32767:
		var b [3]byte
		b[0] = 28
		binary.BigEndian.PutUint16(b[1:], uint16(int16(v)))
		return append(out, b[:]...)
	default:
		var b [5]byte
		b[0] = 29
		binary.BigEndian.PutUint32(b[1:], uint32(int32(v)))
		return append(out, b[:]...)
	}
}

func collectCFFSubrs(charstring []byte, localSubrs, globalSubrs [][]byte, usedLocals, usedGlobals map[int]bool) {
	collectCFFSubrsWithDepth(charstring, localSubrs, globalSubrs, usedLocals, usedGlobals, 0)
}

func collectCFFSubrsWithDepth(charstring []byte, localSubrs, globalSubrs [][]byte, usedLocals, usedGlobals map[int]bool, depth int) {
	if depth > 32 {
		return
	}
	stack := make([]int, 0, 16)
	stems := 0
	for pos := 0; pos < len(charstring); {
		b := charstring[pos]
		if n, ok := type2NumberLen(charstring[pos:]); ok {
			value, _ := parseType2Number(charstring[pos:])
			stack = append(stack, value)
			pos += n
			continue
		}
		if b == 10 || b == 29 {
			if len(stack) > 0 {
				operand := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				if b == 10 {
					idx := operand + cffSubrBias(len(localSubrs))
					if idx >= 0 && idx < len(localSubrs) && !usedLocals[idx] {
						usedLocals[idx] = true
						collectCFFSubrsWithDepth(localSubrs[idx], localSubrs, globalSubrs, usedLocals, usedGlobals, depth+1)
					}
				} else {
					idx := operand + cffSubrBias(len(globalSubrs))
					if idx >= 0 && idx < len(globalSubrs) && !usedGlobals[idx] {
						usedGlobals[idx] = true
						collectCFFSubrsWithDepth(globalSubrs[idx], localSubrs, globalSubrs, usedLocals, usedGlobals, depth+1)
					}
				}
			}
			pos++
			continue
		}
		switch b {
		case 1, 3, 18, 23:
			stems += len(stack) / 2
			stack = stack[:0]
			pos++
			continue
		case 19, 20:
			stems += len(stack) / 2
			pos++
			maskBytes := (stems + 7) / 8
			if pos+maskBytes > len(charstring) {
				return
			}
			pos += maskBytes
			stack = stack[:0]
			continue
		}
		if b == 12 {
			pos += 2
		} else {
			pos++
		}
		stack = stack[:0]
	}
}

func type2NumberLen(data []byte) (int, bool) {
	if len(data) == 0 {
		return 0, false
	}
	b := data[0]
	switch {
	case b == 28:
		return 3, len(data) >= 3
	case b == 255:
		return 5, len(data) >= 5
	case b >= 32 && b <= 246:
		return 1, true
	case b >= 247 && b <= 254:
		return 2, len(data) >= 2
	default:
		return 0, false
	}
}

func parseType2Number(data []byte) (int, bool) {
	b := data[0]
	switch {
	case b >= 32 && b <= 246:
		return int(b) - 139, true
	case b >= 247 && b <= 250:
		return (int(b)-247)*256 + int(data[1]) + 108, true
	case b >= 251 && b <= 254:
		return -(int(b)-251)*256 - int(data[1]) - 108, true
	case b == 28:
		return int(int16(binary.BigEndian.Uint16(data[1:]))), true
	case b == 255:
		return int(int32(binary.BigEndian.Uint32(data[1:])) >> 16), true
	default:
		return 0, false
	}
}

func cffSubrBias(n int) int {
	switch {
	case n < 1240:
		return 107
	case n < 33900:
		return 1131
	default:
		return 32768
	}
}

// pruneCFFSubrs preserves INDEX positions instead of renumbering calls. Unused
// entries become a one-byte return program; used call operands and subroutine
// bias therefore remain valid without rewriting Type 2 charstrings.
func pruneCFFSubrs(subrs [][]byte, used map[int]bool) [][]byte {
	if len(subrs) == 0 {
		return nil
	}
	out := make([][]byte, len(subrs))
	for i, subr := range subrs {
		if used[i] {
			out[i] = subr
		} else {
			out[i] = cffReturn
		}
	}
	return out
}

func buildCFFMaxpTable(numGlyphs uint16) []byte {
	var buf bytes.Buffer
	writeUint32(&buf, 0x00005000)
	writeUint16(&buf, numGlyphs)
	return buf.Bytes()
}

type sfntTable struct {
	tag  string
	data []byte
}

func assembleSFNT(scalar uint32, tables []sfntTable) ([]byte, error) {
	sort.Slice(tables, func(i, j int) bool { return tables[i].tag < tables[j].tag })
	nTables := uint16(len(tables))
	dirEnd := 12 + 16*int(nTables)
	dataStart := dirEnd
	if rem := dataStart % 4; rem != 0 {
		dataStart += 4 - rem
	}
	type record struct {
		tag      [4]byte
		checkSum uint32
		offset   uint32
		length   uint32
		data     []byte
	}
	records := make([]record, len(tables))
	off := uint32(dataStart)
	for i, t := range tables {
		copy(records[i].tag[:], t.tag)
		records[i].length = uint32(len(t.data))
		records[i].offset = off
		records[i].data = append([]byte(nil), t.data...)
		if t.tag == "head" && len(records[i].data) >= 12 {
			binary.BigEndian.PutUint32(records[i].data[8:], 0)
		}
		records[i].checkSum = ttfTableChecksum(records[i].data)
		padded := uint32(len(records[i].data))
		if rem := padded % 4; rem != 0 {
			padded += 4 - rem
		}
		off += padded
	}

	var out bytes.Buffer
	sr, es, rs := ttfSearchRange(nTables)
	writeUint32(&out, scalar)
	writeUint16(&out, nTables)
	writeUint16(&out, sr)
	writeUint16(&out, es)
	writeUint16(&out, rs)
	for _, r := range records {
		out.Write(r.tag[:])
		writeUint32(&out, r.checkSum)
		writeUint32(&out, r.offset)
		writeUint32(&out, r.length)
	}
	for out.Len() < dataStart {
		out.WriteByte(0)
	}
	for _, r := range records {
		out.Write(r.data)
		if pad := len(r.data) % 4; pad != 0 {
			out.Write(make([]byte, 4-pad))
		}
	}
	result := out.Bytes()
	fileSum := ttfTableChecksum(result)
	adj := uint32(0xB1B0AFBA) - fileSum
	for _, r := range records {
		if string(r.tag[:]) == "head" {
			binary.BigEndian.PutUint32(result[r.offset+8:], adj)
			break
		}
	}
	return result, nil
}
