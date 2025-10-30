/* eslint-disable react/no-unescaped-entities */
import { Card } from './Card'
import { Text } from './Text'
import { Button } from './Button'
import { Badge } from './Badge'
import { Icon } from './Icon'

export const BasicUsage = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Basic Card Usage</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Cards are flexible containers that group related content together. They
        provide a consistent visual structure with built-in spacing, border, and
        shadow styling that adapts to both light and dark themes.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Simple Card</h4>
      <Card>
        <Text>This is a basic card with default styling.</Text>
      </Card>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Card with Multiple Elements</h4>
      <Card>
        <Text variant="h3" weight="stronger">
          Card Title
        </Text>
        <Text>
          This card contains multiple text elements, demonstrating how the
          built-in gap spacing works between child elements.
        </Text>
        <Text theme="neutral" variant="subtext">
          Additional secondary information can be included.
        </Text>
      </Card>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Default Styling:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Flex column layout with 6px gap between children</li>
        <li>24px padding on all sides</li>
        <li>Border with rounded corners and subtle shadow</li>
        <li>Automatic dark mode adaptation</li>
      </ul>
    </div>
  </div>
)

export const ContentTypes = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Card Content Types</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Cards can contain various types of content including text, buttons,
        badges, icons, and other components. The built-in spacing automatically
        handles layout between different content types.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Information Card</h4>
      <Card>
        <div className="flex items-center gap-2">
          <Icon variant="Info" size="20" />
          <Text variant="h3" weight="stronger">
            System Status
          </Text>
        </div>
        <Text>All systems are operating normally.</Text>
        <div className="flex gap-2">
          <Badge theme="success">Operational</Badge>
          <Badge theme="info">Last updated: 2 minutes ago</Badge>
        </div>
      </Card>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Action Card</h4>
      <Card>
        <Text variant="h3" weight="stronger">
          Deploy Application
        </Text>
        <Text>
          Ready to deploy your application to production. This will create a new
          release and update your live environment.
        </Text>
        <div className="flex gap-2">
          <Button variant="primary">Deploy Now</Button>
          <Button variant="secondary">Schedule Later</Button>
        </div>
      </Card>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Data Card</h4>
      <Card>
        <Text variant="h3" weight="stronger">
          Performance Metrics
        </Text>
        <div className="grid grid-cols-2 gap-4">
          <div>
            <Text variant="label" theme="neutral">
              Response Time
            </Text>
            <Text variant="h2" weight="stronger">
              245ms
            </Text>
          </div>
          <div>
            <Text variant="label" theme="neutral">
              Success Rate
            </Text>
            <Text variant="h2" weight="stronger" theme="success">
              99.8%
            </Text>
          </div>
        </div>
      </Card>
    </div>
  </div>
)

export const CustomStyling = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Custom Card Styling</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Cards accept a{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          className
        </code>{' '}
        prop to customize styling while maintaining the base card structure.
        Common customizations include background colors, borders, and spacing
        adjustments.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Background Variants</h4>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <Card className="bg-blue-50 dark:bg-blue-950 border-blue-200 dark:border-blue-800">
          <Text variant="h3" weight="stronger" theme="info">
            Info Card
          </Text>
          <Text>Card with blue background theme.</Text>
        </Card>
        <Card className="bg-green-50 dark:bg-green-950 border-green-200 dark:border-green-800">
          <Text variant="h3" weight="stronger" theme="success">
            Success Card
          </Text>
          <Text>Card with green background theme.</Text>
        </Card>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Shadow and Border Variants</h4>
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Card className="shadow-none border-2">
          <Text variant="base" weight="strong">
            Strong Border
          </Text>
          <Text variant="subtext">No shadow, thick border</Text>
        </Card>
        <Card className="shadow-lg">
          <Text variant="base" weight="strong">
            Large Shadow
          </Text>
          <Text variant="subtext">Enhanced shadow depth</Text>
        </Card>
        <Card className="border-0 bg-gray-100 dark:bg-gray-800">
          <Text variant="base" weight="strong">
            No Border
          </Text>
          <Text variant="subtext">Background color only</Text>
        </Card>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Spacing Customization</h4>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <Card className="p-3 gap-3">
          <Text variant="base" weight="strong">
            Compact Card
          </Text>
          <Text variant="subtext">Reduced padding and gap</Text>
        </Card>
        <Card className="p-8 gap-8">
          <Text variant="base" weight="strong">
            Spacious Card
          </Text>
          <Text variant="subtext">Increased padding and gap</Text>
        </Card>
      </div>
    </div>
  </div>
)

export const LayoutPatterns = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Card Layout Patterns</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Cards work well in various layout patterns. Here are common ways to
        arrange and combine cards in your application interfaces.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Card Grid</h4>
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Card>
          <Icon variant="Database" size="24" />
          <Text variant="base" weight="strong">
            Database
          </Text>
          <Text variant="subtext" theme="neutral">
            Manage your data
          </Text>
        </Card>
        <Card>
          <Icon variant="Gear" size="24" />
          <Text variant="base" weight="strong">
            Settings
          </Text>
          <Text variant="subtext" theme="neutral">
            Configure options
          </Text>
        </Card>
        <Card>
          <Icon variant="ChartBar" size="24" />
          <Text variant="base" weight="strong">
            Analytics
          </Text>
          <Text variant="subtext" theme="neutral">
            View insights
          </Text>
        </Card>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Sidebar Cards</h4>
      <div className="flex gap-6">
        <div className="flex-1">
          <Card>
            <Text variant="h3" weight="stronger">
              Main Content Area
            </Text>
            <Text>
              This represents the primary content area of a page layout. Cards
              can be used to structure both main content and supplementary
              information.
            </Text>
            <Text>
              The flexible nature of cards makes them suitable for various
              content types while maintaining visual consistency across your
              application.
            </Text>
          </Card>
        </div>
        <div className="w-64 space-y-4">
          <Card>
            <Text variant="base" weight="strong">
              Quick Actions
            </Text>
            <div className="space-y-2">
              <Button variant="ghost" size="sm" className="w-full justify-start">
                <Icon variant="Plus" size="16" />
                Create New
              </Button>
              <Button variant="ghost" size="sm" className="w-full justify-start">
                <Icon variant="Upload" size="16" />
                Import Data
              </Button>
            </div>
          </Card>
          <Card>
            <Text variant="base" weight="strong">
              Recent Activity
            </Text>
            <Text variant="subtext" theme="neutral">
              No recent activity
            </Text>
          </Card>
        </div>
      </div>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Layout Best Practices:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Use consistent gap spacing between cards in grids and lists</li>
        <li>Consider responsive breakpoints for card grid layouts</li>
        <li>Group related functionality within single cards</li>
        <li>Maintain visual hierarchy with card sizing and content structure</li>
      </ul>
    </div>
  </div>
)
