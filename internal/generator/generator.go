package generator

import (
	"bytes"
	"errors"
	"html/template"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/net/html"

	"github.com/ByteBakersCo/babilema/internal/config"
	"github.com/ByteBakersCo/babilema/internal/parser"
	"github.com/ByteBakersCo/babilema/internal/utils"
)

const maxPreviewLength int = 240

type templateRenderer interface {
	Render(
		templateFilePath string,
		writer io.Writer,
		data interface{},
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

// TODO(moveGeneratedFilesToOutputDir):
// If output directory exists, backup the contents
// If output directory does not exist, create it
// If error occurs, delete generated files, restore backup, and return error
// Make sure no file is deleted before everything is moved (including temp directory)
func moveGeneratedFilesToOutputDir(cfg config.Config) error {
	if cfg.TempDir == "" {
		return errors.New("temp dir not set")
	}

	if cfg.OutputDir == "" {
		return errors.New("output dir not set")
	}

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
		return nil, errors.New("css directory not set")
	}

	if cfg.WebsiteURL == "" {
		return nil, errors.New("website URL not set")
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

func populateTemplateData(
	renderer templateRenderer,
	cfg config.Config,
	headerData interface{},
	footerData interface{},
) (templateData, error) {
	var err error
	data := templateData{}

	data.CSSLinks, err = extractCSSLinks(cfg.CSSDir, cfg)
	if err != nil {
		return templateData{}, err
	}

	var header bytes.Buffer
	err = renderer.Render(cfg.TemplateHeaderFilePath, &header, headerData)
	if err != nil {
		return templateData{}, err
	}

	var footer bytes.Buffer
	err = renderer.Render(cfg.TemplateFooterFilePath, &footer, footerData)
	if err != nil {
		return templateData{}, err
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
		return errors.New("website URL not set")
	}

	if cfg.TemplateIndexFilePath == "" {
		return errors.New("index template file path not set")

	}

	if cfg.TempDir == "" {
		return errors.New("temp dir not set")
	}

	engine, err := newTemplateRenderer(cfg)
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

	data := indexTemplateData{Articles: articles}
	data.templateData, err = populateTemplateData(engine, cfg, nil, nil)
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
		return err
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
		return errors.New("temp dir not set")
	}

	if cfg.OutputDir == "" {
		return errors.New("output dir not set")
	}

	var err error
	engine, err := newTemplateRenderer(cfg)
	if err != nil {
		return err
	}

	var data postTemplateData
	data.templateData, err = populateTemplateData(engine, cfg, nil, nil)
	if err != nil {
		return err
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
				return error
			}

			defer func() {
				if cerr := outputFile.Close(); error == nil && err == nil {
					err = cerr
				}
			}()

			writer = outputFile

			var articleURL string
			articleURL, err = utils.RelativeFilePath(relativePath)
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
		err = engine.Render(cfg.TemplatePostFilePath, writer, data)
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
