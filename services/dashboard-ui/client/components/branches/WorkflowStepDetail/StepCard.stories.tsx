export default {
  title: 'Branches/WorkflowStepDetail/StepCard',
}

import type { ReactNode } from 'react'
import type { DiffSectionData } from '@/components/approvals/plan-diffs/app-config/AppConfigDiff'
import type { TInstallWorkflowStep } from '@/types'
import { AppContext } from '@/providers/app-provider'
import { StepCard } from './StepCard'
import { ConfigStep } from './steps/ConfigStep/ConfigStep'
import { CommitStep } from './steps/CommitStep/CommitStep'
import { BuildStep } from './steps/BuildStep/BuildStep'
import { DeployGroupStep } from './steps/DeployGroupStep/DeployGroupStep'
import { PlanGroupStep } from './steps/PlanGroupStep/PlanGroupStep'
import { PostDeployRunbooksStep } from './steps/PostDeployRunbooksStep/PostDeployRunbooksStep'

const sandboxToml = `terraform_version = '1.13.5'
skip_noops = true

[public_repo]
repo = 'acme/aws-eks-auto-sandbox'
directory = '.'
branch = 'main'

[vars]
cluster_name = 'n-{{.nuon.install.id}}'
cluster_version = '{{ .nuon.inputs.inputs.cluster_version }}'
enable_dns = 'true'`

const sandboxTomlWithVarFile = `${sandboxToml}

[[var_file]]
contents = 'ebs_storage_class = {\\n  enabled = true\\n  name = "ebs-auto"\\n  is_default_class = true\\n}'`

const runnerToml = `type = 'aws-eks'
region = 'us-west-2'

[env_vars]
LOG_LEVEL = 'info'`

const configSections: DiffSectionData[] = [
  {
    name: 'Sandbox',
    sectionKey: 'sandbox',
    grouped: false,
    additions: 1,
    removals: 0,
    changed: 0,
    entities: [],
    fields: [],
    files: [],
    content: { op: 'add', after: sandboxTomlWithVarFile },
  },
  {
    name: 'Runner',
    sectionKey: 'runner',
    grouped: false,
    additions: 1,
    removals: 0,
    changed: 0,
    entities: [],
    fields: [],
    files: [],
    content: { op: 'add', after: runnerToml },
  },
  {
    name: 'Components',
    sectionKey: 'components',
    grouped: true,
    additions: 2,
    removals: 0,
    changed: 0,
    entities: [
      {
        name: 'api',
        op: 'add',
        componentType: 'helm_chart',
        fields: [
          { key: 'chart_name', op: 'add', diff: "'api'" },
          { key: 'image_tag', op: 'add', diff: "'v1.4.2'" },
        ],
      },
      {
        name: 'worker',
        op: 'add',
        componentType: 'docker_build',
        fields: [{ key: 'dockerfile', op: 'add', diff: "'Dockerfile'" }],
      },
    ],
    fields: [],
  },
]

const configStep = {
  id: 'step-config',
  name: 'fetch app config',
  group_idx: 2,
  started_at: '2024-06-15T10:30:00Z',
  execution_time: 11400000000,
  install_workflow_id: 'wf-xyz789',
  status: {
    status: 'success',
    status_human_description: 'Fetched app config from the config repo',
  },
} as TInstallWorkflowStep

export const ConfigSnapshot = () => (
  <StepCard step={configStep}>
    <ConfigStep
      appConfigId="cfg-123"
      status="success"
      sections={configSections}
    />
  </StepCard>
)

export const ConfigSnapshotInProgress = () => (
  <StepCard
    step={
      {
        ...configStep,
        execution_time: undefined,
        status: {
          status: 'in-progress',
          status_human_description: 'Cloning the config repo',
        },
      } as TInstallWorkflowStep
    }
  >
    <ConfigStep appConfigId={undefined} status="in-progress" sections={[]} />
  </StepCard>
)

export const ConfigSnapshotFailed = () => (
  <StepCard
    step={
      {
        ...configStep,
        status: {
          status: 'error',
          status_human_description:
            'Repository acme/config not found or access denied',
        },
      } as TInstallWorkflowStep
    }
  >
    <ConfigStep appConfigId={undefined} status="error" sections={[]} />
  </StepCard>
)

export const ConfigSnapshotLoading = () => (
  <StepCard step={configStep}>
    <ConfigStep
      appConfigId="cfg-123"
      status="success"
      sections={[]}
      isLoading
    />
  </StepCard>
)

export const PaddedBody = () => (
  <StepCard
    step={
      {
        id: 'step-commit',
        name: 'fetch commit',
        group_idx: 1,
        started_at: '2024-06-15T10:29:00Z',
        execution_time: 3000000000,
        status: {
          status: 'success',
          status_human_description: 'Fetched commit from GitHub',
        },
      } as TInstallWorkflowStep
    }
  >
    <CommitStep
      metadata={{
        commit_sha: 'a1b2c3d4e5f6a7b8',
        commit_message: 'feat: app branches and install configs',
        author_name: 'Ada Lovelace',
        branch: 'feature/app-branches',
        repo: 'acme/config',
        files_changed: 2,
        additions: 120,
        deletions: 8,
        changed_files: [
          { path: 'nuon.toml', additions: 12, deletions: 3 },
          { path: 'components/api/nuon.toml', additions: 108, deletions: 5 },
        ],
      }}
    />
  </StepCard>
)

export const NoBody = () => (
  <StepCard
    step={
      {
        id: 'step-unknown',
        name: 'reconcile install state',
        group_idx: 9,
        status: {
          status: 'pending',
          status_human_description: 'Waiting on the previous step',
        },
      } as TInstallWorkflowStep
    }
  />
)

const mkStep = (over: Record<string, any>) =>
  ({
    started_at: '2024-06-15T10:31:00Z',
    execution_time: 25200000000,
    install_workflow_id: 'wf-xyz789',
    status: { status: 'success' },
    ...over,
  }) as TInstallWorkflowStep

const WithApp = ({ children }: { children: ReactNode }) => (
  <AppContext.Provider
    value={{
      app: { id: 'app-1', name: 'demo-app' } as any,
      labelColors: {},
      refresh: () => {},
    }}
  >
    {children}
  </AppContext.Provider>
)

const mkInstall = (over: Record<string, any> = {}): any => ({
  id: 'ins_acme',
  name: 'acme-prod',
  cloud_platform: 'aws',
  aws_account: { region: 'us-east-1' },
  ...over,
})

export const Builds = () => (
  <WithApp>
    <StepCard
      step={mkStep({
        id: 'step-build',
        name: 'building components and sandbox',
        group_idx: 4,
      })}
    >
      <BuildStep
        status="success"
        metadata={{
          builds: [
            {
              component_id: 'c1',
              component_name: 'application_load_balancer',
              component_type: 'helm_chart',
              status: 'success',
              change_reason: 'source_changed',
              duration: 12.4,
            },
            {
              component_id: 'c2',
              component_name: 'certificate',
              component_type: 'terraform_module',
              status: 'success',
              change_reason: 'source_changed',
              duration: 8.1,
            },
            {
              component_id: 'c3',
              component_name: 'observability',
              component_type: 'helm_chart',
              status: 'success',
              change_reason: 'no_changes',
              duration: 0.2,
            },
          ],
        }}
      />
    </StepCard>
  </WithApp>
)

export const DeployGroup = () => (
  <StepCard
    step={mkStep({
      id: 'step-deploy',
      name: 'deploy install group: beta',
      group_idx: 6,
      status: { status: 'in-progress' },
    })}
  >
    <DeployGroupStep
      groupName="beta"
      totalInstalls={3}
      deployedCount={1}
      rows={[
        {
          installId: 'ins_acme',
          install: mkInstall(),
          deployStatus: 'success',
          installHref: '/org_1/installs/ins_acme',
          workflowHref: '/org_1/installs/ins_acme/workflows/wf_1',
        },
        {
          installId: 'ins_globex',
          install: mkInstall({
            id: 'ins_globex',
            name: 'globex-staging',
            cloud_platform: 'azure',
            aws_account: undefined,
            azure_account: { location: 'eastus' },
          }),
          deployStatus: 'in-progress',
          installHref: '/org_1/installs/ins_globex',
        },
        {
          installId: 'ins_initech',
          install: mkInstall({
            id: 'ins_initech',
            name: 'initech-prod',
            cloud_platform: 'gcp',
            aws_account: undefined,
            gcp_account: { region: 'us-central1' },
          }),
          deployStatus: 'pending',
          installHref: '/org_1/installs/ins_initech',
        },
      ]}
    />
  </StepCard>
)

export const PlanGroup = () => (
  <StepCard
    step={mkStep({
      id: 'step-plan',
      name: 'plan install group: beta',
      group_idx: 5,
      status: {
        status: 'awaiting-approval',
        status_human_description: 'Waiting on plan approval',
      },
    })}
  >
    <PlanGroupStep
      groupName="beta"
      hasResponse={false}
      showApproveBar
      isInProgress={false}
      installs={[
        {
          installId: 'ins_acme',
          installName: 'acme-prod',
          installLabels: { env: 'prod' },
          sections: configSections,
          summary: { added: 1, removed: 0, changed: 2 },
        },
        {
          installId: 'ins_globex',
          installName: 'globex-staging',
          sections: [],
          summary: { added: 0, removed: 0, changed: 0 },
        },
      ]}
    />
  </StepCard>
)

export const PostDeployRunbooks = () => (
  <StepCard
    step={mkStep({
      id: 'step-runbooks',
      name: 'post-deploy runbooks: beta',
      group_idx: 7,
    })}
  >
    <PostDeployRunbooksStep
      groupName="beta"
      runbookNames={['smoke-test', 'notify']}
      rows={[
        {
          installId: 'ins_acme',
          install: mkInstall(),
          installHref: '/org_1/installs/ins_acme',
          runbooks: [
            { runbookName: 'smoke-test', status: 'success' },
            { runbookName: 'notify', status: 'success' },
          ],
        },
        {
          installId: 'ins_globex',
          install: mkInstall({ id: 'ins_globex', name: 'globex-staging' }),
          installHref: '/org_1/installs/ins_globex',
          runbooks: [
            { runbookName: 'smoke-test', status: 'success' },
            { runbookName: 'notify', status: 'in-progress' },
          ],
        },
      ]}
    />
  </StepCard>
)
