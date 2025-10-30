/* eslint-disable react/no-unescaped-entities */
import { Loading } from './Loading'
import { Button } from './Button'
import { Card } from './Card'
import { Text } from './Text'

export const Variants = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Loading Variants</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        The{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          variant
        </code>{' '}
        prop controls the size of the loading spinner. Both variants include
        smooth rotation animation and inherit the current text color for
        automatic theming.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Size Variants</h4>
      <div className="flex items-center gap-6">
        <div className="space-y-2 text-center">
          <div className="text-xs font-medium text-gray-500">Default (20px)</div>
          <Loading variant="default" />
        </div>
        <div className="space-y-2 text-center">
          <div className="text-xs font-medium text-gray-500">Large (40px)</div>
          <Loading variant="large" />
        </div>
      </div>
    </div>

    <div className="grid grid-cols-1 md:grid-cols-2 gap-3 text-sm mt-6">
      <div>
        <strong>default:</strong> 20x20px spinner for inline use and compact
        spaces
      </div>
      <div>
        <strong>large:</strong> 40x40px spinner for prominent loading states
        and page-level indicators
      </div>
    </div>
  </div>
)

export const StrokeWidth = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Stroke Width Options</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        The{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          strokeWidth
        </code>{' '}
        prop controls the thickness of the spinner lines. Thicker strokes
        provide better visibility in low contrast environments or when used
        over complex backgrounds.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Stroke Variations</h4>
      <div className="flex items-center gap-8">
        <div className="space-y-2 text-center">
          <div className="text-xs font-medium text-gray-500">Default (2px)</div>
          <Loading strokeWidth="default" />
        </div>
        <div className="space-y-2 text-center">
          <div className="text-xs font-medium text-gray-500">Thick (4px)</div>
          <Loading strokeWidth="thick" />
        </div>
        <div className="space-y-2 text-center">
          <div className="text-xs font-medium text-gray-500">Large + Thick</div>
          <Loading variant="large" strokeWidth="thick" />
        </div>
      </div>
    </div>

    <div className="grid grid-cols-1 md:grid-cols-2 gap-3 text-sm mt-6">
      <div>
        <strong>default:</strong> 2px stroke width for standard visibility and
        clean appearance
      </div>
      <div>
        <strong>thick:</strong> 4px stroke width for enhanced visibility and
        bold presentation
      </div>
    </div>
  </div>
)

export const ColorInheritance = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Color Inheritance</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Loading spinners automatically inherit the current text color using{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          currentColor
        </code>
        . This allows them to adapt to any theme or context without additional
        styling.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Theme Colors</h4>
      <div className="flex flex-wrap gap-6 items-center">
        <div className="space-y-2 text-center">
          <div className="text-xs font-medium text-gray-500">Default</div>
          <div className="text-gray-900 dark:text-gray-100">
            <Loading />
          </div>
        </div>
        <div className="space-y-2 text-center">
          <div className="text-xs font-medium text-gray-500">Blue</div>
          <div className="text-blue-500">
            <Loading />
          </div>
        </div>
        <div className="space-y-2 text-center">
          <div className="text-xs font-medium text-gray-500">Green</div>
          <div className="text-green-500">
            <Loading />
          </div>
        </div>
        <div className="space-y-2 text-center">
          <div className="text-xs font-medium text-gray-500">Red</div>
          <div className="text-red-500">
            <Loading />
          </div>
        </div>
        <div className="space-y-2 text-center">
          <div className="text-xs font-medium text-gray-500">Purple</div>
          <div className="text-purple-500">
            <Loading variant="large" />
          </div>
        </div>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Background Contexts</h4>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div className="bg-gray-100 dark:bg-gray-800 p-4 rounded-lg text-center">
          <Loading variant="large" />
          <Text className="mt-2" variant="subtext" theme="neutral">
            On gray background
          </Text>
        </div>
        <div className="bg-blue-100 dark:bg-blue-950 text-blue-700 dark:text-blue-300 p-4 rounded-lg text-center">
          <Loading variant="large" />
          <Text className="mt-2" variant="subtext">
            On colored background
          </Text>
        </div>
      </div>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Automatic Adaptation:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Inherits text color from parent elements automatically</li>
        <li>Works with dark mode theming without additional configuration</li>
        <li>Adapts to custom color schemes and branded interfaces</li>
        <li>Maintains proper contrast ratios in different contexts</li>
      </ul>
    </div>
  </div>
)

export const ContextualUsage = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Contextual Usage Patterns</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Loading spinners work in various contexts throughout the application.
        Here are common usage patterns and recommended approaches for different
        scenarios.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Inline with Text</h4>
      <div className="space-y-3">
        <div className="flex items-center gap-2">
          <Loading />
          <Text>Loading data...</Text>
        </div>
        <div className="flex items-center gap-2">
          <Text>Processing request</Text>
          <Loading />
        </div>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Button States</h4>
      <div className="flex flex-wrap gap-3">
        <Button disabled className="opacity-75">
          <Loading />
          Saving...
        </Button>
        <Button variant="primary" disabled className="opacity-75">
          <Loading />
          Deploying
        </Button>
        <Button variant="secondary" disabled className="opacity-75 justify-center w-24">
          <Loading />
        </Button>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Card Loading States</h4>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <Card>
          <div className="text-center py-8">
            <Loading variant="large" />
            <Text className="mt-3" variant="subtext" theme="neutral">
              Loading content...
            </Text>
          </div>
        </Card>
        <Card>
          <div className="flex items-center justify-between">
            <Text weight="strong">System Status</Text>
            <Loading />
          </div>
          <Text variant="subtext" theme="neutral">
            Checking all services...
          </Text>
        </Card>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Page-Level Loading</h4>
      <div className="border-2 border-dashed border-gray-300 dark:border-gray-600 rounded-lg p-12 text-center">
        <Loading variant="large" strokeWidth="thick" />
        <Text className="mt-4" variant="base" weight="strong">
          Loading Application
        </Text>
        <Text className="mt-1" variant="subtext" theme="neutral">
          Please wait while we prepare your dashboard...
        </Text>
      </div>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Usage Guidelines:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Use default variant for inline and compact loading states</li>
        <li>Use large variant for prominent loading indicators and page-level states</li>
        <li>Include descriptive text when the loading context isn't obvious</li>
        <li>Consider thick stroke width for better visibility over complex backgrounds</li>
        <li>Disable interactive elements while loading to prevent user confusion</li>
      </ul>
    </div>
  </div>
)
