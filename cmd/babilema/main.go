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
		config.DefaultConfigPath(),
		"Path to the config file",
	)

	flag.Parse()

	if *configFilePath == "" {
		*configFilePath = config.DefaultConfigPath()
	}

	cfg, err := config.LoadConfig(*configFilePath)
	if err != nil {
		log.Fatal("Error loading config:", err)
	}

	log.Println("Config loaded successfully from", *configFilePath)

	parsedIssues, err := parser.ParseIssues(cfg)
	if err != nil {
		log.Fatal("Error parsing issues:", err)
	}

	err = generator.GenerateBlogPosts(parsedIssues, cfg, nil)
	if err != nil {
		log.Fatal("Error generating blog posts:", err)
	}
}
