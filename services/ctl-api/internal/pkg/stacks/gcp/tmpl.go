package gcp

const inputsTmpl = `nuon_install_id          = "{{.Install.ID}}"
nuon_org_id              = "{{.Runner.OrgID}}"
nuon_app_id              = "{{.Install.AppID}}"
{{- if .GCPProjectID}}
gcp_project_id           = "{{.GCPProjectID}}"
{{- end}}
{{- if .GCPRegion}}
gcp_region               = "{{.GCPRegion}}"
{{- end}}
runner_api_url           = "{{.Settings.RunnerAPIURL}}"
{{- if .APIToken}}
runner_api_token         = "{{.APIToken}}"
{{- end}}
runner_id                = "{{.Runner.ID}}"
runner_init_script_url   = "{{.RunnerInitScriptURL}}"
{{- if .Settings.AWSInstanceType}}
runner_machine_type      = "{{.Settings.AWSInstanceType}}"
{{- end}}
phone_home_url           = "{{.CloudFormationStackVersion.PhoneHomeURL}}"
provision_policies = {
{{- range .ProvisionPolicies}}
  "{{.Name}}" = {{.Permissions}}
{{- end}}
}
maintenance_policies = {
{{- range .MaintenancePolicies}}
  "{{.Name}}" = {{.Permissions}}
{{- end}}
}
deprovision_policies = {
{{- range .DeprovisionPolicies}}
  "{{.Name}}" = {{.Permissions}}
{{- end}}
}
provision_predefined_role    = "{{.ProvisionPredefinedRole}}"
maintenance_predefined_role  = "{{.MaintenancePredefinedRole}}"
deprovision_predefined_role  = "{{.DeprovisionPredefinedRole}}"
break_glass_roles = {
{{- range .BreakGlassRoles}}
  "{{.Name}}" = {
    policies = {
    {{- range .Policies}}
      "{{.Name}}" = {{.Permissions}}
    {{- end}}
    }
    predefined_role = "{{.PredefinedRole}}"
    enabled         = false
  }
{{- end}}
}
custom_roles = {
{{- range .CustomRoles}}
  "{{.Name}}" = {
    policies = {
    {{- range .Policies}}
      "{{.Name}}" = {{.Permissions}}
    {{- end}}
    }
    predefined_role = "{{.PredefinedRole}}"
    enabled         = true
  }
{{- end}}
}
install_inputs = {
{{- range .InstallInputs}}
  "{{.Name}}" = "{{.Value}}"
{{- end}}
}
`

// providerInputsTmpl is the slimmed-down tfvars for the Terraform-provider
// flow: the install-stacks module reads runner details, permissions and roles
// from the API via the stack_config data source, so only the API base URL, the
// phone-home ID, the customer GCP project/region, and the install-input names
// need to be supplied here.
const providerInputsTmpl = `nuon_api_url  = "{{.Settings.RunnerAPIURL}}"
phone_home_id = "{{.CloudFormationStackVersion.PhoneHomeID}}"
{{- if .GCPProjectID}}
gcp_project_id = "{{.GCPProjectID}}"
{{- end}}
{{- if .GCPRegion}}
gcp_region     = "{{.GCPRegion}}"
{{- end}}
install_inputs = {
{{- range .InstallInputs}}
  "{{.Name}}" = "{{.Value}}"
{{- end}}
}
`

const secretsTmpl = `auto_generate_secrets = [{{range .AutoGenerateSecrets}}"{{.}}", {{end}}]
secrets = {
{{- range .Secrets}}
  "{{.Name}}" = {
    description = "{{.Description}}"
    required    = {{.Required}}
    value       = "{{.Value}}"
  }
{{- end}}
}
`
