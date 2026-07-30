# Hugo Analysis for Structured Docs

## Purpose

This document evaluates Hugo (https://github.com/gohugoio/hugo) as a potential foundation for the Structured Docs engine. It documents what Hugo provides, what differs from our goals, and why we are not using Hugo as a dependency.

## How Hugo Works

### Content Model

Hugo's primary content source is **Markdown files with YAML/TOML/JSON front matter** in `content/`:

```markdown
---
title: Hello World
date: 2026-07-29
tags: [hello, world]
---
This is the body of my post.
```

- The front matter becomes page parameters (`.Title`, `.Date`, `.Params.tags`)
- The Markdown body becomes `.Content` (rendered to HTML via Goldmark)
- The content type is derived from the section directory (e.g., `content/posts/` → type `posts`)
- **No schema validation** — front matter is parsed as `map[string]any`
- `page` object is the data passed to templates

### Data Files (Secondary)

Files in the `data/` directory (YAML/TOML/JSON/XML/CSV) are supplementary:

```
data/a.yml          → hugo.Data.a
data/c/d.yml        → hugo.Data.c.d
data/myfolder/foo.yml → hugo.Data.MyFolder.foo
```

- Loaded into a `map[string]any` tree keyed by directory structure
- Accessible in templates via `hugo.Data` or deprecated `.Site.Data`
- **Not used for page content** — only for ancillary data (config lists, translations, etc.)
- **No schema validation**

### Template System

Templates live in `layouts/`. Naming convention:

```
[path/]<layout>[.<outputFormat>][.<kind>][.<lang>].<ext>
```

Examples:
- `layouts/_default/single.html`
- `layouts/posts/single.html`
- `layouts/posts/single.rss.xml`

Template lookup uses a **weighted descriptor matching** system:
1. The page builds a `TemplateDescriptor` (kind, layout, output format, media type)
2. Walks the template tree from the page's path upward
3. Computes a weight for each candidate (kind match = 5pts, layout = 4-6pts, format = 4pts)
4. Highest-weight, closest-path match wins
5. Base templates (`baseof.html`) are applied as an overlay

Templates are executed via Go's `html/template` (for HTML) or `text/template` (for plain text).

### Output Formats

A single content page can render to **multiple output formats** (HTML, RSS, JSON, etc.):

- Each page creates one `pageOutput` per configured render format
- Output formats are defined in `output/` as `Format` structs (Name, MediaType, Path, IsPlainText, IsHTML)
- Content rendering results can be shared across formats when the same template is used

### Compilation Pipeline

```
content/foo.md ──┐
                  ├──► Parse front matter → page.Params
                  │    Parse body → page.Content (Markdown→HTML)
                  │
                  ├──► Resolve template (weighted matching)
                  │
                  ├──► Execute template with page as data
                  │    └── Template context: .Title, .Content,
                  │        .Params, .Date, hugo.Data, .Site, etc.
                  │
                  └──► Write output → public/foo/index.html
```

### File Watching

- Uses `fsnotify` with a `Batcher` that debounces events in time intervals (500ms default)
- Supports partial rebuilds: only changed pages/templates are reprocessed
- Falls back to polling if native fsnotify is unavailable

## Key Differences from Structured Docs

| Aspect | Hugo | Structured Docs |
|---|---|---|
| **Content source** | Markdown files with front matter | Pure YAML data files |
| **Content =** | Markdown body + front matter params | Entire YAML file is structured data |
| **Content type** | Convention-based (section directory) | Schema-defined (explicit `type` field) |
| **Validation** | None (front matter is `map[string]any`) | Full type/field validation against schema |
| **Data files role** | Supplementary (hugo.Data) | Primary (they ARE the content) |
| **Template matching** | Path + kind + layout convention | Field-compatibility matching |
| **Template fields** | Page object exposes all fields | AST-based field extraction from templates |
| **Schema system** | Does not exist | YAML-defined type schemas with typed fields |
| **Output format** | Template extension determines format | Same approach (`.template.md` → `.md`) |

### Core Inversion

Hugo: **Markdown body is content, front matter is metadata**

Structured Docs: **YAML data IS the complete content, template transforms it**

## What Hugo Provides That We Don't Need

| Hugo Feature | Reason Not Needed |
|---|---|
| `hugolib/page.go` — Full page rendering pipeline | Our data has no Markdown body to render |
| `hugolib/site_render.go` — Site rendering orchestrator | We have no site structure (no home page, sections, taxonomies) |
| `tpl/tplimpl/templatestore.go` — Template store (6500+ lines) | Our template matching is simpler (field-based, not weighted) |
| `resources/` — Image processing, asset pipeline | Out of scope — pure documentation compilation |
| `markup/` — Goldmark Markdown converter | Output is template-defined, not Markdown-converted |
| `modules/` — Theme/module system | Too heavy for initial scope |
| `commands/` — CLI with server, livereload | We use cobra directly, simpler CLI |
| `hugolib/pagesfromdata/` — Content adapter system | Our whole engine IS a content adapter |
| `livereload/` + `deploy/` | Out of scope |

## What We Share with Hugo's Design

| Aspect | How Hugo Does It | How We Do It |
|---|---|---|
| **Output format by extension** | `.html` → HTML, `.xml` → RSS | `.template.md` → `.md`, `.template.html` → `.html` |
| **File watching** | fsnotify + batched debounce | Same approach |
| **Template engine** | Go `text/template` / `html/template` | Same |
| **Configuration** | Project root config file + sensible defaults | Same |
| **Convention over config** | Directory structure dictates behavior | Same |

## What We Import Instead of Hugo

| Need | Library | Also Used By Hugo |
|---|---|---|
| YAML parsing | `goccy/go-yaml` | Yes |
| File watching | `fsnotify/fsnotify` | Yes |
| CLI framework | `spf13/cobra` | Yes |
| Filesystem abstraction | `spf13/afero` | Yes |
| Template engine | Go stdlib `text/template` | Yes (Hugo uses a fork) |

## Conclusion

Hugo's architecture inspired several design decisions (output format by extension, convention-over-configuration, batched file watching), but its core model — Markdown content with supplementary data files — is the inverse of Structured Docs' approach (pure YAML data as primary content with schema validation). Importing Hugo as a Go dependency would pull in a heavy, tightly-coupled rendering pipeline while providing none of our differentiating features (typed schemas, field-based template matching, validation). We use the same underlying libraries directly for a lighter, focused dependency tree.
