// Package main implements the ldap-manager CLI entrypoint.
package main

import (
	"os"

	"github.com/y0-l0/ldap-server-helm/ldap-manager/pkg/cli"
	initpkg "github.com/y0-l0/ldap-server-helm/ldap-manager/pkg/init"
)

func main() {
	os.Exit(cli.Main(os.Args, os.Stderr, initpkg.Run, cli.RunSidecar))
}
