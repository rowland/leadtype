// Copyright 2026 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ttf

import (
	"reflect"
	"testing"
)

func TestParseCFFFDSelectFormat0(t *testing.T) {
	got, err := parseCFFFDSelect([]byte{0xff, 0, 2, 1, 2}, 1, 3)
	if err != nil {
		t.Fatal(err)
	}
	if want := []uint8{2, 1, 2}; !reflect.DeepEqual(got, want) {
		t.Fatalf("FDSelect = %v, want %v", got, want)
	}
}

func TestParseCFFFDSelectFormat3(t *testing.T) {
	data := []byte{
		0xff,
		3, 0, 2,
		0, 0, 1,
		0, 2, 4,
		0, 5,
	}
	got, err := parseCFFFDSelect(data, 1, 5)
	if err != nil {
		t.Fatal(err)
	}
	if want := []uint8{1, 1, 4, 4, 4}; !reflect.DeepEqual(got, want) {
		t.Fatalf("FDSelect = %v, want %v", got, want)
	}
}

func TestParseCFFCharsetFormats(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want []uint16
	}{
		{name: "format0", data: []byte{0xff, 0xff, 0xff, 0, 0, 10, 0, 20, 0, 30}, want: []uint16{0, 10, 20, 30}},
		{name: "format1", data: []byte{0xff, 0xff, 0xff, 1, 0, 10, 2}, want: []uint16{0, 10, 11, 12}},
		{name: "format2", data: []byte{0xff, 0xff, 0xff, 2, 0, 10, 0, 2}, want: []uint16{0, 10, 11, 12}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseCFFCharset(tt.data, 3, 4)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("charset = %v, want %v", got, tt.want)
			}
		})
	}
}
