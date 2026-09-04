import { ComponentDocs } from '../__stories__/ComponentDocs'
import { Link } from '../atoms/Link'
import { UserDropdown } from './UserDropdown'
import { FocusHeader } from './FocusHeader'

export default {
  title: 'lite/organisms/FocusHeader',
}

const Actions = () => (
  <>
    <Link href="https://docs.nuon.co" external variant="caption">
      Developer docs
    </Link>
    <UserDropdown
      user={{ name: 'Alex Morgan', email: 'alex@example.com' }}
      signOutHref="https://auth.example.com/logout"
    />
  </>
)

export const Overview = () => (
  <ComponentDocs
    name="FocusHeader"
    tier="organism"
    summary="Full-width global chrome for sidebar-free focused flows."
    use={[
      'Use as the sticky header inside the FocusShell scroll region.',
      'Place global documentation and account controls in actions.',
    ]}
    avoid={[
      'Do not place wizard progress, navigation, or skip actions here.',
      'Do not add organization status or dashboard navigation.',
    ]}
    rules={[
      'The Nuon logo always links to a safe application root.',
      'The header keeps a stable height across route changes.',
      'The glass surface and elevation appear only after content scrolls.',
      'Actions remain aligned at narrow widths.',
    ]}
    props={[
      {
        name: 'actions',
        type: 'ReactNode',
        description: 'Global controls aligned opposite the Nuon logo.',
      },
      {
        name: 'homeHref',
        type: 'string',
        default: '/',
        description: 'Safe destination for the Nuon logo.',
      },
    ]}
  />
)

export const Default = () => <FocusHeader actions={<Actions />} />

export const EmptyActions = () => <FocusHeader />

export const LongActions = () => (
  <FocusHeader
    actions={
      <Link
        href="https://docs.nuon.co/getting-started"
        external
        variant="caption"
      >
        Read the developer documentation
      </Link>
    }
  />
)
