package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/y0-l0/ldap-server-helm/ldap-manager/pkg/setup"
	"github.com/y0-l0/ldap-server-helm/ldap-manager/pkg/sidecar"
)

const healthAddr = ":8080"

// Config is the top-level ldap-manager configuration, loaded from a JSON file.
// AdminPW is the only field not sourced from the file — it comes from LDAP_ADMIN_PW.
type Config struct {
	LogLevel    string           `json:"logLevel"`
	DataDir     string           `json:"dataDir"`
	RunDir      string           `json:"runDir"`
	RootpwPath  string           `json:"rootpwPath"`
	LdifSeedDir string           `json:"ldifSeedDir"`
	Connection  ConnectionConfig `json:"connection"`
	AdminPW     string           `json:"-"`
}

// ConnectionConfig holds LDAP connection settings used by the sidecar.
type ConnectionConfig struct {
	URI    string `json:"uri"`
	BindDN string `json:"bindDN"`
}

func (cfg Config) SetupConfig() setup.Config {
	return setup.Config{
		DataDir:    cfg.DataDir,
		RunDir:     cfg.RunDir,
		RootpwPath: cfg.RootpwPath,
		AdminPW:    cfg.AdminPW,
	}
}

func (cfg Config) SidecarConfig() sidecar.Config {
	return sidecar.Config{
		HealthAddr: healthAddr,
		SeedDir:    cfg.LdifSeedDir,
		DataDir:    cfg.DataDir,
		PollDelay:  2 * time.Second,
	}
}

func (cfg Config) ldapCfg() ldapConfig {
	return ldapConfig{
		uri:    cfg.Connection.URI,
		bindDN: cfg.Connection.BindDN,
		bindPW: cfg.AdminPW,
	}
}

// loadConfig reads a JSON config file at path and the LDAP_ADMIN_PW env var.
// All fields are required — missing fields are reported as an error.
func loadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}

	cfg.AdminPW = os.Getenv("LDAP_ADMIN_PW")

	if err := cfg.validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (cfg Config) validate() error {
	var missing []string
	if cfg.LogLevel == "" {
		missing = append(missing, "logLevel")
	}
	if cfg.DataDir == "" {
		missing = append(missing, "dataDir")
	}
	if cfg.RunDir == "" {
		missing = append(missing, "runDir")
	}
	if cfg.RootpwPath == "" {
		missing = append(missing, "rootpwPath")
	}
	if cfg.LdifSeedDir == "" {
		missing = append(missing, "ldifSeedDir")
	}
	if cfg.AdminPW == "" {
		missing = append(missing, "LDAP_ADMIN_PW")
	}
	missing = append(missing, cfg.Connection.missing()...)
	if len(missing) > 0 {
		return fmt.Errorf("missing required fields: %s", strings.Join(missing, ", "))
	}
	return nil
}

func (c ConnectionConfig) missing() []string {
	var missing []string
	if c.URI == "" {
		missing = append(missing, "connection.uri")
	}
	if c.BindDN == "" {
		missing = append(missing, "connection.bindDN")
	}
	return missing
}
