/* eslint-disable react/no-unescaped-entities */
import { Badge } from './Badge'

export const Themes = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Badge Themes</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        The <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">theme</code> prop controls the color scheme of the badge. Each theme includes proper dark mode styling and maintains accessibility contrast ratios.
      </p>
    </div>
    
    <div className="space-y-4">
      <div className="flex flex-wrap gap-4 items-center">
        <Badge theme="brand">Brand</Badge>
        <Badge theme="error">Error</Badge>
        <Badge theme="warn">Warn</Badge>
        <Badge theme="info">Info</Badge>
        <Badge theme="success">Success</Badge>
        <Badge theme="neutral">Neutral</Badge>
        <Badge theme="default">Default</Badge>
      </div>
      
      <div className="grid grid-cols-1 md:grid-cols-2 gap-3 text-sm">
        <div><strong>brand:</strong> Purple primary colors for Nuon platform branding</div>
        <div><strong>error:</strong> Red colors for error states and critical issues</div>
        <div><strong>warn:</strong> Orange colors for warnings and cautions</div>
        <div><strong>info:</strong> Blue colors for informational content</div>
        <div><strong>success:</strong> Green colors for successful operations</div>
        <div><strong>neutral:</strong> Cool grey colors for neutral information</div>
        <div><strong>default:</strong> Standard grey colors (default if no theme specified)</div>
      </div>
    </div>
  </div>
)

export const Sizes = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Badge Sizes</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        The <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">size</code> prop controls the dimensions and typography of the badge. All sizes use -0.2px letter spacing for improved readability.
      </p>
    </div>
    
    <div className="space-y-4">
      <div className="flex gap-4 items-center">
        <Badge size="sm">Small</Badge>
        <Badge size="md">Medium</Badge>
        <Badge size="lg">Large</Badge>
      </div>
      
      <div className="grid grid-cols-1 md:grid-cols-3 gap-3 text-sm">
        <div><strong>sm:</strong> 11px text, 14px line height, 8px/2px padding</div>
        <div><strong>md:</strong> 12px text, 17px line height, 8px/2px padding</div>
        <div><strong>lg:</strong> 12px text, 17px line height, 12px/4px padding (default)</div>
      </div>
    </div>
  </div>
)

export const Variants = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Badge Variants</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        The <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">variant</code> prop controls the visual style and typography of the badge.
      </p>
    </div>
    
    <div className="space-y-4">
      <div className="flex gap-4 items-center">
        <Badge variant="default">Default variant</Badge>
        <Badge variant="code">Code variant</Badge>
      </div>
      
      <div className="grid grid-cols-1 md:grid-cols-2 gap-3 text-sm">
        <div><strong>default:</strong> Sans-serif font with fully rounded corners (default)</div>
        <div><strong>code:</strong> Monospace font with moderately rounded corners for technical content</div>
      </div>
      
      <div className="text-sm text-gray-600 dark:text-gray-400 mt-3">
        <strong>Use case:</strong> The <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">code</code> variant is ideal for displaying version numbers, API endpoints, or technical identifiers.
      </div>
    </div>
  </div>
)
