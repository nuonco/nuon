import type { TAppBranch } from '@/types'
import { latestBranchConfig } from '@/utils/branch-utils'
import { hasLabelSelector } from './label-selector'
import { type IWizardDescriptor, type IWizardStep } from './wizard'

export interface IAppSetupState {
  hasDeploymentPlan: boolean
}

const hasQualifyingGroup = (branch: TAppBranch) =>
  (latestBranchConfig(branch)?.install_groups ?? []).some(
    (group) =>
      Boolean(group.all_installs) || hasLabelSelector(group.label_selector)
  )

export const appSetupStateFromBranches = (
  branches?: Array<TAppBranch | undefined> | null
): IAppSetupState => ({
  hasDeploymentPlan: (branches ?? []).some(
    (branch) => branch && hasQualifyingGroup(branch)
  ),
})

const deploymentPlanStep: IWizardStep<IAppSetupState> = {
  id: 'deployment-plan',
  label: 'Create a deployment plan',
  complete: (state) => state.hasDeploymentPlan,
  render: () => null,
}

export const appSetupDescriptor: IWizardDescriptor<IAppSetupState> = {
  steps: [deploymentPlanStep],
}
