package cli

import (
	"path/filepath"
)

func (s *Unittest) TestLoadConfig_FullConfig() {
	path := filepath.Join(s.T().TempDir(), "config.json")
	s.WriteFile(path, []byte(`{
		"logLevel": "debug",
		"dataDir": "/data",
		"runDir": "/run",
		"rootpwPath": "/etc/rootpw.conf",
		"connection": {
			"uri": "ldap://localhost:389/",
			"baseDN": "dc=example,dc=org",
			"bindDN": "cn=admin,dc=example,dc=org"
		},
		"sidecar": {
			"healthAddr": ":9090",
			"seedDir": "/custom/seed"
		}
	}`))

	cfg, err := loadConfig(path)
	s.Require().NoError(err)
	s.Equal("debug", cfg.LogLevel)
	s.Equal("/data", cfg.DataDir)
	s.Equal("/run", cfg.RunDir)
	s.Equal("/etc/rootpw.conf", cfg.RootpwPath)
	s.Equal("ldap://localhost:389/", cfg.Connection.URI)
	s.Equal("dc=example,dc=org", cfg.Connection.BaseDN)
	s.Equal("cn=admin,dc=example,dc=org", cfg.Connection.BindDN)
	s.Equal(":9090", cfg.Sidecar.HealthAddr)
	s.Equal("/custom/seed", cfg.Sidecar.SeedDir)
}

func (s *Unittest) TestLoadConfig_PartialConfig_KeepsDefaults() {
	path := filepath.Join(s.T().TempDir(), "config.json")
	s.WriteFile(path, []byte(`{"dataDir": "/custom/data"}`))

	cfg, err := loadConfig(path)
	s.Require().NoError(err)
	s.Equal("/custom/data", cfg.DataDir)
	// Fields absent from the file retain their built-in defaults.
	s.Equal("info", cfg.LogLevel)
	s.Equal("/var/run/slapd", cfg.RunDir)
	s.Equal("ldapi:///", cfg.Connection.URI)
	s.Equal(":8080", cfg.Sidecar.HealthAddr)
}

func (s *Unittest) TestLoadConfig_MissingFile() {
	_, err := loadConfig("/nonexistent/config.json")
	s.Require().Error(err)
}

func (s *Unittest) TestLoadConfig_MalformedJSON() {
	path := filepath.Join(s.T().TempDir(), "config.json")
	s.WriteFile(path, []byte(`not json`))

	_, err := loadConfig(path)
	s.Require().Error(err)
}
