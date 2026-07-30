package config

import (
	"fmt"

	"github.com/romayengineer/structured-docs/pkg/structured/fsys"
	"gopkg.in/yaml.v3"
)

type Config struct {
	SchemaDir     string   `yaml:"schema_dir"`
	DataDir       string   `yaml:"data_dir"`
	TemplateDir   string   `yaml:"template_dir"`
	OutputDir     string   `yaml:"output_dir"`
	TemplateOrder []string `yaml:"template_order"`
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
