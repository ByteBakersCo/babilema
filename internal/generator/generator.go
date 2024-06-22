package generator

import (
	"bytes"
	"html/template"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/net/html"

	"github.com/ByteBakersCo/babilema/internal/config"
	"github.com/ByteBakersCo/babilema/internal/parser"
	"github.com/ByteBakersCo/babilema/internal/utils"
)

const maxPreviewLength int = 240

type templateEngine interface {
	Generate(
		templateFilePath string,
		writer io.Writer,
		data interface{},
	) error
}

type templateData struct {
	parser.ParsedIssue
	Header   template.HTML
	Footer   template.HTML
	CSSLinks []string
}

type article struct {
	Image         string
	Title         string
	Author        string
	Preview       template.HTML
	DatePublished time.Time
	URL           string
}

func moveGeneratedFilesToOutputDir(cfg config.Config) error {
	files, err := os.ReadDir(cfg.TempDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		src := filepath.Join(cfg.TempDir, file.Name())
		dest := filepath.Join(cfg.OutputDir, file.Name())

		err = os.Rename(src, dest)
		if err != nil {
			return err
		}
	}

	err = os.RemoveAll(cfg.TempDir)
	if err != nil {
		return err
	}

	return nil
}

func newTemplateEngine(cfg config.Config) (templateEngine, error) {
	switch cfg.TemplateEngine {
	case config.EleventyTemplateEngine:
		return NewEleventyTemplateEngine(cfg)
	default:
		if cfg.TemplateEngine != config.DefaultTemplateEngine &&
			cfg.TemplateEngine != "" {
			log.Printf(
				"Unknown template engine: %s -- using default template engine (html/template)\n",
				cfg.TemplateEngine,
			)
		}
		return NewDefaultTemplateEngine(), nil
	}
}

func extractHTML(
	filePath string,
	engine templateEngine,
	data interface{},
) (template.HTML, error) {
	var buf bytes.Buffer
	err := engine.Generate(filePath, &buf, data)
	if err != nil {
		return "", err
	}

	return template.HTML(buf.String()), nil
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
		return nil, nil
	}

	var cssLinks []string
	err := filepath.Walk(
		cssDir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() && strings.HasSuffix(path, ".css") {
				relativeFilePath, err := utils.RelativeFilePath(path)
				if err != nil {
					return err
				}

				websiteURL, err := url.Parse(cfg.WebsiteURL)
				if err != nil {
					return err
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
		return nil, err
	}

	return cssLinks, nil
}

func generateBlogIndexPage(
	articles []article,
	cfg config.Config,
	testOutputWriter io.Writer, // for testing purposes
) error {
	if len(articles) == 0 {
		return nil
	}

	engine, err := newTemplateEngine(cfg)
	if err != nil {
		return err
	}

	websiteURL, err := url.Parse(cfg.WebsiteURL)
	if err != nil {
		return err
	}

	for i := range articles {
		articles[i].URL = filepath.Join(websiteURL.Path, articles[i].URL)
	}

	data := struct {
		Header   template.HTML
		Footer   template.HTML
		Articles []article
	}{
		Articles: articles,
	}

	// TODO: add possibility to inject custom data to header and footer
	data.Header, err = extractHTML(
		cfg.TemplateHeaderFilePath,
		engine,
		nil,
	)
	if err != nil {
		return err
	}

	data.Footer, err = extractHTML(
		cfg.TemplateFooterFilePath,
		engine,
		nil,
	)
	if err != nil {
		return err
	}

	writer := testOutputWriter
	if writer == nil {
		filename := filepath.Base(cfg.TemplateIndexFilePath)
		path := filepath.Join(cfg.TempDir, filename)

		outputFile, error := os.Create(path)
		if error != nil {
			return error
		}

		defer outputFile.Close()

		writer = outputFile
	}

	log.Println("Generating blog index page...")
	err = engine.Generate(cfg.TemplateIndexFilePath, writer, data)
	if err != nil {
		return err
	}

	return nil
}

func GenerateBlogPosts(
	parsedIssues []parser.ParsedIssue,
	cfg config.Config,
	testOutputWriter io.Writer, // for testing purposes
) error {
	if len(parsedIssues) == 0 {
		return nil
	}

	var err error
	engine, err := newTemplateEngine(cfg)
	if err != nil {
		return err
	}

	data := templateData{}
	data.CSSLinks, err = extractCSSLinks(cfg.CSSDir, cfg)
	if err != nil {
		return err
	}

	// TODO: add possibility to inject custom data to header and footer
	data.Header, err = extractHTML(
		cfg.TemplateHeaderFilePath,
		engine,
		nil,
	)
	if err != nil {
		return err
	}

	data.Footer, err = extractHTML(
		cfg.TemplateFooterFilePath,
		engine,
		nil,
	)
	if err != nil {
		return err
	}

	var articles []article
	for _, issue := range parsedIssues {
		data.ParsedIssue = issue
		writer := testOutputWriter
		filename := issue.Metadata.Slug + ".html"
		path := filepath.Join(cfg.TempDir, filename)

		if writer == nil {
			outputFile, error := os.Create(path)
			if error != nil {
				return error
			}
			defer outputFile.Close()

			writer = outputFile

			var articleURL string
			articleURL, err = utils.RelativeFilePath(
				filepath.Join(cfg.OutputDir, filename),
			)
			if err != nil {
				return err
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
		err = engine.Generate(cfg.TemplatePostFilePath, writer, data)
		if err != nil {
			return err
		}
	}

	if len(articles) > 0 {
		err = generateBlogIndexPage(articles, cfg, testOutputWriter)
		if err != nil {
			return err
		}
	}

	isTest := testOutputWriter != nil
	if !isTest {
		err = moveGeneratedFilesToOutputDir(cfg)
		if err != nil {
			return err
		}
	}

	return nil
}
