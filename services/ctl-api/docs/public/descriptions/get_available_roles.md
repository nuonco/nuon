Returns the roles available in the install's stack outputs, including provision,
deprovision, maintenance, custom, and break-glass roles.

Pass `workflow_type` to preview the parent workflow's default for component and
action steps. Workflow-type defaults are enabled by default, using the same
mapping as execution. Set `USE_LEGACY_MAINTENANCE_ROLE_DEFAULT=true` to keep the
legacy maintenance default instead. The API and worker must use the same flag
value.

With `principal_type`, `operation_type`, and `principal_id`, the `default` marker
also accounts for entity, break-glass, and operation-matrix configuration. Omit
`principal_id` for adhoc actions. Sandbox defaults follow their operation type
independently of the workflow-default flag.

Without a principal, the preview is the workflow's base default; individual steps
may use configured role overrides. Omitting `workflow_type` preserves the legacy
component/action maintenance default.
