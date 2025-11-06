/* eslint-disable react/no-unescaped-entities */
import { Avatar } from './Avatar'
import { Text } from './Text'

export const BasicUsage = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Basic Avatar Usage</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Avatars represent users and entities throughout the application. They
        automatically generate initials from names or display provided images
        with proper fallback behavior. All avatars include smooth transitions
        and responsive design support.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Name-based Avatar</h4>
      <div className="flex items-center gap-4">
        <Avatar name="John Doe" />
        <Text variant="subtext">Automatically generates "JD" initials</Text>
      </div>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Automatic Features:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Generates initials from full names automatically</li>
        <li>Handles both light and dark mode styling</li>
        <li>Smooth transitions and hover effects</li>
        <li>Proper contrast ratios for accessibility</li>
      </ul>
    </div>
  </div>
)

export const Sizes = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Avatar Sizes</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        The{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          size
        </code>{' '}
        prop controls the dimensions and typography of avatars. Each size is
        optimized for different contexts and use cases throughout the
        application.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">All Available Sizes</h4>
      <div className="flex items-center gap-6">
        <div className="text-center space-y-2">
          <Avatar name="John Doe" size="xs" />
          <Text variant="label" className="text-xs">
            XS (24px)
          </Text>
        </div>
        <div className="text-center space-y-2">
          <Avatar name="John Doe" size="sm" />
          <Text variant="label" className="text-xs">
            SM (28px)
          </Text>
        </div>
        <div className="text-center space-y-2">
          <Avatar name="John Doe" size="md" />
          <Text variant="label" className="text-xs">
            MD (32px)
          </Text>
        </div>
        <div className="text-center space-y-2">
          <Avatar name="John Doe" size="lg" />
          <Text variant="label" className="text-xs">
            LG (36px)
          </Text>
        </div>
        <div className="text-center space-y-2">
          <Avatar name="John Doe" size="xl" />
          <Text variant="label" className="text-xs">
            XL (40px)
          </Text>
        </div>
      </div>
    </div>

    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3 text-sm mt-6">
      <div>
        <strong>xs (24px):</strong> Compact listings and dense interfaces
      </div>
      <div>
        <strong>sm (28px):</strong> Secondary areas and sidebar elements
      </div>
      <div>
        <strong>md (32px):</strong> Standard size for most use cases (default)
      </div>
      <div>
        <strong>lg (36px):</strong> User profiles and prominent displays
      </div>
      <div>
        <strong>xl (40px):</strong> Headers, hero sections, and key user areas
      </div>
    </div>
  </div>
)

export const WithImage = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Avatar with Images</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Avatars can display user profile images when provided. Use the{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          src
        </code>{' '}
        prop for image URLs and{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          alt
        </code>{' '}
        for accessibility. Images are automatically cropped and scaled to fit.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Image Examples</h4>
      <div className="flex items-center gap-6">
        <div className="text-center space-y-2">
          <Avatar
            src="https://github.com/nat.png"
            alt="Nat Friedman"
            size="sm"
          />
          <Text variant="label" className="text-xs">
            Small Image
          </Text>
        </div>
        <div className="text-center space-y-2">
          <Avatar src="https://github.com/nat.png" alt="Nat Friedman" />
          <Text variant="label" className="text-xs">
            Default Image
          </Text>
        </div>
        <div className="text-center space-y-2">
          <Avatar
            src="https://github.com/nat.png"
            alt="Nat Friedman"
            size="lg"
          />
          <Text variant="label" className="text-xs">
            Large Image
          </Text>
        </div>
      </div>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Image Handling:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Automatic cropping and scaling using object-cover</li>
        <li>Fallback to initials if image fails to load</li>
        <li>Optimized with Next.js Image component for performance</li>
        <li>Proper alt text support for screen readers</li>
      </ul>
    </div>
  </div>
)

export const LoadingState = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Avatar Loading States</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Use the{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          isLoading
        </code>{' '}
        prop to show loading states while user data is being fetched. The
        loading state includes a subtle pulse animation and neutral styling.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Loading Examples</h4>
      <div className="flex items-center gap-6">
        <div className="text-center space-y-2">
          <Avatar isLoading size="sm" />
          <Text variant="label" className="text-xs">
            Small Loading
          </Text>
        </div>
        <div className="text-center space-y-2">
          <Avatar isLoading />
          <Text variant="label" className="text-xs">
            Default Loading
          </Text>
        </div>
        <div className="text-center space-y-2">
          <Avatar isLoading size="lg" />
          <Text variant="label" className="text-xs">
            Large Loading
          </Text>
        </div>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Loading vs Normal Comparison</h4>
      <div className="flex items-center gap-8">
        <div className="flex items-center gap-3">
          <Avatar name="John Doe" />
          <Text variant="subtext">Normal State</Text>
        </div>
        <div className="flex items-center gap-3">
          <Avatar isLoading />
          <Text variant="subtext">Loading State</Text>
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
        Avatars are used throughout the application in various contexts. Here
        are common patterns and recommended approaches for different scenarios.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">User Profile Section</h4>
      <div className="flex items-center gap-4 p-4 border rounded-lg">
        <Avatar src="https://github.com/nat.png" alt="John Doe" size="lg" />
        <div>
          <Text weight="strong">John Doe</Text>
          <Text variant="subtext" theme="neutral">
            Senior Developer
          </Text>
          <Text variant="subtext" theme="neutral">
            john@company.com
          </Text>
        </div>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">User List</h4>
      <div className="space-y-3">
        <div className="flex items-center gap-3 p-3 border rounded">
          <Avatar name="Alice Johnson" size="sm" />
          <div className="flex-1">
            <Text variant="subtext" weight="strong">
              Alice Johnson
            </Text>
            <Text variant="label" theme="neutral">
              Product Manager
            </Text>
          </div>
        </div>
        <div className="flex items-center gap-3 p-3 border rounded">
          <Avatar name="Bob Smith" size="sm" />
          <div className="flex-1">
            <Text variant="subtext" weight="strong">
              Bob Smith
            </Text>
            <Text variant="label" theme="neutral">
              UX Designer
            </Text>
          </div>
        </div>
        <div className="flex items-center gap-3 p-3 border rounded">
          <Avatar isLoading size="sm" />
          <div className="flex-1">
            <Text variant="subtext" theme="neutral">
              Loading user...
            </Text>
          </div>
        </div>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Navigation Header</h4>
      <div className="flex items-center justify-between p-4 border rounded-lg bg-gray-50 dark:bg-gray-800">
        <Text weight="stronger">Dashboard</Text>
        <div className="flex items-center gap-3">
          <Text variant="subtext">Welcome back, John</Text>
          <Avatar src="https://github.com/nat.png" alt="John Doe" size="sm" />
        </div>
      </div>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Best Practices:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Use appropriate sizes based on the interface context</li>
        <li>Provide meaningful alt text for profile images</li>
        <li>Implement loading states for better user experience</li>
        <li>Ensure proper contrast in both light and dark modes</li>
        <li>Consider fallback initials for users without profile images</li>
      </ul>
    </div>
  </div>
)
