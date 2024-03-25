# Babilema

  
> 🚧 This is a work in progress 🚧  
  

Babilema (Esperanto for _garrulous_) is a minimalist static blog generator that
turns GitHub issues into blog posts.  
It is intended to be used as a GitHub action.

- [Usage](#usage)
- [Configuration file](#configuration-file)
  * [Markdown metadata structure (AKA front matter)](#markdown-metadata-structure-aka-front-matter)
- [Installation](#installation)
  * [Build from source](#build-from-source)
  * [Run from source](#run-from-source)
- [Usage tips](#usage-tips)
  * [Debugging your templates](#debugging-your-templates)
  * [Writing your own templates](#writing-your-own-templates)
  * [robots.txt](#robotstxt)
  * [Who can write?](#who-can-write)
  * [History file](#history-file)
- [Contributing](#contributing)

## Usage
In order to use Babilema, you need to have a GitHub repository with issues that will be used as blog posts.  
You can run Babilema from the command line or as a [GitHub action](https://github.com/ByteBakersCo/babilema-action).  
The minimum requirements are the environment variables `GITHUB_REPOSITORY` and `GITHUB_TOKEN` that need to be set.  

```bash
GITHUB_REPOSITORY="owner/repo" # or {{ github.repository }} in a GitHub action
GITHUB_TOKEN="your_personal_access_token" # or ${{ secrets.GITHUB_TOKEN }} in a GitHub action
```

## Configuration file
Babilema uses a TOML configuration file, by default it will look for
`.babilema.toml` at the root of your repo or wherever you are running the `babilema` command from.  
You can pass a different path as a command line argument.
```bash
babilema --config /path/to/your/config.toml

# or, if you're running from source without any environment variables set

go run cmd/babilema/main.go --config /path/to/your/config.toml
```

The default configuration file would look like this (if it wasn't built at runtime):
```toml 
website_url = "http://localhost:8080"       # The URL of your website
blog_title = ""                             # The title of your blog, can be overwritten per issue
blog_post_issue_prefix = "[BLOG]"           # The prefix of your blog post issues title
output_dir = "{repo_root}/"                 # The directory where the generated html files will be saved
temp_dir = "{repo_root}/tmp"                 # The directory where the temporary files will be saved
template_post_file_path = "{repo_root}/{output_dir}/templates/post.html"
template_header_file_path = "{repo_root}/{output_dir}/templates/header.html"
template_footer_file_path = "{repo_root}/{output_dir}/templates/footer.html"
template_index_file_path = "{repo_root}/{output_dir}/templates/index.html" # Your blog's homepage file
css_dir = "{repo_root}/{output_dir}/templates/css" # The directory where the CSS files are stored (if any)
```

`{repo_root}` will be replaced by the absolute path to the repository or website root (`/`).  

By default, `output_dir` will be used to determine the root of all other paths and it will always be preceded by `{repo_root}`.  
Meaning that if you leave `output_dir` empty, it will be equal to `{repo_root}`, if you set it to `blog/` it will be `{repo_root}/blog/`.  
If you want to use a different directory for your templates, it will NOT be preceded by `output_dir` but only `{repo_root}`.  

### Markdown metadata structure (AKA front matter)
Issues should be written in markdown with a **TOML** front matter.  
The front matter should be at the very top of the file and start/end with `---`.  

**NOTE**: The Markdown body of your issue will be used as the actual content of your blog post.  
The title, author, date published and date modified will be added automatically using the metadata.  

Here's an example of a blog post issue with the currently supported metadata:  
```toml
---
# This is actual TOML so you can add comments
title = "My first blog post" # REQUIRED - <title>...</title> + post title
blog_title= "My super blog" # Added to the <title> tag. Will look like <title>My first blog post - My super blog</title>
slug = "my-first-blog-post"  # REQUIRED - The URL and filename's slug
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

Note that `blog_title` will override the values in the configuration file if it 
is set on a particular issue.  
You can find an example of a blog post in the issues.  

## Installation
### Build from source
To build Babilema from source, you can use the provided `build.sh` script.  
It will create a binary in the `bin` directory.  
```bash
./build.sh
```

Make sure you have Go installed on your system.  
You also need to set the following environment variables to run Babilema:  
```bash
GITHUB_TOKEN="your_personal_access_token"
GITHUB_REPOSITORY="your_username/your_repo"
```

`GITHUB_TOKEN` is your personal access token with `read` access to `Issues`.  

**NEVER SHARE YOUR PERSONAL ACCESS TOKEN**  
Be sure to keep it safe and never commit it to your repository.  

## Usage tips
### Debugging your templates
When generating the HTML files, Babilema will stop at the first error it encounters.  
You can find the generated files in the `temp_dir` directory (default: `{repo_root}/tmp`).  
Once you've fixed the error in your templates, you can re-run Babilema to generate the files again.  
All the generated files will be copied to the `output_dir` directory (default: `{repo_root}/`) and the `temp_dir` will be deleted.  

### Writing your own templates
You can find basic example templates in the `templates` directory.  
They contain all the relevant fields to generate a blog post and your blog's home page.  

### robots.txt
**Don't forget to at least disallow your templates directory path in your
robots.txt file**

e.g:  
```txt
User-agent: *
Disallow: /blog/templates/
```

### Who can write?
Only users with write access to the repository can create blog posts.  
When parsing the issues, Babilema will only consider the ones created by users with the permission "admin" or "write".    

### History file

Babilema generates and uses a `.babilema-history.toml` file in the `output_dir` in order to
check whether a given issue was already parsed and if it was modified since
last time.  If you want to re-generate all blog posts, you can delete this
file.

---
## Contributing
PRs are welcome!  
The goal here is to have just the essentials to have a basic blog that does it's primary job: display blog posts.  

If you want to open a PR, feel free to do so. 
Just make sure to (more or less) follow the [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) standard.  
Also make sure to run `gofmt` on your code before committing.  
Ideally, lines should not exceed 80 characters.  

Please try to keep the code as simple as possible and respect the general style of the project.  
Also don't forget to add tests for your code, update the README, yaddi yadda...

And use US English :D  

### TODO

- [x] Write a better README
- [x] Finish implementing demo template
- [x] Add support for custom templates
- [x] Add support for custom themes (CSS)
- [ ] Add support for custom scripts (JS)
- [ ] Add "related articles" section generator
- [x] Handle `index.html` file
- [ ] Add testing blog on this repo
- [ ] Add support for custom metadata
- [ ] Add support for injecting data in header + footer
- [ ] (CI) Check if it's possible to trigger a rebuild on issue update
- [ ] Make sure the injected paths (HTML) is correct on Windows
- [ ] Add auto-optimization for preview images (?)
- [ ] Use goroutines to speed up the process
