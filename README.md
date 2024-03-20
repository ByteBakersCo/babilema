# Babilema


> 🚧 This is a work in progress 🚧  


Babilema (Esperanto for _garrulous_) is a minimalist static blog generator that
turns GitHub issues into blog posts.  
It is intended to be used as a GitHub action.

## Markdown metadata structure (AKA front matter)

Input issues should be written in markdown with a **TOML** front matter.  
The front matter should be at the very top of the file and start/end with `---`.  

**NOTE**: The Markdown body of your issue will be used as the actual content of your blog post.  
The title, author, date published and date modified will be added automatically using the metadata.  

Here's an example of a blog post issue with the currently supported metadata:  

```toml
---
# This is actual TOML so you can add comments
website_url = "http://localhost:8080" # The URL of your website
title = "My first blog post" # REQUIRED - <title>...</title> + post title
slug = "my-first-blog-post"  # REQUIRED - The URL and filename slug
page_subtitle = "My super blog" # Added to the <title> tag. Will look like <title>My first blog post - My super blog</title>
description = "This is my first blog post" # <meta name="description" content="...">
keywords = ["blog", "first", "post"] # <meta name="keywords" content="...">
author = "John Doe" # <meta name="author" content="..."> + post author
image = "http://example.com/image.jpg" # Or use a relative path # Social media/SEO image
publisher = "John Doe Team" # <meta name="publisher" content="...">
tags = ["blog", "tutorial"] # Will be used to reference other blog posts
---

**This is the content of my first blog post.**  
It uses GitHub issues as its source.  
```

You can find an example of a blog post in the issues.  

## Configuration file

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

The default configuration file would look like this (if it wasn't built at runtime):

```toml 
website_url = "http://localhost:8080"   # The URL of your website
blog_post_issue_prefix = "[BLOG]"       # The prefix of your blog post issues title
output_dir = "{repo_root}/"             # The directory where the generated html files will be saved
template_post_file_path = "{repo_root}/{output_dir}/templates/post.html"
template_header_file_path = "{repo_root}/{output_dir}/templates/header.html"
templateFooterFilePath = "{repo_root}/{output_dir}/templates/footer.html"
css_dir = "{repo_root}/{output_dir}/templates/css" # The directory where the CSS files are stored (if any)
```

`{repo_root}` will be replaced by the absolute path to the repository root.    

By default, `output_dir` will be used to determine the root of all other paths and it will always be preceded by `{repo_root}`.  
Meaning that if you leave `output_dir` empty, it will be equal to `{repo_root}`, if you set it to `blog/` it will be `{repo_root}/blog/`.  
If you want to use a different directory for your templates, it will NOT be preceded by `output_dir` but only `{repo_root}`.  

## Usage tips
### robots.txt
**Don't forget to at least disallow your templates directory path in your
robots.txt file**

e.g:  
```txt
User-agent: *
Disallow: /blog/templates/
```

### History file

Babilema generates and uses a `.babilema-history.toml` file in the `output_dir` in order to
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
