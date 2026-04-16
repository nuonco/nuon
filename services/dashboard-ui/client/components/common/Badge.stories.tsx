export default {
  title: 'Common/Badge',
}

import { Badge } from './Badge'
import { Text } from './Text'
import { Icon } from './Icon'

export const BadgeNeutral = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Badge neutral</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Non-semantic badges for labels, categories, and metadata. Auto-sizes to
        12px / px-3 py-1.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Default variant</h4>
      <div className="flex flex-wrap gap-3 items-center">
        <Badge>Neutral</Badge>
        <Badge theme="default">Default</Badge>
        <Badge theme="brand">Brand</Badge>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Code variant</h4>
      <div className="flex flex-wrap gap-3 items-center">
        <Badge variant="code">v1.0.0</Badge>
        <Badge variant="code">/api/v1/users</Badge>
        <Badge variant="code" theme="brand">main</Badge>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">With icons</h4>
      <div className="flex flex-wrap gap-3 items-center">
        <Badge>
          <Icon variant="Tag" size="12" />
          Component
        </Badge>
        <Badge variant="code">
          <Icon variant="GitBranch" size="12" />
          feature/auth
        </Badge>
      </div>
    </div>
  </div>
)

export const BadgeStatus = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Badge status</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Semantic status badges. Auto-sizes to 11px / px-2 py-0.5.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Default variant</h4>
      <div className="flex flex-wrap gap-3 items-center">
        <Badge theme="success">
          <Icon variant="CheckCircle" size="12" />
          Passed
        </Badge>
        <Badge theme="warn">
          <Icon variant="Warning" size="12" />
          Warning
        </Badge>
        <Badge theme="error">
          <Icon variant="WarningCircle" size="12" />
          Denied
        </Badge>
        <Badge theme="info">
          <Icon variant="Info" size="12" />
          Info
        </Badge>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Code variant</h4>
      <div className="flex flex-wrap gap-3 items-center">
        <Badge variant="code" theme="success">200 OK</Badge>
        <Badge variant="code" theme="warn">deprecated</Badge>
        <Badge variant="code" theme="error">404</Badge>
        <Badge variant="code" theme="info">beta</Badge>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Without icons</h4>
      <div className="flex flex-wrap gap-3 items-center">
        <Badge theme="success">Active</Badge>
        <Badge theme="warn">Maintenance</Badge>
        <Badge theme="error">Failed</Badge>
        <Badge theme="info">Pending</Badge>
      </div>
    </div>
  </div>
)

export const Variants = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Variants</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Two variants: default (pill, sans-serif) for human labels, code
        (rounded-md, monospace) for technical identifiers.
      </p>
    </div>

    <div className="space-y-3">
      <div className="flex items-center gap-4">
        <Badge variant="default" theme="brand">Premium</Badge>
        <Text variant="subtext">Default: sans-serif, fully rounded</Text>
      </div>
      <div className="flex items-center gap-4">
        <Badge variant="code" theme="brand">v2.1.4</Badge>
        <Text variant="subtext">Code: monospace, rounded-md</Text>
      </div>
    </div>
  </div>
)

export const AutoSizing = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Auto-sizing</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Size is determined by theme. Status themes (success, warn, error, info)
        auto-size to sm. Neutral themes auto-size to lg. No manual size prop
        needed.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">
        Status themes → sm (11px, compact)
      </h4>
      <div className="flex flex-wrap gap-3 items-center">
        <Badge theme="success">Active</Badge>
        <Badge theme="warn">Warning</Badge>
        <Badge theme="error">Failed</Badge>
        <Badge theme="info">Beta</Badge>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">
        Neutral themes → lg (12px, standard)
      </h4>
      <div className="flex flex-wrap gap-3 items-center">
        <Badge theme="neutral">Neutral</Badge>
        <Badge theme="default">Default</Badge>
        <Badge theme="brand">Brand</Badge>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Side by side comparison</h4>
      <div className="flex items-center gap-3 p-3 border rounded">
        <Text weight="strong">API Gateway</Text>
        <Badge theme="success">Live</Badge>
        <Badge variant="code">v2.1.4</Badge>
        <Badge theme="warn">1 warning</Badge>
      </div>
    </div>
  </div>
)

export const UsageExamples = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Usage patterns</h3>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Status indicators</h4>
      <div className="space-y-3">
        <div className="flex items-center justify-between p-3 border rounded">
          <div className="flex items-center gap-3">
            <Text weight="strong">Production deployment</Text>
            <Badge theme="success">
              <Icon variant="CheckCircle" size="12" />
              Live
            </Badge>
          </div>
          <Text variant="subtext" theme="neutral">2 hours ago</Text>
        </div>
        <div className="flex items-center justify-between p-3 border rounded">
          <div className="flex items-center gap-3">
            <Text weight="strong">Staging environment</Text>
            <Badge theme="warn">
              <Icon variant="Warning" size="12" />
              Maintenance
            </Badge>
          </div>
          <Text variant="subtext" theme="neutral">5 minutes ago</Text>
        </div>
        <div className="flex items-center justify-between p-3 border rounded">
          <div className="flex items-center gap-3">
            <Text weight="strong">Development build</Text>
            <Badge theme="error">
              <Icon variant="WarningCircle" size="12" />
              Failed
            </Badge>
          </div>
          <Text variant="subtext" theme="neutral">1 hour ago</Text>
        </div>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Technical metadata</h4>
      <div className="p-4 border rounded-lg space-y-3">
        <div className="flex items-center justify-between">
          <Text weight="strong">API Gateway Service</Text>
          <Badge variant="code" theme="brand">v2.1.4</Badge>
        </div>
        <div className="space-y-2">
          <div className="flex items-center gap-2">
            <Text variant="subtext">Endpoint:</Text>
            <Badge variant="code">/api/v1/auth</Badge>
          </div>
          <div className="flex items-center gap-2">
            <Text variant="subtext">Branch:</Text>
            <Badge variant="code">main</Badge>
          </div>
          <div className="flex items-center gap-2">
            <Text variant="subtext">Status:</Text>
            <Badge variant="code" theme="success">200 OK</Badge>
          </div>
        </div>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Category labels</h4>
      <div className="space-y-3">
        <div className="p-3 border rounded">
          <div className="flex items-center justify-between mb-2">
            <Text weight="strong">Advanced analytics dashboard</Text>
            <div className="flex gap-2">
              <Badge theme="brand">Premium</Badge>
              <Badge theme="info">New</Badge>
            </div>
          </div>
          <Text variant="subtext" theme="neutral">
            Comprehensive analytics with real-time data visualization
          </Text>
        </div>
        <div className="p-3 border rounded">
          <div className="flex items-center justify-between mb-2">
            <Text weight="strong">Legacy API integration</Text>
            <div className="flex gap-2">
              <Badge theme="warn">Deprecated</Badge>
              <Badge>Maintenance</Badge>
            </div>
          </div>
          <Text variant="subtext" theme="neutral">
            Scheduled for replacement in Q2 2024
          </Text>
        </div>
      </div>
    </div>
  </div>
)
