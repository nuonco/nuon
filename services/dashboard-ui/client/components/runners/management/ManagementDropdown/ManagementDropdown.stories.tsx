export default {
  title: 'Runners/Management/ManagementDropdown',
}

import { ManagementDropdown } from './ManagementDropdown'

const mockRunner = {
  id: 'runner-1',
  name: 'my-runner',
} as any

const mockSettings = {
  id: 'settings-1',
  container_image_tag: 'v1.2.3',
  binary_version: 'v0.5.0',
} as any

export const BuildRunner = () => (
  <div className="p-4">
    <ManagementDropdown
      runner={mockRunner}
      isInstallRunner={false}
      settings={mockSettings}
    />
  </div>
)

export const InstallRunner = () => (
  <div className="p-4">
    <ManagementDropdown
      runner={mockRunner}
      isInstallRunner
      settings={mockSettings}
    />
  </div>
)

export const InstallRunnerInstanceOnly = () => (
  <div className="p-4">
    <ManagementDropdown
      runner={mockRunner}
      isInstallRunner
      settings={mockSettings}
      hasMngProcess={false}
      hasInstanceProcess
    />
  </div>
)

export const InstallRunnerManagerOnly = () => (
  <div className="p-4">
    <ManagementDropdown
      runner={mockRunner}
      isInstallRunner
      settings={mockSettings}
      hasMngProcess
      hasInstanceProcess={false}
    />
  </div>
)
