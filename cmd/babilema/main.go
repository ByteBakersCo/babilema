package main

import (
	"flag"
	"log"

	"github.com/ByteBakersCo/babilema/internal/config"
	"github.com/ByteBakersCo/babilema/internal/generator"
	"github.com/ByteBakersCo/babilema/internal/parser"
)

func main() {
	configFilePath := flag.String(
		"config",
		"",
		"Path to the config file",
	)

	flag.Parse()

	if *configFilePath == "" {
		defaultCfgPath, err := config.DefaultConfigPath()
		if err != nil {
			return
		}

		*configFilePath = defaultCfgPath
	}

	cfg, err := config.LoadConfig(*configFilePath)
	if err != nil {
		log.Fatalln("Error loading config:", err)
	}

	parsedIssues, err := parser.ParseIssues(cfg)
	if err != nil {
		log.Fatalln("Error parsing issues:", err)
	}

	err = generator.GenerateBlogPosts(parsedIssues, cfg, nil)
	if err != nil {
		log.Fatalln("Error generating blog posts:", err)
	}
}
