/* eslint-disable react/no-unescaped-entities */
import { Link } from './Link'
import { Icon } from './Icon'
import { Text } from './Text'

export const Variants = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Link Variants</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        The{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          variant
        </code>{' '}
        prop controls the visual styling and behavior of links. Each variant is
        optimized for specific use cases and contexts within the application.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">All Link Variants</h4>
      <div className="space-y-3 p-4 border rounded-lg">
        <Link href="#" variant="default">
          Default link variant
        </Link>
        <Link href="#" variant="ghost">
          Ghost link variant
        </Link>
        <Link href="#" variant="nav">
          Nav link variant
        </Link>
        <Link href="#" variant="breadcrumb">
          Breadcrumb link variant
        </Link>
      </div>
    </div>

    <div className="grid grid-cols-1 md:grid-cols-2 gap-3 text-sm mt-6">
      <div>
        <strong>default:</strong> Standard link styling with primary colors and
        hover effects
      </div>
      <div>
        <strong>ghost:</strong> Subtle button-like styling with background hover
        effects
      </div>
      <div>
        <strong>nav:</strong> Navigation link styling optimized for sidebars and
        menus
      </div>
      <div>
        <strong>breadcrumb:</strong> Minimal styling for breadcrumb navigation
        chains
      </div>
    </div>
  </div>
)

export const States = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Link States</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Links support active states using the{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          isActive
        </code>{' '}
        prop. Active states are particularly important for navigation links to
        show the current page or section.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Navigation Links</h4>
      <div className="space-y-2 p-4 border rounded-lg bg-gray-50 dark:bg-gray-800">
        <Link href="#" variant="nav">
          Dashboard
        </Link>
        <Link href="#" variant="nav" isActive>
          Applications (Active)
        </Link>
        <Link href="#" variant="nav">
          Settings
        </Link>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Breadcrumb Links</h4>
      <div className="flex items-center gap-2 p-4 border rounded-lg">
        <Link href="#" variant="breadcrumb">
          Home
        </Link>
        <span className="text-gray-400">/</span>
        <Link href="#" variant="breadcrumb">
          Projects
        </Link>
        <span className="text-gray-400">/</span>
        <Link href="#" variant="breadcrumb" isActive>
          My App
        </Link>
      </div>
    </div>
  </div>
)

export const ExternalLinks = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">External Links</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Use the{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          isExternal
        </code>{' '}
        prop for links that navigate to external websites. External links
        automatically include proper security attributes and open in a new tab.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">External Link Examples</h4>
      <div className="space-y-3 p-4 border rounded-lg">
        <Link href="https://docs.nuon.co" isExternal>
          <Icon variant="ArrowSquareOut" size="16" />
          Documentation
        </Link>
        <Link href="https://github.com/nuonplatform" isExternal variant="ghost">
          <Icon variant="GithubLogo" size="16" />
          GitHub Repository
        </Link>
        <Link href="https://nuon.co/support" isExternal>
          Support Center
          <Icon variant="ArrowSquareOut" size="16" />
        </Link>
      </div>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>External Link Features:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Automatically adds target="_blank" for new tab opening</li>
        <li>Includes rel="noopener noreferrer" for security</li>
        <li>Works with all link variants and styling options</li>
        <li>Consider adding visual indicators (icons) for clarity</li>
      </ul>
    </div>
  </div>
)

export const WithIcons = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Links with Icons</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Links automatically detect and properly space icon components. Icons can
        be placed before or after the link text to enhance meaning and visual
        appeal.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Icon Positioning</h4>
      <div className="space-y-3 p-4 border rounded-lg">
        <Link href="#" variant="default">
          <Icon variant="House" size="16" />
          Home Page
        </Link>
        <Link href="#" variant="ghost">
          <Icon variant="Gear" size="16" />
          Settings
        </Link>
        <Link href="#" variant="nav">
          <Icon variant="ChartBar" size="16" />
          Analytics Dashboard
        </Link>
        <Link href="#" variant="default">
          View Profile
          <Icon variant="ArrowRight" size="16" />
        </Link>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Navigation with Icons</h4>
      <div className="space-y-2 p-4 border rounded-lg bg-gray-50 dark:bg-gray-800">
        <Link href="#" variant="nav">
          <Icon variant="SquaresFour" size="18" />
          Dashboard
        </Link>
        <Link href="#" variant="nav" isActive>
          <Icon variant="Stack" size="18" />
          Applications
        </Link>
        <Link href="#" variant="nav">
          <Icon variant="Users" size="18" />
          Team
        </Link>
        <Link href="#" variant="nav">
          <Icon variant="Gear" size="18" />
          Settings
        </Link>
      </div>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Icon Guidelines:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Use 16px icons for default and ghost variants</li>
        <li>Use 18px icons for nav variants to maintain visual balance</li>
        <li>Place functional icons before text, directional icons after</li>
        <li>Ensure icons are semantically meaningful to the link purpose</li>
      </ul>
    </div>
  </div>
)

export const UsagePatterns = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Common Usage Patterns</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Links are used throughout the application in various contexts. Here are
        common patterns and recommended approaches for different scenarios.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Inline Text Links</h4>
      <div className="p-4 border rounded-lg">
        <Text>
          Welcome to the platform! Check out our{' '}
          <Link href="/docs">documentation</Link> to get started, or visit the{' '}
          <Link href="/examples">examples page</Link> for inspiration. For
          additional help, see our{' '}
          <Link href="https://nuon.co/support" isExternal>
            support center
          </Link>
          .
        </Text>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Action Links</h4>
      <div className="flex flex-wrap gap-3 p-4 border rounded-lg">
        <Link href="/create" variant="ghost">
          <Icon variant="Plus" size="16" />
          Create New
        </Link>
        <Link href="/import" variant="ghost">
          <Icon variant="Upload" size="16" />
          Import Data
        </Link>
        <Link href="/export" variant="ghost">
          <Icon variant="Download" size="16" />
          Export
        </Link>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Breadcrumb Navigation</h4>
      <div className="flex items-center gap-2 p-4 border rounded-lg">
        <Link href="/org" variant="breadcrumb">
          Organization
        </Link>
        <Icon variant="CaretRight" size="12" className="text-gray-400" />
        <Link href="/org/apps" variant="breadcrumb">
          Applications
        </Link>
        <Icon variant="CaretRight" size="12" className="text-gray-400" />
        <Link href="/org/apps/my-app" variant="breadcrumb" isActive>
          My Application
        </Link>
      </div>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Best Practices:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Use default variant for inline text links</li>
        <li>Use ghost variant for action-oriented links</li>
        <li>Use nav variant for navigation menus and sidebars</li>
        <li>Use breadcrumb variant for navigation breadcrumbs</li>
        <li>
          Always provide meaningful link text that describes the destination
        </li>
      </ul>
    </div>
  </div>
)
