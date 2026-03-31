package ltml

// Dir represents the text/layout direction of a container.
type Dir int

const (
	DirLTR = Dir(iota)
	DirRTL
)

var dirStrings = []string{"ltr", "rtl"}

func (d Dir) String() string {
	if int(d) < 0 || int(d) >= len(dirStrings) {
		return "unknown"
	}
	return dirStrings[d]
}

func ParseDir(value string) Dir {
	if value == "rtl" {
		return DirRTL
	}
	return DirLTR
}

func IsRTL(c Container) bool {
	return c.Dir() == DirRTL
}
