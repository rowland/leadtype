package ttf

import "testing"

func TestNewFont(t *testing.T) {
	f, err := NewFont("/Library/Fonts/Arial.ttf")
	if err != nil {
		t.Fatalf("Error creating font: %s", err)
	}
	if f == nil {
		t.Fatal("Font not created")
	}
	
	if f.scalar != 0x00010000 {
		t.Errorf("Scalar: expected %d, got %d", 0x00010000, f.scalar)
	}
	if f.nTables != 0x0018 {
		t.Errorf("nTables: expected %d, got %d", 0x0018, f.nTables)
	}
}
