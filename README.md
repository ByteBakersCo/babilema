# Babilema

> This is a work in progress  

Babilema (Esperanto for _garrulous_) is a minimalist static blog generator that
turns GitHub issues into blog posts. It is intended to be used as a GitHub
action.

### Markdown metadata structure (AKA front matter)

Input issues should be written in markdown with a **TOML** front matter. The
front matter should contain the following fields:

```go Description string Keywords    []string Author      string Title
string Slug        string Image       string Publisher   string Logo
string Tags        []string ```

The following fields will be passed as arguments when executing the action.
```go URL           string DatePublished string DateModified  string ```

You can find an example of a blog post in the issues

### Configuration file

Babilema uses a TOML configuration file, by default it will look for
`.babilema.yml` at the root of your repo. You can pass a different path as an
argument when executing the action. Like so: ```yaml name: Babilema uses:
babilema/babilema@v0.1.0 with: config: .babilema.yml ```

The default configuration file would look like this (if it wasn't built in Go):

```toml website_url = "http://localhost:8080" blog_post_issue_prefix: "[BLOG]"
output_dir = "{repo_root}/blog" template_post_file_path =
"{repo_root}/blog/templates/post.html" template_header_file_path =
"{repo_root}/blog/templates/header.html" templateFooterFilePath =
"{repo_root}/blog/templates/footer.html" css_dir =
"{repo_root}/blog/templates/css"```

**Don't forget to at least disallow your templates directory path in your
robots.txt file**

### History file

Babilema uses a `.babilema-history.toml` file in the `output_dir` in order to
check whether a given issue was already parsed and if it was modified since
last time.  If you want to re-generate all blog posts, you can delete this
file.

---

### TODO

- [ ] Write a better README
- [ ] Finish implementing default template
- [x] Add support for custom templates
- [x] Add support for custom themes (CSS)
- [ ] Add support for custom scripts (JS)
- [ ] Add testing blog on this repo
- [ ] Add support for custom metadata
- [ ] Use goroutines to speed up the process
