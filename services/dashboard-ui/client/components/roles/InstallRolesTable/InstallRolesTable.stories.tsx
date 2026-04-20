import { SurfacesContext } from '@/providers/surfaces-provider'
import { InstallRolesTable } from './InstallRolesTable'

export default {
  title: 'Roles/InstallRolesTable',
}

const mockSurfaces = {
  modals: [],
  panels: [],
  addModal: () => '',
  removeModal: () => {},
  addPanel: () => '',
  removePanel: () => {},
  clearPanels: () => {},
}

const mockRoles = [
  {
    id: 'ir-1',
    install_id: 'inst-1',
    app_role_config_id: 'role-1',
    app_role_config: {
      id: 'role-1',
      display_name: 'Provision role',
      description: 'Used to provision infrastructure in customer accounts',
      name: 'nuon-provision-role',
      type: 'provision',
      created_at: '2024-06-15T10:30:00Z',
      enabled: true,
      arn: 'arn:aws:iam::123456789012:role/nuon-provision-role',
      policies: [
        { id: 'pol-1', managed_policy_name: 'AmazonEKSClusterPolicy' },
      ],
      permissions_boundary: '',
    },
    enabled: true,
    provisioned: true,
    role_id: 'arn:aws:iam::123456789012:role/nuon-provision-role',
    created_at: '2024-06-15T10:30:00Z',
  },
  {
    id: 'ir-2',
    install_id: 'inst-1',
    app_role_config_id: 'role-2',
    app_role_config: {
      id: 'role-2',
      display_name: 'Deprovision role',
      description: 'Used to tear down infrastructure',
      name: 'nuon-deprovision-role',
      type: 'deprovision',
      created_at: '2024-07-01T14:00:00Z',
      enabled: true,
      arn: '',
      policies: [],
      permissions_boundary: '',
    },
    enabled: true,
    provisioned: false,
    role_id: '',
    created_at: '2024-07-01T14:00:00Z',
  },
]

export const Default = () => (
  <SurfacesContext.Provider value={mockSurfaces}>
    <InstallRolesTable roles={mockRoles} />
  </SurfacesContext.Provider>
)

export const Loading = () => (
  <SurfacesContext.Provider value={mockSurfaces}>
    <InstallRolesTable roles={[]} isLoading />
  </SurfacesContext.Provider>
)

export const Empty = () => (
  <SurfacesContext.Provider value={mockSurfaces}>
    <InstallRolesTable roles={[]} />
  </SurfacesContext.Provider>
)
