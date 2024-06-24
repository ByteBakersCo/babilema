package generator

import (
	"bytes"
	"html/template"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/ByteBakersCo/babilema/internal/config"
	"github.com/ByteBakersCo/babilema/internal/parser"
	"github.com/ByteBakersCo/babilema/internal/utils"
)

func normalize(s string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
	}
	return strings.Join(lines, "\n")
}

func TestNewTemplateGenerator(t *testing.T) {
	tests := []struct {
		templateRenderer string
		want             templateRenderer
	}{
		{templateRenderer: "notexisting", want: defaultTemplateRenderer{}},
		{templateRenderer: "", want: defaultTemplateRenderer{}},
		{templateRenderer: "default", want: defaultTemplateRenderer{}},
		{templateRenderer: "Default", want: defaultTemplateRenderer{}},
		{templateRenderer: "eleventy", want: eleventyTemplateRenderer{}},
		{templateRenderer: "Eleventy", want: eleventyTemplateRenderer{}},
	}

	rootDir, err := utils.RootDir()
	if err != nil {
		t.Fatal("could not get root dir:", err)
	}

	for _, tt := range tests {
		cfg := config.Config{
			TemplateRenderer: config.TemplateRenderer(tt.templateRenderer),
			OutputDir:        filepath.Join(rootDir, "test-data"),
		}
		testname := string(cfg.TemplateRenderer)

		t.Run(testname, func(t *testing.T) {
			renderer, err := newTemplateRenderer(cfg)
			if err != nil {
				t.Errorf("could not create new template renderer: %s", err)
			}

			if reflect.TypeOf(renderer) != reflect.TypeOf(tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, renderer)
			}
		})
	}
}

func TestGenerateBlogPosts(t *testing.T) {
	parsedFiles := []parser.ParsedIssue{
		{
			Metadata: parser.Metadata{
				Title:     "Test Title",
				BlogTitle: "Website name",
			},
			Content: template.HTML("<h1>Test HTML</h1>"),
		},
	}

	_, file, _, _ := runtime.Caller(0)
	basePath := filepath.Dir(file)
	var buf bytes.Buffer
	err := GenerateBlogPosts(
		parsedFiles,
		config.Config{
			TemplateRenderer: config.DefaultTemplateRenderer,
			TemplatePostFilePath: filepath.Join(
				basePath,
				"test-data",
				"post.html",
			),
			TemplateHeaderFilePath: filepath.Join(
				basePath,
				"test-data",
				"header.html",
			),
			TemplateFooterFilePath: filepath.Join(
				basePath,
				"test-data",
				"footer.html",
			),
			TemplateIndexFilePath: filepath.Join(
				basePath,
				"test-data",
				"index.html",
			),
			OutputDir:           filepath.Join(basePath, "test-data"),
			CSSDir:              filepath.Join(basePath, "test-data"),
			TempDir:             filepath.Join(basePath, "test-data", "tmp"),
			BlogPostIssuePrefix: "[BLOG]",
			WebsiteURL:          "http://localhost:8080/foo",
		},
		&buf,
	)
	if err != nil {
		t.Errorf("failed to generate blog post: %s", err)
	}

	output := buf.String()

	expectedOutput := `<head>
		<title>Test Title - Website name</title>


		<link rel="stylesheet" type="text/css" href="/foo/internal/generator/test-data/css/bar.css">

	<link rel="stylesheet" type="text/css" href="/foo/internal/generator/test-data/foo.css">


	</head>

	<body>
		<header><div>Test Header</div>
	</header>
		<h1>Test HTML</h1>
		<footer><div>Test Footer</div>
	</footer>
	</body>
`
	if normalize(output) != normalize(expectedOutput) {
		t.Errorf(
			"Expected output to be '%s', got '%s'",
			normalize(expectedOutput),
			normalize(output),
		)
	}
}

// TODO(test): merge the posts generation tests
func TestGenerateBlogPostsWithEleventy(t *testing.T) {
	defer mustCleanupEleventyConfigFile(t)

	parsedFiles := []parser.ParsedIssue{
		{
			Metadata: parser.Metadata{
				Title:     "Test Title",
				BlogTitle: "Website name",
			},
			Content: template.HTML("<h1>Test HTML</h1>"),
		},
	}

	_, file, _, _ := runtime.Caller(0)
	basePath := filepath.Dir(file)
	var buf bytes.Buffer
	err := GenerateBlogPosts(
		parsedFiles,
		config.Config{
			TemplateRenderer: config.EleventyTemplateRenderer,
			TemplatePostFilePath: filepath.Join(
				basePath,
				"test-data",
				"post.liquid",
			),
			TemplateHeaderFilePath: filepath.Join(
				basePath,
				"test-data",
				"header.html",
			),
			TemplateFooterFilePath: filepath.Join(
				basePath,
				"test-data",
				"footer.html",
			),
			TemplateIndexFilePath: filepath.Join(
				basePath,
				"test-data",
				"index.liquid",
			),
			OutputDir:           filepath.Join(basePath, "test-data"),
			CSSDir:              filepath.Join(basePath, "test-data"),
			TempDir:             filepath.Join(basePath, "test-data", "tmp"),
			BlogPostIssuePrefix: "[BLOG]",
			WebsiteURL:          "http://localhost:8080/foo",
		},
		&buf,
	)
	if err != nil {
		t.Errorf("failed to generate blog post: %s", err)
	}

	output := buf.String()

	expectedOutput := `<head>
		<title>Test Title - Website name</title>


		<link rel="stylesheet" type="text/css" href="/foo/internal/generator/test-data/css/bar.css">

	<link rel="stylesheet" type="text/css" href="/foo/internal/generator/test-data/foo.css">


	</head>

	<body>
		<header><div>Test Header</div>
	</header>
		<h1>Test HTML</h1>
		<footer><div>Test Footer</div>
	</footer>
	</body>
`
	if normalize(output) != normalize(expectedOutput) {
		t.Errorf(
			"Expected output to be '%s', got '%s'",
			normalize(expectedOutput),
			normalize(output),
		)
	}
}

func TestExtractPlainText(t *testing.T) {
	expected := "This is a test with an image"
	actual := extractPlainText(template.HTML(`<b>
This is a test with an image <img href="foo.jpg" alt="an image" />
</b>`))

	if actual != expected {
		t.Errorf("Expected '%s', got '%s'", expected, actual)
	}
}

func TestGenerateBlogIndexPage(t *testing.T) {
	articles := []article{
		{
			Image:   filepath.Join("test-data", "image.jpg"),
			Author:  "Test Author",
			Preview: "Test preview",
			Title:   "Test Title 1",
			DatePublished: parser.FormatTime(
				time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
				config.DefaultDateLayout,
			),
			URL: "bar/baz.html",
		},
		{
			Author:  "Test Author",
			Preview: "Test preview without an image",
			Title:   "Test Title 2",
			DatePublished: parser.FormatTime(
				time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
				config.DefaultDateLayout,
			),
			URL: "bar/qux.html",
		},
	}

	_, file, _, _ := runtime.Caller(0)
	basePath := filepath.Dir(file)
	var buf bytes.Buffer
	err := generateBlogIndexPage(
		articles,
		config.Config{
			TemplateIndexFilePath: filepath.Join(
				basePath,
				"test-data",
				"index.html",
			),
			TemplateHeaderFilePath: filepath.Join(
				basePath,
				"test-data",
				"header.html",
			),
			TemplateFooterFilePath: filepath.Join(
				basePath,
				"test-data",
				"footer.html",
			),
			OutputDir:           filepath.Join(basePath, "test-data"),
			WebsiteURL:          "https://localhost:8080/foo",
			TempDir:             ".",
			DateLayout:          "2006-01-02 15:04:05 UTC",
			TemplateRenderer:    config.DefaultTemplateRenderer,
			BlogTitle:           "foo",
			BlogPostIssuePrefix: "bar",
			CSSDir:              filepath.Join(basePath, "test-data", "css"),
		},
		&buf,
	)
	if err != nil {
		t.Errorf("failed to generate blog post: %s", err)
	}

	output := buf.String()

	expectedOutput := `<html>
        
        <body>
        <header><div>Test Header</div>
        </header>
        
        <article>
        <a href="/foo/bar/baz.html">
        <h1>Test Title 1</h1>
        </a>
        <p>Test preview</p>
        <p>Author: Test Author</p>
        <p>Published: 1970-01-01 00:00:00 UTC</p>
        <a href="/foo/bar/baz.html"><img src="test-data/image.jpg" alt="Test Title 1" /></a>
        </article>
        
        <article>
        <a href="/foo/bar/qux.html">
        <h1>Test Title 2</h1>
        </a>
        <p>Test preview without an image</p>
        <p>Author: Test Author</p>
        <p>Published: 1970-01-01 00:00:00 UTC</p>
        <a href="/foo/bar/qux.html"><img src="" alt="Test Title 2" /></a>
        </article>
        
        <footer><div>Test Footer</div>
        </footer>
        </body>
        
        </html>
		`
	if normalize(output) != normalize(expectedOutput) {
		t.Errorf(
			"Expected output to be '%s', got '%s'",
			normalize(expectedOutput),
			normalize(output),
		)
	}
}

// TODO(test): merge the index generation tests
func TestGenerateBlogIndexPageWithEleventy(t *testing.T) {
	defer mustCleanupEleventyConfigFile(t)

	articles := []article{
		{
			Image:   filepath.Join("test-data", "image.jpg"),
			Author:  "Test Author",
			Preview: "Test preview",
			Title:   "Test Title 1",
			DatePublished: parser.FormatTime(
				time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
				config.DefaultDateLayout,
			),
			URL: "bar/baz.html",
		},
		{
			Author:  "Test Author",
			Preview: "Test preview without an image",
			Title:   "Test Title 2",
			DatePublished: parser.FormatTime(
				time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
				config.DefaultDateLayout,
			),
			URL: "bar/qux.html",
		},
	}

	_, file, _, _ := runtime.Caller(0)
	basePath := filepath.Dir(file)
	var buf bytes.Buffer
	err := generateBlogIndexPage(
		articles,
		config.Config{
			TemplateRenderer: config.EleventyTemplateRenderer,
			TemplateIndexFilePath: filepath.Join(
				basePath,
				"test-data",
				"index.liquid",
			),
			TemplateHeaderFilePath: filepath.Join(
				basePath,
				"test-data",
				"header.html",
			),
			TemplateFooterFilePath: filepath.Join(
				basePath,
				"test-data",
				"footer.html",
			),
			OutputDir:           filepath.Join(basePath, "test-data"),
			WebsiteURL:          "https://localhost:8080/foo",
			TempDir:             ".",
			DateLayout:          "2006-01-02 15:04:05 UTC",
			BlogTitle:           "foo",
			BlogPostIssuePrefix: "bar",
			CSSDir:              filepath.Join(basePath, "test-data", "css"),
		},
		&buf,
	)
	if err != nil {
		t.Errorf("failed to generate blog post: %s", err)
	}

	output := buf.String()

	expectedOutput := `<html>
        
        <body>
        <header><div>Test Header</div>
        </header>
        
        <article>
        <a href="/foo/bar/baz.html">
        <h1>Test Title 1</h1>
        </a>
        <p>Test preview</p>
        <p>Author: Test Author</p>
        <p>Published: 1970-01-01 00:00:00 UTC</p>
        <a href="/foo/bar/baz.html"><img src="test-data/image.jpg" alt="Test Title 1" /></a>
        </article>
        
        <article>
        <a href="/foo/bar/qux.html">
        <h1>Test Title 2</h1>
        </a>
        <p>Test preview without an image</p>
        <p>Author: Test Author</p>
        <p>Published: 1970-01-01 00:00:00 UTC</p>
        <a href="/foo/bar/qux.html"><img src="" alt="Test Title 2" /></a>
        </article>
        
        <footer><div>Test Footer</div>
        </footer>
        </body>
        
        </html>
		`
	if normalize(output) != normalize(expectedOutput) {
		t.Errorf(
			"Expected output to be '%s', got '%s'",
			normalize(expectedOutput),
			normalize(output),
		)
	}
}

func mustCleanupEleventyConfigFile(t *testing.T) {
	rootDir, err := utils.RootDir()
	if err != nil {
		t.Errorf("[CLEANUP] failed to get root dir: %s", err)
	}

	err = os.Remove(filepath.Join(rootDir, defaultEleventyConfigFileName))
	if err != nil {
		t.Errorf("[CLEANUP] failed to remove eleventy config file: %s", err)
	}
}
