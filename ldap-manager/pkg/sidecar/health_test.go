package sidecar

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
)

type fakeLDAPChecker struct {
	healthy bool
}

func (f *fakeLDAPChecker) Check(_ context.Context) error {
	if !f.healthy {
		return errors.New("unhealthy")
	}
	return nil
}

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
			checker := &fakeLDAPChecker{healthy: tc.healthy}
			srv := newHealthServer(":0", checker)

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
			checker := &fakeLDAPChecker{healthy: tc.healthy}
			srv := newHealthServer(":0", checker)

			req, _ := http.NewRequestWithContext(s.T().Context(), http.MethodGet, "/readyz", nil)
			rec := httptest.NewRecorder()
			srv.Handler.ServeHTTP(rec, req)

			s.Require().Equal(tc.wantStatus, rec.Code)
		})
	}
}
