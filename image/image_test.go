package image

import (
	"testing"
)

func TestUInt32ToColor(t *testing.T) {
	in := uint32(0x12345678)
	expected := uint32(0x12785634)
	if result := UInt32ToColor(in).Uint32(); result != expected {
		t.Errorf("Expected 0x%x, gave 0x%x", expected, result)
	}
}
