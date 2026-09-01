export default {
  title: 'Branches/WorkflowStepDetail/StepCard',
}

import type { DiffSectionData } from '@/components/approvals/plan-diffs/app-config/AppConfigDiff'
import type { TInstallWorkflowStep } from '@/types'
import { StepCard } from './StepCard'
import { ConfigStep } from './steps/ConfigStep/ConfigStep'
import { CommitStep } from './steps/CommitStep/CommitStep'

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
  <StepCard step={configStep} flush>
    <ConfigStep appConfigId="cfg-123" status="success" sections={configSections} />
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
    flush
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
    flush
  >
    <ConfigStep appConfigId={undefined} status="error" sections={[]} />
  </StepCard>
)

export const ConfigSnapshotLoading = () => (
  <StepCard step={configStep} flush>
    <ConfigStep appConfigId="cfg-123" status="success" sections={[]} isLoading />
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
