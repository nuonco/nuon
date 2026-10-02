export default {
  title: 'Views / Installs / Deployment details',
}

import type { ReactNode } from 'react'
import { AppConfigFilesDiff } from '@/components/branches/ComponentConfigDiff/ComponentConfigDiff'
import { Banner } from '@/components/common/Banner'
import { Text } from '@/components/common/Text'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { DeploymentDetail } from '@/components/installs/DeploymentDetail'
import { WorkflowChangesSummary } from '@/components/workflows/WorkflowChangesSummary'
import { WorkflowStepsComponent } from '@/components/workflows/WorkflowSteps'
import type {
  TInstallDeploymentRecord,
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
      banners={
        <Banner theme="warn">
          <div className="flex flex-col gap-1">
            <Text weight="strong">Terraform plan requires review</Text>
            <Text variant="subtext" theme="neutral">
              Review the proposed infrastructure changes before applying them.
            </Text>
          </div>
        </Banner>
      }
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
