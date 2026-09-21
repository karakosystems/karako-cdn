package server_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/karakosystems/karako-cdn/internal/server"
)

func TestExampleFilesAreServed(t *testing.T) {
	srv, err := server.New(
		server.Config{BaseFQDN: "karakosystems.com", Addr: ":80"},
		os.DirFS("../../example/public"),
	)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	for _, target := range []string{
		"/karako/logos/logo-icon-black.png",
		"/karako/logos/logo-full-navy.png",
	} {
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, target, nil))
		if rec.Code != http.StatusOK {
			t.Errorf("%s: status = %d, want 200 (loaded paths: %v)", target, rec.Code, srv.Paths())
		}
		if ct := rec.Header().Get("Content-Type"); ct != "image/png" {
			t.Errorf("%s: Content-Type = %q", target, ct)
		}
	}

	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/discover.json", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("/discover.json: status = %d, want 200", rec.Code)
	}
	if len(rec.Body.Bytes()) == 0 {
		t.Fatal("/discover.json: empty body")
	}
}
