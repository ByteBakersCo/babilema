package generator

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/net/html"

	"github.com/ByteBakersCo/babilema/internal/config"
	"github.com/ByteBakersCo/babilema/internal/parser"
	"github.com/ByteBakersCo/babilema/internal/utils/copyutils"
	"github.com/ByteBakersCo/babilema/internal/utils/pathutils"
)

const maxPreviewLength int = 240

type templateRenderer interface {
	Render(
		templateFilePath string,
		writer io.Writer,
		data any,
	) error
}

func newTemplateRenderer(cfg config.Config) (templateRenderer, error) {
	cfg.TemplateRenderer = config.TemplateRenderer(
		strings.ToLower(string(cfg.TemplateRenderer)),
	)

	switch cfg.TemplateRenderer {
	case config.EleventyTemplateRenderer:
		return NewEleventyTemplateRenderer(cfg)
	default:
		if cfg.TemplateRenderer != config.DefaultTemplateRenderer &&
			cfg.TemplateRenderer != "" {
			log.Printf(
				"Unknown template engine: %s -- using default template engine (html/template)\n",
				cfg.TemplateRenderer,
			)
		}
		return NewDefaultTemplateRenderer(), nil
	}
}

type templateData struct {
	Header   template.HTML
	Footer   template.HTML
	CSSLinks []string
}

type postTemplateData struct {
	templateData
	parser.ParsedIssue
}

type indexTemplateData struct {
	templateData
	Articles []article
}

type article struct {
	Image         string
	Title         string
	Author        string
	Preview       template.HTML
	DatePublished string
	URL           string
}

func moveGeneratedFilesToOutputDir(cfg config.Config) error {
	err := copyutils.CopyDir(cfg.TempDir, cfg.OutputDir)
	if err != nil {
		return fmt.Errorf("moveGeneratedFilesToOutputDir(): %w", err)
	}

	return os.RemoveAll(cfg.TempDir)
}

func extractPlainText(content template.HTML) string {
	htmlStr := string(content)
	doc, _ := html.Parse(strings.NewReader(htmlStr))

	var text string
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.TextNode {
			text += n.Data
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}

	f(doc)

	result := strings.Join(strings.Fields(text), " ")
	if len(result) > maxPreviewLength {
		result = result[:maxPreviewLength] + "..."
	}

	return result
}

func extractCSSLinks(cssDir string, cfg config.Config) ([]string, error) {
	if cssDir == "" {
		return nil, errors.New(
			"extractCSSLinks(%q, cfg): css directory not set",
		)
	}

	if cfg.WebsiteURL == "" {
		return nil, errors.New(
			"extractCSSLinks(%q, cfg): cfg.WebsiteURL not set",
		)
	}

	var cssLinks []string
	err := filepath.WalkDir(
		cssDir,
		func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return fmt.Errorf(
					"extractCSSLinks(%q, cfg) - WalkDir(%q, %q): %w",
					cssDir,
					path,
					entry,
					err,
				)
			}

			info, err := entry.Info()
			if err != nil {
				return fmt.Errorf(
					"extractCSSLinks(%q, cfg) - WalkDir(%q, %q): %w",
					cssDir,
					path,
					entry,
					err,
				)
			}

			if !info.IsDir() && strings.HasSuffix(path, ".css") {
				relativeFilePath, err := pathutils.RelativeFilePath(path)
				if err != nil {
					return fmt.Errorf(
						"extractCSSLinks(%q, cfg) - WalkDir(%q, %q): %w",
						cssDir,
						path,
						entry,
						err,
					)
				}

				websiteURL, err := url.Parse(cfg.WebsiteURL)
				if err != nil {
					return fmt.Errorf(
						"extractCSSLinks(%q, cfg) - WalkDir(%q, %q): %w",
						cssDir,
						path,
						entry,
						err,
					)
				}

				relativeFilePath = filepath.Join(
					websiteURL.Path,
					relativeFilePath,
				)

				cssLinks = append(cssLinks, relativeFilePath)
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("extractCSSLinks(%q, cfg): %w", cssDir, err)
	}

	return cssLinks, nil
}

func populateTemplateData(
	renderer templateRenderer,
	cfg config.Config,
	headerData any,
	footerData interface{},
) (templateData, error) {
	var err error
	data := templateData{}

	data.CSSLinks, err = extractCSSLinks(cfg.CSSDir, cfg)
	if err != nil {
		return templateData{}, fmt.Errorf("populateTemplateData(): %w", err)
	}

	var header bytes.Buffer
	err = renderer.Render(cfg.TemplateHeaderFilePath, &header, headerData)
	if err != nil {
		return templateData{}, fmt.Errorf("populateTemplateData(): %w", err)
	}

	var footer bytes.Buffer
	err = renderer.Render(cfg.TemplateFooterFilePath, &footer, footerData)
	if err != nil {
		return templateData{}, fmt.Errorf("populateTemplateData(): %w", err)
	}

	data.Header = template.HTML(header.String())
	data.Footer = template.HTML(footer.String())

	return data, nil
}

// TODO: add possibility to inject custom data to header and footer
func generateBlogIndexPage(
	articles []article,
	cfg config.Config,
	testOutputWriter io.Writer, // for testing purposes
) error {
	if len(articles) == 0 {
		return nil
	}

	if cfg.WebsiteURL == "" {
		return errors.New("generateBlogIndexPage(): cfg.WebsiteURL not set")
	}

	if cfg.TemplateIndexFilePath == "" {
		return errors.New(
			"generateBlogIndexPage(): cfg.TemplateIndexFilePath path not set",
		)

	}

	if cfg.TempDir == "" {
		return errors.New("generateBlogIndexPage(): cfg.TempDir not set")
	}

	engine, err := newTemplateRenderer(cfg)
	if err != nil {
		return fmt.Errorf("generateBlogIndexPage(): %w", err)
	}

	websiteURL, err := url.Parse(cfg.WebsiteURL)
	if err != nil {
		return fmt.Errorf("generateBlogIndexPage(): %w", err)
	}

	for i := range articles {
		articles[i].URL = filepath.Join(websiteURL.Path, articles[i].URL)
	}

	data := indexTemplateData{Articles: articles}
	data.templateData, err = populateTemplateData(engine, cfg, nil, nil)
	if err != nil {
		return fmt.Errorf("generateBlogIndexPage(): %w", err)
	}

	writer := testOutputWriter
	if writer == nil {
		filename := filepath.Base(cfg.TemplateIndexFilePath)
		path := filepath.Join(cfg.TempDir, filename)

		outputFile, error := os.Create(path)
		if error != nil {
			return fmt.Errorf("generateBlogIndexPage(): %w", error)
		}

		defer func() {
			if cerr := outputFile.Close(); error == nil && err == nil {
				err = cerr
			}
		}()

		writer = outputFile
	}

	log.Println("Generating blog index page...")
	err = engine.Render(cfg.TemplateIndexFilePath, writer, data)
	if err != nil {
		return fmt.Errorf("generateBlogIndexPage(): %w", err)
	}

	return nil
}

// TODO: add possibility to inject custom data to header and footer
func GenerateBlogPosts(
	parsedIssues []parser.ParsedIssue,
	cfg config.Config,
	testOutputWriter io.Writer, // for testing purposes
) error {
	if len(parsedIssues) == 0 {
		return nil
	}

	if cfg.TempDir == "" {
		return errors.New("GenerateBlogPosts(): cfg.TempDir not set")
	}

	if cfg.OutputDir == "" {
		return errors.New("GeneratePlogPosts(): cfg.OutputDir not set")
	}

	var err error
	engine, err := newTemplateRenderer(cfg)
	if err != nil {
		return fmt.Errorf("GenerateBlogPosts(): %w", err)
	}

	var data postTemplateData
	data.templateData, err = populateTemplateData(engine, cfg, nil, nil)
	if err != nil {
		return fmt.Errorf("GenerateBlogPosts(): %w", err)
	}

	var articles []article
	for _, issue := range parsedIssues {
		data.ParsedIssue = issue
		writer := testOutputWriter

		if writer == nil {
			outputFilePath := filepath.Join(
				cfg.TempDir,
				issue.Metadata.Slug,
				"index.html",
			)
			relativePath := filepath.Join(
				cfg.OutputDir,
				issue.Metadata.Slug,
				"index.html",
			)

			outputFile, error := os.Create(outputFilePath)
			if error != nil {
				return fmt.Errorf("GenerateBlogPosts(): %w", error)
			}

			defer func() {
				if cerr := outputFile.Close(); error == nil && err == nil {
					err = cerr
				}
			}()

			writer = outputFile

			var articleURL string
			articleURL, err = pathutils.RelativeFilePath(relativePath)
			if err != nil {
				return fmt.Errorf("GenerateBlogPosts(): %w", err)
			}

			articles = append(articles, article{
				Image:         data.Metadata.Image,
				Title:         data.Metadata.Title,
				Author:        data.Metadata.Author,
				Preview:       template.HTML(extractPlainText(data.Content)),
				DatePublished: data.Metadata.DatePublished,
				URL:           articleURL,
			})
		}

		log.Println("Generating blog post:", data.Metadata.Slug)
		err = engine.Render(cfg.TemplatePostFilePath, writer, data)
		if err != nil {
			return fmt.Errorf("GenerateBlogPosts(): %w", err)
		}
	}

	if len(articles) > 0 {
		err = generateBlogIndexPage(articles, cfg, testOutputWriter)
		if err != nil {
			return fmt.Errorf("GenerateBlogPosts(): %w", err)
		}
	}

	isTest := testOutputWriter != nil
	if !isTest {
		err = moveGeneratedFilesToOutputDir(cfg)
		if err != nil {
			return fmt.Errorf("GenerateBlogPosts(): %w", err)
		}
	}

	return nil
}
