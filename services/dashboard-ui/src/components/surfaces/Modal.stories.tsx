/* eslint-disable react/no-unescaped-entities */
import { Modal } from './Modal'
import { SurfacesProvider } from '@/providers/surfaces-provider'
import { useSurfaces } from '@/hooks/use-surfaces'
import { Button } from '@/components/common/Button'
import { Text } from '@/components/common/Text'
import { Card } from '@/components/common/Card'
import { Icon } from '@/components/common/Icon'

// Simple modal demo component
const SimpleModalDemo = () => {
  const { addModal } = useSurfaces()

  const openModal = () => {
    addModal(
      <Modal heading="Welcome to Modals">
        <div className="p-6">
          <Text className="mb-4">
            This is a simple modal example that demonstrates basic functionality.
            Modals overlay the main content and capture user focus until dismissed.
          </Text>
          <Text variant="subtext" theme="neutral">
            You can close this modal by clicking the X button, pressing Escape, or clicking outside.
          </Text>
        </div>
      </Modal>
    )
  }

  return <Button onClick={openModal}>Open Simple Modal</Button>
}

// Modal variants demo component
const ModalVariantsDemo = () => {
  const { addModal } = useSurfaces()

  const openBasicModal = () => {
    addModal(
      <Modal heading="Basic Modal">
        <div className="p-6">
          <Text>This is a basic modal with just content and close functionality.</Text>
        </div>
      </Modal>
    )
  }

  const openActionModal = () => {
    addModal(
      <Modal 
        heading="Action Modal"
        primaryActionTrigger={{
          children: 'Save Changes',
          onClick: () => alert('Changes saved successfully!'),
        }}
      >
        <div className="p-6">
          <Text className="mb-4">
            This modal includes a primary action button for user decisions.
          </Text>
          <div className="space-y-3">
            <input
              type="text"
              className="w-full p-2 border rounded"
              placeholder="Enter your name..."
            />
            <textarea
              className="w-full p-2 border rounded"
              rows={3}
              placeholder="Add a description..."
            />
          </div>
        </div>
      </Modal>
    )
  }

  const openFormModal = () => {
    addModal(
      <Modal 
        heading="Create New Project"
        primaryActionTrigger={{
          children: 'Create Project',
          onClick: () => alert('Project created successfully!'),
        }}
      >
        <div className="p-6">
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium mb-1">Project Name</label>
              <input
                type="text"
                className="w-full p-2 border rounded"
                placeholder="My Awesome Project"
              />
            </div>
            <div>
              <label className="block text-sm font-medium mb-1">Description</label>
              <textarea
                className="w-full p-2 border rounded"
                rows={3}
                placeholder="Describe your project..."
              />
            </div>
            <div>
              <label className="block text-sm font-medium mb-1">Category</label>
              <select className="w-full p-2 border rounded">
                <option>Web Application</option>
                <option>Mobile App</option>
                <option>API Service</option>
              </select>
            </div>
          </div>
        </div>
      </Modal>
    )
  }

  const openConfirmModal = () => {
    addModal(
      <Modal 
        heading="Confirm Deletion"
        primaryActionTrigger={{
          children: 'Delete',
          onClick: () => alert('Item deleted successfully!'),
        }}
      >
        <div className="p-6">
          <Text className="mb-2">Are you sure you want to delete this item?</Text>
          <Text variant="subtext" theme="neutral" className="text-red-600">
            This action cannot be undone.
          </Text>
        </div>
      </Modal>
    )
  }

  return (
    <div className="flex flex-wrap gap-3">
      <Button onClick={openBasicModal}>Basic Modal</Button>
      <Button onClick={openActionModal}>Action Modal</Button>
      <Button onClick={openFormModal}>Form Modal</Button>
      <Button onClick={openConfirmModal}>Confirmation Modal</Button>
    </div>
  )
}

// User management demo component
const UserManagementDemo = () => {
  const { addModal } = useSurfaces()

  const openEditProfile = () => {
    addModal(
      <Modal 
        heading="Edit Profile"
        primaryActionTrigger={{
          children: 'Update Profile',
          onClick: () => alert('Profile updated successfully!'),
        }}
      >
        <div className="p-6">
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium mb-1">Display Name</label>
              <input
                type="text"
                className="w-full p-2 border rounded"
                defaultValue="John Doe"
              />
            </div>
            <div>
              <label className="block text-sm font-medium mb-1">Email</label>
              <input
                type="email"
                className="w-full p-2 border rounded"
                defaultValue="john@example.com"
              />
            </div>
            <div>
              <label className="block text-sm font-medium mb-1">Bio</label>
              <textarea
                className="w-full p-2 border rounded"
                rows={3}
                defaultValue="Software developer passionate about creating great user experiences."
              />
            </div>
          </div>
        </div>
      </Modal>
    )
  }

  const openChangePassword = () => {
    addModal(
      <Modal 
        heading="Change Password"
        primaryActionTrigger={{
          children: 'Update Password',
          onClick: () => alert('Password changed successfully!'),
        }}
      >
        <div className="p-6">
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium mb-1">Current Password</label>
              <input
                type="password"
                className="w-full p-2 border rounded"
                placeholder="Enter current password"
              />
            </div>
            <div>
              <label className="block text-sm font-medium mb-1">New Password</label>
              <input
                type="password"
                className="w-full p-2 border rounded"
                placeholder="Enter new password"
              />
            </div>
            <div>
              <label className="block text-sm font-medium mb-1">Confirm Password</label>
              <input
                type="password"
                className="w-full p-2 border rounded"
                placeholder="Confirm new password"
              />
            </div>
          </div>
        </div>
      </Modal>
    )
  }

  return (
    <>
      <Button onClick={openEditProfile}>Edit Profile</Button>
      <Button onClick={openChangePassword}>Change Password</Button>
    </>
  )
}

// Data operations demo component
const DataOperationsDemo = () => {
  const { addModal } = useSurfaces()

  const openCreateContent = () => {
    addModal(
      <Modal 
        heading="Create New Article"
        primaryActionTrigger={{
          children: 'Publish Article',
          onClick: () => alert('Article published successfully!'),
        }}
      >
        <div className="p-6">
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium mb-1">Title</label>
              <input
                type="text"
                className="w-full p-2 border rounded"
                placeholder="Enter article title..."
              />
            </div>
            <div>
              <label className="block text-sm font-medium mb-1">Category</label>
              <select className="w-full p-2 border rounded">
                <option>Technology</option>
                <option>Design</option>
                <option>Business</option>
                <option>Tutorial</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium mb-1">Content</label>
              <textarea
                className="w-full p-2 border rounded"
                rows={6}
                placeholder="Write your article content here..."
              />
            </div>
          </div>
        </div>
      </Modal>
    )
  }

  const openDeleteConfirm = () => {
    addModal(
      <Modal 
        heading="Delete Article"
        primaryActionTrigger={{
          children: 'Delete Forever',
          onClick: () => alert('Article deleted successfully!'),
        }}
      >
        <div className="p-6">
          <Text className="mb-4">
            Are you sure you want to delete &ldquo;How to Build Great UIs&rdquo;?
          </Text>
          <div className="p-3 bg-red-50 border border-red-200 rounded">
            <Text variant="subtext" className="text-red-700">
              <strong>Warning:</strong> This will permanently delete the article and all its associated data. This action cannot be undone.
            </Text>
          </div>
        </div>
      </Modal>
    )
  }

  const openBulkAction = () => {
    addModal(
      <Modal 
        heading="Bulk Update Articles"
        primaryActionTrigger={{
          children: 'Update 5 Articles',
          onClick: () => alert('Bulk update completed!'),
        }}
      >
        <div className="p-6">
          <Text className="mb-4">
            You have selected 5 articles to update. Choose what you&rsquo;d like to change:
          </Text>
          <div className="space-y-3">
            <label className="flex items-center space-x-2">
              <input type="checkbox" className="rounded" />
              <span>Update category to "Featured"</span>
            </label>
            <label className="flex items-center space-x-2">
              <input type="checkbox" className="rounded" />
              <span>Mark as published</span>
            </label>
            <label className="flex items-center space-x-2">
              <input type="checkbox" className="rounded" />
              <span>Add "trending" tag</span>
            </label>
          </div>
        </div>
      </Modal>
    )
  }

  return (
    <>
      <Button onClick={openCreateContent}>Create Article</Button>
      <Button onClick={openDeleteConfirm}>Delete Article</Button>
      <Button onClick={openBulkAction}>Bulk Actions</Button>
    </>
  )
}

// Footer actions demo component
const FooterActionsDemo = () => {
  const { addModal } = useSurfaces()

  const openModalWithInfo = () => {
    addModal(
      <Modal 
        heading="Deploy Component"
        primaryActionTrigger={{
          children: 'Deploy Now',
          onClick: () => alert('Component deployed successfully!'),
        }}
        footerActions={
          <div className="flex items-center gap-2">
            <Icon variant="Info" size={16} />
            <Text variant="subtext" theme="neutral">
              This will take approximately 5 minutes
            </Text>
          </div>
        }
      >
        <div className="p-6">
          <Text className="mb-4">
            Select the build you want to deploy to your environment.
          </Text>
          <div className="space-y-3">
            <div className="p-3 border rounded">
              <Text variant="base" weight="strong">Build #123</Text>
              <Text variant="subtext" theme="neutral">Latest build from main branch</Text>
            </div>
            <div className="p-3 border rounded opacity-50">
              <Text variant="base" weight="strong">Build #122</Text>
              <Text variant="subtext" theme="neutral">Previous stable build</Text>
            </div>
          </div>
        </div>
      </Modal>
    )
  }

  const openModalWithActions = () => {
    addModal(
      <Modal 
        heading="Save Draft"
        primaryActionTrigger={{
          children: 'Save & Continue',
          onClick: () => alert('Draft saved successfully!'),
        }}
        footerActions={
          <div className="flex items-center gap-3">
            <Button 
              variant="ghost" 
              onClick={() => alert('Auto-save enabled!')}
            >
              <Icon variant="Lightning" size={16} />
              Enable Auto-save
            </Button>
            <div className="text-sm text-gray-500">|</div>
            <Text variant="subtext" theme="neutral">
              Last saved: 2 minutes ago
            </Text>
          </div>
        }
      >
        <div className="p-6">
          <Text className="mb-4">
            Your document has unsaved changes. Save as draft to continue later?
          </Text>
          <div className="p-3 bg-yellow-50 border border-yellow-200 rounded">
            <Text variant="subtext" className="text-yellow-800">
              You have 3 unsaved changes that will be preserved in the draft.
            </Text>
          </div>
        </div>
      </Modal>
    )
  }

  const openModalWithStatus = () => {
    addModal(
      <Modal 
        heading="Process File"
        primaryActionTrigger={{
          children: 'Start Processing',
          onClick: () => alert('File processing started!'),
        }}
        footerActions={
          <div className="flex items-center gap-2 text-green-600">
            <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse"></div>
            <Text variant="subtext">
              Ready to process • 2.4MB file size
            </Text>
          </div>
        }
      >
        <div className="p-6">
          <Text className="mb-4">
            Ready to process the uploaded file. This will analyze the content and generate a report.
          </Text>
          <div className="space-y-2 text-sm text-gray-600">
            <div className="flex justify-between">
              <span>File name:</span>
              <span className="font-mono">document.pdf</span>
            </div>
            <div className="flex justify-between">
              <span>File size:</span>
              <span>2.4 MB</span>
            </div>
            <div className="flex justify-between">
              <span>Estimated time:</span>
              <span>~30 seconds</span>
            </div>
          </div>
        </div>
      </Modal>
    )
  }

  return (
    <div className="flex flex-wrap gap-3">
      <Button onClick={openModalWithInfo}>Deploy Modal</Button>
      <Button onClick={openModalWithActions}>Save Draft Modal</Button>
      <Button onClick={openModalWithStatus}>Process File Modal</Button>
    </div>
  )
}




export const BasicUsage = () => (
  <SurfacesProvider>
    <div className="space-y-6">
      <div className="space-y-3">
        <h3 className="text-lg font-semibold">Basic Modal Usage</h3>
        <p className="text-sm text-gray-600 dark:text-gray-400">
          The Modal component creates centered overlay dialogs that capture user focus
          and require interaction before returning to the main interface. Modals are
          ideal for confirmations, forms, detailed information, and critical user
          decisions that need full attention.
        </p>
      </div>

      <div className="space-y-4">
        <h4 className="text-sm font-medium">Simple Modal Example</h4>
        <div className="p-4 border rounded-lg">
          <SimpleModalDemo />
        </div>
        <Text variant="subtext" theme="neutral">
          Click the button above to see a basic modal with close functionality
        </Text>
      </div>

      <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
        <strong>Key Features:</strong>
        <ul className="mt-2 space-y-1 list-disc list-inside">
          <li>Centered overlay with backdrop blur for focus isolation</li>
          <li>Automatic focus management and keyboard navigation support</li>
          <li>Built-in close button and click-outside-to-dismiss behavior</li>
          <li>Portal rendering for proper z-index and accessibility</li>
          <li>Smooth enter/exit animations with fade and scale effects</li>
          <li>Responsive design that adapts to different screen sizes</li>
        </ul>
      </div>
    </div>
  </SurfacesProvider>
)

export const ModalVariants = () => (
  <SurfacesProvider>
    <div className="space-y-6">
      <div className="space-y-3">
        <h3 className="text-lg font-semibold">Modal Variants</h3>
        <p className="text-sm text-gray-600 dark:text-gray-400">
          Modals support different configurations to handle various user interaction
          patterns. The{' '}
          <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
            primaryActionTrigger
          </code>{' '}
          prop adds action buttons to the modal footer for user decisions and form submissions.
        </p>
      </div>

      <div className="space-y-4">
        <h4 className="text-sm font-medium">Available Configurations</h4>
        <div className="p-4 border rounded-lg">
          <ModalVariantsDemo />
        </div>
        <Text variant="subtext" theme="neutral">
          Try different modal types to see how they handle various use cases
        </Text>
      </div>

      <div className="space-y-4">
        <h4 className="text-sm font-medium">Configuration Options</h4>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
          <div className="p-3 border rounded-lg">
            <strong className="text-blue-600 dark:text-blue-400">Basic Modal:</strong>
            <span className="block text-gray-600 dark:text-gray-400 mt-1">
              Simple content display with close functionality only
            </span>
          </div>
          <div className="p-3 border rounded-lg">
            <strong className="text-green-600 dark:text-green-400">Action Modal:</strong>
            <span className="block text-gray-600 dark:text-gray-400 mt-1">
              Includes primary action button for confirmations and submissions
            </span>
          </div>
          <div className="p-3 border rounded-lg">
            <strong className="text-purple-600 dark:text-purple-400">Form Modal:</strong>
            <span className="block text-gray-600 dark:text-gray-400 mt-1">
              Complex forms with multiple fields and validation
            </span>
          </div>
          <div className="p-3 border rounded-lg">
            <strong className="text-orange-600 dark:text-orange-400">Confirmation Modal:</strong>
            <span className="block text-gray-600 dark:text-gray-400 mt-1">
              Critical decisions requiring explicit user confirmation
            </span>
          </div>
        </div>
      </div>
    </div>
  </SurfacesProvider>
)

export const ModalUsageExamples = () => (
  <SurfacesProvider>
    <div className="space-y-6">
      <div className="space-y-3">
        <h3 className="text-lg font-semibold">Modal Usage Examples</h3>
        <p className="text-sm text-gray-600 dark:text-gray-400">
          Modals are essential for user workflows that require focused attention,
          data input, or critical decisions. They should interrupt the user's flow
          only when necessary and provide clear paths to completion or cancellation.
        </p>
      </div>

      <div className="space-y-4">
        <h4 className="text-sm font-medium">User Account Management</h4>
        <div className="p-4 border rounded-lg space-y-3">
          <Text variant="base" className="font-medium">Profile Settings</Text>
          <div className="flex gap-3">
            <UserManagementDemo />
          </div>
        </div>
      </div>

      <div className="space-y-4">
        <h4 className="text-sm font-medium">Data Operations</h4>
        <div className="p-4 border rounded-lg space-y-3">
          <Text variant="base" className="font-medium">Content Management</Text>
          <div className="flex gap-3">
            <DataOperationsDemo />
          </div>
        </div>
      </div>

      <div className="space-y-4">
        <h4 className="text-sm font-medium">Implementation Example</h4>
        <div className="p-4 border rounded-lg">
          <Card>
            <pre className="bg-gray-50 dark:bg-gray-900 p-4 rounded text-sm overflow-x-auto">
              {`import { Modal } from '@/components/surfaces/Modal'
import { useSurfaces } from '@/hooks/use-surfaces'

function DeleteButton({ itemName }) {
  const { addModal } = useSurfaces()
  
  const handleDelete = () => {
    addModal(
      <Modal 
        heading="Confirm Deletion"
        primaryActionTrigger={{
          children: "Delete",
          onClick: () => performDelete()
        }}
      >
        <div className="p-6">
          <p>Are you sure you want to delete "{itemName}"?</p>
          <p className="text-red-600 text-sm mt-2">
            This action cannot be undone.
          </p>
        </div>
      </Modal>
    )
  }
  
  return <Button onClick={handleDelete}>Delete Item</Button>
}`}
            </pre>
          </Card>
        </div>
      </div>

      <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
        <strong>Best Practices:</strong>
        <ul className="mt-2 space-y-1 list-disc list-inside">
          <li>Use modals sparingly - only for critical interactions that need full attention</li>
          <li>Provide clear, descriptive headings that explain the modal's purpose</li>
          <li>Include primary actions for confirmations, submissions, or next steps</li>
          <li>Always allow users to easily dismiss or cancel modal interactions</li>
          <li>Keep modal content focused and avoid overwhelming users with too much information</li>
          <li>Ensure proper keyboard navigation and screen reader accessibility</li>
        </ul>
      </div>
    </div>
  </SurfacesProvider>
)

export const FooterActions = () => (
  <SurfacesProvider>
    <div className="space-y-6">
      <div className="space-y-3">
        <h3 className="text-lg font-semibold">Footer Actions</h3>
        <p className="text-sm text-gray-600 dark:text-gray-400">
          The <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">footerActions</code> prop
          allows you to add custom content to the left side of the modal footer, before the Cancel/Close and
          primary action buttons. This is useful for displaying contextual information, additional actions,
          or status indicators.
        </p>
      </div>

      <div className="space-y-4">
        <h4 className="text-sm font-medium">Footer Actions Examples</h4>
        <div className="p-4 border rounded-lg">
          <FooterActionsDemo />
        </div>
        <Text variant="subtext" theme="neutral">
          Try these examples to see different types of footer actions in practice
        </Text>
      </div>

      <div className="space-y-4">
        <h4 className="text-sm font-medium">Use Cases</h4>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 text-sm">
          <div className="p-3 border rounded-lg">
            <strong className="text-blue-600 dark:text-blue-400">Contextual Info:</strong>
            <span className="block text-gray-600 dark:text-gray-400 mt-1">
              Show processing time, file size, or other relevant details
            </span>
          </div>
          <div className="p-3 border rounded-lg">
            <strong className="text-green-600 dark:text-green-400">Additional Actions:</strong>
            <span className="block text-gray-600 dark:text-gray-400 mt-1">
              Secondary buttons like "Enable auto-save" or quick toggles
            </span>
          </div>
          <div className="p-3 border rounded-lg">
            <strong className="text-purple-600 dark:text-purple-400">Status Indicators:</strong>
            <span className="block text-gray-600 dark:text-gray-400 mt-1">
              Connection status, validation state, or process readiness
            </span>
          </div>
        </div>
      </div>

      <div className="space-y-4">
        <h4 className="text-sm font-medium">Implementation Example</h4>
        <div className="p-4 border rounded-lg">
          <Card>
            <pre className="bg-gray-50 dark:bg-gray-900 p-4 rounded text-sm overflow-x-auto">
              {`<Modal 
  heading="Deploy Component"
  primaryActionTrigger={{
    children: 'Deploy Now',
    onClick: handleDeploy
  }}
  footerActions={
    <div className="flex items-center gap-2">
      <Icon variant="Info" size={16} />
      <Text variant="subtext" theme="neutral">
        This will take approximately 5 minutes
      </Text>
    </div>
  }
>
  <div className="p-6">
    <p>Select the build you want to deploy...</p>
  </div>
</Modal>`}
            </pre>
          </Card>
        </div>
      </div>
    </div>
  </SurfacesProvider>
)
