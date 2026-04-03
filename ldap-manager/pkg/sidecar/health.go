package sidecar

import (
	"log/slog"
	"net/http"
	"time"
)

// newHealthServer returns an http.Server serving /healthz and /readyz.
func newHealthServer(addr string, check checkFunc) *http.Server {
	mux := http.NewServeMux()

	healthcheck := func(w http.ResponseWriter, r *http.Request) {
		if err := check(r.Context()); err != nil {
			slog.Warn("health check failed", "error", err)
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}

	mux.HandleFunc("/healthz", healthcheck)
	mux.HandleFunc("/readyz", healthcheck)

	return &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
}
