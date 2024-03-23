package config

import (
	"path/filepath"
	"testing"

	"github.com/ByteBakersCo/babilema/internal/utils"
)

func TestLoadConfig(t *testing.T) {
	root := utils.RootDir()
	expected := Config{
		WebsiteURL:             "https://babilema.github.io",
		BlogTitle:              "Babilema: A Minimalist Static Blog Generator",
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
			utils.FormatStruct(expected, "Expected output to be"),
			utils.FormatStruct(actual, "\ngot"),
		)
	}
}
