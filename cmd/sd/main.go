package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/romayengineer/structured-docs/pkg/structured/compiler"
	"github.com/romayengineer/structured-docs/pkg/structured/config"
	"github.com/romayengineer/structured-docs/pkg/structured/fsys"
	"github.com/romayengineer/structured-docs/pkg/structured/watcher"
)

const sdocsDirName = ".sd"

var version = "dev" // set via -ldflags at build time

func findSdocsDir() (string, bool) {
	wd, err := os.Getwd()
	if err != nil {
		return "", false
	}
	dir := wd
	for {
		// Check for .sd/structured.yml (created by sd init)
		candidate := filepath.Join(dir, sdocsDirName, "structured.yml")
		if _, err := os.Stat(candidate); err == nil {
			return filepath.Join(dir, sdocsDirName), true
		}
		// Check for structured.yml directly (custom init dir or manual setup)
		candidate = filepath.Join(dir, "structured.yml")
		if _, err := os.Stat(candidate); err == nil {
			return dir, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", false
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "init" {
		dir := sdocsDirName
		if len(os.Args) > 2 {
			dir = os.Args[2]
		}
		if err := initProject(dir); err != nil {
			log.Fatalf("init: %v", err)
		}
		return
	}

	showVersion := flag.Bool("version", false, "print version and exit")
	configPath := flag.String("config", "structured.yml", "path to config file")
	watchMode := flag.Bool("watch", false, "watch for file changes and recompile")
	clean := flag.Bool("clean", false, "remove output directory before compiling")
	flag.Parse()

	if *showVersion {
		fmt.Println("sd version", version)
		os.Exit(0)
	}

	var fs fsys.FS = fsys.OS{}

	// Auto-detect sdocs/ directory when using default config path
	if *configPath == "structured.yml" {
		if sdocsDir, ok := findSdocsDir(); ok {
			if err := os.Chdir(sdocsDir); err != nil {
				log.Fatalf("chdir to %s: %v", sdocsDir, err)
			}
		}
	}

	cfg, err := config.Load(fs, *configPath)
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	if *clean {
		if err := compiler.CleanOutput(fs, cfg); err != nil {
			log.Fatalf("clean: %v", err)
		}
		log.Println("cleaned output directory")
	}

	results, err := compiler.Compile(fs, cfg)
	if err != nil {
		log.Fatalf("compile: %v", err)
	}

	for _, r := range results {
		fmt.Printf("  %s → %s (%s)\n", filepath.Join(cfg.DataDir, r.SourcePath), r.OutputPath, r.Format)
	}
	fmt.Printf("compiled %d file(s)\n", len(results))

	if *watchMode {
		log.Println("watching for changes...")
		done := make(chan struct{})
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt)

		go func() {
			<-sigCh
			close(done)
		}()

		if err := watcher.Watch(fs, cfg, done); err != nil {
			log.Fatalf("watch: %v", err)
		}
	}
}
