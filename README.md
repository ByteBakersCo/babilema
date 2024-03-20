# Babilema


> 🚧 This is a work in progress 🚧  


Babilema (Esperanto for _garrulous_) is a minimalist static blog generator that
turns GitHub issues into blog posts.  
It is intended to be used as a GitHub action.

### Markdown metadata structure (AKA front matter)

Input issues should be written in markdown with a **TOML** front matter. The
front matter should be at the very top of the file and start/end with `---`.  
The following fields are currently supported:

```go
struct {
    Title       string    // REQUIRED - <title>...</title> + post title
    Slug        string    // REQUIRED - The URL and filename slug
    PageSubtitle string   // added to the <title> tag. e.g. "<Blog post title> - My super blog"
    Description string    // <meta name="description" content="...">
    Keywords    []string  // <meta name="keywords" content="...">
    Author      string    // <meta name="author" content="..."> + post author
    Image       string    // Social media/SEO image
    Publisher   string    // <meta name="publisher" content="...">
    Tags        []string  // Will be used to reference other blog posts

    // The following fields will be set automatically at runtime by Babilema:
    URL           string  // infered from the configuration file (see below)
    DatePublished string  // infered from the issue creation date
    DateModified  string  // infered from the issue last update date
}
```

You can find an example of a blog post in the issues.

### Configuration file

Babilema uses a TOML configuration file, by default it will look for
`.babilema.yml` at the root of your repo.  
You can pass a different path as an argument when executing the action.  
Like so:  
```yaml
name: Babilema
uses: babilema/babilema@v0.1.0
with:
  config: 'blog/.config.yml'
```

The default configuration file would look like this (if it wasn't built in Go):

```toml 
website_url = "http://localhost:8080"   # The URL of your website
blog_post_issue_prefix = "[BLOG]"       # The prefix of your blog post issues title
output_dir = "{repo_root}/blog"         # The directory where the generated html files will be saved
template_post_file_path = "{repo_root}/blog/templates/post.html"
template_header_file_path = "{repo_root}/blog/templates/header.html"
templateFooterFilePath = "{repo_root}/blog/templates/footer.html"
css_dir = "{repo_root}/blog/templates/css" # The directory where the CSS files are stored (if any)
```

**Don't forget to at least disallow your templates directory path in your
robots.txt file**

e.g:  
```txt
User-agent: *
Disallow: /blog/templates/
```

### History file

Babilema uses a `.babilema-history.toml` file in the `output_dir` in order to
check whether a given issue was already parsed and if it was modified since
last time.  If you want to re-generate all blog posts, you can delete this
file.

---

### TODO

- [x] Write a better README
- [ ] Finish implementing default template
- [x] Add support for custom templates
- [x] Add support for custom themes (CSS)
- [ ] Add support for custom scripts (JS)
- [ ] Add "related articles" section generator
- [ ] Handle `index.html` file
- [ ] Add testing blog on this repo
- [ ] Add support for custom metadata
- [ ] Use goroutines to speed up the process
