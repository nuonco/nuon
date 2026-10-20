import type { TGroupStepSelection } from '@/components/installs/CreateInstall/GroupStep'
import {
  getEditableInputs,
  isBooleanInput,
  type InstallFormValues,
  type InstallPlatform,
} from '@/components/installs/forms/InstallForm'
import { AWS_ACCOUNT_ID_REGEX } from '@/components/installs/forms/InstallForm/schema'
import type {
  TAppBranchConfig,
  TAppInputConfig,
  TAppBranchInstallGroup,
} from '@/types'

type TLabelValue = { key: string; value: string }

const labelRecord = (labels: TLabelValue[]) =>
  Object.fromEntries(
    labels
      .filter(({ key }) => key.trim())
      .map(({ key, value }) => [key.trim(), value.trim()])
  )

export const assignableGroupLabels = (
  group?: TAppBranchInstallGroup | null
): Record<string, string> =>
  Object.fromEntries(
    Object.entries(group?.label_selector?.match_labels ?? {}).filter(
      ([, value]) => value !== '*'
    )
  )

export const labelsMatchGroup = (
  labels: TLabelValue[],
  group?: TAppBranchInstallGroup | null
) => {
  const required = Object.entries(assignableGroupLabels(group))
  if (required.length === 0) return false
  const values = labelRecord(labels)
  return required.every(([key, value]) => values[key] === value)
}

export const assignmentBlocker = (
  selection: TGroupStepSelection,
  labels: TLabelValue[],
  config?: TAppBranchConfig
) => {
  if (selection.mode === 'labels' && !labelsMatchGroup(labels, selection.group)) {
    return 'Restore the selected group labels or choose pin to group'
  }

  if (
    (selection.mode === 'explicit' || selection.mode === 'default') &&
    (config?.install_groups ?? []).some(
      (group) => !group.default && labelsMatchGroup(labels, group)
    )
  ) {
    return 'Remove the group selector labels or use label-based assignment'
  }

  return undefined
}

export const mergeCarriedLabels = ({
  labels,
  previous,
  next,
}: {
  labels: TLabelValue[]
  previous: Record<string, string>
  next: Record<string, string>
}) => {
  const withoutPrevious = labels.filter(
    ({ key, value }) => previous[key.trim()] !== value.trim()
  )
  const existing = new Set(withoutPrevious.map(({ key }) => key.trim()))
  return [
    ...withoutPrevious,
    ...Object.entries(next)
      .filter(([key]) => !existing.has(key))
      .map(([key, value]) => ({ key, value })),
  ]
}

export const enrollmentBlocker = (
  groupCount: number,
  selection: TGroupStepSelection
) =>
  groupCount > 0 && selection.mode === 'none'
    ? 'Choose an install group'
    : undefined

export const installDetailsBlocker = ({
  values,
  platform,
  requireTargetAccount,
}: {
  values: InstallFormValues
  platform?: InstallPlatform
  requireTargetAccount: boolean
}) => {
  if (!values.name.trim()) return 'Enter the install name and destination'

  if (platform === 'aws') {
    if (!values.region) return 'Enter the install name and region'
    const accountId = values.aws_account_id
    if (accountId && !AWS_ACCOUNT_ID_REGEX.test(accountId)) {
      return 'Enter a 12-digit AWS account ID'
    }
    if (requireTargetAccount && !AWS_ACCOUNT_ID_REGEX.test(accountId)) {
      return 'Enter a 12-digit AWS account ID'
    }
  }

  if (platform === 'azure') {
    if (!values.location) return 'Enter the install name and location'
    if (requireTargetAccount && !values.azure_subscription_id.trim()) {
      return 'Enter the Azure subscription ID'
    }
  }

  if (platform === 'gcp' && requireTargetAccount && !values.gcp_project_id.trim()) {
    return 'Enter the GCP project ID'
  }

  return undefined
}

export const installInputsBlocker = (
  values: InstallFormValues,
  inputConfig?: TAppInputConfig
) => {
  const missing = getEditableInputs(inputConfig).some((input) => {
    if (!input?.required || !input.name || isBooleanInput(input)) return false
    const value = values.inputs[input.name]
    return typeof value !== 'string' || !value.trim()
  })
  return missing ? 'Fill in the required inputs' : undefined
}
