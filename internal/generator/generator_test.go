package generator

import (
	"bytes"
	"html/template"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/ByteBakersCo/babilema/internal/config"
	"github.com/ByteBakersCo/babilema/internal/parser"
)

func normalize(s string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
	}
	return strings.Join(lines, "\n")
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
		t.Fatalf("failed to generate blog post: %s", err)
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
			Image:         filepath.Join("test-data", "image.jpg"),
			Author:        "Test Author",
			Preview:       "Test preview",
			Title:         "Test Title 1",
			DatePublished: time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
			URL:           "bar/baz.html",
		},
		{
			Author:        "Test Author",
			Preview:       "Test preview without an image",
			Title:         "Test Title 2",
			DatePublished: time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
			URL:           "bar/qux.html",
		},
	}

	var buf bytes.Buffer
	err := generateBlogIndexPage(
		articles,
		config.Config{
			TemplateIndexFilePath: filepath.Join(
				".",
				"test-data",
				"index.html",
			),
			TemplateHeaderFilePath: filepath.Join(
				".",
				"test-data",
				"header.html",
			),
			TemplateFooterFilePath: filepath.Join(
				".",
				"test-data",
				"footer.html",
			),
			OutputDir:  filepath.Join(".", "test-data"),
			WebsiteURL: "https://localhost:8080/foo",
		},
		&buf,
	)
	if err != nil {
		t.Fatalf("failed to generate blog post: %s", err)
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
        <p>Published: 1970-01-01 00:00:00 &#43;0000 UTC</p>
        <a href="/foo/bar/baz.html"><img src="test-data/image.jpg" alt="Test Title 1" /></a>
        </article>
        
        <article>
        <a href="/foo/bar/qux.html">
        <h1>Test Title 2</h1>
        </a>
        <p>Test preview without an image</p>
        <p>Author: Test Author</p>
        <p>Published: 1970-01-01 00:00:00 &#43;0000 UTC</p>
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
