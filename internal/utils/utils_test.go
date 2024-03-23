package utils

import (
	"path/filepath"
	"testing"
)

func TestRootDir(t *testing.T) {
	expected := "babilema"
	actual := filepath.Base(RootDir())
	if actual != expected {
		t.Errorf("Expected output to be %s, got %s", expected, actual)
	}
}

func TestRelativeFilePath(t *testing.T) {
	expected := filepath.Join("/", "internal", "utils", "utils.go")
	actual, _ := RelativeFilePath(
		filepath.Join(RootDir(), "internal", "utils", "utils.go"),
	)
	if actual != expected {
		t.Errorf("Expected output to be %s, got %s", expected, actual)
	}
}
