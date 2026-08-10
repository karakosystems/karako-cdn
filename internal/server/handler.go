package server

import (
	"bytes"
	"mime"
	"net/http"
	"path"
	"path/filepath"
	"strings"
)

const missCacheControl = "public, max-age=300"

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	switch r.URL.Path {
	case "/":
		s.redirect(w, r)
		return
	case "/health":
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("healthy\n"))
		return
	case metricsPath:
		s.serveMetrics(w)
		return
	case "/favicon.ico":
		s.notFound(w, r)
		return
	}

	fd, exists := s.files[r.URL.Path]
	if !exists {
		if path.Ext(r.URL.Path) != "" {
			s.notFound(w, r)
		} else {
			s.redirect(w, r)
		}
		return
	}

	if ct := contentTypeFor(r.URL.Path); ct != "" {
		w.Header().Set("Content-Type", ct)
	}
	w.Header().Set("Cache-Control", cacheControlFor(r.URL.Path))

	content, etag := fd.content, s.etags[r.URL.Path]
	if gz, ok := s.gzips[r.URL.Path]; ok {
		w.Header().Set("Vary", "Accept-Encoding")
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			content, etag = gz.content, s.gzipETags[r.URL.Path]
			w.Header().Set("Content-Encoding", "gzip")
		}
	}

	w.Header().Set("ETag", etag)
	if r.Header.Get("If-None-Match") == etag {
		s.metrics.notModified.Add(1)
		w.WriteHeader(http.StatusNotModified)
		return
	}

	s.metrics.ok.Add(1)
	http.ServeContent(w, r, filepath.Base(r.URL.Path), fd.modTime, bytes.NewReader(content))
}

func (s *Server) redirect(w http.ResponseWriter, r *http.Request) {
	s.metrics.redirects.Add(1)
	w.Header().Set("Cache-Control", missCacheControl)
	http.Redirect(w, r, "https://"+s.cfg.BaseFQDN, http.StatusFound)
}

func (s *Server) notFound(w http.ResponseWriter, r *http.Request) {
	s.metrics.notFound.Add(1)
	w.Header().Set("Cache-Control", missCacheControl)
	http.NotFound(w, r)
}

func cacheControlFor(p string) string {
	if p == discoverPath {
		return "public, must-revalidate, max-age=600"
	}
	return "public, max-age=31536000, immutable"
}

var contentTypes = map[string]string{
	".json":  "application/json",
	".png":   "image/png",
	".jpg":   "image/jpeg",
	".jpeg":  "image/jpeg",
	".gif":   "image/gif",
	".webp":  "image/webp",
	".svg":   "image/svg+xml",
	".bmp":   "image/bmp",
	".ico":   "image/x-icon",
	".css":   "text/css; charset=utf-8",
	".js":    "text/javascript; charset=utf-8",
	".txt":   "text/plain; charset=utf-8",
	".pdf":   "application/pdf",
	".woff2": "font/woff2",
	".woff":  "font/woff",
	".ttf":   "font/ttf",
	".otf":   "font/otf",
	".mp4":   "video/mp4",
	".webm":  "video/webm",
}

func contentTypeFor(p string) string {
	ext := strings.ToLower(path.Ext(p))
	if ct, ok := contentTypes[ext]; ok {
		return ct
	}
	return mime.TypeByExtension(ext)
}
