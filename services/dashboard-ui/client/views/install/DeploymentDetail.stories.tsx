export default {
  title: 'Views / Installs / Deployment details',
}

import type { ReactNode } from 'react'
import { Navigate, Outlet, Route, Routes } from 'react-router'
import { ApprovalBanner } from '@/components/approvals/ApprovalBanner'
import { AppConfigFilesDiff } from '@/components/branches/ComponentConfigDiff/ComponentConfigDiff'
import { Banner } from '@/components/common/Banner'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { DeploymentDetail } from '@/components/installs/DeploymentDetail'
import { WorkflowAlertBanners } from '@/components/workflows/WorkflowDetails'
import { WorkflowChangesSummary } from '@/components/workflows/WorkflowChangesSummary'
import { WorkflowStepsComponent } from '@/components/workflows/WorkflowSteps'
import { WorkflowContext } from '@/providers/workflow-provider'
import type {
  TInstallDeploymentRecord,
  TInstallDeploymentStatus,
  TStepChangeSummary,
  TWorkflow,
  TWorkflowStep,
} from '@/types'

const BASE_PATH = '/deployments/wf-deploy-1'

const deployment = (
  status: TInstallDeploymentStatus
): TInstallDeploymentRecord => ({
  id: 'wf-deploy-1',
  type: 'app_branch_update',
  status,
  created_at: '2026-10-02T16:00:00Z',
  title: 'Update production install',
  summary: 'Apply the latest app template to this install.',
  workflow: {
    id: 'wf-deploy-1',
    name: 'Update production install',
    type: 'app_branch_config_update',
  },
  app_branch: {
    id: 'branch-main',
    name: 'main',
    run_id: 'run-main-42',
    sha: 'abc123def456',
  },
  affected_resources: {
    stack: true,
    components: ['api', 'worker'],
    images: ['ghcr.io/acme/api'],
  },
  change_groups: [],
})

const workflow = (
  status: string,
  description: string,
  metadata?: Record<string, unknown>
): TWorkflow =>
  ({
    id: 'wf-deploy-1',
    name: 'Update production install',
    type: 'app_branch_config_update',
    created_at: '2026-10-02T16:00:00Z',
    status: {
      status,
      status_human_description: description,
      metadata,
    },
  }) as TWorkflow

const step = (
  id: string,
  name: string,
  groupIdx: number,
  status: string,
  extra?: Record<string, unknown>
): TWorkflowStep =>
  ({
    id,
    name,
    execution_type: extra?.execution_type ?? 'system',
    group_idx: groupIdx,
    status: { status, ...(extra?.status as object) },
    finished: extra?.finished,
    execution_time: extra?.execution_time,
    approval: extra?.approval,
  }) as TWorkflowStep

const done = (
  id: string,
  name: string,
  groupIdx: number,
  nanoseconds: number
) =>
  step(id, name, groupIdx, 'success', {
    finished: true,
    execution_time: nanoseconds,
  })

const summary = (
  fields: Omit<TStepChangeSummary, 'counts' | 'countsState' | 'hasDetail'> &
    Partial<Pick<TStepChangeSummary, 'counts' | 'countsState' | 'hasDetail'>>
): TStepChangeSummary => ({
  counts: { create: 0, update: 0, delete: 0, replace: 0, noop: 0 },
  countsState: 'ok',
  hasDetail: true,
  ...fields,
})

const approvalStep = step(
  'step-stack-approval',
  'Plan stack',
  1,
  'approval-awaiting',
  {
    execution_type: 'approval',
    approval: { id: 'approval-1', type: 'terraform_plan' },
  }
)

const failedStep = step('step-deploy', 'Deploy components', 2, 'error', {
  finished: true,
  status: {
    status: 'error',
    status_human_description: 'Helm upgrade failed for api',
  },
})

const inProgressSummaries: TStepChangeSummary[] = [
  summary({
    stepId: 'step-stack',
    stepName: 'Plan stack',
    componentName: 'Install stack',
    planType: 'terraform_plan',
    status: 'approved',
    counts: { create: 2, update: 3, delete: 0, replace: 0, noop: 8 },
  }),
  summary({
    stepId: 'step-api',
    stepName: 'Plan api',
    componentName: 'api',
    planType: 'helm_approval',
    status: 'generating',
    countsState: 'unknown',
    hasDetail: false,
  }),
]

const awaitingSummaries: TStepChangeSummary[] = [
  summary({
    stepId: 'step-stack',
    stepName: 'Plan stack',
    componentName: 'Install stack',
    planType: 'terraform_plan',
    status: 'pending-approval',
    counts: { create: 2, update: 3, delete: 0, replace: 0, noop: 8 },
  }),
  summary({
    stepId: 'step-api',
    stepName: 'Plan api',
    componentName: 'api',
    planType: 'helm_approval',
    status: 'pending-approval',
    counts: { create: 0, update: 2, delete: 0, replace: 0, noop: 5 },
  }),
]

const succeededSummaries: TStepChangeSummary[] = [
  summary({
    stepId: 'step-stack',
    stepName: 'Plan stack',
    componentName: 'Install stack',
    planType: 'terraform_plan',
    status: 'applied',
    counts: { create: 2, update: 3, delete: 0, replace: 0, noop: 8 },
  }),
  summary({
    stepId: 'step-api',
    stepName: 'Plan api',
    componentName: 'api',
    planType: 'helm_approval',
    status: 'applied',
    counts: { create: 0, update: 2, delete: 0, replace: 0, noop: 5 },
  }),
]

const failedSummaries: TStepChangeSummary[] = [
  summary({
    stepId: 'step-stack',
    stepName: 'Plan stack',
    componentName: 'Install stack',
    planType: 'terraform_plan',
    status: 'applied',
    counts: { create: 2, update: 3, delete: 0, replace: 0, noop: 8 },
  }),
  summary({
    stepId: 'step-api',
    stepName: 'Plan api',
    componentName: 'api',
    planType: 'helm_approval',
    status: 'error',
    countsState: 'unknown',
    hasDetail: true,
  }),
]

const inProgressSteps = [
  done('step-sync', 'Sync install configuration', 0, 1_800_000_000),
  done('step-stack', 'Plan stack', 1, 8_200_000_000),
  step('step-deploy', 'Deploy components', 2, 'in-progress'),
]

const awaitingSteps = [
  done('step-sync', 'Sync install configuration', 0, 1_800_000_000),
  approvalStep,
]

const succeededSteps = [
  done('step-sync', 'Sync install configuration', 0, 1_800_000_000),
  done('step-stack', 'Plan stack', 1, 8_200_000_000),
  done('step-deploy', 'Deploy components', 2, 12_400_000_000),
]

const failedSteps = [
  done('step-sync', 'Sync install configuration', 0, 1_800_000_000),
  done('step-stack', 'Plan stack', 1, 8_200_000_000),
  failedStep,
]

type TScenario = {
  approvalPrompt?: boolean
  banners?: ReactNode
  deployment: TInstallDeploymentRecord
  steps: TWorkflowStep[]
  summaries: TStepChangeSummary[]
  workflow: TWorkflow
}

const ChangesPane = ({ summaries }: { summaries: TStepChangeSummary[] }) => (
  <WorkflowChangesSummary
    summaries={summaries}
    renderDetail={(item) => (
      <Banner theme="neutral">{item.componentName} plan details</Banner>
    )}
  />
)

const TemplatePane = () => (
  <AppConfigFilesDiff
    title="Template updates"
    previousVersion="abc123d"
    currentVersion="def456a"
    configSections={[
      {
        name: 'Components',
        sectionKey: 'components',
        additions: 0,
        removals: 0,
        changed: 1,
        grouped: true,
        fields: [],
        entities: [
          {
            name: 'api',
            op: 'change',
            componentType: 'helm_chart',
            fields: [],
            files: [{ name: 'components/api.yaml', op: 'change' }],
          },
        ],
      },
    ]}
    files={[
      {
        path: 'components/api.yaml',
        kind: 'file',
        change: 'modified',
        before: 'replicas: 2\nimage: ghcr.io/acme/api:v1\n',
        after: 'replicas: 3\nimage: ghcr.io/acme/api:v2\n',
      },
    ]}
  />
)

const WorkflowPane = ({
  approvalPrompt,
  steps,
}: {
  approvalPrompt?: boolean
  steps: TWorkflowStep[]
}) => (
  <div className="flex flex-col gap-4">
    <SectionHeader
      title="Workflow steps"
      description="Follow each step in this deployment."
    />
    <WorkflowStepsComponent
      workflowSteps={steps}
      eagerStepsLoaded
      allStepsLoaded
      approvalPrompt={approvalPrompt}
    />
  </div>
)

const Frame = ({ scenario }: { scenario: TScenario }) => (
  <DeploymentDetail
    banners={scenario.banners}
    basePath={BASE_PATH}
    branchHref="/branches/branch-main"
    deployment={scenario.deployment}
    workflow={scenario.workflow}
  >
    <Outlet />
  </DeploymentDetail>
)

const workflowContext = (scenario: TScenario) => {
  const failed = scenario.steps.filter(
    (item) => item.status?.status === 'error'
  )
  const pending = scenario.steps.filter(
    (item) => item.status?.status === 'approval-awaiting'
  )
  const completed = scenario.steps.filter(
    (item) => item.status?.status === 'success'
  )

  return {
    workflow: scenario.workflow,
    stopPolling: () => {},
    workflowSteps: scenario.steps,
    hasApprovals: pending.length > 0,
    failedSteps: failed,
    pendingApprovals: pending,
    discardedSteps: [],
    completedSteps: completed,
    stepsWithPolicyViolations: [],
    totalSteps: scenario.steps.length,
    pendingApprovalsCount: pending.length,
    discardedStepsCount: 0,
    completedStepsCount: completed.length,
    failedStepsCount: failed.length,
    policyViolationsCount: 0,
  }
}

const DeploymentView = ({ scenario }: { scenario: TScenario }) => (
  <WorkflowContext.Provider value={workflowContext(scenario)}>
    <div className="mx-auto w-full max-w-6xl">
      <Routes>
        <Route path={BASE_PATH} element={<Frame scenario={scenario} />}>
          <Route
            index
            element={<ChangesPane summaries={scenario.summaries} />}
          />
          <Route path="template-updates" element={<TemplatePane />} />
          <Route
            path="workflow"
            element={
              <WorkflowPane
                approvalPrompt={scenario.approvalPrompt}
                steps={scenario.steps}
              />
            }
          />
        </Route>
        <Route path="*" element={<Navigate to={BASE_PATH} replace />} />
      </Routes>
    </div>
  </WorkflowContext.Provider>
)

export const InProgress = () => (
  <DeploymentView
    scenario={{
      deployment: deployment('in-progress'),
      workflow: workflow('in-progress', 'Deploying components'),
      summaries: inProgressSummaries,
      steps: inProgressSteps,
    }}
  />
)

export const AwaitingApproval = () => (
  <DeploymentView
    scenario={{
      approvalPrompt: true,
      banners: <ApprovalBanner step={approvalStep} />,
      deployment: deployment('pending'),
      workflow: workflow(
        'approval-awaiting',
        'Waiting for plan approval'
      ),
      summaries: awaitingSummaries,
      steps: awaitingSteps,
    }}
  />
)

export const Succeeded = () => (
  <DeploymentView
    scenario={{
      deployment: deployment('success'),
      workflow: workflow('success', 'Deployment finished'),
      summaries: succeededSummaries,
      steps: succeededSteps,
    }}
  />
)

export const Failed = () => {
  const failedWorkflow = workflow('error', 'Helm upgrade failed for api')

  return (
    <DeploymentView
      scenario={{
        banners: (
          <WorkflowAlertBanners
            workflow={failedWorkflow}
            failedSteps={[failedStep]}
          />
        ),
        deployment: deployment('error'),
        workflow: failedWorkflow,
        summaries: failedSummaries,
        steps: failedSteps,
      }}
    />
  )
}
