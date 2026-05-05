import { SurfacesContext } from '@/providers/surfaces-provider'
import { AppRolesTable } from './AppRolesTable'

export default {
  title: 'Roles/AppRolesTable',
}

const mockSurfaces = {
  modals: [],
  panels: [],
  addModal: () => '',
  removeModal: () => {},
  addPanel: () => '',
  updatePanel: () => {},
  removePanel: () => {},
  clearPanels: () => {},
}

const mockRoles = [
  {
    id: 'role-1',
    display_name: 'Provision role',
    description: 'Used to provision infrastructure in customer accounts',
    name: 'nuon-provision-role',
    type: 'provision',
    created_at: '2024-06-15T10:30:00Z',
    policies: [
      { id: 'pol-1', managed_policy_name: 'AmazonEKSClusterPolicy' },
    ],
    permissions_boundary: undefined,
  },
  {
    id: 'role-2',
    display_name: 'Deprovision role',
    description: 'Used to tear down infrastructure',
    name: 'nuon-deprovision-role',
    type: 'deprovision',
    created_at: '2024-07-01T14:00:00Z',
    policies: [],
    permissions_boundary: undefined,
  },
  {
    id: 'role-3',
    display_name: 'Maintenance role',
    description: 'Used for ongoing maintenance tasks',
    name: 'nuon-maintenance-role',
    type: 'maintenance',
    created_at: '2024-08-10T09:15:00Z',
    policies: [
      { id: 'pol-2', managed_policy_name: 'AmazonS3ReadOnlyAccess' },
      { id: 'pol-3', name: 'custom-policy', contents: btoa('{}') },
    ],
    permissions_boundary: undefined,
  },
]

export const Default = () => (
  <SurfacesContext.Provider value={mockSurfaces}>
    <AppRolesTable roles={mockRoles} />
  </SurfacesContext.Provider>
)

export const Loading = () => (
  <SurfacesContext.Provider value={mockSurfaces}>
    <AppRolesTable roles={[]} isLoading />
  </SurfacesContext.Provider>
)

export const Empty = () => (
  <SurfacesContext.Provider value={mockSurfaces}>
    <AppRolesTable roles={[]} />
  </SurfacesContext.Provider>
)
