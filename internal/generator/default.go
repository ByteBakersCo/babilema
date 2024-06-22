package generator

import (
	"html/template"
	"io"
)

type defaultTemplateEngine struct {
}

func (defaultTemplateEngine) Generate(
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

func NewDefaultTemplateEngine() defaultTemplateEngine {
	return defaultTemplateEngine{}
}
