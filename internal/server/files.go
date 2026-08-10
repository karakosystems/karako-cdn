package server

import (
	"crypto/md5"
	"fmt"
	"io/fs"
	"path"
	"strings"
	"time"
)

type fileData struct {
	content []byte
	modTime time.Time
}

func loadFiles(fsys fs.FS, root string, files map[string]*fileData, etags map[string]string) error {
	now := time.Now()
	return fs.WalkDir(fsys, root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		name := path.Base(p)
		if strings.HasPrefix(name, ".") {
			return nil
		}
		content, err := fs.ReadFile(fsys, p)
		if err != nil {
			return fmt.Errorf("reading %s: %w", p, err)
		}
		urlPath := strings.TrimPrefix(p, root)
		files[urlPath] = &fileData{content: content, modTime: now}
		etags[urlPath] = fmt.Sprintf("\"%x\"", md5.Sum(content))
		return nil
	})
}
