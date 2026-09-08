import { Outlet } from 'react-router'
import { useConfig } from '@/hooks/use-config'
import { Text } from '../components/atoms/Text'
import type { INavItem } from '../components/molecules/NavLink'
import { Breadcrumb } from '../components/molecules/Breadcrumb'
import { OrgProfile } from '../components/molecules/OrgProfile'
import { OrgSwitcherMenu } from '../components/organisms/OrgSwitcherMenu'
import { UserDropdownContainer as UserDropdown } from '../components/organisms/UserDropdownContainer'
import { SurfaceHost } from '../components/organisms/surfaces'
import { DashboardShell } from '../components/templates/DashboardShell'
import { useBreadcrumbItems } from '../hooks/use-breadcrumbs'
import { useCurrentUser } from '../hooks/use-current-user'
import { useStatusBarContent } from '../hooks/use-status-bar'
import { BreadcrumbProvider } from '../providers/breadcrumb-provider'
import { OrgProvider, useOrg } from '../providers/org-provider'
import { StatusBarProvider } from '../providers/status-bar-provider'

export const orgNavigation = (orgId: string) => {
  const primary: INavItem[] = [
    {
      href: `/${orgId}`,
      label: 'Dashboard',
      icon: 'HouseIcon',
      shortcut: 'g d',
      end: true,
    },
    {
      href: `/${orgId}/apps`,
      label: 'Apps',
      icon: 'AppWindowIcon',
      shortcut: 'g a',
    },
    {
      href: `/${orgId}/installs`,
      label: 'Installs',
      icon: 'CubeIcon',
      shortcut: 'g i',
    },
  ]
  const secondary: INavItem[] = [
    {
      href: `/${orgId}/teams`,
      label: 'Team',
      icon: 'UsersThreeIcon',
      shortcut: 'g t',
    },
    {
      href: `/${orgId}/settings`,
      label: 'Settings',
      icon: 'GearIcon',
      shortcut: 'g s',
    },
    {
      href: 'https://docs.nuon.co',
      label: 'Developer docs',
      icon: 'BookOpenTextIcon',
      external: true,
    },
  ]

  return { primary, secondary }
}

const OrgShell = () => {
  const config = useConfig()
  const { org, orgId, isLoading, error } = useOrg()
  const { user, isLoading: isLoadingUser } = useCurrentUser()
  const breadcrumbs = useBreadcrumbItems()
  const statusBarContent = useStatusBarContent()
  const navigation = orgNavigation(orgId ?? '')

  return (
    <DashboardShell
      primaryNav={navigation.primary}
      secondaryNav={navigation.secondary}
      homeHref={`/${orgId ?? ''}`}
      headerLeading={<Breadcrumb items={breadcrumbs} />}
      userMenu={
        <UserDropdown
          user={user}
          loading={isLoadingUser}
          signOutHref={`${config.authServiceUrl ?? ''}/logout`}
          stretch
          org={org}
          orgLoading={isLoading}
          orgSwitcher={<OrgSwitcherMenu />}
        />
      }
      statusBar={
        <div className="flex h-7 items-stretch justify-between">
          <div className="flex min-w-0 items-stretch">
            <span className="modeline-point-right flex items-center bg-surface-modeline pr-5 pl-3">
              <OrgProfile org={org} loading={isLoading} variant="modeline" />
            </span>
            {statusBarContent ? (
              <span className="flex min-w-0 items-center px-2">
                {statusBarContent}
              </span>
            ) : null}
          </div>
          <span className="modeline-point-left flex shrink-0 items-center bg-surface-modeline pr-3 pl-5">
            <Text variant="label" family="mono" color="tertiary">
              {error ? 'disconnected' : `v${config.version ?? 'dev'}`}
            </Text>
          </span>
        </div>
      }
    >
      <Outlet />
    </DashboardShell>
  )
}

export const OrgLayout = () => (
  <OrgProvider>
    <BreadcrumbProvider>
      <StatusBarProvider>
        <SurfaceHost scope="org">
          <OrgShell />
        </SurfaceHost>
      </StatusBarProvider>
    </BreadcrumbProvider>
  </OrgProvider>
)
