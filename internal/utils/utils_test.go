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
