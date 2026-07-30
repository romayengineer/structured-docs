# Feature Ideas

## Backlog

- [ ] Compiler: create REDME.md as an index file with all the links to all .md files created
- [ ] Validator: check that all hyperlinks in generated markdown resolve (link checker)
- [ ] Validator: enforce max file size for generated `.md` output
- [ ] Validator: flag broken internal anchors (`[](#nonexistent-section)`)
- [ ] CLI: `--strict` flag to accumulate all validation errors before failing (vs stop on first)
- [ ] CLI: `--manifest` flag to generate a `_index.json` alongside `.md` files
- [ ] Template: support `include` / partial template composition
- [ ] Template: support `html/template` output in addition to `text/template`
- [ ] Config: validate that every schema field has a human-readable description
- [ ] Config: warn on unused fields in data files (fields in data but not in schema)
- [ ] Output: generate table of contents from `##` headers
- [ ] Output: diff mode — show what changed between two runs
- [ ] Watch: debounce interval configurable via config file
- [ ] Watch: only recompile affected files instead of full rebuild

## Implemented

- [x] Pluggable validator interface (`Validator`) with `DefaultValidators()`
- [x] Validator: header spacing (`lines_between_headers` for h1–h4)
- [x] Validator: file name style (kebab, snake, camel, lowercase + custom regex)
- [x] Validator: `allow_leading_digit` option (default true, false rejects digit-starting names)
- [x] Validator: skip headers inside fenced code blocks (``` fences)
- [x] CLI: `-version` flag with version injected via `-ldflags` (dev / git hash)
- [x] Compiler: validate rendered content and output path **before** writing the file
- [x] Tests: integration tests (`//go:build integration`) run separately from unit tests
