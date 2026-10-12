package nuon

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func TestListServiceAccountsPreservesManagementMetadata(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/service-accounts" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		for key, want := range map[string]string{
			"include_runners": "true",
			"include_stacks":  "false",
			"offset":          "20",
			"limit":           "5",
		} {
			if got := r.URL.Query().Get(key); got != want {
				t.Errorf("%s = %q, want %q", key, got, want)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Nuon-Page-Next", "true")
		fmt.Fprint(w, `[
			{"id":"acct-managed","name":"acme collector","system_account":true,"purposes":["install_telemetry_collector"],"managed_service_account":{"owner_type":"install","owner_id":"inl_acme","owner_name":"acme","purpose":"install_telemetry_collector","instance_key":"collector"}},
			{"id":"acct-user","name":"acme automation","system_account":false,"purposes":[]}
		]`)
	}))
	defer server.Close()

	c, err := New(WithURL(server.URL), WithOrgID("org-acme"))
	if err != nil {
		t.Fatal(err)
	}
	var api Client = c
	accounts, more, err := api.ListServiceAccounts(context.Background(), true, false, &models.GetPaginatedQuery{Offset: 20, Limit: 5})
	if err != nil {
		t.Fatal(err)
	}
	if !more || len(accounts) != 2 {
		t.Fatalf("accounts = %#v, more = %v", accounts, more)
	}
	managed := accounts[0]
	if managed.ID != "acct-managed" || managed.Name != "acme collector" || !managed.SystemAccount || len(managed.Purposes) != 1 || managed.Purposes[0] != "install_telemetry_collector" {
		t.Fatalf("managed account = %#v", managed)
	}
	owner := managed.ManagedServiceAccount
	if owner == nil || owner.OwnerType != "install" || owner.OwnerID != "inl_acme" || owner.OwnerName != "acme" || owner.Purpose != "install_telemetry_collector" || owner.InstanceKey != "collector" {
		t.Fatalf("managed account ownership = %#v", owner)
	}
	user := accounts[1]
	if user.ID != "acct-user" || user.SystemAccount || len(user.Purposes) != 0 || user.ManagedServiceAccount != nil {
		t.Fatalf("user account = %#v", user)
	}
}
