/* eslint-disable react/no-unescaped-entities */
import { Button } from './Button'
import { Icon } from './Icon'

export const Variants = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Button Variants</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        The{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          variant
        </code>{' '}
        prop controls the visual style and semantic meaning of the button. Each
        variant includes appropriate colors, hover states, focus indicators, and
        dark mode styling.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">All Variants</h4>
      <div className="flex flex-wrap gap-4 items-center">
        <Button variant="primary">Primary</Button>
        <Button variant="secondary">Secondary</Button>
        <Button variant="danger">Danger</Button>
        <Button variant="ghost">Ghost</Button>
        <Button variant="tab">Tab</Button>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Active State</h4>
      <div className="flex flex-wrap gap-4 items-center">
        <Button variant="primary" isActive>
          Primary Active
        </Button>
        <Button variant="secondary" isActive>
          Secondary Active
        </Button>
        <Button variant="danger" isActive>
          Danger Active
        </Button>
        <Button variant="ghost" isActive>
          Ghost Active
        </Button>
        <Button variant="tab" isActive>
          Tab Active
        </Button>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Disabled State</h4>
      <div className="flex flex-wrap gap-4 items-center">
        <Button variant="primary" disabled>
          Primary Disabled
        </Button>
        <Button variant="secondary" disabled>
          Secondary Disabled
        </Button>
        <Button variant="danger" disabled>
          Danger Disabled
        </Button>
        <Button variant="ghost" disabled>
          Ghost Disabled
        </Button>
        <Button variant="tab" disabled>
          Tab Disabled
        </Button>
      </div>
    </div>

    <div className="grid grid-cols-1 md:grid-cols-2 gap-3 text-sm mt-6">
      <div>
        <strong>primary:</strong> Purple filled button for main actions and CTAs
      </div>
      <div>
        <strong>secondary:</strong> Outlined button for secondary actions
        (default)
      </div>
      <div>
        <strong>danger:</strong> Red outlined button for destructive actions
      </div>
      <div>
        <strong>ghost:</strong> Transparent button with minimal styling for
        subtle actions
      </div>
      <div>
        <strong>tab:</strong> Special variant with bottom border for tab
        navigation
      </div>
    </div>
  </div>
)

export const Sizes = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Button Sizes</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        The{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          size
        </code>{' '}
        prop controls the dimensions, padding, and typography of the button.
        Default size is{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          md
        </code>{' '}
        if no size prop is provided.
      </p>
    </div>

    <div className="space-y-4">
      <div className="flex flex-wrap gap-4 items-center">
        <Button size="xs">Extra Small</Button>
        <Button size="sm">Small</Button>
        <Button size="md">Medium</Button>
        <Button size="lg">Large</Button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-3 text-sm mt-4">
        <div>
          <strong>xs:</strong> 12px text, 16px height, minimal padding
        </div>
        <div>
          <strong>sm:</strong> 12px text, 24px height, 8px horizontal padding
        </div>
        <div>
          <strong>md:</strong> 14px text, 32px height, 12px horizontal padding
          (default)
        </div>
        <div>
          <strong>lg:</strong> 14px text, 36px height, 12px horizontal padding
        </div>
      </div>
    </div>
  </div>
)

export const Links = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Button as Link</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        When you provide an{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          href
        </code>{' '}
        prop, the Button automatically renders as a link while maintaining
        button styling. Internal links use Next.js Link component, while
        external links use standard anchor tags with proper security attributes.
      </p>
    </div>

    <div className="space-y-4">
      <div className="flex flex-wrap gap-4 items-center">
        <Button href="/">Internal Link</Button>
        <Button href="/dashboard" variant="primary">
          Dashboard Link
        </Button>
        <Button
          href="https://nuon.co"
          target="_blank"
          rel="noopener noreferrer"
        >
          External Link
        </Button>
        <Button
          href="https://docs.nuon.co"
          target="_blank"
          rel="noopener noreferrer"
          variant="secondary"
        >
          Documentation
        </Button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-3 text-sm mt-4">
        <div>
          <strong>Internal Links:</strong> Use &quot;/&quot; prefix, rendered
          with Next.js Link for client-side navigation
        </div>
        <div>
          <strong>External Links:</strong> Use full URLs, rendered as anchor
          tags with security attributes
        </div>
        <div>
          <strong>Accessibility:</strong> Automatically includes proper ARIA
          attributes and keyboard navigation
        </div>
        <div>
          <strong>Styling:</strong> Maintains all button variants, sizes, and
          states when used as links
        </div>
      </div>
    </div>
  </div>
)

export const WithIcons = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Buttons with Icons</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Buttons automatically support icons through CSS detection. Include{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          Icon
        </code>{' '}
        components as children and the button will apply proper spacing and
        alignment. Icons can be placed before or after text.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Icons with Text</h4>
      <div className="flex flex-wrap gap-4 items-center">
        <Button variant="primary">
          <Icon variant="Plus" size="16" />
          Create New
        </Button>
        <Button variant="secondary">
          <Icon variant="Download" size="16" />
          Download
        </Button>
        <Button variant="danger">
          <Icon variant="Trash" size="16" />
          Delete
        </Button>
        <Button variant="ghost">
          <Icon variant="Gear" size="16" />
          Settings
        </Button>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Icon Only Buttons</h4>
      <div className="flex flex-wrap gap-4 items-center">
        <Button className="!p-2" variant="primary">
          <Icon variant="Plus" size="18" />
        </Button>
        <Button className="!p-2" variant="secondary">
          <Icon variant="Eraser" size="18" />
        </Button>
        <Button className="!p-2" variant="danger">
          <Icon variant="X" size="18" />
        </Button>
        <Button className="!p-2" variant="ghost">
          <Icon variant="DotsThreeVertical" size="18" weight="bold" />
        </Button>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Icon Position Examples</h4>
      <div className="flex flex-wrap gap-4 items-center">
        <Button variant="primary">
          <Icon variant="ArrowLeft" size="16" />
          Back
        </Button>
        <Button variant="primary">
          Next
          <Icon variant="ArrowRight" size="16" />
        </Button>
        <Button variant="secondary">
          <Icon variant="ArrowSquareOut" size="16" />
          Open External
        </Button>
        <Button variant="secondary">
          <Icon variant="FloppyDisk" size="16" />
          Save Changes
        </Button>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Different Sizes with Icons</h4>
      <div className="flex flex-wrap gap-4 items-center">
        <Button size="xs" variant="ghost">
          <Icon variant="Info" size="12" />
        </Button>
        <Button size="sm">
          <Icon variant="CheckCircle" size="14" />
          Confirm
        </Button>
        <Button size="md" variant="primary">
          <Icon variant="Upload" size="16" />
          Upload File
        </Button>
        <Button size="lg" variant="secondary">
          <Icon variant="Database" size="18" />
          View Database
        </Button>
      </div>
    </div>

    <div className="grid grid-cols-1 md:grid-cols-2 gap-3 text-sm mt-6">
      <div>
        <strong>Automatic Spacing:</strong> CSS selector detects SVG icons and
        applies proper gap
      </div>
      <div>
        <strong>Icon Sizing:</strong> Match icon size to button size (xs=12px,
        sm=14px, md=16px, lg=18px)
      </div>
      <div>
        <strong>Icon Position:</strong> Place before text for leading icons,
        after text for trailing icons
      </div>
      <div>
        <strong>Icon-Only:</strong> Use custom padding for square icon-only
        buttons
      </div>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Best Practices:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Use consistent icon sizes that match the button size</li>
        <li>Place icons before text for actions, after text for navigation</li>
        <li>Ensure icon-only buttons have accessible labels or tooltips</li>
        <li>Choose icons that clearly represent the button's action</li>
        <li>
          Use the same icon variant across similar actions for consistency
        </li>
      </ul>
    </div>
  </div>
)
