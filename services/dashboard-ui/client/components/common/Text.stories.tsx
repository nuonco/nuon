import { Text } from './Text'

export const Themes = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Text Themes</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        The{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          theme
        </code>{' '}
        prop controls the color scheme of text elements. Each theme includes
        appropriate colors and dark mode styling while maintaining accessibility
        contrast ratios.
      </p>
    </div>

    <div className="space-y-4">
      <div className="flex flex-col gap-4">
        <Text>Default text</Text>
        <Text theme="neutral">Neutral text</Text>
        <Text theme="info">Info text</Text>
        <Text theme="success">Success text</Text>
        <Text theme="warn">Warning text</Text>
        <Text theme="error">Error text</Text>
        <Text theme="brand">Brand text</Text>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-3 text-sm mt-6">
        <div>
          <strong>default:</strong> Uses default text color (no theme override)
        </div>
        <div>
          <strong>neutral:</strong> Cool grey colors for secondary information
        </div>
        <div>
          <strong>info:</strong> Blue colors for informational content
        </div>
        <div>
          <strong>success:</strong> Green colors for successful operations
        </div>
        <div>
          <strong>warn:</strong> Orange colors for warnings and cautions
        </div>
        <div>
          <strong>error:</strong> Red colors for error states and critical
          issues
        </div>
        <div>
          <strong>brand:</strong> Purple primary colors for Nuon platform
          branding
        </div>
      </div>
    </div>
  </div>
)

export const Families = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Text Font Families</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        The{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          family
        </code>{' '}
        prop controls the font family used for the text. Monospace headings
        automatically get adjusted letter spacing for better readability.
      </p>
    </div>

    <div className="space-y-4">
      <div className="flex flex-col gap-4">
        <Text family="sans">Sans serif family (Inter)</Text>
        <Text family="mono">Monospace family (Hack)</Text>
        <Text family="mono" variant="h2">
          Mono heading with adjusted tracking
        </Text>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-3 text-sm mt-6">
        <div>
          <strong>sans:</strong> Inter font for general text content (default)
        </div>
        <div>
          <strong>mono:</strong> Hack font for code, technical content, and data
          display
        </div>
      </div>

      <div className="text-sm text-gray-600 dark:text-gray-400 mt-4 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
        <strong>Special Behavior:</strong>
        <ul className="mt-2 space-y-1 list-disc list-inside">
          <li>
            Monospace headings (h1, h2, h3) automatically get reduced letter
            spacing
          </li>
          <li>
            Sans serif uses optimized letter spacing for different text sizes
          </li>
          <li>Both families support full dark mode theming</li>
        </ul>
      </div>
    </div>
  </div>
)

export const Weights = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Text Weights</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        The{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          weight
        </code>{' '}
        prop controls the font weight of text elements. These weights are
        optimized for the Inter and Hack font families used in the design
        system.
      </p>
    </div>

    <div className="space-y-4">
      <div className="flex flex-col gap-4">
        <Text weight="normal">Normal weight (400)</Text>
        <Text weight="strong">Strong weight (500)</Text>
        <Text weight="stronger">Stronger weight (600)</Text>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-3 text-sm mt-6">
        <div>
          <strong>normal:</strong> Font weight 400 for body text (default)
        </div>
        <div>
          <strong>strong:</strong> Font weight 500 for emphasis and labels
        </div>
        <div>
          <strong>stronger:</strong> Font weight 600 for headings and strong
          emphasis
        </div>
      </div>

      <div className="text-sm text-gray-600 dark:text-gray-400 mt-4 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
        <strong>Usage Guidelines:</strong>
        <ul className="mt-2 space-y-1 list-disc list-inside">
          <li>
            Use <strong>normal</strong> for most body text and content
          </li>
          <li>
            Use <strong>strong</strong> for labels, form fields, and moderate
            emphasis
          </li>
          <li>
            Use <strong>stronger</strong> for headings and critical information
          </li>
        </ul>
      </div>
    </div>
  </div>
)

export const Variants = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Text Variants</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        The{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          variant
        </code>{' '}
        prop controls the typography scale including font size, line height, and
        letter spacing. Each variant is optimized for specific content types and
        hierarchy levels.
      </p>
    </div>

    <div className="space-y-4">
      <div className="flex flex-col gap-4">
        <Text variant="h1">H1 variant - Main page titles</Text>
        <Text variant="h2">H2 variant - Section headings</Text>
        <Text variant="h3">H3 variant - Subsection headings</Text>
        <Text variant="base">Base variant - Standard text content</Text>
        <Text variant="body">Body variant - Main body text (default)</Text>
        <Text variant="subtext">Subtext variant - Secondary information</Text>
        <Text variant="label">Label variant - Form labels and small text</Text>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-3 text-sm mt-6">
        <div>
          <strong>h1:</strong> 34px, 40px line height, -0.8px tracking
        </div>
        <div>
          <strong>h2:</strong> 24px, 30px line height, -0.8px tracking
        </div>
        <div>
          <strong>h3:</strong> 18px, 27px line height, -0.2px tracking
        </div>
        <div>
          <strong>base:</strong> 16px, 24px line height, -0.2px tracking
        </div>
        <div>
          <strong>body:</strong> 14px, 24px line height, -0.2px tracking
          (default)
        </div>
        <div>
          <strong>subtext:</strong> 12px, 17px line height, -0.2px tracking
        </div>
        <div>
          <strong>label:</strong> 11px, 14px line height, -0.2px tracking
        </div>
      </div>

      <div className="text-sm text-gray-600 dark:text-gray-400 mt-4 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
        <strong>Typography Hierarchy:</strong>
        <ul className="mt-2 space-y-1 list-disc list-inside">
          <li>
            Larger variants (h1, h2) use increased negative tracking for optical
            balance
          </li>
          <li>All variants include optimized line heights for readability</li>
          <li>Letter spacing adjusts automatically based on font family</li>
        </ul>
      </div>
    </div>
  </div>
)

export const SemanticRoles = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Semantic HTML Roles</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        The{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          role
        </code>{' '}
        prop automatically renders the appropriate HTML element while
        maintaining visual styling. This ensures proper semantic markup and
        accessibility support for screen readers.
      </p>
    </div>

    <div className="space-y-4">
      <div className="flex flex-col gap-4">
        <Text role="heading" level={1} variant="h1">
          Semantic heading level 1 (renders as &lt;h1&gt;)
        </Text>
        <Text role="heading" level={2} variant="h2">
          Semantic heading level 2 (renders as &lt;h2&gt;)
        </Text>
        <Text role="heading" level={3} variant="h3">
          Semantic heading level 3 (renders as &lt;h3&gt;)
        </Text>
        <Text role="paragraph" variant="body">
          This is a semantic paragraph element (renders as &lt;p&gt;)
        </Text>
        <Text role="code" family="mono" variant="body">
          const num = 400 (renders as &lt;code&gt;)
        </Text>
        <Text role="time" variant="subtext">
          2024-01-15T10:30:00Z (renders as &lt;time&gt;)
        </Text>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-3 text-sm mt-6">
        <div>
          <strong>heading + level:</strong> Renders as h1-h6 with proper ARIA
          attributes
        </div>
        <div>
          <strong>paragraph:</strong> Renders as &lt;p&gt; tag for block-level
          text
        </div>
        <div>
          <strong>code:</strong> Renders as &lt;code&gt; for inline code
          snippets
        </div>
        <div>
          <strong>time:</strong> Renders as &lt;time&gt; for dates and
          timestamps
        </div>
      </div>

      <div className="text-sm text-gray-600 dark:text-gray-400 mt-4 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
        <strong>Accessibility Benefits:</strong>
        <ul className="mt-2 space-y-1 list-disc list-inside">
          <li>
            Screen readers can properly identify content structure and hierarchy
          </li>
          <li>Keyboard navigation respects semantic document outline</li>
          <li>
            Search engines better understand content meaning and importance
          </li>
          <li>
            Visual styling remains completely customizable via variant and theme
            props
          </li>
        </ul>
      </div>
    </div>
  </div>
)
