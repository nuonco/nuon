package azureroles

import "testing"

func TestGUID(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"Azure Kubernetes Service RBAC Reader", "7f6c6a51-bcf8-42ba-9220-52d62157d7db"},
		{"Azure Kubernetes Service RBAC Cluster Admin", "b1ff04bb-8a4e-4dc4-8eb5-8693973ce19b"},
		{"Network Contributor", "4d97b98b-1d4f-4787-a291-c67834d212e7"},
		// Already a GUID: passed through so a config can name a role the map
		// does not carry.
		{"7f6c6a51-bcf8-42ba-9220-52d62157d7db", "7f6c6a51-bcf8-42ba-9220-52d62157d7db"},
	}
	for _, c := range cases {
		if got := GUID(c.in); got != c.want {
			t.Errorf("GUID(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestResolvable(t *testing.T) {
	ok := []string{
		"Azure Kubernetes Service RBAC Reader",
		"Azure Kubernetes Service RBAC Writer",
		"Azure Kubernetes Service RBAC Admin",
		"Owner",
		"7f6c6a51-bcf8-42ba-9220-52d62157d7db",
		"7F6C6A51-BCF8-42BA-9220-52D62157D7DB",
	}
	for _, v := range ok {
		if !Resolvable(v) {
			t.Errorf("Resolvable(%q) = false, want true", v)
		}
	}

	bad := []string{
		"",
		"Azure Kubernetes Service RBAC Readr", // typo
		"Azure Kubernetes Service RBAC SuperAdmin", // not a real role
		"7f6c6a51-bcf8-42ba-9220",                  // truncated GUID
		"7f6c6a51bcf842ba922052d62157d7db",         // GUID without dashes
	}
	for _, v := range bad {
		if Resolvable(v) {
			t.Errorf("Resolvable(%q) = true, want false", v)
		}
	}
}

// The AKS data-plane roles are the ones an Azure app config reaches for, and
// their absence is what let an invalid name reach a customer's stack deploy.
func TestAKSDataPlaneRolesAreMapped(t *testing.T) {
	for _, name := range []string{
		"Azure Kubernetes Service RBAC Reader",
		"Azure Kubernetes Service RBAC Writer",
		"Azure Kubernetes Service RBAC Admin",
		"Azure Kubernetes Service RBAC Cluster Admin",
	} {
		if GUID(name) == name {
			t.Errorf("%q did not resolve to a GUID", name)
		}
	}
}
