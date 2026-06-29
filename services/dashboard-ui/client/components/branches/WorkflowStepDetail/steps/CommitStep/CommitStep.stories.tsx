export default {
  title: 'Branches/WorkflowStepDetail/CommitStep',
}

import { CommitStep } from './CommitStep'

const baseMetadata = {
  commit_sha: 'a1b2c3d4e5f6a7b8c9d0',
  commit_message: 'feat: add deployment plan editor\n\nIntroduces install groups with manual and label-based selection.',
  author_name: 'Ada Lovelace',
  branch: 'feature/deploy-plans',
  base_branch: 'main',
}

export const WithPullRequest = () => (
  <CommitStep
    metadata={{
      ...baseMetadata,
      pr_number: 482,
      pr_title: 'Add deployment plan editor',
      pr_status: 'open',
      pr_reviewer_count: 2,
      pr_url: 'https://github.com/example/repo/pull/482',
      files_changed: 6,
      additions: 214,
      deletions: 37,
      changed_files: [
        { path: 'client/components/branches/DeploymentPlanEditor.tsx', additions: 180, deletions: 12 },
        { path: 'client/lib/ctl-api/apps/branches/index.ts', additions: 34, deletions: 25 },
      ],
    }}
  />
)

export const NoPullRequest = () => (
  <CommitStep
    metadata={{
      ...baseMetadata,
      files_changed: 1,
      additions: 8,
      deletions: 0,
      changed_files: [{ path: 'nuon.toml', additions: 8, deletions: 0 }],
    }}
  />
)

export const MessageOnly = () => <CommitStep metadata={baseMetadata} />

export const Fetching = () => <CommitStep metadata={{}} />
