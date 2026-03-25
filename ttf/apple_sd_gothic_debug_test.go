package ttf

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDebugAppleSDGothicNeoLoad(t *testing.T) {
	path := findSystemFontFile("AppleSDGothicNeo.ttc")
	if path == "" {
		t.Skip("AppleSDGothicNeo.ttc not found in system font directories")
	}

	infos, err := LoadFontInfosFromTTC(path)
	if err != nil {
		t.Skipf("LoadFontInfosFromTTC(%q): %v", path, err)
	}
	if len(infos) == 0 {
		t.Skipf("no TTC faces found in %q", path)
	}

	for i, fi := range infos {
		if i >= 4 {
			break
		}
		if _, err := LoadFontAtOffset(fi.Filename(), fi.TTCOffset()); err != nil {
			t.Fatalf("LoadFontAtOffset(%s, %d) for %s/%s: %v", fi.Filename(), fi.TTCOffset(), fi.Family(), fi.Style(), err)
		}
	}
}

func findSystemFontFile(base string) string {
	dirs := []string{}
	if home, err := os.UserHomeDir(); err == nil {
		dirs = append(dirs, filepath.Join(home, "Library", "Fonts"))
	}
	dirs = append(dirs,
		"/Library/Fonts",
		"/System/Library/Fonts/Supplemental",
		"/System/Library/Fonts",
	)
	for _, dir := range dirs {
		path := filepath.Join(dir, base)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}
