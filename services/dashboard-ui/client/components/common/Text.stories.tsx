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

export const FlexLayout = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Flex layout</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        The{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          flex
        </code>{' '}
        prop adds <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">inline-flex items-center gap-1.5</code> for
        inline icon + text patterns. Override the default gap via className.
      </p>
    </div>

    <div className="space-y-4">
      <div className="flex flex-col gap-4">
        <Text flex>
          <span>🔵</span> Default gap (1.5)
        </Text>
        <Text flex className="gap-1">
          <span>🟢</span> Tighter gap (1)
        </Text>
        <Text flex className="gap-4">
          <span>🟡</span> Wider gap (4)
        </Text>
        <Text flex theme="info">
          <span>ℹ️</span> Info theme with flex
        </Text>
        <Text flex variant="h3" weight="stronger">
          <span>⚡</span> Heading variant with flex
        </Text>
      </div>

      <div className="text-sm text-gray-600 dark:text-gray-400 mt-4 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
        <strong>Usage:</strong>
        <ul className="mt-2 space-y-1 list-disc list-inside">
          <li>Use for icon + text pairs instead of <code>className="!flex items-center gap-1"</code></li>
          <li>Default gap is <code>gap-1.5</code>, overridable via className</li>
          <li>Works with any variant, theme, or weight</li>
        </ul>
      </div>
    </div>
  </div>
)

export const Nowrap = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Text wrapping</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        The{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          nowrap
        </code>{' '}
        prop prevents text from wrapping. By default, Text uses{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">text-wrap</code>.
      </p>
    </div>

    <div className="space-y-4">
      <div className="flex flex-col gap-4 max-w-[200px] border border-dashed border-gray-300 dark:border-gray-600 p-4 rounded">
        <Text>This text will wrap when it overflows the container because text-wrap is the default</Text>
        <Text nowrap>This text will not wrap and will overflow the container instead</Text>
        <Text nowrap className="truncate">This text will truncate with an ellipsis instead of wrapping</Text>
      </div>
    </div>
  </div>
)

export const AsElement = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Element override</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        The{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          as
        </code>{' '}
        prop renders an explicit HTML element regardless of role or variant.
        Useful when you want heading styling on a non-heading element, or need a
        block element without a semantic role.
      </p>
    </div>

    <div className="space-y-4">
      <div className="flex flex-col gap-4">
        <Text as="p" variant="h3" weight="stronger">
          Heading style on a &lt;p&gt; element
        </Text>
        <Text as="div" className="flex items-center justify-between">
          <span>Div element</span>
          <span>with flex layout</span>
        </Text>
        <Text as="label" weight="strong">
          Label element with strong weight
        </Text>
      </div>

      <div className="text-sm text-gray-600 dark:text-gray-400 mt-4 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
        <strong>Precedence:</strong> <code>as</code> takes priority over <code>role</code> and <code>level</code> for
        element resolution. Styling props (variant, weight, theme) still apply normally.
      </div>
    </div>
  </div>
)

export const LevelHeadings = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Level-based headings</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        The{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          level
        </code>{' '}
        prop alone now implies a heading element, so you no longer need{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          role="heading"
        </code>{' '}
        alongside it. This lets you decouple visual variant from semantic level.
      </p>
    </div>

    <div className="space-y-4">
      <div className="flex flex-col gap-4">
        <Text variant="h3" weight="stronger" level={1}>
          Visual h3, semantic &lt;h1&gt; (level alone)
        </Text>
        <Text variant="h3" weight="stronger" level={2}>
          Visual h3, semantic &lt;h2&gt; (level alone)
        </Text>
        <Text role="heading" level={3} variant="h3">
          role="heading" + level still works
        </Text>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-3 text-sm mt-6">
        <div>
          <strong>level alone:</strong> Renders as h1-h6 with heading role and aria-level
        </div>
        <div>
          <strong>role="heading" + level:</strong> Same result, backward compatible
        </div>
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
