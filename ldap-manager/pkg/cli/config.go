package cli

import (
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/y0-l0/ldap-server-helm/ldap-manager/pkg/setup"
	"github.com/y0-l0/ldap-server-helm/ldap-manager/pkg/sidecar"
)

// Config is the top-level ldap-manager configuration, loaded from a JSON file.
type Config struct {
	LogLevel   string           `json:"logLevel"`
	DataDir    string           `json:"dataDir"`
	RunDir     string           `json:"runDir"`
	RootpwPath string           `json:"rootpwPath"`
	Connection ConnectionConfig `json:"connection"`
	Sidecar    SidecarConfig    `json:"sidecar"`
}

// ConnectionConfig holds LDAP connection settings used by the sidecar.
type ConnectionConfig struct {
	URI    string `json:"uri"`
	BaseDN string `json:"baseDN"`
	BindDN string `json:"bindDN"`
}

// SidecarConfig holds settings exclusive to the sidecar subcommand.
type SidecarConfig struct {
	HealthAddr string `json:"healthAddr"`
	SeedDir    string `json:"seedDir"`
}

var errMissingAdminPW = errors.New("LDAP_ADMIN_PW is required")

func (cfg Config) toSetup() (setup.Config, error) {
	adminPW := os.Getenv("LDAP_ADMIN_PW")
	if adminPW == "" {
		return setup.Config{}, errMissingAdminPW
	}

	return setup.Config{
		DataDir:    cfg.DataDir,
		RunDir:     cfg.RunDir,
		RootpwPath: cfg.RootpwPath,
		AdminPW:    adminPW,
	}, nil
}

func (cfg Config) toSidecar() sidecar.Config {
	return sidecar.Config{
		HealthAddr: cfg.Sidecar.HealthAddr,
		SeedDir:    cfg.Sidecar.SeedDir,
		DataDir:    cfg.DataDir,
		PollDelay:  2 * time.Second,
	}
}

func (cfg Config) toLDAP() ldapConfig {
	bindDN := cfg.Connection.BindDN
	if bindDN == "" {
		bindDN = "cn=admin," + cfg.Connection.BaseDN
	}

	return ldapConfig{
		uri:    cfg.Connection.URI,
		bindDN: bindDN,
		bindPW: os.Getenv("LDAP_ADMIN_PW"),
	}
}

// loadConfig reads a JSON config file at path. Fields absent from the file
// retain their built-in default values.
func loadConfig(path string) (Config, error) {
	cfg := configDefaults()
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func configDefaults() Config {
	return Config{
		LogLevel:   "info",
		DataDir:    "/var/lib/ldap",
		RunDir:     "/var/run/slapd",
		RootpwPath: "/etc/ldap/auth/rootpw.conf",
		Connection: ConnectionConfig{
			URI: "ldapi:///",
		},
		Sidecar: SidecarConfig{
			HealthAddr: ":8080",
			SeedDir:    "/seed",
		},
	}
}
