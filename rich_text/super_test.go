package rich_text

import (
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/rowland/leadtype/font"
	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/ttf_fonts"
)

type SuperTest struct {
	*testing.T
}

var (
	testTTFFontSourceOnce sync.Once
	testTTFFontSource     *ttf_fonts.TtfFonts
	testTTFFontSourceErr  error
)

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

func mustTTFFamilies(families ...string) (fonts []*font.Font) {
	fc := mustTTFFontSource()
	for _, family := range families {
		if family == "" {
			fonts = append(fonts, nil)
			continue
		}
		f, err := font.New(family, options.Options{}, font.FontSources{fc})
		if err != nil {
			panic(err)
		}
		fonts = append(fonts, f)
	}
	return
}
