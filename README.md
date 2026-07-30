# Structured Docs

Generate structured documentation from YAML schemas, data, and Go templates.

## Install

```sh
go install github.com/romayengineer/structured-docs/cmd/sd@latest
```

If the module proxy hasn't picked up the latest commit yet:

```sh
GOPROXY=direct go install github.com/romayengineer/structured-docs/cmd/sd@main
```

## Quick Start

```sh
cd example
sd
```

Output:

```
data/blog/hello-world.yml → output/blog/hello-world.md (md)
compiled 1 file(s)
```

## Usage

```
sd [flags]

Flags:
  -config string    path to config file (default "structured.yml")
  -watch            watch for file changes and recompile
  -clean            remove output directory before compiling
```
