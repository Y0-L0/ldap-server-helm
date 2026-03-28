package sidecar

import (
	"context"
	"fmt"

	"github.com/go-ldap/ldap/v3"
)

// RealLDAP implements LDAPChecker and LDAPSeeder using a real LDAP connection.
type RealLDAP struct {
	URI    string
	BindDN string
	BindPW string
}

// Check connects to slapd and performs a root DSE search.
func (r *RealLDAP) Check(_ context.Context) error {
	conn, err := ldap.DialURL(r.URI)
	if err != nil {
		return fmt.Errorf("connecting to %s: %w", r.URI, err)
	}
	defer conn.Close()

	req := ldap.NewSearchRequest(
		"",
		ldap.ScopeBaseObject,
		ldap.NeverDerefAliases,
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
	conn, err := ldap.DialURL(r.URI)
	if err != nil {
		return fmt.Errorf("connecting to %s: %w", r.URI, err)
	}
	defer conn.Close()

	if err := conn.Bind(r.BindDN, r.BindPW); err != nil {
		return fmt.Errorf("binding as %s: %w", r.BindDN, err)
	}

	addReq := ldap.NewAddRequest(dn, nil)
	for key, values := range attrs {
		addReq.Attribute(key, values)
	}

	if err := conn.Add(addReq); err != nil {
		return fmt.Errorf("adding %s: %w", dn, err)
	}

	return nil
}
