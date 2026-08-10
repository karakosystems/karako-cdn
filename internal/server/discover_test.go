package server

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"testing"
	"testing/fstest"
)

type discoverPayload struct {
	Resources []struct {
		Path        string `json:"path"`
		ContentType string `json:"contentType"`
		Size        int    `json:"size"`
		ETag        string `json:"etag"`
	} `json:"resources"`
}

func TestDiscoverListsResourcesWithMetadata(t *testing.T) {
	srv := newTestServer(t)

	rec := do(t, srv, http.MethodGet, "/discover.json", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}

	var payload discoverPayload
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(payload.Resources) != 2 {
		t.Fatalf("resources = %d, want 2 (logo.png and app.json)", len(payload.Resources))
	}
	if !sort.SliceIsSorted(payload.Resources, func(i, j int) bool {
		return payload.Resources[i].Path < payload.Resources[j].Path
	}) {
		t.Error("resources must be sorted by path")
	}
	for _, res := range payload.Resources {
		if res.Path == "/discover.json" {
			t.Error("discover.json must not list itself")
		}
	}

	logo := payload.Resources[1] // sorted: /json/... < /karako/...
	if logo.Path != "https://cdn.karakosystems.com/karako/logos/logo.png" {
		t.Fatalf("first path = %q, want absolute URL", logo.Path)
	}
	if logo.ContentType != "image/png" {
		t.Errorf("contentType = %q", logo.ContentType)
	}
	if logo.Size != len("png-bytes") {
		t.Errorf("size = %d, want %d", logo.Size, len("png-bytes"))
	}
	wantETag := strings.Trim(do(t, srv, http.MethodGet, "/karako/logos/logo.png", nil).Header().Get("ETag"), `"`)
	if logo.ETag != wantETag {
		t.Errorf("etag = %q, want %q (served ETag without quotes)", logo.ETag, wantETag)
	}
}

func TestDiscoverPayloadIsCompactJSON(t *testing.T) {
	srv := newTestServer(t)
	body := do(t, srv, http.MethodGet, "/discover.json", nil).Body.String()
	if strings.Contains(body, "\n") {
		t.Error("discover.json must be compact (no indentation)")
	}
}

func TestDiscoverKeepsShortCacheLifetime(t *testing.T) {
	srv := newTestServer(t)
	rec := do(t, srv, http.MethodGet, "/discover.json", nil)
	if cc := rec.Header().Get("Cache-Control"); cc != "public, must-revalidate, max-age=600" {
		t.Errorf("Cache-Control = %q, want public, must-revalidate, max-age=600", cc)
	}
}

func TestDiscoverSupports304(t *testing.T) {
	srv := newTestServer(t)
	etag := do(t, srv, http.MethodGet, "/discover.json", nil).Header().Get("ETag")
	if etag == "" {
		t.Fatal("discover.json must carry an ETag")
	}
	rec := do(t, srv, http.MethodGet, "/discover.json", http.Header{"If-None-Match": {etag}})
	if rec.Code != http.StatusNotModified {
		t.Fatalf("status = %d, want 304", rec.Code)
	}
}

func TestDiscoverCollisionFailsAtStartup(t *testing.T) {
	assets := fstest.MapFS{
		"assets/discover.json": {Data: []byte(`{"fake":true}`)},
	}
	if _, err := New(Config{BaseFQDN: "karakosystems.com", Addr: ":80"}, assets); err == nil {
		t.Fatal("New must fail when an embedded asset collides with /discover.json")
	}
}
