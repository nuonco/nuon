/* eslint-disable react/no-unescaped-entities */
import { Divider } from './Divider'
import { Text } from './Text'
import { Button } from './Button'
import { Icon } from './Icon'

export const BasicUsage = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Basic Divider Usage</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Dividers provide visual separation between content sections. They render
        as horizontal lines with optional text labels. The component
        automatically handles light and dark mode styling with proper contrast
        ratios.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Simple Divider</h4>
      <div className="w-full max-w-md border p-4 rounded">
        <Text>Content above divider</Text>
        <Divider />
        <Text>Content below divider</Text>
      </div>
      <Text variant="subtext" theme="neutral">
        Basic horizontal line separator without text
      </Text>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Key Features:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Responsive design that adapts to container width</li>
        <li>Automatic uppercase transformation for divider words</li>
        <li>Shadow and border styling for text labels</li>
        <li>Dark mode support with appropriate contrast</li>
      </ul>
    </div>
  </div>
)

export const WithLabels = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Dividers with Labels</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Use the{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          dividerWord
        </code>{' '}
        prop to add text labels to dividers. The text is automatically converted
        to uppercase and styled with a background, border, and shadow for better
        visibility against the divider line.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Common Divider Labels</h4>
      <div className="space-y-6 max-w-md">
        <div className="border p-4 rounded">
          <Text>Login with email</Text>
          <Divider dividerWord="OR" />
          <Text>Login with OAuth provider</Text>
        </div>

        <div className="border p-4 rounded">
          <Text>Primary content section</Text>
          <Divider dividerWord="AND" />
          <Text>Secondary content section</Text>
        </div>

        <div className="border p-4 rounded">
          <Text>Free tier features</Text>
          <Divider dividerWord="PLUS" />
          <Text>Premium tier features</Text>
        </div>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Alternative Patterns</h4>
      <div className="space-y-4 max-w-md">
        <div className="border p-4 rounded">
          <Text>Step 1: Configuration</Text>
          <Divider dividerWord="NEXT" />
          <Text>Step 2: Deployment</Text>
        </div>

        <div className="border p-4 rounded">
          <Text>Public repositories</Text>
          <Divider dividerWord="PRIVATE" />
          <Text>Private repositories</Text>
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
        Dividers are commonly used in forms, authentication interfaces, and
        content sections. Here are typical usage patterns for different
        interface contexts.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Authentication Form</h4>
      <div className="max-w-md border rounded-lg p-6 space-y-4">
        <div className="space-y-2">
          <Text variant="label" weight="strong">
            Email
          </Text>
          <div className="w-full px-3 py-2 border rounded-md bg-gray-50 dark:bg-gray-800">
            <Text variant="subtext" theme="neutral">
              user@example.com
            </Text>
          </div>
        </div>
        <div className="space-y-2">
          <Text variant="label" weight="strong">
            Password
          </Text>
          <div className="w-full px-3 py-2 border rounded-md bg-gray-50 dark:bg-gray-800">
            <Text variant="subtext" theme="neutral">
              ••••••••••
            </Text>
          </div>
        </div>
        <Button variant="primary" className="w-full">
          Sign In
        </Button>

        <Divider dividerWord="OR" />

        <Button variant="secondary" className="w-full">
          <Icon variant="Globe" size="16" />
          Continue with Google
        </Button>
        <Button variant="secondary" className="w-full">
          <Icon variant="GithubLogo" size="16" />
          Continue with GitHub
        </Button>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Feature Comparison</h4>
      <div className="max-w-md border rounded-lg p-6 space-y-4">
        <div>
          <Text weight="stronger">Free Plan Features</Text>
          <div className="mt-2 space-y-1">
            <Text variant="subtext">• 5 projects included</Text>
            <Text variant="subtext">• 1GB storage per project</Text>
            <Text variant="subtext">• Community support</Text>
          </div>
        </div>

        <Divider dividerWord="UPGRADE" />

        <div>
          <Text weight="stronger">Pro Plan Features</Text>
          <div className="mt-2 space-y-1">
            <Text variant="subtext">• Unlimited projects</Text>
            <Text variant="subtext">• 100GB storage per project</Text>
            <Text variant="subtext">• Priority support</Text>
            <Text variant="subtext">• Advanced analytics</Text>
          </div>
        </div>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Multi-Step Process</h4>
      <div className="max-w-md border rounded-lg p-6 space-y-4">
        <div>
          <Text weight="stronger">Step 1: Repository Setup</Text>
          <Text variant="subtext" theme="neutral" className="mt-1">
            Configure your repository settings and choose deployment options.
          </Text>
        </div>

        <Divider dividerWord="THEN" />

        <div>
          <Text weight="stronger">Step 2: Environment Configuration</Text>
          <Text variant="subtext" theme="neutral" className="mt-1">
            Set up environment variables and deployment targets.
          </Text>
        </div>

        <Divider dividerWord="FINALLY" />

        <div>
          <Text weight="stronger">Step 3: Deploy Application</Text>
          <Text variant="subtext" theme="neutral" className="mt-1">
            Review settings and trigger your first deployment.
          </Text>
        </div>
      </div>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Best Practices:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Use dividers to separate logically different content sections</li>
        <li>Keep divider words short and meaningful (OR, AND, PLUS, etc.)</li>
        <li>
          Consider the visual weight - too many dividers can clutter the
          interface
        </li>
        <li>Ensure adequate spacing above and below dividers</li>
        <li>Use labeled dividers for alternative actions or workflow steps</li>
      </ul>
    </div>
  </div>
)
