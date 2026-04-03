// Package main implements the ldap-manager CLI entrypoint.
package main

import (
	"os"

	"github.com/y0-l0/ldap-server-helm/ldap-manager/pkg/cli"
	"github.com/y0-l0/ldap-server-helm/ldap-manager/pkg/setup"
)

func main() {
	os.Exit(cli.Main(os.Args, os.Stderr, setup.Run, cli.RunSidecar))
}
