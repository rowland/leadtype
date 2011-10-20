package ttf

import "testing"

func TestFontCollection_Len(t *testing.T) {
	
	var fc FontCollection
	
	if err := fc.Add("/Library/Fonts/*.ttf"); err != nil {
		t.Errorf(err.String())
	}
	
	expectI(t, "Len", 102, fc.Len())
	
}
