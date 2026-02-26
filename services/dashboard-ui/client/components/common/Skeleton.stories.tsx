/* eslint-disable react/no-unescaped-entities */
import { Skeleton } from './Skeleton'
import { Text } from './Text'

export const BasicUsage = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Basic Skeleton Usage</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Skeleton components provide loading placeholders that mimic the
        structure of content being loaded. They help reduce perceived loading
        time and provide visual feedback during data fetching. The component
        uses smooth animations and responsive design.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Single Line Skeleton</h4>
      <div className="p-4 border rounded">
        <Skeleton />
      </div>
      <Text variant="subtext" theme="neutral">
        Default single line with automatic width and height
      </Text>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Key Features:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Smooth pulse animation that respects user preferences</li>
        <li>Automatic dark mode support with appropriate contrast</li>
        <li>Responsive design that adapts to container width</li>
        <li>Customizable dimensions and multiple line support</li>
      </ul>
    </div>
  </div>
)

export const MultipleLines = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Multiple Line Skeletons</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Use the{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          lines
        </code>{' '}
        prop to create multi-line skeleton placeholders. This is useful for
        paragraph text, lists, and other multi-line content structures.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Different Line Counts</h4>
      <div className="space-y-6">
        <div className="p-4 border rounded">
          <Text variant="label" weight="strong" className="mb-3">
            2 Lines
          </Text>
          <Skeleton lines={2} />
        </div>

        <div className="p-4 border rounded">
          <Text variant="label" weight="strong" className="mb-3">
            3 Lines
          </Text>
          <Skeleton lines={3} />
        </div>

        <div className="p-4 border rounded">
          <Text variant="label" weight="strong" className="mb-3">
            5 Lines
          </Text>
          <Skeleton lines={5} />
        </div>
      </div>
    </div>
  </div>
)

export const CustomDimensions = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Custom Dimensions</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Skeleton dimensions can be customized using{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          width
        </code>{' '}
        and{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          height
        </code>{' '}
        props. Width can be an array to vary line widths for more natural
        text-like appearance.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Variable Width Lines</h4>
      <div className="space-y-4">
        <div className="p-4 border rounded">
          <Text variant="label" weight="strong" className="mb-3">
            Paragraph Style
          </Text>
          <Skeleton lines={3} width={['100%', '85%', '60%']} />
        </div>

        <div className="p-4 border rounded">
          <Text variant="label" weight="strong" className="mb-3">
            Short Lines
          </Text>
          <Skeleton lines={2} width={['50%', '75%']} height="2rem" />
        </div>

        <div className="p-4 border rounded">
          <Text variant="label" weight="strong" className="mb-3">
            Custom Heights
          </Text>
          <Skeleton lines={1} width="30%" height="3rem" />
        </div>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Specific Use Cases</h4>
      <div className="space-y-4">
        <div className="p-4 border rounded">
          <Text variant="label" weight="strong" className="mb-3">
            Button Skeleton
          </Text>
          <Skeleton width="120px" height="40px" />
        </div>

        <div className="p-4 border rounded">
          <Text variant="label" weight="strong" className="mb-3">
            Title Skeleton
          </Text>
          <Skeleton width="60%" height="1.5rem" />
        </div>

        <div className="p-4 border rounded">
          <Text variant="label" weight="strong" className="mb-3">
            Image Placeholder
          </Text>
          <Skeleton width="200px" height="150px" />
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
        Skeletons are commonly used in loading states for cards, lists,
        profiles, and content areas. Here are typical patterns for different
        interface contexts.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Card Loading State</h4>
      <div className="p-6 border rounded-lg space-y-4">
        <div className="flex items-center gap-4">
          <Skeleton width="48px" height="48px" />
          <div className="flex-1">
            <Skeleton width="60%" height="1.2rem" />
            <div className="mt-2">
              <Skeleton width="40%" height="1rem" />
            </div>
          </div>
        </div>
        <div>
          <Skeleton lines={3} width={['100%', '95%', '80%']} />
        </div>
        <div className="flex gap-2">
          <Skeleton width="80px" height="32px" />
          <Skeleton width="60px" height="32px" />
        </div>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">List Item Loading</h4>
      <div className="space-y-3">
        <div className="flex items-center gap-3 p-3 border rounded">
          <Skeleton width="32px" height="32px" />
          <div className="flex-1">
            <Skeleton width="70%" height="1rem" />
            <div className="mt-1">
              <Skeleton width="50%" height="0.875rem" />
            </div>
          </div>
          <Skeleton width="24px" height="24px" />
        </div>

        <div className="flex items-center gap-3 p-3 border rounded">
          <Skeleton width="32px" height="32px" />
          <div className="flex-1">
            <Skeleton width="85%" height="1rem" />
            <div className="mt-1">
              <Skeleton width="60%" height="0.875rem" />
            </div>
          </div>
          <Skeleton width="24px" height="24px" />
        </div>

        <div className="flex items-center gap-3 p-3 border rounded">
          <Skeleton width="32px" height="32px" />
          <div className="flex-1">
            <Skeleton width="75%" height="1rem" />
            <div className="mt-1">
              <Skeleton width="45%" height="0.875rem" />
            </div>
          </div>
          <Skeleton width="24px" height="24px" />
        </div>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Table Loading State</h4>
      <div className="border rounded-lg overflow-hidden">
        <div className="bg-gray-50 dark:bg-gray-800 p-4 border-b">
          <Skeleton width="150px" height="1.25rem" />
        </div>
        <div className="divide-y">
          <div className="p-4 grid grid-cols-4 gap-4">
            <Skeleton width="80%" />
            <Skeleton width="60%" />
            <Skeleton width="90%" />
            <Skeleton width="50%" />
          </div>
          <div className="p-4 grid grid-cols-4 gap-4">
            <Skeleton width="70%" />
            <Skeleton width="85%" />
            <Skeleton width="75%" />
            <Skeleton width="65%" />
          </div>
          <div className="p-4 grid grid-cols-4 gap-4">
            <Skeleton width="90%" />
            <Skeleton width="55%" />
            <Skeleton width="80%" />
            <Skeleton width="70%" />
          </div>
        </div>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Profile Loading State</h4>
      <div className="p-6 border rounded-lg">
        <div className="flex items-start gap-4">
          <Skeleton width="80px" height="80px" />
          <div className="flex-1 space-y-3">
            <Skeleton width="200px" height="1.5rem" />
            <Skeleton width="150px" height="1rem" />
            <div className="space-y-2">
              <Skeleton lines={2} width={['100%', '85%']} />
            </div>
            <div className="flex gap-2">
              <Skeleton width="100px" height="32px" />
              <Skeleton width="80px" height="32px" />
            </div>
          </div>
        </div>
      </div>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Best Practices:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Match skeleton structure to actual content layout</li>
        <li>Use varied line widths to mimic natural text patterns</li>
        <li>Include skeletons for images, buttons, and interactive elements</li>
        <li>Keep animation subtle to avoid distraction</li>
        <li>Replace skeletons promptly once content loads</li>
        <li>Consider skeleton duration - not too short or too long</li>
      </ul>
    </div>
  </div>
)
