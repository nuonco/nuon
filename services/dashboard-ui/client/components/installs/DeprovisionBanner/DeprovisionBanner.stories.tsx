export default {
  title: 'Installs/DeprovisionBanner',
}

import { DeprovisionBanner } from './DeprovisionBanner'

const mockProvisioning = {
  id: 'inst123',
  name: 'test-install',
  lifecycle_status: {
    status: 'provisioning',
    status_human_description: 'Install is being provisioned',
  },
} as any

const mockDeprovisioning = {
  id: 'inst123',
  name: 'test-install',
  lifecycle_status: {
    status: 'deprovisioning',
    status_human_description: 'Install is being deprovisioned',
  },
} as any

const mockDeprovisioned = {
  id: 'inst123',
  name: 'test-install',
  lifecycle_status: {
    status: 'deprovisioned',
    status_human_description: 'Install has been deprovisioned',
  },
} as any

export const Provisioning = () => (
  <DeprovisionBanner
    install={mockProvisioning}
    orgId="org123"
    workflowId="wf123"
  />
)

export const Deprovisioning = () => (
  <DeprovisionBanner
    install={mockDeprovisioning}
    orgId="org123"
    workflowId="wf123"
  />
)

export const Deprovisioned = () => (
  <DeprovisionBanner install={mockDeprovisioned} orgId="org123" />
)
