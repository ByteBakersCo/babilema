package generator

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ByteBakersCo/babilema/internal/config"
	"github.com/ByteBakersCo/babilema/internal/utils"
)

const eleventyCommand string = "npx eleventy"
const defaultEleventyConfigFileName string = ".eleventy.js"

type eleventyTemplateRenderer struct {
	outputDir      string
	configFilePath string
}

func NewEleventyTemplateRenderer(
	cfg config.Config,
) (eleventyTemplateRenderer, error) {
	configFilePath, err := findConfigFile(cfg)
	if err != nil {
		return eleventyTemplateRenderer{}, err
	}

	return eleventyTemplateRenderer{
		outputDir:      cfg.OutputDir,
		configFilePath: configFilePath,
	}, nil
}

// JSON returned by Eleventy
type parsedData struct {
	URL        string `json:"url"`
	InputPath  string `json:"inputPath"`
	OutputPath string `json:"outputPath"`
	Content    string `json:"content"`
}

func (renderer eleventyTemplateRenderer) Render(
	templateFilePath string,
	writer io.Writer,
	data interface{},
) error {
	if !utils.IsFileAndExists(templateFilePath) {
		return fmt.Errorf("template file not found: %s", templateFilePath)
	}

	if !hasNode() {
		return fmt.Errorf("node is not installed")
	}

	if !hasEleventy() {
		return fmt.Errorf("eleventy is not installed")
	}

	// Data population
	var fileName string
	var dataFilePath string
	var err error
	if data != nil {
		fileName = extractFileName(writer)
		if fileName == "" {
			// receiving a buffer
			fileName = templateFilePath
		}

		dataFilePath, err = createDataFile(fileName, data)
		if err != nil {
			return err
		}

		log.Println("Eleventy data file created:", dataFilePath)
	}

	inputDir := fmt.Sprintf("--input=%s", templateFilePath)
	configPath := fmt.Sprintf("--config=%s", renderer.configFilePath)

	cmd := exec.Command(
		"npx",
		"eleventy",
		inputDir,
		configPath,
		"--to=ndjson",
	)

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err = cmd.Run(); err != nil {
		return fmt.Errorf("%s - %v", out.String(), err)
	}

	var jsonData parsedData
	if err = json.Unmarshal(out.Bytes(), &jsonData); err != nil {
		return err
	}

	log.Printf("Eleventy output: %s", jsonData.OutputPath)

	if _, err = io.WriteString(writer, jsonData.Content); err != nil {
		return err
	}

	if dataFilePath != "" {
		if err = os.Remove(dataFilePath); err != nil {
			return err
		}
	}

	return nil
}

func createDataFile(parentFilePath string, data interface{}) (string, error) {
	dataFilePath := strings.TrimSuffix(
		parentFilePath,
		filepath.Ext(parentFilePath),
	) + ".11tydata.json"

	file, err := os.Create(dataFilePath)
	if err != nil {
		return "", err
	}

	encoder := json.NewEncoder(file)
	if err = encoder.Encode(data); err != nil {
		return "", err
	}

	if err = file.Close(); err != nil {
		return "", err
	}

	return dataFilePath, nil
}

func createConfigFile(path string, cfg config.Config) error {
	if cfg.OutputDir == "" {
		return errors.New("output directory is not set")
	}

	content := fmt.Sprintf(`module.exports = function(eleventyConfig) {
	return {
		dir: {
			output: "%s",
			markdownTemplateEngine: "liquid,md",
		}
	}
};`, cfg.OutputDir)

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err = file.WriteString(content); err != nil {
		return err
	}

	return file.Sync()
}

func findConfigFile(cfg config.Config) (string, error) {
	rootDir, err := utils.RootDir()
	if err != nil {
		return "", err
	}

	configFileNames := []string{
		filepath.Join(rootDir, defaultEleventyConfigFileName),
		filepath.Join(rootDir, "eleventy.config.js"),
		filepath.Join(rootDir, "eleventy.config.cjs"),
	}

	for _, filename := range configFileNames {
		if utils.IsFileAndExists(filename) {
			return filename, nil
		}
	}

	log.Println(
		"Config file not found. Creating new config file at",
		configFileNames[0],
	)
	err = createConfigFile(configFileNames[0], cfg)
	if err != nil {
		return "", err
	}

	return configFileNames[0], nil
}

func extractFileName(writer io.Writer) string {
	if file, ok := writer.(*os.File); ok {
		return file.Name()
	}

	return ""
}

func hasNode() bool {
	cmd := "node --version"
	return utils.IsCommandAvailable(cmd)
}

func hasEleventy() bool {
	cmd := eleventyCommand + " --help"
	return utils.IsCommandAvailable(cmd)
}
