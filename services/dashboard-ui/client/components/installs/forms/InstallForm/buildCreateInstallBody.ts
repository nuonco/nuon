import type { TCreateAppInstallBody } from '@/lib'
import type { InstallFormValues, InstallPlatform } from './schema'

export const buildCreateInstallBody = (
  values: InstallFormValues,
  platform?: InstallPlatform
): TCreateAppInstallBody => {
  const inputs: Record<string, string> = {}
  for (const [key, value] of Object.entries(values.inputs)) {
    inputs[key] = typeof value === 'boolean' ? String(value) : value
  }

  const installConfig: NonNullable<TCreateAppInstallBody['install_config']> = {
    approval_option: values.autoApprove ? 'approve-all' : 'prompt',
  }
  const vpc = values.vpcTemplateUrl?.trim()
  const runner = values.runnerTemplateUrl?.trim()
  if (vpc) installConfig.vpc_nested_template_url = vpc
  if (runner) installConfig.runner_nested_template_url = runner

  const labels: Record<string, string> = {}
  for (const { key, value } of values.labels) {
    const trimmed = key.trim()
    if (trimmed) labels[trimmed] = value.trim()
  }

  const body: TCreateAppInstallBody = {
    name: values.name.trim(),
    inputs: Object.keys(inputs).length > 0 ? inputs : undefined,
    install_config: installConfig,
    labels: Object.keys(labels).length > 0 ? labels : undefined,
    metadata: { managed_by: 'nuon/dashboard' },
  }

  if (platform === 'aws' && values.region) {
    body.aws_account = {
      iam_role_arn: '',
      region: values.region,
      connection_id: values.aws_connection_id || undefined,
      account_id: values.aws_account_id || undefined,
    }
  } else if (platform === 'azure' && values.location) {
    body.azure_account = {
      location: values.location,
      service_principal_app_id: '',
      service_principal_password: '',
      subscription_id: values.azure_subscription_id || undefined,
      subscription_tenant_id: '',
    }
  } else if (platform === 'gcp') {
    body.gcp_account = {
      project_id: values.gcp_project_id || undefined,
    }
  }

  return body
}
