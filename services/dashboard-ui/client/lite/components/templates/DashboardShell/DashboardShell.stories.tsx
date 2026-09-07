import { ComponentDocs } from '../../__stories__/ComponentDocs'
import { Card } from '../../atoms/Card'
import { Text } from '../../atoms/Text'
import { UserDropdown } from '../../organisms/UserDropdown'
import { DashboardShell } from './DashboardShell'

export default {
  title: 'lite/templates/DashboardShell',
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
    href: '/team',
    label: 'Team',
    icon: 'UsersThreeIcon' as const,
    shortcut: 'g t',
  },
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

const AccountControl = () => (
  <UserDropdown
    user={USER}
    signOutHref="https://auth.example.com/logout"
    side="top"
  />
)

const StatusBar = () => (
  <div className="flex h-8 items-center px-4">
    <Text variant="label" color="secondary">
      All systems operational
    </Text>
  </div>
)

const Content = ({ rows = 6 }: { rows?: number }) => (
  <div className="mx-auto flex w-full max-w-5xl flex-col gap-4">
    <div>
      <Text as="h1" variant="title">
        Dashboard
      </Text>
      <Text as="p" variant="caption" color="secondary">
        Review applications and installations across your organization.
      </Text>
    </div>
    {Array.from({ length: rows }, (_, index) => (
      <Card key={index} className="min-h-28">
        <Text weight="semibold">Workspace section {index + 1}</Text>
        <Text as="p" variant="caption" color="secondary">
          Page content scrolls independently from the sidebar, header, and
          status bar.
        </Text>
      </Card>
    ))}
  </div>
)

const Shell = ({
  initialDesktopExpanded,
  rows,
  status = true,
}: {
  initialDesktopExpanded?: boolean
  rows?: number
  status?: boolean
}) => (
  <DashboardShell
    primaryNav={PRIMARY}
    secondaryNav={SECONDARY}
    userMenu={<AccountControl />}
    headerLeading={
      <Text variant="caption" weight="semibold">
        Example organization
      </Text>
    }
    statusBar={status ? <StatusBar /> : undefined}
    initialDesktopExpanded={initialDesktopExpanded}
  >
    <Content rows={rows} />
  </DashboardShell>
)

export const Overview = () => (
  <ComponentDocs
    name="DashboardShell"
    tier="template"
    summary="The persistent sidebar and content frame for the Lite dashboard."
    use={[
      'Use as the route layout around organization dashboard pages.',
      'Supply resolved navigation and chrome controls from an outer container.',
    ]}
    avoid={[
      'Do not put page headings, tabs, or wizard progress in the shell API.',
      'Do not create another page-level scroll container inside it.',
    ]}
    rules={[
      'Desktop sidebar preference persists independently from mobile drawer state.',
      'The user menu moves between desktop header and mobile sidebar footer.',
      'The header sticks within the page scroll region and gains its glass surface after scrolling.',
      'The sidebar and full-width status bar remain outside the page scroll region.',
    ]}
    props={[
      {
        name: 'primaryNav',
        type: 'INavItem[]',
        description: 'Primary dashboard destinations and shortcuts.',
      },
      {
        name: 'secondaryNav',
        type: 'INavItem[]',
        description: 'Secondary sidebar destinations.',
      },
      {
        name: 'userMenu',
        type: 'ReactNode',
        description: 'User control moved to the correct responsive location.',
      },
      {
        name: 'statusBar',
        type: 'ReactNode',
        description: 'Pinned status content below the main scroll region.',
      },
    ]}
  />
)

export const Expanded = () => <Shell initialDesktopExpanded />
Expanded.meta = { fullBleed: true }

export const Collapsed = () => <Shell initialDesktopExpanded={false} />
Collapsed.meta = { fullBleed: true }

export const Scrolling = () => <Shell initialDesktopExpanded rows={18} />
Scrolling.meta = { fullBleed: true }

export const WithoutStatusBar = () => (
  <Shell initialDesktopExpanded status={false} />
)
WithoutStatusBar.meta = { fullBleed: true }

export const Responsive = () => <Shell initialDesktopExpanded />
Responsive.meta = { fullBleed: true }
