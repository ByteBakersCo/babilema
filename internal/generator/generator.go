package generator

import (
	"html/template"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ByteBakersCo/babilema/internal/config"
	"github.com/ByteBakersCo/babilema/internal/parser"
)

type templateData struct {
	parser.ParsedIssue
	Header   template.HTML
	Footer   template.HTML
	CSSLinks []string
}

func extractHTML(filePath string) (template.HTML, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	return template.HTML(content), nil
}

func extractCSSLinks(cssDir string) ([]string, error) {
	if cssDir == "" {
		return nil, nil
	}

	var cssLinks []string
	err := filepath.Walk(
		cssDir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() && strings.HasSuffix(path, ".css") {
				cssLinks = append(cssLinks, path)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return cssLinks, nil
}

func GenerateBlogPosts(
	parsedIssues []parser.ParsedIssue,
	cfg config.Config,
	testOutputWriter io.Writer, // for testing purposes
) error {
	templatePostFilePath := cfg.TemplatePostFilePath
	templateHeaderFilePath := cfg.TemplateHeaderFilePath
	templateFooterFilePath := cfg.TemplateFooterFilePath
	cssDir := cfg.CSSDir
	outputDir := cfg.OutputDir

	tmpl, err := template.ParseFiles(templatePostFilePath)
	if err != nil {
		return err
	}

	data := templateData{}
	data.CSSLinks, err = extractCSSLinks(cssDir)
	if err != nil {
		return err
	}

	data.Header, err = extractHTML(templateHeaderFilePath)
	if err != nil {
		return err
	}

	data.Footer, err = extractHTML(templateFooterFilePath)
	if err != nil {
		return err
	}

	for _, issue := range parsedIssues {
		data.ParsedIssue = issue
		writer := testOutputWriter

		if writer == nil {
			outputFile, error := os.Create(
				outputDir + data.Metadata.Slug + ".html",
			)
			if error != nil {
				return error
			}

			defer outputFile.Close()

			writer = outputFile
		}

		log.Println("Generating blog post: " + data.Metadata.Slug)
		err = tmpl.Execute(writer, data)
		if err != nil {
			return err
		}
	}

	return nil
}
