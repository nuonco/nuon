import { ComponentDocs } from '../__stories__/ComponentDocs'
import { Button } from '../atoms/Button'
import { Text } from '../atoms/Text'
import { DashboardShellProvider } from '../../providers/dashboard-shell-provider'
import { UserDropdown } from './UserDropdown'
import { DashboardHeader } from './DashboardHeader'

export default {
  title: 'lite/organisms/DashboardHeader',
}

const USER = {
  name: 'Alex Morgan',
  email: 'alex@example.com',
}

export const Overview = () => (
  <ComponentDocs
    name="DashboardHeader"
    tier="organism"
    summary="Sticky global chrome at the top of the dashboard content region."
    use={[
      'Use inside the DashboardShell scroll region above page content.',
      'Place global context in leading and global controls in actions.',
    ]}
    avoid={[
      'Do not put page headings, detail identity, or wizard actions here.',
      'Do not render a second mobile user menu in the header.',
    ]}
    rules={[
      'The sidebar control changes behavior at the desktop breakpoint.',
      'The desktop user menu occupies the trailing edge.',
      'The header gains its glass surface after content begins scrolling.',
    ]}
    props={[
      {
        name: 'leading',
        type: 'ReactNode',
        description: 'Global context shown after the sidebar control.',
      },
      {
        name: 'actions',
        type: 'ReactNode',
        description: 'Global actions shown before the user menu.',
      },
      {
        name: 'userMenu',
        type: 'ReactNode',
        description: 'Desktop account control supplied by DashboardShell.',
      },
    ]}
  />
)

export const Default = () => (
  <DashboardShellProvider initialDesktopExpanded>
    <DashboardHeader
      leading={<Text weight="semibold">Dashboard</Text>}
      actions={
        <Button variant="ghost" size="sm">
          Search
        </Button>
      }
      userMenu={
        <UserDropdown
          user={USER}
          signOutHref="https://auth.example.com/logout"
          compact
        />
      }
    />
  </DashboardShellProvider>
)

export const CollapsedSidebar = () => (
  <DashboardShellProvider initialDesktopExpanded={false}>
    <DashboardHeader
      leading={<Text weight="semibold">Dashboard</Text>}
      userMenu={
        <UserDropdown
          user={USER}
          signOutHref="https://auth.example.com/logout"
          compact
        />
      }
    />
  </DashboardShellProvider>
)
