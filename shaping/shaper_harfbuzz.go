// Copyright 2024 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

//go:build harfbuzz

package shaping

// #cgo pkg-config: harfbuzz
// #include <hb.h>
// #include <stdlib.h>
import "C"
import (
	"fmt"
	"unsafe"
)

// NewShaper returns a CGO Arabic shaper backed by the system libharfbuzz.
// Build with: go build -tags harfbuzz ./...
// Requires: libharfbuzz-dev (or equivalent) installed on the build host.
func NewShaper() Shaper { return &hbShaper{} }

type hbShaper struct{}

// Shape shapes a run of Arabic text using the system libharfbuzz via CGO.
// The returned glyphs are in visual (display) order.
func (h *hbShaper) Shape(text []rune, fontBytes []byte, ppem float32) ([]GlyphPosition, error) {
	if len(fontBytes) == 0 {
		return nil, fmt.Errorf("shaping: empty font data")
	}

	// Load font from raw bytes.
	blob := C.hb_blob_create(
		(*C.char)(unsafe.Pointer(&fontBytes[0])),
		C.uint(len(fontBytes)),
		C.HB_MEMORY_MODE_READONLY,
		nil,
		nil,
	)
	defer C.hb_blob_destroy(blob)

	face := C.hb_face_create(blob, 0)
	defer C.hb_face_destroy(face)

	hbFont := C.hb_font_create(face)
	defer C.hb_font_destroy(hbFont)

	// Scale in 1/64th units to match the 26.6 fixed-point convention used
	// by the rest of the shaping package.
	scale := C.int(ppem * 64)
	C.hb_font_set_scale(hbFont, scale, scale)

	// Build the shaping buffer.
	buf := C.hb_buffer_create()
	defer C.hb_buffer_destroy(buf)

	C.hb_buffer_set_direction(buf, C.HB_DIRECTION_RTL)
	C.hb_buffer_set_script(buf, C.HB_SCRIPT_ARABIC)

	lang := C.hb_language_from_string(C.CString("ar"), -1)
	C.hb_buffer_set_language(buf, lang)

	// Encode the rune slice as UTF-8 and add it to the buffer.
	utf8 := []byte(string(text))
	cs := (*C.char)(C.CBytes(utf8))
	defer C.free(unsafe.Pointer(cs))
	C.hb_buffer_add_utf8(buf, cs, C.int(len(utf8)), 0, C.int(len(utf8)))

	// Shape.
	C.hb_shape(hbFont, buf, nil, 0)

	// Retrieve results.
	var count C.uint
	infoPtr := C.hb_buffer_get_glyph_infos(buf, &count)
	posPtr := C.hb_buffer_get_glyph_positions(buf, &count)

	n := int(count)
	infos := unsafe.Slice(infoPtr, n)
	positions := unsafe.Slice(posPtr, n)

	// Build a byte-offset → rune-index mapping so ClusterIndex matches
	// the []rune indices used by the rest of the shaping package.
	// HarfBuzz sets cluster to the UTF-8 byte offset of the source character.
	byteToRune := make(map[int]int, len(text))
	ri := 0
	for bi := range string(text) {
		byteToRune[bi] = ri
		ri++
	}

	glyphs := make([]GlyphPosition, n)
	for i := range glyphs {
		glyphs[i] = GlyphPosition{
			GlyphID:      uint16(infos[i].codepoint),
			XAdvance:     int32(positions[i].x_advance),
			YAdvance:     int32(positions[i].y_advance),
			XOffset:      int32(positions[i].x_offset),
			YOffset:      int32(positions[i].y_offset),
			ClusterIndex: byteToRune[int(infos[i].cluster)],
		}
	}
	return glyphs, nil
}
