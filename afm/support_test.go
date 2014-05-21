// Copyright 2011-2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package afm

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func where(skip int) (file string, line int, name string) {
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

func expect(t *testing.T, name string, condition bool) {
	if !condition {
		file, line, _ := where(2)
		t.Errorf("%s:%d: %s: failed condition", file, line, name)
	}
}

func expectI(t *testing.T, name string, expected, actual int) {
	if expected != actual {
		file, line, _ := where(2)
		t.Errorf("%s:%d: %s: expected %d, got %d", file, line, name, expected, actual)
	}
}

func expectI8(t *testing.T, name string, expected, actual int8) {
	if expected != actual {
		file, line, _ := where(2)
		t.Errorf("%s:%d: %s: expected %d, got %d", file, line, name, expected, actual)
	}
}

func expectI16(t *testing.T, name string, expected, actual int16) {
	if expected != actual {
		file, line, _ := where(2)
		t.Errorf("%s:%d: %s: expected %d, got %d", file, line, name, expected, actual)
	}
}

func expectI32(t *testing.T, name string, expected, actual int32) {
	if expected != actual {
		file, line, _ := where(2)
		t.Errorf("%s:%d: %s: expected %d, got %d", file, line, name, expected, actual)
	}
}

func expectUI32(t *testing.T, name string, expected, actual uint32) {
	if expected != actual {
		file, line, _ := where(2)
		t.Errorf("%s:%d: %s: expected %d, got %d", file, line, name, expected, actual)
	}
}

func expectUI16(t *testing.T, name string, expected, actual uint16) {
	if expected != actual {
		file, line, _ := where(2)
		t.Errorf("%s:%d: %s: expected %d, got %d", file, line, name, expected, actual)
	}
}

func expectS(t *testing.T, name string, expected, actual string) {
	if expected != actual {
		file, line, _ := where(2)
		t.Errorf("%s:%d: %s: expected \"%s\", got \"%s\"", file, line, name, expected, actual)
	}
}

func expectF(t *testing.T, name string, expected, actual float64) {
	if expected != actual {
		file, line, _ := where(2)
		t.Errorf("%s:%d: %s: expected %f, got %f", file, line, name, expected, actual)
	}
}
