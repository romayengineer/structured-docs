# Structured Docs — Plan

## Overview

A Go-based documentation compilation engine. Reads typed YAML data files, matches them to templates based on field compatibility, and compiles structured .md and .html output.

## Project Structure

```
structured-docs/
├── main.go                       # CLI entry point
├── go.mod
├── structured.yml                # Project configuration
├── schema/                       # Type schema definitions (.yml)
│   └── post.yml
├── data/                         # Content data files (.yml, mirrors output tree)
│   └── blog/
│       └── hello-world.yml
├── templates/                    # Template files
│   ├── post.template.md          # → outputs .md
│   └── post.template.html        # → outputs .html
├── output/                       # Compiled output (mirrors data/)
│   └── blog/
│       ├── hello-world.md
│       └── hello-world.html
└── PLAN.md
```

## Configuration (`structured.yml`)

```yaml
schema_dir: schema          # default
data_dir: data              # default
template_dir: templates     # default
output_dir: output          # default
template_order:             # required — implicit resolution order
  - post.template.md
  - post.template.html
  - default.template.md
```

All directory fields have sensible defaults; only `template_order` is required.

## Type Schema

Defined as YAML files in `schema/`. The filename before extension is the type name (or the `type` field can override it).

```yaml
# schema/post.yml
type: post
description: "A blog post"
fields:
  - name: title
    type: string
    required: true
  - name: date
    type: string
  - name: body
    type: string
    required: true
  - name: tags
    type: "[]string"
```

Supported field types: `string`, `int`, `float`, `bool`, `[]string`, `[]int`, `map[string]string`.

## Data File

YAML files in `data/`. The directory structure under `data/` is mirrored in `output/`.

```yaml
# data/blog/hello-world.yml
type: post
# template: post.template.md     # optional explicit override
title: Hello World
date: 2026-07-29
body: This is my first post!
tags: [hello, world]
```

Validation at compile time: required fields present, types match, no unknown fields.

## Template System

- Files use `.template.{lang}` extension (`.template.md`, `.template.html`)
- Content is standard Go `text/template` syntax
- The engine parses the template with `text/template/parse`, walks the AST, and collects all `FieldNode` identifiers (`.title`, `.body`, etc.) — these are the template's required fields
- Template extension determines output format:
  - `.template.md` → compiled to `.md`
  - `.template.html` → compiled to `.html`

## Template Resolution

Two strategies:

1. **Explicit**: Data file has a `template:` field naming the template file to use.
2. **Implicit**: Iterate `template_order` from config. First template whose auto-detected field set is a subset of the data type's field set wins.

A single data file can match both a `.template.md` and a `.template.html` template independently (same or different templates), producing both `.md` and `.html` output.

## Engine Architecture (`pkg/structured/`)

| Package    | Responsibility                                            |
|------------|-----------------------------------------------------------|
| `config`   | Load and validate `structured.yml`                        |
| `schema`   | Parse schema YAML files, build type registry              |
| `data`     | Load data YAML files, validate against their type schema  |
| `template` | Load template files, extract field references via AST walk |
| `resolver` | Match data entries to templates (explicit + implicit)     |
| `renderer` | Execute Go `text/template` with validated data            |
| `compiler` | Orchestrate the full pipeline                             |
| `watcher`  | fsnotify-based file watcher for `--watch`                 |

## CLI

```
sd [flags]

Flags:
  --config path   Path to structured.yml (default: ./structured.yml)
  --watch, -w     Watch for file changes and recompile
  --clean         Remove output directory before compiling
```

Exit code 1 on validation errors.

## Deliverables

1. Working CLI that compiles a project with schema + data + templates
2. Watch mode for live recompilation
3. Go library (`pkg/structured/`) importable by external projects
4. Input validation: schema conformance, type checking, field completeness
5. Example project with `schema/post.yml`, `data/blog/hello-world.yml`, `templates/post.template.md`
