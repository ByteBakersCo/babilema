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
	"github.com/ByteBakersCo/babilema/internal/utils/pathutils"
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
		return eleventyTemplateRenderer{}, fmt.Errorf(
			"NewEleventyTemplateRenderer(): %w",
			err,
		)
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
	data any,
) error {
	info, err := os.Stat(templateFilePath)
	if err != nil {
		return fmt.Errorf("Render(%q): %w", templateFilePath, err)
	}

	if info.IsDir() {
		return fmt.Errorf("Render(%q): is a directory", templateFilePath)
	}

	if !hasNode() {
		return fmt.Errorf("node is not installed")
	}

	if !hasEleventy() {
		return fmt.Errorf("eleventy is not installed")
	}

	// Data population
	var filename string
	var dataFilePath string
	if data != nil {
		filename = extractFileName(writer)
		if filename == "" {
			// receiving a buffer
			filename = templateFilePath
		}

		dataFilePath, err = createDataFile(filename, data)
		if err != nil {
			return fmt.Errorf("Render(%q): %w", templateFilePath, err)
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
		return fmt.Errorf("Render(%q): %w", templateFilePath, err)
	}

	log.Printf("Eleventy output: %s", jsonData.OutputPath)

	if _, err = io.WriteString(writer, jsonData.Content); err != nil {
		return fmt.Errorf("Render(%q): %w", templateFilePath, err)
	}

	if dataFilePath != "" {
		if err = os.Remove(dataFilePath); err != nil {
			return fmt.Errorf("Render(%q): %w", templateFilePath, err)
		}
	}

	return nil
}

func createDataFile(parentFilePath string, data any) (string, error) {
	dataFilePath := strings.TrimSuffix(
		parentFilePath,
		filepath.Ext(parentFilePath),
	) + ".11tydata.json"

	file, err := os.Create(dataFilePath)
	if err != nil {
		return "", fmt.Errorf("createDataFile(%q, data): %w", dataFilePath, err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err = encoder.Encode(data); err != nil {
		return "", fmt.Errorf("createDataFile(%q, data): %w", dataFilePath, err)
	}

	if err = file.Close(); err != nil {
		return "", fmt.Errorf("createDataFile(%q, data): %w", dataFilePath, err)
	}

	return dataFilePath, nil
}

func createConfigFile(path string, cfg config.Config) error {
	if cfg.OutputDir == "" {
		return errors.New("createConfigFile(): cfg.OutputDir is not set")
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
		return fmt.Errorf("createConfigFile(%q): %w", path, err)
	}
	defer file.Close()

	if _, err = file.WriteString(content); err != nil {
		return fmt.Errorf("createConfigFile(%q): %w", path, err)
	}

	return file.Sync()
}

func findConfigFile(cfg config.Config) (string, error) {
	rootDir, err := pathutils.RootDir()
	if err != nil {
		return "", fmt.Errorf("findConfigFile(): %w", err)
	}

	configFileNames := []string{
		filepath.Join(rootDir, defaultEleventyConfigFileName),
		filepath.Join(rootDir, "eleventy.config.js"),
		filepath.Join(rootDir, "eleventy.config.cjs"),
	}

	var info os.FileInfo
	for _, filename := range configFileNames {
		info, err = os.Stat(filename)
		if err != nil && errors.Is(err, os.ErrNotExist) {
			continue
		}

		if err != nil {
			return "", fmt.Errorf("findConfigFile(): %w", err)
		}

		if info.IsDir() {
			return "", fmt.Errorf(
				"findConfigFile(): %q is a directory",
				filename,
			)
		}

		return filename, nil
	}

	log.Println(
		"Config file not found. Creating new config file at",
		configFileNames[0],
	)
	err = createConfigFile(configFileNames[0], cfg)
	if err != nil {
		return "", fmt.Errorf("findConfigFile(): %w", err)
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
	command := exec.Command("/bin/sh", "-c", "node --version")
	return command.Run() == nil
}

func hasEleventy() bool {
	cmd := eleventyCommand + " --help"
	command := exec.Command("/bin/sh", "-c", cmd)
	return command.Run() == nil
}
