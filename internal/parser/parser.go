package parser

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/gomarkdown/markdown"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"

	"github.com/ByteBakersCo/babilema/internal/config"
	"github.com/ByteBakersCo/babilema/internal/history"
)

type Metadata struct {
	// Determined at runtime
	DatePublished string
	DateModified  string

	// Determined at runtime (WebsiteURL + Slug)
	URL string

	Title       string // required
	Slug        string // required
	BlogTitle   string
	Description string
	Author      string
	Image       string
	Publisher   string
	Keywords    []string
	Tags        []string
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

func FormatTime(t time.Time, layout string) string {
	timezone, _ := t.Zone()
	return (t.Format(layout) + " " + timezone)
}

func checkRequiredMetadata(metadata Metadata) error {
	missingFields := []string{}
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
	if cfg.DateLayout == "" {
		return Metadata{}, errors.New("date layout not set")
	}

	if cfg.WebsiteURL == "" {
		return Metadata{}, errors.New("website URL not set")
	}

	if cfg.BlogTitle == "" {
		return Metadata{}, errors.New("blog title not set")
	}

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

	err = checkRequiredMetadata(metadata)
	if err != nil {
		return Metadata{}, err
	}

	metadata.DatePublished = FormatTime(issue.GetCreatedAt(), cfg.DateLayout)
	metadata.DateModified = FormatTime(issue.GetUpdatedAt(), cfg.DateLayout)

	// Since the actual url is {domain}/slug/index.html we need to remove the file extension
	metadata.Slug = strings.TrimSuffix(
		metadata.Slug,
		filepath.Ext(metadata.Slug),
	)
	// This is important, the user cannot set a different URL on their post
	metadata.URL = fmt.Sprintf("%s/%s/", cfg.WebsiteURL, metadata.Slug)

	if metadata.BlogTitle == "" {
		metadata.BlogTitle = cfg.BlogTitle
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
	if cfg.TempDir == "" {
		return nil, errors.New("temp dir not set")
	}

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

	var permissionLevel *github.RepositoryPermissionLevel
	var metadata Metadata
	var content []byte
	var parsedIssues []ParsedIssue
	for _, issue := range issues {
		if cfg.BlogPostIssuePrefix != "" &&
			!strings.HasPrefix(issue.GetTitle(), cfg.BlogPostIssuePrefix) {
			continue
		}

		permissionLevel, _, err = client.Repositories.GetPermissionLevel(
			ctx,
			owner,
			repo,
			issue.GetUser().GetLogin(),
		)
		if err != nil {
			return nil, err
		}

		hasWritePermission := permissionLevel.GetPermission() == "write" ||
			permissionLevel.GetPermission() == "admin"

		if !hasWritePermission {
			continue
		}

		metadata, err = extractMetadata(*issue, cfg)
		if err != nil {
			return nil, err
		}

		if _, ok := postsHistory[metadata.Slug]; ok {
			isUpdated := issue.GetUpdatedAt().After(postsHistory[metadata.Slug])
			if !isUpdated {
				continue
			}
		}

		postsHistory[metadata.Slug] = issue.GetUpdatedAt()

		content, err = extractMarkdown([]byte(issue.GetBody()))
		if err != nil {
			return nil, err
		}

		content = markdown.ToHTML(content, nil, nil)

		parsedIssues = append(parsedIssues, ParsedIssue{
			Content:  template.HTML(content),
			Metadata: metadata,
		})
	}

	log.Printf("Found %d blog posts to generate.\n", len(parsedIssues))

	if len(parsedIssues) > 0 {
		err = os.MkdirAll(cfg.TempDir, os.ModePerm)
		if err != nil {
			return nil, err
		}

		err = history.UpdateHistoryFile(postsHistory, cfg)
		if err != nil {
			return nil, err
		}
	}

	return parsedIssues, nil
}
