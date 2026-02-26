/* eslint-disable react/no-unescaped-entities */
import { ClickToCopy, ClickToCopyButton } from './ClickToCopy'
import { Text } from './Text'
import { Badge } from './Badge'
import { Icon } from './Icon'

export const BasicUsage = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Basic ClickToCopy Usage</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        ClickToCopy components provide one-click copying functionality with
        visual feedback. They automatically detect text content from children
        and show a temporary "Copied" notification after successful copying.
        Both inline and button variants are available.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Inline Text Copying</h4>
      <div className="space-y-3">
        <div className="p-3 border rounded">
          <ClickToCopy>Simple text to copy</ClickToCopy>
        </div>
        <div className="p-3 border rounded">
          <ClickToCopy>
            <Text weight="strong">Formatted text content</Text>
          </ClickToCopy>
        </div>
        <div className="p-3 border rounded">
          <ClickToCopy>
            <span>Text inside nested elements</span>
          </ClickToCopy>
        </div>
      </div>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Automatic Features:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Intelligent text extraction from React children</li>
        <li>Copy and Check icon states with smooth transitions</li>
        <li>5-second "Copied" notification with automatic dismissal</li>
        <li>Hover effects and proper cursor styling</li>
      </ul>
    </div>
  </div>
)

export const ButtonVariant = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">ClickToCopyButton Variant</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        The{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          ClickToCopyButton
        </code>{' '}
        component provides a more prominent interface for copying functionality.
        It requires explicit{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          textToCopy
        </code>{' '}
        prop and renders as an interactive button with hover effects.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Button Examples</h4>
      <div className="flex gap-4 items-center flex-wrap">
        <ClickToCopyButton textToCopy="Hello World" />
        <ClickToCopyButton textToCopy="API_KEY_12345" />
        <ClickToCopyButton textToCopy="user@example.com" />
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Custom Notification Styling</h4>
      <div className="flex gap-4 items-center flex-wrap">
        <ClickToCopyButton
          textToCopy="Custom blue notification"
          noticeClassName="!bg-blue-500 text-white"
        />
        <ClickToCopyButton
          textToCopy="Custom green notification"
          noticeClassName="!bg-green-500 text-white"
        />
        <ClickToCopyButton
          textToCopy="Custom purple notification"
          noticeClassName="!bg-purple-500 text-white"
        />
      </div>
    </div>

    <div className="grid grid-cols-1 md:grid-cols-2 gap-3 text-sm mt-6">
      <div>
        <strong>Button Style:</strong> Border, padding, hover effects with
        rounded corners
      </div>
      <div>
        <strong>Notification:</strong> Positioned absolutely with shadow and
        custom styling support
      </div>
    </div>
  </div>
)

export const TechnicalExamples = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Technical Content Examples</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        ClickToCopy components are particularly useful for technical content
        like API keys, commands, URLs, and code snippets where users frequently
        need to copy exact values.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">API Keys and Tokens</h4>
      <div className="space-y-3">
        <div className="flex items-center justify-between p-3 border rounded">
          <div>
            <Text variant="label" weight="strong">
              API Key
            </Text>
            <div className="flex items-center gap-2 mt-1">
              <Badge variant="code" size="sm" theme="neutral">
                sk_live_4eC39HqLyjWDarjtT1zdp7dc
              </Badge>
              <ClickToCopyButton textToCopy="sk_live_4eC39HqLyjWDarjtT1zdp7dc" />
            </div>
          </div>
        </div>
        <div className="flex items-center justify-between p-3 border rounded">
          <div>
            <Text variant="label" weight="strong">
              Access Token
            </Text>
            <div className="flex items-center gap-2 mt-1">
              <ClickToCopy>
                <Badge variant="code" size="sm" theme="info">
                  eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9
                </Badge>
              </ClickToCopy>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Commands and URLs</h4>
      <div className="space-y-3">
        <div className="p-3 border rounded bg-gray-50 dark:bg-gray-800">
          <Text variant="label" weight="strong">
            Installation Command
          </Text>
          <div className="mt-2 flex items-center gap-2">
            <ClickToCopy>
              <code className="text-sm bg-black text-green-400 p-2 rounded">
                npm install @nuon/dashboard-ui
              </code>
            </ClickToCopy>
          </div>
        </div>
        <div className="p-3 border rounded bg-gray-50 dark:bg-gray-800">
          <Text variant="label" weight="strong">
            API Endpoint
          </Text>
          <div className="mt-2 flex items-center gap-2">
            <ClickToCopy>
              <Badge variant="code" theme="brand">
                https://api.nuon.co/v1/organizations
              </Badge>
            </ClickToCopy>
          </div>
        </div>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Configuration Values</h4>
      <div className="grid grid-cols-1 gap-3">
        <div className="p-3 border rounded">
          <div className="flex items-center justify-between">
            <Text variant="subtext" weight="strong">
              Database URL
            </Text>
            <ClickToCopyButton textToCopy="postgresql://user:pass@localhost:5432/db" />
          </div>
        </div>
        <div className="p-3 border rounded">
          <div className="flex items-center justify-between">
            <Text variant="subtext" weight="strong">
              Redis Connection
            </Text>
            <ClickToCopyButton textToCopy="redis://localhost:6379" />
          </div>
        </div>
        <div className="p-3 border rounded">
          <div className="flex items-center justify-between">
            <Text variant="subtext" weight="strong">
              Webhook URL
            </Text>
            <ClickToCopyButton textToCopy="https://webhook.site/123e4567-e89b-12d3-a456-426614174000" />
          </div>
        </div>
      </div>
    </div>
  </div>
)

export const UsageExamples = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Common Usage Patterns</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        ClickToCopy components integrate seamlessly with other interface
        elements to provide copying functionality in various contexts. Here are
        recommended patterns for different use cases.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Settings and Configuration</h4>
      <div className="space-y-3 border rounded-lg p-4">
        <div className="flex items-center justify-between">
          <div>
            <Text weight="strong">Organization ID</Text>
            <Text variant="subtext" theme="neutral">
              Unique identifier for your organization
            </Text>
          </div>
          <ClickToCopy>
            <Badge variant="code" theme="brand">
              org_123e4567e89b
            </Badge>
          </ClickToCopy>
        </div>
        <div className="flex items-center justify-between">
          <div>
            <Text weight="strong">Webhook Secret</Text>
            <Text variant="subtext" theme="neutral">
              Use this to verify webhook signatures
            </Text>
          </div>
          <ClickToCopyButton textToCopy="whsec_1234567890abcdef" />
        </div>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Documentation and Help</h4>
      <div className="p-4 border rounded-lg bg-blue-50 dark:bg-blue-950/20">
        <div className="flex items-start gap-3">
          <Icon
            variant="Info"
            size="16"
            className="text-blue-600 dark:text-blue-400 mt-0.5"
          />
          <div className="flex-1">
            <Text weight="strong">Quick Start</Text>
            <Text variant="subtext" theme="neutral" className="mt-1">
              Copy this command to get started with the Nuon CLI:
            </Text>
            <div className="mt-3">
              <ClickToCopy>
                <code className="bg-white dark:bg-gray-800 p-2 rounded border text-sm">
                  curl -fsSL https://install.nuon.co | sh
                </code>
              </ClickToCopy>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Error Messages and Support</h4>
      <div className="p-4 border rounded-lg bg-red-50 dark:bg-red-950/20">
        <div className="flex items-start gap-3">
          <Icon
            variant="WarningCircle"
            size="16"
            className="text-red-600 dark:text-red-400 mt-0.5"
          />
          <div className="flex-1">
            <Text weight="strong">Error Details</Text>
            <Text variant="subtext" theme="neutral" className="mt-1">
              Share this error ID with support for faster resolution:
            </Text>
            <div className="mt-3 flex items-center gap-2">
              <Badge variant="code" theme="error" size="sm">
                ERR_123456789
              </Badge>
              <ClickToCopyButton textToCopy="ERR_123456789" />
            </div>
          </div>
        </div>
      </div>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Best Practices:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Use inline ClickToCopy for text that's already visible</li>
        <li>Use ClickToCopyButton for dedicated copy actions</li>
        <li>Provide clear context about what will be copied</li>
        <li>Consider custom notification styling for themed interfaces</li>
        <li>Combine with Badges or code formatting for technical content</li>
      </ul>
    </div>
  </div>
)
