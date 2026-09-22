export default {
  title: 'Sandbox/InstallSandbox',
}

import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { SandboxConfigCardComponent } from '@/components/sandbox/SandboxConfigCard'
import type { TInstallSandbox, TSandboxConfig, TSandboxRun } from '@/types'
import { InstallSandbox } from './InstallSandbox'

const sandbox = {
  id: 'sandbox-1',
  status: 'active',
  status_v2: { status: 'active' },
  terraform_workspace: { id: 'workspace-1' },
} as TInstallSandbox

const latestRun = {
  id: 'run-1',
  run_type: 'provision',
  created_at: '2026-09-22T14:00:00Z',
} as TSandboxRun

const config = {
  type: 'terraform',
  terraform_version: '1.9.0',
  drift_schedule: '0 6 * * *',
  variables: {
    region: 'us-east-1',
    environment: 'production',
  },
} as TSandboxConfig

const actions = (
  <>
    <Button variant="secondary">
      <Icon variant="StackIcon" size={16} />
      Terraform state
    </Button>
    <Button variant="secondary">Sandbox controls</Button>
  </>
)

export const Default = () => (
  <InstallSandbox
    sandbox={sandbox}
    latestRun={latestRun}
    actions={actions}
    config={<SandboxConfigCardComponent config={config} />}
  />
)

export const Drifted = () => (
  <InstallSandbox
    sandbox={sandbox}
    latestRun={latestRun}
    actions={actions}
    driftBanner={<Banner theme="warn">Sandbox drift detected</Banner>}
    config={<SandboxConfigCardComponent config={config} />}
  />
)

export const NotProvisioned = () => (
  <InstallSandbox config={<SandboxConfigCardComponent config={config} />} />
)

export const Loading = () => (
  <InstallSandbox loading config={<SandboxConfigCardComponent loading />} />
)
