/* eslint-disable react/no-unescaped-entities */
import { Panel } from './Panel'
import { SurfacesProvider } from '@/providers/surfaces-provider'
import { useSurfaces } from '@/hooks/use-surfaces'
import { Button } from '@/components/common/Button'
import { Text } from '@/components/common/Text'
import { Card } from '@/components/common/Card'

// Simple panel demo component
const SimplePanelDemo = () => {
  const { addPanel } = useSurfaces()

  const openPanel = () => {
    addPanel(
      <Panel heading="Welcome to Panels">
        <div className="p-6">
          <Text className="mb-4">
            This is a simple panel example that demonstrates basic slide-out functionality.
            Panels appear from the right side and overlay the main content.
          </Text>
          <Text variant="subtext" theme="neutral">
            You can close this panel by clicking the X button, pressing Escape, or clicking the overlay.
          </Text>
        </div>
      </Panel>
    )
  }

  return <Button onClick={openPanel}>Open Simple Panel</Button>
}

// Panel sizes demo component
const PanelSizesDemo = () => {
  const { addPanel } = useSurfaces()

  const openDefaultPanel = () => {
    addPanel(
      <Panel heading="Default Panel">
        <div className="p-6">
          <Text className="mb-4">
            This is a default-sized panel with fixed width. Perfect for forms, 
            user profiles, and detailed information that doesn't require a lot of space.
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
              placeholder="Add notes..."
            />
          </div>
        </div>
      </Panel>
    )
  }

  const openHalfPanel = () => {
    addPanel(
      <Panel size="half" heading="Half-Width Panel">
        <div className="p-6">
          <Text className="mb-4">
            This panel takes up half the screen width, making it ideal for 
            side-by-side workflows or detailed content views.
          </Text>
          
          <div className="space-y-4">
            <Card>
              <Text variant="base" className="mb-4">Project Statistics</Text>
              <div className="grid grid-cols-2 gap-4">
                <div className="text-center p-3 bg-blue-50 rounded">
                  <Text variant="base" className="text-blue-600 font-semibold">42</Text>
                  <Text variant="subtext">Active Tasks</Text>
                </div>
                <div className="text-center p-3 bg-green-50 rounded">
                  <Text variant="base" className="text-green-600 font-semibold">128</Text>
                  <Text variant="subtext">Completed</Text>
                </div>
              </div>
            </Card>
          </div>
        </div>
      </Panel>
    )
  }

  const openFullPanel = () => {
    addPanel(
      <Panel size="full" heading="Full-Width Panel">
        <div className="p-6">
          <Text className="mb-6">
            This panel takes up the entire screen width, providing maximum space 
            for complex dashboards, data tables, or comprehensive forms.
          </Text>
          
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            <Card>
              <Text variant="base" className="mb-3">Performance</Text>
              <div className="space-y-2">
                <div className="flex justify-between">
                  <span>CPU Usage</span>
                  <span className="text-blue-600">45%</span>
                </div>
                <div className="flex justify-between">
                  <span>Memory</span>
                  <span className="text-green-600">62%</span>
                </div>
                <div className="flex justify-between">
                  <span>Storage</span>
                  <span className="text-orange-600">78%</span>
                </div>
              </div>
            </Card>
            
            <Card>
              <Text variant="base" className="mb-3">Recent Activity</Text>
              <div className="space-y-2">
                <div className="p-2 bg-gray-50 rounded text-sm">User login detected</div>
                <div className="p-2 bg-gray-50 rounded text-sm">Task completed</div>
                <div className="p-2 bg-gray-50 rounded text-sm">File uploaded</div>
              </div>
            </Card>
            
            <Card>
              <Text variant="base" className="mb-3">Settings</Text>
              <div className="space-y-3">
                <label className="flex items-center space-x-2">
                  <input type="checkbox" />
                  <span>Enable notifications</span>
                </label>
                <label className="flex items-center space-x-2">
                  <input type="checkbox" />
                  <span>Auto-save changes</span>
                </label>
              </div>
            </Card>
          </div>
        </div>
      </Panel>
    )
  }

  return (
    <div className="flex flex-wrap gap-3">
      <Button onClick={openDefaultPanel}>Default Size</Button>
      <Button onClick={openHalfPanel}>Half Width</Button>
      <Button onClick={openFullPanel}>Full Width</Button>
    </div>
  )
}

// Panel usage examples demo component
const PanelUsageDemo = () => {
  const { addPanel } = useSurfaces()

  const openUserProfile = () => {
    addPanel(
      <Panel heading="User Profile" size="default">
        <div className="p-6">
          <div className="space-y-4">
            <div className="flex items-center space-x-4">
              <div className="w-16 h-16 bg-blue-500 rounded-full flex items-center justify-center text-white font-semibold text-xl">
                JD
              </div>
              <div>
                <Text variant="base" className="font-semibold">John Doe</Text>
                <Text variant="subtext" theme="neutral">john@example.com</Text>
              </div>
            </div>
            
            <div className="space-y-3">
              <div>
                <label className="block text-sm font-medium mb-1">Display Name</label>
                <input
                  type="text"
                  className="w-full p-2 border rounded"
                  defaultValue="John Doe"
                />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">Bio</label>
                <textarea
                  className="w-full p-2 border rounded"
                  rows={3}
                  defaultValue="Product designer passionate about creating intuitive user experiences."
                />
              </div>
            </div>
            
            <div className="pt-4 border-t">
              <Button>Update Profile</Button>
            </div>
          </div>
        </div>
      </Panel>
    )
  }

  const openProjectDetails = () => {
    addPanel(
      <Panel heading="Project Details" size="half">
        <div className="p-6">
          <div className="space-y-6">
            <div>
              <Text variant="base" className="font-semibold mb-2">Project Overview</Text>
              <Text variant="subtext" theme="neutral">
                A comprehensive dashboard redesign focused on improving user experience 
                and streamlining key workflows.
              </Text>
            </div>
            
            <div>
              <Text variant="base" className="font-semibold mb-3">Team Members</Text>
              <div className="space-y-2">
                <div className="flex items-center space-x-3">
                  <div className="w-8 h-8 bg-purple-500 rounded-full flex items-center justify-center text-white text-sm">A</div>
                  <span>Alice Johnson - Project Lead</span>
                </div>
                <div className="flex items-center space-x-3">
                  <div className="w-8 h-8 bg-green-500 rounded-full flex items-center justify-center text-white text-sm">B</div>
                  <span>Bob Smith - Frontend Developer</span>
                </div>
                <div className="flex items-center space-x-3">
                  <div className="w-8 h-8 bg-orange-500 rounded-full flex items-center justify-center text-white text-sm">C</div>
                  <span>Carol Brown - Designer</span>
                </div>
              </div>
            </div>
            
            <div>
              <Text variant="base" className="font-semibold mb-3">Timeline</Text>
              <div className="space-y-2 text-sm">
                <div className="flex justify-between">
                  <span>Started</span>
                  <span>March 15, 2024</span>
                </div>
                <div className="flex justify-between">
                  <span>Due Date</span>
                  <span>June 30, 2024</span>
                </div>
                <div className="flex justify-between">
                  <span>Progress</span>
                  <span className="text-green-600">65% Complete</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </Panel>
    )
  }

  const openSettingsPanel = () => {
    addPanel(
      <Panel heading="Application Settings" size="full">
        <div className="p-6">
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            <Card>
              <Text variant="base" className="font-semibold mb-4">General</Text>
              <div className="space-y-3">
                <div>
                  <label className="block text-sm font-medium mb-1">Language</label>
                  <select className="w-full p-2 border rounded">
                    <option>English</option>
                    <option>Spanish</option>
                    <option>French</option>
                  </select>
                </div>
                <div>
                  <label className="block text-sm font-medium mb-1">Timezone</label>
                  <select className="w-full p-2 border rounded">
                    <option>UTC-5 (Eastern)</option>
                    <option>UTC-8 (Pacific)</option>
                    <option>UTC+0 (GMT)</option>
                  </select>
                </div>
              </div>
            </Card>
            
            <Card>
              <Text variant="base" className="font-semibold mb-4">Notifications</Text>
              <div className="space-y-3">
                <label className="flex items-center space-x-2">
                  <input type="checkbox" defaultChecked />
                  <span>Email notifications</span>
                </label>
                <label className="flex items-center space-x-2">
                  <input type="checkbox" defaultChecked />
                  <span>Push notifications</span>
                </label>
                <label className="flex items-center space-x-2">
                  <input type="checkbox" />
                  <span>SMS notifications</span>
                </label>
              </div>
            </Card>
            
            <Card>
              <Text variant="base" className="font-semibold mb-4">Privacy</Text>
              <div className="space-y-3">
                <label className="flex items-center space-x-2">
                  <input type="checkbox" defaultChecked />
                  <span>Analytics tracking</span>
                </label>
                <label className="flex items-center space-x-2">
                  <input type="checkbox" />
                  <span>Share usage data</span>
                </label>
                <label className="flex items-center space-x-2">
                  <input type="checkbox" defaultChecked />
                  <span>Cookie consent</span>
                </label>
              </div>
            </Card>
          </div>
          
          <div className="mt-6 pt-6 border-t">
            <div className="flex gap-3">
              <Button>Save Settings</Button>
              <Button>Reset to Defaults</Button>
            </div>
          </div>
        </div>
      </Panel>
    )
  }

  return (
    <div className="flex flex-wrap gap-3">
      <Button onClick={openUserProfile}>View Profile</Button>
      <Button onClick={openProjectDetails}>Project Details</Button>
      <Button onClick={openSettingsPanel}>App Settings</Button>
    </div>
  )
}

export const BasicUsage = () => (
  <SurfacesProvider>
    <div className="space-y-6">
      <div className="space-y-3">
        <h3 className="text-lg font-semibold">Basic Panel Usage</h3>
        <p className="text-sm text-gray-600 dark:text-gray-400">
          The Panel component creates slide-out overlays that appear from the right 
          side of the screen. Panels are ideal for secondary content, detailed views,
          forms, and contextual information that doesn't require a full page navigation.
          They maintain context while providing focused interaction space.
        </p>
      </div>

      <div className="space-y-4">
        <h4 className="text-sm font-medium">Simple Panel Example</h4>
        <div className="p-4 border rounded-lg">
          <SimplePanelDemo />
        </div>
        <Text variant="subtext" theme="neutral">
          Click the button above to see a basic panel with slide-out functionality
        </Text>
      </div>

      <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
        <strong>Key Features:</strong>
        <ul className="mt-2 space-y-1 list-disc list-inside">
          <li>Smooth slide-out animation from the right side of the screen</li>
          <li>Automatic overlay with click-to-close and escape key support</li>
          <li>Built-in close button with proper accessibility attributes</li>
          <li>Portal rendering for correct z-index layering and focus management</li>
          <li>Three size options: default (fixed), half-width, and full-width</li>
          <li>Responsive behavior that adapts to different screen sizes</li>
        </ul>
      </div>
    </div>
  </SurfacesProvider>
)

export const PanelSizeVariants = () => (
  <SurfacesProvider>
    <div className="space-y-6">
      <div className="space-y-3">
        <h3 className="text-lg font-semibold">Panel Size Variants</h3>
        <p className="text-sm text-gray-600 dark:text-gray-400">
          Panels support three different size variants to accommodate different content
          types and use cases. The{' '}
          <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
            size
          </code>{' '}
          prop controls the width of the panel, allowing you to choose the most
          appropriate layout for your content.
        </p>
      </div>

      <div className="space-y-4">
        <h4 className="text-sm font-medium">Available Sizes</h4>
        <div className="p-4 border rounded-lg">
          <PanelSizesDemo />
        </div>
        <Text variant="subtext" theme="neutral">
          Try different panel sizes to see how they accommodate various content layouts
        </Text>
      </div>

      <div className="space-y-4">
        <h4 className="text-sm font-medium">Size Descriptions</h4>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 text-sm">
          <div className="p-3 border rounded-lg">
            <strong className="text-blue-600 dark:text-blue-400">Default:</strong>
            <span className="block text-gray-600 dark:text-gray-400 mt-1">
              Fixed width (w-104) - Perfect for forms, user profiles, and focused content
            </span>
          </div>
          <div className="p-3 border rounded-lg">
            <strong className="text-purple-600 dark:text-purple-400">Half:</strong>
            <span className="block text-gray-600 dark:text-gray-400 mt-1">
              50% screen width - Ideal for side-by-side workflows and detailed views
            </span>
          </div>
          <div className="p-3 border rounded-lg">
            <strong className="text-orange-600 dark:text-orange-400">Full:</strong>
            <span className="block text-gray-600 dark:text-gray-400 mt-1">
              100% screen width - Great for dashboards, settings, and complex layouts
            </span>
          </div>
        </div>
      </div>
    </div>
  </SurfacesProvider>
)

export const PanelUsageExamples = () => (
  <SurfacesProvider>
    <div className="space-y-6">
      <div className="space-y-3">
        <h3 className="text-lg font-semibold">Panel Usage Examples</h3>
        <p className="text-sm text-gray-600 dark:text-gray-400">
          Panels excel at providing contextual interfaces that don't disrupt the main
          workflow. They're commonly used for editing content, displaying details,
          configuring settings, and managing user accounts while keeping the primary
          interface visible and accessible.
        </p>
      </div>

      <div className="space-y-4">
        <h4 className="text-sm font-medium">Common Use Cases</h4>
        <div className="p-4 border rounded-lg space-y-3">
          <Text variant="base" className="font-medium">Interactive Examples</Text>
          <PanelUsageDemo />
        </div>
        <Text variant="subtext" theme="neutral">
          Click any button above to see real-world panel implementations
        </Text>
      </div>

      <div className="space-y-4">
        <h4 className="text-sm font-medium">Implementation Example</h4>
        <div className="p-4 border rounded-lg">
          <Card>
            <pre className="bg-gray-50 dark:bg-gray-900 p-4 rounded text-sm overflow-x-auto">
              {`import { Panel } from '@/components/surfaces/Panel'
import { useSurfaces } from '@/hooks/use-surfaces'

function UserList() {
  const { addPanel } = useSurfaces()
  
  const viewUserProfile = (user) => {
    addPanel(
      <Panel heading={\`\${user.name} Profile\`} size="half">
        <div className="p-6">
          <div className="space-y-4">
            <img src={user.avatar} alt={user.name} />
            <h3>{user.name}</h3>
            <p>{user.email}</p>
            {/* Additional profile content */}
          </div>
        </div>
      </Panel>
    )
  }
  
  return (
    <div>
      {users.map(user => (
        <button onClick={() => viewUserProfile(user)}>
          View {user.name}
        </button>
      ))}
    </div>
  )
}`}
            </pre>
          </Card>
        </div>
      </div>

      <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
        <strong>Best Practices:</strong>
        <ul className="mt-2 space-y-1 list-disc list-inside">
          <li>Use panels for secondary content that supports the main interface</li>
          <li>Choose appropriate sizes based on content complexity and user needs</li>
          <li>Provide clear headings that describe the panel's purpose and content</li>
          <li>Keep panel content focused and avoid overwhelming users with information</li>
          <li>Consider mobile responsiveness - panels adapt automatically to smaller screens</li>
          <li>Ensure proper keyboard navigation and screen reader accessibility</li>
        </ul>
      </div>
    </div>
  </SurfacesProvider>
)
