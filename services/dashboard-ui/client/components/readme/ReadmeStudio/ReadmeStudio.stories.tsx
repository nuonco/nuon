export default {
  title: 'Readme/ReadmeStudio',
}

import { ReadmeStudio } from './ReadmeStudio'

const mockState = {
  app: { name: 'control-plane' },
  install: { name: 'prod-cluster', status: 'active' },
  sandbox: {
    status: 'finished',
    type: 'aws-eks',
    outputs: { cluster: { version: '1.31' } },
  },
  domain: {
    public_domain: 'prod.example.com',
    internal_domain: 'internal.example.com',
  },
  runner: { status: 'active' },
  components: [
    { name: 'ctl-api', status: 'active', updated_at: '2026-07-20T09:00:00Z' },
    { name: 'dashboard', status: 'active', updated_at: '2026-07-19T15:00:00Z' },
    { name: 'workers', status: 'error', updated_at: '2026-07-18T12:00:00Z' },
  ],
}

const mockInstalls = [
  { id: 'inst-1', name: 'prod-cluster' },
  { id: 'inst-2', name: 'staging-cluster' },
]

const mockRunbooks = [
  { id: 'rbk-1', name: 'Rotate credentials' },
  { id: 'rbk-2', name: 'Scale up cluster' },
  { id: 'rbk-3', name: 'Restore from backup' },
]

const mockActions = [
  { id: 'act-1', name: 'restart-api' },
  { id: 'act-2', name: 'flush-cache' },
]

const mockComponents = [
  { id: 'cmp-1', name: 'ctl-api' },
  { id: 'cmp-2', name: 'dashboard' },
  { id: 'cmp-3', name: 'workers' },
]

export const Default = () => (
  <ReadmeStudio
    installs={mockInstalls}
    runbooks={mockRunbooks}
    actions={mockActions}
    components={mockComponents}
    previewInstallId="inst-1"
    previewInstallState={mockState}
  />
)

export const NoInstallSelected = () => (
  <ReadmeStudio
    installs={mockInstalls}
    runbooks={mockRunbooks}
    actions={mockActions}
    components={mockComponents}
  />
)

export const NoEntities = () => <ReadmeStudio installs={mockInstalls} />

export const LoadError = () => (
  <ReadmeStudio installs={mockInstalls} loadingError />
)
