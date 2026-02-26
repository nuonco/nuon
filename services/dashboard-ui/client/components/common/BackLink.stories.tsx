/* eslint-disable react/no-unescaped-entities */
import { BackLink } from './BackLink'
import { Text } from './Text'
import { Icon } from './Icon'

export const BasicUsage = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Basic BackLink Usage</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        BackLink provides a consistent way to navigate back to the previous page
        in the browser history. It automatically includes a left-pointing caret
        icon and uses the browser's back navigation functionality. The component
        includes proper hover and focus states.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Default BackLink</h4>
      <div className="p-4 border rounded-lg">
        <BackLink />
      </div>
      <Text variant="subtext" theme="neutral">
        Default text with left caret icon
      </Text>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Default Behavior:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Navigates to the previous page using browser history</li>
        <li>Includes hover and focus states with primary color scheme</li>
        <li>Automatically displays left caret icon and "Back" text</li>
        <li>Proper keyboard accessibility support</li>
      </ul>
    </div>
  </div>
)

export const CustomContent = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Custom BackLink Content</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        BackLink accepts custom children to override the default "Back" text and
        icon. You can provide any React content while maintaining the navigation
        behavior and styling.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Custom Text Examples</h4>
      <div className="space-y-3 p-4 border rounded-lg">
        <BackLink>Go back</BackLink>
        <BackLink>Return to dashboard</BackLink>
        <BackLink>← Previous page</BackLink>
        <BackLink>
          <Icon variant="House" size="16" />
          Back to home
        </BackLink>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Contextual BackLinks</h4>
      <div className="space-y-3 p-4 border rounded-lg">
        <BackLink>
          <Icon variant="CaretLeft" weight="bold" />
          Back to applications
        </BackLink>
        <BackLink>
          <Icon variant="CaretLeft" weight="bold" />
          Return to project list
        </BackLink>
        <BackLink>
          <Icon variant="ArrowLeft" size="16" />
          Back to settings
        </BackLink>
      </div>
    </div>
  </div>
)

export const TextVariants = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">BackLink Text Variants</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        BackLink inherits all Text component props, allowing you to customize
        the typography using{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          variant
        </code>
        ,{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          weight
        </code>
        , and other text properties while maintaining navigation functionality.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Different Text Variants</h4>
      <div className="space-y-3 p-4 border rounded-lg">
        <BackLink variant="subtext">Small back link</BackLink>
        <BackLink variant="base">Base back link</BackLink>
        <BackLink variant="body" weight="normal">
          Normal weight back link
        </BackLink>
        <BackLink variant="body" weight="stronger">
          Stronger weight back link
        </BackLink>
      </div>
    </div>

    <div className="grid grid-cols-1 md:grid-cols-2 gap-3 text-sm mt-6">
      <div>
        <strong>variant:</strong> Controls text size and typography scale
      </div>
      <div>
        <strong>weight:</strong> Controls font weight (normal, strong, stronger)
      </div>
    </div>
  </div>
)

export const UsageExamples = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Common Usage Patterns</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        BackLink is commonly used in page headers, form interfaces, and detail
        views where users need to navigate back to a previous context. Here are
        typical usage patterns.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Page Header</h4>
      <div className="border rounded-lg">
        <div className="p-4 border-b bg-gray-50 dark:bg-gray-800">
          <BackLink>Back to applications</BackLink>
        </div>
        <div className="p-4">
          <Text variant="h2" weight="stronger">
            Application Details
          </Text>
          <Text variant="subtext" theme="neutral">
            Manage settings and configuration for your application
          </Text>
        </div>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Form Context</h4>
      <div className="border rounded-lg p-4">
        <BackLink variant="subtext">Back to user list</BackLink>
        <div className="mt-4 space-y-4">
          <Text variant="h3" weight="stronger">
            Edit User Profile
          </Text>
          <div className="space-y-3">
            <div>
              <Text variant="label" weight="strong">
                Full Name
              </Text>
              <div className="mt-1 p-2 border rounded bg-gray-50 dark:bg-gray-800">
                John Doe
              </div>
            </div>
            <div>
              <Text variant="label" weight="strong">
                Email Address
              </Text>
              <div className="mt-1 p-2 border rounded bg-gray-50 dark:bg-gray-800">
                john@example.com
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Breadcrumb Alternative</h4>
      <div className="border rounded-lg p-4">
        <BackLink>
          <Icon variant="CaretLeft" weight="bold" />
          Projects / My Application
        </BackLink>
        <div className="mt-4">
          <Text variant="h3" weight="stronger">
            Deployment Configuration
          </Text>
          <Text variant="subtext" theme="neutral">
            Configure deployment settings for production environment
          </Text>
        </div>
      </div>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Best Practices:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Place BackLinks at the top of pages or forms</li>
        <li>Use contextual text that describes where users will return</li>
        <li>
          Consider using smaller text variants for less prominent placement
        </li>
        <li>Ensure BackLinks are easily discoverable and clickable</li>
        <li>Test navigation flow to ensure proper browser history behavior</li>
      </ul>
    </div>
  </div>
)
