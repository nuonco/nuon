import { useState, type ReactNode } from 'react'
import { useLocation } from 'react-router'
import { DateTime } from 'luxon'
import { ApprovalBanner } from '@/components/approvals/ApprovalBanner'
import { ApprovePlanModal } from '@/components/approvals/ApprovePlan/ApprovePlan'
import { DenyPlanModal } from '@/components/approvals/DenyPlan/DenyPlan'
import { RetryPlanModal } from '@/components/approvals/RetryPlan/RetryPlan'
import type { DiffSectionData } from '@/components/branches/AppConfigDiff'
import { AppConfigFilesDiff } from '@/components/branches/ComponentConfigDiff/ComponentConfigDiff'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { CheckboxFilterDropdown } from '@/components/common/CheckboxFilterDropdown'
import { CodeBlock } from '@/components/common/CodeBlock'
import { EmptyState } from '@/components/common/EmptyState'
import { Select } from '@/components/common/form/Select'
import { ID } from '@/components/common/ID'
import { Icon } from '@/components/common/Icon'
import { RadioFilterDropdown } from '@/components/common/RadioFilterDropdown'
import { SearchInput } from '@/components/common/SearchInput'
import { Tabs } from '@/components/common/Tabs'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Timeline } from '@/components/common/Timeline'
import { DetailHeader } from '@/components/layout/DetailHeader'
import { DetailPage } from '@/components/layout/DetailPage'
import { ListPage } from '@/components/layout/ListPage'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { DeploymentDetail } from '@/components/installs/DeploymentDetail'
import { DeploymentRunStatus as PreviewRunStatus } from './DeploymentProgress'
import { DeploymentRow } from '@/components/installs/DeploymentsList/DeploymentRow'
import {
  DEPLOYMENT_TYPE_LABELS,
  type IDeploymentFilter,
} from '@/components/installs/DeploymentsList/DeploymentsListPresenter'
import { Panel, type IPanel } from '@/components/surfaces/Panel'
import { WorkflowChangesSummary } from '@/components/workflows/WorkflowChangesSummary'
import { WorkflowStepsComponent } from '@/components/workflows/WorkflowSteps'
import { useSurfaces } from '@/hooks/use-surfaces'
import {
  APPROVAL_MODAL_COPY,
  DENY_MODAL_COPY,
  RETRY_MODAL_COPY,
} from '@/utils/approval-utils'
import {
  datePresetQueryParameter,
  WORKFLOW_DATE_LABELS,
  WORKFLOW_STATUS_GROUPS,
  WORKFLOW_STATUS_LABELS,
  workflowStatusOptions,
  type TWorkflowDatePreset,
  type TWorkflowStatusOption,
} from '@/utils/workflow-filters'
import type {
  TInstallDeploymentRecord,
  TInstallDeploymentRecordType,
  TStepChangeSummary,
  TWorkflow,
  TWorkflowStep,
} from '@/types'

const deployment: TInstallDeploymentRecord = {
  id: 'wf-deploy-1',
  type: 'app_branch_update',
  status: 'in-progress',
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
}

const workflow = {
  id: 'wf-deploy-1',
  name: 'Update production install',
  type: 'app_branch_config_update',
  created_at: '2026-10-02T16:00:00Z',
  status: {
    status: 'approval-awaiting',
    status_human_description: 'Waiting for plan approval',
  },
} as TWorkflow

const summaries: TStepChangeSummary[] = [
  {
    stepId: 'step-stack',
    stepName: 'Plan stack',
    componentName: 'Install stack',
    planType: 'terraform_plan',
    status: 'pending-approval',
    counts: { create: 2, update: 3, delete: 0, replace: 0, noop: 8 },
    countsState: 'ok',
    hasDetail: true,
  },
  {
    stepId: 'step-api',
    stepName: 'Plan api',
    componentName: 'api',
    planType: 'helm_approval',
    status: 'approved',
    counts: { create: 0, update: 2, delete: 0, replace: 0, noop: 5 },
    countsState: 'ok',
    hasDetail: true,
  },
]

const steps = [
  {
    id: 'step-sync',
    name: 'Sync install configuration',
    execution_type: 'system',
    group_idx: 0,
    finished: true,
    execution_time: 1_800_000_000,
    status: { status: 'success' },
  },
  {
    id: 'step-stack',
    name: 'Plan stack',
    execution_type: 'system',
    group_idx: 1,
    finished: true,
    execution_time: 8_200_000_000,
    status: { status: 'success' },
  },
  {
    id: 'step-deploy',
    name: 'Deploy components',
    execution_type: 'system',
    group_idx: 2,
    status: { status: 'in-progress' },
  },
] as TWorkflowStep[]

const approvalStep = {
  id: 'step-stack-approval',
  name: 'Plan stack',
  execution_type: 'approval',
  status: { status: 'approval-awaiting' },
  approval: { id: 'approval-1', type: 'terraform_plan' },
} as TWorkflowStep

const Page = ({
  activeTabIndex,
  children,
}: {
  activeTabIndex: number
  children: ReactNode
}) => (
  <div className="mx-auto w-full max-w-6xl">
    <DeploymentDetail
      activeTabIndex={activeTabIndex}
      banners={<ApprovalBanner step={approvalStep} />}
      basePath="/deployments/wf-deploy-1"
      branchHref="/branches/branch-main"
      deployment={deployment}
      workflow={workflow}
    >
      {children}
    </DeploymentDetail>
  </div>
)

export const Changes = () => (
  <Page activeTabIndex={0}>
    <WorkflowChangesSummary
      summaries={summaries}
      renderDetail={(summary) => (
        <Banner theme="neutral">{summary.componentName} plan details</Banner>
      )}
    />
  </Page>
)

export const TemplateUpdates = () => (
  <Page activeTabIndex={1}>
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
  </Page>
)

export const Workflow = () => (
  <Page activeTabIndex={2}>
    <div className="flex flex-col gap-4">
      <SectionHeader
        title="Workflow steps"
        description="Follow each step in this deployment."
      />
      <WorkflowStepsComponent
        workflowSteps={steps}
        eagerStepsLoaded
        allStepsLoaded
        approvalPrompt
      />
    </div>
  </Page>
)

const sandboxSections: DiffSectionData[] = [
  {
    name: 'Components',
    sectionKey: 'components',
    additions: 1,
    removals: 1,
    changed: 1,
    grouped: true,
    fields: [],
    entities: [
      {
        name: 'api',
        op: 'change',
        componentType: 'helm_chart',
        fields: [],
        files: [{ name: 'components/api/values.yaml', op: 'change' }],
      },
      {
        name: 'worker',
        op: 'add',
        componentType: 'docker_build',
        fields: [],
        files: [{ name: 'components/worker/Dockerfile', op: 'add' }],
      },
      {
        name: 'legacy',
        op: 'remove',
        componentType: 'kubernetes_manifest',
        fields: [],
        files: [{ name: 'components/legacy/service.yaml', op: 'remove' }],
      },
    ],
  },
  {
    name: 'Sandbox',
    sectionKey: 'sandbox',
    additions: 0,
    removals: 0,
    changed: 1,
    grouped: false,
    fields: [],
    entities: [],
    files: [{ name: 'sandbox/main.tf', op: 'change' }],
  },
]

const sandboxFiles = [
  {
    path: 'components/api/values.yaml',
    kind: 'helm values',
    change: 'modified' as const,
    before:
      'replicaCount: 2\nimage:\n  repository: ghcr.io/acme/api\n  tag: v1.2.3\nresources:\n  requests:\n    cpu: 250m\n    memory: 512Mi\n',
    after:
      'replicaCount: 3\nimage:\n  repository: ghcr.io/acme/api\n  tag: v1.3.0\nresources:\n  requests:\n    cpu: 500m\n    memory: 1Gi\n',
  },
  {
    path: 'components/worker/Dockerfile',
    kind: 'dockerfile',
    change: 'added' as const,
    after:
      'FROM ghcr.io/acme/worker-base:2\nCOPY bin/worker /usr/local/bin/worker\nENTRYPOINT ["/usr/local/bin/worker"]\n',
  },
  {
    path: 'components/legacy/service.yaml',
    kind: 'kubernetes manifest',
    change: 'removed' as const,
    before:
      'apiVersion: v1\nkind: Service\nmetadata:\n  name: legacy\nspec:\n  ports:\n    - port: 8080\n',
  },
  {
    path: 'sandbox/main.tf',
    kind: 'sandbox terraform',
    change: 'modified' as const,
    before:
      'module "sandbox" {\n  source = "./modules/eks"\n  instance_type = "t3.medium"\n  min_size = 1\n  max_size = 3\n}\n',
    after:
      'module "sandbox" {\n  source = "./modules/eks"\n  instance_type = "t3.large"\n  min_size = 2\n  max_size = 6\n}\n',
  },
]

type TPreviewOutcome = {
  category: 'Stack' | 'Sandbox' | 'Components'
  name: string
  status: 'success' | 'error' | 'not-started' | 'in-progress' | 'unknown'
  detail: string
}

type TPreviewRun = {
  id: string
  title: string
  type: TInstallDeploymentRecordType
  resources: string[]
  status: 'in-progress' | 'success' | 'error'
  scope: 'components' | 'stack' | 'mixed'
  activity: string
  note?: string
  outcomes: TPreviewOutcome[]
  changesApplied?: boolean
  created_at: string
  steps: TWorkflowStep[]
}

const previewRuns: TPreviewRun[] = [
  {
    id: 'wf-preview-running',
    title: 'Deploy api + worker',
    type: 'component_deploy',
    resources: ['api', 'worker'],
    status: 'in-progress',
    scope: 'components',
    activity: 'Running api post-deploy readiness check',
    outcomes: [
      {
        category: 'Components',
        name: 'api',
        status: 'in-progress',
        detail: 'Readiness check running',
      },
      {
        category: 'Components',
        name: 'worker',
        status: 'not-started',
        detail: 'Not started',
      },
    ],
    minutesAgo: 2,
    names: [
      'Sync configuration',
      'Plan api',
      'Deploy api',
      'Run api post-deploy readiness check',
      'Deploy worker',
      'Verify install health',
    ],
    states: [
      'success',
      'success',
      'success',
      'in-progress',
      'pending',
      'pending',
    ],
  },
  {
    id: 'wf-preview-components-pending',
    title: 'Roll out template v14',
    type: 'app_branch_update',
    resources: ['stack', 'sandbox', 'api', 'worker', 'legacy'],
    status: 'in-progress',
    scope: 'mixed',
    activity: 'Prepare component deployment',
    outcomes: [
      {
        category: 'Stack',
        name: 'Stack',
        status: 'success',
        detail: 'Update completed',
      },
      {
        category: 'Sandbox',
        name: 'Sandbox',
        status: 'success',
        detail: 'Update completed',
      },
      {
        category: 'Components',
        name: 'api',
        status: 'not-started',
        detail: 'Not started',
      },
      {
        category: 'Components',
        name: 'worker',
        status: 'not-started',
        detail: 'Not started',
      },
      {
        category: 'Components',
        name: 'legacy',
        status: 'not-started',
        detail: 'Removal not started',
      },
    ],
    minutesAgo: 3,
    names: [
      'Sync configuration',
      'Apply stack',
      'Apply sandbox',
      'Validate install inputs',
      'Prepare component deployment',
      'Deploy api',
      'Deploy worker',
      'Remove legacy',
      'Verify install health',
    ],
    states: [
      'success',
      'success',
      'success',
      'success',
      'in-progress',
      'pending',
      'pending',
      'pending',
      'pending',
    ],
  },
  {
    id: 'wf-preview-approval',
    title: 'Roll out template v15',
    type: 'app_branch_update',
    resources: ['stack', 'sandbox', 'api', 'worker', 'legacy'],
    status: 'in-progress',
    scope: 'mixed',
    activity: 'Waiting for component plan approval',
    outcomes: [
      {
        category: 'Stack',
        name: 'Stack',
        status: 'success',
        detail: 'Update completed',
      },
      {
        category: 'Sandbox',
        name: 'Sandbox',
        status: 'success',
        detail: 'Update completed',
      },
      {
        category: 'Components',
        name: 'api',
        status: 'not-started',
        detail: 'Not started — awaiting plan approval',
      },
      {
        category: 'Components',
        name: 'worker',
        status: 'not-started',
        detail: 'Not started — awaiting plan approval',
      },
      {
        category: 'Components',
        name: 'legacy',
        status: 'not-started',
        detail: 'Removal not started — awaiting plan approval',
      },
    ],
    minutesAgo: 4,
    names: [
      'Sync configuration',
      'Apply stack',
      'Apply sandbox',
      'Generate component plan',
      'Component plan approval',
      'Deploy api',
      'Deploy worker',
      'Remove legacy',
      'Verify install health',
    ],
    states: [
      'success',
      'success',
      'success',
      'success',
      'approval-awaiting',
      'pending',
      'pending',
      'pending',
      'pending',
    ],
  },
  {
    id: 'wf-preview-partial',
    title: 'Roll out template v13',
    type: 'app_branch_update',
    resources: ['stack', 'sandbox', 'api', 'worker', 'scheduler'],
    status: 'error',
    scope: 'mixed',
    activity: 'worker: post-deploy readiness check failed',
    changesApplied: true,
    outcomes: [
      {
        category: 'Stack',
        name: 'Stack',
        status: 'success',
        detail: 'Update completed',
      },
      {
        category: 'Sandbox',
        name: 'Sandbox',
        status: 'success',
        detail: 'Update completed',
      },
      {
        category: 'Components',
        name: 'api',
        status: 'success',
        detail: 'Rollout completed',
      },
      {
        category: 'Components',
        name: 'worker',
        status: 'error',
        detail: 'Deployed; readiness check failed',
      },
      {
        category: 'Components',
        name: 'scheduler',
        status: 'not-started',
        detail: 'Not started',
      },
    ],
    minutesAgo: 5,
    names: [
      'Sync configuration',
      'Apply stack',
      'Apply sandbox',
      'Deploy api',
      'Verify api readiness',
      'Deploy worker',
      'Verify worker readiness',
      'Deploy scheduler',
    ],
    states: [
      'success',
      'success',
      'success',
      'success',
      'success',
      'success',
      'error',
      'not-started',
    ],
  },
  {
    id: 'wf-preview-fixed',
    title: 'Update stack permissions',
    type: 'stack_update',
    resources: ['stack'],
    status: 'success',
    scope: 'stack',
    activity: 'Stack update completed',
    note: 'Previous stack update failed',
    outcomes: [
      {
        category: 'Stack',
        name: 'Stack',
        status: 'success',
        detail: 'Update completed',
      },
    ],
    minutesAgo: 35,
    names: ['Sync configuration', 'Validate stack policy', 'Apply stack'],
    states: ['success', 'success', 'success'],
  },
  {
    id: 'wf-preview-failed',
    title: 'Update stack permissions',
    type: 'stack_update',
    resources: ['stack'],
    status: 'error',
    scope: 'stack',
    activity: 'Validate stack policy failed — invalid resource ARN',
    changesApplied: false,
    outcomes: [
      {
        category: 'Stack',
        name: 'Stack',
        status: 'error',
        detail: 'Validation failed before apply; no changes applied',
      },
    ],
    minutesAgo: 50,
    names: ['Sync configuration', 'Validate stack policy', 'Apply stack'],
    states: ['success', 'error', 'not-started'],
  },
  {
    id: 'wf-preview-completed',
    title: 'Roll out template v12',
    type: 'app_branch_update',
    resources: ['stack', 'sandbox', 'api', 'worker', 'legacy'],
    status: 'success',
    scope: 'mixed',
    activity: 'Stack, sandbox and component updates completed',
    outcomes: [
      {
        category: 'Stack',
        name: 'Stack',
        status: 'success',
        detail: 'Update completed',
      },
      {
        category: 'Sandbox',
        name: 'Sandbox',
        status: 'success',
        detail: 'Update completed',
      },
      {
        category: 'Components',
        name: 'api',
        status: 'success',
        detail: 'Rollout completed',
      },
      {
        category: 'Components',
        name: 'worker',
        status: 'success',
        detail: 'Rollout completed',
      },
      {
        category: 'Components',
        name: 'legacy',
        status: 'success',
        detail: 'Removal completed',
      },
    ],
    minutesAgo: 2880,
    names: [
      'Sync configuration',
      'Apply stack',
      'Apply sandbox',
      'Deploy api',
      'Deploy worker',
      'Verify install health',
    ],
    states: ['success', 'success', 'success', 'success', 'success', 'success'],
  },
].map(({ minutesAgo, names, states, ...run }) => ({
  ...run,
  type: run.type as TInstallDeploymentRecordType,
  status: run.status as TPreviewRun['status'],
  scope: run.scope as TPreviewRun['scope'],
  outcomes: run.outcomes as TPreviewOutcome[],
  created_at: DateTime.now().minus({ minutes: minutesAgo }).toISO(),
  steps: names.map((name, index) => ({
    id: `${run.id}-step-${index}`,
    name,
    execution_type: 'system',
    group_idx: index,
    finished: ['success', 'error'].includes(states[index]),
    execution_time: ['success', 'error'].includes(states[index])
      ? (index + 1) * 2_000_000_000
      : undefined,
    status: { status: states[index] },
  })) as TWorkflowStep[],
}))

type TPreviewTab = 'template' | 'workflow' | 'changes'
const previewTabs = {
  template: { path: '/template-updates', text: 'Template updates' },
  workflow: { path: '/workflow', text: 'Workflow' },
  changes: { path: '/', text: 'Change summary' },
}
const tabOrder = (run: TPreviewRun): TPreviewTab[] =>
  run.status === 'success'
    ? ['changes', 'workflow', 'template']
    : ['template', 'workflow', 'changes']
const detailBase = (run: TPreviewRun) =>
  `/org-preview/apps/app-preview/installs/install-preview/deployments/${run.id}`
const detailHref = (run: TPreviewRun, tab: TPreviewTab) =>
  `${detailBase(run)}${tab === 'changes' ? '' : previewTabs[tab].path}`

const previewStatus = (run: TPreviewRun) =>
  run.status === 'in-progress' &&
  run.steps.some((step) => step.status.status === 'approval-awaiting')
    ? 'awaiting-approval'
    : run.status

const PreviewApprovalBanner = () => {
  const { addModal, removeModal } = useSurfaces()
  const [response, setResponse] = useState('')
  const actions = [
    {
      label: 'Retry plan',
      Modal: RetryPlanModal,
      copy: RETRY_MODAL_COPY.helm_approval,
    },
    {
      label: 'Deny plan',
      Modal: DenyPlanModal,
      copy: DENY_MODAL_COPY.helm_approval,
    },
    {
      label: 'Approve plan',
      Modal: ApprovePlanModal,
      copy: APPROVAL_MODAL_COPY.helm_approval,
    },
  ]

  return (
    <Banner className="@container" theme="warn">
      <div className="flex flex-col gap-2">
        <div className="flex flex-col">
          <Text weight="strong">Helm chart plan requires review</Text>
          <Text variant="subtext" theme="neutral">
            This Helm chart plan includes updates to your deployment. Review the
            changes to ensure your release will work as intended.
          </Text>
        </div>
        <div className="flex flex-wrap self-end gap-2">
          {actions.map(({ label, Modal, copy }) => (
            <Button
              key={label}
              variant={label === 'Approve plan' ? 'primary' : undefined}
              onClick={() => {
                const modalId = addModal(
                  <Modal
                    modalCopy={{
                      ...copy,
                      message:
                        'Storybook preview only. No deployment will be changed.',
                    }}
                    isPending={false}
                    onSubmit={() => {
                      setResponse(
                        `${label} simulated. The fixture is unchanged.`
                      )
                      removeModal(modalId)
                    }}
                  />
                )
              }}
            >
              {label}
            </Button>
          ))}
        </div>
        <div role="status">
          <Text variant="subtext" theme="neutral">
            {response || 'Preview only · Approval actions are simulated.'}
          </Text>
        </div>
      </div>
    </Banner>
  )
}

const PreviewTabContent = ({
  run,
  tab,
}: {
  run: TPreviewRun
  tab: TPreviewTab
}) => {
  if (tab === 'workflow')
    return (
      <div className="flex flex-col gap-4">
        <SectionHeader title="Workflow steps" description={run.activity} />
        {previewStatus(run) === 'awaiting-approval' ? (
          <PreviewApprovalBanner />
        ) : null}
        <WorkflowStepsComponent
          workflowSteps={run.steps}
          eagerStepsLoaded
          allStepsLoaded
        />
      </div>
    )
  if (tab === 'template') {
    const stackFile = {
      path: 'stack/permissions.yaml',
      kind: 'stack policy',
      change: 'modified' as const,
      before: 'Action: s3:GetObject\nResource: arn:aws:s3:::acme-artifacts\n',
      after: 'Action: s3:GetObject\nResource: arn:aws:s3:::acme-artifacts/*\n',
    }
    const stackSection: DiffSectionData = {
      name: 'Stack',
      sectionKey: 'stack',
      additions: 0,
      removals: 0,
      changed: 1,
      grouped: false,
      fields: [],
      entities: [],
      files: [{ name: stackFile.path, op: 'change' }],
    }
    const schedulerFile = {
      path: 'components/scheduler/values.yaml',
      kind: 'helm values',
      change: 'modified' as const,
      before: 'replicaCount: 1\n',
      after: 'replicaCount: 2\n',
    }
    const includesScheduler = run.resources.includes('scheduler')
    const componentSection: DiffSectionData = {
      ...sandboxSections[0],
      removals: 0,
      changed: includesScheduler ? 2 : 1,
      entities: [
        ...sandboxSections[0].entities.filter(
          (entity) => entity.name !== 'legacy'
        ),
        ...(includesScheduler
          ? [
              {
                name: 'scheduler',
                op: 'change' as const,
                componentType: 'helm_chart',
                fields: [],
                files: [{ name: schedulerFile.path, op: 'change' as const }],
              },
            ]
          : []),
      ],
    }
    return (
      <AppConfigFilesDiff
        title="Template updates"
        previousVersion="abc123d"
        currentVersion="def456a"
        configSections={
          run.scope === 'stack'
            ? [stackSection]
            : run.scope === 'components'
              ? [componentSection]
              : [
                  stackSection,
                  includesScheduler ? componentSection : sandboxSections[0],
                  sandboxSections[1],
                ]
        }
        files={
          run.scope === 'stack'
            ? [stackFile]
            : run.scope === 'components'
              ? sandboxFiles.slice(0, 2)
              : includesScheduler
                ? [
                    stackFile,
                    ...sandboxFiles.filter(
                      (file) => file.path !== 'components/legacy/service.yaml'
                    ),
                    schedulerFile,
                  ]
                : [stackFile, ...sandboxFiles]
        }
      />
    )
  }
  if (run.status === 'error' && run.changesApplied === false)
    return (
      <EmptyState
        emptyTitle="No changes applied"
        emptyMessage="Stack policy validation failed before the apply step. See Workflow for the failed step."
      />
    )
  if (run.status === 'error')
    return (
      <SectionHeader
        title="Change summary"
        description="Some changes were applied. worker was deployed but failed readiness; scheduler was not started."
      />
    )
  const planned =
    run.scope === 'stack'
      ? [summaries[0]]
      : run.scope === 'components'
        ? [summaries[1]]
        : summaries
  return (
    <div className="flex flex-col gap-4">
      <SectionHeader
        title="Change summary"
        description={
          run.status === 'success'
            ? 'Resource changes from this completed deployment.'
            : previewStatus(run) === 'awaiting-approval'
              ? 'Stack and sandbox updates completed. Component changes are planned and awaiting approval; they have not been applied.'
              : 'Planned changes. This deployment is still running; these are not final results.'
        }
      />
      <WorkflowChangesSummary
        summaries={planned.map((summary) => ({
          ...summary,
          status:
            previewStatus(run) === 'awaiting-approval' &&
            summary.stepId !== 'step-stack'
              ? 'pending-approval'
              : 'approved',
        }))}
        renderDetail={(summary) => (
          <CodeBlock language="diff">
            {summary.stepId === 'step-stack'
              ? '+ aws_iam_role.worker\n+ aws_iam_role_policy.worker\n~ aws_iam_policy.artifacts\n~ aws_iam_role.api\n~ aws_iam_role_policy.api\n'
              : '- Deployment/api replicas: 2\n+ Deployment/api replicas: 3\n- Service/api targetPort: 8080\n+ Service/api targetPort: 8081\n'}
          </CodeBlock>
        )}
      />
    </div>
  )
}

const PreviewPanel = ({ run, ...props }: IPanel & { run: TPreviewRun }) => (
  <Panel
    heading={run.title}
    size="3/4"
    aria-label="Deployment details"
    {...props}
  >
    <div className="flex flex-wrap items-center gap-3">
      <PreviewRunStatus run={run} />
      <ID>{run.id}</ID>
      <Time time={run.created_at} format="relative" variant="subtext" />
    </div>
    <Text
      variant="subtext"
      theme={run.status === 'error' ? 'error' : 'neutral'}
    >
      {run.activity}
    </Text>
    <Tabs
      naturalHeight
      initActiveTab={run.status === 'success' ? 'changes' : 'workflow'}
      tabsClassName="mt-4"
      tabControlsClassName="!gap-2 md:!gap-6 [&>button]:px-1 md:[&>button]:px-3"
      tabLabels={Object.fromEntries(
        Object.entries(previewTabs).map(([key, value]) => [key, value.text])
      )}
      tabs={Object.fromEntries(
        tabOrder(run).map((tab) => [
          tab,
          <div key={tab} className="flex flex-col gap-4">
            <div className="flex justify-end">
              <Button href={detailHref(run, tab)} variant="ghost" size="sm">
                Open full page <Icon variant="ArrowSquareOutIcon" />
              </Button>
            </div>
            <PreviewTabContent run={run} tab={tab} />
          </div>,
        ])
      )}
    />
  </Panel>
)

const emptyFilter: IDeploymentFilter = {
  search: '',
  status: new Set(),
  type: new Set(),
}
const previewScenarios = {
  'running-components': {
    label: 'Running components',
    runIds: ['wf-preview-running'],
  },
  'components-not-started': {
    label: 'Components pending',
    runIds: ['wf-preview-components-pending'],
  },
  'pending-approval': {
    label: 'Pending approval',
    runIds: ['wf-preview-approval'],
  },
  'partial-failure': {
    label: 'Partial failure',
    runIds: ['wf-preview-partial'],
  },
  'stack-recovery': {
    label: 'Stack failure and recovery',
    runIds: ['wf-preview-fixed', 'wf-preview-failed'],
  },
  completed: {
    label: 'Completed rollout',
    runIds: ['wf-preview-completed'],
  },
  'interactive-sandbox': {
    label: 'All deployments',
    runIds: previewRuns.map((run) => run.id),
  },
}

export const DeploymentsPreview = ({
  scenario,
}: {
  scenario: keyof typeof previewScenarios
}) => {
  const { addPanel } = useSurfaces()
  const { pathname } = useLocation()
  const [filter, setFilter] = useState<IDeploymentFilter>(emptyFilter)
  const scenarioRuns = previewRuns.filter((run) =>
    previewScenarios[scenario].runIds.includes(run.id)
  )
  const resourceOptions = [
    ...new Set(scenarioRuns.flatMap((run) => run.resources)),
  ].map((value) => ({ value, label: value }))
  const directLinkRun =
    scenario === 'interactive-sandbox'
      ? scenarioRuns.find((run) => run.id === 'wf-preview-completed')!
      : scenarioRuns[0]
  const selected = scenarioRuns.find(
    (run) =>
      pathname === detailBase(run) || pathname.startsWith(`${detailBase(run)}/`)
  )
  const activeTab: TPreviewTab = pathname.endsWith('/template-updates')
    ? 'template'
    : pathname.endsWith('/workflow')
      ? 'workflow'
      : 'changes'
  const since = datePresetQueryParameter(filter.date ?? null)
  const hasActiveFilters =
    !!filter.search ||
    filter.status.size > 0 ||
    filter.type.size > 0 ||
    !!filter.resource ||
    !!filter.date
  const runs = scenarioRuns.filter(
    (run) =>
      `${run.title} ${run.activity}`
        .toLowerCase()
        .includes(filter.search.toLowerCase()) &&
      (!filter.status.size ||
        [...filter.status].some((status) =>
          WORKFLOW_STATUS_GROUPS[status].includes(run.status)
        )) &&
      (!filter.type.size || filter.type.has(run.type)) &&
      (!filter.resource || run.resources.includes(filter.resource)) &&
      (!since || DateTime.fromISO(run.created_at) >= DateTime.fromISO(since))
  )

  return (
    <div className="mx-auto flex w-full max-w-6xl flex-col gap-6 p-4 md:p-6">
      <div className="flex flex-wrap items-center justify-between gap-2 border-b pb-3">
        <div className="flex w-80 max-w-full min-w-0 flex-col gap-2">
          <Text variant="subtext" theme="neutral">
            Interactive preview · Fixture data only
          </Text>
          <Select
            id="preview-scenario"
            labelProps={{ labelText: 'Preview scenario' }}
            value={scenario}
            options={Object.entries(previewScenarios).map(([value, entry]) => ({
              value,
              label: entry.label,
            }))}
            onChange={(value) => {
              const url = new URL(window.location.href)
              url.searchParams.set(
                'story',
                value === 'interactive-sandbox'
                  ? 'playground--installs--deployments--interactive-sandbox'
                  : `features--installs--deployment-details--${value}`
              )
              window.location.assign(url.toString())
            }}
          />
        </div>
        {!selected ? (
          <Button
            variant="ghost"
            size="sm"
            href={detailHref(
              directLinkRun,
              directLinkRun.status === 'success' ? 'changes' : 'workflow'
            )}
          >
            Preview an existing full-page link{' '}
            <Icon variant="ArrowSquareOutIcon" />
          </Button>
        ) : null}
      </div>
      {selected ? (
        <>
          <Button className="self-start" variant="ghost" href="/">
            <Icon variant="ArrowLeftIcon" /> Back to deployments
          </Button>
          <main aria-label="Deployment full page">
            <DetailPage
              className="[&>.tab-nav]:gap-2 md:[&>.tab-nav]:gap-6 [&>.tab-nav>a]:px-1 md:[&>.tab-nav>a]:px-3"
              header={
                <DetailHeader
                  backLink={false}
                  title={selected.title}
                  status={<PreviewRunStatus run={selected} />}
                  id={selected.id}
                  description={selected.activity}
                />
              }
              tabNav={{
                basePath: detailBase(selected),
                tabs: tabOrder(selected).map((tab) => previewTabs[tab]),
              }}
            >
              <PreviewTabContent
                key={`${selected.id}-${activeTab}`}
                run={selected}
                tab={activeTab}
              />
            </DetailPage>
          </main>
        </>
      ) : (
        <ListPage
          title="Deployments"
          description="Follow rollouts for acme-preview. Open details to inspect a workflow or its changes."
        >
          <div className="flex min-w-0 flex-wrap items-center gap-2">
            <SearchInput
              aria-label="Search deployments"
              placeholder="Search deployments"
              value={filter.search}
              onChange={(search) =>
                setFilter((current) => ({ ...current, search }))
              }
              onClear={() =>
                setFilter((current) => ({ ...current, search: '' }))
              }
            />
            <CheckboxFilterDropdown
              id="deployments-filter-status"
              label="Status"
              options={workflowStatusOptions().map((value) => ({
                value,
                label: WORKFLOW_STATUS_LABELS[value],
              }))}
              selected={filter.status}
              onChange={(status) =>
                setFilter((current) => ({
                  ...current,
                  status: status as Set<TWorkflowStatusOption>,
                }))
              }
            />
            <CheckboxFilterDropdown
              id="deployments-filter-type"
              label="Type"
              options={(
                Object.keys(
                  DEPLOYMENT_TYPE_LABELS
                ) as TInstallDeploymentRecordType[]
              ).map((value) => ({
                value,
                label: DEPLOYMENT_TYPE_LABELS[value],
              }))}
              selected={filter.type}
              onChange={(type) =>
                setFilter((current) => ({
                  ...current,
                  type: type as Set<TInstallDeploymentRecordType>,
                }))
              }
            />
            <RadioFilterDropdown
              id="deployments-filter-resource"
              label="Resource"
              options={resourceOptions}
              selected={filter.resource}
              onChange={(resource) =>
                setFilter((current) => ({ ...current, resource }))
              }
            />
            <RadioFilterDropdown
              id="deployments-filter-date"
              label="Date"
              options={(
                Object.keys(WORKFLOW_DATE_LABELS) as TWorkflowDatePreset[]
              ).map((value) => ({ value, label: WORKFLOW_DATE_LABELS[value] }))}
              selected={filter.date}
              onChange={(date) =>
                setFilter((current) => ({
                  ...current,
                  date: date as TWorkflowDatePreset | undefined,
                }))
              }
            />
            {hasActiveFilters ? (
              <Button variant="ghost" onClick={() => setFilter(emptyFilter)}>
                Clear filters
              </Button>
            ) : null}
          </div>
          {runs.length ? (
            <Timeline
              className="w-full"
              events={runs}
              groupByDate={false}
              pagination={{ hasNext: false, offset: 0, limit: 20 }}
              getEventKey={(run) => run.id}
              renderEvent={(run) => (
                <DeploymentRow
                  run={run}
                  title={run.title}
                  createdAt={run.created_at}
                  href={`/deployments/${run.id}`}
                  onViewDetails={() => addPanel(<PreviewPanel run={run} />)}
                >
                  {run.note ? (
                    <Text variant="subtext" theme="neutral">
                      {run.note}
                    </Text>
                  ) : null}
                </DeploymentRow>
              )}
            />
          ) : (
            <EmptyState
              emptyTitle="No deployments found"
              emptyMessage="Try a different search or status."
            />
          )}
        </ListPage>
      )}
    </div>
  )
}
