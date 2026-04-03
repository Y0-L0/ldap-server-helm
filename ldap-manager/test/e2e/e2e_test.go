package e2e

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"text/template"
	"time"

	"github.com/y0-l0/ldap-server-helm/ldap-manager/pkg/ldap"
	"github.com/y0-l0/ldap-server-helm/ldap-manager/pkg/setup"
)

func TestMain(m *testing.M) {
	if os.Getenv("INTEGRATION") == "" {
		fmt.Println("skipping e2e tests; set INTEGRATION=1 to run")
		os.Exit(0)
	}
	os.Exit(m.Run())
}

const (
	baseDN  = "dc=example,dc=org"
	adminDN = "cn=admin," + baseDN
	adminPW = "admin"
)

var slapdConfTmpl = template.Must(template.New("slapd.conf").Parse(`
include   /etc/ldap/schema/core.schema
include   /etc/ldap/schema/cosine.schema
include   /etc/ldap/schema/inetorgperson.schema
include   /etc/ldap/schema/nis.schema

modulepath /usr/lib/ldap
moduleload back_mdb

database  mdb
suffix    "{{ .BaseDN }}"
rootdn    "{{ .AdminDN }}"
include   {{ .RootpwPath }}

directory {{ .DataDir }}

index     objectClass eq

access to *
    by dn.exact="{{ .AdminDN }}" write
    by * read
`))

type slapdConf struct {
	BaseDN     string
	AdminDN    string
	RootpwPath string
	DataDir    string
}

func (s *E2E) SetupSuite() {
	s.tmpDir = s.T().TempDir()
	s.dataDir = filepath.Join(s.tmpDir, "data")
	s.seedDir = filepath.Join(s.tmpDir, "seed")

	runDir := filepath.Join(s.tmpDir, "run")
	rootpwPath := filepath.Join(s.tmpDir, "rootpw.conf")
	confPath := filepath.Join(s.tmpDir, "slapd.conf")

	// Step 1: run real init code
	err := setup.Run(setup.Config{
		DataDir:    s.dataDir,
		RunDir:     runDir,
		RootpwPath: rootpwPath,
		AdminPW:    adminPW,
	})
	s.Require().NoError(err, "setup.Run")

	// Step 2: generate slapd.conf
	f, err := os.Create(confPath)
	s.Require().NoError(err)
	err = slapdConfTmpl.Execute(f, slapdConf{
		BaseDN:     baseDN,
		AdminDN:    adminDN,
		RootpwPath: rootpwPath,
		DataDir:    s.dataDir,
	})
	s.Require().NoError(err)
	s.Require().NoError(f.Close())

	// Step 3: copy seed LDIF into seedDir
	s.Require().NoError(os.MkdirAll(s.seedDir, 0o750))
	seedData, err := os.ReadFile("testdata/seed.ldif")
	s.Require().NoError(err)
	seedPath := filepath.Join(s.seedDir, "seed.ldif")
	s.Require().NoError(os.WriteFile(seedPath, seedData, 0o600)) //nolint:gosec // test path, not user input

	// Step 4: offline-load seed data before slapd starts (same as vanilla entrypoint)
	slapadd := exec.CommandContext(context.Background(), "slapadd",
		"-f", confPath,
		"-l", seedPath,
		"-c",
	)
	out, err := slapadd.CombinedOutput()
	s.Require().NoError(err, "slapadd: %s", string(out))

	// Step 5: find a free port and start slapd
	port := freePort(s.T())
	s.ldapURI = fmt.Sprintf("ldap://127.0.0.1:%d", port)

	ctx := context.Background()
	s.slapd = exec.CommandContext(ctx, "slapd",
		"-f", confPath,
		"-h", s.ldapURI,
		"-d", "0",
	)
	s.slapd.Stdout = os.Stdout
	s.slapd.Stderr = os.Stderr
	s.Require().NoError(s.slapd.Start(), "slapd start")

	// Step 6: wait for slapd readiness
	s.backend = &ldap.RealLDAP{
		URI:    s.ldapURI,
		BindDN: adminDN,
		BindPW: adminPW,
	}
	s.waitForSlapd()
}

func (s *E2E) TearDownSuite() {
	if s.slapd != nil && s.slapd.Process != nil {
		_ = s.slapd.Process.Kill()
		_ = s.slapd.Wait()
	}
}

func (s *E2E) waitForSlapd() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for {
		if err := s.backend.Check(ctx); err == nil {
			return
		}
		select {
		case <-ctx.Done():
			s.Require().Fail("slapd did not become ready within 10s")
		case <-time.After(100 * time.Millisecond):
		}
	}
}

func freePort(t *testing.T) int {
	t.Helper()
	lc := net.ListenConfig{}
	l, err := lc.Listen(context.Background(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("freePort: %v", err)
	}
	addr, ok := l.Addr().(*net.TCPAddr)
	if !ok {
		t.Fatal("freePort: unexpected address type")
	}
	port := addr.Port
	if err := l.Close(); err != nil {
		t.Fatalf("freePort close: %v", err)
	}
	return port
}
