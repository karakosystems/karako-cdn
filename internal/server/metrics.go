package server

import (
	"net/http"
	"strconv"
	"strings"
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
	counters := []struct {
		class string
		value uint64
	}{
		{"ok", s.metrics.ok.Load()},
		{"not_modified", s.metrics.notModified.Load()},
		{"redirect", s.metrics.redirects.Load()},
		{"not_found", s.metrics.notFound.Load()},
	}

	var b strings.Builder
	b.WriteString("# HELP karako_cdn_responses_total Responses served, by class.\n")
	b.WriteString("# TYPE karako_cdn_responses_total counter\n")
	for _, c := range counters {
		b.WriteString(`karako_cdn_responses_total{class="`)
		b.WriteString(c.class)
		b.WriteString(`"} `)
		b.WriteString(strconv.FormatUint(c.value, 10))
		b.WriteByte('\n')
	}
	b.WriteString("# HELP karako_cdn_resources Number of resources currently served.\n")
	b.WriteString("# TYPE karako_cdn_resources gauge\n")
	b.WriteString("karako_cdn_resources ")
	b.WriteString(strconv.Itoa(len(s.files)))
	b.WriteByte('\n')

	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	//nosemgrep: go.lang.security.audit.xss.no-direct-write-to-responsewriter.no-direct-write-to-responsewriter
	_, _ = w.Write([]byte(b.String()))
}
