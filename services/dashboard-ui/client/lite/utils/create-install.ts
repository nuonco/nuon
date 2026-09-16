import type { TCreateAppInstallBody } from '@/lib'
import type { TAPIError, TAppConfig, TCloudPlatform } from '@/types'

export interface ICreateInstallInputValue {
  name: string
  value: string | boolean
}

export interface ICreateInstallValues {
  name: string
  location: string
  accountId: string
  autoApprove: boolean
  stackOnly: boolean
  labels: Array<{ key: string; value: string }>
  inputs: ICreateInstallInputValue[]
  branchId: string
  installGroupId: string
}

export const normalizeInstallPlatform = (value?: string): TCloudPlatform => {
  const normalized = value?.toLowerCase() ?? ''
  if (normalized.startsWith('aws')) return 'aws'
  if (normalized.startsWith('azure')) return 'azure'
  if (normalized.startsWith('gcp')) return 'gcp'
  return 'unknown'
}

export const isDuplicateInstallNameError = (
  error?: Error | TAPIError | null
) => {
  if (!error) return false
  const apiError = 'error' in error ? error : undefined
  const text = [
    apiError?.error,
    apiError?.description,
    'message' in error ? error.message : undefined,
  ]
    .filter(Boolean)
    .join(' ')
    .toLowerCase()

  return (
    text.includes('duplicate key') ||
    text.includes('duplicated key') ||
    text.includes('already exists')
  )
}

export const pickInstallConfig = (
  configs?: TAppConfig[],
  requireUnbranched = false
) =>
  configs?.find((config) => {
    if (config?.status && config.status !== 'active') return false
    if (config?.labels?.source === 'git-preview-run') return false
    return !requireUnbranched || !config?.app_branch_id
  })

export const buildCreateInstallBody = ({
  values,
  platform,
  groupLabels,
}: {
  values: ICreateInstallValues
  platform: TCloudPlatform
  groupLabels?: Record<string, string>
}): TCreateAppInstallBody => {
  const labels = Object.fromEntries(
    values.labels.flatMap(({ key, value }) => {
      const trimmed = key.trim()
      return trimmed ? [[trimmed, value.trim()]] : []
    })
  )
  const inputs = Object.fromEntries(
    values.inputs.map(({ name, value }) => [name, String(value)])
  )

  const body: TCreateAppInstallBody = {
    name: values.name.trim(),
    inputs: Object.keys(inputs).length ? inputs : undefined,
    install_config: {
      approval_option: values.autoApprove ? 'approve-all' : 'prompt',
    },
    labels:
      Object.keys(labels).length || Object.keys(groupLabels ?? {}).length
        ? { ...labels, ...groupLabels }
        : undefined,
    metadata: { managed_by: 'nuon/dashboard' },
    stack_only: values.stackOnly || undefined,
    app_branch_id: values.branchId || undefined,
  }

  if (platform === 'aws') {
    body.aws_account = {
      account_id: values.accountId.trim() || undefined,
      iam_role_arn: '',
      region: values.location,
    }
  } else if (platform === 'azure') {
    body.azure_account = {
      location: values.location,
      service_principal_app_id: '',
      service_principal_password: '',
      subscription_id: values.accountId.trim() || undefined,
      subscription_tenant_id: '',
    }
  } else if (platform === 'gcp') {
    body.gcp_account = {
      project_id: values.accountId.trim() || undefined,
      region: values.location,
    }
  }

  return body
}
