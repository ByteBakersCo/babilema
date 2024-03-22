package config

import (
	"testing"

	"github.com/ByteBakersCo/babilema/internal/utils"
)

func TestLoadConfig(t *testing.T) {
	root := utils.RootDir()
	expected := Config{
		WebsiteURL:             "https://babilema.github.io",
		BlogTitle:              "Babilema: A Minimalist Static Blog Generator",
		BlogPostIssuePrefix:    "[BLOG]",
		CommitMessage:          "Babilema: generate blog",
		TemplatePostFilePath:   root + "/templates/post.html",
		TemplateHeaderFilePath: root + "/templates/header.html",
		TemplateFooterFilePath: root + "/templates/footer.html",
		TemplateIndexFilePath:  root + "/templates/index.html",
		CSSDir:                 root + "/templates/css",
		OutputDir:              root,
	}

	actual, err := LoadConfig(root + DefaultConfigFileName)
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
