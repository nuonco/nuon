import { Modal } from './Modal'
import { SurfacesProvider } from '@/providers/surfaces-provider'
import { useSurfaces } from '@/hooks/use-surfaces'
import { Button } from '@/components/common/Button'
import { Text } from '@/components/common/Text'
import { Card } from '@/components/common/Card'

// Component that uses the surfaces context for interactive demos
const InteractiveModalDemo = () => {
  const { addModal } = useSurfaces()

  const openDefaultModal = () => {
    addModal(
      <Modal heading="Default Modal">
        <div className="p-6">
          <Text>This is the content of the modal.</Text>
          <div className="mt-4">
            <input
              type="text"
              className="w-full p-2 border rounded"
              placeholder="Enter some text..."
            />
          </div>
        </div>
      </Modal>
    )
  }

  const openModalWithPrimaryAction = () => {
    addModal(
      <Modal
        heading="Modal with Primary Action"
        primaryActionTrigger={{
          children: 'Save Changes',
          onClick: () => alert('Changes saved!'),
        }}
      >
        <div className="p-6">
          <Text className="mb-4">
            This modal includes a primary action button that appears in the
            footer.
          </Text>
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium mb-1">Name</label>
              <input
                type="text"
                className="w-full p-2 border rounded"
                placeholder="Enter your name..."
              />
            </div>
            <div>
              <label className="block text-sm font-medium mb-1">Email</label>
              <input
                type="email"
                className="w-full p-2 border rounded"
                placeholder="Enter your email..."
              />
            </div>
          </div>
        </div>
      </Modal>
    )
  }

  const openComplexModal = () => {
    addModal(
      <Modal
        heading="Complex Modal Example"
        primaryActionTrigger={{
          children: 'Create Project',
          onClick: () => alert('Project created!'),
        }}
      >
        <div className="p-6">
          <Text className="mb-6 text-gray-600">
            This is an example of a more complex modal with various content
            types.
          </Text>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <Card>
              <Text variant="base" className="mb-4">
                Project Details
              </Text>
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium mb-1">
                    Project Name
                  </label>
                  <input
                    type="text"
                    className="w-full p-2 border rounded"
                    placeholder="My Awesome Project"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium mb-1">
                    Description
                  </label>
                  <textarea
                    className="w-full p-2 border rounded"
                    rows={3}
                    placeholder="Describe your project..."
                  />
                </div>
              </div>
            </Card>

            <Card>
              <Text variant="base" className="mb-4">
                Configuration
              </Text>
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium mb-1">
                    Template
                  </label>
                  <select className="w-full p-2 border rounded">
                    <option>React App</option>
                    <option>Next.js App</option>
                    <option>Node.js API</option>
                  </select>
                </div>
                <div className="space-y-2">
                  <Text variant="label" weight="strong">
                    Features
                  </Text>
                  <div className="space-y-2">
                    <div className="flex items-center gap-2">
                      <input type="checkbox" id="typescript" defaultChecked />
                      <label htmlFor="typescript" className="text-sm">
                        TypeScript
                      </label>
                    </div>
                    <div className="flex items-center gap-2">
                      <input type="checkbox" id="testing" />
                      <label htmlFor="testing" className="text-sm">
                        Testing Setup
                      </label>
                    </div>
                    <div className="flex items-center gap-2">
                      <input type="checkbox" id="linting" defaultChecked />
                      <label htmlFor="linting" className="text-sm">
                        ESLint & Prettier
                      </label>
                    </div>
                  </div>
                </div>
              </div>
            </Card>
          </div>
        </div>
      </Modal>
    )
  }

  return (
    <div className="space-y-4">
      <Text variant="h3" className="mb-4">
        Interactive Modal Examples
      </Text>
      <Text className="mb-6 text-gray-600">
        Click the buttons below to open modals with different configurations.
        Modals appear as overlays with backdrop blur and automatic focus
        management.
      </Text>

      <div className="flex flex-wrap gap-4">
        <Button onClick={openDefaultModal}>Open Default Modal</Button>

        <Button onClick={openModalWithPrimaryAction}>
          Modal with Primary Action
        </Button>

        <Button onClick={openComplexModal}>Complex Modal</Button>
      </div>

      <div className="mt-8 p-4 bg-yellow-50 rounded-lg border border-yellow-200">
        <Text variant="label" weight="strong" className="text-yellow-800">
          Note:
        </Text>
        <Text className="text-yellow-700 mt-1">
          Modals automatically include close functionality via the X button and
          clicking outside the modal. They use transitions and are rendered in a
          portal for proper layering.
        </Text>
      </div>
    </div>
  )
}

export const ModalInteractive = () => (
  <SurfacesProvider>
    <InteractiveModalDemo />
    <Modal
      triggerButton={{ children: 'Open single modal' }}
      heading="Single Modal Component"
    >
      <div className="p-6">
        <Text>Single modal component usage</Text>
      </div>
    </Modal>
  </SurfacesProvider>
)

// Component to demonstrate modal features without opening them
export const ModalFeatures = () => (
  <div className="space-y-6">
    <Text variant="h3">Modal Component Features</Text>

    <div className="space-y-4">
      <Card>
        <Text variant="base" className="mb-4">
          Key Features
        </Text>
        <div className="space-y-2">
          <div className="flex items-center gap-2">
            <div className="w-2 h-2 bg-green-500 rounded-full"></div>
            <Text>Centered overlay with backdrop blur</Text>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-2 h-2 bg-green-500 rounded-full"></div>
            <Text>Automatic focus management and keyboard navigation</Text>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-2 h-2 bg-green-500 rounded-full"></div>
            <Text>Built-in close button and click-outside-to-close</Text>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-2 h-2 bg-green-500 rounded-full"></div>
            <Text>Optional primary action button</Text>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-2 h-2 bg-green-500 rounded-full"></div>
            <Text>Portal rendering for proper z-index layering</Text>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-2 h-2 bg-green-500 rounded-full"></div>
            <Text>URL integration with modal keys</Text>
          </div>
        </div>
      </Card>

      <Card>
        <Text variant="base" className="mb-4">
          Usage Options
        </Text>
        <div className="space-y-3">
          <div>
            <Text variant="label" weight="strong">
              Via useSurfaces hook
            </Text>
            <Text variant="subtext" className="ml-2">
              Programmatically open modals from any component
            </Text>
          </div>
          <div>
            <Text variant="label" weight="strong">
              With triggerButton prop
            </Text>
            <Text variant="subtext" className="ml-2">
              Declarative modal that opens when button is clicked
            </Text>
          </div>
        </div>
      </Card>
    </div>
  </div>
)

export const ModalUsageExample = () => (
  <div className="space-y-6">
    <Text variant="h3">Usage Example</Text>
    <Text className="text-gray-600">How to use modals in your components:</Text>

    <Card>
      <pre className="bg-gray-50 p-4 rounded text-sm overflow-x-auto">
        {`import { Modal } from '@/components';
import { useSurfaces } from '@/hooks';

function MyComponent() {
  const { addModal } = useSurfaces();
  
  const openConfirmDialog = () => {
    addModal(
      <Modal 
        heading="Confirm Action"
        primaryActionTrigger={{
          children: "Confirm",
          onClick: () => handleConfirm()
        }}
      >
        <div className="p-6">
          <p>Are you sure you want to proceed?</p>
        </div>
      </Modal>
    );
  };
  
  return (
    <button onClick={openConfirmDialog}>
      Delete Item
    </button>
  );
}`}
      </pre>
    </Card>

    <div className="p-4 bg-blue-50 rounded-lg border border-blue-200">
      <Text variant="label" weight="strong" className="text-blue-800">
        Remember:
      </Text>
      <Text className="text-blue-700 mt-1">
        Your app needs to be wrapped with SurfacesProvider and have a div with
        id=modal-root for modals to render properly.
      </Text>
    </div>
  </div>
)
