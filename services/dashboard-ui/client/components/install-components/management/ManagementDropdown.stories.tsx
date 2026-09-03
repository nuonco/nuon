export default {
  title: 'Install Components/ManagementDropdown',
}

import { ManagementDropdown } from './ManagementDropdown'
import type { TComponent, TComponentConfig, TInstallComponent } from '@/types'

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

const inConfig = { component_id: 'comp-1' } as TComponentConfig

export const ActiveInConfig = () => (
  <ManagementDropdown
    component={mockComponent}
    componentConfig={inConfig}
    currentBuildId="build-1"
    currentDeployStatus="active"
    installComponent={mockInstallComponent}
  />
)

export const InactiveInConfig = () => (
  <ManagementDropdown
    component={mockComponent}
    componentConfig={inConfig}
    currentBuildId="build-1"
    currentDeployStatus="inactive"
    installComponent={mockInstallComponent}
  />
)

export const InactiveRemovedFromConfig = () => (
  <ManagementDropdown
    component={mockComponent}
    currentBuildId="build-1"
    currentDeployStatus="inactive"
    installComponent={mockInstallComponent}
  />
)

export const ConfigLoading = () => (
  <ManagementDropdown
    component={mockComponent}
    currentBuildId="build-1"
    currentDeployStatus="active"
    installComponent={mockInstallComponent}
    isConfigLoading
  />
)

export const TerraformComponent = () => (
  <ManagementDropdown
    component={{ ...mockComponent, type: 'terraform_module' } as TComponent}
    componentConfig={inConfig}
    currentBuildId="build-1"
    currentDeployStatus="active"
    installComponent={
      {
        ...mockInstallComponent,
        terraform_workspace: { id: 'ws-1' },
      } as TInstallComponent
    }
  />
)
