import { AppBranchRunCard } from './AppBranchRunCard'

export default {
  title: 'Branches/AppBranchRunCard',
}

const run = {
  id: 'arn123',
  workflow_id: 'inw123',
  status: 'active',
  created_at: '2026-09-16T20:00:00Z',
  app_branch: { id: 'abn123', name: 'main' },
  vcs_connection_commit: {
    sha: '5aeca0565df689db877c5aad7c0535e81264801d',
    message: 'chore: update branch configuration',
    author_name: 'Example Developer',
    created_at: '2026-09-16T19:55:00Z',
  },
}

export const DifferentRepository = () => (
  <div className="max-w-3xl">
    <AppBranchRunCard
      appId="app123"
      orgId="org123"
      buildStatus="active"
      sourceRepo="acme/sandbox"
      sourceHref="#"
      sourceCommit={{
        sha: '6987a43568abc8222c96d100284e50c045964258',
        message: 'feat: update sandbox networking',
        author_name: 'Example Developer',
        created_at: '2026-09-16T19:58:00Z',
      }}
      run={run}
    />
  </div>
)

export const SameRepository = () => (
  <div className="max-w-3xl">
    <AppBranchRunCard
      appId="app123"
      orgId="org123"
      buildStatus="active"
      sourceRepo="acme/app"
      sourceHref="#"
      sourceCommit={run.vcs_connection_commit}
      run={run}
    />
  </div>
)
