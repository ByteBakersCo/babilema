package config

import (
	"path/filepath"
	"testing"

	"github.com/ByteBakersCo/babilema/internal/utils/pathutils"
	"github.com/ByteBakersCo/babilema/internal/utils/testutils"
)

func TestLoadConfig(t *testing.T) {
	root, _ := pathutils.RootDir()
	expected := Config{
		TemplateRenderer:       DefaultTemplateRenderer,
		DateLayout:             DefaultDateLayout,
		WebsiteURL:             "http://localhost:8080",
		BlogTitle:              "",
		BlogPostIssuePrefix:    "[BLOG]",
		TemplatePostFilePath:   filepath.Join(root, "templates", "post.html"),
		TemplateHeaderFilePath: filepath.Join(root, "templates", "header.html"),
		TemplateFooterFilePath: filepath.Join(root, "templates", "footer.html"),
		TemplateIndexFilePath:  filepath.Join(root, "templates", "index.html"),
		CSSDir:                 filepath.Join(root, "templates", "css"),
		OutputDir:              root,
		TempDir:                filepath.Join(root, "tmp"),
	}

	configPath := filepath.Join(root, DefaultConfigFileName)

	actual, err := LoadConfig(configPath)
	if err != nil {
		t.Error(err)
	}

	if expected != actual {
		t.Error(
			testutils.FormatStruct(expected, "Expected output to be"),
			testutils.FormatStruct(actual, "\ngot"),
		)
	}

	// Test with a non-existent config file

	expected = defaultConfig(root)
	configPath = filepath.Join(root, "non-existent-config.toml")
	actual, err = LoadConfig(configPath)
	if err != nil {
		t.Error(err)
	}

	if expected != actual {
		t.Error(
			testutils.FormatStruct(expected, "Expected output to be"),
			testutils.FormatStruct(actual, "\ngot"),
		)
	}
}

func TestTrimPath(t *testing.T) {
	rootDir, _ := pathutils.RootDir()

	expected := "foo/bar/baz"
	actual, err := trimPath(filepath.Join(rootDir, "foo", "bar", "baz"))
	if err != nil {
		t.Fatal(err)
	}

	if expected != actual {
		t.Errorf("expected %s, got %s", expected, actual)
	}
}
