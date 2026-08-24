export default {
  title: 'Common/Link',
}

import { Link } from './Link'
import { Icon } from './Icon'
import { Text } from './Text'

export const EntityLink = () => (
  <div className="flex flex-col gap-4">
    <Text variant="subtext">
      Entity link: the resource's own name is the link text. No verb, no icon —
      the name navigates.
    </Text>
    <div className="flex flex-col gap-3 p-4 border rounded-lg">
      <Link href="#">acme-payments</Link>
      <Link href="#" className="font-mono">
        <Icon variant="GitBranchIcon" size="1em" />
        main
      </Link>
    </div>
  </div>
)

export const ViewLink = () => (
  <div className="flex flex-col gap-4">
    <Text variant="subtext">
      View link: a standalone "go see more" link. Sentence case, no icon. Wrap in
      a subtext Text so it inherits a small size.
    </Text>
    <div className="flex flex-col gap-3 p-4 border rounded-lg">
      <Text variant="subtext">
        <Link href="#">View plan</Link>
      </Text>
      <Text variant="subtext">
        <Link href="#">View all runs</Link>
      </Text>
      <Text variant="subtext">
        <Link href="#">View details</Link>
      </Text>
    </div>
  </div>
)

export const ExternalLink = () => (
  <div className="flex flex-col gap-4">
    <Text variant="subtext">
      External link: anything leaving the app. Set <code>isExternal</code> — the
      new-tab icon renders automatically, sized to the surrounding text. Never
      hand-place the icon.
    </Text>
    <div className="flex flex-col gap-3 p-4 border rounded-lg">
      <Link href="https://docs.nuon.co" isExternal>
        View documentation
      </Link>
      <Link href="https://github.com/acme/payments" isExternal className="font-mono">
        <Icon variant="GitHub" size="1em" />
        acme/payments
      </Link>
      <Text>
        Read more in the{' '}
        <Link href="https://docs.nuon.co" isExternal>
          docs
        </Link>
        .
      </Text>
    </div>
  </div>
)

export const InlineLink = () => (
  <div className="flex flex-col gap-4">
    <Text variant="subtext">
      Inline links inherit the surrounding text size and color — never set their
      own size.
    </Text>
    <div className="p-4 border rounded-lg">
      <Text>
        Welcome to the platform. Check out the{' '}
        <Link href="/docs">documentation</Link> to get started, or browse the{' '}
        <Link href="/examples">examples</Link>. For help, see the{' '}
        <Link href="https://nuon.co/support" isExternal>
          support center
        </Link>
        .
      </Text>
    </div>
  </div>
)

export const NavAndBreadcrumb = () => (
  <div className="flex flex-col gap-6">
    <Text variant="subtext">
      Nav-chrome variants (out of scope for the content-link taxonomy) — sidebar
      nav links and breadcrumbs.
    </Text>
    <div className="flex flex-col gap-2 p-4 border rounded-lg bg-gray-50 dark:bg-gray-800">
      <Link href="#" variant="nav">
        <Icon variant="SquaresFourIcon" size="18" />
        Dashboard
      </Link>
      <Link href="#" variant="nav" isActive>
        <Icon variant="StackIcon" size="18" />
        Applications
      </Link>
      <Link href="#" variant="nav">
        <Icon variant="GearIcon" size="18" />
        Settings
      </Link>
    </div>
    <div className="flex items-center gap-2 p-4 border rounded-lg">
      <Link href="#" variant="breadcrumb">
        Organization
      </Link>
      <Icon variant="CaretRightIcon" size="12" className="text-gray-400" />
      <Link href="#" variant="breadcrumb">
        Applications
      </Link>
      <Icon variant="CaretRightIcon" size="12" className="text-gray-400" />
      <Link href="#" variant="breadcrumb" isActive>
        acme-payments
      </Link>
    </div>
  </div>
)
