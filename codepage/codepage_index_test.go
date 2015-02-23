// Copyright 2014 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package codepage

import "testing"
import "reflect"

func TestCodepageIndex_Codepage(t *testing.T) {
	for idx := idx_CP1250; idx <= idx_CP874; idx++ {
		if !reflect.DeepEqual(idx.Codepage(), Codepages[idx]) {
			t.Errorf("Expected %#v, got %#v", Codepages[idx], idx)
		}
		if idx.String() != codepageNames[idx] {
			t.Errorf("Expected %s, got %s", codepageNames[idx], idx.String())
		}
	}
}

func TestCodepageIndex_Map(t *testing.T) {
	m := idx_ISO_8859_1.Map()
	for i := 0; i < 256; i++ {
		if m[i] != rune(i) {
			t.Errorf("Expected %d, got %d", i, m[i])
		}
	}
}

func TestCodepageIndex_Map_default(t *testing.T) {
	bogus := CodepageIndex(-1)
	defer func() {
		if p := recover(); p != nil {
			t.Error(p)
		}
	}()
	if !reflect.DeepEqual(bogus.Map(), ISO_8859_1_Map) {
		t.Error("Expected map ISO-8859-1 for invalid CodepageIndex.")
	}
}

// 162 ns when CodepageIndex returns slices
// 173 ns when CodepageIndex returns pointers to slices
func BenchmarkCharForCodepointForEachCodepage(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for idx := idx_CP1250; idx <= idx_CP874; idx++ {
			_, _ = idx.Codepage().CharForCodepoint(' ')
		}
	}
}
