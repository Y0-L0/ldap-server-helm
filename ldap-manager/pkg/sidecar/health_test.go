package sidecar

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
)

type fakeHealthBackend struct {
	healthy bool
}

func (f *fakeHealthBackend) Check(_ context.Context) error {
	if !f.healthy {
		return errors.New("unhealthy")
	}
	return nil
}

func (f *fakeHealthBackend) Add(_ string, _ map[string][]string) error { return nil }

func (s *Unittest) TestHealthz() {
	tests := []struct {
		name       string
		healthy    bool
		wantStatus int
	}{
		{"healthy returns 200", true, http.StatusOK},
		{"unhealthy returns 503", false, http.StatusServiceUnavailable},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			backend := &fakeHealthBackend{healthy: tc.healthy}
			srv := newHealthServer(":0", backend)

			req, _ := http.NewRequestWithContext(s.T().Context(), http.MethodGet, "/healthz", nil)
			rec := httptest.NewRecorder()
			srv.Handler.ServeHTTP(rec, req)

			s.Require().Equal(tc.wantStatus, rec.Code)
		})
	}
}

func (s *Unittest) TestReadyz() {
	tests := []struct {
		name       string
		healthy    bool
		wantStatus int
	}{
		{"healthy returns 200", true, http.StatusOK},
		{"unhealthy returns 503", false, http.StatusServiceUnavailable},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			backend := &fakeHealthBackend{healthy: tc.healthy}
			srv := newHealthServer(":0", backend)

			req, _ := http.NewRequestWithContext(s.T().Context(), http.MethodGet, "/readyz", nil)
			rec := httptest.NewRecorder()
			srv.Handler.ServeHTTP(rec, req)

			s.Require().Equal(tc.wantStatus, rec.Code)
		})
	}
}
