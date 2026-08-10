package server

import (
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"fmt"
	"strings"
)

const gzipMinSize = 256

func compressible(contentType string) bool {
	switch {
	case strings.HasPrefix(contentType, "text/"),
		strings.HasPrefix(contentType, "application/json"),
		strings.HasPrefix(contentType, "application/xml"),
		strings.HasPrefix(contentType, "application/pdf"),
		strings.HasPrefix(contentType, "image/svg+xml"):
		return true
	}
	return false
}

func (s *Server) compressAll() error {
	for p, fd := range s.files {
		if len(fd.content) < gzipMinSize || !compressible(contentTypeFor(p)) {
			continue
		}
		var buf bytes.Buffer
		zw, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
		if err != nil {
			return err
		}
		if _, err := zw.Write(fd.content); err != nil {
			return fmt.Errorf("compressing %s: %w", p, err)
		}
		if err := zw.Close(); err != nil {
			return fmt.Errorf("compressing %s: %w", p, err)
		}
		if buf.Len() > len(fd.content)*9/10 {
			continue
		}
		s.gzips[p] = &fileData{content: buf.Bytes(), modTime: fd.modTime}
		s.gzipETags[p] = fmt.Sprintf("\"%x\"", md5.Sum(buf.Bytes()))
	}
	return nil
}
