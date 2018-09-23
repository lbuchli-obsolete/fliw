package cli

import "testing"

func TestGetDir(t *testing.T) {
	input := "/home/test/go/src/testproject/testfile.go"
	expected := "/home/test/go/src/testproject/"

	result := getDir(input)

	if result != expected {
		t.Error("Expected: ", expected, ", Gave: ", result)
	}
}
