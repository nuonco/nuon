export default {
  title: 'Studio/OperationsStudio',
}

import type { TRunbook } from '@/lib/ctl-api/apps/runbooks/get-runbooks'
import { RunbookNotebook } from './RunbookNotebook'

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

const mockActions = [
  { id: 'act-1', name: 'restart-api' },
  { id: 'act-2', name: 'flush-cache' },
  { id: 'act-3', name: 'verify-health' },
]

const mockComponents = [
  { id: 'cmp-1', name: 'ctl-api' },
  { id: 'cmp-2', name: 'dashboard' },
  { id: 'cmp-3', name: 'workers' },
]

const mockRunbooks = [
  {
    id: 'rbk-1',
    name: 'Rotate credentials',
    configs: [
      {
        steps: [
          {
            idx: 0,
            name: 'Deploy API',
            type: 'component_deploy',
            component_name: 'ctl-api',
          },
          {
            idx: 1,
            name: 'Verify health',
            type: 'action',
            action_workflow_id: 'act-3',
          },
        ],
      },
    ],
  },
  { id: 'rbk-2', name: 'Scale up cluster', configs: [{ steps: [] }] },
] as unknown as TRunbook[]

export const Default = () => (
  <RunbookNotebook
    appId="app-runbook-proto"
    components={mockComponents}
    actions={mockActions}
    runbooks={mockRunbooks}
    installs={mockInstalls}
    previewInstallId="inst-1"
    previewInstallState={mockState}
    onPreviewInstallChange={() => {}}
  />
)

export const LoadingError = () => (
  <RunbookNotebook
    appId=""
    loadingError
    components={[]}
    actions={[]}
    runbooks={[]}
  />
)
