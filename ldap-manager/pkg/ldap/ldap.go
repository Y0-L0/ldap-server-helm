// Package ldap implements the LDAP adapter using go-ldap.
package ldap

import (
	"context"
	"errors"
	"fmt"
	"sync"

	goldap "github.com/go-ldap/ldap/v3"
)

// RealLDAP implements sidecar.Backend using a persistent LDAP connection.
type RealLDAP struct {
	URI    string
	BindDN string
	BindPW string

	mu   sync.Mutex
	conn *goldap.Conn
}

func (r *RealLDAP) ensureConn() error {
	if r.conn != nil {
		return nil
	}
	conn, err := goldap.DialURL(r.URI)
	if err != nil {
		return fmt.Errorf("connecting to %s: %w", r.URI, err)
	}
	r.conn = conn
	return nil
}

// isConnError returns true if the error is a network/connection-level failure
// rather than an LDAP protocol error. On connection errors the connection
// should be discarded so the next call reconnects.
func isConnError(err error) bool {
	var ldapErr *goldap.Error
	return !errors.As(err, &ldapErr)
}

func (r *RealLDAP) handleErr(err error) error {
	if isConnError(err) {
		r.conn.Close()
		r.conn = nil
	}
	return err
}

// Check performs a root DSE search to verify slapd is reachable.
func (r *RealLDAP) Check(_ context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.ensureConn(); err != nil {
		return err
	}

	req := goldap.NewSearchRequest(
		"",
		goldap.ScopeBaseObject,
		goldap.NeverDerefAliases,
		0, 0, false,
		"(objectClass=*)",
		[]string{"namingContexts"},
		nil,
	)

	if _, err := r.conn.Search(req); err != nil {
		return r.handleErr(fmt.Errorf("root DSE search: %w", err))
	}

	return nil
}

// Add binds and adds an entry to the directory.
func (r *RealLDAP) Add(dn string, attrs map[string][]string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.ensureConn(); err != nil {
		return err
	}

	if err := r.conn.Bind(r.BindDN, r.BindPW); err != nil {
		return r.handleErr(fmt.Errorf("binding as %s: %w", r.BindDN, err))
	}

	addReq := goldap.NewAddRequest(dn, nil)
	for key, values := range attrs {
		addReq.Attribute(key, values)
	}

	if err := r.conn.Add(addReq); err != nil {
		return r.handleErr(fmt.Errorf("adding %s: %w", dn, err))
	}

	return nil
}
