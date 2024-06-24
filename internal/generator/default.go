package generator

import (
	"fmt"
	"html/template"
	"io"

	"github.com/ByteBakersCo/babilema/internal/utils"
)

type defaultTemplateRenderer struct{}

func NewDefaultTemplateRenderer() defaultTemplateRenderer {
	return defaultTemplateRenderer{}
}

func (defaultTemplateRenderer) Render(
	templateFilePath string,
	writer io.Writer,
	data interface{},
) error {
	if !utils.IsFileAndExists(templateFilePath) {
		return fmt.Errorf("template file not found: %s", templateFilePath)
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
