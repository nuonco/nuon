import { useOutletContext } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { RunbookStepCard } from '@/components/runbooks/RunbookStepCard'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getWorkflowSteps } from '@/lib'
import type { TInstallRunbookOutletContext } from './types'

export const RunbookLatestRunTab = () => {
  const { installRunbook } = useOutletContext<TInstallRunbookOutletContext>()
  const { org } = useOrg()
  const { install } = useInstall()

  const runs = installRunbook?.runs ?? []
  const lastRun = runs[0]
  const workflowId =
    lastRun?.install_workflow_id ?? lastRun?.install_workflow?.id

  const { data: workflowSteps, isLoading } = useQuery({
    queryKey: ['runbook-workflow-steps', workflowId],
    queryFn: () =>
      getWorkflowSteps({ workflowId: workflowId!, orgId: org!.id }),
    enabled: !!workflowId && !!org?.id,
  })

  if (!lastRun) {
    return <Text theme="neutral">No runs yet.</Text>
  }

  const steps = (workflowSteps ?? [])
    .filter(
      (s) =>
        s.step_target_type === 'install_deploys' ||
        s.step_target_type === 'install_action_workflow_runs'
    )
    .sort((a, b) => {
      const aTime = a.created_at ?? ''
      const bTime = b.created_at ?? ''
      return aTime.localeCompare(bTime)
    })

  if (isLoading) {
    return (
      <div className="flex flex-col gap-4">
        <Skeleton height="160px" width="100%" />
        <Skeleton height="160px" width="100%" />
      </div>
    )
  }

  const workflowUrl = `/${org?.id}/installs/${install?.id}/workflows/${workflowId}`

  if (!steps.length) {
    return <Text theme="neutral">No step data available for this run.</Text>
  }

  return (
    <div className="flex flex-col gap-4">
      {steps.map((step) => (
        <RunbookStepCard
          key={step.id}
          step={step}
          installId={install!.id}
          orgId={org!.id}
          workflowUrl={workflowUrl}
        />
      ))}
    </div>
  )
}
