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

const mockInstallMixed = {
  ...mockInstall,
  runner_status: 'failed',
  sandbox_status: 'executing',
  composite_component_status: 'active',
  drifted_objects: [
    {
      target_id: 'deploy-1',
      target_type: 'install_deploy',
      component_name: 'web-server',
      install_workflow_id: 'wf-1',
    },
  ],
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

export const WideContainer = () => (
  <div className="w-[600px] border border-dashed p-2">
    <InstallStatuses install={mockInstall} variant="badge" />
  </div>
)

export const NarrowContainer = () => (
  <div className="w-64 border border-dashed p-2">
    <InstallStatuses install={mockInstall} variant="badge" />
  </div>
)

export const NarrowMixedStatuses = () => (
  <div className="w-64 border border-dashed p-2">
    <InstallStatuses install={mockInstallMixed} variant="badge" />
  </div>
)

export const Simple = () => (
  <SimpleInstallStatuses install={mockInstall} />
)

export const SimpleLabelHidden = () => (
  <SimpleInstallStatuses install={mockInstall} isLabelHidden />
)
