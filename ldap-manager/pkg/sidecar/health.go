// Package sidecar implements the ldap-manager sidecar logic: health checks and LDAP seeding.
package sidecar

import (
	"context"
	"net/http"
	"time"
)

// LDAPChecker tests whether slapd is responsive.
type LDAPChecker interface {
	Check(ctx context.Context) error
}

// NewHealthServer returns an http.Server serving /healthz and /readyz.
func NewHealthServer(addr string, checker LDAPChecker) *http.Server {
	mux := http.NewServeMux()

	check := func(w http.ResponseWriter, r *http.Request) {
		if err := checker.Check(r.Context()); err != nil {
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
