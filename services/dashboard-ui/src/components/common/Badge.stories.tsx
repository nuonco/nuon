/* eslint-disable react/no-unescaped-entities */
import { Badge } from './Badge'
import { Text } from './Text'
import { Icon } from './Icon'

export const BasicUsage = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Basic Badge Usage</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Badges are compact elements used to display metadata, status, or
        categorization information. They feature subtle backgrounds, borders,
        and typography designed to integrate seamlessly with the interface
        while providing clear visual hierarchy.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Default Badge</h4>
      <div className="flex items-center gap-4">
        <Badge>Default Badge</Badge>
        <Text variant="subtext">Sans-serif typography, fully rounded corners</Text>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">With Icons</h4>
      <div className="flex flex-wrap gap-3">
        <Badge theme="success">
          <Icon variant="CheckCircle" size="12" />
          Verified
        </Badge>
        <Badge theme="error">
          <Icon variant="WarningCircle" size="12" />
          Failed
        </Badge>
        <Badge theme="info">
          <Icon variant="Info" size="12" />
          Beta
        </Badge>
        <Badge theme="warn">
          <Icon variant="Warning" size="12" />
          Deprecated
        </Badge>
      </div>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Key Features:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Automatic flexbox layout with proper gap spacing</li>
        <li>Shrink and grow control for consistent sizing</li>
        <li>Dark mode support with appropriate contrast ratios</li>
        <li>Icon integration with automatic alignment</li>
      </ul>
    </div>
  </div>
)

export const Themes = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Badge Themes</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        The{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          theme
        </code>{' '}
        prop controls the color scheme of the badge. Each theme includes proper
        dark mode styling and maintains accessibility contrast ratios. All themes
        use semantic colors that align with their intended meaning.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">All Available Themes</h4>
      <div className="flex flex-wrap gap-4 items-center">
        <Badge theme="brand">Brand</Badge>
        <Badge theme="error">Error</Badge>
        <Badge theme="warn">Warn</Badge>
        <Badge theme="info">Info</Badge>
        <Badge theme="success">Success</Badge>
        <Badge theme="neutral">Neutral</Badge>
        <Badge theme="default">Default</Badge>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-3 text-sm mt-6">
        <div>
          <strong>brand:</strong> Purple primary colors for Nuon platform
          branding and product features
        </div>
        <div>
          <strong>error:</strong> Red colors for error states, failures, and critical
          issues requiring attention
        </div>
        <div>
          <strong>warn:</strong> Orange colors for warnings, cautions, and
          situations needing review
        </div>
        <div>
          <strong>info:</strong> Blue colors for informational content,
          tips, and neutral updates
        </div>
        <div>
          <strong>success:</strong> Green colors for successful operations,
          completed states, and positive feedback
        </div>
        <div>
          <strong>neutral:</strong> Cool grey colors for neutral information
          and secondary content
        </div>
        <div>
          <strong>default:</strong> Standard grey colors used when no theme
          is specified (default)
        </div>
      </div>
    </div>
  </div>
)

export const Sizes = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Badge Sizes</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        The{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          size
        </code>{' '}
        prop controls the dimensions and typography of the badge. All sizes use
        -0.2px letter spacing for improved readability and consistent visual
        weight across different contexts.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Size Comparison</h4>
      <div className="flex gap-6 items-center">
        <div className="text-center space-y-2">
          <Badge size="sm" theme="brand">Small</Badge>
          <Text variant="label" className="text-xs">SM</Text>
        </div>
        <div className="text-center space-y-2">
          <Badge size="md" theme="brand">Medium</Badge>
          <Text variant="label" className="text-xs">MD</Text>
        </div>
        <div className="text-center space-y-2">
          <Badge size="lg" theme="brand">Large</Badge>
          <Text variant="label" className="text-xs">LG (Default)</Text>
        </div>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Sizes with Icons</h4>
      <div className="flex gap-4 items-center">
        <Badge size="sm" theme="success">
          <Icon variant="CheckCircle" size="10" />
          Done
        </Badge>
        <Badge size="md" theme="info">
          <Icon variant="Info" size="11" />
          Notice
        </Badge>
        <Badge size="lg" theme="error">
          <Icon variant="WarningCircle" size="12" />
          Error
        </Badge>
      </div>
    </div>

    <div className="grid grid-cols-1 md:grid-cols-3 gap-3 text-sm mt-6">
      <div>
        <strong>sm:</strong> 11px text, 14px line height, 8px/2px padding.
        Ideal for dense interfaces and compact layouts.
      </div>
      <div>
        <strong>md:</strong> 12px text, 17px line height, 8px/2px padding.
        Good balance for moderate density interfaces.
      </div>
      <div>
        <strong>lg:</strong> 12px text, 17px line height, 12px/4px padding
        (default). Best for standard layouts and readability.
      </div>
    </div>
  </div>
)

export const Variants = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Badge Variants</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        The{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          variant
        </code>{' '}
        prop controls the visual style and typography of the badge. Each variant
        is optimized for different types of content and contexts.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Variant Comparison</h4>
      <div className="space-y-3">
        <div className="flex items-center gap-4">
          <Badge variant="default" theme="brand">User Interface</Badge>
          <Text variant="subtext">Default: Sans-serif, fully rounded</Text>
        </div>
        <div className="flex items-center gap-4">
          <Badge variant="code" theme="brand">v2.1.4</Badge>
          <Text variant="subtext">Code: Monospace, moderately rounded</Text>
        </div>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Code Variant Examples</h4>
      <div className="flex flex-wrap gap-3">
        <Badge variant="code" theme="neutral">v1.0.0</Badge>
        <Badge variant="code" theme="info">main</Badge>
        <Badge variant="code" theme="success">200 OK</Badge>
        <Badge variant="code" theme="error">404</Badge>
        <Badge variant="code" theme="warn">deprecated</Badge>
        <Badge variant="code" theme="brand">/api/v1/users</Badge>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Default Variant Examples</h4>
      <div className="flex flex-wrap gap-3">
        <Badge variant="default" theme="success">
          <Icon variant="CheckCircle" size="12" />
          Verified
        </Badge>
        <Badge variant="default" theme="brand">Premium</Badge>
        <Badge variant="default" theme="info">New</Badge>
        <Badge variant="default" theme="warn">Limited</Badge>
        <Badge variant="default" theme="error">Urgent</Badge>
      </div>
    </div>

    <div className="grid grid-cols-1 md:grid-cols-2 gap-3 text-sm mt-6">
      <div>
        <strong>default:</strong> Sans-serif font with fully rounded corners.
        Best for general labels, status indicators, and UI categories.
      </div>
      <div>
        <strong>code:</strong> Monospace font with moderately rounded corners.
        Perfect for version numbers, API endpoints, and technical identifiers.
      </div>
    </div>
  </div>
)

export const UsageExamples = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Common Usage Patterns</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Badges are versatile components used throughout the interface for
        categorization, status indication, and metadata display. Here are
        common patterns and recommended approaches.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Status Indicators</h4>
      <div className="space-y-3">
        <div className="flex items-center justify-between p-3 border rounded">
          <div className="flex items-center gap-3">
            <Text weight="strong">Production Deployment</Text>
            <Badge theme="success" size="sm">
              <Icon variant="CheckCircle" size="10" />
              Live
            </Badge>
          </div>
          <Text variant="subtext" theme="neutral">2 hours ago</Text>
        </div>
        <div className="flex items-center justify-between p-3 border rounded">
          <div className="flex items-center gap-3">
            <Text weight="strong">Staging Environment</Text>
            <Badge theme="warn" size="sm">
              <Icon variant="Warning" size="10" />
              Maintenance
            </Badge>
          </div>
          <Text variant="subtext" theme="neutral">5 minutes ago</Text>
        </div>
        <div className="flex items-center justify-between p-3 border rounded">
          <div className="flex items-center gap-3">
            <Text weight="strong">Development Build</Text>
            <Badge theme="error" size="sm">
              <Icon variant="WarningCircle" size="10" />
              Failed
            </Badge>
          </div>
          <Text variant="subtext" theme="neutral">1 hour ago</Text>
        </div>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Version and Technical Info</h4>
      <div className="p-4 border rounded-lg space-y-3">
        <div className="flex items-center justify-between">
          <Text weight="strong">API Gateway Service</Text>
          <Badge variant="code" theme="brand" size="sm">v2.1.4</Badge>
        </div>
        <div className="space-y-2">
          <div className="flex items-center gap-2">
            <Text variant="subtext">Endpoint:</Text>
            <Badge variant="code" theme="neutral" size="sm">/api/v1/auth</Badge>
          </div>
          <div className="flex items-center gap-2">
            <Text variant="subtext">Branch:</Text>
            <Badge variant="code" theme="info" size="sm">main</Badge>
          </div>
          <div className="flex items-center gap-2">
            <Text variant="subtext">Status:</Text>
            <Badge variant="code" theme="success" size="sm">200 OK</Badge>
          </div>
        </div>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Feature and Category Labels</h4>
      <div className="space-y-3">
        <div className="p-3 border rounded">
          <div className="flex items-center justify-between mb-2">
            <Text weight="strong">Advanced Analytics Dashboard</Text>
            <div className="flex gap-2">
              <Badge theme="brand" size="sm">Premium</Badge>
              <Badge theme="info" size="sm">New</Badge>
            </div>
          </div>
          <Text variant="subtext" theme="neutral">
            Comprehensive analytics with real-time data visualization
          </Text>
        </div>
        <div className="p-3 border rounded">
          <div className="flex items-center justify-between mb-2">
            <Text weight="strong">Legacy API Integration</Text>
            <div className="flex gap-2">
              <Badge theme="warn" size="sm">Deprecated</Badge>
              <Badge theme="neutral" size="sm">Maintenance</Badge>
            </div>
          </div>
          <Text variant="subtext" theme="neutral">
            Scheduled for replacement in Q2 2024
          </Text>
        </div>
      </div>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Best Practices:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Use semantic themes that match the content meaning</li>
        <li>Choose appropriate sizes based on visual hierarchy and space</li>
        <li>Combine with icons for better visual communication</li>
        <li>Use the code variant for technical identifiers and values</li>
        <li>Keep badge text concise and scannable</li>
        <li>Ensure proper contrast in both light and dark modes</li>
      </ul>
    </div>
  </div>
)
