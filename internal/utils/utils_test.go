package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRootDir(t *testing.T) {
	wd, _ := os.Getwd()
	parent := filepath.Dir(wd)
	grandparent := filepath.Dir(parent)
	expected := filepath.Base(grandparent)
	rootDir, _ := RootDir()
	actual := filepath.Base(rootDir)
	if actual != string(expected) {
		t.Errorf("Expected output to be %s, got %s", expected, actual)
	}
}

func TestRelativeFilePath(t *testing.T) {
	rootDir, _ := RootDir()
	expected := filepath.Join("/", "internal", "utils", "utils.go")
	actual, _ := RelativeFilePath(
		filepath.Join(rootDir, "internal", "utils", "utils.go"),
	)
	if actual != expected {
		t.Errorf("Expected output to be %s, got %s", expected, actual)
	}
}

func TestIsCommandAvailable(t *testing.T) {
	cmd := "echo hello"
	if !IsCommandAvailable(cmd) {
		t.Errorf("Expected command to be available")
	}

	cmd = "qwfpbt"
	if IsCommandAvailable(cmd) {
		t.Errorf("Expected command to not be available")
	}
}
