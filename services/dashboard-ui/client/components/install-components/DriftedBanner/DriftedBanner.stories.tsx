export default {
  title: 'Install Components/DriftedBanner',
}

import { DriftedBanner } from './DriftedBanner'

const mockDrifted = {
  target_type: 'install_deploy',
  install_workflow_id: 'wf-1',
} as any

export const Deploy = () => (
  <DriftedBanner
    drifted={mockDrifted}
    orgId="org-1"
    installId="install-1"
  />
)

export const Sandbox = () => (
  <DriftedBanner
    drifted={{ ...mockDrifted, target_type: 'install_sandbox_run' }}
    orgId="org-1"
    installId="install-1"
  />
)
