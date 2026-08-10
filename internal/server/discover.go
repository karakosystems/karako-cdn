package server

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"sort"
	"time"
)

const discoverPath = "/discover.json"

type resourceInfo struct {
	Path        string `json:"path"`
	ContentType string `json:"contentType"`
	Size        int    `json:"size"`
	ETag        string `json:"etag"`
}

func (s *Server) buildDiscover() error {
	if _, exists := s.files[discoverPath]; exists {
		return fmt.Errorf("embedded asset %s collides with the generated discovery endpoint", discoverPath)
	}

	resources := make([]resourceInfo, 0, len(s.files))
	for p, fd := range s.files {
		resources = append(resources, resourceInfo{
			Path:        p,
			ContentType: contentTypeFor(p),
			Size:        len(fd.content),
			ETag:        s.etags[p],
		})
	}
	sort.Slice(resources, func(i, j int) bool { return resources[i].Path < resources[j].Path })

	payload, err := json.MarshalIndent(map[string][]resourceInfo{"resources": resources}, "", "  ")
	if err != nil {
		return err
	}

	s.files[discoverPath] = &fileData{content: payload, modTime: time.Now()}
	s.etags[discoverPath] = fmt.Sprintf("\"%x\"", md5.Sum(payload))
	return nil
}
