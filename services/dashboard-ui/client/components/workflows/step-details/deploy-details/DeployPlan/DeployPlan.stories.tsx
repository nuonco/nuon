export default {
  title: 'Workflows/DeployPlan',
}

import { DeployPlan } from './DeployPlan'
import type { TWorkflowStep } from '@/types'

const terraformStep = {
  id: 'step-1',
  approval: { type: 'terraform_plan', id: 'apr-1' },
} as TWorkflowStep

const helmStep = {
  id: 'step-2',
  approval: { type: 'helm_approval', id: 'apr-2' },
} as TWorkflowStep

export const TerraformLoading = () => (
  <DeployPlan step={terraformStep} plan={null} isLoading={true} />
)

export const HelmLoading = () => (
  <DeployPlan step={helmStep} plan={null} isLoading={true} />
)

export const TerraformWithPlan = () => (
  <DeployPlan
    step={terraformStep}
    plan={{ resource_changes: [] }}
    isLoading={false}
  />
)
