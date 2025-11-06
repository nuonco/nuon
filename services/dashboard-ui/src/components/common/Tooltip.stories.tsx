/* eslint-disable react/no-unescaped-entities */
import { Button } from './Button'
import { Icon } from './Icon'
import { Status } from './Status'
import { Text } from './Text'
import { Tooltip } from './Tooltip'
import { Badge } from './Badge'

export const BasicUsage = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Basic Tooltip Usage</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Tooltips provide contextual information when users hover over elements.
        They automatically position themselves and include smooth fade-in/out
        animations. Content appears on hover and disappears when the mouse
        leaves the trigger area.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Simple Tooltip</h4>
      <div className="flex gap-8 p-8 border rounded-lg">
        <Tooltip tipContent="This is a simple tooltip with helpful information">
          <Text className="cursor-help border-b border-dashed">Hover me</Text>
        </Tooltip>
      </div>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Accessibility:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Tooltips have proper ARIA role="tooltip" attributes</li>
        <li>Content is accessible to screen readers</li>
        <li>Keyboard navigation support through focus events</li>
        <li>Non-intrusive design that doesn't block interface elements</li>
      </ul>
    </div>
  </div>
)

export const Positioning = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Tooltip Positioning</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        The{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          position
        </code>{' '}
        prop controls where the tooltip appears relative to its trigger element.
        Positioning is automatically calculated and includes proper spacing for
        optimal readability.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">All Position Options</h4>
      <div className="grid grid-cols-2 md:grid-cols-4 gap-8 p-8 border rounded-lg">
        <div className="text-center">
          <Tooltip tipContent="Tooltip positioned above" position="top">
            <Button variant="secondary">Top</Button>
          </Tooltip>
        </div>
        <div className="text-center">
          <Tooltip tipContent="Tooltip positioned below" position="bottom">
            <Button variant="secondary">Bottom</Button>
          </Tooltip>
        </div>
        <div className="text-center">
          <Tooltip tipContent="Tooltip positioned to the left" position="left">
            <Button variant="secondary">Left</Button>
          </Tooltip>
        </div>
        <div className="text-center">
          <Tooltip
            tipContent="Tooltip positioned to the right"
            position="right"
          >
            <Button variant="secondary">Right</Button>
          </Tooltip>
        </div>
      </div>
    </div>

    <div className="grid grid-cols-1 md:grid-cols-2 gap-3 text-sm mt-6">
      <div>
        <strong>top:</strong> Appears above the trigger with downward arrow
        (default)
      </div>
      <div>
        <strong>bottom:</strong> Appears below the trigger with upward arrow
      </div>
      <div>
        <strong>left:</strong> Appears to the left with rightward arrow
      </div>
      <div>
        <strong>right:</strong> Appears to the right with leftward arrow
      </div>
    </div>
  </div>
)

export const WithIcon = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Tooltips with Icons</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Enable the{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          showIcon
        </code>{' '}
        prop to automatically add a question mark icon next to the trigger
        content. This provides a visual cue that additional information is
        available on hover.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Help Icons</h4>
      <div className="space-y-4 p-6 border rounded-lg">
        <div className="flex items-center gap-4">
          <Text weight="strong">Username</Text>
          <Tooltip
            tipContent="Your username must be unique and contain only letters, numbers, and underscores"
            showIcon
          >
            <Text variant="subtext">What's this?</Text>
          </Tooltip>
        </div>
        <div className="flex items-center gap-4">
          <Text weight="strong">API Rate Limit</Text>
          <Tooltip
            tipContent="Maximum number of API requests allowed per minute. Premium users get higher limits."
            showIcon
            position="right"
          >
            <Text variant="subtext">Learn more</Text>
          </Tooltip>
        </div>
        <div className="flex items-center gap-4">
          <Badge theme="info">Beta Feature</Badge>
          <Tooltip
            tipContent="This feature is in beta testing. Report any issues to support."
            showIcon
            position="bottom"
          >
            <Text variant="subtext">Beta info</Text>
          </Tooltip>
        </div>
      </div>
    </div>
  </div>
)

export const ContentTypes = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Tooltip Content Types</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Tooltips can contain various types of content including plain text,
        formatted text, and even complex components. Use the{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          tipContentClassName
        </code>{' '}
        prop to customize the styling of the tooltip container.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Different Content Examples</h4>
      <div className="flex flex-wrap gap-6 p-6 border rounded-lg">
        <Tooltip tipContent="Simple text tooltip">
          <Button variant="secondary">Text Only</Button>
        </Tooltip>

        <Tooltip
          tipContent={
            <div className="space-y-1">
              <Text weight="strong" variant="subtext">
                Formatted Content
              </Text>
              <Text variant="label">Multiple lines with styling</Text>
            </div>
          }
          tipContentClassName="!whitespace-normal !w-48"
        >
          <Button variant="secondary">Rich Text</Button>
        </Tooltip>

        <Tooltip
          tipContent={
            <div className="flex items-center gap-2">
              <Status status="success" isWithoutText />
              <Text variant="subtext">System healthy</Text>
            </div>
          }
        >
          <Button variant="secondary">With Components</Button>
        </Tooltip>

        <Tooltip
          tipContent={
            <div className="space-y-2">
              <Text weight="strong" variant="subtext">
                Quick Actions
              </Text>
              <div className="space-y-1">
                <Text variant="label">• View details</Text>
                <Text variant="label">• Edit settings</Text>
                <Text variant="label">• Delete item</Text>
              </div>
            </div>
          }
          tipContentClassName="!whitespace-normal !w-40"
          position="bottom"
        >
          <Button variant="secondary">Action List</Button>
        </Tooltip>
      </div>
    </div>
  </div>
)

export const InteractiveTooltips = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Interactive Tooltip Content</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        For complex interactions, tooltips can contain interactive elements like
        buttons and lists. Use the{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          isOpen
        </code>{' '}
        prop to control tooltip visibility programmatically.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Complex Tooltip Example</h4>
      <div className="flex gap-6 p-6 border rounded-lg">
        <Tooltip
          isOpen
          position="right"
          tipContentClassName="!p-0"
          tipContent={
            <div className="flex flex-col w-52">
              <div className="px-3 py-2 border-b bg-gray-50 dark:bg-gray-800">
                <div className="flex items-center justify-between">
                  <Text variant="subtext" weight="strong">
                    Dependencies
                  </Text>
                  <Badge theme="info">5</Badge>
                </div>
              </div>
              <div className="flex flex-col divide-y max-h-32 overflow-y-auto">
                {[
                  { status: 'success', name: 'auth_service' },
                  { status: 'error', name: 'redis_cache' },
                  { status: 'warn', name: 'file_storage' },
                  { status: 'success', name: 'api_gateway' },
                  { status: 'info', name: 'notification_service' },
                ].map((dep) => (
                  <div
                    key={dep.name}
                    className="flex items-center justify-between p-2 hover:bg-gray-50 dark:hover:bg-gray-800"
                  >
                    <div className="flex items-center gap-2">
                      <Status status={dep.status} isWithoutText />
                      <Text variant="label" className="truncate">
                        {dep.name}
                      </Text>
                    </div>
                    <Icon variant="CaretRight" size="12" />
                  </div>
                ))}
              </div>
            </div>
          }
        >
          <Button variant="secondary">View Dependencies</Button>
        </Tooltip>
      </div>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Interactive Content Guidelines:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Keep interactive tooltips simple and focused</li>
        <li>Consider using modals or popovers for complex interactions</li>
        <li>Ensure tooltip content is accessible via keyboard navigation</li>
        <li>Use consistent styling with your application's design system</li>
      </ul>
    </div>
  </div>
)
