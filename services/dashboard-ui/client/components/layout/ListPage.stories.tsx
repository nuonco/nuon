export default {
  title: 'Layout/ListPage',
}

import type { ReactNode } from 'react'
import { Button } from '@/components/common/Button'
import { Table } from '@/components/common/Table'
import { BreadcrumbContext } from '@/providers/breadcrumb-provider'
import { NotificationContext } from '@/providers/notification-provider'
import { SidebarContext } from '@/providers/sidebar-provider'
import { ListPage } from './ListPage'

type TRow = { name: string; status: string }

const columns = [
  { accessorKey: 'name', header: 'Name' },
  { accessorKey: 'status', header: 'Status' },
]

const data: TRow[] = [
  { name: 'api', status: 'Active' },
  { name: 'worker', status: 'Active' },
  { name: 'web', status: 'Provisioning' },
]

const mockBreadcrumb = {
  breadcrumbLinks: [
    { path: '/org-001', text: 'My org' },
    { path: '/org-001/installs', text: 'Installs' },
  ],
  isLoading: false,
  updateBreadcrumb: () => {},
}

const mockSidebar = {
  isSidebarOpen: true,
  closeSidebar: () => {},
  openSidebar: () => {},
  toggleSidebar: () => {},
}

const mockNotifications = {
  emitNotification: async () => false,
  permission: 'default' as NotificationPermission,
  requestPermission: async () => 'default' as NotificationPermission,
  isSupported: false,
  settings: { permissionRequested: false },
  hasRequestedPermission: false,
  muted: false,
  toggleMute: () => {},
}

const ShellProviders = ({ children }: { children: ReactNode }) => (
  <NotificationContext.Provider value={mockNotifications}>
    <BreadcrumbContext.Provider value={mockBreadcrumb}>
      <SidebarContext.Provider value={mockSidebar}>
        {children}
      </SidebarContext.Provider>
    </BreadcrumbContext.Provider>
  </NotificationContext.Provider>
)

export const Page = () => (
  <ShellProviders>
    <ListPage
      variant="page"
      title="Installs"
      description="View and manage all deployed installs here."
      createAction={<Button variant="primary">Create install</Button>}
    >
      <Table
        columns={columns}
        data={data}
        pagination={{ hasNext: true, offset: 0, limit: 20 }}
      />
    </ListPage>
  </ShellProviders>
)

export const PageWithTableFilters = () => (
  <ShellProviders>
    <ListPage
      variant="page"
      title="Installs"
      description="View and manage all deployed installs here."
      createAction={<Button variant="primary">Create install</Button>}
    >
      <Table
        columns={columns}
        data={data}
        filterActions={
          <div className="flex items-center gap-3">
            <Button variant="secondary">Labels</Button>
            <Button variant="secondary">Branches</Button>
          </div>
        }
        pagination={{ hasNext: true, offset: 0, limit: 20 }}
      />
    </ListPage>
  </ShellProviders>
)

export const Section = () => (
  <ListPage
    title="App components"
    description="Manage the components that make up your application."
    createAction={<Button variant="primary">Create component</Button>}
  >
    <Table columns={columns} data={data} />
  </ListPage>
)

export const SectionEmpty = () => (
  <ListPage
    title="App components"
    description="Manage the components that make up your application."
    createAction={<Button variant="primary">Create component</Button>}
  >
    <Table
      columns={columns}
      data={[]}
      emptyStateProps={{
        emptyTitle: 'No components yet',
        emptyMessage: 'Sync your app config to add components.',
      }}
    />
  </ListPage>
)

export const SectionLoading = () => (
  <ListPage
    title="App components"
    description="Manage the components that make up your application."
  >
    <Table columns={columns} data={[]} isLoading />
  </ListPage>
)
