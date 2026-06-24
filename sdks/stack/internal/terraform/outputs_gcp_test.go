package terraform

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-exec/tfexec"
)

// TestGCPOutputsToCore locks the module output names to install-stacks/gcp
// outputs.tf and verifies typed decoding into core.Outputs, including null
// service-account outputs.
func TestGCPOutputsToCore(t *testing.T) {
	raw := map[string]string{
		"project_id":                       `"my-proj"`,
		"region":                           `"us-central1"`,
		"network_name":                     `"nuon-net"`,
		"runner_service_account_email":     `"runner@my-proj.iam.gserviceaccount.com"`,
		"runner_service_account_unique_id": `"123456789"`,
		"provision_sa_email":               `"prov@my-proj.iam.gserviceaccount.com"`,
		"maintenance_sa_email":             `null`,
		"custom_sa_emails":                 `{"inst-certs":"certs@my-proj.iam.gserviceaccount.com"}`,
		"secret_names":                     `{"db_secret_name":"projects/my-proj/secrets/db/versions/1"}`,
		"install_inputs":                   `{"github_app_id":"4121567"}`,
	}
	meta := map[string]tfexec.OutputMeta{}
	for k, v := range raw {
		meta[k] = tfexec.OutputMeta{Value: json.RawMessage(v)}
	}

	out, err := gcpOutputsToCore(meta)
	if err != nil {
		t.Fatalf("gcpOutputsToCore: %v", err)
	}

	if out.Cloud != "gcp" {
		t.Errorf("Cloud = %q, want gcp", out.Cloud)
	}
	if out.GCP == nil {
		t.Fatal("GCP outputs nil")
	}
	if out.GCP.ProjectID != "my-proj" {
		t.Errorf("ProjectID = %q", out.GCP.ProjectID)
	}
	if out.GCP.RunnerSAEmail != "runner@my-proj.iam.gserviceaccount.com" {
		t.Errorf("RunnerSAEmail = %q", out.GCP.RunnerSAEmail)
	}
	if out.GCP.MaintenanceSAEmail != "" {
		t.Errorf("MaintenanceSAEmail (null) = %q, want empty", out.GCP.MaintenanceSAEmail)
	}
	if out.GCP.CustomSAEmails["inst-certs"] == "" {
		t.Errorf("CustomSAEmails missing inst-certs: %v", out.GCP.CustomSAEmails)
	}
	if out.GCP.SecretNames["db_secret_name"] == "" {
		t.Errorf("SecretNames missing db_secret_name: %v", out.GCP.SecretNames)
	}
	if out.InstallInputs["github_app_id"] != "4121567" {
		t.Errorf("InstallInputs = %v", out.InstallInputs)
	}
}
