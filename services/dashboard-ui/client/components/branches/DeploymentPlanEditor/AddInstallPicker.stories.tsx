export default {
  title: 'Branches/DeploymentPlanEditor/AddInstallPicker',
}

import { AddInstallPicker } from './AddInstallPicker'

const noop = () => {}

const installs = [
  { id: 'i1', name: 'acme-prod' },
  { id: 'i2', name: 'globex-prod' },
  { id: 'i3', name: 'initech-staging' },
  { id: 'i4', name: 'umbrella-dev' },
  { id: 'i5', name: 'soylent-prod' },
  { id: 'i6', name: 'hooli-staging' },
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
