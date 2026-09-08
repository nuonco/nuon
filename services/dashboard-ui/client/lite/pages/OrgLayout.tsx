import { Outlet } from 'react-router'
import { useConfig } from '@/hooks/use-config'
import { Text } from '../components/atoms/Text'
import type { INavItem } from '../components/molecules/NavLink'
import { ThemeSwitcher } from '../components/molecules/ThemeSwitcher'
import { Breadcrumb } from '../components/molecules/Breadcrumb'
import { UserDropdown } from '../components/organisms/UserDropdown'
import { SurfaceHost } from '../components/organisms/surfaces'
import { DashboardShell } from '../components/templates/DashboardShell'
import { useBreadcrumbItems } from '../hooks/use-breadcrumbs'
import { useCurrentUser } from '../hooks/use-current-user'
import { BreadcrumbProvider } from '../providers/breadcrumb-provider'
import { OrgProvider, useOrg } from '../providers/org-provider'

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
  const navigation = orgNavigation(orgId ?? '')

  return (
    <DashboardShell
      primaryNav={navigation.primary}
      secondaryNav={navigation.secondary}
      homeHref={`/${orgId ?? ''}`}
      headerLeading={<Breadcrumb items={breadcrumbs} />}
      headerActions={<ThemeSwitcher />}
      userMenu={
        <UserDropdown
          user={user}
          loading={isLoadingUser}
          signOutHref={`${config.authServiceUrl ?? ''}/logout`}
          stretch
        />
      }
      statusBar={
        <div className="flex h-8 items-center justify-between gap-4 px-4">
          <Text
            variant="label"
            color="secondary"
            loading={isLoading}
            loadingWidth={16}
          >
            {org?.name ?? 'Organization unavailable'}
          </Text>
          <Text variant="label" color="tertiary">
            {error ? 'Connection issue' : `Version ${config.version ?? 'dev'}`}
          </Text>
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
      <SurfaceHost scope="org">
        <OrgShell />
      </SurfaceHost>
    </BreadcrumbProvider>
  </OrgProvider>
)
