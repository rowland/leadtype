package pdf

import (
	"bytes"
	"flag"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/rowland/leadtype/font"
	"github.com/rowland/leadtype/ttf_fonts"
)

var update = flag.Bool("update", false, "regenerate golden files")

var (
	testTTFFontSourceOnce sync.Once
	testTTFFontSource     *ttf_fonts.TtfFonts
	testTTFFontSourceErr  error
)

// normaliseDate replaces the CreationDate value in a PDF byte slice with a
// fixed placeholder so golden-file comparisons are not date-sensitive.
var creationDateRe = regexp.MustCompile(`/CreationDate \(D:[^)]+\)`)

func normaliseDate(data []byte) []byte {
	return creationDateRe.ReplaceAll(data, []byte("/CreationDate (D:20000101000000)"))
}

// compareGolden normalises timestamps and compares got against the file at
// path. When -update is set the golden file is written instead.
func compareGolden(t *testing.T, got []byte, path string) {
	t.Helper()
	got = normaliseDate(got)
	if *update {
		if err := os.WriteFile(path, got, 0644); err != nil {
			t.Fatalf("writing golden file: %v", err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading golden file %s: %v\n(run `go test -update` to generate it)", path, err)
	}
	if !bytes.Equal(got, normaliseDate(want)) {
		t.Errorf("output differs from golden file %s\nrun `go test -update` to regenerate", path)
	}
}

type SuperTest struct {
	*testing.T
}

func (st *SuperTest) AlmostEqual(expected, actual, delta float64, msg ...string) {
	if math.Abs(expected-actual) > delta {
		st.fail(expected, actual, false, msg...)
	}
}

func (st *SuperTest) Equal(expected, actual interface{}, msg ...string) {
	if expected != actual {
		st.fail(expected, actual, false, msg...)
	}
}

func (st *SuperTest) Must(condition bool, msg ...string) {
	if !condition {
		file, line, name := st.where(2)
		st.Fatalf("\t%s:%d: %s: %v\n", file, line, name, msg)
	}
}

func (st *SuperTest) MustNot(condition bool, msg ...string) {
	if condition {
		file, line, name := st.where(2)
		st.Fatalf("\t%s:%d: %s: %v\n", file, line, name, msg)
	}
}

func (st *SuperTest) False(actual bool, msg ...string) {
	if actual {
		st.fail(false, true, false, msg...)
	}
}

func (st *SuperTest) True(actual bool, msg ...string) {
	if !actual {
		st.fail(true, false, false, msg...)
	}
}

func (st *SuperTest) where(skip int) (file string, line int, name string) {
	pc, file, line, ok := runtime.Caller(skip)
	if ok {
		name = runtime.FuncForPC(pc).Name()
		if i := strings.LastIndex(name, "."); i >= 0 {
			name = name[i+1:]
		}
		file = filepath.Base(file)
	} else {
		name = "unknown func"
		file = "unknown file"
		line = 1
	}
	return
}

func (st *SuperTest) fail(expected, actual interface{}, must bool, msg ...string) {
	file, line, name := st.where(3)
	if must {
		st.Fatalf("\t%s:%d: %s: Expected <%v>, got <%v>. %v\n", file, line, name, expected, actual, msg)
	} else {
		st.Errorf("\t%s:%d: %s: Expected <%v>, got <%v>. %v\n", file, line, name, expected, actual, msg)
	}
}

// testFontSource loads a TTF fixture from the given glob pattern and returns
// a FontSource. If the file cannot be found or loaded the test is skipped.
// Use paths relative to the module root (e.g. "../ttf/testdata/minimal.ttf").
func testFontSource(t *testing.T, pattern string) font.FontSource {
	t.Helper()
	fc, err := ttf_fonts.New(pattern)
	if err != nil || len(fc.FontInfos) == 0 {
		t.Skipf("fixture font not found at %s: %v", pattern, err)
	}
	return fc
}

func skipIfNoTTFFonts(t *testing.T) {
	t.Helper()
	if _, err := os.Stat("/Library/Fonts/Arial.ttf"); err != nil {
		t.Skip("macOS TTF fonts not available")
	}
}

func mustTTFFontSource() *ttf_fonts.TtfFonts {
	testTTFFontSourceOnce.Do(func() {
		testTTFFontSource, testTTFFontSourceErr = ttf_fonts.NewFromSystemFonts()
	})
	if testTTFFontSourceErr != nil {
		panic(testTTFFontSourceErr)
	}
	return testTTFFontSource
}
