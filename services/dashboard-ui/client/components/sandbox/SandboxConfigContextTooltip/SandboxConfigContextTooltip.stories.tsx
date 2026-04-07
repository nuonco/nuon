export default {
  title: 'Sandbox/SandboxConfigContextTooltip',
}

import { SandboxConfigContextTooltip } from './SandboxConfigContextTooltip'

const mockConfig = {
  terraform_version: '1.5.0',
  cloud_platform: 'aws',
} as any

const noop = () => ''

export const Default = () => (
  <SandboxConfigContextTooltip
    appConfigId="config-123"
    orgId="org-456"
    appId="app-789"
    config={mockConfig}
    isLoading={false}
    error={null}
    addModal={noop}
  />
)

export const Loading = () => (
  <SandboxConfigContextTooltip
    appConfigId="config-123"
    orgId="org-456"
    appId="app-789"
    config={undefined}
    isLoading={true}
    error={null}
    addModal={noop}
  />
)
