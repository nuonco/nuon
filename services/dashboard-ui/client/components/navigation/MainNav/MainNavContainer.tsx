import { useConfig } from '@/hooks/use-config'
import { useDashboardPreferences } from '@/hooks/use-dashboard-preferences'
import { useOrg } from '@/hooks/use-org'
import { useSidebar } from '@/hooks/use-sidebar'
import { MainNav } from './MainNav'

export const MainNavContainer = () => {
  const { org } = useOrg()
  const { datadogEnv } = useConfig()
  const { isSidebarOpen } = useSidebar()
  const { isInstallsTabEnabled } = useDashboardPreferences()

  if (!org) return null

  const customerPortalUrl =
    datadogEnv === 'stage'
      ? 'https://customers.stage.nuon.co'
      : datadogEnv === 'local'
        ? 'http://localhost:8080'
        : 'https://customers.nuon.co'

  return (
    <MainNav
      org={org}
      isSidebarOpen={isSidebarOpen}
      showInstalls={isInstallsTabEnabled}
      hasCustomerPortal={false}
      customerPortalUrl={customerPortalUrl}
    />
  )
}
