import { useMemo } from 'react'
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
import { useNuonStaff } from '../hooks/use-nuon-staff'
import { useStatusBarContent } from '../hooks/use-status-bar'
import { BreadcrumbProvider } from '../providers/breadcrumb-provider'
import { ModulesProvider, useModules } from '../providers/modules-provider'
import { OrgProvider, useOrg } from '../providers/org-provider'
import { StatusBarProvider } from '../providers/status-bar-provider'
import { modulesHref } from '../utils/hrefs'
import {
  ALL_MODULES,
  MODULES,
  moduleHref,
  type IModule,
  type TModuleId,
} from '../utils/modules'
import { settingsNavigation } from './SettingsLayout'

const navItemFor = (orgId: string, module: IModule): INavItem => ({
  href: moduleHref(orgId, module),
  label: module.nav?.label ?? module.name,
  icon: module.icon,
  shortcut: module.nav?.shortcut,
})

export const orgNavigation = (
  orgId: string,
  enabled: ReadonlySet<TModuleId> = ALL_MODULES
) => {
  const modules = MODULES.filter(
    (module) => module.nav && enabled.has(module.id)
  )
  const primary: INavItem[] = [
    {
      href: `/${orgId}`,
      label: 'Dashboard',
      icon: 'HouseIcon',
      shortcut: 'g d',
      end: true,
    },
    ...modules
      .filter((module) => module.nav?.group === 'primary')
      .map((module) => navItemFor(orgId, module)),
  ]
  const secondary: INavItem[] = modules
    .filter((module) => module.nav?.group === 'secondary')
    .map((module) => navItemFor(orgId, module))

  const settings = settingsNavigation(orgId, enabled)
  if (settings.length) {
    secondary.push({
      href: settings[0].href,
      label: 'Settings',
      icon: 'GearIcon',
      shortcut: 'g s',
    })
  }
  secondary.push({
    href: 'https://docs.nuon.co',
    label: 'Developer docs',
    icon: 'BookOpenTextIcon',
    external: true,
  })

  return { primary, secondary }
}

const OrgShell = () => {
  const config = useConfig()
  const { org, orgId, loading, error } = useOrg()
  const { enabled } = useModules()
  const { user, isLoading: isLoadingUser } = useCurrentUser()
  const { staff } = useNuonStaff()
  const breadcrumbs = useBreadcrumbItems()
  const statusBarContent = useStatusBarContent()
  const navigation = useMemo(
    () => orgNavigation(orgId ?? '', enabled),
    [enabled, orgId]
  )

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
          orgLoading={loading}
          orgSwitcher={<OrgSwitcherMenu />}
          manageModulesHref={staff && orgId ? modulesHref(orgId) : undefined}
        />
      }
      statusBar={
        <div className="flex h-7 items-stretch justify-between">
          <div className="flex min-w-0 items-stretch">
            <span className="modeline-point-right flex items-center bg-surface-modeline pr-5 pl-3">
              <OrgProfile org={org} loading={loading} variant="modeline" />
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
    <ModulesProvider>
      <BreadcrumbProvider>
        <StatusBarProvider>
          <SurfaceHost scope="org">
            <OrgShell />
          </SurfaceHost>
        </StatusBarProvider>
      </BreadcrumbProvider>
    </ModulesProvider>
  </OrgProvider>
)
