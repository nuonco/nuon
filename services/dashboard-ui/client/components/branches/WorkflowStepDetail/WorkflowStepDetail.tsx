import type { TInstallWorkflowStep } from '@/types'
import { StepCard } from './StepCard'
import { CommitStep } from './steps/CommitStep'
import { ConfigStep } from './steps/ConfigStep'
import { BuildStep } from './steps/BuildStep'
import { ComparisonStep } from './steps/ComparisonStep'
import { PlanGroupStep } from './steps/PlanGroupStep'
import { DeployGroupStep } from './steps/DeployGroupStep'
import { PostDeployRunbooksStep } from './steps/PostDeployRunbooksStep'

interface IWorkflowStepDetail {
  step: TInstallWorkflowStep
  appBranchId?: string
  appBranchRunId?: string
  onClose: () => void
}

export const WorkflowStepDetail = ({
  step,
  appBranchId,
  appBranchRunId,
  onClose: _onClose,
}: IWorkflowStepDetail) => {
  const metadata = step.status?.metadata || {}
  const status = step.status?.status
  const name = step.name?.toLowerCase() ?? ''

  const isConfigStep =
    name.includes('config') &&
    !name.includes('diff') &&
    !name.includes('differences')

  const body = (() => {
    if (name.includes('commit')) return <CommitStep metadata={metadata} />
    if (isConfigStep) {
      return <ConfigStep metadata={metadata} status={status} />
    }
    if (
      name.includes('compute differences') ||
      name.includes('compute run comparison')
    ) {
      return (
        <ComparisonStep
          metadata={metadata}
          status={status}
          appBranchId={appBranchId}
          appBranchRunId={appBranchRunId}
        />
      )
    }
    if (name.includes('build')) {
      return (
        <BuildStep
          metadata={metadata}
          status={status}
          appBranchId={appBranchId}
          appBranchRunId={appBranchRunId}
        />
      )
    }
    if (
      name.includes('plan install group') ||
      name === 'plan preview install'
    ) {
      return <PlanGroupStep step={step} metadata={metadata} />
    }
    if (
      name.includes('deploy install group') ||
      name === 'apply preview install'
    ) {
      return <DeployGroupStep step={step} metadata={metadata} />
    }
    if (name.includes('post-deploy runbooks')) {
      return <PostDeployRunbooksStep step={step} metadata={metadata} />
    }
    return null
  })()

  return <StepCard step={step}>{body}</StepCard>
}
