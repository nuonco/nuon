export default {
  title: 'Common/SplitButton',
}

import { Menu } from './Menu'
import { Button } from './Button'
import { SplitButton } from './SplitButton'
import { Text } from './Text'
import { Icon } from './Icon'

export const BasicUsage = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Basic Split Button</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        A Split Button combines a primary action button with a dropdown menu for
        secondary actions. The left side performs the main action, while the
        right side (three dots) opens a dropdown with additional options.
      </p>
    </div>

    <div className="space-y-4">
      <SplitButton
        buttonProps={{ children: 'Primary Action' }}
        dropdownProps={{
          children: (
            <Menu>
              <Text>Controls</Text>
              <Button>Secondary action</Button>
              <Button>Tertiary action</Button>
              <hr />
              <Text>Danger zone</Text>
              <Button variant="danger">Delete action</Button>
            </Menu>
          ),
          id: 'basic-example',
        }}
      />
    </div>
  </div>
)

export const Variants = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Split Button Variants</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        The{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          variant
        </code>{' '}
        prop controls the visual style of both the main button and dropdown
        trigger. Split buttons support primary and secondary variants.
      </p>
    </div>

    <div className="space-y-4">
      <div className="flex flex-wrap gap-4 items-center">
        <SplitButton
          variant="primary"
          buttonProps={{ children: 'Primary' }}
          dropdownProps={{
            children: (
              <Menu>
                <Button>Option 1</Button>
                <Button>Option 2</Button>
                <Button>Option 3</Button>
              </Menu>
            ),
            id: 'primary-variant',
          }}
        />

        <SplitButton
          variant="secondary"
          buttonProps={{ children: 'Secondary' }}
          dropdownProps={{
            children: (
              <Menu>
                <Button>Option 1</Button>
                <Button>Option 2</Button>
                <Button>Option 3</Button>
              </Menu>
            ),
            id: 'secondary-variant',
          }}
        />
      </div>
    </div>

    <div className="grid grid-cols-1 md:grid-cols-2 gap-3 text-sm mt-6">
      <div>
        <strong>primary:</strong> Purple filled style for main CTAs
      </div>
      <div>
        <strong>secondary:</strong> Outlined style for secondary actions
        (default)
      </div>
    </div>
  </div>
)

export const WithIcons = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Split Buttons with Icons</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Split buttons support icons in the main button through the{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          buttonProps
        </code>{' '}
        prop. Icons help users quickly identify the primary action.
      </p>
    </div>

    <div className="space-y-4">
      <div className="flex flex-wrap gap-4 items-center">
        <SplitButton
          variant="primary"
          buttonProps={{
            children: (
              <>
                <Icon variant="PlusIcon" size="16" />
                Create
              </>
            ),
          }}
          dropdownProps={{
            children: (
              <Menu>
                <Text>Create from template</Text>
                <Button>
                  Duplicate existing
                  <Icon variant="CopyIcon" size="16" />
                </Button>
                <Button>
                  Import from file
                  <Icon variant="UploadIcon" size="16" />
                </Button>
                <hr />
                <Button>
                  Generate with AI
                  <Icon variant="MagicWandIcon" size="16" />
                </Button>
              </Menu>
            ),
            id: 'create-with-icon',
          }}
        />

        <SplitButton
          variant="secondary"
          buttonProps={{
            children: (
              <>
                <Icon variant="DownloadIcon" size="16" />
                Export
              </>
            ),
          }}
          dropdownProps={{
            children: (
              <Menu>
                <Text>Export formats</Text>
                <Button>
                  Export as PDF
                  <Icon variant="FileTextIcon" size="16" />
                </Button>
                <Button>
                  Export as CSV
                  <Icon variant="FileCsvIcon" size="16" />
                </Button>
                <Button>
                  Export as ZIP
                  <Icon variant="FileZipIcon" size="16" />
                </Button>
              </Menu>
            ),
            id: 'export-with-icon',
          }}
        />

        <SplitButton
          variant="secondary"
          buttonProps={{
            children: (
              <>
                <Icon variant="ShareIcon" size="16" />
                Share
              </>
            ),
          }}
          dropdownProps={{
            children: (
              <Menu>
                <Text>Share options</Text>
                <Button>
                  Copy link
                  <Icon variant="LinkIcon" size="16" />
                </Button>
                <Button>
                  Send via email
                  <Icon variant="EnvelopeIcon" size="16" />
                </Button>
                <hr />
                <Button>
                  Team access
                  <Icon variant="UsersIcon" size="16" />
                </Button>
              </Menu>
            ),
            id: 'share-with-icon',
          }}
        />
      </div>
    </div>
  </div>
)

export const ComplexDropdowns = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Complex Dropdown Content</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Split button dropdowns can contain complex menu structures with
        sections, dividers, and different action types. Use{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          Text
        </code>{' '}
        components for section headers and{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          hr
        </code>{' '}
        elements for visual separation.
      </p>
    </div>

    <div className="space-y-4">
      <div className="flex flex-wrap gap-4 items-center">
        <SplitButton
          variant="primary"
          buttonProps={{ children: 'Deploy' }}
          dropdownProps={{
            children: (
              <Menu>
                <Text>Quick Deploy</Text>
                <Button>
                  Deploy to staging
                  <Icon variant="RocketIcon" size="16" />
                </Button>
                <Button>
                  Deploy to production
                  <Icon variant="CloudCheckIcon" size="16" />
                </Button>
                <hr />
                <Text>Advanced Options</Text>
                <Button>
                  Custom deployment
                  <Icon variant="GearIcon" size="16" />
                </Button>
                <Button>
                  Schedule deployment
                  <Icon variant="TimerIcon" size="16" />
                </Button>
                <Button>
                  Deploy from branch
                  <Icon variant="GitBranchIcon" size="16" />
                </Button>
                <hr />
                <Text>Rollback</Text>
                <Button variant="danger">
                  Rollback last deployment
                  <Icon variant="ArrowCounterClockwiseIcon" size="16" />
                </Button>
              </Menu>
            ),
            id: 'complex-deploy',
          }}
        />

        <SplitButton
          variant="secondary"
          buttonProps={{ children: 'Manage' }}
          dropdownProps={{
            children: (
              <Menu>
                <Text>Configuration</Text>
                <Button>
                  Settings
                  <Icon variant="SlidersIcon" size="16" />
                </Button>
                <Button>
                  Environment variables
                  <Icon variant="KeyIcon" size="16" />
                </Button>
                <Button>
                  Database settings
                  <Icon variant="DatabaseIcon" size="16" />
                </Button>
                <hr />
                <Text>Monitoring</Text>
                <Button>
                  View metrics
                  <Icon variant="ChartLineIcon" size="16" />
                </Button>
                <Button>
                  Error tracking
                  <Icon variant="BugIcon" size="16" />
                </Button>
                <Button>
                  Health checks
                  <Icon variant="PulseIcon" size="16" />
                </Button>
                <hr />
                <Text>Team</Text>
                <Button>
                  Invite members
                  <Icon variant="UserPlusIcon" size="16" />
                </Button>
                <Button>
                  Manage permissions
                  <Icon variant="CrownIcon" size="16" />
                </Button>
              </Menu>
            ),
            id: 'complex-manage',
          }}
        />
      </div>
    </div>
  </div>
)

export const UsageGuidelines = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Usage Guidelines</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Split buttons are ideal when you have one primary action with several
        related secondary actions. They help reduce interface complexity while
        keeping common actions easily accessible.
      </p>
    </div>

    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
      <div className="space-y-4">
        <h4 className="text-sm font-medium text-green-600 dark:text-green-400">
          ✅ Good Use Cases
        </h4>
        <div className="space-y-3 text-sm">
          <div>
            <strong>Create actions:</strong> Main "Create" button with template
            options in dropdown
          </div>
          <div>
            <strong>Export functions:</strong> Default export with format
            options in dropdown
          </div>
          <div>
            <strong>Deploy operations:</strong> Quick deploy with environment
            choices in dropdown
          </div>
          <div>
            <strong>Share features:</strong> Primary share with additional share
            methods in dropdown
          </div>
        </div>
      </div>

      <div className="space-y-4">
        <h4 className="text-sm font-medium text-red-600 dark:text-red-400">
          ❌ Avoid When
        </h4>
        <div className="space-y-3 text-sm">
          <div>
            <strong>Unrelated actions:</strong> Grouping unrelated functionality
            together
          </div>
          <div>
            <strong>Too many options:</strong> Dropdown with more than 8-10
            items
          </div>
          <div>
            <strong>Critical actions:</strong> Placing destructive actions as
            secondary options
          </div>
          <div>
            <strong>Simple cases:</strong> When a regular dropdown would be
            clearer
          </div>
        </div>
      </div>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Best Practices:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Make the primary action the most common or important choice</li>
        <li>Group related secondary actions logically with section headers</li>
        <li>Use dividers to separate different types of actions</li>
        <li>Keep dropdown menus concise and scannable</li>
        <li>Place destructive actions at the bottom with danger styling</li>
        <li>Ensure the primary action works independently of the dropdown</li>
      </ul>
    </div>
  </div>
)
