import { useState } from 'react'
import { DateTime } from 'luxon'
import { Text } from '@/components/common/Text'
import { Timeline } from '@/components/common/Timeline'
import { Select } from '@/components/common/form/Select'
import {
  ResourceOutcomes,
  type TDeploymentRun,
} from '@/components/installs/DeploymentDetail/DeploymentProgress'
import {
  ACTIVE_DEPLOYMENT_STATUSES,
  deploymentOutcomes,
  deploymentSteps,
  stepAffectedResources,
} from '@/components/installs/DeploymentDetail/deployment-progress'
import { DeploymentCard } from '@/components/installs/DeploymentsList/DeploymentCard'
import { DeploymentPolicySummary } from '@/components/installs/DeploymentsList/DeploymentPolicySummary'
import { DeploymentRow } from '@/components/installs/DeploymentsList/DeploymentRow'
import { ListPage } from '@/components/layout/ListPage'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { Panel, type IPanel } from '@/components/surfaces/Panel'
import { PolicyViolations } from '@/components/workflows/step-details/PolicyViolations'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useTheme } from '@/hooks/use-theme'
import type { TWorkflowStep } from '@/types'
import { getPolicyViolationCounts } from '@/utils/workflow-utils'

type TPolicyPreviewRun = Omit<TDeploymentRun, 'steps'> & {
  id: string
  title: string
  created_at: string
  steps: TWorkflowStep[]
}

const warningMetadata = {
  deny_violations: [],
  warn_violations: [
    {
      policy_id: 'policy-replicas',
      severity: 'warn',
      message: 'One replica configured; two recommended.',
    },
  ],
  passed_policy_ids: ['policy-image-tag'],
}

const manyWarningsMetadata = {
  ...warningMetadata,
  warn_violations: [
    ...warningMetadata.warn_violations,
    ...[
      'The image tag is mutable; pin a digest for reproducible deployments.',
      'CPU requests are below the recommended minimum.',
      'Memory requests are below the recommended minimum.',
      'A readiness probe is recommended.',
      'A liveness probe is recommended.',
      'A startup probe is recommended for slow-starting containers.',
      'A pod disruption budget is recommended.',
      'Spread replicas across availability zones.',
      'Spread replicas across nodes.',
      'A network policy is recommended.',
      'Consider a read-only root filesystem.',
      'An explicit non-root user is recommended.',
      'Drop unused Linux capabilities.',
      'A seccomp profile is recommended.',
      'Disable service account token automount when unused.',
      'An explicit termination grace period is recommended.',
      'A graceful shutdown hook is recommended.',
      'An autoscaling policy is recommended.',
      'A minimum replica count of two is recommended for autoscaling.',
      'Resource ownership labels are missing.',
      'Resource environment labels are missing.',
      'A metrics endpoint is recommended.',
      'Structured application logs are recommended.',
      'Review the configured ingress timeout for long-running requests.',
    ].map((message, index) => ({
      policy_id: `policy-recommendation-${index + 1}`,
      severity: 'warn',
      message,
    })),
  ],
}

const deniedMetadata = {
  ...warningMetadata,
  deny_violations: [
    {
      policy_id: 'policy-resource-limits',
      severity: 'deny',
      message: 'CPU and memory limits are required.',
    },
    {
      policy_id: 'policy-privileged',
      severity: 'deny',
      message: 'Privileged containers are not allowed.',
    },
  ],
}

const componentRun = (
  id: string,
  component: string,
  status: 'in-progress' | 'success' | 'error',
  minutesAgo: number,
  policyMetadata?: NonNullable<TWorkflowStep['status']>['metadata']
): TPolicyPreviewRun => {
  const names = [
    `fetch ${component} config`,
    `render ${component} config`,
    `plan ${component}`,
    `apply ${component}`,
    'verify readiness',
  ]
  const statuses = {
    'in-progress': ['success', 'success', 'success', 'in-progress', 'pending'],
    success: ['success', 'success', 'success', 'success', 'success'],
    error: ['success', 'success', 'error', 'pending', 'pending'],
  }[status]
  const steps = names.map(
    (name, index) =>
      ({
        id: `${id}-step-${index}`,
        name,
        execution_type: index === 2 ? 'approval' : 'system',
        group_idx: index + 1,
        metadata: { component_name: component },
        status: {
          status: statuses[index],
          metadata: index === 2 ? policyMetadata : undefined,
          history: [],
        },
      }) as TWorkflowStep
  )
  const evidence = {
    status: { status },
    finished: status !== 'in-progress',
    steps,
  }
  return {
    id,
    title: `Deploy ${component}`,
    created_at: DateTime.now().minus({ minutes: minutesAgo }).toISO(),
    status,
    activity: status === 'error' ? 'Policy check failed' : '',
    steps,
    outcomes: deploymentOutcomes(
      { affected_resources: stepAffectedResources(steps) },
      evidence
    ),
  }
}

const PolicyDetails = ({
  run,
  ...props
}: IPanel & { run: TPolicyPreviewRun }) => (
  <Panel heading={run.title} aria-label="Deployment policy details" {...props}>
    <Text theme="neutral" variant="subtext">
      Fixture data only. Denials stop the deployment before apply; warnings do
      not block deployment.
    </Text>
    <ResourceOutcomes run={run} />
    {run.steps
      .filter((step) => getPolicyViolationCounts(step).hasViolations)
      .map((step) => (
        <PolicyViolations key={step.id} step={step} />
      ))}
  </Panel>
)

export const DeploymentPoliciesPreview = () => {
  const { addPanel } = useSurfaces()
  const { preference, setPreference } = useTheme()
  const [warningCount, setWarningCount] = useState('25')
  const warnings =
    warningCount === '25' ? manyWarningsMetadata : warningMetadata
  const runs = [
    componentRun('running-worker', 'worker', 'in-progress', 12, warnings),
    componentRun('running-api', 'api', 'in-progress', 4),
    componentRun('denied-api', 'api', 'error', 8, deniedMetadata),
    componentRun('completed-worker', 'worker', 'success', 24, warnings),
    componentRun('completed-scheduler', 'scheduler', 'success', 60),
  ]
  const active = runs.filter((run) =>
    ACTIVE_DEPLOYMENT_STATUSES.has(run.status)
  )
  const history = runs.filter(
    (run) => !ACTIVE_DEPLOYMENT_STATUSES.has(run.status)
  )
  const summaries = (run: TPolicyPreviewRun, history = false) => (
    <DeploymentPolicySummary
      history={history}
      steps={deploymentSteps({ steps: run.steps }).map((step) => {
        const { denyCount, warnCount, warnViolations } =
          getPolicyViolationCounts(step)
        return {
          policy: {
            deny_count: denyCount,
            warn_count: warnCount,
            first_warn_message: warnViolations[0]?.message,
          },
        }
      })}
    />
  )

  return (
    <main
      aria-label="Deployment policies preview"
      className="@container mx-auto w-full max-w-7xl p-4 md:p-6"
    >
      <ListPage
        title="Deployments"
        description="Follow rollouts for acme-preview. View details to inspect policy results."
        actions={
          <>
            <Select
              id="policy-preview-warning-count"
              labelProps={{ labelText: 'Warnings' }}
              value={warningCount}
              onChange={setWarningCount}
              options={[
                { value: '1', label: '1 warning' },
                { value: '25', label: '25 warnings' },
              ]}
            />
            <Select
              id="policy-preview-theme"
              labelProps={{ labelText: 'Theme' }}
              value={preference}
              onChange={(value) => {
                if (value === 'system' || value === 'light' || value === 'dark')
                  setPreference(value)
              }}
              options={[
                { value: 'system', label: 'System' },
                { value: 'light', label: 'Light' },
                { value: 'dark', label: 'Dark' },
              ]}
            />
          </>
        }
      >
        <section aria-label="In progress" className="flex flex-col gap-3">
          <SectionHeader
            title={`In progress (${active.length})`}
            description="Warnings do not block deployment. Policy denials are terminal and appear in History."
          />
          <div className="grid grid-cols-1 gap-3 @4xl:grid-cols-2">
            {active.map((run) => (
              <DeploymentCard
                key={run.id}
                run={run}
                title={run.title}
                typeLabel="Component deploy"
                createdAt={run.created_at}
                onViewDetails={() => addPanel(<PolicyDetails run={run} />)}
              >
                {summaries(run)}
              </DeploymentCard>
            ))}
          </div>
        </section>
        <section aria-label="History" className="flex flex-col gap-3">
          <SectionHeader
            title="History"
            description="Policy results remain visible after the deployment finishes."
          />
          <Timeline
            events={history}
            groupByDate={false}
            pagination={{ hasNext: false, offset: 0 }}
            getEventKey={(run) => run.id}
            renderEvent={(run) => (
              <DeploymentRow
                run={run}
                title={run.title}
                createdAt={run.created_at}
                onViewDetails={() => addPanel(<PolicyDetails run={run} />)}
                history
              >
                {summaries(run, true)}
              </DeploymentRow>
            )}
          />
        </section>
      </ListPage>
    </main>
  )
}
