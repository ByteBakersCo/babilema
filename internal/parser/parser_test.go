package parser

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/ByteBakersCo/babilema/internal/config"
	"github.com/google/go-github/github"
)

// TODO: need to validate that all the critical metadata is present
// TODO: need to add conditional HTML rendering for metadata, etc

func mockIssue() github.Issue {
	body := `---
Description = "This is a test post for Babilema"
Keywords = ["test", "post", "babilema"]
Author = "Babilema team"
Title = "Test post"
Slug = "test-post"
Image = "test-post.jpg"
Publisher = "Babilema team"
Logo = "babilema-logo.png"
tags = ["test", "post", "babilema"]
URL = "example.com"
---

# Test post

This is a simple test post for Babilema`

	createdAt := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	updatedAt := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)

	return github.Issue{
		Body:      &body,
		CreatedAt: &createdAt,
		UpdatedAt: &updatedAt,
	}
}

func badMockIssue() github.Issue {
	body := `---
	Description = "This is a test post for Babilema"
	---

	# Test post
	`

	return github.Issue{
		Body: &body,
	}
}

func TestExtractMarkdown(t *testing.T) {
	expected := []byte(`# Test post

This is a simple test post for Babilema`)

	actual, err := extractMarkdown([]byte(*mockIssue().Body))
	if err != nil {
		t.Errorf("extractMarkdown failed: %s", err)
	}

	if string(actual) != string(expected) {
		t.Errorf(
			"Expected output to be\n %v\n\t------\n\tgot\n %v",
			(expected),
			(actual),
		)
	}
}

func TestExtractMetadata(t *testing.T) {
	expected := Metadata{
		Description:   "This is a test post for Babilema",
		Keywords:      []string{"test", "post", "babilema"},
		Author:        "Babilema team",
		Title:         "Test post",
		Slug:          "test-post",
		Image:         "test-post.jpg",
		Publisher:     "Babilema team",
		Logo:          "babilema-logo.png",
		Tags:          []string{"test", "post", "babilema"},
		URL:           "example.com",
		DatePublished: time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
		DateModified:  time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	actual, err := extractMetadata(mockIssue(), config.Config{})
	if err != nil {
		t.Errorf("extractMetadata failed: %s", err)
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(
			"Expected output to be %+v\n------\ngot %+v",
			expected,
			actual,
		)
	}

	// Sad path
	badActual, err := extractMetadata(badMockIssue(), config.Config{})
	if err == nil {
		t.Errorf("Expected error, got %+v", badActual)
	}
	missingMetadata := []string{
		"URL",
		"Slug",
		"Title",
	}

	for _, field := range missingMetadata {
		if !strings.Contains(err.Error(), field) {
			t.Errorf("Expected error to contain '%s', got '%s'", field, err)
		}
	}

	badActualWithURL, err := extractMetadata(
		badMockIssue(),
		config.Config{WebsiteURL: "example.com"},
	)
	if err == nil {
		t.Errorf("Expected error, got %+v", badActualWithURL)
	}

	if strings.Contains(err.Error(), "URL") {
		t.Errorf("Expected error NOT to contain 'URL', got '%s'", err)
	}
}
