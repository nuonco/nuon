import { useOutletContext } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { Button } from '@/components/common/Button'
import { EmptyState } from '@/components/common/EmptyState'
import { Icon } from '@/components/common/Icon'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { RunbookRunTimeline } from '@/components/runbooks/RunbookRunTimeline'
import { RunbookStepCard } from '@/components/runbooks/RunbookStepCard'
import { Panel } from '@/components/surfaces/Panel'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { getWorkflowSteps } from '@/lib'
import type { TInstallRunbookOutletContext } from './types'

export const RunbookHistoryTab = () => {
  const { installRunbook } = useOutletContext<TInstallRunbookOutletContext>()
  const { org } = useOrg()
  const { install } = useInstall()
  const { addPanel } = useSurfaces()

  const runbook = installRunbook?.runbook
  const runs = installRunbook?.runs ?? []
  const lastRun = runs[0]
  const workflowId =
    lastRun?.install_workflow_id ?? lastRun?.install_workflow?.id
  const basePath = `/${org?.id}/installs/${install?.id}`

  const { data: workflowSteps, isLoading: isLoadingSteps } = useQuery({
    queryKey: ['runbook-workflow-steps', workflowId],
    queryFn: () =>
      getWorkflowSteps({ workflowId: workflowId!, orgId: org!.id }),
    enabled: !!workflowId && !!org?.id,
    refetchInterval: 10000,
  })

  if (!lastRun) {
    return (
      <EmptyState
        className="mt-12"
        variant="history"
        size="sm"
        emptyTitle="No runs yet"
        emptyMessage="This runbook has not been run yet. Trigger a run to see history here."
      />
    )
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

  const workflowUrl = `${basePath}/workflows/${workflowId}`

  const timeline = (
    <RunbookRunTimeline
      runs={runs}
      runbookName={runbook?.name ?? ''}
      basePath={basePath}
    />
  )

  return (
    <div className="@container">
      <div className="grid grid-cols-1 @3xl:grid-cols-12 gap-6">
        <div className="@3xl:col-span-8 flex flex-col gap-4">
          <div className="flex items-center justify-between">
            <Text variant="base" weight="strong">
              Latest run
            </Text>
            <div className="@3xl:hidden">
              <Button
                variant="secondary"
                size="sm"
                onClick={() =>
                  addPanel(
                    <Panel heading="Run history">
                      {timeline}
                    </Panel>
                  )
                }
              >
                <Icon variant="ClockCounterClockwiseIcon" size={16} />
                Run history
              </Button>
            </div>
          </div>

          {isLoadingSteps ? (
            <div className="flex flex-col gap-4">
              <Skeleton height="160px" width="100%" />
              <Skeleton height="160px" width="100%" />
            </div>
          ) : steps.length > 0 ? (
            steps.map((step) => (
              <RunbookStepCard
                key={step.id}
                step={step}
                installId={install!.id}
                orgId={org!.id}
                workflowUrl={workflowUrl}
              />
            ))
          ) : (
            <Text theme="neutral">No step data available for this run.</Text>
          )}
        </div>

        <div className="hidden @3xl:flex flex-col @3xl:col-span-4 gap-4">
          <Text variant="base" weight="strong">
            Run history
          </Text>
          {timeline}
        </div>
      </div>
    </div>
  )
}
