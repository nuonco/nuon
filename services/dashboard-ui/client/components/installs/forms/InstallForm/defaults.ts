import type { TAppInputConfig, TInstall } from '@/types'
import {
  getEditableInputs,
  isBooleanInput,
  type InstallFormMode,
  type InstallFormValues,
} from './schema'

export interface BuildInstallDefaultsParams {
  mode: InstallFormMode
  inputConfig?: TAppInputConfig
  install?: TInstall
  defaultAutoApprove?: boolean
}

export const buildInstallInputDefaults = (
  inputConfig?: TAppInputConfig,
  install?: TInstall
): Record<string, string | boolean> => {
  const installInputs = install?.install_inputs?.at(0)?.values ?? {}
  const inputs: Record<string, string | boolean> = {}

  for (const input of getEditableInputs(inputConfig)) {
    if (!input?.name) continue
    const existing = installInputs?.[input.name]
    if (isBooleanInput(input)) {
      inputs[input.name] =
        existing != null ? existing === 'true' : input.default === 'true'
    } else {
      inputs[input.name] = existing ?? input.default ?? ''
    }
  }

  return inputs
}

export const buildInstallDefaults = ({
  inputConfig,
  install,
  defaultAutoApprove,
}: BuildInstallDefaultsParams): InstallFormValues => ({
  name: install?.name ?? '',
  region: '',
  aws_connection_id: '',
  aws_account_id: '',
  location: '',
  azure_subscription_id: '',
  gcp_project_id: '',
  autoApprove: Boolean(defaultAutoApprove),
  vpcTemplateUrl: '',
  runnerTemplateUrl: '',
  labels: [],
  role: '',
  deployDependents: true,
  stackOnly: false,
  inputsOnly: false,
  inputs: buildInstallInputDefaults(inputConfig, install),
})

export const mergeDraftValues = (
  base: InstallFormValues,
  draft: Partial<InstallFormValues> | null | undefined
): InstallFormValues => {
  if (!draft) return base
  return {
    ...base,
    ...draft,
    inputs: { ...base.inputs, ...draft.inputs },
  }
}
