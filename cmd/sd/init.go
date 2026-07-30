package main

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed skeleton
var skeletonFS embed.FS

func initProject(dir string) error {
	walkDir := "skeleton"
	err := fs.WalkDir(skeletonFS, walkDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(walkDir, path)
		target := filepath.Join(dir, rel)

		if d.IsDir() {
			return os.MkdirAll(target, 0755)
		}

		src, err := skeletonFS.Open(path)
		if err != nil {
			return fmt.Errorf("opening %s: %w", path, err)
		}
		defer src.Close()

		dst, err := os.Create(target)
		if err != nil {
			return fmt.Errorf("creating %s: %w", target, err)
		}
		defer dst.Close()

		if _, err := io.Copy(dst, src); err != nil {
			return fmt.Errorf("writing %s: %w", target, err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("copying skeleton: %w", err)
	}

	fmt.Printf("created project in %s\n", dir)
	fmt.Println("  run `sd` from that directory to generate the documentation")
	return nil
}
