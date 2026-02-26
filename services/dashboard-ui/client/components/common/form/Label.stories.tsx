/* eslint-disable react/no-unescaped-entities */
import { Label } from './Label'
import { Input } from './Input'
import { Text } from '@/components/common/Text'

export default {
  title: 'Forms/Label',
  component: Label,
}

export const BasicUsage = () => (
  <div className="space-y-6 max-w-md">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Basic Label Usage</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        The Label component provides a semantic wrapper for form inputs with
        proper accessibility attributes. It automatically handles the htmlFor
        attribute when used with form controls.
      </p>
    </div>

    <div className="space-y-4">
      <div>
        <Label htmlFor="basic-input">
          <Text variant="body" weight="strong">Basic Label</Text>
        </Label>
        <input
          id="basic-input"
          type="text"
          placeholder="Enter text here..."
          className="mt-1 w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-primary-500"
        />
      </div>

      <div>
        <Label htmlFor="email-input">
          <Text variant="body" weight="strong">Email Address</Text>
        </Label>
        <input
          id="email-input"
          type="email"
          placeholder="user@example.com"
          className="mt-1 w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-primary-500"
        />
      </div>

      <div>
        <Label htmlFor="password-input">
          <Text variant="body" weight="strong">Password</Text>
        </Label>
        <input
          id="password-input"
          type="password"
          placeholder="••••••••"
          className="mt-1 w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-primary-500"
        />
      </div>
    </div>
  </div>
)

export const WithText = () => (
  <div className="space-y-6 max-w-md">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Labels with Text Component</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Labels work seamlessly with the Text component to provide consistent
        typography and styling. You can use different text variants, weights,
        and themes within labels.
      </p>
    </div>

    <div className="space-y-4">
      <div>
        <Label htmlFor="username" className="block">
          <Text variant="body" weight="strong" theme="default">
            Username
          </Text>
        </Label>
        <input
          id="username"
          type="text"
          placeholder="Enter username"
          className="mt-1 w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-primary-500"
        />
      </div>

      <div>
        <Label htmlFor="description" className="block">
          <Text variant="h3" weight="strong" theme="brand">
            Project Description
          </Text>
        </Label>
        <textarea
          id="description"
          placeholder="Describe your project..."
          rows={3}
          className="mt-1 w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-primary-500"
        />
      </div>

      <div>
        <Label htmlFor="tags" className="block">
          <Text variant="body" weight="strong" theme="neutral">
            Tags (Optional)
          </Text>
        </Label>
        <input
          id="tags"
          type="text"
          placeholder="react, typescript, ui"
          className="mt-1 w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-primary-500"
        />
      </div>
    </div>

    <div className="grid grid-cols-1 gap-3 text-sm mt-6">
      <div>
        <strong>Text Integration:</strong> Use Text component for consistent typography
      </div>
      <div>
        <strong>Weight Options:</strong> normal, strong, stronger for emphasis
      </div>
      <div>
        <strong>Theme Support:</strong> All Text themes available (brand, neutral, error, etc.)
      </div>
    </div>
  </div>
)

export const CustomStyling = () => (
  <div className="space-y-6 max-w-md">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Custom Label Styling</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Labels accept className and all standard HTML attributes for custom
        styling. They can be styled for different layouts and visual hierarchies.
      </p>
    </div>

    <div className="space-y-4">
      <div>
        <Label 
          htmlFor="styled-1" 
          className="block p-2 bg-blue-50 dark:bg-blue-900/20 rounded-md border border-blue-200 dark:border-blue-800"
        >
          <Text variant="body" weight="strong" theme="info">
            Highlighted Label
          </Text>
        </Label>
        <input
          id="styled-1"
          type="text"
          placeholder="Enter text..."
          className="mt-1 w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-primary-500"
        />
      </div>

      <div>
        <Label 
          htmlFor="styled-2" 
          className="flex items-center justify-between p-2 bg-gray-50 dark:bg-gray-800 rounded-md"
        >
          <Text variant="body" weight="strong">
            Label with Badge
          </Text>
          <span className="px-2 py-1 text-xs bg-green-100 dark:bg-green-900/30 text-green-800 dark:text-green-300 rounded-full">
            Required
          </span>
        </Label>
        <input
          id="styled-2"
          type="text"
          placeholder="Required field..."
          required
          className="mt-1 w-full px-3 py-2 border border-red-300 rounded-md focus:ring-2 focus:ring-red-500"
        />
      </div>

      <div>
        <Label 
          htmlFor="styled-3" 
          className="block border-l-4 border-primary-500 pl-3"
        >
          <Text variant="body" weight="strong" theme="brand">
            Accented Label
          </Text>
        </Label>
        <input
          id="styled-3"
          type="text"
          placeholder="Important field..."
          className="mt-1 w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-primary-500"
        />
      </div>
    </div>

    <div className="grid grid-cols-1 gap-3 text-sm mt-6">
      <div>
        <strong>Background Styling:</strong> Use background colors for emphasis
      </div>
      <div>
        <strong>Border Accents:</strong> Add visual hierarchy with borders
      </div>
      <div>
        <strong>Flexbox Layouts:</strong> Create complex label layouts
      </div>
      <div>
        <strong>Badge Integration:</strong> Combine with status indicators
      </div>
    </div>
  </div>
)

export const InlineLabels = () => (
  <div className="space-y-6 max-w-md">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Inline Label Patterns</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Labels can be used in inline layouts for compact forms, checkboxes,
        radio buttons, and toggle controls. Proper spacing and alignment
        ensure good usability.
      </p>
    </div>

    <div className="space-y-4">
      <div className="flex items-center gap-3">
        <Label htmlFor="inline-1" className="flex items-center gap-2 cursor-pointer">
          <input
            id="inline-1"
            type="checkbox"
            className="accent-primary-600"
          />
          <Text variant="body">
            Accept terms and conditions
          </Text>
        </Label>
      </div>

      <div className="flex items-center gap-3">
        <Label htmlFor="inline-2" className="flex items-center gap-2 cursor-pointer">
          <input
            id="inline-2"
            type="radio"
            name="option"
            value="option1"
            className="accent-primary-600"
          />
          <Text variant="body">
            Option 1
          </Text>
        </Label>
      </div>

      <div className="flex items-center gap-3">
        <Label htmlFor="inline-3" className="flex items-center gap-2 cursor-pointer">
          <input
            id="inline-3"
            type="radio"
            name="option"
            value="option2"
            className="accent-primary-600"
          />
          <Text variant="body">
            Option 2
          </Text>
        </Label>
      </div>

      <div className="flex items-center gap-4">
        <Label htmlFor="toggle" className="text-sm font-medium">
          Enable feature:
        </Label>
        <Label htmlFor="toggle" className="relative inline-flex cursor-pointer">
          <input
            id="toggle"
            type="checkbox"
            className="sr-only peer"
          />
          <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-primary-300 dark:peer-focus:ring-primary-800 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-primary-600"></div>
        </Label>
      </div>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Inline Label Best Practices:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Use cursor-pointer for clickable labels</li>
        <li>Ensure adequate spacing between elements</li>
        <li>Align items properly for visual consistency</li>
        <li>Test keyboard navigation and screen reader compatibility</li>
        <li>Group related options with consistent styling</li>
      </ul>
    </div>
  </div>
)

export const WithErrors = () => (
  <div className="space-y-6 max-w-md">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Labels with Error States</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Labels can be styled to indicate error states and provide visual
        feedback for form validation. Use consistent color and styling patterns
        for error indication.
      </p>
    </div>

    <div className="space-y-4">
      <div>
        <Label htmlFor="error-1" className="block">
          <Text variant="body" weight="strong" theme="error">
            Username *
          </Text>
        </Label>
        <input
          id="error-1"
          type="text"
          placeholder="Enter username"
          aria-invalid="true"
          aria-describedby="error-1-msg"
          className="mt-1 w-full px-3 py-2 border border-red-500 rounded-md focus:ring-2 focus:ring-red-500"
        />
        <Text id="error-1-msg" variant="subtext" theme="error" className="mt-1 block">
          Username is required
        </Text>
      </div>

      <div>
        <Label htmlFor="error-2" className="block p-2 bg-red-50 dark:bg-red-900/20 rounded-md border border-red-200 dark:border-red-800">
          <Text variant="body" weight="strong" theme="error">
            Email Address *
          </Text>
        </Label>
        <input
          id="error-2"
          type="email"
          placeholder="user@example.com"
          aria-invalid="true"
          aria-describedby="error-2-msg"
          className="mt-1 w-full px-3 py-2 border border-red-500 rounded-md focus:ring-2 focus:ring-red-500"
        />
        <Text id="error-2-msg" variant="subtext" theme="error" className="mt-1 block">
          Please enter a valid email address
        </Text>
      </div>

      <div>
        <Label htmlFor="warning" className="block">
          <Text variant="body" weight="strong" theme="warn">
            Password
          </Text>
        </Label>
        <input
          id="warning"
          type="password"
          placeholder="••••••••"
          className="mt-1 w-full px-3 py-2 border border-orange-300 rounded-md focus:ring-2 focus:ring-orange-500"
        />
        <Text variant="subtext" theme="warn" className="mt-1 block">
          Password should be at least 8 characters
        </Text>
      </div>
    </div>

    <div className="grid grid-cols-1 gap-3 text-sm mt-6">
      <div>
        <strong>Error Styling:</strong> Use theme="error" for error text
      </div>
      <div>
        <strong>Background Hints:</strong> Colored backgrounds for severe errors
      </div>
      <div>
        <strong>ARIA Attributes:</strong> Link errors with aria-describedby
      </div>
      <div>
        <strong>Warning States:</strong> Use theme="warn" for warnings
      </div>
    </div>
  </div>
)