export default {
  title: 'Layout/DetailPage',
}

import type { ReactNode } from 'react'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { Duration } from '@/components/common/Duration'
import { LabeledStatus } from '@/components/common/LabeledStatus'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { BreadcrumbContext } from '@/providers/breadcrumb-provider'
import { NotificationContext } from '@/providers/notification-provider'
import { SidebarContext } from '@/providers/sidebar-provider'
import { DetailHeader } from './DetailHeader'
import { DetailPage } from './DetailPage'
import { HistoryPanelButton, HistoryRail } from './HistoryRail'

const created = '2026-08-24T10:12:00Z'
const updated = '2026-08-24T10:19:42Z'

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

const runHeader = (
  <DetailHeader
    title="api deploy"
    id="dep01hzk8t3fqp2r9x4m7wcn5vb"
    identity={
      <Time time={created} format="relative" variant="subtext" theme="info" />
    }
    actions={<Button variant="secondary">Manage</Button>}
    metadata={
      <>
        <LabeledStatus label="Status" statusProps={{ status: 'active' }} />
        <LabeledValue label="Duration">
          <Duration variant="subtext" beginTime={created} endTime={updated} />
        </LabeledValue>
        <LabeledValue label="Install">
          <Link href="#">acme-payments</Link>
        </LabeledValue>
      </>
    }
  />
)

const history = (
  <div className="flex flex-col gap-3">
    {['Deployed 2 hours ago', 'Deployed 1 day ago', 'Deployed 3 days ago'].map(
      (entry) => (
        <Card key={entry} className="!p-4 !gap-2">
          <Text variant="subtext">{entry}</Text>
        </Card>
      )
    )}
  </div>
)

export const Run = () => (
  <DetailPage
    header={runHeader}
    tabNav={{
      basePath: '',
      tabs: [
        { path: '/', text: 'Summary' },
        { path: '/logs', text: 'Logs' },
        { path: '/plan', text: 'Plan' },
      ],
    }}
  >
    <Text theme="neutral">Tab content</Text>
  </DetailPage>
)

export const RunWithBanner = () => (
  <DetailPage
    header={runHeader}
    banners={
      <Banner theme="warn">
        <Text>This deploy is waiting on an approval.</Text>
      </Banner>
    }
    tabNav={{
      basePath: '',
      tabs: [
        { path: '/', text: 'Summary' },
        { path: '/logs', text: 'Logs' },
      ],
    }}
  >
    <Text theme="neutral">Tab content</Text>
  </DetailPage>
)

export const EntityWithHistoryRail = () => (
  <DetailPage
    header={
      <DetailHeader
        backLink={false}
        title="Sandbox details"
        id="sbx01hzk8t3fqp2r9x4m7wcn5vb"
        actions={
          <>
            <HistoryPanelButton title="Sandbox history" history={history} />
            <Button variant="secondary">Manage</Button>
          </>
        }
      />
    }
  >
    <HistoryRail title="Sandbox history" history={history}>
      <Card>
        <Text variant="base" weight="strong">
          Configuration
        </Text>
        <Text theme="neutral">Section body</Text>
      </Card>
      <Card>
        <Text variant="base" weight="strong">
          Terraform workspace
        </Text>
        <Text theme="neutral">Section body</Text>
      </Card>
    </HistoryRail>
  </DetailPage>
)

export const Document = () => (
  <DetailPage
    header={
      <DetailHeader
        backLink={false}
        title="Install state"
        description="Raw state data for this install."
        actions={<Button variant="secondary">Download state</Button>}
      />
    }
  >
    <Card>
      <Text theme="neutral">Document body</Text>
    </Card>
  </DetailPage>
)

export const Page = () => (
  <ShellProviders>
    <DetailPage variant="page" header={runHeader}>
      <Text theme="neutral">Page body</Text>
    </DetailPage>
  </ShellProviders>
)
