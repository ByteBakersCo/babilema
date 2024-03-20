package config

import (
	"testing"

	"github.com/ByteBakersCo/babilema/internal/utils"
)

func TestLoadConfig(t *testing.T) {
	root := utils.RootDir()
	expected := Config{
		WebsiteURL:             "babilema.github.io",
		BlogPostIssuePrefix:         "[BLOG]",
		TemplatePostFilePath:   root + "/templates/post.html",
		TemplateHeaderFilePath: root + "/templates/header.html",
		TemplateFooterFilePath: root + "/templates/footer.html",
		CSSDir:                 root + "/templates/css",
		OutputDir:              root,
	}

	cfg, err := LoadConfig(root + DefaultConfigPath)
	if err != nil {
		t.Error(err)
	}

	if expected != cfg {
		t.Errorf("expected %+v, got %+v", expected, cfg)
	}
}
