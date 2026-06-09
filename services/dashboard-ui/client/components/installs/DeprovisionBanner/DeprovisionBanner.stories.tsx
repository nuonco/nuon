export default {
  title: 'Installs/DeprovisionBanner',
}

import { DeprovisionBanner } from './DeprovisionBanner'

const mockProvisioning = {
  id: 'inst123',
  name: 'test-install',
  lifecycle_phase: {
    phase: 'provisioning',
    description: 'Setting up runner and sandbox resources',
  },
} as any

const mockDeprovisioning = {
  id: 'inst123',
  name: 'test-install',
  lifecycle_phase: {
    phase: 'deprovisioning',
    description: 'Tearing down components and cloud resources',
  },
} as any

const mockDeprovisioned = {
  id: 'inst123',
  name: 'test-install',
  lifecycle_phase: {
    phase: 'deprovisioned',
    description: 'Deprovision workflow completed',
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
