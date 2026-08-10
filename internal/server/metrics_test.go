package server

import (
	"net/http"
	"strings"
	"testing"
	"testing/fstest"
)

func TestMetricsCountsResponseClasses(t *testing.T) {
	srv := newTestServer(t)

	etag := do(t, srv, http.MethodGet, "/karako/logos/logo.png", nil).Header().Get("ETag")
	do(t, srv, http.MethodGet, "/karako/logos/logo.png", http.Header{"If-None-Match": {etag}})
	do(t, srv, http.MethodGet, "/", nil)
	do(t, srv, http.MethodGet, "/missing.png", nil)

	rec := do(t, srv, http.MethodGet, metricsPath, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	for _, want := range []string{
		`karako_cdn_responses_total{class="ok"} 1`,
		`karako_cdn_responses_total{class="not_modified"} 1`,
		`karako_cdn_responses_total{class="redirect"} 1`,
		`karako_cdn_responses_total{class="not_found"} 1`,
		"karako_cdn_resources 3",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("metrics output missing %q\n%s", want, body)
		}
	}
}

func TestMetricsCollisionFailsAtStartup(t *testing.T) {
	assets := fstest.MapFS{
		"assets/metrics": {Data: []byte("shadowing")},
	}
	if _, err := New(Config{BaseFQDN: "karakosystems.com"}, assets); err == nil {
		t.Fatal("New must fail when an embedded asset collides with /metrics")
	}
}
