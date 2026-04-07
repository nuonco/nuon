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
} as any

export const BuildRunner = () => (
  <div className="p-4">
    <ManagementDropdown
      runner={mockRunner}
      isManaged={false}
      isInstallRunner={false}
      settings={mockSettings}
    />
  </div>
)

export const InstallRunner = () => (
  <div className="p-4">
    <ManagementDropdown
      runner={mockRunner}
      isManaged
      isInstallRunner
      settings={mockSettings}
    />
  </div>
)
