export default { title: 'Cloud connections/Concepts/Preset flow' }

import type { ReactNode } from 'react'
import { BreadcrumbContext } from '@/providers/breadcrumb-provider'
import { NotificationContext } from '@/providers/notification-provider'
import { SidebarContext } from '@/providers/sidebar-provider'
import { PresetFlow } from './PresetFlow'
import { CloudConnectionDetailPage } from './CloudConnectionDetailPage'
import { CloudConnectionsOverview } from './CloudConnectionsOverview'
import { CLOUD_CONNECTIONS } from './overviewMockData'

const OverviewProviders = ({ children }: { children: ReactNode }) => (
  <NotificationContext.Provider
    value={{
      emitNotification: async () => false,
      permission: 'default',
      requestPermission: async () => 'default',
      isSupported: false,
      settings: { permissionRequested: false },
      hasRequestedPermission: false,
      muted: false,
      toggleMute: () => {},
    }}
  >
    <BreadcrumbContext.Provider
      value={{
        breadcrumbLinks: [
          { path: '/org-mock-001', text: 'Example org' },
          {
            path: '/org-mock-001/cloud-connections',
            text: 'Cloud connections',
          },
        ],
        isLoading: false,
        updateBreadcrumb: () => {},
      }}
    >
      <SidebarContext.Provider
        value={{
          isSidebarOpen: true,
          closeSidebar: () => {},
          openSidebar: () => {},
          toggleSidebar: () => {},
        }}
      >
        {children}
      </SidebarContext.Provider>
    </BreadcrumbContext.Provider>
  </NotificationContext.Provider>
)

export const OverviewPopulated = () => (
  <OverviewProviders>
    <CloudConnectionsOverview connections={CLOUD_CONNECTIONS} />
  </OverviewProviders>
)
export const DetailOverviewStacks = () => (
  <OverviewProviders>
    <CloudConnectionDetailPage connection={CLOUD_CONNECTIONS[0]} />
  </OverviewProviders>
)
export const DetailOverviewCustom = () => (
  <OverviewProviders>
    <CloudConnectionDetailPage connection={CLOUD_CONNECTIONS[2]} />
  </OverviewProviders>
)
export const DetailInstalls = () => (
  <OverviewProviders>
    <CloudConnectionDetailPage
      connection={CLOUD_CONNECTIONS[0]}
      selectedTab="installs"
    />
  </OverviewProviders>
)
export const DetailInstallsEmpty = () => (
  <OverviewProviders>
    <CloudConnectionDetailPage
      connection={CLOUD_CONNECTIONS[0]}
      selectedTab="installs"
      installsEmpty
    />
  </OverviewProviders>
)
export const DetailVerificationVerified = () => (
  <OverviewProviders>
    <CloudConnectionDetailPage
      connection={CLOUD_CONNECTIONS[0]}
      selectedTab="verification"
    />
  </OverviewProviders>
)
export const DetailVerificationError = () => (
  <OverviewProviders>
    <CloudConnectionDetailPage
      connection={CLOUD_CONNECTIONS[3]}
      selectedTab="verification"
    />
  </OverviewProviders>
)
export const DetailVerifying = () => (
  <OverviewProviders>
    <CloudConnectionDetailPage
      connection={CLOUD_CONNECTIONS[0]}
      selectedTab="verification"
      initialVerification="verifying"
    />
  </OverviewProviders>
)
export const OverviewEmpty = () => (
  <OverviewProviders>
    <CloudConnectionsOverview connections={[]} />
  </OverviewProviders>
)
export const OverviewLoading = () => (
  <OverviewProviders>
    <CloudConnectionsOverview connections={[]} isLoading />
  </OverviewProviders>
)

export const CloudAndAccount = () => <PresetFlow initialStep={1} />
export const AccessPreset = () => <PresetFlow initialStep={2} />
export const PresetSelected = () => (
  <PresetFlow initialStep={2} initialAccess="preset" />
)
export const CustomSelected = () => (
  <PresetFlow initialStep={2} initialAccess="custom" showTrustPolicy />
)
export const RunInCloudTerraform = () => (
  <PresetFlow
    initialStep={3}
    initialAccess="preset"
    initialFormat="terraform"
  />
)
export const RunInCloudAWSCLI = () => (
  <PresetFlow initialStep={3} initialAccess="preset" initialFormat="cli" />
)
export const RunInCloudCloudFormation = () => (
  <PresetFlow
    initialStep={3}
    initialAccess="preset"
    initialFormat="cloudformation"
  />
)
export const Verify = () => (
  <PresetFlow initialStep={4} initialAccess="preset" />
)
export const Verifying = () => (
  <PresetFlow initialStep={4} initialAccess="preset" verification="verifying" />
)
export const Verified = () => (
  <PresetFlow initialStep={4} initialAccess="preset" verification="verified" />
)
export const FailedAtTrust = () => (
  <PresetFlow
    initialStep={4}
    initialAccess="preset"
    verification="failed-trust"
  />
)
export const FailedAtPermissions = () => (
  <PresetFlow
    initialStep={4}
    initialAccess="preset"
    verification="failed-permissions"
  />
)

CloudAndAccount.meta = { fullBleed: true }
OverviewPopulated.meta = { fullBleed: true }
DetailOverviewStacks.meta = { fullBleed: true }
DetailOverviewCustom.meta = { fullBleed: true }
DetailInstalls.meta = { fullBleed: true }
DetailInstallsEmpty.meta = { fullBleed: true }
DetailVerificationVerified.meta = { fullBleed: true }
DetailVerificationError.meta = { fullBleed: true }
DetailVerifying.meta = { fullBleed: true }
OverviewEmpty.meta = { fullBleed: true }
OverviewLoading.meta = { fullBleed: true }
AccessPreset.meta = { fullBleed: true }
PresetSelected.meta = { fullBleed: true }
CustomSelected.meta = { fullBleed: true }
RunInCloudTerraform.meta = { fullBleed: true }
RunInCloudAWSCLI.meta = { fullBleed: true }
RunInCloudCloudFormation.meta = { fullBleed: true }
Verify.meta = { fullBleed: true }
Verifying.meta = { fullBleed: true }
Verified.meta = { fullBleed: true }
FailedAtTrust.meta = { fullBleed: true }
FailedAtPermissions.meta = { fullBleed: true }
