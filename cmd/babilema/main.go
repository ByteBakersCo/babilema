package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/ByteBakersCo/babilema/internal/config"
	"github.com/ByteBakersCo/babilema/internal/generator"
	"github.com/ByteBakersCo/babilema/internal/parser"
	"github.com/ByteBakersCo/babilema/internal/utils"
)

func main() {
	configFilePath := flag.String(
		"config",
		utils.RootDir()+"/.babilema.yml",
		"Path to the config file",
	)

	flag.Parse()

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

	fmt.Println("Blog posts generated successfully!")
}
