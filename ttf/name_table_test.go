package ttf

import "testing"

func TestNameTable(t *testing.T) {
	f, err := LoadFont("/Library/Fonts/Arial.ttf")
	if err != nil {
		t.Fatalf("Error loading font: %s", err)
	}
	expectS(t, "postScriptName", "ArialMT", f.nameTable.postScriptName)
	expectS(t, "fullName", "Arial", f.nameTable.fullName)
	expectS(t, "family", "Arial", f.nameTable.fontFamily)
	expectS(t, "fontSubfamily", "Regular", f.nameTable.fontSubfamily)
	expectS(t, "version", "Version 5.01.2x", f.nameTable.version)
	expectS(t, "uniqueSubfamily", "Monotype:Arial Regular:Version 5.01 (Microsoft)", f.nameTable.uniqueSubfamily)
	expectS(t, "manufacturerName", "The Monotype Corporation", f.nameTable.manufacturerName)
	expectS(t, "designer", "Monotype Type Drawing Office - Robin Nicholas, Patricia Saunders 1982", f.nameTable.designer)
	expectS(t, "copyrightNotice", "© 2006 The Monotype Corporation. All Rights Reserved.", f.nameTable.copyrightNotice)
	expectS(t, "trademarkNotice", "Arial is a trademark of The Monotype Corporation in the United States and/or other countries.", f.nameTable.trademarkNotice)
	expectS(t, "licenseDescription", "You may use this font to display and print content as permitted by the license terms for the product in which this font is included. You may only (i) embed this font in content as permitted by the embedding restrictions included in this font; and (ii) temporarily download this font to a printer or other output device to help print content.", f.nameTable.licenseDescription)
}
