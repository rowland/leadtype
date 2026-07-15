package ltml

import "testing"

func mustPreferredHeight(t testing.TB, widget Widget, writer Writer) float64 {
	t.Helper()
	value, err := widget.PreferredHeight(writer)
	if err != nil {
		t.Fatalf("PreferredHeight() error = %v", err)
	}
	return value
}

func mustPreferredWidth(t testing.TB, widget Widget, writer Writer) float64 {
	t.Helper()
	value, err := widget.PreferredWidth(writer)
	if err != nil {
		t.Fatalf("PreferredWidth() error = %v", err)
	}
	return value
}
