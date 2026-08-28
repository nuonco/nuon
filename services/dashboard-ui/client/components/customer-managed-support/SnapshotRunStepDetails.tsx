import { DateTime } from 'luxon'
import { PlanComponent } from '@/components/approvals/Plan'
import { Badge } from '@/components/common/Badge'
import { EmptyState } from '@/components/common/EmptyState'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { LabeledValue } from '@/components/common/LabeledValue'
import { PropertyGrid } from '@/components/common/PropertyGrid'
import { Tabs } from '@/components/common/Tabs'
import { Text } from '@/components/common/Text'
import { StepDetailPanelComponent } from '@/components/workflows/step-details/StepDetailPanel'
import type { IPanel } from '@/components/surfaces/Panel'
import type {
  TCustomerManagedSnapshotJobLog,
  TCustomerManagedSnapshotRunStep,
} from '@/lib'
import type { TWorkflowStep, TWorkflowStepApprovalType } from '@/types'
import { snakeToWords } from '@/utils/string-utils'
import { CustomerManagedSnapshotLogViewer } from './SnapshotLogViewer'
import { SnapshotCloudFormationPlan } from './SnapshotCloudFormationPlan'

type DriftResource = {
  address?: string
  action?: string
  drifted?: boolean
}

const planApprovalTypes: Record<string, TWorkflowStepApprovalType> = {
  terraform: 'terraform_plan',
  helm: 'helm_approval',
  kubernetes_manifest: 'kubernetes_manifest_approval',
  pulumi: 'pulumi_plan',
}

const workflowStepStatus = (
  status: string
): NonNullable<TWorkflowStep['status']>['status'] => {
  if (status === 'finished' || status === 'completed') return 'success'
  if (status === 'failed') return 'error'
  return status as NonNullable<TWorkflowStep['status']>['status']
}

const executionTime = (step: TCustomerManagedSnapshotRunStep) => {
  if (!step.started_at || !step.finished_at) return undefined
  const startedAt = DateTime.fromISO(step.started_at)
  const finishedAt = DateTime.fromISO(step.finished_at)
  if (!startedAt.isValid || !finishedAt.isValid) return undefined
  return finishedAt.diff(startedAt).as('milliseconds') * 1e6
}

export const toCustomerManagedWorkflowStep = (
  step: TCustomerManagedSnapshotRunStep
): TWorkflowStep => {
  const approvalType = step.plan ? planApprovalTypes[step.plan.kind] : undefined
  const createdAt = step.started_at ?? step.finished_at
  const createdAtSeconds = createdAt
    ? DateTime.fromISO(createdAt).toSeconds()
    : undefined

  return {
    id: step.id,
    name: snakeToWords(step.name),
    created_at: createdAt,
    created_by: { email: 'Customer runner' },
    execution_type:
      step.kind === 'skipped' ? 'skipped' : approvalType ? 'approval' : 'user',
    execution_time: executionTime(step),
    finished: !!step.finished_at,
    started_at: step.started_at,
    finished_at: step.finished_at,
    result_directive: step.result_directive,
    approval: approvalType
      ? {
          id: `${step.id}-captured-plan`,
          type: approvalType,
          response: { source: 'support-snapshot' },
        }
      : undefined,
    status: {
      status: workflowStepStatus(step.status),
      status_human_description: step.status_description,
      created_at_ts: createdAtSeconds,
    },
  }
}

const CapturedDrift = ({ step }: { step: TCustomerManagedSnapshotRunStep }) => {
  const drift = step.drift
  const resources = (drift?.resources ?? []) as DriftResource[]

  if (!drift) {
    return (
      <EmptyState
        variant="history"
        emptyTitle="No plan captured"
        emptyMessage="Plan details will appear when the customer captures a run containing a resource plan."
      />
    )
  }

  return (
    <div className="flex flex-col gap-6 pt-4">
      <HeadingGroup>
        <span className="flex items-center gap-2">
          <Text variant="base" weight="strong">
            Resource drift
          </Text>
          <Badge theme={drift.drifted ? 'warn' : 'success'}>
            {drift.drifted ? 'Drift detected' : 'No drift'}
          </Badge>
        </span>
        {typeof drift.summary === 'string' ? (
          <Text variant="subtext" theme="neutral">
            {drift.summary}
          </Text>
        ) : null}
      </HeadingGroup>
      <div className="flex flex-wrap gap-6">
        <LabeledValue label="Resource changes">
          <Text variant="base">{String(drift.resource_changes ?? 0)}</Text>
        </LabeledValue>
        <LabeledValue label="Output changes">
          <Text variant="base">{String(drift.output_changes ?? 0)}</Text>
        </LabeledValue>
        <LabeledValue label="Drifted resources">
          <Text variant="base">{String(drift.resource_drift ?? 0)}</Text>
        </LabeledValue>
      </div>
      <PropertyGrid<DriftResource>
        values={resources}
        columns={[
          { key: 'address', header: 'Resource' },
          {
            key: 'action',
            header: 'Action',
            render: (value) => (
              <Badge variant="code">{String(value ?? 'unknown')}</Badge>
            ),
          },
          {
            key: 'drifted',
            header: 'Result',
            render: (value, resource) =>
              value ? (
                <Badge theme="warn">Drifted</Badge>
              ) : resource.action === 'noop' ? (
                <Badge theme="success">No change</Badge>
              ) : (
                <Badge theme="neutral">Planned change</Badge>
              ),
          },
        ]}
        emptyStateProps={{
          variant: 'table',
          emptyTitle: 'No resource changes captured',
          emptyMessage:
            'Resource changes will appear when the plan detects drift.',
        }}
      />
      {drift.resources_truncated ? (
        <Badge theme="warn">Resource list truncated</Badge>
      ) : null}
    </div>
  )
}

const CapturedStepDetails = ({
  workflowStep,
  snapshotStep,
  log,
  capturedAt,
}: {
  workflowStep: TWorkflowStep
  snapshotStep: TCustomerManagedSnapshotRunStep
  log?: TCustomerManagedSnapshotJobLog
  capturedAt?: string
}) => {
  const tabs: Record<string, React.ReactNode> = {}
  if (snapshotStep.plan) {
    tabs.plan = (
      <div className="mt-4">
        {snapshotStep.plan.kind === 'cloudformation' ? (
          <SnapshotCloudFormationPlan plan={snapshotStep.plan.content} />
        ) : (
          <PlanComponent
            step={workflowStep}
            plan={snapshotStep.plan.content}
            isLoading={false}
            error={undefined}
          />
        )}
      </div>
    )
  } else if (snapshotStep.drift) {
    tabs.plan = <CapturedDrift step={snapshotStep} />
  }
  if (log) {
    tabs.logs = (
      <div className="pt-4">
        <CustomerManagedSnapshotLogViewer log={log} capturedAt={capturedAt} />
      </div>
    )
  }

  if (!Object.keys(tabs).length) {
    return (
      <EmptyState
        variant="history"
        emptyTitle="No plan or logs captured"
        emptyMessage="Details will appear when the customer captures a run containing a plan or job logs."
      />
    )
  }

  return <Tabs tabs={tabs} />
}

export const CustomerManagedSnapshotRunStepDetails = ({
  step,
  log,
  capturedAt,
  ...props
}: {
  step: TCustomerManagedSnapshotRunStep
  log?: TCustomerManagedSnapshotJobLog
  capturedAt?: string
} & IPanel) => {
  const workflowStep = toCustomerManagedWorkflowStep(step)

  return (
    <StepDetailPanelComponent
      step={workflowStep}
      planOnly
      size="3/4"
      {...props}
    >
      <CapturedStepDetails
        workflowStep={workflowStep}
        snapshotStep={step}
        log={log}
        capturedAt={capturedAt}
      />
    </StepDetailPanelComponent>
  )
}
