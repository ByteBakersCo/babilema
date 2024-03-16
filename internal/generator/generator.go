package generator

import (
	"html/template"
	"io"
	"os"
	"strings"

	"github.com/ByteBakersCo/babilema/internal/parser"
)

func GenerateBlogPosts(
	parsedIssues []parser.ParsedIssue,
	outputDir string,
	templateFilePath string,
	outputWriter io.Writer, // for testing purposes
) error {
	path := strings.Split(templateFilePath, "/")
	templateFileName := path[len(path)-1]
	tmpl, err := template.New(templateFileName).ParseFiles(templateFilePath)
	if err != nil {
		return err
	}

	for _, file := range parsedIssues {
		writer := outputWriter

		if writer == nil {
			outputFile, error := os.Create(
				outputDir + file.Metadata.Slug + ".html",
			)
			if error != nil {
				return error
			}

			defer outputFile.Close()

			writer = outputFile
		}

		err = tmpl.Execute(writer, file)
		if err != nil {
			return err
		}
	}

	return nil
}
