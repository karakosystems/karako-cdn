package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
)

func newTestServer(t *testing.T) *Server {
	t.Helper()
	assets := fstest.MapFS{
		"json/config/app.json":  {Data: []byte(`{"ok":true}`)},
		"json/.gitkeep":         {Data: nil},
		"karako/logos/logo.png": {Data: []byte("png-bytes")},
	}
	srv, err := New(Config{BaseFQDN: "karakosystems.com", Addr: ":80"}, assets)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return srv
}

func do(t *testing.T, srv *Server, method, target string, header http.Header) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(method, target, nil)
	for k, vs := range header {
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}
	srv.ServeHTTP(rec, req)
	return rec
}

func TestServesEmbeddedFiles(t *testing.T) {
	srv := newTestServer(t)

	rec := do(t, srv, http.MethodGet, "/karako/logos/logo.png", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "image/png" {
		t.Errorf("Content-Type = %q, want image/png", ct)
	}
	if rec.Header().Get("ETag") == "" {
		t.Error("missing ETag")
	}
	if cc := rec.Header().Get("Cache-Control"); cc != "public, max-age=31536000, immutable" {
		t.Errorf("Cache-Control = %q", cc)
	}
	if rec.Body.String() != "png-bytes" {
		t.Errorf("unexpected body: %q", rec.Body.String())
	}

	rec = do(t, srv, http.MethodGet, "/json/config/app.json", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("JSON status = %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("JSON Content-Type = %q", ct)
	}
	if rec.Body.String() != `{"ok":true}` {
		t.Errorf("unexpected JSON body: %q", rec.Body.String())
	}
}

func TestETagReturns304(t *testing.T) {
	srv := newTestServer(t)
	etag := do(t, srv, http.MethodGet, "/karako/logos/logo.png", nil).Header().Get("ETag")
	rec := do(t, srv, http.MethodGet, "/karako/logos/logo.png", http.Header{"If-None-Match": {etag}})
	if rec.Code != http.StatusNotModified {
		t.Fatalf("status = %d, want 304", rec.Code)
	}
}

func TestRedirectsBarePaths(t *testing.T) {
	srv := newTestServer(t)
	for _, target := range []string{"/", "/unknown", "/karako/logos"} {
		rec := do(t, srv, http.MethodGet, target, nil)
		if rec.Code != http.StatusFound {
			t.Errorf("%s: status = %d, want 302", target, rec.Code)
		}
		if loc := rec.Header().Get("Location"); loc != "https://karakosystems.com" {
			t.Errorf("%s: Location = %q, want https://karakosystems.com", target, loc)
		}
		if cc := rec.Header().Get("Cache-Control"); cc != "public, max-age=300" {
			t.Errorf("%s: Cache-Control = %q, want public, max-age=300", target, cc)
		}
	}
}

func TestUnknownAssetPathsReturn404(t *testing.T) {
	srv := newTestServer(t)
	for _, target := range []string{"/unknown.json", "/karako/logos/missing.png"} {
		rec := do(t, srv, http.MethodGet, target, nil)
		if rec.Code != http.StatusNotFound {
			t.Errorf("%s: status = %d, want 404", target, rec.Code)
		}
		if cc := rec.Header().Get("Cache-Control"); cc != "public, max-age=300" {
			t.Errorf("%s: Cache-Control = %q, want public, max-age=300", target, cc)
		}
	}
}

func TestHealth(t *testing.T) {
	srv := newTestServer(t)
	rec := do(t, srv, http.MethodGet, "/health", nil)
	if rec.Code != http.StatusOK || rec.Body.String() != "healthy\n" {
		t.Fatalf("health: %d %q", rec.Code, rec.Body.String())
	}
}

func TestCORSPreflight(t *testing.T) {
	srv := newTestServer(t)
	rec := do(t, srv, http.MethodOptions, "/karako/logos/logo.png", nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("OPTIONS: status = %d, want 204", rec.Code)
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("missing Access-Control-Allow-Origin")
	}
}

func TestFaviconNotFound(t *testing.T) {
	srv := newTestServer(t)
	if rec := do(t, srv, http.MethodGet, "/favicon.ico", nil); rec.Code != http.StatusNotFound {
		t.Fatalf("favicon: status = %d, want 404", rec.Code)
	}
}

func TestDotfilesNotServed(t *testing.T) {
	srv := newTestServer(t)
	rec := do(t, srv, http.MethodGet, "/json/.gitkeep", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("/json/.gitkeep: status = %d, want 404 (not loaded, no content)", rec.Code)
	}
}

func TestContentTypeFallsBackToBuiltinTable(t *testing.T) {
	if ct := contentTypeFor("/fonts/brand.woff2"); ct != "font/woff2" {
		t.Errorf("woff2 = %q, want font/woff2", ct)
	}
	if ct := contentTypeFor("/styles/site.css"); ct != "text/css; charset=utf-8" {
		t.Errorf("css = %q, want text/css; charset=utf-8", ct)
	}
}
