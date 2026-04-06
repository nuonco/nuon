export default {
  title: 'Actions/InstallActionRunTimeline',
}

import { InstallActionRunTimeline } from './InstallActionRunTimeline'

const mockRuns = Array.from({ length: 5 }, (_, i) => ({
  id: `run-${i + 1}`,
  created_at: new Date(Date.now() - i * 3600000).toISOString(),
  status: i === 0 ? 'succeeded' : i === 1 ? 'running' : 'succeeded',
  triggered_by_type: i % 2 === 0 ? 'manual' : 'post-deploy-component',
  run_env_vars: { COMPONENT_NAME: 'web-app', COMPONENT_ID: `comp-${i}` },
  status_v2: { status: i === 0 ? 'succeeded' : 'running' },
  created_by: { email: 'user@example.com' },
})) as any

export const Default = () => (
  <InstallActionRunTimeline
    actionId="action-1"
    actionName="deploy-step"
    runs={mockRuns}
    basePath="/org-1/installs/install-1"
    pagination={{ hasNext: true, offset: 0, limit: 10 }}
  />
)

export const Empty = () => (
  <InstallActionRunTimeline
    actionId="action-1"
    actionName="deploy-step"
    runs={[]}
    basePath="/org-1/installs/install-1"
    pagination={{ hasNext: false, offset: 0, limit: 10 }}
  />
)
