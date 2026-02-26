/* eslint-disable react/no-unescaped-entities */
import { ToastProvider } from '@/providers/toast-provider'
import { useToast } from '@/hooks/use-toast'
import { Button } from '@/components/common/Button'
import { Text } from '@/components/common/Text'
import { Toast } from './Toast'

const ToastTrigger = ({ theme, children, heading, content }) => {
  const { addToast } = useToast()
  return (
    <Button
      onClick={() => {
        addToast(
          <Toast theme={theme} heading={heading || `${theme} toast`}>
            {content || `This is a ${theme} toast notification.`}
          </Toast>
        )
      }}
    >
      {children}
    </Button>
  )
}

export const BasicUsage = () => (
  <ToastProvider>
    <div className="space-y-6">
      <div className="space-y-3">
        <h3 className="text-lg font-semibold">Basic Toast Usage</h3>
        <p className="text-sm text-gray-600 dark:text-gray-400">
          The Toast component displays temporary notifications that appear in the
          bottom-right corner of the screen. Toasts automatically dismiss after a
          configurable timeout and can be manually dismissed by users. They provide
          important feedback for user actions and system events.
        </p>
      </div>

      <div className="space-y-4">
        <h4 className="text-sm font-medium">Simple Toast Example</h4>
        <div className="p-4 border rounded-lg">
          <ToastTrigger 
            theme="default"
            heading="Welcome!"
            content="This is a simple toast notification that demonstrates the basic functionality."
          >
            Show Default Toast
          </ToastTrigger>
        </div>
        <Text variant="subtext" theme="neutral">
          Click the button above to see a toast notification in action
        </Text>
      </div>

      <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
        <strong>Key Features:</strong>
        <ul className="mt-2 space-y-1 list-disc list-inside">
          <li>Automatic dismissal after 5 seconds (customizable timeout)</li>
          <li>Pause timer on hover to allow reading</li>
          <li>Manual close button for user control</li>
          <li>Proper ARIA live regions for screen reader accessibility</li>
          <li>Fixed positioning with smooth animations</li>
          <li>Multiple toast stacking with proper spacing</li>
        </ul>
      </div>
    </div>
  </ToastProvider>
)

export const ToastThemes = () => (
  <ToastProvider>
    <div className="space-y-6">
      <div className="space-y-3">
        <h3 className="text-lg font-semibold">Toast Themes</h3>
        <p className="text-sm text-gray-600 dark:text-gray-400">
          The{' '}
          <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
            theme
          </code>{' '}
          prop controls the visual appearance and semantic meaning of toast notifications.
          Each theme uses appropriate colors, semantic meaning, and ARIA attributes
          to convey the right level of urgency and context to users.
        </p>
      </div>

      <div className="space-y-4">
        <h4 className="text-sm font-medium">Available Themes</h4>
        <div className="p-4 border rounded-lg">
          <div className="flex flex-wrap gap-3 mb-6">
            <ToastTrigger theme="brand" heading="Brand Update" content="Your brand settings have been updated successfully.">Brand</ToastTrigger>
            <ToastTrigger theme="success" heading="Success!" content="Your changes have been saved successfully.">Success</ToastTrigger>
            <ToastTrigger theme="info" heading="Information" content="Here's some helpful information for you.">Info</ToastTrigger>
            <ToastTrigger theme="warn" heading="Warning" content="Please review your settings before continuing.">Warning</ToastTrigger>
            <ToastTrigger theme="error" heading="Error" content="An error occurred while processing your request.">Error</ToastTrigger>
            <ToastTrigger theme="neutral" heading="Notification" content="You have a new notification.">Neutral</ToastTrigger>
            <ToastTrigger theme="default" heading="Default" content="This is a default notification.">Default</ToastTrigger>
          </div>
        </div>
        <Text variant="subtext" theme="neutral">
          Click any button above to see the corresponding toast theme in action
        </Text>
      </div>

      <div className="space-y-4">
        <h4 className="text-sm font-medium">Theme Descriptions</h4>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
          <div className="p-3 border rounded-lg">
            <strong className="text-purple-600 dark:text-purple-400">brand:</strong>
            <span className="block text-gray-600 dark:text-gray-400 mt-1">
              Purple primary colors for Nuon platform-specific notifications and branding
            </span>
          </div>
          <div className="p-3 border rounded-lg">
            <strong className="text-green-600 dark:text-green-400">success:</strong>
            <span className="block text-gray-600 dark:text-gray-400 mt-1">
              Green colors for successful operations, completions, and positive feedback
            </span>
          </div>
          <div className="p-3 border rounded-lg">
            <strong className="text-blue-600 dark:text-blue-400">info:</strong>
            <span className="block text-gray-600 dark:text-gray-400 mt-1">
              Blue colors for informational content and helpful tips
            </span>
          </div>
          <div className="p-3 border rounded-lg">
            <strong className="text-orange-600 dark:text-orange-400">warn:</strong>
            <span className="block text-gray-600 dark:text-gray-400 mt-1">
              Orange colors for warnings that require user attention
            </span>
          </div>
          <div className="p-3 border rounded-lg">
            <strong className="text-red-600 dark:text-red-400">error:</strong>
            <span className="block text-gray-600 dark:text-gray-400 mt-1">
              Red colors for critical issues and error states
            </span>
          </div>
          <div className="p-3 border rounded-lg">
            <strong className="text-gray-600 dark:text-gray-400">neutral:</strong>
            <span className="block text-gray-600 dark:text-gray-400 mt-1">
              Cool grey colors for neutral information and general notifications
            </span>
          </div>
          <div className="p-3 border rounded-lg">
            <strong className="text-gray-600 dark:text-gray-400">default:</strong>
            <span className="block text-gray-600 dark:text-gray-400 mt-1">
              Standard grey colors - the default theme when no theme is specified
            </span>
          </div>
        </div>
      </div>

      <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
        <strong>Accessibility Behavior:</strong>
        <ul className="mt-2 space-y-1 list-disc list-inside">
          <li>
            <strong>Error &amp; Warn:</strong> Use <code>role="alert"</code> and{' '}
            <code>aria-live="assertive"</code> for immediate screen reader attention
          </li>
          <li>
            <strong>Other themes:</strong> Use <code>role="status"</code> and{' '}
            <code>aria-live="polite"</code> for non-urgent notifications
          </li>
          <li>
            <strong>Dark mode:</strong> All themes automatically adapt colors for dark mode compatibility
          </li>
          <li>
            <strong>Semantic meaning:</strong> Colors and ARIA attributes convey proper urgency levels
          </li>
        </ul>
      </div>
    </div>
  </ToastProvider>
)

export const ToastUsageExamples = () => (
  <ToastProvider>
    <div className="space-y-6">
      <div className="space-y-3">
        <h3 className="text-lg font-semibold">Toast Usage Examples</h3>
        <p className="text-sm text-gray-600 dark:text-gray-400">
          Toasts are commonly used throughout applications to provide immediate
          feedback for user actions, system events, and important notifications.
          They should be used sparingly and only for information that requires
          immediate user attention.
        </p>
      </div>

      <div className="space-y-4">
        <h4 className="text-sm font-medium">Form Submission Feedback</h4>
        <div className="p-4 border rounded-lg space-y-3">
          <Text variant="base" className="font-medium">User Registration Form</Text>
          <div className="flex gap-3">
            <ToastTrigger 
              theme="success" 
              heading="Registration Successful!" 
              content="Welcome to Nuon! Your account has been created and you're now logged in."
            >
              Register Success
            </ToastTrigger>
            <ToastTrigger 
              theme="error" 
              heading="Registration Failed" 
              content="Email address is already in use. Please try with a different email."
            >
              Register Error
            </ToastTrigger>
          </div>
        </div>
      </div>

      <div className="space-y-4">
        <h4 className="text-sm font-medium">System Status Updates</h4>
        <div className="p-4 border rounded-lg space-y-3">
          <Text variant="base" className="font-medium">Deployment Pipeline</Text>
          <div className="flex gap-3">
            <ToastTrigger 
              theme="info" 
              heading="Deployment Started" 
              content="Your application deployment has begun. This may take a few minutes."
            >
              Deploy Started
            </ToastTrigger>
            <ToastTrigger 
              theme="success" 
              heading="Deployment Complete" 
              content="Your application has been successfully deployed to production."
            >
              Deploy Success
            </ToastTrigger>
            <ToastTrigger 
              theme="warn" 
              heading="Deployment Warning" 
              content="Deployment completed with warnings. Check logs for details."
            >
              Deploy Warning
            </ToastTrigger>
          </div>
        </div>
      </div>

      <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
        <strong>Best Practices:</strong>
        <ul className="mt-2 space-y-1 list-disc list-inside">
          <li>Use toasts for temporary, non-critical information that doesn't require user action</li>
          <li>Keep toast messages concise and actionable</li>
          <li>Choose appropriate themes that match the semantic meaning of the message</li>
          <li>For critical errors that require user action, consider using modals instead</li>
          <li>Don't overwhelm users with too many simultaneous toasts</li>
          <li>Ensure toast content is accessible and readable by screen readers</li>
        </ul>
      </div>
    </div>
  </ToastProvider>
)
