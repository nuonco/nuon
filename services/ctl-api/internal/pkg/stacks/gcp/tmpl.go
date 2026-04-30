package gcp

const tmpl = `nuon_install_id          = "{{.Install.ID}}"
nuon_org_id              = "{{.Runner.OrgID}}"
nuon_app_id              = "{{.Install.AppID}}"
{{- if .Install.GCPAccount}}
{{- if .Install.GCPAccount.ProjectID}}
gcp_project_id           = "{{.Install.GCPAccount.ProjectID}}"
{{- end}}
{{- if .Install.GCPAccount.Region}}
gcp_region               = "{{.Install.GCPAccount.Region}}"
{{- end}}
{{- end}}
runner_api_url           = "{{.Settings.RunnerAPIURL}}"
{{- if .APIToken}}
runner_api_token         = "{{.APIToken}}"
{{- end}}
runner_id                = "{{.Runner.ID}}"
runner_init_script_url   = "{{.RunnerInitScriptURL}}"
phone_home_url           = "{{.CloudFormationStackVersion.PhoneHomeURL}}"
provision_permissions    = {{.ProvisionPermissions}}
maintenance_permissions  = {{.MaintenancePermissions}}
deprovision_permissions  = {{.DeprovisionPermissions}}
provision_predefined_role    = "{{.ProvisionPredefinedRole}}"
maintenance_predefined_role  = "{{.MaintenancePredefinedRole}}"
deprovision_predefined_role  = "{{.DeprovisionPredefinedRole}}"
break_glass_roles = {
{{- range .BreakGlassRoles}}
  "{{.Name}}" = {
    permissions     = {{.Permissions}}
    predefined_role = "{{.PredefinedRole}}"
    enabled         = false
  }
{{- end}}
}
custom_roles = {
{{- range .CustomRoles}}
  "{{.Name}}" = {
    permissions     = {{.Permissions}}
    predefined_role = "{{.PredefinedRole}}"
    enabled         = true
  }
{{- end}}
}
install_inputs = {
{{- range .InstallInputs}}
  "{{.}}" = ""
{{- end}}
}
auto_generate_secrets = [{{range .AutoGenerateSecrets}}"{{.}}", {{end}}]
secrets = {
{{- range .Secrets}}
  "{{.Name}}" = {
    description = "{{.Description}}"
    required    = {{.Required}}
    value       = "{{.Default}}"
  }
{{- end}}
}
`
