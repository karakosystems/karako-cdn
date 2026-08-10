package server

import (
	"testing"
	"testing/fstest"
)

func TestLoadFilesMapsPathsAndETags(t *testing.T) {
	fsys := fstest.MapFS{
		"assets/karako/logos/logo.png": {Data: []byte("png-bytes")},
		"assets/karako/notes.txt":      {Data: []byte("any file is served")},
		"assets/json/.gitkeep":         {Data: nil},
	}
	files := map[string]*fileData{}
	etags := map[string]string{}

	if err := loadFiles(fsys, "assets", files, etags); err != nil {
		t.Fatalf("loadFiles: %v", err)
	}

	fd, ok := files["/karako/logos/logo.png"]
	if !ok {
		t.Fatalf("missing /karako/logos/logo.png, got: %v", files)
	}
	if string(fd.content) != "png-bytes" {
		t.Errorf("unexpected content: %q", fd.content)
	}
	wantETag := etagFor([]byte("png-bytes"))
	if etags["/karako/logos/logo.png"] != wantETag {
		t.Errorf("etag = %q, want %q", etags["/karako/logos/logo.png"], wantETag)
	}
	if _, ok := files["/karako/notes.txt"]; !ok {
		t.Error("non-image files under assets/ must be served too")
	}
	if _, ok := files["/json/.gitkeep"]; ok {
		t.Error("dotfiles must not be loaded")
	}
}
