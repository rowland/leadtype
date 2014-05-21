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

// 162 ns when CodepageIndex returns slices
// 173 ns when CodepageIndex returns pointers to slices
func BenchmarkCharForCodepointForEachCodepage(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for idx := idx_CP1250; idx <= idx_CP874; idx++ {
			_, _ = idx.Codepage().CharForCodepoint(' ')
		}
	}
}
