package config

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/BurntSushi/toml"

	"github.com/ByteBakersCo/babilema/internal/utils"
)

const DefaultConfigFileName string = ".babilema.toml"

type Config struct {
	WebsiteURL             string `toml:"website_url"`
	BlogTitle              string `toml:"blog_title"`
	BlogPostIssuePrefix    string `toml:"blog_post_issue_prefix"`
	TemplatePostFilePath   string `toml:"template_post_file_path"`
	TemplateHeaderFilePath string `toml:"template_header_file_path"`
	TemplateFooterFilePath string `toml:"template_footer_file_path"`
	TemplateIndexFilePath  string `toml:"template_index_file_path"`
	CSSDir                 string `toml:"css_dir"`
	OutputDir              string `toml:"output_dir"`
	TempDir                string `toml:"temp_dir"`
}

func DefaultConfigPath() string {
	return filepath.Join(utils.RootDir(), DefaultConfigFileName)
}

func defaultConfig(root string) Config {
	return Config{
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
}

func trimPath(path string) string {
	path = strings.TrimPrefix(path, utils.RootDir())
	path = strings.TrimPrefix(path, ".")
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimSuffix(path, "/")
	path = strings.TrimSuffix(path, ".")

	return path
}

func fillEmptyConfigFields(cfg Config) Config {
	cfg.OutputDir = trimPath(cfg.OutputDir)
	cfg.OutputDir = filepath.Join(utils.RootDir(), cfg.OutputDir)

	defaultCfg := defaultConfig(cfg.OutputDir)

	cfgVal := reflect.ValueOf(&cfg).Elem()
	defaultVal := reflect.ValueOf(defaultCfg)

	for i := 0; i < cfgVal.NumField(); i++ {
		if reflect.DeepEqual(
			cfgVal.Field(i).Interface(),
			reflect.Zero(cfgVal.Field(i).Type()).Interface(),
		) {
			cfgVal.Field(i).Set(defaultVal.Field(i))
		}
	}

	return cfg
}

func fixPaths(cfg Config) Config {
	cfg.TemplatePostFilePath = trimPath(cfg.TemplatePostFilePath)
	cfg.TemplateHeaderFilePath = trimPath(cfg.TemplateHeaderFilePath)
	cfg.TemplateFooterFilePath = trimPath(cfg.TemplateFooterFilePath)
	cfg.TemplateIndexFilePath = trimPath(cfg.TemplateIndexFilePath)
	cfg.CSSDir = trimPath(cfg.CSSDir)
	cfg.OutputDir = trimPath(cfg.OutputDir)
	cfg.TemplatePostFilePath = filepath.Join(
		utils.RootDir(),
		cfg.TemplatePostFilePath,
	)
	cfg.TemplateHeaderFilePath = filepath.Join(
		utils.RootDir(),
		cfg.TemplateHeaderFilePath,
	)
	cfg.TemplateFooterFilePath = filepath.Join(
		utils.RootDir(),
		cfg.TemplateFooterFilePath,
	)
	cfg.TemplateIndexFilePath = filepath.Join(
		utils.RootDir(),
		cfg.TemplateIndexFilePath,
	)
	cfg.CSSDir = filepath.Join(utils.RootDir(), cfg.CSSDir)
	cfg.OutputDir = filepath.Join(utils.RootDir(), cfg.OutputDir)

	return cfg
}

func LoadConfig(configFilePath string) (Config, error) {
	if _, err := os.Stat(configFilePath); errors.Is(err, os.ErrNotExist) {
		log.Println("Config file not found. Using default config.")
		return defaultConfig(utils.RootDir()), nil
	}

	cfg := Config{}
	_, err := toml.DecodeFile(configFilePath, &cfg)
	if err != nil {
		return Config{}, err
	}

	cfg = fillEmptyConfigFields(cfg)
	cfg = fixPaths(cfg)

	log.Println("Config loaded successfully from", configFilePath)

	return cfg, nil
}
