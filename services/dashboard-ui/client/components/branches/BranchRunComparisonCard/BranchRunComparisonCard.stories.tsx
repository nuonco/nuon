import { BranchRunComparisonCard } from './BranchRunComparisonCard'
import type { TBranchRunComparisonRunSummary } from '@/lib/ctl-api/apps/branches/get-branch-run-comparison'

export default {
  title: 'Branches/BranchRunComparisonCard',
}

const sampleRun: TBranchRunComparisonRunSummary = {
  id: 'arn7u0fajsl41xojhuar5yg6nh',
  workflow_id: 'inwabc123workflow00000001',
  status: 'success',
  created_at: '2026-08-27T03:26:48.704Z',
  pr_number: 32,
  base_branch: 'main',
  event_type: 'pull_request',
  vcs_connection_commit: {
    sha: '4fef897276aefa90b17737c5ee8fc515b252281e',
    message: 'chore: update domains',
    author_name: 'Example Author',
    author_avatar_url: 'https://github.com/example.png',
  },
}

export const PreviousRun = () => (
  <div className="max-w-md">
    <BranchRunComparisonCard
      label="Previous run"
      run={sampleRun}
      runHref="/org/apps/app/branches/branch/runs/inwabc123workflow00000001"
      repoSlug="nuonco/kitchen-sink"
    />
  </div>
)

export const ThisRun = () => (
  <div className="max-w-md">
    <BranchRunComparisonCard
      label="Current run"
      run={{
        ...sampleRun,
        id: 'arnqjlnpio97zdra35v6pn5zvv',
        vcs_connection_commit: {
          ...sampleRun.vcs_connection_commit!,
          sha: '89fa6893290f17de337eba1ee672d7314056f813',
        },
      }}
      repoSlug="nuonco/kitchen-sink"
    />
  </div>
)
