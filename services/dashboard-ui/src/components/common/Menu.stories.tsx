/* eslint-disable react/no-unescaped-entities */
import { Menu } from './Menu'
import { Button } from './Button'
import { Icon } from './Icon'
import { Link } from './Link'
import { Text } from './Text'
import { Status } from './Status'
import { Badge } from './Badge'

export const BasicUsage = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Basic Menu Usage</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Menus provide a consistent container for organizing related actions,
        links, and content. They're commonly used within dropdowns, popovers,
        and other contextual interfaces. Content is automatically styled and
        spaced appropriately.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Simple Menu</h4>
      <Menu className="w-56">
        <Button variant="ghost">Edit Item</Button>
        <Button variant="ghost">Duplicate</Button>
        <Link href="#" className="px-3 py-2">
          View Details
        </Link>
        <hr />
        <Text variant="label" theme="neutral">
          More Actions
        </Text>
        <Button variant="ghost">Share</Button>
        <Button variant="ghost">Export</Button>
      </Menu>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Menu Features:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Automatic styling and spacing for menu items</li>
        <li>Support for buttons, links, text, and custom content</li>
        <li>Horizontal dividers (hr) for visual grouping</li>
        <li>Consistent padding and hover states</li>
      </ul>
    </div>
  </div>
)

export const ComplexMenu = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Complex Menu Example</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Menus can contain various types of content including grouped sections,
        status indicators, badges, and differently styled actions. This
        demonstrates a realistic menu for managing deployment settings.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Deployment Actions Menu</h4>
      <Menu className="w-64">
        <Text variant="label" theme="neutral">
          Quick Actions
        </Text>
        <Button variant="ghost">
          <Icon variant="PencilSimpleLine" size="16" />
          Edit Configuration
        </Button>
        <Button variant="ghost">
          <Icon variant="ListChecks" size="16" />
          Auto Approve Changes
        </Button>
        <Button variant="ghost">
          <Icon variant="CodeBlock" size="16" />
          View Current State
        </Button>
        <hr />
        <Text variant="label" theme="neutral">
          Deployment Controls
        </Text>
        <Button variant="ghost">
          <Icon variant="ArrowUp" size="16" />
          Reprovision Install
        </Button>
        <Button variant="ghost">
          <Icon variant="ArrowDown" size="16" />
          Deprovision Install
        </Button>
        <Button variant="ghost">
          <Icon variant="StackMinus" size="16" />
          Remove Stack
        </Button>
        <hr />
        <div className="px-3 py-2 flex items-center justify-between">
          <Text variant="subtext">Status</Text>
          <Status status="success" />
        </div>
        <div className="px-3 py-2 flex items-center justify-between">
          <Text variant="subtext">Environment</Text>
          <Badge theme="info" size="sm">
            Production
          </Badge>
        </div>
        <hr />
        <Text variant="label" theme="neutral">
          Danger Zone
        </Text>
        <Button variant="ghost">
          <Icon variant="Key" size="16" />
          Break Glass Access
        </Button>
        <Button variant="ghost" className="!text-red-600 dark:!text-red-400">
          <Icon variant="Trash" size="16" />
          Delete Installation
        </Button>
      </Menu>
    </div>
  </div>
)

export const MenuSections = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Menu Sections and Grouping</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Use text labels and horizontal dividers to organize menu items into
        logical groups. This improves usability and makes it easier for users
        to find specific actions.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">User Account Menu</h4>
      <Menu className="w-52">
        <div className="px-3 py-2 border-b">
          <Text weight="strong" variant="subtext">
            John Doe
          </Text>
          <Text variant="label" theme="neutral">
            john@example.com
          </Text>
        </div>
        <Text variant="label" theme="neutral">
          Account
        </Text>
        <Button variant="ghost">
          <Icon variant="User" size="16" />
          Profile Settings
        </Button>
        <Button variant="ghost">
          <Icon variant="Bell" size="16" />
          Notifications
        </Button>
        <Button variant="ghost">
          <Icon variant="CreditCard" size="16" />
          Billing
        </Button>
        <hr />
        <Text variant="label" theme="neutral">
          Organization
        </Text>
        <Button variant="ghost">
          <Icon variant="Users" size="16" />
          Team Members
        </Button>
        <Button variant="ghost">
          <Icon variant="Gear" size="16" />
          Organization Settings
        </Button>
        <hr />
        <Text variant="label" theme="neutral">
          Help & Support
        </Text>
        <Link href="#" className="px-3 py-2">
          <Icon variant="Book" size="16" />
          Documentation
        </Link>
        <Link href="#" className="px-3 py-2">
          <Icon variant="ChatCircle" size="16" />
          Contact Support
        </Link>
        <hr />
        <Button variant="ghost" className="!text-red-600 dark:!text-red-400">
          <Icon variant="SignOut" size="16" />
          Sign Out
        </Button>
      </Menu>
    </div>
  </div>
)

export const MenuWithStatus = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Menus with Status Information</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Menus can display status information, metrics, and other contextual
        data alongside actionable items. This is useful for system dashboards
        and monitoring interfaces.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">System Status Menu</h4>
      <Menu className="w-56">
        <Text variant="label" theme="neutral">
          System Health
        </Text>
        <div className="px-3 py-2 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Status status="success" isWithoutText />
            <Text variant="subtext">Database</Text>
          </div>
          <Text variant="label" theme="neutral">
            2ms
          </Text>
        </div>
        <div className="px-3 py-2 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Status status="success" isWithoutText />
            <Text variant="subtext">API Gateway</Text>
          </div>
          <Text variant="label" theme="neutral">
            15ms
          </Text>
        </div>
        <div className="px-3 py-2 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Status status="warn" isWithoutText />
            <Text variant="subtext">Cache Server</Text>
          </div>
          <Text variant="label" theme="neutral">
            120ms
          </Text>
        </div>
        <hr />
        <Text variant="label" theme="neutral">
          Quick Actions
        </Text>
        <Button variant="ghost">
          <Icon variant="ArrowClockwise" size="16" />
          Refresh Status
        </Button>
        <Link href="#" className="px-3 py-2">
          <Icon variant="ChartLine" size="16" />
          View Detailed Metrics
        </Link>
        <hr />
        <div className="px-3 py-2">
          <Text variant="label" theme="neutral">
            Last Updated: 2 minutes ago
          </Text>
        </div>
      </Menu>
    </div>
  </div>
)

export const CustomStyling = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Menu Customization</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Menus can be customized with different widths, background colors, and
        styling using standard CSS classes. The menu component provides a
        consistent base that can be adapted to various design needs.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Different Menu Styles</h4>
      <div className="flex gap-6">
        <div>
          <Text variant="label" className="mb-2 block">
            Compact Menu
          </Text>
          <Menu className="w-40">
            <Button variant="ghost" size="sm">
              <Icon variant="Eye" size="14" />
              View
            </Button>
            <Button variant="ghost" size="sm">
              <Icon variant="PencilSimple" size="14" />
              Edit
            </Button>
            <Button variant="ghost" size="sm">
              <Icon variant="Trash" size="14" />
              Delete
            </Button>
          </Menu>
        </div>
        <div>
          <Text variant="label" className="mb-2 block">
            Wide Menu
          </Text>
          <Menu className="w-72">
            <Text variant="label" theme="neutral">
              Recent Documents
            </Text>
            <Button variant="ghost" className="justify-between">
              <span>Project Proposal.pdf</span>
              <Text variant="label" theme="neutral">
                2 hours ago
              </Text>
            </Button>
            <Button variant="ghost" className="justify-between">
              <span>Design System Guide.figma</span>
              <Text variant="label" theme="neutral">
                1 day ago
              </Text>
            </Button>
            <hr />
            <Link href="#" className="px-3 py-2">
              <Icon variant="FolderOpen" size="16" />
              Browse All Documents
            </Link>
          </Menu>
        </div>
      </div>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Customization Guidelines:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Use appropriate widths based on content length</li>
        <li>Maintain consistent padding and spacing</li>
        <li>Group related actions with dividers and labels</li>
        <li>Consider menu item hierarchy and visual weight</li>
      </ul>
    </div>
  </div>
)
