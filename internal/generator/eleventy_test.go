package generator

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/ByteBakersCo/babilema/internal/config"
	"github.com/ByteBakersCo/babilema/internal/utils"
)

/*
1. has eleventy?
	y - run eleventy
	n - return error
2. run eleventy has templating engine
*/

func TestNewEleventyTemplateRenderer(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	basePath := filepath.Dir(file)

	cfg := config.Config{
		OutputDir: filepath.Join(basePath, "test-data"),
	}

	expected := eleventyTemplateRenderer{
		outputDir:      cfg.OutputDir,
		configFilePath: "",
	}

	engine, err := NewEleventyTemplateRenderer(cfg)
	if err != nil {
		t.Fatal("could not create new Eleventy template engine:", err)
	}

	if reflect.DeepEqual(engine, expected) {
		t.Fatalf("expected %v, got %v", expected, engine)
	}

}

func TestEleventyRender(t *testing.T) {
	t.Skip("TODO: implement eleventy render test")
}

func TestExtractFileName(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	basePath := filepath.Dir(file)

	tests := []struct {
		path   string
		writer io.Writer
		want   string
	}{
		{
			path: filepath.Join(basePath, "test-data", "data.html"),
			want: filepath.Join(basePath, "test-data", "data.html"),
		},
		{
			path: "notexist.txt",
			want: "",
		},
	}

	var err error
	for _, tt := range tests {
		testName := tt.path

		if tt.want != "" {
			tt.writer, err = os.Create(tt.path)
			if err != nil {
				t.Fatal("could not create test file:", err)
			}
		} else {
			tt.writer = new(bytes.Buffer)
		}

		t.Run(testName, func(t *testing.T) {
			result := extractFileName(tt.writer)

			if tt.want != result {
				t.Errorf("expected %v, got %v", tt.want, result)
			}
		})
	}

	if err := os.Remove(tests[0].path); err != nil {
		t.Fatal("could not remove test file:", err)
	}

}

func TestCreateDataFile(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	basePath := filepath.Dir(file)

	parentFilePath := filepath.Join(basePath, "test-data", "data.html")
	data := []byte(`{"test": "data"}`)

	expected := filepath.Join(basePath, "test-data", "data.11tydata.json")
	result, err := createDataFile(parentFilePath, data)
	if err != nil {
		t.Fatal("could not create data file:", err)
	}

	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}

	if !utils.IsFileAndExists(result) {
		t.Fatal("data file does not exist")
	}

	if err = os.Remove(result); err != nil {
		t.Fatal("could not remove data file:", err)
	}
}

func TestCreateConfigFile(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	basePath := filepath.Dir(file)

	filePath := filepath.Join(basePath, "test-data", ".eleventy.js")
	cfg := config.Config{OutputDir: filepath.Join(basePath, "test-data")}

	var err error
	if err = createConfigFile(filePath, cfg); err != nil {
		t.Fatal("could not create config file:", err)
	}

	if !utils.IsFileAndExists(filePath) {
		t.Fatal("config file does not exist")
	}

	if err = os.Remove(filePath); err != nil {
		t.Fatal("could not remove config file:", err)
	}
}

func TestFindConfigFile(t *testing.T) {
	rootDir, err := utils.RootDir()
	if err != nil {
		t.Fatal("could not get root directory:", err)
	}

	filePath := filepath.Join(rootDir, ".eleventy.js")
	cfg := config.Config{OutputDir: rootDir}

	if err = createConfigFile(filePath, cfg); err != nil {
		t.Fatal("could not create config file:", err)
	}

	result, err := findConfigFile(cfg)
	if err != nil {
		t.Fatal("could not find config file:", err)
	}

	if result != filePath {
		t.Errorf("expected %s, got %s", filePath, result)
	}

	if err = os.Remove(result); err != nil {
		t.Fatal("could not remove config file:", err)
	}

}
