package karakocdn_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	karakocdn "github.com/karakosystems/karako-cdn"
	"github.com/karakosystems/karako-cdn/internal/server"
)

func TestRealEmbeddedLogosAreServed(t *testing.T) {
	srv, err := server.New(
		server.Config{BaseFQDN: "karakosystems.com", Addr: ":80"},
		karakocdn.Assets,
	)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	for _, target := range []string{
		"/images/karako/logos/logo-icon-black.png",
		"/images/karako/logos/logo-full-navy.png",
		"/images/antwan/logos/logo-full-color.png",
		"/images/antwan/logos/logo-icon-color.png",
		"/images/karako/og-image.png",
		"/images/antwan/og-image.png",
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
