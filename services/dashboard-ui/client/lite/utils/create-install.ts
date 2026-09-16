import type { TCreateAppInstallBody } from '@/lib'
import type {
  TAPIError,
  TAppConfig,
  TAppInput,
  TAppInputConfig,
  TCloudPlatform,
} from '@/types'
import { COMPONENT_OVERRIDE_INPUT_GROUP } from '@/utils/install-utils'
import type { IInstallSetupInput } from '../components/organisms/InstallSetup/InstallSetup'

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

const inputFromApi = (input: TAppInput): IInstallSetupInput | undefined => {
  if (!input?.name || input?.source === 'customer') return undefined
  const boolean =
    input?.type === 'bool' ||
    input?.default === 'true' ||
    input?.default === 'false'
  const type = boolean
    ? 'boolean'
    : input?.type === 'number'
      ? 'number'
      : input?.sensitive
        ? 'password'
        : 'text'

  return {
    name: input.name,
    label: input?.display_name ?? input.name,
    description: input?.description,
    required: input?.required,
    type,
    defaultValue: boolean ? input?.default === 'true' : (input?.default ?? ''),
  }
}

const byIndex = <T extends { index?: number }>(a: T, b: T) =>
  (a?.index ?? 0) - (b?.index ?? 0)

export const installSetupInputs = (
  inputConfig?: TAppInputConfig
): IInstallSetupInput[] => {
  const all = inputConfig?.inputs ?? []
  if (!all.length) return []

  return [...(inputConfig?.input_groups ?? [])]
    .filter((group) => group?.name !== COMPONENT_OVERRIDE_INPUT_GROUP)
    .sort(byIndex)
    .flatMap((group) =>
      all
        .filter((input) => input?.group_id === group?.id)
        .sort(byIndex)
        .flatMap((input) => {
          const resolved = inputFromApi(input)
          return resolved ? [resolved] : []
        })
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

export const resolveInstallConfig = (
  configs?: TAppConfig[],
  selectedBranchId?: string
): { config?: TAppConfig; branchId: string } => {
  if (selectedBranchId) {
    return { config: pickInstallConfig(configs), branchId: selectedBranchId }
  }

  const unbranched = pickInstallConfig(configs, true)
  if (unbranched) return { config: unbranched, branchId: '' }

  const latest = pickInstallConfig(configs)
  return { config: latest, branchId: latest?.app_branch_id ?? '' }
}

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
