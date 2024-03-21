package parser

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/go-github/github"

	"github.com/ByteBakersCo/babilema/internal/config"
	"github.com/ByteBakersCo/babilema/internal/utils"
)

func mockIssue() github.Issue {
	body := `---
Title = "Test post"
Slug = "test-post"
BlogTitle = "Overwritten Website name"
Description = "This is a test post for Babilema"
Keywords = ["test", "post", "babilema"]
Author = "Babilema team"
Image = "test-post.jpg"
Publisher = "Babilema team"
Logo = "babilema-logo.png" # does not exist
tags = ["test", "post", "babilema"]
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
			"Expected output to be\n %v\n\t------\n\tgot\n %v\n\n\n\n%s\n\t------\n\t%s\n",
			expected,
			actual,
			string(expected),
			string(actual),
		)
	}
}

func TestExtractMetadata(t *testing.T) {
	expected := Metadata{
		Title:         "Test post",
		Slug:          "test-post",
		BlogTitle:     "Overwritten Website name",
		Description:   "This is a test post for Babilema",
		Keywords:      []string{"test", "post", "babilema"},
		Author:        "Babilema team",
		Image:         "test-post.jpg",
		Publisher:     "Babilema team",
		Tags:          []string{"test", "post", "babilema"},
		URL:           "example.com/test-post",
		DatePublished: time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
		DateModified:  time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	actual, err := extractMetadata(
		mockIssue(),
		config.Config{
			BlogTitle:  "This should be ignored",
			WebsiteURL: "example.com",
		},
	)
	if err != nil {
		t.Errorf("extractMetadata failed: %s", err)
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Error(
			utils.FormatStruct(expected, "Expected output to be"),
			utils.FormatStruct(actual, "\ngot"),
		)
	}

	// Sad path
	badActual, err := extractMetadata(badMockIssue(), config.Config{})
	if err == nil {
		t.Error(utils.FormatStruct(badActual, "Expected error, got"))
	}
	missingMetadata := []string{
		"Slug",
		"Title",
	}

	for _, field := range missingMetadata {
		if !strings.Contains(err.Error(), field) {
			t.Errorf("Expected error to contain '%s', got '%s'", field, err)
		}
	}
}
