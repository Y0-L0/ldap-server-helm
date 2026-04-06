package cli

import (
	"path/filepath"
)

func (s *Unittest) TestLoadConfig_FullConfig() {
	jsonConfig := []byte(`{
		"logLevel": "debug",
		"dataDir": "/data",
		"runDir": "/run",
		"rootpwPath": "/etc/rootpw.conf",
		"ldifSeedDir": "/custom/seed",
		"connection": {
			"uri": "ldap://localhost:389/",
			"bindDN": "cn=admin,dc=example,dc=org"
		}
	}`)
	expected := Config{
		LogLevel:    "debug",
		DataDir:     "/data",
		RunDir:      "/run",
		RootpwPath:  "/etc/rootpw.conf",
		LdifSeedDir: "/custom/seed",
		Connection: ConnectionConfig{
			URI:    "ldap://localhost:389/",
			BindDN: "cn=admin,dc=example,dc=org",
		},
		AdminPW: "secret",
	}

	s.T().Setenv("LDAP_ADMIN_PW", "secret")

	path := filepath.Join(s.T().TempDir(), "config.json")
	s.WriteFile(path, jsonConfig)

	actual, err := loadConfig(path)

	s.Require().NoError(err)
	s.Equal(expected, actual)
}

func (s *Unittest) TestLoadConfig_AdminPW_NotInJSON() {
	s.T().Setenv("LDAP_ADMIN_PW", "secret")
	path := filepath.Join(s.T().TempDir(), "config.json")
	s.WriteFile(path, []byte(`{"adminPW": "should-be-ignored"}`))

	actual, err := loadConfig(path)

	// Validation will fail on missing fields, but AdminPW must come from env.
	s.Require().Error(err)
	s.NotEqual("should-be-ignored", actual.AdminPW)
}

func (s *Unittest) TestLoadConfig_MissingFields() {
	s.T().Setenv("LDAP_ADMIN_PW", "")
	path := filepath.Join(s.T().TempDir(), "config.json")
	s.WriteFile(path, []byte(`{}`))

	_, err := loadConfig(path)
	s.Require().Error(err)
	s.Contains(err.Error(), "missing required fields")
	s.Contains(err.Error(), "LDAP_ADMIN_PW")
	s.Contains(err.Error(), "logLevel")
	s.Contains(err.Error(), "connection.bindDN")
}

func (s *Unittest) TestLoadConfig_MissingFile() {
	_, err := loadConfig("/nonexistent/config.json")
	s.Require().Error(err)
}

func (s *Unittest) TestLoadConfig_MalformedJSON() {
	s.T().Setenv("LDAP_ADMIN_PW", "secret")
	path := filepath.Join(s.T().TempDir(), "config.json")
	s.WriteFile(path, []byte(`not json`))

	_, err := loadConfig(path)
	s.Require().Error(err)
}
