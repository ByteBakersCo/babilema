package generator

import (
	"bytes"
	"html/template"
	"strings"
	"testing"

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

func TestGenerate(t *testing.T) {
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
