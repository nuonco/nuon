package aws

const tmpl = `nuon_install_id          = "{{.Install.ID}}"
nuon_org_id              = "{{.Runner.OrgID}}"
nuon_app_id              = "{{.Install.AppID}}"
{{- if .Install.AWSAccount}}
{{- if .Install.AWSAccount.Region}}
aws_region               = "{{.Install.AWSAccount.Region}}"
{{- end}}
{{- end}}
runner_api_url           = "{{.Settings.RunnerAPIURL}}"
{{- if .APIToken}}
runner_api_token         = "{{.APIToken}}"
{{- end}}
runner_id                = "{{.Runner.ID}}"
runner_init_script_url   = "{{.RunnerInitScriptURL}}"
phone_home_url           = "{{.CloudFormationStackVersion.PhoneHomeURL}}"
nuon_control_plane_account_ids = {{.ControlPlaneAccountIDs}}
provision_permissions          = {{.ProvisionPermissions}}
maintenance_permissions        = {{.MaintenancePermissions}}
deprovision_permissions        = {{.DeprovisionPermissions}}
provision_managed_policy_arns   = {{.ProvisionManagedPolicyArns}}
maintenance_managed_policy_arns = {{.MaintenanceManagedPolicyArns}}
deprovision_managed_policy_arns = {{.DeprovisionManagedPolicyArns}}
break_glass_roles = {
{{- range .BreakGlassRoles}}
  "{{.Name}}" = {
    permissions         = {{.Permissions}}
    managed_policy_arns = {{.ManagedPolicyArns}}
    enabled             = false
  }
{{- end}}
}
custom_roles = {
{{- range .CustomRoles}}
  "{{.Name}}" = {
    permissions         = {{.Permissions}}
    managed_policy_arns = {{.ManagedPolicyArns}}
    enabled             = true
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
