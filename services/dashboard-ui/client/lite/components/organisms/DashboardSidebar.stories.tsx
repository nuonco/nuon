import { useEffect } from 'react'
import { ComponentDocs } from '../__stories__/ComponentDocs'
import {
  DashboardShellProvider,
  useDashboardShell,
} from '../../providers/dashboard-shell-provider'
import { UserDropdown } from './UserDropdown'
import { DashboardSidebar } from './DashboardSidebar'

export default {
  title: 'lite/organisms/DashboardSidebar',
}

const PRIMARY = [
  {
    href: '/',
    label: 'Dashboard',
    icon: 'HouseIcon' as const,
    shortcut: 'g d',
    end: true,
  },
  {
    href: '/apps',
    label: 'Apps',
    icon: 'AppWindowIcon' as const,
    shortcut: 'g a',
  },
  {
    href: '/installs',
    label: 'Installs',
    icon: 'CubeIcon' as const,
    shortcut: 'g i',
  },
]

const SECONDARY = [
  {
    href: '/settings',
    label: 'Settings',
    icon: 'GearIcon' as const,
    shortcut: 'g s',
  },
  {
    href: 'https://docs.nuon.co',
    label: 'Developer docs',
    icon: 'BookOpenTextIcon' as const,
    external: true,
  },
]

const USER = {
  name: 'Alex Morgan',
  email: 'alex@example.com',
}

export const Overview = () => (
  <ComponentDocs
    name="DashboardSidebar"
    tier="organism"
    summary="The floating glass navigation surface for DashboardShell."
    use={[
      'Use once as the persistent navigation column in DashboardShell.',
      'Supply navigation and user controls from containers.',
    ]}
    avoid={[
      'Do not fetch user data inside the sidebar.',
      'Do not place organization switching in the sidebar.',
      'Do not persist mobile drawer visibility.',
    ]}
    rules={[
      'Desktop supports expanded and icon-rail layouts.',
      'The logo and navigation labels remain mounted while width and opacity animate.',
      'Mobile uses a modal drawer with its user menu in the footer.',
      'The glass surface is composed from Card.',
    ]}
    props={[
      {
        name: 'primaryNav',
        type: 'INavItem[]',
        description: 'Primary dashboard destinations.',
      },
      {
        name: 'secondaryNav',
        type: 'INavItem[]',
        description: 'Secondary destinations anchored near the footer.',
      },
      {
        name: 'userMenu',
        type: 'ReactNode',
        description: 'Mobile user control.',
      },
    ]}
  />
)

export const Expanded = () => (
  <DashboardShellProvider initialDesktopExpanded>
    <div className="h-screen">
      <DashboardSidebar primaryNav={PRIMARY} secondaryNav={SECONDARY} />
    </div>
  </DashboardShellProvider>
)
Expanded.meta = { fullBleed: true }

export const Collapsed = () => (
  <DashboardShellProvider initialDesktopExpanded={false}>
    <div className="h-screen">
      <DashboardSidebar primaryNav={PRIMARY} secondaryNav={SECONDARY} />
    </div>
  </DashboardShellProvider>
)
Collapsed.meta = { fullBleed: true }

const MobileSidebar = () => {
  const { desktop, openMobileSidebar } = useDashboardShell()

  useEffect(() => {
    if (!desktop) openMobileSidebar()
  }, [desktop, openMobileSidebar])

  return (
    <DashboardSidebar
      primaryNav={PRIMARY}
      secondaryNav={SECONDARY}
      userMenu={
        <UserDropdown
          user={USER}
          signOutHref="https://auth.example.com/logout"
          side="top"
          align="start"
          stretch
        />
      }
    />
  )
}

export const MobileDrawer = () => (
  <DashboardShellProvider>
    <div className="h-screen">
      <MobileSidebar />
    </div>
  </DashboardShellProvider>
)
MobileDrawer.meta = { fullBleed: true }
