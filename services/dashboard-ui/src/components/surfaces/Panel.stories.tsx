import { Panel } from './Panel'
import { SurfacesProvider } from '@/providers/surfaces-provider'
import { useSurfaces } from '@/hooks/use-surfaces'
import { Button } from '@/components/common/Button'
import { Text } from '@/components/common/Text'
import { Card } from '@/components/common/Card'

// Component that uses the surfaces context for interactive demos
const InteractivePanelDemo = () => {
  const { addPanel } = useSurfaces()

  const openDefaultPanel = () => {
    addPanel(
      <Panel heading="Default Size Panel">
        <Text>
          This is a default-sized panel (w-104). Perfect for forms, details, or
          quick actions.
        </Text>
        <div className="space-y-4">
          <input
            type="text"
            className="w-full p-2 border rounded"
            placeholder="Enter some text..."
          />
          <textarea
            className="w-full p-2 border rounded"
            rows={3}
            placeholder="Add a description..."
          />
        </div>
      </Panel>
    )
  }

  const openHalfPanel = () => {
    addPanel(
      <Panel size="half">
        <div className="p-6">
          <Text variant="h3" className="mb-4">
            Half Width Panel
          </Text>
          <Text className="mb-6">
            This panel takes up half the screen width. Great for side-by-side
            workflows.
          </Text>

          <div className="space-y-4">
            <Card>
              <Text variant="base" className="mb-4">
                Quick Stats
              </Text>
              <div className="grid grid-cols-2 gap-4">
                <div className="text-center p-3 bg-blue-50 rounded">
                  <Text variant="base" className="text-blue-600">
                    24
                  </Text>
                  <Text variant="subtext">Active</Text>
                </div>
                <div className="text-center p-3 bg-green-50 rounded">
                  <Text variant="base" className="text-green-600">
                    156
                  </Text>
                  <Text variant="subtext">Complete</Text>
                </div>
              </div>
            </Card>

            <div className="space-y-2">
              <Text variant="label" weight="strong">
                Recent Activity
              </Text>
              <div className="space-y-1">
                <div className="p-2 bg-gray-50 rounded text-sm">
                  User logged in
                </div>
                <div className="p-2 bg-gray-50 rounded text-sm">
                  Task completed
                </div>
                <div className="p-2 bg-gray-50 rounded text-sm">
                  File uploaded
                </div>
              </div>
            </div>
          </div>
        </div>
      </Panel>
    )
  }

  const openFullPanel = () => {
    addPanel(
      <Panel size="full">
        <Text variant="h2" className="mb-6">
          Full Width Panel
        </Text>
        <Text className="mb-8 text-gray-600">
          This panel takes up the entire screen width. Perfect for detailed
          views, dashboards, or complex forms.
        </Text>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          <Card>
            <Text variant="base" className="mb-4">
              Overview
            </Text>
            <div className="space-y-3">
              <div className="flex justify-between">
                <Text variant="subtext">Total Users:</Text>
                <Text weight="strong">1,234</Text>
              </div>
              <div className="flex justify-between">
                <Text variant="subtext">Active Sessions:</Text>
                <Text weight="strong">89</Text>
              </div>
              <div className="flex justify-between">
                <Text variant="subtext">Revenue:</Text>
                <Text weight="strong" className="text-green-600">
                  $12,345
                </Text>
              </div>
            </div>
          </Card>

          <Card>
            <Text variant="base" className="mb-4">
              Performance
            </Text>
            <div className="space-y-4">
              <div>
                <div className="flex justify-between mb-1">
                  <Text variant="subtext">CPU Usage</Text>
                  <Text variant="subtext">72%</Text>
                </div>
                <div className="w-full bg-gray-200 rounded-full h-2">
                  <div
                    className="bg-blue-500 h-2 rounded-full"
                    style={{ width: '72%' }}
                  ></div>
                </div>
              </div>
              <div>
                <div className="flex justify-between mb-1">
                  <Text variant="subtext">Memory</Text>
                  <Text variant="subtext">45%</Text>
                </div>
                <div className="w-full bg-gray-200 rounded-full h-2">
                  <div
                    className="bg-green-500 h-2 rounded-full"
                    style={{ width: '45%' }}
                  ></div>
                </div>
              </div>
            </div>
          </Card>

          <Card>
            <Text variant="base" className="mb-4">
              Recent Actions
            </Text>
            <div className="space-y-2">
              <div className="flex items-center gap-2 p-2 bg-blue-50 rounded">
                <div className="w-2 h-2 bg-blue-500 rounded-full"></div>
                <Text variant="subtext">Deploy completed</Text>
              </div>
              <div className="flex items-center gap-2 p-2 bg-green-50 rounded">
                <div className="w-2 h-2 bg-green-500 rounded-full"></div>
                <Text variant="subtext">Backup successful</Text>
              </div>
              <div className="flex items-center gap-2 p-2 bg-yellow-50 rounded">
                <div className="w-2 h-2 bg-yellow-500 rounded-full"></div>
                <Text variant="subtext">Update pending</Text>
              </div>
            </div>
          </Card>
        </div>

        <div className="mt-8 pt-6 border-t">
          <Text variant="base" className="mb-4">
            Configuration
          </Text>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium mb-1">
                  API Endpoint
                </label>
                <input
                  type="url"
                  className="w-full p-2 border rounded"
                  placeholder="https://api.example.com"
                />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">
                  Timeout (seconds)
                </label>
                <input
                  type="number"
                  className="w-full p-2 border rounded"
                  placeholder="30"
                />
              </div>
            </div>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium mb-1">
                  Environment
                </label>
                <select className="w-full p-2 border rounded">
                  <option>Production</option>
                  <option>Staging</option>
                  <option>Development</option>
                </select>
              </div>
              <div className="flex items-center gap-2">
                <input type="checkbox" id="debug" />
                <label htmlFor="debug" className="text-sm">
                  Enable debug mode
                </label>
              </div>
            </div>
          </div>
        </div>
      </Panel>
    )
  }

  return (
    <div className="space-y-4">
      <Text variant="h3" className="mb-4">
        Interactive Panel Examples
      </Text>
      <Text className="mb-6 text-gray-600">
        Click the buttons below to open panels with different sizes and content.
        Panels appear as slide-out overlays from the right side of the screen.
      </Text>

      <div className="flex flex-wrap gap-4">
        <Button onClick={openDefaultPanel}>Open Default Panel</Button>

        <Button onClick={openHalfPanel}>Open Half Panel</Button>

        <Button onClick={openFullPanel}>Open Full Panel</Button>
      </div>

      <div className="mt-8 p-4 bg-yellow-50 rounded-lg border border-yellow-200">
        <Text variant="label" weight="strong" className="text-yellow-800">
          Note:
        </Text>
        <Text className="text-yellow-700 mt-1">
          Panels automatically include a close button and overlay click-to-close
          functionality. They use transitions and are rendered in a portal for
          proper layering.
        </Text>
      </div>
    </div>
  )
}

export const PanelSizes = () => (
  <SurfacesProvider>
    <InteractivePanelDemo />
    <Panel
      triggerButton={{ children: 'Open single panel' }}
      heading={
        <>
          <Text variant="h3">Heading</Text>
          <Text variant="subtext" theme="neutral">
            Sub heading
          </Text>
        </>
      }
    >
      <div className="text-sm">single panel component</div>
    </Panel>
  </SurfacesProvider>
)

// Component to demonstrate panel features without opening them
export const PanelFeatures = () => (
  <div className="space-y-6">
    <Text variant="h3">Panel Component Features</Text>

    <div className="space-y-4">
      <Card>
        <Text variant="base" className="mb-4">
          Size Options
        </Text>
        <div className="space-y-3">
          <div>
            <Text variant="label" weight="strong">
              default
            </Text>
            <Text variant="subtext" className="ml-2">
              Fixed width (w-104) - Good for forms and detail views
            </Text>
          </div>
          <div>
            <Text variant="label" weight="strong">
              half
            </Text>
            <Text variant="subtext" className="ml-2">
              Half screen width - Perfect for side-by-side workflows
            </Text>
          </div>
          <div>
            <Text variant="label" weight="strong">
              full
            </Text>
            <Text variant="subtext" className="ml-2">
              Full screen width - Ideal for dashboards and complex layouts
            </Text>
          </div>
        </div>
      </Card>

      <Card>
        <Text variant="base" className="mb-4">
          Key Features
        </Text>
        <div className="space-y-2">
          <div className="flex items-center gap-2">
            <div className="w-2 h-2 bg-green-500 rounded-full"></div>
            <Text>Slide-out animation from the right side</Text>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-2 h-2 bg-green-500 rounded-full"></div>
            <Text>Automatic overlay with click-to-close</Text>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-2 h-2 bg-green-500 rounded-full"></div>
            <Text>Built-in close button</Text>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-2 h-2 bg-green-500 rounded-full"></div>
            <Text>Portal rendering for proper z-index layering</Text>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-2 h-2 bg-green-500 rounded-full"></div>
            <Text>Dark mode support</Text>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-2 h-2 bg-green-500 rounded-full"></div>
            <Text>Custom styling support via className</Text>
          </div>
        </div>
      </Card>
    </div>
  </div>
)

export const PanelUsageExample = () => (
  <div className="space-y-6">
    <Text variant="h3">Usage Example</Text>
    <Text className="text-gray-600">How to use panels in your components:</Text>

    <Card>
      <pre className="bg-gray-50 p-4 rounded text-sm overflow-x-auto">
        {`import { Panel } from '@/components';
import { useSurfaces } from '@/hooks';

function MyComponent() {
  const { addPanel } = useSurfaces();
  
  const openUserProfile = () => {
    addPanel(
      <Panel size="half">
        <div className="p-6">
          <h2>User Profile</h2>
          {/* Your content here */}
        </div>
      </Panel>
    );
  };
  
  return (
    <button onClick={openUserProfile}>
      View Profile
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
        id=panel-root for panels to render properly.
      </Text>
    </div>
  </div>
)
