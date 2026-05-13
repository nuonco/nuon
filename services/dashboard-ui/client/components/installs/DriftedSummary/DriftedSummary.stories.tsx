export default {
  title: 'Installs/DriftedSummary',
}

import { DriftedSummary } from './DriftedSummary'
import type { TDriftedObject } from '@/types/ctl-api.types'

const mockOrgId = 'orgX'
const mockInstallId = 'inliY'

const componentDrift = (
  id: string,
  name: string,
  workflowId: string
): TDriftedObject => ({
  target_id: id,
  target_type: 'install_deploy',
  component_name: name,
  install_workflow_id: workflowId,
})

const sandboxDrift = (id: string, workflowId: string): TDriftedObject => ({
  target_id: id,
  target_type: 'install_sandbox_run',
  install_workflow_id: workflowId,
})

export const SingleComponent = () => (
  <DriftedSummary
    orgId={mockOrgId}
    installId={mockInstallId}
    driftedObjects={[componentDrift('1', 'kubelogstream', 'wf1')]}
  />
)

export const TwoComponents = () => (
  <DriftedSummary
    orgId={mockOrgId}
    installId={mockInstallId}
    driftedObjects={[
      componentDrift('1', 'kubelogstream', 'wf1'),
      componentDrift('2', 'observability', 'wf2'),
    ]}
  />
)

export const ManyComponents = () => (
  <DriftedSummary
    orgId={mockOrgId}
    installId={mockInstallId}
    driftedObjects={[
      componentDrift('1', 'kubelogstream', 'wf1'),
      componentDrift('2', 'observability', 'wf2'),
      componentDrift('3', 'certificate', 'wf3'),
      componentDrift('4', 'coder', 'wf4'),
      componentDrift('5', 'application_load_balancer', 'wf5'),
      componentDrift('6', 'rds_cluster_coder', 'wf6'),
      componentDrift('7', 'rds_subnet', 'wf7'),
    ]}
  />
)

export const SandboxOnly = () => (
  <DriftedSummary
    orgId={mockOrgId}
    installId={mockInstallId}
    driftedObjects={[sandboxDrift('1', 'wf1')]}
  />
)

export const Empty = () => (
  <DriftedSummary
    orgId={mockOrgId}
    installId={mockInstallId}
    driftedObjects={[]}
  />
)
