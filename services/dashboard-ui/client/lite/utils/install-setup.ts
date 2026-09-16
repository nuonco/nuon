import type { TInstall } from '@/types'
import { type IWizardDescriptor, type IWizardStep } from './wizard'

export interface IInstallSetupState {
  lifecyclePhase?: string
}

export const installSetupStateFromInstall = (
  install?: Pick<TInstall, 'lifecycle_phase'> | null
): IInstallSetupState => ({
  lifecyclePhase: install?.lifecycle_phase?.phase,
})

const provisionStep: IWizardStep<IInstallSetupState> = {
  id: 'provision',
  label: 'Provision install',
  complete: (state) => state.lifecyclePhase !== 'provisioning',
  render: () => null,
}

export const installSetupDescriptor: IWizardDescriptor<IInstallSetupState> = {
  steps: [provisionStep],
}
