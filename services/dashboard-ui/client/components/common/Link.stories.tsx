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
      the name navigates. Inside a sized context (table cell, sentence), use
      variant="inline" so it inherits.
    </Text>
    <div className="flex flex-col gap-3 p-4 border rounded-lg">
      <Link href="#">acme-payments</Link>
      <Link href="#" className="font-mono">
        <Icon variant="GitBranchIcon" size="1em" />
        main
      </Link>
      <Text variant="body">
        In a body-sized row: <Link href="#" variant="inline">acme-payments</Link>
      </Text>
    </div>
  </div>
)

export const ViewLink = () => (
  <div className="flex flex-col gap-4">
    <Text variant="subtext">
      View link: a standalone "go see more" link. Sentence case, no icon. The
      default Link renders at subtext size on its own — no Text wrapper needed;
      use textVariant to explicitly size up.
    </Text>
    <div className="flex flex-col gap-3 p-4 border rounded-lg">
      <Link href="#">View plan</Link>
      <Link href="#">View all runs</Link>
      <Link href="#">View details</Link>
      <Link href="#" textVariant="base">
        View details (textVariant="base")
      </Link>
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
      variant="inline" inherits the surrounding text size — for links inside
      sentences, table cells, and other sized contexts.
    </Text>
    <div className="p-4 border rounded-lg">
      <Text>
        Welcome to the platform. Check out the{' '}
        <Link href="/docs" variant="inline">documentation</Link> to get started, or browse the{' '}
        <Link href="/examples" variant="inline">examples</Link>. For help, see the{' '}
        <Link href="https://nuon.co/support" isExternal variant="inline">
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
