export default {
  title: 'Common/Dropdown',
}

import { Button } from './Button'
import { Dropdown } from './Dropdown'
import { Icon } from './Icon'
import { Link } from './Link'
import { Menu } from './Menu'
import { Text } from './Text'
import { Badge } from './Badge'
import { Status } from './Status'

export const BasicUsage = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Basic Dropdown Usage</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Dropdowns provide a contextual menu that appears when users click or
        focus on a trigger button. They automatically handle positioning, focus
        management, and click-outside-to-close behavior. Content is typically
        wrapped in a Menu component for proper styling.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Simple Dropdown</h4>
      <div className="flex gap-4">
        <Dropdown id="basic-dropdown" buttonText="Actions">
          <Menu className="min-w-48">
            <Text variant="label" theme="neutral">
              Quick Actions
            </Text>
            <Button variant="ghost">Edit Settings</Button>
            <Button variant="ghost">View Details</Button>
            <Link href="#" className="px-3 py-2">
              Documentation
            </Link>
            <hr />
            <Text variant="label" theme="neutral">
              Advanced
            </Text>
            <Button variant="ghost">Export Data</Button>
          </Menu>
        </Dropdown>
      </div>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Behavior:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Opens on click or keyboard focus of the trigger button</li>
        <li>Closes when clicking outside or pressing Escape</li>
        <li>Maintains focus management for accessibility</li>
        <li>Supports keyboard navigation within the menu</li>
      </ul>
    </div>
  </div>
)

export const Positioning = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Dropdown Positioning</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        The{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          position
        </code>{' '}
        prop controls where the dropdown content appears relative to the trigger
        button. Positioning includes proper spacing and collision detection.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Position Options</h4>
      <div className="flex flex-wrap gap-4 justify-center p-8">
        <Dropdown id="below-dropdown" buttonText="Below" position="below">
          <Menu className="min-w-40">
            <Button variant="ghost">Option 1</Button>
            <Button variant="ghost">Option 2</Button>
            <Button variant="ghost">Option 3</Button>
          </Menu>
        </Dropdown>

        <Dropdown id="above-dropdown" buttonText="Above" position="above">
          <Menu className="min-w-40">
            <Button variant="ghost">Option 1</Button>
            <Button variant="ghost">Option 2</Button>
            <Button variant="ghost">Option 3</Button>
          </Menu>
        </Dropdown>

        <Dropdown
          id="beside-left-dropdown"
          buttonText="Beside"
          position="beside"
        >
          <Menu className="min-w-40">
            <Button variant="ghost">Option 1</Button>
            <Button variant="ghost">Option 2</Button>
            <Button variant="ghost">Option 3</Button>
          </Menu>
        </Dropdown>
      </div>
    </div>

    <div className="grid grid-cols-1 md:grid-cols-3 gap-3 text-sm mt-6">
      <div>
        <strong>below:</strong> Appears below the trigger button (default)
      </div>
      <div>
        <strong>above:</strong> Appears above the trigger button
      </div>
      <div>
        <strong>beside:</strong> Appears to the side of the trigger button
      </div>
    </div>
  </div>
)

export const Alignment = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Dropdown Alignment</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        The{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          alignment
        </code>{' '}
        prop controls how the dropdown content aligns relative to the trigger
        button. This works in combination with the position prop to provide
        precise placement control.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Alignment Options</h4>
      <div className="flex justify-center gap-8 p-8">
        <Dropdown
          id="left-align-dropdown"
          buttonText="Left Aligned"
          alignment="left"
        >
          <Menu className="min-w-48">
            <Text variant="label" theme="neutral">
              Left Alignment
            </Text>
            <Button variant="ghost">This content aligns to the left</Button>
            <Button variant="ghost">Of the trigger button</Button>
          </Menu>
        </Dropdown>

        <Dropdown
          id="right-align-dropdown"
          buttonText="Right Aligned"
          alignment="right"
        >
          <Menu className="min-w-48">
            <Text variant="label" theme="neutral">
              Right Alignment
            </Text>
            <Button variant="ghost">This content aligns to the right</Button>
            <Button variant="ghost">Of the trigger button</Button>
          </Menu>
        </Dropdown>
      </div>
    </div>

    <div className="grid grid-cols-1 md:grid-cols-2 gap-3 text-sm mt-6">
      <div>
        <strong>left:</strong> Dropdown content aligns to the left edge of the
        trigger button
      </div>
      <div>
        <strong>right:</strong> Dropdown content aligns to the right edge of the
        trigger button
      </div>
    </div>
  </div>
)

export const CustomContent = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Custom Dropdown Content</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Dropdown content can include various components like buttons, links,
        text, status indicators, and dividers. The Menu component provides
        consistent styling and spacing for dropdown items.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Rich Content Example</h4>
      <div className="flex gap-4">
        <Dropdown id="rich-dropdown" buttonText="User Actions">
          <Menu className="min-w-52">
            <div className="px-3 py-2 border-b">
              <Text weight="strong">John Doe</Text>
              <Text variant="subtext" theme="neutral">
                john@example.com
              </Text>
            </div>
            <Text variant="label" theme="neutral">
              Account
            </Text>
            <Button variant="ghost">
              <Icon variant="User" size="16" />
              Profile Settings
            </Button>
            <Button variant="ghost">
              <Icon variant="Bell" size="16" />
              Notifications
            </Button>
            <hr />
            <Text variant="label" theme="neutral">
              Status
            </Text>
            <div className="px-3 py-2 flex items-center justify-between">
              <Text variant="subtext">System Health</Text>
              <Status status="success" isWithoutText />
            </div>
            <div className="px-3 py-2 flex items-center justify-between">
              <Text variant="subtext">Account Type</Text>
              <Badge theme="brand" size="sm">
                Premium
              </Badge>
            </div>
            <hr />
            <Button
              variant="ghost"
              className="!text-red-600 dark:!text-red-400"
            >
              <Icon variant="SignOut" size="16" />
              Sign Out
            </Button>
          </Menu>
        </Dropdown>

        <Dropdown id="status-dropdown" buttonText="View Status">
          <Menu className="min-w-48">
            <Text variant="label" theme="neutral">
              Service Status
            </Text>
            <div className="px-3 py-1 flex items-center justify-between">
              <Text variant="subtext">Database</Text>
              <Status status="success" />
            </div>
            <div className="px-3 py-1 flex items-center justify-between">
              <Text variant="subtext">API Gateway</Text>
              <Status status="success" />
            </div>
            <div className="px-3 py-1 flex items-center justify-between">
              <Text variant="subtext">Cache</Text>
              <Status status="warn" />
            </div>
            <hr />
            <Link href="#" className="px-3 py-2">
              <Icon variant="ArrowSquareOut" size="16" />
              View Full Status Page
            </Link>
          </Menu>
        </Dropdown>
      </div>
    </div>
  </div>
)

export const NestedDropdowns = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Nested Dropdowns</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Dropdowns can be nested within other dropdowns to create hierarchical
        menus. Nested dropdowns typically use the "beside" position and include
        visual indicators like arrow icons to show they contain submenus.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Multi-level Menu</h4>
      <div className="flex gap-4">
        <Dropdown id="nested-main" buttonText="Main Menu">
          <Menu className="min-w-48">
            <Text variant="label" theme="neutral">
              Quick Actions
            </Text>
            <Button variant="ghost">New Document</Button>
            <Button variant="ghost">Recent Files</Button>
            <hr />
            <Text variant="label" theme="neutral">
              Settings
            </Text>
            <Dropdown
              id="nested-settings"
              buttonText="Preferences"
              position="beside"
              alignment="right"
              icon={<Icon variant="CaretRight" />}
              variant="ghost"
              className="w-full justify-between"
            >
              <Menu className="min-w-40">
                <Button variant="ghost">Theme</Button>
                <Button variant="ghost">Language</Button>
                <Button variant="ghost">Shortcuts</Button>
                <hr />
                <Dropdown
                  id="nested-advanced"
                  buttonText="Advanced"
                  position="beside"
                  alignment="right"
                  icon={<Icon variant="CaretRight" />}
                  variant="ghost"
                  className="w-full justify-between"
                >
                  <Menu className="min-w-36">
                    <Button variant="ghost">Debug Mode</Button>
                    <Button variant="ghost">Performance</Button>
                    <Button variant="ghost">Experiments</Button>
                  </Menu>
                </Dropdown>
              </Menu>
            </Dropdown>
            <Button variant="ghost">Account Settings</Button>
          </Menu>
        </Dropdown>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Deep nesting (10 levels)</h4>
      <div className="flex gap-4">
        <Dropdown id="deep-root" buttonText="Level 1">
          <Menu className="min-w-48">
            <Button variant="ghost">Action at level 1</Button>
            <Dropdown
              id="deep-2"
              buttonText="Level 2"
              position="beside"
              alignment="right"
              icon={<Icon variant="CaretRight" />}
              variant="ghost"
              className="w-full justify-between"
            >
              <Menu className="min-w-44">
                <Button variant="ghost">Action at level 2</Button>
                <Dropdown
                  id="deep-3"
                  buttonText="Level 3"
                  position="beside"
                  alignment="left"
                  icon={<Icon variant="CaretLeft" />}
                  variant="ghost"
                  className="w-full justify-between"
                >
                  <Menu className="min-w-44">
                    <Button variant="ghost">Action at level 3</Button>
                    <Dropdown
                      id="deep-4"
                      buttonText="Level 4"
                      position="beside"
                      alignment="right"
                      icon={<Icon variant="CaretRight" />}
                      variant="ghost"
                      className="w-full justify-between"
                    >
                      <Menu className="min-w-44">
                        <Button variant="ghost">Action at level 4</Button>
                        <Dropdown
                          id="deep-5"
                          buttonText="Level 5"
                          position="beside"
                          alignment="left"
                          icon={<Icon variant="CaretLeft" />}
                          variant="ghost"
                          className="w-full justify-between"
                        >
                          <Menu className="min-w-44">
                            <Button variant="ghost">Action at level 5</Button>
                            <Dropdown
                              id="deep-6"
                              buttonText="Level 6"
                              position="beside"
                              alignment="right"
                              icon={<Icon variant="CaretRight" />}
                              variant="ghost"
                              className="w-full justify-between"
                            >
                              <Menu className="min-w-44">
                                <Button variant="ghost">
                                  Action at level 6
                                </Button>
                                <Dropdown
                                  id="deep-7"
                                  buttonText="Level 7"
                                  position="beside"
                                  alignment="left"
                                  icon={<Icon variant="CaretLeft" />}
                                  variant="ghost"
                                  className="w-full justify-between"
                                >
                                  <Menu className="min-w-44">
                                    <Button variant="ghost">
                                      Action at level 7
                                    </Button>
                                    <Dropdown
                                      id="deep-8"
                                      buttonText="Level 8"
                                      position="beside"
                                      alignment="right"
                                      icon={<Icon variant="CaretRight" />}
                                      variant="ghost"
                                      className="w-full justify-between"
                                    >
                                      <Menu className="min-w-44">
                                        <Button variant="ghost">
                                          Action at level 8
                                        </Button>
                                        <Dropdown
                                          id="deep-9"
                                          buttonText="Level 9"
                                          position="beside"
                                          alignment="left"
                                          icon={<Icon variant="CaretLeft" />}
                                          variant="ghost"
                                          className="w-full justify-between"
                                        >
                                          <Menu className="min-w-44">
                                            <Button variant="ghost">
                                              Action at level 9
                                            </Button>
                                            <Dropdown
                                              id="deep-10"
                                              buttonText="Level 10"
                                              position="beside"
                                              alignment="right"
                                              icon={
                                                <Icon variant="CaretRight" />
                                              }
                                              variant="ghost"
                                              className="w-full justify-between"
                                            >
                                              <Menu className="min-w-44">
                                                <Button variant="ghost">
                                                  Action at level 10
                                                </Button>
                                              </Menu>
                                            </Dropdown>
                                          </Menu>
                                        </Dropdown>
                                      </Menu>
                                    </Dropdown>
                                  </Menu>
                                </Dropdown>
                              </Menu>
                            </Dropdown>
                          </Menu>
                        </Dropdown>
                      </Menu>
                    </Dropdown>
                  </Menu>
                </Dropdown>
              </Menu>
            </Dropdown>
          </Menu>
        </Dropdown>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">
        Multiple siblings with nested children
      </h4>
      <div className="flex gap-4">
        <Dropdown id="sibling-root" buttonText="File menu">
          <Menu className="min-w-48">
            <Button variant="ghost">New file</Button>
            <Dropdown
              id="sib-open"
              buttonText="Open recent"
              position="beside"
              alignment="right"
              icon={<Icon variant="CaretRight" />}
              variant="ghost"
              className="w-full justify-between"
            >
              <Menu className="min-w-44">
                <Button variant="ghost">project-a.ts</Button>
                <Button variant="ghost">config.json</Button>
                <Dropdown
                  id="sib-open-more"
                  buttonText="More files"
                  position="beside"
                  alignment="right"
                  icon={<Icon variant="CaretRight" />}
                  variant="ghost"
                  className="w-full justify-between"
                >
                  <Menu className="min-w-40">
                    <Button variant="ghost">readme.md</Button>
                    <Button variant="ghost">package.json</Button>
                  </Menu>
                </Dropdown>
              </Menu>
            </Dropdown>
            <Dropdown
              id="sib-export"
              buttonText="Export as"
              position="beside"
              alignment="right"
              icon={<Icon variant="CaretRight" />}
              variant="ghost"
              className="w-full justify-between"
            >
              <Menu className="min-w-44">
                <Button variant="ghost">PDF</Button>
                <Button variant="ghost">HTML</Button>
                <Dropdown
                  id="sib-export-img"
                  buttonText="Image"
                  position="beside"
                  alignment="right"
                  icon={<Icon variant="CaretRight" />}
                  variant="ghost"
                  className="w-full justify-between"
                >
                  <Menu className="min-w-40">
                    <Button variant="ghost">PNG</Button>
                    <Button variant="ghost">SVG</Button>
                    <Button variant="ghost">JPEG</Button>
                  </Menu>
                </Dropdown>
              </Menu>
            </Dropdown>
            <Dropdown
              id="sib-share"
              buttonText="Share with"
              position="beside"
              alignment="right"
              icon={<Icon variant="CaretRight" />}
              variant="ghost"
              className="w-full justify-between"
            >
              <Menu className="min-w-44">
                <Button variant="ghost">Copy link</Button>
                <Dropdown
                  id="sib-share-team"
                  buttonText="Team"
                  position="beside"
                  alignment="right"
                  icon={<Icon variant="CaretRight" />}
                  variant="ghost"
                  className="w-full justify-between"
                >
                  <Menu className="min-w-40">
                    <Button variant="ghost">Engineering</Button>
                    <Button variant="ghost">Design</Button>
                    <Dropdown
                      id="sib-share-team-eng"
                      buttonText="Sub-teams"
                      position="beside"
                      alignment="left"
                      icon={<Icon variant="CaretLeft" />}
                      variant="ghost"
                      className="w-full justify-between"
                    >
                      <Menu className="min-w-40">
                        <Button variant="ghost">Frontend</Button>
                        <Button variant="ghost">Backend</Button>
                        <Button variant="ghost">Infra</Button>
                      </Menu>
                    </Dropdown>
                  </Menu>
                </Dropdown>
                <Dropdown
                  id="sib-share-ext"
                  buttonText="External"
                  position="beside"
                  alignment="right"
                  icon={<Icon variant="CaretRight" />}
                  variant="ghost"
                  className="w-full justify-between"
                >
                  <Menu className="min-w-40">
                    <Button variant="ghost">Email</Button>
                    <Button variant="ghost">Slack</Button>
                  </Menu>
                </Dropdown>
              </Menu>
            </Dropdown>
            <hr />
            <Button variant="ghost">Close</Button>
          </Menu>
        </Dropdown>
      </div>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Nested Dropdown Guidelines:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Use "beside" positioning for nested dropdowns</li>
        <li>Include visual indicators (arrows) to show expandable items</li>
        <li>Keep nesting levels to a minimum (2-3 levels max)</li>
        <li>Ensure adequate spacing and hover/focus states</li>
      </ul>
    </div>
  </div>
)
