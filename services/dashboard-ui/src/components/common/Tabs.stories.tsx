import { Tabs } from './Tabs'

export const Basic = () => (
  <Tabs
    tabs={{
      overview: <div className="p-4">This is the overview tab content</div>,
      settings: <div className="p-4">This is the settings tab content</div>,
      help: <div className="p-4">This is the help tab content</div>,
    }}
  />
)

export const WithInitialTab = () => (
  <Tabs
    initActiveTab="settings"
    tabs={{
      overview: (
        <div className="p-4">Overview content - not initially active</div>
      ),
      settings: <div className="p-4">Settings content - initially active!</div>,
      help: <div className="p-4">Help content</div>,
    }}
  />
)

export const CamelCaseKeys = () => (
  <Tabs
    tabs={{
      userProfile: (
        <div className="p-4">
          <h3>User Profile</h3>
          <p>This tab key was userProfile and gets converted to User Profile</p>
        </div>
      ),
      accountSettings: (
        <div className="p-4">
          <h3>Account Settings</h3>
          <p>
            This tab key was accountSettings and gets converted to Account
            Settings
          </p>
        </div>
      ),
      billingInfo: (
        <div className="p-4">
          <h3>Billing Info</h3>
          <p>This tab key was billingInfo and gets converted to Billing Info</p>
        </div>
      ),
    }}
  />
)

export const RichContent = () => (
  <Tabs
    tabs={{
      dashboard: (
        <div className="p-4 space-y-4">
          <h2 className="text-xl font-bold">Dashboard</h2>
          <div className="grid grid-cols-2 gap-4">
            <div className="p-4 bg-gray-100 rounded">Metric 1</div>
            <div className="p-4 bg-gray-100 rounded">Metric 2</div>
          </div>
        </div>
      ),
      analytics: (
        <div className="p-4 space-y-4">
          <h2 className="text-xl font-bold">Analytics</h2>
          <div className="space-y-2">
            <div className="w-full bg-gray-200 rounded-full h-2">
              <div className="bg-blue-500 h-2 rounded-full w-3/4"></div>
            </div>
            <p>75% completion rate</p>
          </div>
        </div>
      ),
      reports: (
        <div className="p-4">
          <h2 className="text-xl font-bold">Reports</h2>
          <table className="w-full mt-4">
            <thead>
              <tr className="border-b">
                <th className="text-left p-2">Date</th>
                <th className="text-left p-2">Value</th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <td className="p-2">2024-01-01</td>
                <td className="p-2">$1,000</td>
              </tr>
              <tr>
                <td className="p-2">2024-01-02</td>
                <td className="p-2">$1,500</td>
              </tr>
            </tbody>
          </table>
        </div>
      ),
    }}
  />
)

export const ManyTabs = () => (
  <Tabs
    tabs={{
      tab1: <div className="p-4">Content for Tab 1</div>,
      tab2: <div className="p-4">Content for Tab 2</div>,
      tab3: <div className="p-4">Content for Tab 3</div>,
      tab4: <div className="p-4">Content for Tab 4</div>,
      tab5: <div className="p-4">Content for Tab 5</div>,
      tab6: <div className="p-4">Content for Tab 6</div>,
      veryLongTabName: (
        <div className="p-4">
          This tab has a very long name that gets converted properly
        </div>
      ),
    }}
  />
)

export const SingleTab = () => (
  <Tabs
    tabs={{
      onlyTab: (
        <div className="p-4">
          <h3>Single Tab</h3>
          <p>This demonstrates the component with only one tab.</p>
        </div>
      ),
    }}
  />
)

export const CustomStyling = () => (
  <Tabs
    className="border rounded-lg"
    tabsClassName="min-h-[200px] p-2"
    tabs={{
      styled: (
        <div className="p-4 text-center">
          <p>
            This Tabs component has custom className and tabsClassName applied.
          </p>
          <p>The wrapper has a border and rounded corners.</p>
          <p>The tabs container has a minimum height and padding.</p>
        </div>
      ),
      another: (
        <div className="p-4">
          <p>Another tab with the same custom styling applied.</p>
        </div>
      ),
    }}
  />
)
