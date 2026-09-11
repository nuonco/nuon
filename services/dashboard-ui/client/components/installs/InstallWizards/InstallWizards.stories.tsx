export default {
  title: 'Installs/Install wizards',
}

import type { ReactNode } from 'react'
import { BreadcrumbContext } from '@/providers/breadcrumb-provider'
import { NotificationContext } from '@/providers/notification-provider'
import { SidebarContext } from '@/providers/sidebar-provider'
import { InstallDeprovisionWizard, InstallSetupWizard } from './InstallWizards'

const breadcrumb = (current: string) => ({
  breadcrumbLinks: [
    { path: '/org-mock-001', text: 'Mock organization' },
    { path: '/org-mock-001/installs', text: 'Installs' },
    { path: '', text: current },
  ],
  isLoading: false,
  updateBreadcrumb: () => {},
})

const sidebar = {
  isSidebarOpen: true,
  closeSidebar: () => {},
  openSidebar: () => {},
  toggleSidebar: () => {},
}

const notifications = {
  emitNotification: async () => false,
  permission: 'default' as NotificationPermission,
  requestPermission: async () => 'default' as NotificationPermission,
  isSupported: false,
  settings: { permissionRequested: false },
  hasRequestedPermission: false,
  muted: false,
  toggleMute: () => {},
}

const PageStory = ({
  children,
  current = 'Install setup',
}: {
  children: ReactNode
  current?: string
}) => (
  <NotificationContext.Provider value={notifications}>
    <BreadcrumbContext.Provider value={breadcrumb(current)}>
      <SidebarContext.Provider value={sidebar}>
        <div className="flex h-screen min-h-0">{children}</div>
      </SidebarContext.Provider>
    </BreadcrumbContext.Provider>
  </NotificationContext.Provider>
)

const provisionComplete = [
  { id: 'stack', label: 'Create install stack', status: 'success' },
  { id: 'runner', label: 'Connect runner', status: 'success' },
  { id: 'sandbox', label: 'Provision sandbox', status: 'success' },
  { id: 'components', label: 'Deploy components', status: 'success' },
]

const provisioning = [
  { id: 'stack', label: 'Create install stack', status: 'success' },
  { id: 'runner', label: 'Connect runner', status: 'success' },
  { id: 'sandbox', label: 'Provision sandbox', status: 'executing' },
  { id: 'components', label: 'Deploy components', status: 'pending' },
]

const teardownComplete = [
  { id: 'components', label: 'Teardown components', status: 'success' },
  { id: 'sandbox', label: 'Teardown sandbox', status: 'success' },
  { id: 'runner', label: 'Stop runner', status: 'success' },
]

const teardownInProgress = [
  { id: 'components', label: 'Teardown components', status: 'success' },
  { id: 'sandbox', label: 'Teardown sandbox', status: 'executing' },
  { id: 'runner', label: 'Stop runner', status: 'pending' },
]

export const SetupCustomerManaged = () => (
  <PageStory>
    <InstallSetupWizard
      initialStep="stack"
      initialStackOwnership="customer"
      progress={provisionComplete}
      onComplete={() => {}}
    />
  </PageStory>
)
SetupCustomerManaged.meta = { fullBleed: true }

export const SetupNuonManaged = () => (
  <PageStory>
    <InstallSetupWizard
      initialStep="stack"
      initialStackOwnership="nuon"
      initialRoleArn="arn:aws:iam::123456789012:role/nuon-stack-manager"
      progress={provisionComplete}
      onComplete={() => {}}
    />
  </PageStory>
)
SetupNuonManaged.meta = { fullBleed: true }

export const Provisioning = () => (
  <PageStory>
    <InstallSetupWizard
      initialStep="provision"
      initialStackOwnership="nuon"
      initialRoleArn="arn:aws:iam::123456789012:role/nuon-stack-manager"
      progress={provisioning}
      onComplete={() => {}}
    />
  </PageStory>
)
Provisioning.meta = { fullBleed: true }

export const ProvisionFailed = () => (
  <PageStory>
    <InstallSetupWizard
      initialStep="provision"
      initialStackOwnership="nuon"
      initialRoleArn="arn:aws:iam::123456789012:role/nuon-stack-manager"
      progress={provisioning}
      error="The sandbox could not assume its install role. Check the trust policy and retry."
      onComplete={() => {}}
    />
  </PageStory>
)
ProvisionFailed.meta = { fullBleed: true }

export const DeprovisionConfirm = () => (
  <PageStory current="Deprovision">
    <InstallDeprovisionWizard
      initialStackOwnership="customer"
      progress={teardownComplete}
      onComplete={() => {}}
    />
  </PageStory>
)
DeprovisionConfirm.meta = { fullBleed: true }

export const TeardownInProgress = () => (
  <PageStory current="Deprovision">
    <InstallDeprovisionWizard
      initialStackOwnership="nuon"
      initialStep="teardown"
      progress={teardownInProgress}
      onComplete={() => {}}
    />
  </PageStory>
)
TeardownInProgress.meta = { fullBleed: true }

export const AwaitingManualStackDestroy = () => (
  <PageStory current="Deprovision">
    <InstallDeprovisionWizard
      initialStackOwnership="customer"
      initialStep="stack"
      progress={teardownComplete}
      onComplete={() => {}}
    />
  </PageStory>
)
AwaitingManualStackDestroy.meta = { fullBleed: true }

export const DestroyingManagedStack = () => (
  <PageStory current="Deprovision">
    <InstallDeprovisionWizard
      initialStackOwnership="nuon"
      initialStep="stack"
      progress={teardownComplete}
      stackDestroying
      onComplete={() => {}}
    />
  </PageStory>
)
DestroyingManagedStack.meta = { fullBleed: true }

export const DeprovisionComplete = () => (
  <PageStory current="Deprovision">
    <InstallDeprovisionWizard
      initialStackOwnership="nuon"
      initialStep="complete"
      progress={teardownComplete}
      onComplete={() => {}}
    />
  </PageStory>
)
DeprovisionComplete.meta = { fullBleed: true }
