package server

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"testing"
	"testing/fstest"
)

func newGzipTestServer(t *testing.T) (*Server, string) {
	t.Helper()
	payload := strings.Repeat(`{"key":"value","key":"value"},`, 100)
	assets := fstest.MapFS{
		"assets/json/big.json":  {Data: []byte(payload)},
		"assets/images/pix.png": {Data: []byte(strings.Repeat("png", 200))},
	}
	srv, err := New(Config{BaseFQDN: "karakosystems.com"}, assets)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return srv, payload
}

func TestGzipVariantServedWhenAccepted(t *testing.T) {
	srv, payload := newGzipTestServer(t)

	rec := do(t, srv, http.MethodGet, "/json/big.json", http.Header{"Accept-Encoding": {"gzip, br"}})
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if enc := rec.Header().Get("Content-Encoding"); enc != "gzip" {
		t.Fatalf("Content-Encoding = %q, want gzip", enc)
	}
	if vary := rec.Header().Get("Vary"); vary != "Accept-Encoding" {
		t.Errorf("Vary = %q, want Accept-Encoding", vary)
	}
	if rec.Body.Len() >= len(payload) {
		t.Errorf("gzip body (%d) not smaller than original (%d)", rec.Body.Len(), len(payload))
	}
	zr, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatalf("gzip.NewReader: %v", err)
	}
	decoded, err := io.ReadAll(zr)
	if err != nil {
		t.Fatalf("decompress: %v", err)
	}
	if string(decoded) != payload {
		t.Error("decompressed body differs from original payload")
	}
}

func TestIdentityServedWithoutAcceptEncoding(t *testing.T) {
	srv, payload := newGzipTestServer(t)

	rec := do(t, srv, http.MethodGet, "/json/big.json", nil)
	if enc := rec.Header().Get("Content-Encoding"); enc != "" {
		t.Fatalf("Content-Encoding = %q, want empty", enc)
	}
	if vary := rec.Header().Get("Vary"); vary != "Accept-Encoding" {
		t.Errorf("Vary = %q, want Accept-Encoding even on identity responses", vary)
	}
	if rec.Body.String() != payload {
		t.Error("identity body differs from original payload")
	}
}

func TestGzipVariantHasOwnETagAnd304(t *testing.T) {
	srv, _ := newGzipTestServer(t)

	identityETag := do(t, srv, http.MethodGet, "/json/big.json", nil).Header().Get("ETag")
	gzETag := do(t, srv, http.MethodGet, "/json/big.json", http.Header{"Accept-Encoding": {"gzip"}}).Header().Get("ETag")
	if identityETag == gzETag {
		t.Error("gzip variant must carry a distinct ETag")
	}

	rec := do(t, srv, http.MethodGet, "/json/big.json", http.Header{
		"Accept-Encoding": {"gzip"},
		"If-None-Match":   {gzETag},
	})
	if rec.Code != http.StatusNotModified {
		t.Fatalf("status = %d, want 304 for matching gzip ETag", rec.Code)
	}
}

func TestAlreadyCompressedTypesGetNoVariant(t *testing.T) {
	srv, _ := newGzipTestServer(t)
	rec := do(t, srv, http.MethodGet, "/images/pix.png", http.Header{"Accept-Encoding": {"gzip"}})
	if enc := rec.Header().Get("Content-Encoding"); enc != "" {
		t.Errorf("png Content-Encoding = %q, want empty (no gzip variant)", enc)
	}
}
