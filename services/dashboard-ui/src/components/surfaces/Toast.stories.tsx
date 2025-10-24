/* eslint-disable react/no-unescaped-entities */
import { ToastProvider } from '@/providers/toast-provider'
import { useToast } from '@/hooks/use-toast'
import { Button } from '@/components/common/Button'
import { Toast } from './Toast'

const ToastTrigger = ({ theme, children }) => {
  const { addToast } = useToast()
  return (
    <Button
      onClick={() => {
        addToast(
          <Toast theme={theme} heading={`${theme} toast`}>
            This is a {theme} toast.
          </Toast>
        )
      }}
    >
      {children}
    </Button>
  )
}

export const Default = () => (
  <ToastProvider>
    <div className="space-y-6">
      <div className="space-y-3">
        <h3 className="text-lg font-semibold">Default Toast</h3>
        <p className="text-sm text-gray-600 dark:text-gray-400">
          Click the button below to trigger a default toast notification. Toasts
          appear in the bottom-right corner and automatically dismiss after 5
          seconds unless paused by hovering.
        </p>
      </div>

      <div>
        <ToastTrigger theme="default">Show Default Toast</ToastTrigger>
      </div>

      <div className="text-sm text-gray-600 dark:text-gray-400 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
        <strong>Toast Features:</strong>
        <ul className="mt-2 space-y-1 list-disc list-inside">
          <li>
            Auto-dismiss after 5 seconds (customizable via <code>timeout</code>{' '}
            prop)
          </li>
          <li>Pause timer on hover to allow reading</li>
          <li>Manual close button appears on hover</li>
          <li>Proper ARIA live regions for screen readers</li>
          <li>Fixed positioning in bottom-right corner</li>
        </ul>
      </div>
    </div>
  </ToastProvider>
)

export const Themes = () => (
  <ToastProvider>
    <div className="space-y-6">
      <div className="space-y-3">
        <h3 className="text-lg font-semibold">Toast Themes</h3>
        <p className="text-sm text-gray-600 dark:text-gray-400">
          The{' '}
          <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
            theme
          </code>{' '}
          prop controls the color scheme and accessibility behavior of the
          toast. Each theme includes appropriate colors, dark mode styling, and
          semantic ARIA attributes for screen readers.
        </p>
      </div>

      <div className="space-y-4">
        <div className="flex flex-wrap gap-4">
          <ToastTrigger theme="brand">Brand</ToastTrigger>
          <ToastTrigger theme="error">Error</ToastTrigger>
          <ToastTrigger theme="warn">Warn</ToastTrigger>
          <ToastTrigger theme="info">Info</ToastTrigger>
          <ToastTrigger theme="success">Success</ToastTrigger>
          <ToastTrigger theme="neutral">Neutral</ToastTrigger>
          <ToastTrigger theme="default">Default</ToastTrigger>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-3 text-sm mt-6">
          <div>
            <strong>brand:</strong> Purple primary colors for Nuon platform
            notifications
          </div>
          <div>
            <strong>error:</strong> Red colors with assertive ARIA for critical
            issues
          </div>
          <div>
            <strong>warn:</strong> Orange colors with assertive ARIA for
            warnings
          </div>
          <div>
            <strong>info:</strong> Blue colors with polite ARIA for
            informational content
          </div>
          <div>
            <strong>success:</strong> Green colors with polite ARIA for
            successful operations
          </div>
          <div>
            <strong>neutral:</strong> Cool grey colors with polite ARIA for
            neutral information
          </div>
          <div>
            <strong>default:</strong> Standard grey colors with polite ARIA
            (default theme)
          </div>
        </div>

        <div className="text-sm text-gray-600 dark:text-gray-400 mt-4 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
          <strong>Accessibility Behavior:</strong>
          <ul className="mt-2 space-y-1 list-disc list-inside">
            <li>
              <strong>Error &amp; Warn:</strong> Use <code>role="alert"</code>{' '}
              and <code>aria-live="assertive"</code> for immediate attention
            </li>
            <li>
              <strong>Other themes:</strong> Use <code>role="status"</code> and{' '}
              <code>aria-live="polite"</code> for non-urgent notifications
            </li>
            <li>
              <strong>Screen readers:</strong> Automatically announce toast
              content based on urgency level
            </li>
          </ul>
        </div>
      </div>
    </div>
  </ToastProvider>
)
