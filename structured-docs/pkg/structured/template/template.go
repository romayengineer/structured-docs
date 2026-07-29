package template

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template/parse"
)

type Template struct {
	FileName       string
	Content        string
	RequiredFields []string
	OutputExt      string
}

func LoadAll(templateDir string) ([]*Template, error) {
	var templates []*Template

	err := filepath.Walk(templateDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		name := info.Name()
		if !strings.HasSuffix(name, ".template.md") && !strings.HasSuffix(name, ".template.html") {
			return nil
		}

		b, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading template %s: %w", path, err)
		}

		content := string(b)

		fields, err := extractFields(content)
		if err != nil {
			return fmt.Errorf("extracting fields from template %s: %w", path, err)
		}

		ext := outputExt(name)

		relName, err := filepath.Rel(templateDir, path)
		if err != nil {
			return fmt.Errorf("getting relative path for %s: %w", path, err)
		}

		templates = append(templates, &Template{
			FileName:       relName,
			Content:        content,
			RequiredFields: fields,
			OutputExt:      ext,
		})

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("walking template dir %s: %w", templateDir, err)
	}

	return templates, nil
}

func outputExt(name string) string {
	base := strings.TrimSuffix(name, ".template.md")
	if base != name {
		return ".md"
	}
	return ".html"
}

func extractFields(content string) ([]string, error) {
	trees, err := parse.Parse("_", content, "{{", "}}", parse.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parsing template: %w", err)
	}

	seen := make(map[string]struct{})
	var fields []string

	for _, tree := range trees {
		if tree.Root != nil {
			walkNodes(tree.Root.Nodes, func(n parse.Node) {
				if fn, ok := n.(*parse.FieldNode); ok {
					if len(fn.Ident) > 0 && fn.Ident[0] != "" {
						name := fn.Ident[0]
						if _, exists := seen[name]; !exists {
							seen[name] = struct{}{}
							fields = append(fields, name)
						}
					}
				}
			})
		}
	}

	return fields, nil
}

func walkNodes(nodes []parse.Node, fn func(parse.Node)) {
	for _, n := range nodes {
		fn(n)
		switch n := n.(type) {
		case *parse.ListNode:
			walkNodes(n.Nodes, fn)
		case *parse.ActionNode:
			walkPipeNode(n.Pipe, fn)
		case *parse.IfNode:
			walkPipeNode(n.Pipe, fn)
			if n.List != nil {
				walkNodes(n.List.Nodes, fn)
			}
			if n.ElseList != nil {
				walkNodes(n.ElseList.Nodes, fn)
			}
		case *parse.RangeNode:
			walkPipeNode(n.Pipe, fn)
			if n.List != nil {
				walkNodes(n.List.Nodes, fn)
			}
			if n.ElseList != nil {
				walkNodes(n.ElseList.Nodes, fn)
			}
		case *parse.WithNode:
			walkPipeNode(n.Pipe, fn)
			if n.List != nil {
				walkNodes(n.List.Nodes, fn)
			}
			if n.ElseList != nil {
				walkNodes(n.ElseList.Nodes, fn)
			}
		case *parse.TemplateNode:
			if n.Pipe != nil {
				walkPipeNode(n.Pipe, fn)
			}
		case *parse.DefineNode:
			if n.List != nil {
				walkNodes(n.List.Nodes, fn)
			}
		}
	}
}

func walkPipeNode(pipe *parse.PipeNode, fn func(parse.Node)) {
	if pipe == nil {
		return
	}
	for _, cmd := range pipe.Cmds {
		for _, arg := range cmd.Args {
			fn(arg)
		}
	}
	for _, decl := range pipe.Decl {
		fn(decl)
	}
}
