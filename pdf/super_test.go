package pdf

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

type SuperTest struct {
	*testing.T
}

func (st *SuperTest) Equal(expected interface{}, actual interface{}, msg ...string) {
	if expected != actual {
		st.fail(expected, actual, false, msg...)
	}
}

func (st *SuperTest) MustEqual(expected interface{}, actual interface{}, msg ...string) {
	if expected != actual {
		st.fail(expected, actual, true, msg...)
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

func (st *SuperTest) Nil(actual interface{}, msg ...string) {
	if actual != nil {
		st.fail(nil, actual, false, msg...)
	}
}

func (st *SuperTest) fail(expected, actual interface{}, must bool, msg ...string) {
	pc, file, line, ok := runtime.Caller(2)
	var name string
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
	if must {
		st.Fatalf("\t%s:%d: %s: Expected %v, got %v. %v\n", file, line, name, expected, actual, msg)
	} else {
		st.Errorf("\t%s:%d: %s: Expected %v, got %v. %v\n", file, line, name, expected, actual, msg)
	}
}
