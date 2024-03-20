package history

import (
	"maps"
	"os"
	"testing"
	"time"

	"github.com/ByteBakersCo/babilema/internal/config"
	"github.com/ByteBakersCo/babilema/internal/utils"
)

func cleanup() {
	os.Remove(historyFileName)
}

func TestUpdateHistoryFile(t *testing.T) {
	history := map[string]string{
		"foo": time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
	}
	err := UpdateHistoryFile(history, config.Config{
		OutputDir: ".",
	})
	if err != nil {
		t.Error(err)
	}

	expected := `# This file is auto-generated by Babilema. Do not edit manually.

[history]
  foo = "1970-01-01T00:00:00Z"
`

	content, err := os.ReadFile(historyFileName)
	if err != nil {
		t.Error(err)
	}

	if string(content) != expected {
		t.Errorf("expected %q, got %q", expected, string(content))
	}
}

func TestParseHistoryFile(t *testing.T) {
	defer cleanup()

	expected := map[string]string{
		"foo": time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
	}

	actual, err := ParseHistoryFile(config.Config{
		OutputDir: ".",
	})
	if err != nil {
		t.Error(err)
	}

	if !maps.Equal(expected, actual) {
		t.Error(
			utils.FormatStruct(expected, "Expected output to be"),
			utils.FormatStruct(actual, "\ngot"),
		)
	}
}
