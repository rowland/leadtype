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
		if idx.String() != CodepageNames[idx] {
			t.Errorf("Expected %s, got %s", CodepageNames[idx], idx.String())
		}
	}
}
