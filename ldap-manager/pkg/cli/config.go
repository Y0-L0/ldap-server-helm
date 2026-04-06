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

// Config is the top-level ldap-manager configuration, loaded from a JSON file.
// AdminPW is the only field not sourced from the file — it comes from LDAP_ADMIN_PW.
type Config struct {
	LogLevel   string           `json:"logLevel"`
	DataDir    string           `json:"dataDir"`
	RunDir     string           `json:"runDir"`
	RootpwPath string           `json:"rootpwPath"`
	Connection ConnectionConfig `json:"connection"`
	Sidecar    SidecarConfig    `json:"sidecar"`
	AdminPW    string           `json:"-"`
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
		HealthAddr: cfg.Sidecar.HealthAddr,
		SeedDir:    cfg.Sidecar.SeedDir,
		DataDir:    cfg.DataDir,
		PollDelay:  2 * time.Second,
	}
}

func (cfg Config) LDAPConfig() LDAPConfig {
	return LDAPConfig{
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
	required := []struct {
		name string
		val  string
	}{
		{"logLevel", cfg.LogLevel},
		{"dataDir", cfg.DataDir},
		{"runDir", cfg.RunDir},
		{"rootpwPath", cfg.RootpwPath},
		{"connection.uri", cfg.Connection.URI},
		{"connection.baseDN", cfg.Connection.BaseDN},
		{"connection.bindDN", cfg.Connection.BindDN},
		{"sidecar.healthAddr", cfg.Sidecar.HealthAddr},
		{"sidecar.seedDir", cfg.Sidecar.SeedDir},
		{"LDAP_ADMIN_PW", cfg.AdminPW},
	}

	var missing []string
	for _, f := range required {
		if f.val == "" {
			missing = append(missing, f.name)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required fields: %s", strings.Join(missing, ", "))
	}

	return nil
}
