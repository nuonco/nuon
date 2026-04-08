export default {
  title: 'Approvals/Plan',
}

import { mockWorkflowStep, mockHelmStep, mockK8sStep } from '@/components/__fixtures__/workflows'
import { Plan } from './Plan'

export const LoadingTerraform = () => (
  <Plan step={mockWorkflowStep} plan={null} isLoading={true} error={null} />
)

export const LoadingHelm = () => (
  <Plan step={mockHelmStep} plan={null} isLoading={true} error={null} />
)

export const LoadingKubernetes = () => (
  <Plan step={mockK8sStep} plan={null} isLoading={true} error={null} />
)

export const NoPlanGenerated = () => (
  <Plan step={mockWorkflowStep} plan={null} isLoading={false} error={null} />
)

export const FailedToLoad = () => (
  <Plan
    step={mockWorkflowStep}
    plan={null}
    isLoading={false}
    error={{ message: 'Network error' }}
  />
)

export const NoApprovalOnStep = () => (
  <Plan
    step={{
      ...mockWorkflowStep,
      approval: undefined,
      finished: true,
    } as any}
    plan={null}
    isLoading={false}
    error={null}
  />
)

export const NoApprovalStillRunning = () => (
  <Plan
    step={{
      ...mockWorkflowStep,
      approval: undefined,
      finished: false,
    } as any}
    plan={null}
    isLoading={false}
    error={null}
  />
)
