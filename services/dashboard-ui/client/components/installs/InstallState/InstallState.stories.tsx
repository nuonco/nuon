export default {
  title: 'Installs/InstallState',
}

import { InstallState } from './InstallState'

const mockState = `{
  "install": {
    "id": "inst-1",
    "name": "acme-production"
  },
  "sandbox": {
    "status": "active",
    "outputs": {
      "vpc_id": "vpc-00000000000000000"
    }
  },
  "components": {
    "api": {
      "status": "active"
    }
  }
}`

export const Default = () => (
  <InstallState
    value={mockState}
    filename="acme-production-state.json"
    onDownload={() => undefined}
  />
)

export const Loading = () => <InstallState loading />

export const WithError = () => (
  <InstallState error="Unable to load install state." />
)

export const Empty = () => <InstallState />
