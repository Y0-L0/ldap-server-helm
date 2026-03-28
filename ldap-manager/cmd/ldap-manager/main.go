// Package main implements the ldap-manager CLI entrypoint.
package main

import (
	"io"
	"os"

	"github.com/y0-l0/ldap-server-helm/ldap-manager/pkg/cli"
)

var (
	osExit           = os.Exit
	stdout io.Writer = os.Stdout
	stderr io.Writer = os.Stderr
)

func main() {
	osExit(cli.Main(os.Args, stdout, stderr))
}
