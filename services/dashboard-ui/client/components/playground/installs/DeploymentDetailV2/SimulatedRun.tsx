import { useMemo, useState } from 'react'
import { Button } from '@/components/common/Button'
import type {
  TInstallDeploymentRecord,
  TWorkflow,
  TWorkflowStep,
} from '@/types'
import { HelmDiff, TerraformDiff } from '@/components/diffs/plan-diff-switch'
import { singleImageTagChangePlan } from '@/lib/fixtures/plan-diffs/helm'
import { securityGroupUpdatePlan } from '@/lib/fixtures/plan-diffs/terraform'
import { Toggle } from '@/components/common/form/Toggle'
import { DeploymentDetailV2 } from './DeploymentDetailV2'
import { deploymentScenario } from './fixtures'
import { RunView } from './RunView'
import { useSimulatedRun, type ISimulatedRun } from './use-simulated-run'

const RunControls = ({
  autoApprove,
  onAutoApproveChange,
  onRestart,
}: {
  autoApprove: boolean
  onAutoApproveChange: (value: boolean) => void
  onRestart: () => void
}) => (
  <div className="flex w-full items-center justify-end gap-4 px-4 pt-4 md:px-6">
    <Toggle
      checked={autoApprove}
      label="Auto-approve plans"
      onChange={onAutoApproveChange}
    />
    <Button onClick={onRestart}>Restart story</Button>
  </div>
)

const renderPlan = (step: TWorkflowStep) =>
  step.approval?.type === 'terraform_plan' ? (
    <TerraformDiff plan={securityGroupUpdatePlan} />
  ) : (
    <HelmDiff plan={singleImageTagChangePlan} />
  )

const runViewOf = (run: ISimulatedRun, autoApprove: boolean) => (
  <RunView
    autoApprove={autoApprove}
    steps={run.steps}
    logs={run.logs}
    attempts={run.attempts}
    onApprove={run.approve}
    onDeny={run.deny}
    onRetry={run.retry}
    onSkip={run.skip}
    renderPlan={renderPlan}
  />
)

export const SimulatedRunView = () => {
  const [autoApprove, setAutoApprove] = useState(false)
  const run = useSimulatedRun({ autoApprove })

  return (
    <>
      <RunControls
        autoApprove={autoApprove}
        onAutoApproveChange={setAutoApprove}
        onRestart={run.restart}
      />
      <div className="w-full p-4 md:p-6">{runViewOf(run, autoApprove)}</div>
    </>
  )
}

export const SimulatedDeployment = () => {
  const [autoApprove, setAutoApprove] = useState(false)
  const run = useSimulatedRun({ autoApprove })
  const base = useMemo(() => deploymentScenario('active'), [])

  const workflow = {
    ...base.workflow,
    finished: run.isFinished,
    status: {
      ...base.workflow.status,
      status: run.status as TWorkflow['status']['status'],
    },
    steps: run.steps,
  }
  const deployment = {
    ...base.deployment,
    status: run.status as TInstallDeploymentRecord['status'],
  }

  return (
    <>
      <RunControls
        autoApprove={autoApprove}
        onAutoApproveChange={setAutoApprove}
        onRestart={run.restart}
      />
      <DeploymentDetailV2
        branch={base.branch}
        deployment={deployment}
        liveView={runViewOf(run, autoApprove)}
        workflow={workflow}
      />
    </>
  )
}
