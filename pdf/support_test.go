package pdf

import "testing"

func check(t *testing.T, condition bool, msg string) {
	if !condition {
		t.Error(msg)
	}
}

func expectF(t *testing.T, expected, actual float32) {
	if expected != actual {
		t.Errorf("Expected %f, got %f", expected, actual)
	}
}

func expectI(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected %d, got %d", expected, actual)
	}
}

