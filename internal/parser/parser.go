package parser

import (
	"context"
	"errors"
	"html/template"
	"log"
	"os"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/ByteBakersCo/babilema/internal/config"
	"github.com/ByteBakersCo/babilema/internal/history"
	"github.com/gomarkdown/markdown"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type Metadata struct {
	Title        string // required
	Slug         string // required
	PageSubtitle string
	Description  string
	Keywords     []string
	Author       string
	Image        string
	Publisher    string
	Tags         []string

	// Determined from configuration file
	URL string

	// Determined at runtime
	DatePublished time.Time
	DateModified  time.Time
}

type ParsedIssue struct {
	Content  template.HTML
	Metadata Metadata
}

func trimAllSpaces(array []string) []string {
	cleanSlice := make([]string, 0, len(array))

	for _, str := range array {
		str = strings.TrimSpace(str)
		cleanSlice = append(cleanSlice, str)
	}

	return cleanSlice
}

func checkRequiredMetadata(metadata Metadata) error {
	missingFields := []string{}
	if metadata.URL == "" {
		missingFields = append(missingFields, "URL")
	}

	if metadata.Slug == "" {
		missingFields = append(missingFields, "Slug")
	}

	if metadata.Title == "" {
		missingFields = append(missingFields, "Title")
	}

	if len(missingFields) > 0 {
		msg := "missing required metadata fields: " + strings.Join(
			missingFields,
			", ",
		)
		return errors.New(msg)
	}

	return nil
}

func extractMetadata(issue github.Issue, cfg config.Config) (Metadata, error) {
	content := issue.GetBody()

	lines := strings.Split(content, "\n")
	lines = trimAllSpaces(lines)

	hasMetadataHeader := len(lines) > 3 || lines[0] == "---"

	if !hasMetadataHeader {
		return Metadata{}, errors.New("no TOML header found")
	}

	endOfHeader := 1
	for ; endOfHeader < len(lines) && lines[endOfHeader] != "---"; endOfHeader++ {
	}

	var metadata Metadata
	err := toml.Unmarshal(
		[]byte(strings.Join(lines[1:endOfHeader], "\n")),
		&metadata,
	)
	if err != nil {
		return Metadata{}, err
	}

	createdAt := issue.GetCreatedAt()
	updatedAt := issue.GetUpdatedAt()

	metadata.DatePublished = createdAt
	metadata.DateModified = updatedAt

	if metadata.URL == "" {
		metadata.URL = cfg.WebsiteURL
	}

	err = checkRequiredMetadata(metadata)
	if err != nil {
		return Metadata{}, err
	}

	return metadata, nil
}

func extractMarkdown(content []byte) ([]byte, error) {
	lines := strings.Split(string(content), "\n")
	lines = trimAllSpaces(lines)
	hasMetadataHeader := len(lines) > 3 && lines[0] == "---"

	if !hasMetadataHeader {
		return content, nil
	}

	hasReachedEndOfHeader := false
	endOfHeader := 1
	for {
		if endOfHeader >= len(lines) {
			err := errors.New("no closing --- found in metadata")
			if hasReachedEndOfHeader {
				err = errors.New("blog post has no content")
			}

			return nil, err
		}

		if lines[endOfHeader] == "---" {
			hasReachedEndOfHeader = true
		}

		endOfHeader++

		// Removing extra new lines after metadata header
		if hasReachedEndOfHeader && lines[endOfHeader] != "" {
			break
		}
	}

	return []byte(strings.Join(lines[endOfHeader:], "\n")), nil
}

func ParseIssues(cfg config.Config) ([]ParsedIssue, error) {
	ghToken := os.Getenv("GITHUB_TOKEN")
	if ghToken == "" {
		return nil, errors.New("GITHUB_TOKEN not set")
	}

	ctx := context.Background()
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: ghToken})
	tokenClient := oauth2.NewClient(ctx, tokenSource)
	client := github.NewClient(tokenClient)

	repo := os.Getenv("GITHUB_REPOSITORY")
	if repo == "" {
		return nil, errors.New("GITHUB_REPOSITORY not set")
	}

	parts := strings.Split(repo, "/")
	owner, repo := parts[0], parts[1]

	issues, _, err := client.Issues.ListByRepo(ctx, owner, repo, nil)
	if err != nil {
		return nil, err
	}

	postsHistory, err := history.ParseHistoryFile(cfg)
	if err != nil {
		return nil, err
	}

	var parsedIssues []ParsedIssue
	for _, issue := range issues {
		if !strings.HasPrefix(issue.GetTitle(), cfg.BlogPostIssuePrefix) {
			continue
		}

		permissionLevel, _, err := client.Repositories.GetPermissionLevel(
			ctx,
			owner,
			repo,
			issue.GetUser().GetName(),
		)
		if err != nil {
			return nil, err
		}

		hasWritePermission := permissionLevel.GetPermission() == "write" ||
			permissionLevel.GetPermission() == "admin"

		if !hasWritePermission {
			continue
		}

		metadata, err := extractMetadata(*issue, cfg)
		if err != nil {
			return nil, err
		}

		if _, ok := postsHistory[metadata.Slug]; ok {
			isUpdated := issue.GetUpdatedAt().After(metadata.DateModified)
			if !isUpdated {
				continue
			}

			postsHistory[metadata.Slug] = issue.GetUpdatedAt().
				Format(time.RFC3339)
		}

		content, err := extractMarkdown([]byte(issue.GetBody()))
		if err != nil {
			return nil, err
		}

		content = markdown.ToHTML(content, nil, nil)

		parsedIssues = append(parsedIssues, ParsedIssue{
			Content:  template.HTML(content),
			Metadata: metadata,
		})
	}

	log.Printf("Found %d blog posts to generate\n", len(parsedIssues))

	if len(parsedIssues) > 0 {
		history.UpdateHistoryFile(postsHistory, cfg)
	}

	return parsedIssues, nil
}
