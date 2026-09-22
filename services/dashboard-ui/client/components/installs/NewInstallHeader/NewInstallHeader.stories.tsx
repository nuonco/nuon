export default {
  title: 'Installs/NewInstallHeader',
}

import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { InstallStatuses } from '@/components/installs/InstallStatuses'
import { NewInstallHeader } from './NewInstallHeader'

const orgId = 'org-1'

const mockInstall = {
  id: 'inst-1',
  org_id: orgId,
  app_id: 'app-1',
  app: { id: 'app-1', name: 'payments' },
  app_branch: { id: 'brnch-1', name: 'main' },
  name: 'acme-production',
  updated_at: '2026-09-20T18:04:00Z',
  labels: { env: 'production', team: 'platform' },
  metadata: { managed_by: 'nuon/cli/install-config' },
  runner_id: 'runner-1',
  runner_type: 'aws',
  runner_status: 'active',
  sandbox_status: 'active',
  composite_component_status: 'active',
  composite_health_status: 'active',
  drifted_objects: [],
  install_components: [],
  install_sandbox_runs: [],
} as any

const branchAction = (
  <Button variant="secondary" size="sm">
    <Icon variant="PencilSimpleLineIcon" size={16} />
    Edit
  </Button>
)

const settingsAction = (
  <Button variant="secondary" aria-label="Install settings">
    <Icon variant="GearIcon" size={16} />
  </Button>
)

export const Default = () => (
  <NewInstallHeader
    install={mockInstall}
    orgId={orgId}
    branchAction={branchAction}
    settingsAction={settingsAction}
    statuses={<InstallStatuses install={mockInstall} />}
  />
)

export const ManagedByDashboard = () => (
  <NewInstallHeader
    install={{ ...mockInstall, metadata: { managed_by: 'nuon/dashboard' } }}
    orgId={orgId}
    branchAction={branchAction}
    settingsAction={settingsAction}
    statuses={<InstallStatuses install={mockInstall} />}
  />
)

export const WithDriftAndNoLabels = () => {
  const install = {
    ...mockInstall,
    labels: {},
    composite_component_status: 'error',
    composite_health_status: 'degraded',
    drifted_objects: [
      {
        target_id: 'dply-1',
        target_type: 'install_deploy',
        component_name: 'api',
        install_workflow_id: 'wkfl-1',
      },
    ],
  }

  return (
    <NewInstallHeader
      install={install}
      orgId={orgId}
      branchAction={branchAction}
      settingsAction={settingsAction}
      statuses={<InstallStatuses install={install} />}
    />
  )
}
