package config

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/ByteBakersCo/babilema/internal/utils/pathutils"
)

const DefaultConfigFileName string = ".babilema.toml"
const DefaultDateLayout string = "2006-01-02 15:04:05"

type TemplateRenderer string

const (
	DefaultTemplateRenderer  TemplateRenderer = "default"
	EleventyTemplateRenderer TemplateRenderer = "eleventy"
)

type Config struct {
	TemplateRenderer       TemplateRenderer `toml:"template_renderer"`
	DateLayout             string           `toml:"date_layout"`
	WebsiteURL             string           `toml:"website_url"`
	BlogTitle              string           `toml:"blog_title"`
	BlogPostIssuePrefix    string           `toml:"blog_post_issue_prefix"`
	TemplatePostFilePath   string           `toml:"template_post_file_path"`
	TemplateHeaderFilePath string           `toml:"template_header_file_path"`
	TemplateFooterFilePath string           `toml:"template_footer_file_path"`
	TemplateIndexFilePath  string           `toml:"template_index_file_path"`
	CSSDir                 string           `toml:"css_dir"`
	OutputDir              string           `toml:"output_dir"`
	TempDir                string           `toml:"temp_dir"`
}

func DefaultConfigPath() (string, error) {
	rootDir, err := pathutils.RootDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(rootDir, DefaultConfigFileName), nil
}

func defaultConfig(root string) Config {
	return Config{
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
}

func trimPath(path string) (string, error) {
	rootDir, err := pathutils.RootDir()
	if err != nil {
		return "", err
	}

	path = strings.TrimPrefix(path, rootDir)
	path = strings.TrimPrefix(path, ".")
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimSuffix(path, "/")
	path = strings.TrimSuffix(path, ".")

	return path, nil
}

func fillEmptyConfigFields(cfg Config) (Config, error) {
	if cfg.OutputDir == "" {
		return Config{}, errors.New("output directory not set")
	}

	outputDir, err := trimPath(cfg.OutputDir)
	if err != nil {
		return Config{}, err
	}

	rootDir, _ := pathutils.RootDir()
	cfg.OutputDir = filepath.Join(rootDir, outputDir)

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

	return cfg, nil
}

func fixPaths(cfg Config) (Config, error) {
	if cfg.TemplatePostFilePath == "" {
		return Config{}, errors.New("post template file path not set")
	}

	if cfg.TemplateHeaderFilePath == "" {
		return Config{}, errors.New("header template file path not set")
	}

	if cfg.TemplateFooterFilePath == "" {
		return Config{}, errors.New("footer template file path not set")
	}

	if cfg.TemplateIndexFilePath == "" {
		return Config{}, errors.New("index template file path not set")
	}

	if cfg.CSSDir == "" {
		return Config{}, errors.New("css directory not set")
	}

	if cfg.OutputDir == "" {
		return Config{}, errors.New("output directory not set")
	}

	if cfg.TempDir == "" {
		return Config{}, errors.New("temp directory not set")
	}

	rootDir, err := pathutils.RootDir()
	if err != nil {
		return Config{}, err
	}

	cfg.TemplatePostFilePath, _ = trimPath(cfg.TemplatePostFilePath)
	cfg.TemplateHeaderFilePath, _ = trimPath(cfg.TemplateHeaderFilePath)
	cfg.TemplateFooterFilePath, _ = trimPath(cfg.TemplateFooterFilePath)
	cfg.TemplateIndexFilePath, _ = trimPath(cfg.TemplateIndexFilePath)
	cfg.CSSDir, _ = trimPath(cfg.CSSDir)
	cfg.OutputDir, _ = trimPath(cfg.OutputDir)
	cfg.TempDir, _ = trimPath(cfg.TempDir)

	cfg.TemplatePostFilePath = filepath.Join(
		rootDir,
		cfg.TemplatePostFilePath,
	)
	cfg.TemplateHeaderFilePath = filepath.Join(
		rootDir,
		cfg.TemplateHeaderFilePath,
	)
	cfg.TemplateFooterFilePath = filepath.Join(
		rootDir,
		cfg.TemplateFooterFilePath,
	)
	cfg.TemplateIndexFilePath = filepath.Join(
		rootDir,
		cfg.TemplateIndexFilePath,
	)
	cfg.CSSDir = filepath.Join(rootDir, cfg.CSSDir)
	cfg.OutputDir = filepath.Join(rootDir, cfg.OutputDir)
	cfg.TempDir = filepath.Join(rootDir, cfg.TempDir)

	return cfg, nil
}

func LoadConfig(configFilePath string) (Config, error) {
	rootDir, err := pathutils.RootDir()
	if err != nil {
		return Config{}, err
	}

	if _, err = os.Stat(configFilePath); errors.Is(err, os.ErrNotExist) {
		log.Println("Config file not found. Using default config.")
		return defaultConfig(rootDir), nil
	}

	cfg := Config{}
	_, err = toml.DecodeFile(configFilePath, &cfg)
	if err != nil {
		return Config{}, err
	}

	cfg, _ = fillEmptyConfigFields(cfg)
	cfg, _ = fixPaths(cfg)

	cfg.TemplateRenderer = TemplateRenderer(
		strings.ToLower(string(cfg.TemplateRenderer)),
	)

	log.Println("Config loaded successfully from", configFilePath)

	return cfg, nil
}
