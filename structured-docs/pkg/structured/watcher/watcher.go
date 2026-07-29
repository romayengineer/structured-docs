package watcher

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/romayengineer/structured-docs/pkg/structured/compiler"
	"github.com/romayengineer/structured-docs/pkg/structured/config"
	"github.com/romayengineer/structured-docs/pkg/structured/fsys"
)

func Watch(fsys fsys.FS, cfg *config.Config, done chan struct{}) error {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("creating watcher: %w", err)
	}
	defer w.Close()

	dirs := []string{cfg.SchemaDir, cfg.DataDir, cfg.TemplateDir}
	for _, dir := range dirs {
		if err := fsys.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return w.Add(path)
			}
			return nil
		}); err != nil {
			return fmt.Errorf("watching directory %s: %w", dir, err)
		}
	}

	var debounceTimer *time.Timer

	for {
		select {
		case event, ok := <-w.Events:
			if !ok {
				return nil
			}
			if event.Op&(fsnotify.Create|fsnotify.Write|fsnotify.Remove|fsnotify.Rename) != 0 {
				if debounceTimer != nil {
					debounceTimer.Stop()
				}
				debounceTimer = time.AfterFunc(500*time.Millisecond, func() {
					log.Println("change detected, recompiling...")
					results, err := compiler.Compile(fsys, cfg)
					if err != nil {
						log.Printf("compile error: %v", err)
						return
					}
					for _, r := range results {
						log.Printf("  wrote %s (%s)", r.OutputPath, r.Format)
					}
				})
			}

		case err, ok := <-w.Errors:
			if !ok {
				return nil
			}
			log.Printf("watch error: %v", err)

		case <-done:
			return nil
		}
	}
}
