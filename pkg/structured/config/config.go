package config

import (
	"fmt"

	"github.com/romayengineer/structured-docs/pkg/structured/fsys"
	"gopkg.in/yaml.v3"
)

type Config struct {
	SchemaDir     string    `yaml:"schema_dir"`
	DataDir       string    `yaml:"data_dir"`
	TemplateDir   string    `yaml:"template_dir"`
	OutputDir     string    `yaml:"output_dir"`
	TemplateOrder   []string  `yaml:"template_order"`
	Validate        *Validate `yaml:"validate"`
	GenerateReadme  *bool     `yaml:"generate_readme"`
}

type Validate struct {
	LinesBetweenHeaders *HeaderSpacing `yaml:"lines_between_headers"`
	FileName            *FileNameRule  `yaml:"file_name"`
}

type FileNameRule struct {
	Style             string `yaml:"style"`
	Pattern           string `yaml:"pattern"`
	AllowLeadingDigit *bool  `yaml:"allow_leading_digit"`
}

type HeaderSpacing struct {
	H1      int `yaml:"h1"`
	H2      int `yaml:"h2"`
	H3      int `yaml:"h3"`
	H4      int `yaml:"h4"`
	Default int `yaml:"default"`
}

func Default() *Config {
	return &Config{
		SchemaDir:   "schema",
		DataDir:     "data",
		TemplateDir: "templates",
		OutputDir:   "output",
	}
}

func Load(fsys fsys.FS, path string) (*Config, error) {
	cfg := Default()

	b, err := fsys.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	if err := yaml.Unmarshal(b, cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	if len(cfg.TemplateOrder) == 0 {
		return nil, fmt.Errorf("template_order is required in config")
	}

	return cfg, nil
}
