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
runner_id                = "{{.Runner.ID}}"
phone_home_url           = "{{.CloudFormationStackVersion.PhoneHomeURL}}"
nuon_support_iam_role_arns = {{.ControlPlaneAccountIDs}}
provision_permissions              = {{.ProvisionPermissions}}
maintenance_permissions            = {{.MaintenancePermissions}}
deprovision_permissions            = {{.DeprovisionPermissions}}
provision_inline_policy_document   = {{if .ProvisionInlinePolicyDocument}}{{.ProvisionInlinePolicyDocument}}{{else}}""{{end}}
maintenance_inline_policy_document = {{if .MaintenanceInlinePolicyDocument}}{{.MaintenanceInlinePolicyDocument}}{{else}}""{{end}}
deprovision_inline_policy_document = {{if .DeprovisionInlinePolicyDocument}}{{.DeprovisionInlinePolicyDocument}}{{else}}""{{end}}
provision_managed_policy_arns      = {{.ProvisionManagedPolicyArns}}
maintenance_managed_policy_arns    = {{.MaintenanceManagedPolicyArns}}
deprovision_managed_policy_arns    = {{.DeprovisionManagedPolicyArns}}
break_glass_roles = {
{{- range .BreakGlassRoles}}
  "{{.Name}}" = {
    permissions            = {{.Permissions}}
    inline_policy_document = {{if .InlinePolicyDocument}}{{.InlinePolicyDocument}}{{else}}""{{end}}
    managed_policy_arns    = {{.ManagedPolicyArns}}
    enabled                = false
  }
{{- end}}
}
custom_roles = {
{{- range .CustomRoles}}
  "{{.Name}}" = {
    permissions            = {{.Permissions}}
    inline_policy_document = {{if .InlinePolicyDocument}}{{.InlinePolicyDocument}}{{else}}""{{end}}
    managed_policy_arns    = {{.ManagedPolicyArns}}
    enabled                = true
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
