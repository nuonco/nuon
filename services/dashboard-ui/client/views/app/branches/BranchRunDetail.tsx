import { useParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { Badge } from '@/components/common/Badge'
import { Card } from '@/components/common/Card'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { ID } from '@/components/common/ID'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { WorkflowStepsPipeline } from '@/components/branches/WorkflowStepsPipeline'
import { WorkflowStepDetail } from '@/components/branches/WorkflowStepDetail'
import { useOrg } from '@/hooks/use-org'
import { useApp } from '@/hooks/use-app'
import { useBranch } from '@/hooks/use-branch'
import { BranchProvider } from '@/providers/branch-provider'
import { getBranchWorkflowRun } from '@/lib'
import { useEffect, useState } from 'react'
import type { TInstallWorkflowStep } from '@/types'

function statusTheme(status?: string) {
  if (status === 'success') return 'success'
  if (status === 'error') return 'error'
  if (status === 'in-progress') return 'info'
  return 'neutral'
}

const BranchRunDetailContent = () => {
  const { org } = useOrg()
  const { app } = useApp()
  const { branch } = useBranch()
  const params = useParams()
  const orgId = params.orgId as string
  const appId = params.appId as string
  const branchId = params.branchId as string
  const runId = params.runId as string
  const [selectedStep, setSelectedStep] = useState<TInstallWorkflowStep | null>(null)

  const { data: run, isLoading } = useQuery({
    queryKey: ['branch-run', orgId, appId, branchId, runId],
    queryFn: () => getBranchWorkflowRun({ orgId, appId, branchId, runId }),
    enabled: !!orgId && !!appId && !!branchId && !!runId,
    refetchInterval: 5000,
  })

  const steps = run?.steps || []

  useEffect(() => {
    if (steps.length > 0 && !selectedStep) {
      const inProgressStep = steps.find(
        (step) => step.status?.status === 'in-progress'
      )
      setSelectedStep(inProgressStep || steps[0])
    }
  }, [steps, selectedStep])

  if (isLoading || !run) {
    return (
      <PageSection>
        <Text variant="body" theme="neutral">
          Loading workflow run...
        </Text>
      </PageSection>
    )
  }

  const status = run.status?.status || 'unknown'
  const statusDescription = run.status?.status_human_description || ''

  return (
    <PageSection className="max-w-full">
      <PageTitle title={`Run | ${app?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/apps`, text: 'Apps' },
          { path: `/${org?.id}/apps/${app?.id}`, text: app?.name },
          { path: `/${org?.id}/apps/${app?.id}/branches`, text: 'Branches' },
          { path: `/${org?.id}/apps/${app?.id}/branches/${branchId}`, text: branch?.name },
          { path: `/${org?.id}/apps/${app?.id}/branches/${branchId}/runs/${runId}`, text: 'Run' },
        ]}
      />
      <div className="flex items-start justify-between">
        <HeadingGroup>
          <Text variant="h3" weight="strong">
            Workflow run
          </Text>
          <ID>{runId}</ID>
          <div className="flex items-center gap-3 mt-2">
            <Badge theme={statusTheme(status)} size="sm">
              {status}
            </Badge>
            {statusDescription && (
              <Text variant="subtext" theme="neutral">
                {statusDescription}
              </Text>
            )}
          </div>
        </HeadingGroup>
        <div className="flex flex-col items-end gap-1">
          <Text variant="subtext" theme="neutral">
            Created <Time time={run.created_at} format="relative" />
          </Text>
          {run.started_at && (
            <Text variant="subtext" theme="neutral">
              Started <Time time={run.started_at} format="relative" />
            </Text>
          )}
          {run.finished_at && (
            <Text variant="subtext" theme="neutral">
              Finished <Time time={run.finished_at} format="relative" />
            </Text>
          )}
        </div>
      </div>

      <Card>
        <div className="p-6 min-w-0">
          <div className="flex items-center justify-between mb-4">
            <Text variant="h3" weight="strong">
              Workflow progress
            </Text>
            <Text variant="subtext" theme="neutral">
              Scroll horizontally or use trackpad to navigate
            </Text>
          </div>

          <WorkflowStepsPipeline
            steps={steps}
            selectedStepId={selectedStep?.id}
            onSelectStep={setSelectedStep}
          />
        </div>
      </Card>

      {selectedStep && (
        <WorkflowStepDetail
          step={selectedStep}
          onClose={() => setSelectedStep(null)}
        />
      )}
    </PageSection>
  )
}

export const BranchRunDetail = () => {
  const params = useParams()
  const branchId = params.branchId as string

  return (
    <BranchProvider branchId={branchId}>
      <BranchRunDetailContent />
    </BranchProvider>
  )
}
