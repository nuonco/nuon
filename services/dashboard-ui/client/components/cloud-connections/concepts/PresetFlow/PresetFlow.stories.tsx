export default { title: 'Cloud connections/Concepts/Preset flow' }

import { useEffect, useRef, type ReactNode } from 'react'
import { BreadcrumbContext } from '@/providers/breadcrumb-provider'
import { NotificationContext } from '@/providers/notification-provider'
import { SidebarContext } from '@/providers/sidebar-provider'
import { SurfacesProvider } from '@/providers/surfaces-provider'
import { useSurfaces } from '@/hooks/use-surfaces'
import { PresetFlow } from './PresetFlow'
import { TRUST_POLICY } from './mockData'
import {
  CloudConnectionDetailsPanel,
  type TVerificationDisplay,
} from './CloudConnectionDetailsPanel'
import { CloudConnectionsOverview } from './CloudConnectionsOverview'
import {
  CLOUD_CONNECTIONS,
  type TCloudConnectionOverview,
} from './overviewMockData'

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
        <SurfacesProvider>{children}</SurfacesProvider>
      </SidebarContext.Provider>
    </BreadcrumbContext.Provider>
  </NotificationContext.Provider>
)

const OverviewWithOpenDetails = ({
  connection,
  verification = 'current',
  initialSection = 'summary',
}: {
  connection: TCloudConnectionOverview
  verification?: TVerificationDisplay
  initialSection?: 'summary' | 'policy' | 'verification'
}) => {
  const { addPanel } = useSurfaces()
  const opened = useRef(false)
  useEffect(() => {
    if (opened.current) return
    opened.current = true
    addPanel(
      <CloudConnectionDetailsPanel
        connection={connection}
        initialVerification={verification}
        initialSection={initialSection}
      />
    )
  }, [addPanel, connection, initialSection, verification])

  return <CloudConnectionsOverview connections={CLOUD_CONNECTIONS} />
}

export const OverviewPopulated = () => (
  <OverviewProviders>
    <CloudConnectionsOverview connections={CLOUD_CONNECTIONS} />
  </OverviewProviders>
)
export const OverviewDetailsStacks = () => (
  <OverviewProviders>
    <OverviewWithOpenDetails connection={CLOUD_CONNECTIONS[0]} />
  </OverviewProviders>
)
export const OverviewDetailsCustom = () => (
  <OverviewProviders>
    <OverviewWithOpenDetails
      connection={CLOUD_CONNECTIONS[2]}
      initialSection="policy"
    />
  </OverviewProviders>
)
export const OverviewDetailsError = () => (
  <OverviewProviders>
    <OverviewWithOpenDetails
      connection={CLOUD_CONNECTIONS[3]}
      initialSection="verification"
    />
  </OverviewProviders>
)
export const OverviewDetailsVerifying = () => (
  <OverviewProviders>
    <OverviewWithOpenDetails
      connection={CLOUD_CONNECTIONS[0]}
      verification="verifying"
      initialSection="verification"
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
export const TrustPolicyEditing = () => (
  <PresetFlow
    initialStep={2}
    initialAccess="custom"
    showTrustPolicy
    startTrustEditing
    initialTrustPolicy={TRUST_POLICY.replace(
      '"sts.amazonaws.com"',
      '"example.invalid"'
    )}
  />
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
OverviewDetailsStacks.meta = { fullBleed: true }
OverviewDetailsCustom.meta = { fullBleed: true }
OverviewDetailsError.meta = { fullBleed: true }
OverviewDetailsVerifying.meta = { fullBleed: true }
OverviewEmpty.meta = { fullBleed: true }
OverviewLoading.meta = { fullBleed: true }
AccessPreset.meta = { fullBleed: true }
PresetSelected.meta = { fullBleed: true }
CustomSelected.meta = { fullBleed: true }
TrustPolicyEditing.meta = { fullBleed: true }
RunInCloudTerraform.meta = { fullBleed: true }
RunInCloudAWSCLI.meta = { fullBleed: true }
RunInCloudCloudFormation.meta = { fullBleed: true }
Verify.meta = { fullBleed: true }
Verifying.meta = { fullBleed: true }
Verified.meta = { fullBleed: true }
FailedAtTrust.meta = { fullBleed: true }
FailedAtPermissions.meta = { fullBleed: true }
