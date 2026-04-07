export default {
  title: 'Installs/InstallStatuses',
}

import { InstallStatuses, SimpleInstallStatuses } from './InstallStatuses'

const mockInstall = {
  id: 'inst-123',
  org_id: 'org-456',
  runner_id: 'runner-789',
  runner_type: 'aws',
  runner_status: 'active',
  sandbox_status: 'active',
  composite_component_status: 'active',
  drifted_objects: [],
  install_components: [],
  install_sandbox_runs: [],
} as any

export const BadgeVariant = () => (
  <InstallStatuses install={mockInstall} variant="badge" />
)

export const IconVariant = () => (
  <InstallStatuses install={mockInstall} variant="icon" />
)

export const LabelHidden = () => (
  <InstallStatuses install={mockInstall} isLabelHidden />
)

export const Simple = () => (
  <SimpleInstallStatuses install={mockInstall} />
)

export const SimpleLabelHidden = () => (
  <SimpleInstallStatuses install={mockInstall} isLabelHidden />
)
