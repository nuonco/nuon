package gcp

const tmpl = `nuon_install_id          = "{{.Install.ID}}"
nuon_org_id              = "{{.Runner.OrgID}}"
nuon_app_id              = "{{.Install.AppID}}"
runner_api_url           = "{{.Settings.RunnerAPIURL}}"
runner_api_token         = "{{.APIToken}}"
runner_id                = "{{.Runner.ID}}"
runner_init_script_url   = "{{.RunnerInitScriptURL}}"
phone_home_url           = "{{.CloudFormationStackVersion.PhoneHomeURL}}"
provision_permissions    = {{.ProvisionPermissions}}
maintenance_permissions  = {{.MaintenancePermissions}}
deprovision_permissions  = {{.DeprovisionPermissions}}
provision_predefined_role    = "{{.ProvisionPredefinedRole}}"
maintenance_predefined_role  = "{{.MaintenancePredefinedRole}}"
deprovision_predefined_role  = "{{.DeprovisionPredefinedRole}}"
has_break_glass          = {{.HasBreakGlass}}{{if .HasBreakGlass}}
break_glass_permissions  = {{.BreakGlassPermissions}}
break_glass_predefined_role  = "{{.BreakGlassPredefinedRole}}"{{end}}
`
