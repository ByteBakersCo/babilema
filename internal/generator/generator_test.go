package generator

import (
	"bytes"
	"html/template"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/ByteBakersCo/babilema/internal/config"
	"github.com/ByteBakersCo/babilema/internal/parser"
	"github.com/ByteBakersCo/babilema/internal/utils"
)

func normalize(s string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
	}
	return strings.Join(lines, "\n")
}

func TestNewTemplateRenderer(t *testing.T) {
	tests := []struct {
		templateRenderer string
		want             templateRenderer
	}{
		{templateRenderer: "notexisting", want: defaultTemplateRenderer{}},
		{templateRenderer: "", want: defaultTemplateRenderer{}},
		{templateRenderer: "default", want: defaultTemplateRenderer{}},
		{templateRenderer: "Default", want: defaultTemplateRenderer{}},
		{templateRenderer: "eleventy", want: eleventyTemplateRenderer{}},
		{templateRenderer: "Eleventy", want: eleventyTemplateRenderer{}},
	}

	_, file, _, _ := runtime.Caller(0)
	basePath := filepath.Dir(file)
	for _, test := range tests {
		cfg := config.Config{
			TemplateRenderer: config.TemplateRenderer(test.templateRenderer),
			OutputDir:        filepath.Join(basePath, "test-data"),
		}
		testname := string(cfg.TemplateRenderer)

		t.Run(testname, func(t *testing.T) {
			renderer, err := newTemplateRenderer(cfg)
			if err != nil {
				t.Errorf("could not create new template renderer: %v", err)
			}

			if reflect.TypeOf(renderer) != reflect.TypeOf(test.want) {
				t.Fatalf(
					"NewTemplateRenderer() typeOf: %v, want %v",
					renderer,
					test.want,
				)
			}
		})
	}
}

func TestGenerateBlogPosts(t *testing.T) {
	parsedFiles := []parser.ParsedIssue{
		{
			Metadata: parser.Metadata{
				Title:     "Test Title",
				BlogTitle: "Website name",
			},
			Content: template.HTML("<h1>Test HTML</h1>"),
		},
	}

	_, file, _, _ := runtime.Caller(0)
	basePath := filepath.Dir(file)
	var buf bytes.Buffer
	err := GenerateBlogPosts(
		parsedFiles,
		config.Config{
			TemplateRenderer: config.DefaultTemplateRenderer,
			TemplatePostFilePath: filepath.Join(
				basePath,
				"test-data",
				"post.html",
			),
			TemplateHeaderFilePath: filepath.Join(
				basePath,
				"test-data",
				"header.html",
			),
			TemplateFooterFilePath: filepath.Join(
				basePath,
				"test-data",
				"footer.html",
			),
			TemplateIndexFilePath: filepath.Join(
				basePath,
				"test-data",
				"index.html",
			),
			OutputDir:           filepath.Join(basePath, "test-data"),
			CSSDir:              filepath.Join(basePath, "test-data"),
			TempDir:             filepath.Join(basePath, "test-data", "tmp"),
			BlogPostIssuePrefix: "[BLOG]",
			WebsiteURL:          "http://localhost:8080/foo",
		},
		&buf,
	)
	if err != nil {
		t.Errorf("failed to generate blog post: %v", err)
	}

	got := normalize(buf.String())
	want := normalize(`<head>
		<title>Test Title - Website name</title>


		<link rel="stylesheet" type="text/css" href="/foo/internal/generator/test-data/css/bar.css">

	<link rel="stylesheet" type="text/css" href="/foo/internal/generator/test-data/foo.css">


	</head>

	<body>
		<header><div>Test Header</div>
	</header>
		<h1>Test HTML</h1>
		<footer><div>Test Footer</div>
	</footer>
	</body>
`)

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("GenerateBlogPosts(default) mismatch (-want +got):\n%s", diff)
	}
}

// TODO(test): merge the posts generation tests
func TestGenerateBlogPostsWithEleventy(t *testing.T) {
	defer cleanupEleventyConfigFile(t)

	parsedFiles := []parser.ParsedIssue{
		{
			Metadata: parser.Metadata{
				Title:     "Test Title",
				BlogTitle: "Website name",
			},
			Content: template.HTML("<h1>Test HTML</h1>"),
		},
	}

	_, file, _, _ := runtime.Caller(0)
	basePath := filepath.Dir(file)
	var buf bytes.Buffer
	err := GenerateBlogPosts(
		parsedFiles,
		config.Config{
			TemplateRenderer: config.EleventyTemplateRenderer,
			TemplatePostFilePath: filepath.Join(
				basePath,
				"test-data",
				"post.liquid",
			),
			TemplateHeaderFilePath: filepath.Join(
				basePath,
				"test-data",
				"header.html",
			),
			TemplateFooterFilePath: filepath.Join(
				basePath,
				"test-data",
				"footer.html",
			),
			TemplateIndexFilePath: filepath.Join(
				basePath,
				"test-data",
				"index.liquid",
			),
			OutputDir:           filepath.Join(basePath, "test-data"),
			CSSDir:              filepath.Join(basePath, "test-data"),
			TempDir:             filepath.Join(basePath, "test-data", "tmp"),
			BlogPostIssuePrefix: "[BLOG]",
			WebsiteURL:          "http://localhost:8080/foo",
		},
		&buf,
	)
	if err != nil {
		t.Errorf("failed to generate blog post: %v", err)
	}

	got := normalize(buf.String())
	want := normalize(`<head>
		<title>Test Title - Website name</title>


		<link rel="stylesheet" type="text/css" href="/foo/internal/generator/test-data/css/bar.css">

	<link rel="stylesheet" type="text/css" href="/foo/internal/generator/test-data/foo.css">


	</head>

	<body>
		<header><div>Test Header</div>
	</header>
		<h1>Test HTML</h1>
		<footer><div>Test Footer</div>
	</footer>
	</body>
`)

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf(
			"GenerateBlogPosts(eleventy) mismatch (-want +got):\n%s",
			diff,
		)
	}
}

func TestExtractPlainText(t *testing.T) {
	input := `<b>
This is a test with an image <img href="foo.jpg" alt="an image" />
</b>`

	want := normalize("This is a test with an image")
	got := normalize(extractPlainText(template.HTML(input)))

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf(
			"extractPlainText(%s) mismatch (-want +got):\n%s",
			input,
			diff,
		)
	}
}

func TestGenerateBlogIndexPage(t *testing.T) {
	articles := []article{
		{
			Image:   filepath.Join("test-data", "image.jpg"),
			Author:  "Test Author",
			Preview: "Test preview",
			Title:   "Test Title 1",
			DatePublished: parser.FormatTime(
				time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
				config.DefaultDateLayout,
			),
			URL: "bar/baz.html",
		},
		{
			Author:  "Test Author",
			Preview: "Test preview without an image",
			Title:   "Test Title 2",
			DatePublished: parser.FormatTime(
				time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
				config.DefaultDateLayout,
			),
			URL: "bar/qux.html",
		},
	}

	_, file, _, _ := runtime.Caller(0)
	basePath := filepath.Dir(file)
	var buf bytes.Buffer
	err := generateBlogIndexPage(
		articles,
		config.Config{
			TemplateIndexFilePath: filepath.Join(
				basePath,
				"test-data",
				"index.html",
			),
			TemplateHeaderFilePath: filepath.Join(
				basePath,
				"test-data",
				"header.html",
			),
			TemplateFooterFilePath: filepath.Join(
				basePath,
				"test-data",
				"footer.html",
			),
			OutputDir:           filepath.Join(basePath, "test-data"),
			WebsiteURL:          "https://localhost:8080/foo",
			TempDir:             ".",
			DateLayout:          "2006-01-02 15:04:05 UTC",
			TemplateRenderer:    config.DefaultTemplateRenderer,
			BlogTitle:           "foo",
			BlogPostIssuePrefix: "bar",
			CSSDir:              filepath.Join(basePath, "test-data", "css"),
		},
		&buf,
	)
	if err != nil {
		t.Errorf("failed to generate blog post: %v", err)
	}

	got := normalize(buf.String())
	want := normalize(`<html>
        
        <body>
        <header><div>Test Header</div>
        </header>
        
        <article>
        <a href="/foo/bar/baz.html">
        <h1>Test Title 1</h1>
        </a>
        <p>Test preview</p>
        <p>Author: Test Author</p>
        <p>Published: 1970-01-01 00:00:00 UTC</p>
        <a href="/foo/bar/baz.html"><img src="test-data/image.jpg" alt="Test Title 1" /></a>
        </article>
        
        <article>
        <a href="/foo/bar/qux.html">
        <h1>Test Title 2</h1>
        </a>
        <p>Test preview without an image</p>
        <p>Author: Test Author</p>
        <p>Published: 1970-01-01 00:00:00 UTC</p>
        <a href="/foo/bar/qux.html"><img src="" alt="Test Title 2" /></a>
        </article>
        
        <footer><div>Test Footer</div>
        </footer>
        </body>
        
        </html>
		`)

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf(
			"generateBlogIndexPage(default) mismatch (-want +got):\n%s",
			diff,
		)
	}
}

// TODO(test): merge the index generation tests
func TestGenerateBlogIndexPageWithEleventy(t *testing.T) {
	defer cleanupEleventyConfigFile(t)

	articles := []article{
		{
			Image:   filepath.Join("test-data", "image.jpg"),
			Author:  "Test Author",
			Preview: "Test preview",
			Title:   "Test Title 1",
			DatePublished: parser.FormatTime(
				time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
				config.DefaultDateLayout,
			),
			URL: "bar/baz.html",
		},
		{
			Author:  "Test Author",
			Preview: "Test preview without an image",
			Title:   "Test Title 2",
			DatePublished: parser.FormatTime(
				time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
				config.DefaultDateLayout,
			),
			URL: "bar/qux.html",
		},
	}

	_, file, _, _ := runtime.Caller(0)
	basePath := filepath.Dir(file)
	var buf bytes.Buffer
	err := generateBlogIndexPage(
		articles,
		config.Config{
			TemplateRenderer: config.EleventyTemplateRenderer,
			TemplateIndexFilePath: filepath.Join(
				basePath,
				"test-data",
				"index.liquid",
			),
			TemplateHeaderFilePath: filepath.Join(
				basePath,
				"test-data",
				"header.html",
			),
			TemplateFooterFilePath: filepath.Join(
				basePath,
				"test-data",
				"footer.html",
			),
			OutputDir:           filepath.Join(basePath, "test-data"),
			WebsiteURL:          "https://localhost:8080/foo",
			TempDir:             ".",
			DateLayout:          "2006-01-02 15:04:05 UTC",
			BlogTitle:           "foo",
			BlogPostIssuePrefix: "bar",
			CSSDir:              filepath.Join(basePath, "test-data", "css"),
		},
		&buf,
	)
	if err != nil {
		t.Errorf("failed to generate blog post: %v", err)
	}

	got := normalize(buf.String())
	want := normalize(`<html>
        
        <body>
        <header><div>Test Header</div>
        </header>
        
        <article>
        <a href="/foo/bar/baz.html">
        <h1>Test Title 1</h1>
        </a>
        <p>Test preview</p>
        <p>Author: Test Author</p>
        <p>Published: 1970-01-01 00:00:00 UTC</p>
        <a href="/foo/bar/baz.html"><img src="test-data/image.jpg" alt="Test Title 1" /></a>
        </article>
        
        <article>
        <a href="/foo/bar/qux.html">
        <h1>Test Title 2</h1>
        </a>
        <p>Test preview without an image</p>
        <p>Author: Test Author</p>
        <p>Published: 1970-01-01 00:00:00 UTC</p>
        <a href="/foo/bar/qux.html"><img src="" alt="Test Title 2" /></a>
        </article>
        
        <footer><div>Test Footer</div>
        </footer>
        </body>
        
        </html>
		`)

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf(
			"generateBlogIndexPage(eleventy) mismatch (-want +got):\n%s",
			diff,
		)
	}
}

// TODO: not really useful anymore, should test the utils
func TestMoveGeneratedFilesToOutputDir(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	basePath := filepath.Dir(file)

	tempDir := filepath.Join(basePath, "test-data", "tmp")
	outputDir := filepath.Join(basePath, "test-data", "out")
	cfg := config.Config{
		TempDir:   tempDir,
		OutputDir: outputDir,
	}

	t.Run("HAPPY PATH", func(t *testing.T) {
		t.Cleanup(func() {
			os.RemoveAll(tempDir)
			os.RemoveAll(outputDir)
		})

		err := os.Mkdir(tempDir, 0755)
		if err != nil {
			if os.IsExist(err) {
				err = os.RemoveAll(tempDir)
				if err != nil {
					t.Errorf(
						"moveGeneratedFilesToOutputDir() - could not remove temp dir: %v",
						err,
					)
				}
			} else {
				t.Errorf(
					"moveGeneratedFilesToOutputDir() - could not create temp dir: %v",
					err,
				)
			}
		}

		filenames := []string{"foo.txt", "bar.qux.md"}
		newFilePaths := []string{
			filepath.Join(tempDir, "subdir", filenames[0]),
			filepath.Join(tempDir, filenames[1]),
		}

		for _, path := range newFilePaths {
			err = os.MkdirAll(filepath.Dir(path), 0755)
			if err != nil {
				t.Errorf(
					"moveGeneratedFilesToOutputDir() - could not create subdir: %v",
					err,
				)
			}

			_, err = os.Create(path)
			if err != nil {
				t.Errorf(
					"moveGeneratedFilesToOutputDir() - could not create test file: %v",
					err,
				)
			}
		}

		if _, err = os.Stat(cfg.OutputDir); os.IsExist(err) {
			err = os.RemoveAll(cfg.OutputDir)
			if err != nil {
				t.Errorf(
					"moveGeneratedFilesToOutputDir() - could not remove existing output dir: %v",
					err,
				)
			}
		}

		err = moveGeneratedFilesToOutputDir(cfg)
		if err != nil {
			t.Errorf("moveGeneratedFilesToOutputDir(): %v", err)
		}

		if _, err = os.Stat(cfg.OutputDir); os.IsNotExist(err) {
			t.Errorf(
				"moveGeneratedFilesToOutputDir() - output dir was not created: %v",
				err,
			)
		}

		if _, err = os.Stat(cfg.TempDir); os.IsExist(err) {
			t.Errorf(
				"moveGeneratedFilesToOutputDir() - temp dir was not removed: %v",
				err,
			)
		}

		movedFiles, err := os.ReadDir(cfg.OutputDir)
		if err != nil {
			t.Errorf(
				"moveGeneratedFilesToOutputDir() - could not read output dir: %v",
				err,
			)
		}

		if len(movedFiles) != 2 {
			t.Errorf(
				"moveGeneratedFilesToOutputDir() - There are %d moved files instead of 2",
				len(movedFiles),
			)
		}

		for _, file := range movedFiles {
			t.Logf("file: %v", file.Name())
			t.Logf("file.IsDir(): %v", file.IsDir())

			path := filepath.Join(cfg.OutputDir, file.Name())

			if file.IsDir() {
				subFiles, err := os.ReadDir(path)
				if err != nil {
					t.Errorf(
						"moveGeneratedFilesToOutputDir() - could not read subdir: %v",
						err,
					)
				}

				if len(subFiles) != 1 {
					t.Errorf(
						"moveGeneratedFilesToOutputDir() - There are %d files in subdir instead of 1",
						len(subFiles),
					)
				}

				if subFiles[0].Name() != filenames[0] {
					t.Errorf(
						"moveGeneratedFilesToOutputDir() in subdir, file found: %q, want %q",
						subFiles[0].Name(),
						filenames[0],
					)
				}
			} else {
				if file.Name() != filenames[1] {
					t.Errorf(
						"moveGeneratedFilesToOutputDir() in output dir, file found: %q, want %q", file.Name(), filenames[1],
					)
				}
			}
		}

	})

	// happy path
	// Create cfg.TempDir
	// in dir 1: create file, dir/file, dir/subdir/file
	// if cfg.Outputdir does not exists -> create it
	// all files should be moved to outputdir
	// tempdir should have been removed at the end

	// sad path
	// if cfg.TempDir == "" -> return error
	// if cfg.OutputDir == "" -> return error
	// if cfg.TempDir does not exist -> return error
	// if cfg.TempDir is not a directory -> return error
	// if cfg.Outputdir is not a directory -> return error

}

func cleanupEleventyConfigFile(t *testing.T) {
	rootDir, err := utils.RootDir()
	if err != nil {
		t.Errorf("[CLEANUP] failed to get root dir: %v", err)
	}

	err = os.Remove(filepath.Join(rootDir, defaultEleventyConfigFileName))
	if err != nil {
		t.Errorf("[CLEANUP] failed to remove eleventy config file: %v", err)
	}
}
