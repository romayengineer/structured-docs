# AGENTS.md — Structured Docs (sd)

> **Agent instruction:** Whenever you interact with this project, you MUST ALWAYS communicate in English — both in responses and in any generated/written content.

## Overview

Structured Docs (`sd`) is a Go CLI tool that compiles YAML data files into Markdown/HTML documents using Go `text/template` templates, validated against YAML schema definitions.

**Pipeline:**
```
YAML data files  →  Template matching  →  Go text/template  →  .md / .html output
(schema-validated)  (field-compatibility)  (AST-extracted fields)
```

## Directory Layout

```
structured-docs/
├── cmd/sd/main.go                  # CLI entry point (flags: -config, -watch, -clean)
├── pkg/structured/
│   ├── config/                     # structured.yml config loading
│   ├── schema/                     # YAML schema parsing & field type validation
│   ├── data/                       # YAML data loading & type coercion
│   ├── template/                   # Template loading & AST field extraction
│   ├── resolver/                   # Matching data files to templates
│   ├── renderer/                   # Executing text/template
│   ├── compiler/                   # Orchestrator with pluggable interfaces
│   ├── watcher/                    # fsnotify-based --watch mode
│   └── fsys/                       # FS abstraction (OS + MemFS for testing)
├── example/                        # Working example project
│   ├── schema/post.yml
│   ├── data/blog/hello-world.yml
│   ├── templates/post.template.md
│   └── structured.yml
├── docs/
│   ├── PLAN.md                     # Architecture plan
│   └── HUGO-ANALYSIS.md            # Hugo comparison analysis
├── go.mod / go.sum
├── Makefile
└── README.md
```

## Quick Start

```sh
cd example
sd
# → data/blog/hello-world.yml → output/blog/hello-world.md (md)
# → compiled 1 file(s)
```

## Build & Test

| Command | Purpose |
|---|---|
| `make build` | Build `sd` binary |
| `make test` | `go test ./...` |
| `make cover` | Test with coverage report |
| `make cover-html` | Coverage as HTML |
| `make clean` | `go clean` + remove binary |
| `make run` | Build + run example |
| `make install` | Copy binary to `~/.local/bin/` |
| `go vet ./...` | Static analysis |

```sh
# Build and run directly
go build -o sd ./cmd/sd && ./sd -config example/structured.yml
```

## CLI

```
sd [flags]

Flags:
  -config string   path to structured.yml (default "structured.yml")
  -watch           watch for file changes and recompile
  -clean           remove output directory before compiling
```

Exit code 1 on validation errors.

## Architecture

### Compilation Pipeline

```
config.Load(fs, path)
  │
  ▼
SchemaLoader  ──►  map[string]*TypeDefinition   (schema/*.yml)
TemplateLoader ──►  []*Template                  (templates/*.template.{md,html})
DataLoader     ──►  []*DataFile                  (data/**/*.yml)
  │
  ▼
Resolver  ──►  []*Job  (pairs each DataFile with a compatible Template)
  │
  ▼
Renderer ──► rendered string → fs.WriteFile()
```

### Key Types

| Type | Package | Description |
|---|---|---|
| `Config` | `config` | `SchemaDir`, `DataDir`, `TemplateDir`, `OutputDir`, `TemplateOrder` |
| `TypeDefinition` | `schema` | `Name`, `Description`, `[]FieldDefinition` (name, type, required) |
| `DataFile` | `data` | `SourcePath`, `TypeName`, `ExplicitTemplate`, `Fields` |
| `Template` | `template` | `FileName`, `Content`, `RequiredFields`, `OutputExt` |
| `Job` | `resolver` | `*DataFile` + `*Template` pair |
| `Result` | `compiler` | `SourcePath`, `OutputPath`, `Format` |
| `FS` | `fsys` | Interface: `ReadFile`, `WriteFile`, `ReadDir`, `MkdirAll`, `RemoveAll`, `Walk` |

### Interfaces (pluggable pipeline)

The `Compiler` struct wires together each stage via interfaces. Every interface has a corresponding `*Func` adapter for easy mocking:

```go
type SchemaLoader interface   { LoadSchema(fsys.FS, string) (map[string]*schema.TypeDefinition, error) }
type TemplateLoader interface { LoadTemplates(fsys.FS, string) ([]*template.Template, error) }
type DataLoader interface     { LoadData(fsys.FS, string, map[string]*schema.TypeDefinition) ([]*data.DataFile, error) }
type Resolver interface       { Resolve([]*data.DataFile, []*template.Template, []string, map[string]*schema.TypeDefinition) ([]*resolver.Job, error) }
type Renderer interface       { Render(*resolver.Job) (string, error) }
```

```go
compiler := &compiler.Compiler{
    FS:       fsys.OS{},
    Schema:   compiler.SchemaLoaderFunc(schema.LoadAll),
    Template: compiler.TemplateLoaderFunc(template.LoadAll),
    Data:     compiler.DataLoaderFunc(data.LoadAll),
    Resolve:  compiler.ResolverFunc(resolver.ResolveAll),
    Render:   compiler.RendererFunc(renderer.Render),
}
results, err := compiler.Compile(cfg)
```

## Schemas

Defined as YAML files in the schema directory. The filename (minus extension) is the type name.

```yaml
# schema/post.yml
description: "A blog post"
fields:
  - name: title
    type: string
    required: true
  - name: tags
    type: "[]string"
    required: false
```

Valid field types: `string`, `int`, `float`, `bool`, `[]string`, `[]int`, `map[string]string`

**IMPORTANT:** A `nil` `Required` field defaults to `true`. Set `required: false` explicitly for optional fields.

## Data Files

```yaml
# data/blog/my-post.yml
type: post                          # must match a schema filename
template: post.template.md          # optional — explicit template override
title: My Post
body: |
  Markdown content here...
```

- `type` field is **required** and must match a schema file name (case-sensitive)
- All `required: true` fields must be present
- Unknown fields are rejected
- Values are coerced to the declared type (e.g., `int` accepts float64 truncated)
- Directory structure under `data/` is mirrored in `output/`

## Templates

```gotemplate
# templates/post.template.md
# {{ .title }}

*Date: {{ .date }}*

{{ .body }}

{{ if .tags }}Tags: {{ range .tags }}{{ . }} {{ end }}{{ end }}
```

- Files use `.template.md` or `.template.html` suffix
- Uses Go's `text/template` syntax
- Output extension is derived from the template suffix: `.template.md` → `.md`, `.template.html` → `.html`
- Only top-level field access is supported by the AST extractor (e.g., `.author.name` registers as required field `author`)

### CRITICAL: No template functions

The `extractFields` parser uses `text/template/parse` which does **not** register Go template functions. **No function calls are allowed** — not `eq`, `and`, `not`, `printf`, `len`, etc.

```gotemplate
# BROKEN — all fail to parse:
{{ if eq .status "published" }}
{{ if not .featured }}
{{ if and .a .b }}
{{ printf "%s" .title }}
```

Only use:
- Field access: `{{ .field }}`, `{{ .nested.field }}`
- Actions: `if`, `range`, `with`
- Variables: `{{ $var := .field }}`

## Template Resolution

Two strategies, checked in order:

1. **Explicit**: Data file has a `template:` field — uses that template file directly
2. **Implicit**: Iterates `template_order` from config; picks the **first** template whose auto-extracted `RequiredFields` are all present in the data type's schema

The `RequiredFields` list is extracted by walking the template's AST with `text/template/parse` — only `FieldNode` identifiers are collected.

**Strategy:** Put more specific templates first in `template_order`.

## Configuration (`structured.yml`)

```yaml
schema_dir: schema        # default: "schema"
data_dir: data            # default: "data"
template_dir: templates   # default: "templates"
output_dir: output        # default: "output"
template_order:           # REQUIRED — no default
  - post.template.md
```

Only `template_order` is required; all directory fields have sensible defaults.

## Filesystem Abstraction

All file I/O goes through the `fsys.FS` interface. Two implementations:

- **`fsys.OS{}`** — delegates to `os`/`filepath` (used in production, created in `main.go`)
- **`fsys.MemFS{}`** — in-memory map (used in tests, hermetic)

```go
type FS interface {
    ReadFile(name string) ([]byte, error)
    WriteFile(name string, data []byte, perm os.FileMode) error
    ReadDir(name string) ([]fs.DirEntry, error)
    MkdirAll(path string, perm os.FileMode) error
    RemoveAll(path string) error
    Walk(root string, fn filepath.WalkFunc) error
}
```

**Rule:** Never use `os`/`ioutil`/`filepath` directly in library packages — always go through `fsys.FS`.

## Smart Clean Behavior

`compiler.CleanOutput()` protects source directories when they overlap with the output directory:

1. If no source directory (`SchemaDir`, `DataDir`, `TemplateDir`) is inside `OutputDir` → simple `RemoveAll(OutputDir)`
2. If a source directory is inside `OutputDir` → removes everything **except** `.git`, `.gitignore`, and paths that are or contain source directories

## Coding Conventions

- **Interface + Func adapter pattern**: Each pipeline stage has an interface and a `*Func` type for DI
- **In-memory FS for tests**: All tests use `fsys.NewMemFS()` — never touch the real filesystem
- **No template functions in parser**: AST-based field extraction cannot handle them
- **YAML via gopkg.in/yaml.v3**: Use `yaml.Unmarshal` / `yaml.Marshal`
- **Return errors, never panic**: Validation errors return with exit code 1
- **Pointer receiver for Required**: `*bool` with nil meaning `true` (use `BoolPtr()` helper in tests)
- **Table-driven tests**: Standard Go testing style with `memFS` setup

## Dependencies

| Library | Purpose |
|---|---|
| `gopkg.in/yaml.v3` | YAML parsing |
| `github.com/fsnotify/fsnotify` | File watching (--watch mode) |
| Go stdlib `text/template` | Template engine |
| Go stdlib `text/template/parse` | AST-based field extraction |

**Do not add new dependencies without checking `go.mod` first.**

## Common Pitfalls

| Pitfall | Fix |
|---|---|
| `"function not defined"` when testing templates | Don't use any template functions — only `if`, `range`, `with`, field access |
| `missing required field` for an optional field | Add `required: false` to the schema field definition |
| Wrong template selected for a data file | Reorder `template_order` — more specific templates first |
| Using `os.ReadFile` instead of `fs.ReadFile` | Always use `fsys.FS` methods in library code |
| Adding new field type without adding it to `isValidFieldType` | Update both the validator and `normalizeValue` |
| Forgetting `CleanOutput` overlapping source/output dirs | Test with `MemFS` where source dirs are inside output dir |
| Tests failing on `go test ./...` | Run `go vet ./...` and ensure all packages compile first |
