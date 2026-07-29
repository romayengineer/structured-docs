package watcher

import (
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/romayengineer/structured-docs/pkg/structured/compiler"
	"github.com/romayengineer/structured-docs/pkg/structured/config"
)

type WatchDirs struct {
	SchemaDir   string
	DataDir     string
	TemplateDir string
}

func Watch(cfg *config.Config, done chan struct{}) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("creating watcher: %w", err)
	}
	defer watcher.Close()

	dirs := []string{cfg.SchemaDir, cfg.DataDir, cfg.TemplateDir}
	for _, dir := range dirs {
		if err := filepath.Walk(dir, func(path string, info any, err error) error {
			if err != nil {
				return err
			}
			return watcher.Add(path)
		}); err != nil {
			return fmt.Errorf("watching directory %s: %w", dir, err)
		}
	}

	var debounceTimer *time.Timer

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			if event.Op&(fsnotify.Create|fsnotify.Write|fsnotify.Remove|fsnotify.Rename) != 0 {
				if debounceTimer != nil {
					debounceTimer.Stop()
				}
				debounceTimer = time.AfterFunc(500*time.Millisecond, func() {
					log.Println("change detected, recompiling...")
					results, err := compiler.Compile(cfg)
					if err != nil {
						log.Printf("compile error: %v", err)
						return
					}
					for _, r := range results {
						log.Printf("  wrote %s (%s)", r.OutputPath, r.Format)
					}
				})
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			log.Printf("watch error: %v", err)

		case <-done:
			return nil
		}
	}
}
