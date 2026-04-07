export default {
  title: 'Install Components/ManagementDropdown',
}

import { ManagementDropdown } from './ManagementDropdown'
import type { TComponent, TInstallComponent } from '@/types'

const mockComponent: TComponent = {
  id: 'comp-1',
  name: 'api',
  type: 'helm_chart',
} as TComponent

const mockInstallComponent: TInstallComponent = {
  id: 'ic-1',
  component: mockComponent,
  terraform_workspace: undefined,
} as TInstallComponent

export const Default = () => (
  <ManagementDropdown
    component={mockComponent}
    currentBuildId="build-1"
    currentDeployStatus="active"
    installComponent={mockInstallComponent}
  />
)

export const Inactive = () => (
  <ManagementDropdown
    component={mockComponent}
    currentBuildId="build-1"
    currentDeployStatus="inactive"
    installComponent={mockInstallComponent}
  />
)

export const TerraformComponent = () => (
  <ManagementDropdown
    component={{ ...mockComponent, type: 'terraform_module' } as TComponent}
    currentBuildId="build-1"
    currentDeployStatus="active"
    installComponent={{ ...mockInstallComponent, terraform_workspace: { id: 'ws-1' } } as TInstallComponent}
  />
)
