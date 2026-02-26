/* eslint-disable react/no-unescaped-entities */
import { ContextTooltip } from './ContextTooltip'
import { Button } from './Button'
import { Icon } from './Icon'
import { Text } from './Text'

export const BasicUsage = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Basic ContextTooltip Usage</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        ContextTooltip provides rich contextual menus that appear on hover. They
        support multiple interaction types including navigation links, click
        handlers, and informational displays. The tooltip automatically adapts
        its layout based on the number of items and available space.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Single Action</h4>
      <div className="flex justify-center p-8 border rounded">
        <ContextTooltip
          title="Actions"
          items={[
            {
              id: 'edit',
              title: 'Edit Configuration',
              subtitle: 'Modify settings',
              leftContent: <Icon variant="Pencil" />,
              // eslint-disable-next-line
              onClick: () => console.log('Edit clicked'),
            },
          ]}
        >
          <Button>Single Action Tooltip</Button>
        </ContextTooltip>
      </div>
      <Text variant="subtext" theme="neutral">
        Simple tooltip with one interactive item
      </Text>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Key Features:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Automatic positioning with collision detection</li>
        <li>Support for both click handlers and navigation links</li>
        <li>Rich content with titles, subtitles, and icons</li>
        <li>Scrollable content for large item lists</li>
      </ul>
    </div>
  </div>
)

export const MultipleActions = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Multiple Actions</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        ContextTooltips shine with multiple related actions. Each item can have
        different interaction types - some navigate using{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          href
        </code>
        , others execute functions with{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          onClick
        </code>
        . Icons and warning indicators provide visual context.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Standard Action Menu</h4>
      <div className="flex justify-center p-8 border rounded">
        <ContextTooltip
          title="Quick Actions"
          items={[
            {
              id: 'view',
              title: 'View Details',
              subtitle: 'See full information',
              leftContent: <Icon variant="Eye" />,
              href: '/details',
            },
            {
              id: 'edit',
              title: 'Edit Configuration',
              subtitle: 'Modify settings',
              leftContent: <Icon variant="Pencil" />,
              // eslint-disable-next-line
              onClick: () => console.log('Edit clicked'),
            },
            {
              id: 'duplicate',
              title: 'Duplicate Item',
              subtitle: 'Create a copy',
              leftContent: <Icon variant="Copy" />,
              // eslint-disable-next-line
              onClick: () => console.log('Duplicate clicked'),
            },
            {
              id: 'delete',
              title: 'Delete Item',
              subtitle: 'Remove permanently',
              leftContent: <Icon variant="Trash" />,
              rightContent: <Icon variant="Warning" />,
              // eslint-disable-next-line
              onClick: () => console.log('Delete clicked'),
            },
          ]}
        >
          <Button>Multiple Actions</Button>
        </ContextTooltip>
      </div>
      <Text variant="subtext" theme="neutral">
        Common CRUD operations with visual indicators for destructive actions
      </Text>
    </div>
  </div>
)

export const LargeItemLists = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Large Item Lists</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        For extensive lists, ContextTooltip provides scrollable content with
        customizable dimensions. Use{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          maxHeight
        </code>{' '}
        and{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          width
        </code>{' '}
        props to control the tooltip size and prevent overwhelming the
        interface.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Infrastructure Components List</h4>
      <div className="flex justify-center p-8 border rounded">
        <ContextTooltip
          title="All Components"
          position="bottom"
          maxHeight="max-h-64"
          width="w-64"
          items={[
            {
              id: 'app1',
              title: 'Frontend Application',
              subtitle: 'React web app',
              leftContent: <Icon variant="Globe" />,
              href: '/components/app1',
            },
            {
              id: 'api1',
              title: 'REST API Service',
              subtitle: 'Node.js backend',
              leftContent: <Icon variant="HardDrives" />,
              href: '/components/api1',
            },
            {
              id: 'db1',
              title: 'PostgreSQL Database',
              subtitle: 'Primary database',
              leftContent: <Icon variant="Database" />,
              href: '/components/db1',
            },
            {
              id: 'cache1',
              title: 'Redis Cache',
              subtitle: 'In-memory cache',
              leftContent: <Icon variant="Lightning" />,
              href: '/components/cache1',
            },
            {
              id: 'worker1',
              title: 'Background Workers',
              subtitle: 'Job processing',
              leftContent: <Icon variant="Cpu" />,
              href: '/components/worker1',
            },
            {
              id: 'queue1',
              title: 'Message Queue',
              subtitle: 'RabbitMQ service',
              leftContent: <Icon variant="List" />,
              href: '/components/queue1',
            },
            {
              id: 'auth1',
              title: 'Authentication Service',
              subtitle: 'OAuth provider',
              leftContent: <Icon variant="Shield" />,
              href: '/components/auth1',
            },
            {
              id: 'monitor1',
              title: 'Monitoring Stack',
              subtitle: 'Prometheus & Grafana',
              leftContent: <Icon variant="ChartBar" />,
              href: '/components/monitor1',
            },
            {
              id: 'logs1',
              title: 'Log Aggregation',
              subtitle: 'ELK stack',
              leftContent: <Icon variant="FileText" />,
              href: '/components/logs1',
            },
            {
              id: 'cdn1',
              title: 'Content Delivery Network',
              subtitle: 'CloudFront CDN',
              leftContent: <Icon variant="Cloud" />,
              href: '/components/cdn1',
            },
            {
              id: 'lb1',
              title: 'Load Balancer',
              subtitle: 'Application load balancer',
              leftContent: <Icon variant="Shuffle" />,
              href: '/components/lb1',
            },
            {
              id: 'storage1',
              title: 'Object Storage',
              subtitle: 'S3 compatible storage',
              leftContent: <Icon variant="HardDrive" />,
              href: '/components/storage1',
            },
            {
              id: 'backup1',
              title: 'Backup Service',
              subtitle: 'Automated backups',
              leftContent: <Icon variant="Archive" />,
              href: '/components/backup1',
            },
            {
              id: 'search1',
              title: 'Search Engine',
              subtitle: 'Elasticsearch cluster',
              leftContent: <Icon variant="MagnifyingGlass" />,
              href: '/components/search1',
            },
            {
              id: 'mail1',
              title: 'Email Service',
              subtitle: 'SMTP provider',
              leftContent: <Icon variant="Envelope" />,
              href: '/components/mail1',
            },
          ]}
        >
          <Button>15 Infrastructure Components</Button>
        </ContextTooltip>
      </div>
      <Text variant="subtext" theme="neutral">
        Scrollable list with custom width and height limits
      </Text>
    </div>
  </div>
)

export const CustomConfiguration = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Custom Configuration</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        ContextTooltip supports extensive customization including positioning,
        sizing, and content display options. The{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          showCount
        </code>{' '}
        prop controls whether item counts are displayed, and{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          position
        </code>{' '}
        handles collision detection and optimal placement.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">
        Left-Positioned with Custom Content
      </h4>
      <div className="flex justify-center p-8 border rounded">
        <ContextTooltip
          title="Custom Options"
          showCount={false}
          width="w-72"
          position="left"
          items={[
            {
              id: 'option1',
              title: 'Wide tooltip without count',
              subtitle: 'Positioned to the left',
              leftContent: <Icon variant="Faders" />,
              // eslint-disable-next-line
              onClick: () => console.log('Option 1'),
            },
            {
              id: 'option2',
              title: 'Custom right content',
              subtitle: 'With status indicator',
              leftContent: <Icon variant="CheckCircle" />,
              rightContent: (
                <div className="w-2 h-2 bg-green-500 rounded-full" />
              ),
              // eslint-disable-next-line
              onClick: () => console.log('Option 2'),
            },
          ]}
        >
          <Button variant="primary">Custom Configuration</Button>
        </ContextTooltip>
      </div>
      <Text variant="subtext" theme="neutral">
        Wide tooltip positioned left with custom status indicators
      </Text>
    </div>
  </div>
)

export const InteractiveHandlers = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Interactive Event Handlers</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        ContextTooltip supports both individual item{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          onClick
        </code>{' '}
        handlers and a global{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          onItemClick
        </code>{' '}
        handler that receives the clicked item. This allows for centralized
        event handling and analytics tracking.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Dual Handler Example</h4>
      <div className="flex justify-center p-8 border rounded">
        <ContextTooltip
          title="Interactive Items"
          onItemClick={(item) => alert(`Global handler: ${item.title}`)}
          items={[
            {
              id: 'action1',
              title: 'Action with handler',
              subtitle: 'Triggers both handlers',
              leftContent: <Icon variant="Play" />,
              // eslint-disable-next-line
              onClick: () => console.log('Individual handler executed'),
            },
            {
              id: 'action2',
              title: 'Another action',
              subtitle: 'Only global handler',
              leftContent: <Icon variant="Square" />,
            },
          ]}
        >
          <Button variant="secondary">Interactive Handlers</Button>
        </ContextTooltip>
      </div>
      <Text variant="subtext" theme="neutral">
        Items can have individual onClick handlers plus a global onItemClick
        handler
      </Text>
    </div>
  </div>
)

export const UsageExamples = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Common Usage Patterns</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        ContextTooltips serve multiple purposes from action menus to
        informational displays. Here are common patterns and recommended
        approaches for different scenarios.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">System Status Display</h4>
      <div className="flex justify-center p-8 border rounded">
        <ContextTooltip
          title="System Status"
          items={[
            {
              id: 'status1',
              title: 'API Response Time',
              subtitle: '245ms average',
            },
            {
              id: 'status2',
              title: 'Database Connections',
              subtitle: '12/100 active',
            },
            {
              id: 'status3',
              title: 'Memory Usage',
              subtitle: '68% of 8GB',
            },
            {
              id: 'status4',
              title: 'CPU Utilization',
              subtitle: '23% across 4 cores',
            },
            {
              id: 'status5',
              title: 'Disk Space',
              subtitle: '156GB of 500GB used',
            },
          ]}
        >
          <Button variant="ghost">System Metrics</Button>
        </ContextTooltip>
      </div>
      <Text variant="subtext" theme="neutral">
        Informational tooltip without click handlers - pure data display
      </Text>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Quick Navigation Menu</h4>
      <div className="flex justify-center p-8 border rounded">
        <ContextTooltip
          title="Quick Access"
          items={[
            {
              id: 'dashboard',
              title: 'Dashboard',
              subtitle: 'Overview and metrics',
              leftContent: <Icon variant="House" />,
              href: '/dashboard',
            },
            {
              id: 'apps',
              title: 'Applications',
              subtitle: 'Manage your apps',
              leftContent: <Icon variant="GridFour" />,
              href: '/apps',
            },
            {
              id: 'settings',
              title: 'Settings',
              subtitle: 'Configuration and preferences',
              leftContent: <Icon variant="Gear" />,
              href: '/settings',
            },
            {
              id: 'support',
              title: 'Support',
              subtitle: 'Help and documentation',
              leftContent: <Icon variant="Question" />,
              href: '/support',
            },
          ]}
        >
          <Button variant="primary">Navigation Menu</Button>
        </ContextTooltip>
      </div>
      <Text variant="subtext" theme="neutral">
        Navigation-focused tooltip with href links for routing
      </Text>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Resource Management Actions</h4>
      <div className="flex justify-center p-8 border rounded">
        <ContextTooltip
          title="Resource Actions"
          onItemClick={(item) => {
            // eslint-disable-next-line
            console.log(`Action: ${item.id}`)
          }}
          items={[
            {
              id: 'start',
              title: 'Start Service',
              subtitle: 'Boot up the service',
              leftContent: <Icon variant="Play" />,
              rightContent: (
                <div className="w-2 h-2 bg-gray-400 rounded-full" />
              ),
            },
            {
              id: 'restart',
              title: 'Restart Service',
              subtitle: 'Graceful restart',
              leftContent: <Icon variant="ArrowClockwise" />,
              rightContent: (
                <div className="w-2 h-2 bg-yellow-500 rounded-full" />
              ),
            },
            {
              id: 'stop',
              title: 'Stop Service',
              subtitle: 'Shutdown gracefully',
              leftContent: <Icon variant="Stop" />,
              rightContent: <div className="w-2 h-2 bg-red-500 rounded-full" />,
            },
            {
              id: 'logs',
              title: 'View Logs',
              subtitle: 'Recent activity',
              leftContent: <Icon variant="FileText" />,
              href: '/logs',
            },
          ]}
        >
          <Button variant="secondary">Manage Service</Button>
        </ContextTooltip>
      </div>
      <Text variant="subtext" theme="neutral">
        Mixed interaction types with status indicators and centralized handling
      </Text>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Best Practices:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Group related actions logically in the same tooltip</li>
        <li>Use consistent icon patterns for similar action types</li>
        <li>Provide clear titles and descriptive subtitles</li>
        <li>Use href for navigation and onClick for state changes</li>
        <li>Consider visual indicators (colors, icons) for action states</li>
        <li>Implement both individual and global click handlers when needed</li>
      </ul>
    </div>
  </div>
)
