// Copyright 2026 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ttf

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"sort"
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

type cffSubsetData struct {
	name       []byte
	topDict    []byte
	strings    [][]byte
	globalSubr [][]byte
	charsets   []uint16
	charstring [][]byte
	private    []byte
	localSubr  [][]byte
}

type cffTopDictInfo struct {
	charsetOffset     int
	charStringsOffset int
	privateOffset     int
	privateLength     int
	isCID             bool
}

type cffPrivateInfo struct {
	subrsOffset int
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

func cffSubsetKeepTable(tag string) bool {
	switch tag {
	case "CFF ", "cmap", "head", "hhea", "hmtx", "maxp", "name", "OS/2", "post":
		return true
	default:
		return false
	}
}

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
	if topInfo.isCID {
		return nil, fmt.Errorf("subset cff: CID-keyed CFF fonts are not supported yet")
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
	if topInfo.privateLength > 0 {
		if topInfo.privateOffset < 0 || topInfo.privateOffset+topInfo.privateLength > len(data) {
			return nil, fmt.Errorf("subset cff: invalid Private DICT bounds")
		}
		private = data[topInfo.privateOffset : topInfo.privateOffset+topInfo.privateLength]
		privInfo, err := parseCFFPrivateDict(private)
		if err != nil {
			return nil, err
		}
		if privInfo.subrsOffset > 0 {
			localSubrs, _, err = parseCFFIndex(data, topInfo.privateOffset+privInfo.subrsOffset)
			if err != nil {
				return nil, err
			}
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
	}, nil
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
		} else if len(op) == 2 && op[0] == 12 && op[1] == 30 {
			info.isCID = true
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
