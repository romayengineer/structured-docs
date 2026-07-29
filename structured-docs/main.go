package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/romayengineer/structured-docs/pkg/structured/compiler"
	"github.com/romayengineer/structured-docs/pkg/structured/config"
	"github.com/romayengineer/structured-docs/pkg/structured/watcher"
)

func main() {
	configPath := flag.String("config", "structured.yml", "path to config file")
	watchMode := flag.Bool("watch", false, "watch for file changes and recompile")
	clean := flag.Bool("clean", false, "remove output directory before compiling")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	if *clean {
		if err := compiler.CleanOutput(cfg); err != nil {
			log.Fatalf("clean: %v", err)
		}
		log.Println("cleaned output directory")
	}

	results, err := compiler.Compile(cfg)
	if err != nil {
		log.Fatalf("compile: %v", err)
	}

	for _, r := range results {
		fmt.Printf("  %s → %s (%s)\n", r.SourcePath, r.OutputPath, r.Format)
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

		if err := watcher.Watch(cfg, done); err != nil {
			log.Fatalf("watch: %v", err)
		}
	}
}
