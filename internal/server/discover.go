package server

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
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

	base := "https://" + s.cfg.CDNFQDN
	resources := make([]resourceInfo, 0, len(s.files))
	for p, fd := range s.files {
		resources = append(resources, resourceInfo{
			Path:        base + p,
			ContentType: contentTypeFor(p),
			Size:        len(fd.content),
			ETag:        strings.Trim(s.etags[p], `"`),
		})
	}
	sort.Slice(resources, func(i, j int) bool { return resources[i].Path < resources[j].Path })

	payload, err := json.Marshal(map[string][]resourceInfo{"resources": resources})
	if err != nil {
		return err
	}

	s.files[discoverPath] = &fileData{content: payload, modTime: time.Now()}
	s.etags[discoverPath] = etagFor(payload)
	return nil
}
