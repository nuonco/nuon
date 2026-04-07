import type { Meta, StoryObj } from '@ladle/react'
import { OperationRolesList } from './OperationRolesList'

export default {
  title: 'Common/OperationRolesList',
} satisfies Meta

export const WithRoles: StoryObj = {
  render: () => (
    <OperationRolesList
      operationRoles={{
        provision: 'arn:aws:iam::123456789012:role/ProvisionRole',
        deprovision: 'arn:aws:iam::123456789012:role/DeprovisionRole',
        deploy: 'arn:aws:iam::123456789012:role/DeployRole',
      }}
    />
  ),
}

export const AllOperations: StoryObj = {
  render: () => (
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
  ),
}

export const Empty: StoryObj = {
  render: () => <OperationRolesList operationRoles={null} />,
}

export const CustomEmptyMessage: StoryObj = {
  render: () => (
    <OperationRolesList
      operationRoles={{}}
      emptyMessage="No roles have been assigned yet"
    />
  ),
}
