package pdf

import (
	"testing"
)

func TestLinePatterns(t *testing.T) {
	expectS(t, "[] 0", LinePatterns["solid"].String())
	expectS(t, "[1 2] 0", LinePatterns["dotted"].String())
	expectS(t, "[4 2] 0", LinePatterns["dashed"].String())
	check(t, LinePatterns["bogus"] == nil, "No pattern should be returned for bogus")
}
