package utils

import (
	"strings"
	"testing"
)

func TestRootDir(t *testing.T) {
	expected := "babilema"
	actual := strings.Split(RootDir(), "/")[len(strings.Split(RootDir(), "/"))-1]
	if actual != expected {
		t.Errorf("Expected output to be %s, got %s", expected, actual)
	}
}

func TestRelativeFilePath(t *testing.T) {
	expected := "/internal/utils/utils.go"
	actual, _ := RelativeFilePath(RootDir() + "/internal/utils/utils.go")
	if actual != expected {
		t.Errorf("Expected output to be %s, got %s", expected, actual)
	}
}
