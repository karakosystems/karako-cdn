package server

import (
	"fmt"
	"io/fs"
	"sort"
)

type Config struct {
	BaseFQDN string
	CDNFQDN  string
	Addr     string
}

type Server struct {
	cfg       Config
	files     map[string]*fileData
	etags     map[string]string
	gzips     map[string]*fileData
	gzipETags map[string]string
	metrics   metrics
}

func New(cfg Config, assets fs.FS) (*Server, error) {
	if cfg.CDNFQDN == "" {
		cfg.CDNFQDN = "cdn." + cfg.BaseFQDN
	}
	s := &Server{
		cfg:       cfg,
		files:     make(map[string]*fileData),
		etags:     make(map[string]string),
		gzips:     make(map[string]*fileData),
		gzipETags: make(map[string]string),
	}
	if err := loadFiles(assets, s.files, s.etags); err != nil {
		return nil, err
	}
	if err := s.buildDiscover(); err != nil {
		return nil, err
	}
	if _, exists := s.files[metricsPath]; exists {
		return nil, fmt.Errorf("asset %s collides with the metrics endpoint", metricsPath)
	}
	if err := s.compressAll(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Server) Paths() []string {
	paths := make([]string, 0, len(s.files))
	for p := range s.files {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	return paths
}
