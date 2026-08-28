import { DateTime } from 'luxon'
import { useParams, useSearchParams } from 'react-router'
import { Banner } from '@/components/common/Banner'
import { EmptyState } from '@/components/common/EmptyState'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { WorkflowStepsComponent } from '@/components/workflows/WorkflowSteps'
import { WorkflowHeader } from '@/components/workflows/workflow-details/WorkflowHeader'
import { WorkflowMetrics } from '@/components/workflows/workflow-details/WorkflowMetrics'
import { WorkflowStatusSection } from '@/components/workflows/workflow-details/WorkflowStatusSection'
import { WorkflowDetailsSection } from '@/components/workflows/workflow-details/WorkflowDetailsSection'
import { useCustomerManagedSupportSnapshot } from '@/hooks/use-customer-managed-support-snapshot'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import type { TCustomerManagedSnapshotRun } from '@/lib'
import type { TWorkflow, TWorkflowStep } from '@/types'
import { snakeToWords, toSentenceCase } from '@/utils/string-utils'
import { CustomerManagedSnapshotContent } from './SnapshotEmpty'
import {
  CustomerManagedSnapshotRunStepDetails,
  toCustomerManagedWorkflowStep,
} from './SnapshotRunStepDetails'

const runExecutionTime = (run: TCustomerManagedSnapshotRun) => {
  if (!run.started_at || !run.finished_at) return undefined
  const startedAt = DateTime.fromISO(run.started_at)
  const finishedAt = DateTime.fromISO(run.finished_at)
  if (!startedAt.isValid || !finishedAt.isValid) return undefined
  return finishedAt.diff(startedAt).as('milliseconds') * 1e6
}

const toWorkflow = (
  run: TCustomerManagedSnapshotRun,
  steps: TWorkflowStep[]
): TWorkflow => ({
  id: run.run_id,
  name: run.ref_name || run.ref_id || 'Install run',
  type: run.source as TWorkflow['type'],
  created_at: run.started_at,
  started_at: run.started_at,
  finished_at: run.finished_at,
  finished: !!run.finished_at,
  execution_time: runExecutionTime(run),
  created_by: { email: 'Customer runner' },
  result_directive: run.result_directive,
  status: {
    status:
      run.status === 'finished' || run.status === 'completed'
        ? 'success'
        : run.status === 'failed'
          ? 'error'
          : (run.status as NonNullable<TWorkflow['status']>['status']),
    status_human_description:
      run.error || toSentenceCase(snakeToWords(run.status)),
  },
  steps,
})

export const CustomerManagedSnapshotWorkflowDetail = () => {
  const { workflowId } = useParams()
  const { org } = useOrg()
  const { install } = useInstall()
  const { snapshot } = useCustomerManagedSupportSnapshot()
  const { addPanel } = useSurfaces()
  const [searchParams] = useSearchParams()
  const run = snapshot?.snapshot.runs.find(
    ({ run_id }) => run_id === workflowId
  )
  const snapshotId = searchParams.get('snapshot')
  const runsHref = `/${org.id}/installs/${install.id}/workflows${snapshotId ? `?snapshot=${snapshotId}` : ''}`

  if (!run) {
    return (
      <PageSection>
        <CustomerManagedSnapshotContent>
          <EmptyState
            variant="table"
            emptyTitle="Run not found"
            emptyMessage="Select a run captured in this support bundle."
          />
        </CustomerManagedSnapshotContent>
      </PageSection>
    )
  }

  const workflowSteps = run.steps.map(toCustomerManagedWorkflowStep)
  const workflow = toWorkflow(run, workflowSteps)
  const completedSteps = workflowSteps.filter(
    ({ status }) => status?.status === 'success'
  ).length
  const discardedSteps = workflowSteps.filter(({ status }) =>
    ['discarded', 'skipped', 'user-skipped', 'auto-skipped'].includes(
      status?.status ?? ''
    )
  ).length
  const runName = workflow.name || 'Install run'

  const openStepDetails = (workflowStep: TWorkflowStep) => {
    const step = run.steps.find(({ id }) => id === workflowStep.id)
    if (!step) return
    const logs = snapshot?.snapshot.logs ?? []
    const log = logs.find(({ job_id }) => job_id === step.job_id)
    addPanel(
      <CustomerManagedSnapshotRunStepDetails
        panelKey={`${run.run_id}-${step.id}`}
        step={step}
        log={log}
        capturedAt={snapshot?.captured_at}
      />,
      `${run.run_id}-${step.id}`
    )
  }

  return (
    <PageSection className="!gap-2">
      <PageTitle title={`${runName} | ${install.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org.id}`, text: org.name },
          { path: `/${org.id}/installs`, text: 'Installs' },
          { path: `/${org.id}/installs/${install.id}`, text: install.name },
          { path: runsHref, text: 'Runs' },
          {
            path: `/${org.id}/installs/${install.id}/workflows/${run.run_id}`,
            text: runName,
          },
        ]}
      />

      <CustomerManagedSnapshotContent>
        <div className="flex flex-col gap-2">
          <WorkflowHeader workflow={workflow} install={install} readOnly />

          <WorkflowMetrics
            workflow={workflow}
            pendingApprovalsCount={0}
            policyViolationsCount={0}
            discardedStepsCount={discardedSteps}
            completedStepsCount={completedSteps}
            totalSteps={workflowSteps.length}
          />

          <WorkflowStatusSection workflow={workflow} />

          {run.error ? (
            <Banner theme="error">
              <Text weight="strong">Run failed</Text>
              <Text variant="subtext">{run.error}</Text>
            </Banner>
          ) : null}

          <WorkflowDetailsSection
            workflow={workflow}
            orgId={org.id}
            install={install}
            noun="run"
          />
        </div>

        <div className="flex flex-col gap-6 mt-6">
          <Text variant="h3" weight="strong">
            Run steps
          </Text>
          <WorkflowStepsComponent
            workflowSteps={workflowSteps}
            readOnly
            noun="run"
            onViewDetails={openStepDetails}
            eagerStepsLoaded
            allStepsLoaded
          />
        </div>
      </CustomerManagedSnapshotContent>
    </PageSection>
  )
}
