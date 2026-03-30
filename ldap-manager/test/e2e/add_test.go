package e2e

func (s *E2E) TestAddValidEntry() {
	err := s.backend.Add("cn=testuser,ou=people,"+baseDN, map[string][]string{
		"objectClass": {"inetOrgPerson"},
		"cn":          {"testuser"},
		"sn":          {"User"},
	})
	s.Require().NoError(err)
}

func (s *E2E) TestAddDuplicateEntry() {
	dn := "cn=dupuser,ou=people," + baseDN
	attrs := map[string][]string{
		"objectClass": {"inetOrgPerson"},
		"cn":          {"dupuser"},
		"sn":          {"Duplicate"},
	}

	err := s.backend.Add(dn, attrs)
	s.Require().NoError(err)

	err = s.backend.Add(dn, attrs)
	s.Require().Error(err, "adding duplicate DN should fail")
}

func (s *E2E) TestAddSchemaViolation() {
	// inetOrgPerson requires sn — omitting it should fail
	err := s.backend.Add("cn=baduser,ou=people,"+baseDN, map[string][]string{
		"objectClass": {"inetOrgPerson"},
		"cn":          {"baduser"},
	})
	s.Require().Error(err, "missing required attribute sn should fail")
}
