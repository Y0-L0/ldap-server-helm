// Package ldap implements the LDAP adapter using go-ldap.
package ldap

import (
	"context"
	"fmt"

	goldap "github.com/go-ldap/ldap/v3"
)

// RealLDAP implements sidecar.Backend using a real LDAP connection.
type RealLDAP struct {
	URI    string
	BindDN string
	BindPW string
}

// Check connects to slapd and performs a root DSE search.
func (r *RealLDAP) Check(_ context.Context) error {
	conn, err := goldap.DialURL(r.URI)
	if err != nil {
		return fmt.Errorf("connecting to %s: %w", r.URI, err)
	}
	defer conn.Close()

	req := goldap.NewSearchRequest(
		"",
		goldap.ScopeBaseObject,
		goldap.NeverDerefAliases,
		0, 0, false,
		"(objectClass=*)",
		[]string{"namingContexts"},
		nil,
	)

	_, err = conn.Search(req)
	if err != nil {
		return fmt.Errorf("root DSE search: %w", err)
	}

	return nil
}

// Add connects to slapd, binds, and adds an entry.
func (r *RealLDAP) Add(dn string, attrs map[string][]string) error {
	conn, err := goldap.DialURL(r.URI)
	if err != nil {
		return fmt.Errorf("connecting to %s: %w", r.URI, err)
	}
	defer conn.Close()

	if err := conn.Bind(r.BindDN, r.BindPW); err != nil {
		return fmt.Errorf("binding as %s: %w", r.BindDN, err)
	}

	addReq := goldap.NewAddRequest(dn, nil)
	for key, values := range attrs {
		addReq.Attribute(key, values)
	}

	if err := conn.Add(addReq); err != nil {
		return fmt.Errorf("adding %s: %w", dn, err)
	}

	return nil
}
