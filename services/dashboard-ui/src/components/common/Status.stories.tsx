/* eslint-disable react/no-unescaped-entities */
import { Status } from './Status'

export const Variants = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Status Variants</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        The{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          variant
        </code>{' '}
        prop controls the visual presentation of the status indicator. Each
        variant serves different UI contexts while maintaining semantic meaning
        through status types and colors.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">All Variants</h4>
      <div className="flex flex-wrap gap-6 items-center">
        <div className="space-y-2">
          <div className="text-xs font-medium text-gray-500">Default</div>
          <Status status="success" variant="default" />
        </div>
        <div className="space-y-2">
          <div className="text-xs font-medium text-gray-500">Badge</div>
          <Status status="success" variant="badge" />
        </div>
        <div className="space-y-2">
          <div className="text-xs font-medium text-gray-500">Timeline</div>
          <Status status="success" variant="timeline" />
        </div>
      </div>
    </div>

    <div className="grid grid-cols-1 md:grid-cols-3 gap-3 text-sm mt-6">
      <div>
        <strong>default:</strong> Simple dot indicator with text, minimal
        styling for inline use
      </div>
      <div>
        <strong>badge:</strong> Bordered container with indicator and text,
        great for tags and labels
      </div>
      <div>
        <strong>timeline:</strong> Icon-based indicator with themed background,
        ideal for process steps
      </div>
    </div>
  </div>
)

export const StatusTypes = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Status Types</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        The{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          status
        </code>{' '}
        prop determines the semantic meaning and color scheme. Status text is
        automatically generated from the status value, or you can provide custom
        children content.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Default Variant</h4>
      <div className="flex flex-wrap gap-4 items-center">
        <Status status="default" />
        <Status status="success" />
        <Status status="error" />
        <Status status="warn" />
        <Status status="info" />
        <Status status="brand" />
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Badge Variant</h4>
      <div className="flex flex-wrap gap-4 items-center">
        <Status status="default" variant="badge" />
        <Status status="success" variant="badge" />
        <Status status="error" variant="badge" />
        <Status status="warn" variant="badge" />
        <Status status="info" variant="badge" />
        <Status status="brand" variant="badge" />
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Timeline Variant</h4>
      <div className="flex flex-wrap gap-4 items-center">
        <Status status="default" variant="timeline" />
        <Status status="success" variant="timeline" />
        <Status status="error" variant="timeline" />
        <Status status="warn" variant="timeline" />
        <Status status="info" variant="timeline" />
        <Status status="brand" variant="timeline" />
      </div>
    </div>

    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3 text-sm mt-6">
      <div>
        <strong>default:</strong> Neutral gray color for inactive or pending
        states
      </div>
      <div>
        <strong>success:</strong> Green color indicating completed or successful
        states
      </div>
      <div>
        <strong>error:</strong> Red color for failed, error, or critical states
      </div>
      <div>
        <strong>warn:</strong> Orange color for warning or cautionary states
      </div>
      <div>
        <strong>info:</strong> Blue color for informational or in-progress
        states
      </div>
      <div>
        <strong>brand:</strong> Primary brand color for special or highlighted
        states
      </div>
    </div>
  </div>
)

export const CustomContent = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Custom Content</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        By default, status text is automatically generated from the status value
        (e.g., "in-progress" becomes "In Progress"). You can override this by
        providing custom children content, or hide text entirely with the{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          isWithoutText
        </code>{' '}
        prop.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Auto-generated Text</h4>
      <div className="flex flex-wrap gap-4 items-center">
        <Status status="in-progress" />
        <Status status="deployment-ready" variant="badge" />
        <Status status="build-failed" variant="timeline" />
        <Status status="pending-approval" />
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Custom Text</h4>
      <div className="flex flex-wrap gap-4 items-center">
        <Status status="success">Deploy Complete</Status>
        <Status status="error" variant="badge">
          Failed to Connect
        </Status>
        <Status status="warn" variant="timeline">
          Needs Attention
        </Status>
        <Status status="info">Running Health Check</Status>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Icon Only (No Text)</h4>
      <div className="flex flex-wrap gap-4 items-center">
        <Status status="success" isWithoutText />
        <Status status="error" variant="badge" isWithoutText />
        <Status status="warn" variant="timeline" isWithoutText />
        <Status status="info" isWithoutText />
      </div>
    </div>

    <div className="grid grid-cols-1 md:grid-cols-2 gap-3 text-sm mt-6">
      <div>
        <strong>Auto-generation:</strong> Status values are converted from
        kebab-case to sentence case automatically
      </div>
      <div>
        <strong>Custom Children:</strong> Override default text with any React
        content for specific use cases
      </div>
      <div>
        <strong>Icon Only:</strong> Use isWithoutText prop for compact displays
        where space is limited
      </div>
      <div>
        <strong>Accessibility:</strong> Status meaning is conveyed through color
        and text for screen readers
      </div>
    </div>
  </div>
)

export const IconSizing = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Icon Sizing</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        The timeline variant displays icons inside the status indicator. Use the{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          iconSize
        </code>{' '}
        prop to control icon dimensions. Default size is 18px, and the indicator
        container automatically adjusts to accommodate the icon.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Different Icon Sizes</h4>
      <div className="flex flex-wrap gap-6 items-center">
        <div className="space-y-2">
          <div className="text-xs font-medium text-gray-500">Small (14px)</div>
          <Status status="success" variant="timeline" iconSize={14} />
        </div>
        <div className="space-y-2">
          <div className="text-xs font-medium text-gray-500">
            Default (18px)
          </div>
          <Status status="success" variant="timeline" iconSize={18} />
        </div>
        <div className="space-y-2">
          <div className="text-xs font-medium text-gray-500">Large (24px)</div>
          <Status status="success" variant="timeline" iconSize={24} />
        </div>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Icons in Different States</h4>
      <div className="flex flex-wrap gap-4 items-center">
        <Status status="success" variant="timeline" iconSize={20} />
        <Status status="error" variant="timeline" iconSize={20} />
        <Status status="warn" variant="timeline" iconSize={20} />
        <Status status="info" variant="timeline" iconSize={20} />
      </div>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Note:</strong> Icon sizing only applies to the timeline variant.
      Default and badge variants use simple dot indicators without icons. The
      icon variant is automatically selected based on the status type.
    </div>
  </div>
)

export const UsageExamples = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Common Usage Patterns</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Status components are commonly used throughout the application for
        displaying states, progress, and system feedback. Here are typical usage
        patterns for different contexts.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Deployment Status</h4>
      <div className="space-y-2">
        <div className="flex items-center gap-3">
          <Status status="pending" variant="badge" />
          <span className="text-sm text-gray-600">Deployment queued</span>
        </div>
        <div className="flex items-center gap-3">
          <Status status="in-progress" variant="badge">
            Deploying
          </Status>
          <span className="text-sm text-gray-600">
            Currently deploying to production
          </span>
        </div>
        <div className="flex items-center gap-3">
          <Status status="success" variant="badge">
            Live
          </Status>
          <span className="text-sm text-gray-600">Successfully deployed</span>
        </div>
        <div className="flex items-center gap-3">
          <Status status="error" variant="badge">
            Failed
          </Status>
          <span className="text-sm text-gray-600">
            Deployment failed - check logs
          </span>
        </div>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Process Timeline</h4>
      <div className="space-y-3">
        <div className="flex items-center gap-3">
          <Status status="success" variant="timeline" />
          <span className="text-sm">Repository cloned</span>
        </div>
        <div className="flex items-center gap-3">
          <Status status="success" variant="timeline" />
          <span className="text-sm">Dependencies installed</span>
        </div>
        <div className="flex items-center gap-3">
          <Status status="info" variant="timeline">
            Building
          </Status>
          <span className="text-sm">Building application...</span>
        </div>
        <div className="flex items-center gap-3">
          <Status status="default" variant="timeline" />
          <span className="text-sm text-gray-500">Deploy to production</span>
        </div>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Inline Status Indicators</h4>
      <div className="space-y-2">
        <div className="flex items-center gap-2 text-sm">
          <span>API Service</span>
          <Status status="success" isWithoutText />
          <span className="text-gray-500">Healthy</span>
        </div>
        <div className="flex items-center gap-2 text-sm">
          <span>Database</span>
          <Status status="warn" isWithoutText />
          <span className="text-gray-500">High latency</span>
        </div>
        <div className="flex items-center gap-2 text-sm">
          <span>Cache</span>
          <Status status="error" isWithoutText />
          <span className="text-gray-500">Connection failed</span>
        </div>
      </div>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Best Practices:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Use consistent status types across similar features</li>
        <li>
          Choose variants based on context: badge for labels, timeline for
          processes, default for inline
        </li>
        <li>Provide meaningful text that describes the current state</li>
        <li>Consider using isWithoutText for compact layouts</li>
        <li>
          Ensure color meanings are consistent throughout your application
        </li>
      </ul>
    </div>
  </div>
)
