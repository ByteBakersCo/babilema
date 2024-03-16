package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/ByteBakersCo/babilema/internal/generator"
	"github.com/ByteBakersCo/babilema/internal/parser"
)

func main() {
	outputDir := flag.String(
		"output-dir",
		"generated/",
		"Directory where the generated HTML files should be saved",
	)
	templateFile := flag.String(
		"template-file",
		"templates/post.html",
		"HTML template file to use for the blog posts",
	)

	flag.Parse()

	if *outputDir == "" || *templateFile == "" {
		log.Fatal(
			"Error: You must provide an output directory and template file.",
		)
	}

	parsedIssues, err := parser.ParseIssues()
	if err != nil {
		log.Fatal("Error parsing issues:", err)
	}

	err = generator.GenerateBlogPosts(
		parsedIssues,
		*outputDir,
		*templateFile,
		nil,
	)
	if err != nil {
		log.Fatal("Error generating blog posts:", err)
	}

	fmt.Println("Blog posts generated successfully!")
}
