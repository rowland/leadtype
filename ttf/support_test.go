package ttf

import "testing"

func expectI(t *testing.T, name string, expected, actual int) {
	if expected != actual {
		t.Errorf("%s: expected %d, got %d", name, expected, actual)
	}
}

func expectI8(t *testing.T, name string, expected, actual int8) {
	if expected != actual {
		t.Errorf("%s: expected %d, got %d", name, expected, actual)
	}
}

func expectI16(t *testing.T, name string, expected, actual int16) {
	if expected != actual {
		t.Errorf("%s: expected %d, got %d", name, expected, actual)
	}
}

func expectUI32(t *testing.T, name string, expected, actual uint32) {
	if expected != actual {
		t.Errorf("%s: expected %d, got %d", name, expected, actual)
	}
}

func expectUI16(t *testing.T, name string, expected, actual uint16) {
	if expected != actual {
		t.Errorf("%s: expected %d, got %d", name, expected, actual)
	}
}

func expectS(t *testing.T, name string, expected, actual string) {
	if expected != actual {
		t.Errorf("%s: expected %s, got %s", expected, actual)
	}
}