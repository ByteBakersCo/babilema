package generator

import (
	"bytes"
	"html/template"
	"io"
	"log"
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

func extractHTML(filePath string, data interface{}) (template.HTML, error) {
	tmpl, err := template.ParseFiles(filePath)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
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

func extractCSSLinks(cssDir string) ([]string, error) {
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

	data := struct {
		Header   template.HTML
		Footer   template.HTML
		Articles []article
	}{
		Articles: articles,
	}

	indexTemplate, err := template.ParseFiles(cfg.TemplateIndexFilePath)
	if err != nil {
		return err
	}

	// TODO: add possibility to inject custom data to header and footer
	data.Header, err = extractHTML(cfg.TemplateHeaderFilePath, nil)
	if err != nil {
		return err
	}

	data.Footer, err = extractHTML(cfg.TemplateFooterFilePath, nil)
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
	err = indexTemplate.Execute(writer, data)
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

	postTemplate, err := template.ParseFiles(cfg.TemplatePostFilePath)
	if err != nil {
		return err
	}

	data := templateData{}
	data.CSSLinks, err = extractCSSLinks(cfg.CSSDir)
	if err != nil {
		return err
	}

	// TODO: add possibility to inject custom data to header and footer
	data.Header, err = extractHTML(cfg.TemplateHeaderFilePath, nil)
	if err != nil {
		return err
	}

	data.Footer, err = extractHTML(cfg.TemplateFooterFilePath, nil)
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
			err = os.MkdirAll(cfg.TempDir, os.ModePerm)
			if err != nil {
				return err
			}

			outputFile, error := os.Create(path)
			if error != nil {
				return error
			}
			defer outputFile.Close()

			writer = outputFile

			var url string
			url, err = utils.RelativeFilePath(
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
				URL:           url,
			})
		}

		log.Println("Generating blog post:", data.Metadata.Slug)
		err = postTemplate.Execute(writer, data)
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
