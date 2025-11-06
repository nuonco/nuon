/* eslint-disable react/no-unescaped-entities */
import { LabeledValue } from './LabeledValue'
import { Text } from './Text'
import { Badge } from './Badge'
import { Button } from './Button'
import { Code } from './Code'
import { ID } from './ID'

export const BasicUsage = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Basic LabeledValue Usage</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        LabeledValue provides a consistent layout for displaying labeled content
        with proper spacing and typography. It's perfect for forms,
        configuration displays, and any interface where you need to associate
        labels with values or interactive elements.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Simple Text Values</h4>
      <div className="space-y-3 max-w-md">
        <LabeledValue label="Name">
          <Text>John Doe</Text>
        </LabeledValue>
        <LabeledValue label="Email Address">
          <Text theme="info">john.doe@example.com</Text>
        </LabeledValue>
        <LabeledValue label="Role">
          <Text weight="strong">Administrator</Text>
        </LabeledValue>
      </div>
      <Text variant="subtext" theme="neutral">
        Labels provide clear identification for associated content
      </Text>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Key Features:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Consistent spacing and alignment for labeled content</li>
        <li>Flexible content support - accepts any React children</li>
        <li>Typography scaling that works with the design system</li>
        <li>Proper semantic structure for accessibility</li>
      </ul>
    </div>
  </div>
)

export const DifferentContentTypes = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Different Content Types</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        LabeledValue can contain any React content including badges, buttons,
        code blocks, IDs, and complex layouts. This makes it versatile for
        various interface patterns.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Various Content Examples</h4>
      <div className="space-y-4 max-w-2xl">
        <LabeledValue label="Status">
          <Badge theme="success">Active</Badge>
        </LabeledValue>

        <LabeledValue label="Actions">
          <div className="flex gap-2">
            <Button variant="primary" size="sm">
              Edit
            </Button>
            <Button variant="secondary" size="sm">
              Delete
            </Button>
          </div>
        </LabeledValue>

        <LabeledValue label="User ID">
          <ID>usr_1234567890abcdef</ID>
        </LabeledValue>

        <LabeledValue label="API Key">
          <Code>sk-proj-abc123def456ghi789</Code>
        </LabeledValue>

        <LabeledValue label="Tags">
          <div className="flex flex-wrap gap-2">
            <Badge theme="brand" size="sm">
              Frontend
            </Badge>
            <Badge theme="info" size="sm">
              React
            </Badge>
            <Badge theme="neutral" size="sm">
              TypeScript
            </Badge>
          </div>
        </LabeledValue>

        <LabeledValue label="Description">
          <Text>
            This is a longer description that demonstrates how LabeledValue
            handles multi-line content. The layout remains consistent and
            readable even with longer text blocks.
          </Text>
        </LabeledValue>
      </div>
    </div>
  </div>
)

export const FormLayouts = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Form Layout Examples</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        LabeledValue is commonly used in form layouts, settings panels, and
        configuration interfaces where labels need to be clearly associated with
        input fields or display values.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">User Profile Form</h4>
      <div className="p-4 border rounded-lg max-w-lg">
        <Text variant="h3" weight="stronger" className="mb-4">
          Profile Settings
        </Text>
        <div className="space-y-4">
          <LabeledValue label="Display Name">
            <input
              type="text"
              defaultValue="John Doe"
              className="w-full px-3 py-2 border rounded-md focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
            />
          </LabeledValue>

          <LabeledValue label="Email">
            <input
              type="email"
              defaultValue="john.doe@example.com"
              className="w-full px-3 py-2 border rounded-md focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
            />
          </LabeledValue>

          <LabeledValue label="Role">
            <select className="w-full px-3 py-2 border rounded-md focus:ring-2 focus:ring-blue-500 focus:border-blue-500">
              <option>Administrator</option>
              <option>User</option>
              <option>Guest</option>
            </select>
          </LabeledValue>

          <LabeledValue label="Notifications">
            <div className="space-y-2">
              <label className="flex items-center gap-2">
                <input type="checkbox" defaultChecked />
                <Text variant="subtext">Email notifications</Text>
              </label>
              <label className="flex items-center gap-2">
                <input type="checkbox" />
                <Text variant="subtext">SMS notifications</Text>
              </label>
            </div>
          </LabeledValue>
        </div>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Read-Only Configuration Display</h4>
      <div className="p-4 border rounded-lg max-w-lg">
        <Text variant="h3" weight="stronger" className="mb-4">
          System Configuration
        </Text>
        <div className="space-y-3">
          <LabeledValue label="Environment">
            <Badge theme="success">Production</Badge>
          </LabeledValue>

          <LabeledValue label="Region">
            <Text>us-west-2</Text>
          </LabeledValue>

          <LabeledValue label="Instance Type">
            <Text family="mono">t3.medium</Text>
          </LabeledValue>

          <LabeledValue label="Auto Scaling">
            <Badge theme="info">Enabled</Badge>
          </LabeledValue>

          <LabeledValue label="Backup Schedule">
            <Text variant="subtext">Daily at 2:00 AM UTC</Text>
          </LabeledValue>

          <LabeledValue label="Monitoring">
            <div className="flex items-center gap-2">
              <Badge theme="success" size="sm">
                Active
              </Badge>
              <Button variant="ghost" size="sm">
                View Dashboard
              </Button>
            </div>
          </LabeledValue>
        </div>
      </div>
    </div>
  </div>
)

export const ResponsiveLayouts = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Responsive Layouts</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        LabeledValue works well in responsive grid layouts and can adapt to
        different screen sizes while maintaining proper label-content
        relationships.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Grid Layout Example</h4>
      <div className="p-4 border rounded-lg">
        <Text variant="h3" weight="stronger" className="mb-4">
          Project Details
        </Text>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <LabeledValue label="Project Name">
            <Text weight="strong">E-commerce Platform</Text>
          </LabeledValue>

          <LabeledValue label="Status">
            <Badge theme="success">Active</Badge>
          </LabeledValue>

          <LabeledValue label="Created">
            <Text variant="subtext">March 15, 2024</Text>
          </LabeledValue>

          <LabeledValue label="Last Updated">
            <Text variant="subtext">2 hours ago</Text>
          </LabeledValue>

          <LabeledValue label="Team Members">
            <Text>12 active</Text>
          </LabeledValue>

          <LabeledValue label="Priority">
            <Badge theme="warn">High</Badge>
          </LabeledValue>

          <LabeledValue label="Repository">
            <Button variant="ghost" size="sm">
              View on GitHub
            </Button>
          </LabeledValue>

          <LabeledValue label="Deployment">
            <div className="flex items-center gap-2">
              <Badge theme="info" size="sm">
                v2.1.0
              </Badge>
              <Button variant="ghost" size="sm">
                Deploy
              </Button>
            </div>
          </LabeledValue>
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
        LabeledValue is versatile and works well in settings panels, profile
        pages, configuration displays, and any interface requiring labeled
        content presentation.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">API Documentation Panel</h4>
      <div className="p-4 border rounded-lg">
        <Text variant="h3" weight="stronger" className="mb-4">
          API Endpoint Details
        </Text>
        <div className="space-y-3 max-w-2xl">
          <LabeledValue label="Method">
            <Badge theme="info">POST</Badge>
          </LabeledValue>

          <LabeledValue label="Endpoint">
            <Code>/api/v1/users</Code>
          </LabeledValue>

          <LabeledValue label="Authentication">
            <Text>Bearer Token Required</Text>
          </LabeledValue>

          <LabeledValue label="Rate Limit">
            <Text>100 requests per minute</Text>
          </LabeledValue>

          <LabeledValue label="Response Format">
            <Text family="mono">application/json</Text>
          </LabeledValue>

          <LabeledValue label="Status Codes">
            <div className="space-y-1">
              <div className="flex items-center gap-2">
                <Badge theme="success" size="sm">
                  200
                </Badge>
                <Text variant="subtext">Success</Text>
              </div>
              <div className="flex items-center gap-2">
                <Badge theme="warn" size="sm">
                  400
                </Badge>
                <Text variant="subtext">Bad Request</Text>
              </div>
              <div className="flex items-center gap-2">
                <Badge theme="error" size="sm">
                  401
                </Badge>
                <Text variant="subtext">Unauthorized</Text>
              </div>
            </div>
          </LabeledValue>
        </div>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">System Information</h4>
      <div className="p-4 border rounded-lg">
        <Text variant="h3" weight="stronger" className="mb-4">
          Server Information
        </Text>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          <LabeledValue label="Hostname">
            <Code>api.example.com</Code>
          </LabeledValue>

          <LabeledValue label="IP Address">
            <Code>192.168.1.100</Code>
          </LabeledValue>

          <LabeledValue label="Operating System">
            <Text>Ubuntu 22.04 LTS</Text>
          </LabeledValue>

          <LabeledValue label="Uptime">
            <Text>15 days, 4 hours</Text>
          </LabeledValue>

          <LabeledValue label="Load Average">
            <Text family="mono">0.25, 0.30, 0.28</Text>
          </LabeledValue>

          <LabeledValue label="Memory Usage">
            <div className="flex items-center gap-2">
              <Text>6.2GB / 16GB</Text>
              <Badge theme="success" size="sm">
                39%
              </Badge>
            </div>
          </LabeledValue>
        </div>
      </div>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Best Practices:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Use clear, descriptive labels that identify the content</li>
        <li>Maintain consistent spacing in groups of labeled values</li>
        <li>
          Choose appropriate content types (text, badges, buttons) for the
          context
        </li>
        <li>Consider responsive layouts for groups of labeled values</li>
        <li>Use semantic HTML structure for accessibility</li>
        <li>Group related labeled values together logically</li>
      </ul>
    </div>
  </div>
)
