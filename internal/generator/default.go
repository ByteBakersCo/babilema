package generator

import (
	"fmt"
	"html/template"
	"io"
	"os"
)

type defaultTemplateRenderer struct{}

func NewDefaultTemplateRenderer() defaultTemplateRenderer {
	return defaultTemplateRenderer{}
}

func (defaultTemplateRenderer) Render(
	templateFilePath string,
	writer io.Writer,
	data any,
) error {
	info, err := os.Stat(templateFilePath)
	if err != nil {
		return fmt.Errorf("Render(%q): %w", templateFilePath, err)
	}

	if info.IsDir() {
		return fmt.Errorf("Render(%q): is a directory", templateFilePath)
	}

	tmplt, err := template.ParseFiles(templateFilePath)
	if err != nil {
		return err
	}

	if err = tmplt.Execute(writer, data); err != nil {
		return err
	}

	return nil
}
