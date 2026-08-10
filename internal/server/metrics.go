package server

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

const metricsPath = "/metrics"

type metrics struct {
	ok          atomic.Uint64
	notModified atomic.Uint64
	redirects   atomic.Uint64
	notFound    atomic.Uint64
}

func (s *Server) serveMetrics(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	fmt.Fprintln(w, "# HELP karako_cdn_responses_total Responses served, by class.")
	fmt.Fprintln(w, "# TYPE karako_cdn_responses_total counter")
	fmt.Fprintf(w, "karako_cdn_responses_total{class=\"ok\"} %d\n", s.metrics.ok.Load())
	fmt.Fprintf(w, "karako_cdn_responses_total{class=\"not_modified\"} %d\n", s.metrics.notModified.Load())
	fmt.Fprintf(w, "karako_cdn_responses_total{class=\"redirect\"} %d\n", s.metrics.redirects.Load())
	fmt.Fprintf(w, "karako_cdn_responses_total{class=\"not_found\"} %d\n", s.metrics.notFound.Load())
	fmt.Fprintln(w, "# HELP karako_cdn_resources Number of resources currently served.")
	fmt.Fprintln(w, "# TYPE karako_cdn_resources gauge")
	fmt.Fprintf(w, "karako_cdn_resources %d\n", len(s.files))
}
