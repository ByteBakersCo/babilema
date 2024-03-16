# Babilema

> This is a work in progress  

Babilema (Esperanto for _garrulous_) is a minimalist static blog generator that turns GitHub issues into blog posts. It is intended to be used as a GitHub action.

### Markdown structure

Input issues should be written in markdown with a **TOML** front matter. The front matter should contain the following fields:

```go
Description string
Keywords    []string
Author      string
Title       string
Slug        string
Image       string
Publisher   string
Logo        string
Tags        []string
```

The following fields will be passed as arguments when executing the action.
```go
URL           string
DatePublished string
DateModified  string
```

You can find an example of a blog post in the issues

---

### TODO

- [ ] Write a better README
- [ ] Finish implementing default template
- [ ] Add better support for custom templates
- [ ] Add support for custom themes (CSS)
- [ ] Add support for custom scripts (JS)
- [ ] Add support for custom fonts
- [ ] Add testing blog on this repo
