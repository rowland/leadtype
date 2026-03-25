package afm

import (
	"embed"
	"io/fs"
	"path"
	"strings"
)

//go:embed data/fonts/*.afm data/fonts/*.inf
var defaultData embed.FS

func defaultResourcePath(name string) string {
	name = path.Clean(strings.ReplaceAll(name, "\\", "/"))
	name = strings.TrimPrefix(name, "../")
	name = strings.TrimPrefix(name, "afm/")
	if strings.HasPrefix(name, "data/fonts/") {
		return name
	}
	if strings.Contains(name, "/data/fonts/") {
		if i := strings.Index(name, "/data/fonts/"); i >= 0 {
			return name[i+1:]
		}
	}
	if strings.HasPrefix(name, "fonts/") {
		return "data/" + name
	}
	if !strings.Contains(name, "/") {
		return "data/fonts/" + name
	}
	return name
}

func DefaultFontPaths() ([]string, error) {
	return fs.Glob(defaultData, "data/fonts/*.afm")
}
