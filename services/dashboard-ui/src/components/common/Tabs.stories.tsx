/* eslint-disable react/no-unescaped-entities */
import { Tabs } from './Tabs'
import { Text } from './Text'
import { Button } from './Button'
import { Card } from './Card'
import { Badge } from './Badge'
import { Icon } from './Icon'

export const BasicUsage = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Basic Tab Usage</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Tabs provide a way to organize content into separate, switchable
        sections. Tab labels are automatically generated from the object keys
        using camelCase to sentence case conversion. The component includes
        smooth animations and hover effects.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Simple Tab Example</h4>
      <Tabs
        tabs={{
          overview: (
            <div className="p-4 space-y-3">
              <Text variant="h3" weight="stronger">
                Overview
              </Text>
              <Text>
                This is the overview tab content. Tab keys are automatically
                converted to readable labels.
              </Text>
            </div>
          ),
          settings: (
            <div className="p-4 space-y-3">
              <Text variant="h3" weight="stronger">
                Settings
              </Text>
              <Text>
                Configure your application settings here. Each tab can contain
                any React content.
              </Text>
            </div>
          ),
          help: (
            <div className="p-4 space-y-3">
              <Text variant="h3" weight="stronger">
                Help
              </Text>
              <Text>
                Find documentation and support resources in this section.
              </Text>
            </div>
          ),
        }}
      />
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Key Features:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Automatic tab label generation from object keys</li>
        <li>Smooth transitions between tab content</li>
        <li>Dynamic height adjustment based on content</li>
        <li>Hover and focus animations on tab buttons</li>
      </ul>
    </div>
  </div>
)

export const TabKeyConversion = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Tab Key Conversion</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Tab labels are automatically generated from object keys using camelCase
        to sentence case conversion. This allows you to use developer-friendly
        key names while displaying user-friendly labels.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">CamelCase Key Examples</h4>
      <Tabs
        tabs={{
          userProfile: (
            <div className="p-4">
              <Text variant="h3" weight="stronger">
                User Profile
              </Text>
              <Text>
                This tab key was{' '}
                <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
                  userProfile
                </code>{' '}
                and gets converted to "User Profile"
              </Text>
            </div>
          ),
          accountSettings: (
            <div className="p-4">
              <Text variant="h3" weight="stronger">
                Account Settings
              </Text>
              <Text>
                This tab key was{' '}
                <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
                  accountSettings
                </code>{' '}
                and gets converted to "Account Settings"
              </Text>
            </div>
          ),
          billingAndPayments: (
            <div className="p-4">
              <Text variant="h3" weight="stronger">
                Billing And Payments
              </Text>
              <Text>
                This tab key was{' '}
                <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
                  billingAndPayments
                </code>{' '}
                and gets converted to "Billing And Payments"
              </Text>
            </div>
          ),
          apiIntegrations: (
            <div className="p-4">
              <Text variant="h3" weight="stronger">
                Api Integrations
              </Text>
              <Text>
                This tab key was{' '}
                <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
                  apiIntegrations
                </code>{' '}
                and gets converted to "Api Integrations"
              </Text>
            </div>
          ),
        }}
      />
    </div>
  </div>
)

export const InitialActiveTab = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Initial Active Tab</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Use the{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          initActiveTab
        </code>{' '}
        prop to specify which tab should be active when the component first
        loads. If not specified, the first tab in the object will be active by
        default.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Settings Tab Initially Active</h4>
      <Tabs
        initActiveTab="settings"
        tabs={{
          overview: (
            <div className="p-4">
              <Text>Overview content - not initially active</Text>
            </div>
          ),
          settings: (
            <div className="p-4 space-y-3">
              <div className="flex items-center gap-2">
                <Icon variant="Gear" size="20" />
                <Text weight="strong">Settings - Initially Active!</Text>
              </div>
              <Text>
                This tab was set as the initial active tab using the
                initActiveTab prop.
              </Text>
            </div>
          ),
          help: (
            <div className="p-4">
              <Text>Help content</Text>
            </div>
          ),
        }}
      />
    </div>
  </div>
)

export const RichTabContent = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Rich Tab Content</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Tabs can contain any React content including complex layouts, forms,
        charts, tables, and interactive elements. The container automatically
        adjusts its height based on the active tab's content.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Dashboard Example</h4>
      <Tabs
        tabs={{
          dashboard: (
            <div className="p-4 space-y-4">
              <Text variant="h2" weight="stronger">
                Dashboard Overview
              </Text>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                <Card>
                  <Text variant="label" theme="neutral">
                    Total Users
                  </Text>
                  <Text variant="h2" weight="stronger">
                    2,847
                  </Text>
                  <Badge theme="success">+12% this month</Badge>
                </Card>
                <Card>
                  <Text variant="label" theme="neutral">
                    Revenue
                  </Text>
                  <Text variant="h2" weight="stronger">
                    $45,210
                  </Text>
                  <Badge theme="info">+8% this month</Badge>
                </Card>
                <Card>
                  <Text variant="label" theme="neutral">
                    Active Projects
                  </Text>
                  <Text variant="h2" weight="stronger">
                    23
                  </Text>
                  <Badge theme="warn">-2 this week</Badge>
                </Card>
              </div>
            </div>
          ),
          analytics: (
            <div className="p-4 space-y-4">
              <Text variant="h2" weight="stronger">
                Analytics
              </Text>
              <div className="space-y-3">
                <div>
                  <div className="flex justify-between mb-1">
                    <Text variant="subtext">Completion Rate</Text>
                    <Text variant="subtext">78%</Text>
                  </div>
                  <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
                    <div className="bg-blue-500 h-2 rounded-full w-3/4"></div>
                  </div>
                </div>
                <div>
                  <div className="flex justify-between mb-1">
                    <Text variant="subtext">User Satisfaction</Text>
                    <Text variant="subtext">92%</Text>
                  </div>
                  <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
                    <div className="bg-green-500 h-2 rounded-full w-11/12"></div>
                  </div>
                </div>
              </div>
            </div>
          ),
          reports: (
            <div className="p-4 space-y-4">
              <div className="flex justify-between items-center">
                <Text variant="h2" weight="stronger">
                  Reports
                </Text>
                <Button variant="primary" size="sm">
                  Generate Report
                </Button>
              </div>
              <div className="border rounded-lg overflow-hidden">
                <table className="w-full">
                  <thead className="bg-gray-50 dark:bg-gray-800">
                    <tr>
                      <th className="text-left p-3 font-medium">Date</th>
                      <th className="text-left p-3 font-medium">Revenue</th>
                      <th className="text-left p-3 font-medium">Users</th>
                      <th className="text-left p-3 font-medium">Status</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr className="border-t">
                      <td className="p-3">2024-01-01</td>
                      <td className="p-3">$1,250</td>
                      <td className="p-3">45</td>
                      <td className="p-3">
                        <Badge theme="success">Complete</Badge>
                      </td>
                    </tr>
                    <tr className="border-t">
                      <td className="p-3">2024-01-02</td>
                      <td className="p-3">$1,890</td>
                      <td className="p-3">67</td>
                      <td className="p-3">
                        <Badge theme="success">Complete</Badge>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          ),
        }}
      />
    </div>
  </div>
)

export const CustomStyling = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Custom Tab Styling</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Customize the appearance of tabs using{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          className
        </code>
        ,{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          tabsClassName
        </code>
        , and{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          tabControlsClassName
        </code>{' '}
        props for different parts of the component.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Styled Tab Container</h4>
      <Tabs
        className="border rounded-lg shadow-sm"
        tabControlsClassName="px-4 bg-gray-50 dark:bg-gray-800"
        tabsClassName="min-h-[200px] p-4"
        tabs={{
          customStyle: (
            <div className="space-y-3">
              <Text variant="h3" weight="stronger">
                Custom Styled Tabs
              </Text>
              <Text>
                This tab component has custom styling applied with borders,
                background colors, and padding adjustments.
              </Text>
              <Text variant="subtext" theme="neutral">
                The container has a border and shadow, the tab controls have a
                background color, and the content area has minimum height and
                padding.
              </Text>
            </div>
          ),
          anotherTab: (
            <div className="space-y-3">
              <Text variant="h3" weight="stronger">
                Another Tab
              </Text>
              <Text>
                This tab shares the same custom styling, demonstrating how the
                styling props affect the entire component.
              </Text>
            </div>
          ),
        }}
      />
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Styling Props:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>
          <strong>className:</strong> Styles the entire tab component wrapper
        </li>
        <li>
          <strong>tabControlsClassName:</strong> Styles the tab button container
        </li>
        <li>
          <strong>tabsClassName:</strong> Styles the tab content area
        </li>
        <li>All props support responsive and dark mode classes</li>
      </ul>
    </div>
  </div>
)
