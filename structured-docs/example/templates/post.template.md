# {{ .title }}

*Date: {{ .date }}*

{{ .body }}

{{ if .tags }}
Tags: {{ range .tags }}{{ . }} {{ end }}
{{ end }}
