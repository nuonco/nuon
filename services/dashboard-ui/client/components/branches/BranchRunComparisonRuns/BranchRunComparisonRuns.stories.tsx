import { BranchRunComparisonRuns } from './BranchRunComparisonRuns'
import type { TBranchRunComparisonRunSummary } from '@/lib/ctl-api/apps/branches/get-branch-run-comparison'

export default {
  title: 'Branches/BranchRunComparisonRuns',
}

const baseRun: TBranchRunComparisonRunSummary = {
  id: 'arn7u0fajsl41xojhuar5yg6nh',
  workflow_id: 'inwabc123workflow00000001',
  status: 'success',
  created_at: '2026-08-27T03:26:48.704Z',
  pr_number: 32,
  base_branch: 'main',
  vcs_connection_commit: {
    sha: '4fef897276aefa90b17737c5ee8fc515b252281e',
    message: 'chore: baseline',
    author_name: 'Example Author',
  },
}

const headRun: TBranchRunComparisonRunSummary = {
  id: 'arnqjlnpio97zdra35v6pn5zvv',
  workflow_id: 'inwb1m469fdl66986yvvj0xilk',
  status: 'success',
  created_at: '2026-08-27T03:33:40.801Z',
  pr_number: 32,
  base_branch: 'main',
  vcs_connection_commit: {
    sha: '89fa6893290f17de337eba1ee672d7314056f813',
    message: 'chore: update domains',
    author_name: 'Example Author',
  },
}

export const WithBaseline = () => (
  <BranchRunComparisonRuns
    orgId="orgdh5k8rpmuoboqctgh6akl4i"
    appId="apph4vzqj0xv2ppb2hpp9huqxg"
    branchId="abr2y1a7izahw066cjpwhc7xol"
    baseRun={baseRun}
    headRun={headRun}
    repoSlug="nuonco/kitchen-sink"
    currentGithubHref="https://github.com/nuonco/kitchen-sink/pull/32"
  />
)

export const FirstRunNoBaseline = () => (
  <BranchRunComparisonRuns
    orgId="orgdh5k8rpmuoboqctgh6akl4i"
    appId="apph4vzqj0xv2ppb2hpp9huqxg"
    branchId="abr2y1a7izahw066cjpwhc7xol"
    headRun={headRun}
    repoSlug="nuonco/kitchen-sink"
  />
)
