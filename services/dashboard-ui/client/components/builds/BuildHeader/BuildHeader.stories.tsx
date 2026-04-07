export default {
  title: 'Builds/BuildHeader',
}

import { BuildHeader } from './BuildHeader'

const mockBuild = {
  id: 'bld-abc123',
  component_id: 'comp-1',
  created_at: '2024-01-15T10:00:00Z',
  updated_at: '2024-01-15T10:05:00Z',
  status_v2: { status: 'active', status_human_description: 'Build is running' },
  vcs_connection_commit: null,
  runner_job: null,
  component_config_connection: { id: 'config-conn-1' },
} as any

const mockComponent = {
  id: 'comp-1',
  name: 'api-service',
  type: 'helm_chart',
  app_id: 'app-1',
} as any

const mockApp = {
  id: 'app-1',
  name: 'My App',
  org_id: 'org-1',
} as any

export const Default = () => (
  <BuildHeader
    component={mockComponent}
    build={mockBuild}
    app={mockApp}
  />
)
