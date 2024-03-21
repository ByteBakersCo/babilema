package generator

import (
	"bytes"
	"html/template"
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
				Title:        "Test Title",
				PageSubtitle: "Website name",
			},
			Content: template.HTML("<h1>Test HTML</h1>"),
		},
	}

	var buf bytes.Buffer
	err := GenerateBlogPosts(
		parsedFiles,
		config.Config{
			TemplatePostFilePath:   "./test-data/post.html",
			TemplateHeaderFilePath: "./test-data/header.html",
			TemplateFooterFilePath: "./test-data/footer.html",
			TemplateIndexFilePath:  "./test-data/index.html",
			OutputDir:              "./test-data",
			CSSDir:                 "./test-data",
			BlogPostIssuePrefix:    "[BLOG]",
			WebsiteURL:             "http://localhost:8080",
		},
		&buf,
	)
	if err != nil {
		t.Fatalf("failed to generate blog post: %s", err)
	}

	output := buf.String()

	expectedOutput := `<head>
		<title>Test Title - Website name</title>


		<link rel="stylesheet" type="text/css" href="test-data/css/bar.css">

	<link rel="stylesheet" type="text/css" href="test-data/foo.css">


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
			Image:         "test-data/image.jpg",
			Author:        "Test Author",
			Preview:       "Test preview",
			Title:         "Test Title 1",
			DatePublished: time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
			URL:           "example.com",
		},
		{
			Author:        "Test Author",
			Preview:       "Test preview without an image",
			Title:         "Test Title 2",
			DatePublished: time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
			URL:           "example.com",
		},
	}

	var buf bytes.Buffer
	err := generateBlogIndexPage(
		articles,
		config.Config{
			TemplateIndexFilePath: "./test-data/index.html",
			OutputDir:             "./test-data",
		},
		&buf,
	)
	if err != nil {
		t.Fatalf("failed to generate blog post: %s", err)
	}

	output := buf.String()

	expectedOutput := `<html>
        
        <body>
        
        <article>
        <a href="example.com">
        <h1>Test Title 1</h1>
        </a>
        <p>Test preview</p>
        <p>Author: Test Author</p>
        <p>Published: 1970-01-01 00:00:00 &#43;0000 UTC</p>
        <a href="example.com"><img src="test-data/image.jpg" alt="Test Title 1" /></a>
        </article>
        
        <article>
        <a href="example.com">
        <h1>Test Title 2</h1>
        </a>
        <p>Test preview without an image</p>
        <p>Author: Test Author</p>
        <p>Published: 1970-01-01 00:00:00 &#43;0000 UTC</p>
        <a href="example.com"><img src="" alt="Test Title 2" /></a>
        </article>
        
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
