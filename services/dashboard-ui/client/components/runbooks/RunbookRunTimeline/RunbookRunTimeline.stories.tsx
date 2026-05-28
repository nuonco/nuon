export default {
  title: 'Runbooks/RunbookRunTimeline',
}

import { RunbookRunTimeline } from './RunbookRunTimeline'

const statuses = ['completed', 'in-progress', 'completed', 'error', 'queued']

const mockRuns = Array.from({ length: 5 }, (_, i) => ({
  id: `run-${i + 1}`,
  created_at: new Date(Date.now() - i * 86400000).toISOString(),
  status: statuses[i],
  install_workflow_id: `wf-${i + 1}`,
  install_workflow: {
    id: `wf-${i + 1}`,
    status: { status: statuses[i] },
  },
  created_by: { email: 'user@example.com' },
})) as any

export const Default = () => (
  <RunbookRunTimeline
    runbookName="restart-and-check"
    runs={mockRuns}
    basePath="/org-1/installs/install-1"
  />
)

export const Empty = () => (
  <RunbookRunTimeline
    runbookName="restart-and-check"
    runs={[]}
    basePath="/org-1/installs/install-1"
  />
)
