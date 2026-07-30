# AGENTS.md

> **Agent instruction:** Whenever you interact with this project, you MUST ALWAYS communicate in English — both in responses and in any generated/written content.

For full documentation of the `sd` tool visit:  
https://github.com/romayengineer/structured-docs

---

## Overview

This project uses [`sd`](https://github.com/romayengineer/structured-docs) to generate Markdown documentation from structured YAML data and Go templates.

**Flow:**
```
YAML data files  →  Go templates  →  compiled .md output
(schemas validate data shape)
```

## Directory Layout

```
.sd/
├── schema/                  # YAML schemas (field definitions)
│   └── post.yml
├── data/                    # YAML data files (the actual content)
│   └── blog/
│       └── hello-world.yml
├── templates/               # Go templates (*.template.md)
│   └── post.template.md
├── output/                  # Generated output (DO NOT edit manually)
├── structured.yml           # Configuration file
└── AGENTS.md                # This file
```

## Workflow

1. **Define a schema** — create `schema/<type>.yml` with fields, types, and `required` flags
2. **Create a template** — create `templates/<name>.template.md` using Go template syntax
3. **Write data** — create `data/<category>/<slug>.yml` with `type: <schema-name>` and field values
4. **Generate** — run `sd` (auto-detects `.sd/`)
5. **Preview** — `.md` files open in VS Code/any Markdown viewer

## Schemas

Available field types: `string`, `int`, `float`, `bool`, `[]string`, `[]int`, `map[string]string`

```yaml
# schema/my-type.yml
description: "Description of this document type"
fields:
  - name: title
    type: string
    required: true
  - name: tags
    type: "[]string"
    required: false
```

**IMPORTANT:** Fields default to `required: true`. Always set `required: false` explicitly for optional fields.

## Templates

- File naming: `*.template.md` (produces `.md` output)
- Uses Go's `text/template` syntax
- Data fields are accessed as `{{ .fieldName }}`

### CRITICAL: No template functions

The `extractFields` parser uses `text/template/parse` which does **not** register Go template functions. **No function calls are allowed** — not `eq`, `ne`, `and`, `not`, `printf`, `len`, etc.

```gotemplate
# BROKEN — all of these will fail:
{{ if eq .status "published" }}
{{ if not .featured }}
{{ if and .a .b }}
{{ printf "%s" .title }}
```

Only use:
- Field access: `{{ .field }}`, `{{ .nested.field }}`
- Actions: `if`, `range`, `with`
- Variables: `{{ $var := .field }}`, `{{ range $k, $v := .map }}`

### Whitespace Control

Three patterns to prevent blank lines when fields are absent:

1. **Section-level conditionals** (headers, badges):
   ```gotemplate
   {{- if .field }}
   Content here
   {{- end }}
   ```

2. **Table rows** (must stay contiguous):
   ```gotemplate
   {{ if .field }}| **Row** | {{ .field }} |
   {{ end }}
   ```

3. **Range loops** (lists — keep the newline between items):
   ```gotemplate
   {{ range .items }}- {{ . }}
   {{ end }}
   ```

## Validation Rules

The `structured.yml` config has validation enabled. When you run `sd`, it checks:

- **Header spacing** — exactly 1 blank line before `##`/`###`/`####`/`#####`/`######` headers
- **File name style** — must be kebab-case (lowercase, hyphen-separated, e.g., `my-post.md`)
- **Leading digit** — file names must NOT start with a digit

If validation fails, `sd` exits with code 1 and the output file is not written.

## Data Files

```yaml
# data/blog/my-post.yml
type: post                      # must match a schema filename (minus .yml)
title: My Post
body: |
  Markdown content here...

  ## Section

  ```go
  code blocks work
  ```

tags:
  - tag1
  - tag2
```

- The `type` field is **required** and must match a schema file name (case-sensitive)
- All `required: true` fields must be present
- Optional fields can be omitted entirely

## Running sd

```sh
# From any directory under the project (auto-detects .sd/):
sd

# Or explicitly from the .sd/ directory:
cd .sd && sd

# Watch for changes:
sd -watch

# Clean and regenerate:
sd -clean
```
