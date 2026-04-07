import { OperationRolesList } from './OperationRolesList'

export default { title: 'Common/OperationRolesList' }

export const WithRoles = () => (
  <OperationRolesList
    operationRoles={{
      provision: 'arn:aws:iam::123456789012:role/ProvisionRole',
      deprovision: 'arn:aws:iam::123456789012:role/DeprovisionRole',
      deploy: 'arn:aws:iam::123456789012:role/DeployRole',
    }}
  />
)

export const AllOperations = () => (
  <OperationRolesList
    operationRoles={{
      provision: 'arn:aws:iam::123456789012:role/ProvisionRole',
      deprovision: 'arn:aws:iam::123456789012:role/DeprovisionRole',
      deploy: 'arn:aws:iam::123456789012:role/DeployRole',
      teardown: 'arn:aws:iam::123456789012:role/TeardownRole',
      reprovision: 'arn:aws:iam::123456789012:role/ReprovisionRole',
      trigger: 'arn:aws:iam::123456789012:role/TriggerRole',
    }}
  />
)

export const Empty = () => <OperationRolesList operationRoles={null} />

export const CustomEmptyMessage = () => (
  <OperationRolesList
    operationRoles={{}}
    emptyMessage="No roles have been assigned yet"
  />
)
