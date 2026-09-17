export default {
  title: 'Branches/DeploymentPlanEditor/AddInstallPicker',
}

import { AddInstallPicker } from './AddInstallPicker'

const noop = () => {}

const installs = [
  { id: 'instk1a9x2m4b7c0d3e6f9g2h5j', name: 'acme-prod' },
  { id: 'inst8n1p4q7r0s3t6u9v2w5x8y', name: 'payments-prod' },
  { id: 'instz2b5c8d1e4f7g0h3j6k9m2', name: 'payments-staging' },
  { id: 'instq4r7s0t3u6v9w2x5y8z1a4', name: 'analytics-dev' },
  { id: 'instm6n9p2q5r8s1t4u7v0w3x6', name: 'search-prod' },
  { id: 'instc8d1e4f7g0h3j6k9m2n5p8', name: 'search-staging' },
] as any

export const WithInstalls = () => (
  <AddInstallPicker groupId="group-1" unassignedInstalls={installs.slice(0, 3)} onAdd={noop} />
)

export const Searchable = () => (
  <AddInstallPicker groupId="group-1" unassignedInstalls={installs} onAdd={noop} />
)

export const AllAssigned = () => (
  <AddInstallPicker groupId="group-1" unassignedInstalls={[]} onAdd={noop} />
)
