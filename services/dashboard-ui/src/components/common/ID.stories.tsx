/* eslint-disable react/no-unescaped-entities */
import { ID } from './ID'
import { Text } from './Text'

export const BasicUsage = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Basic ID Usage</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        ID components display technical identifiers with monospace typography
        and built-in click-to-copy functionality. They automatically handle
        long IDs with proper text wrapping and provide visual feedback when
        copied to clipboard.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Simple IDs</h4>
      <div className="space-y-3 p-4 border rounded">
        <div className="flex items-center gap-3">
          <ID>abc-123-def</ID>
          <Text variant="subtext" theme="neutral">Basic identifier</Text>
        </div>
        <div className="flex items-center gap-3">
          <ID>user-456789</ID>
          <Text variant="subtext" theme="neutral">User identifier</Text>
        </div>
        <div className="flex items-center gap-3">
          <ID>app_config_12345</ID>
          <Text variant="subtext" theme="neutral">Config identifier</Text>
        </div>
      </div>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Key Features:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Monospace font for consistent character alignment</li>
        <li>Automatic click-to-copy functionality with visual feedback</li>
        <li>Proper text wrapping for long identifiers</li>
        <li>Inherits Text component properties for consistent styling</li>
      </ul>
    </div>
  </div>
)

export const IDTypes = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Different ID Types</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        ID components work with various identifier formats commonly used in
        software systems including UUIDs, database IDs, hashes, and cloud
        resource identifiers.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Common ID Formats</h4>
      <div className="space-y-3">
        <div className="p-3 border rounded">
          <Text variant="label" weight="strong">UUID</Text>
          <div className="mt-1">
            <ID>f47ac10b-58cc-4372-a567-0e02b2c3d479</ID>
          </div>
        </div>
        
        <div className="p-3 border rounded">
          <Text variant="label" weight="strong">Short ID</Text>
          <div className="mt-1">
            <ID>abc123</ID>
          </div>
        </div>
        
        <div className="p-3 border rounded">
          <Text variant="label" weight="strong">Database ID</Text>
          <div className="mt-1">
            <ID>1234567890</ID>
          </div>
        </div>
        
        <div className="p-3 border rounded">
          <Text variant="label" weight="strong">SHA256 Hash</Text>
          <div className="mt-1">
            <ID>sha256:a3b5c2d7e9f1234567890abcdef1234567890abcdef1234567890abcdef</ID>
          </div>
        </div>
      </div>
    </div>
  </div>
)

export const TextVariants = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Text Variants</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        ID components inherit all Text component properties, allowing you to
        customize typography while maintaining the monospace font and
        click-to-copy behavior.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Size Variants</h4>
      <div className="space-y-3">
        <div className="flex items-center gap-3">
          <ID variant="subtext">small-id-123</ID>
          <Text variant="subtext" theme="neutral">Subtext variant</Text>
        </div>
        <div className="flex items-center gap-3">
          <ID variant="base">base-id-456</ID>
          <Text variant="subtext" theme="neutral">Base variant (default)</Text>
        </div>
        <div className="flex items-center gap-3">
          <ID variant="h3">large-id-789</ID>
          <Text variant="subtext" theme="neutral">H3 variant</Text>
        </div>
      </div>
    </div>
  </div>
)

export const CustomStyling = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Custom Styling</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        The{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          clickToCopyProps
        </code>{' '}
        allows customization of the underlying ClickToCopy component,
        including container styling and notification appearance.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Container Styling</h4>
      <div className="space-y-3">
        <div className="p-3 border rounded">
          <Text variant="label" weight="strong">Custom Container</Text>
          <div className="mt-2">
            <ID clickToCopyProps={{ className: 'bg-blue-50 dark:bg-blue-950 p-2 rounded border' }}>
              styled-container-id
            </ID>
          </div>
        </div>
        
        <div className="p-3 border rounded">
          <Text variant="label" weight="strong">Custom Notification</Text>
          <div className="mt-2">
            <ID clickToCopyProps={{ noticeClassName: '!bg-green-500 text-white' }}>
              custom-notice-id
            </ID>
          </div>
          <Text variant="subtext" theme="neutral" className="mt-1">
            Click to see green notification
          </Text>
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
        ID components are commonly used in technical interfaces, resource
        listings, and configuration displays. Here are typical patterns
        for different contexts.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Resource Details</h4>
      <div className="space-y-3 border rounded-lg p-4">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <Text variant="label" weight="strong">Application ID</Text>
            <div className="mt-1">
              <ID>app-prod-web-frontend</ID>
            </div>
          </div>
          <div>
            <Text variant="label" weight="strong">Deployment ID</Text>
            <div className="mt-1">
              <ID>deploy-2024-03-15-v2-1-0</ID>
            </div>
          </div>
          <div className="md:col-span-2">
            <Text variant="label" weight="strong">Container Image</Text>
            <div className="mt-1">
              <ID>registry.example.com/my-app:sha256-a1b2c3d4e5f6</ID>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Long Cloud Identifiers</h4>
      <div className="space-y-3 max-w-2xl">
        <div className="p-3 border rounded">
          <Text variant="label" weight="strong">AWS IAM Role ARN</Text>
          <div className="mt-1">
            <ID>arn:aws:iam::123456789012:role/service-role/MyLambdaRole</ID>
          </div>
        </div>
        
        <div className="p-3 border rounded">
          <Text variant="label" weight="strong">Google Cloud Instance</Text>
          <div className="mt-1">
            <ID>projects/my-project/locations/us-central1/instances/my-instance</ID>
          </div>
        </div>
        
        <div className="p-3 border rounded">
          <Text variant="label" weight="strong">Azure Resource ID</Text>
          <div className="mt-1">
            <ID>/subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/my-rg/providers/Microsoft.Web/sites/my-app</ID>
          </div>
        </div>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">API Keys and Tokens</h4>
      <div className="space-y-3 max-w-lg">
        <div className="p-3 border rounded bg-gray-50 dark:bg-gray-800">
          <Text variant="label" weight="strong">API Key</Text>
          <div className="mt-1">
            <ID variant="subtext">sk_live_4eC39HqLyjWDarjtT1zdp7dc</ID>
          </div>
        </div>
        
        <div className="p-3 border rounded bg-gray-50 dark:bg-gray-800">
          <Text variant="label" weight="strong">JWT Token (truncated)</Text>
          <div className="mt-1">
            <ID variant="subtext">eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...</ID>
          </div>
        </div>
      </div>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Best Practices:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Use appropriate text variants based on visual hierarchy</li>
        <li>Provide context labels for complex or unfamiliar ID formats</li>
        <li>Consider truncating very long IDs with tooltips for full values</li>
        <li>Group related IDs in structured layouts for better scanning</li>
        <li>Use consistent styling for similar types of identifiers</li>
      </ul>
    </div>
  </div>
)
