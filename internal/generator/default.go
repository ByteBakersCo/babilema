package generator

import (
	"html/template"
	"io"
)

type defaultTemplateRenderer struct{}

func NewDefaultTemplateRenderer() defaultTemplateRenderer {
	return defaultTemplateRenderer{}
}

func (defaultTemplateRenderer) Render(
	tempalteFilePath string,
	writer io.Writer,
	data interface{},
) error {
	tmplt, err := template.ParseFiles(tempalteFilePath)
	if err != nil {
		return err
	}

	err = tmplt.Execute(writer, data)
	if err != nil {
		return err
	}

	return nil
}
