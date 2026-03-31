package sidecar

import (
	"log/slog"
	"net/http"
	"time"
)

// newHealthServer returns an http.Server serving /healthz and /readyz.
func newHealthServer(addr string, backend Backend) *http.Server {
	mux := http.NewServeMux()

	check := func(w http.ResponseWriter, r *http.Request) {
		if err := backend.Check(r.Context()); err != nil {
			slog.Warn("health check failed", "error", err)
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}

	mux.HandleFunc("/healthz", check)
	mux.HandleFunc("/readyz", check)

	return &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
}
