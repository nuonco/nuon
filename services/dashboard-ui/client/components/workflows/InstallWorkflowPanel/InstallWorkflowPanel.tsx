import type { ReactNode } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Duration } from '@/components/common/Duration'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import {
  DeploymentAlerts,
  DeploymentWorkflowContent,
} from '@/components/installs/DeploymentDetail/DeploymentDetailContent'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { RunRunbookButton } from '@/components/runbooks/RunRunbook'
import { Panel, type IPanel } from '@/components/surfaces/Panel'
import { RunningClock } from '@/components/workflows/RunningClock'
import {
  ActionRunStepDetails,
  ActionRunStepDetailsComponent,
} from '@/components/workflows/step-details/action-run-details/ActionRunStepDetails'
import { WorkflowStepsSkeleton } from '@/components/workflows/WorkflowSteps'
import { useInstallLink } from '@/hooks/use-install-path'
import { useOrg } from '@/hooks/use-org'
import { useWorkflow } from '@/hooks/use-workflow'
import { getInstallRunbook, getInstallRunbookRun } from '@/lib'

export const InstallWorkflowPanel = ({
  children,
  ...props
}: Partial<IPanel> & { children: ReactNode }) => (
  <Panel heading="Workflow details" size="3/4" {...props}>
    {children}
  </Panel>
)

export const InstallWorkflowPanelContent = () => {
  const { workflow } = useWorkflow()
  const status = workflow?.status?.status ?? ''
  const activity = workflow?.status?.status_human_description
  const actionRunStep =
    workflow?.type === 'action_workflow_run'
      ? workflow?.steps?.find(
          (step) => step.step_target_type === 'install_action_workflow_runs'
        )
      : undefined

  if (actionRunStep) {
    return <ActionRunStepDetails step={actionRunStep} />
  }
  if (workflow?.type === 'action_workflow_run') {
    return <ActionRunStepDetailsComponent isLoading error={null} />
  }
  if (workflow?.type === 'runbook_run') {
    return <RunbookRunPanelContent />
  }

  return (
    <>
      <WorkflowPanelHeader title={workflow?.name || 'Workflow'} />
      {activity && status !== 'error' ? (
        <Text variant="subtext" theme="neutral">
          {activity}
        </Text>
      ) : null}
      <DeploymentAlerts />
      <div className="mt-4">
        <DeploymentWorkflowContent />
      </div>
    </>
  )
}

export const InstallWorkflowPanelLoading = ({
  status,
  createdAt,
}: {
  status?: string
  createdAt?: string
}) => (
  <>
    <div className="flex flex-wrap items-center gap-3">
      {status ? (
        <Status status={status} variant="badge" />
      ) : (
        <Status loading variant="badge" loadingWidth={8} />
      )}
      <ID loading loadingWidth={16} />
    </div>
    <div className="flex flex-wrap items-start gap-6">
      <LabeledValue label="Started">
        {createdAt ? (
          <Time time={createdAt} format="relative" variant="subtext" />
        ) : (
          <Time loading format="relative" variant="subtext" loadingWidth={10} />
        )}
      </LabeledValue>
    </div>
    <div className="mt-4 flex flex-col gap-4">
      <SectionHeader title="Workflow steps" />
      <WorkflowStepsSkeleton />
    </div>
  </>
)

const WorkflowTiming = () => {
  const { workflow } = useWorkflow()

  return (
    <>
      <LabeledValue label="Started">
        <Time time={workflow.created_at} format="relative" variant="subtext" />
      </LabeledValue>
      {workflow?.finished ? (
        <LabeledValue label="Ran for">
          <Duration variant="subtext" nanoseconds={workflow.execution_time} />
        </LabeledValue>
      ) : (
        <LabeledValue label="Running for">
          <RunningClock
            label=""
            theme="default"
            startedAt={workflow?.started_at || workflow?.created_at}
          />
        </LabeledValue>
      )}
    </>
  )
}

const WorkflowPanelHeader = ({
  title,
  links,
  children,
}: {
  title: string
  links?: ReactNode
  children?: ReactNode
}) => {
  const { workflow } = useWorkflow()
  const { org } = useOrg()
  const installLink = useInstallLink()
  const installId = workflow?.metadata?.install_id || workflow?.owner_id
  const triggeredBy = workflow?.created_by?.email || workflow?.created_by?.name

  return (
    <>
      <div className="flex items-center gap-4">
        <Text variant="base" weight="strong">
          {title}
        </Text>
        <div className="ml-auto flex items-center gap-4">
          {links}
          {installId && org?.id && workflow?.id ? (
            <Link
              href={installLink({
                orgId: org.id,
                installId,
                suffix: `/workflows/${workflow.id}`,
              })}
            >
              View details
            </Link>
          ) : null}
        </div>
      </div>
      <div className="flex flex-col gap-4">
        <div className="flex flex-wrap items-start gap-x-6 gap-y-3">
          <LabeledValue label="Status">
            <Status status={workflow?.status?.status} variant="badge" />
          </LabeledValue>
          <LabeledValue label="Triggered by">
            <Text variant="subtext">{triggeredBy ?? '—'}</Text>
          </LabeledValue>
          <WorkflowTiming />
        </div>
        {children}
      </div>
    </>
  )
}

const RunbookRunPanelContent = () => {
  const { workflow } = useWorkflow()
  const { org } = useOrg()
  const installLink = useInstallLink()
  const runbookName = workflow?.metadata?.runbook_name
  const installId = workflow?.metadata?.install_id || workflow?.owner_id
  const runbookHref =
    runbookName && installId && org?.id
      ? installLink({
          orgId: org.id,
          installId,
          suffix: `/runbooks/${encodeURIComponent(runbookName)}`,
        })
      : undefined
  const { data: installRunbook } = useQuery({
    queryKey: ['install-runbook', org?.id, installId, runbookName],
    queryFn: () =>
      getInstallRunbook({
        orgId: org!.id,
        installId: installId!,
        runbookId: runbookName!,
      }),
    enabled: !!org?.id && !!installId && !!runbookName,
  })
  const runbookRunId = workflow?.metadata?.install_runbook_run_id
  const { data: runbookRun } = useQuery({
    queryKey: ['install-runbook-run', org?.id, installId, runbookRunId],
    queryFn: () =>
      getInstallRunbookRun({
        orgId: org!.id,
        installId: installId!,
        runId: runbookRunId!,
      }),
    enabled: !!org?.id && !!installId && !!runbookRunId,
  })

  return (
    <>
      <WorkflowPanelHeader
        title={runbookName ? `Runbook: ${runbookName}` : 'Runbook run'}
        links={
          runbookHref ? <Link href={runbookHref}>View runbook</Link> : null
        }
      >
        {installRunbook ? (
          <div className="self-end">
            <RunRunbookButton
              installRunbook={installRunbook}
              initialRole={workflow?.role}
              initialValues={{
                inputs: runbookRun?.runbook_inputs,
                steps: runbookRun?.step_selections,
              }}
            >
              Re-run runbook
            </RunRunbookButton>
          </div>
        ) : null}
      </WorkflowPanelHeader>
      <DeploymentAlerts />
      <div className="mt-4">
        <DeploymentWorkflowContent />
      </div>
    </>
  )
}
