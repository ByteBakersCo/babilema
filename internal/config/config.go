package config

import "github.com/BurntSushi/toml"

type Config struct {
	WebsiteURL     string `toml:"website_url"`
	BlogPostPrefix string `toml:"blog_post_prefix"`
}

func LoadConfig() (Config, error) {
	var cfg Config
	_, err := toml.DecodeFile("babilema.toml", &cfg)
	if err != nil {
		return Config{}, err
	}

	if cfg.BlogPostPrefix == "" {
		cfg.BlogPostPrefix = "[BLOG]"
	}

	return cfg, nil
}
