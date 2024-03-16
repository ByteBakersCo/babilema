package generator

import (
	"bytes"
	"html/template"
	"testing"

	"github.com/ByteBakersCo/babilema/internal/parser"
)

func TestGenerate(t *testing.T) {
	parsedFiles := []parser.ParsedIssue{
		{
			Metadata: parser.Metadata{
				Title: "Test Title",
			},
			Content: template.HTML("<h1>Test HTML</h1>"),
		},
	}

	var buf bytes.Buffer
	err := GenerateBlogPosts(
		parsedFiles,
		"",
		"./test-post.html",
		&buf,
	)
	if err != nil {
		t.Fatalf("failed to generate blog post: %s", err)
	}

	output := buf.String()

	expectedOutput := `<head>
    <title>Test Title</title>
</head>

<body><h1>Test HTML</h1></body>
`
	if output != expectedOutput {
		t.Errorf("Expected output to be '%s', got '%s'", expectedOutput, output)
	}
}
