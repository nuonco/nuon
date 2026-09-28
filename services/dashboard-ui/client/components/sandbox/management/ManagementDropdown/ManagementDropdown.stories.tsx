export default {
  title: 'Features / Sandbox / Management / Management dropdown',
}

import { ManagementDropdown } from './ManagementDropdown'

export const Default = () => (
  <div className="p-4">
    <ManagementDropdown />
  </div>
)

export const WithWorkspace = () => (
  <div className="p-4">
    <ManagementDropdown workspaceId="workspace-1" />
  </div>
)
