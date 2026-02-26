/* eslint-disable react/no-unescaped-entities */
import { RadioInput } from './RadioInput'

export default {
  title: 'Forms/RadioInput',
  component: RadioInput,
}

export const BasicUsage = () => (
  <div className="space-y-6 max-w-md">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Basic Radio Input Usage</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        RadioInput components provide an integrated radio button with label,
        hover states, and focus indicators. They automatically handle proper
        accessibility attributes and styling.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Individual Radio Inputs</h4>
      <RadioInput
        name="example"
        value="option1"
        labelProps={{ labelText: 'Option 1' }}
      />
      <RadioInput
        name="example"
        value="option2"
        labelProps={{ labelText: 'Option 2' }}
        defaultChecked
      />
      <RadioInput
        name="example"
        value="option3"
        labelProps={{ labelText: 'Option 3 (disabled)' }}
        disabled
      />
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-4 p-3 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Note:</strong> Radio inputs in the same group should have the same
      name attribute to ensure only one can be selected at a time.
    </div>
  </div>
)

export const RadioGroups = () => (
  <div className="space-y-6 max-w-md">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Radio Button Groups</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Radio buttons are typically used in groups where only one option can be
        selected. Group related options with visual hierarchy and clear labels.
      </p>
    </div>

    <div className="space-y-6">
      <div className="space-y-3">
        <h4 className="text-sm font-medium">Deployment Environment</h4>
        <div className="space-y-2 pl-2 border-l-2 border-gray-200 dark:border-gray-700">
          <RadioInput
            name="environment"
            value="development"
            labelProps={{ labelText: 'Development' }}
            defaultChecked
          />
          <RadioInput
            name="environment"
            value="staging"
            labelProps={{ labelText: 'Staging' }}
          />
          <RadioInput
            name="environment"
            value="production"
            labelProps={{ labelText: 'Production' }}
          />
        </div>
      </div>

      <div className="space-y-3">
        <h4 className="text-sm font-medium">Instance Size</h4>
        <div className="space-y-2 pl-2 border-l-2 border-gray-200 dark:border-gray-700">
          <RadioInput
            name="size"
            value="small"
            labelProps={{ labelText: 'Small (1 CPU, 2GB RAM)' }}
          />
          <RadioInput
            name="size"
            value="medium"
            labelProps={{ labelText: 'Medium (2 CPU, 4GB RAM)' }}
            defaultChecked
          />
          <RadioInput
            name="size"
            value="large"
            labelProps={{ labelText: 'Large (4 CPU, 8GB RAM)' }}
          />
          <RadioInput
            name="size"
            value="xlarge"
            labelProps={{ labelText: 'Extra Large (8 CPU, 16GB RAM)' }}
          />
        </div>
      </div>

      <div className="space-y-3">
        <h4 className="text-sm font-medium">Billing Frequency</h4>
        <div className="space-y-2 pl-2 border-l-2 border-gray-200 dark:border-gray-700">
          <RadioInput
            name="billing"
            value="monthly"
            labelProps={{ labelText: 'Monthly ($29/month)' }}
          />
          <RadioInput
            name="billing"
            value="yearly"
            labelProps={{ labelText: 'Yearly ($290/year - Save 17%)' }}
            defaultChecked
          />
        </div>
      </div>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Radio Group Best Practices:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Use consistent naming within each group</li>
        <li>Provide descriptive labels with additional context</li>
        <li>Set appropriate default selections</li>
        <li>Use visual grouping (borders, spacing) to show relationships</li>
        <li>Order options logically (size, frequency, importance)</li>
      </ul>
    </div>
  </div>
)

export const States = () => (
  <div className="space-y-6 max-w-md">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Radio Input States</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Radio inputs support various states including selected, unselected,
        disabled, and required states. The component provides visual feedback
        and proper accessibility attributes.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Basic States</h4>
      <RadioInput
        name="states1"
        value="unselected"
        labelProps={{ labelText: 'Unselected state (default)' }}
      />
      <RadioInput
        name="states1"
        value="selected"
        labelProps={{ labelText: 'Selected state' }}
        defaultChecked
      />
      
      <h4 className="text-sm font-medium">Disabled States</h4>
      <RadioInput
        name="states2"
        value="disabled-unselected"
        labelProps={{ labelText: 'Disabled unselected' }}
        disabled
      />
      <RadioInput
        name="states2"
        value="disabled-selected"
        labelProps={{ labelText: 'Disabled selected' }}
        disabled
        defaultChecked
      />

      <h4 className="text-sm font-medium">Required Field</h4>
      <RadioInput
        name="states3"
        value="required1"
        labelProps={{ labelText: 'Option A *' }}
        required
      />
      <RadioInput
        name="states3"
        value="required2"
        labelProps={{ labelText: 'Option B *' }}
        required
      />
    </div>

    <div className="grid grid-cols-1 gap-3 text-sm mt-6">
      <div>
        <strong>Unselected:</strong> Default state with hover effects
      </div>
      <div>
        <strong>Selected:</strong> Filled with primary color accent
      </div>
      <div>
        <strong>Disabled:</strong> Muted styling, no user interaction
      </div>
      <div>
        <strong>Required:</strong> Semantic HTML required attribute
      </div>
    </div>
  </div>
)

export const CustomStyling = () => (
  <div className="space-y-6 max-w-md">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Custom Styling</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        RadioInput supports custom styling through labelTextProps and className
        overrides. Customize text appearance, container styling, and add visual
        enhancements.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Text Styling Variants</h4>
      <RadioInput
        name="custom1"
        value="bold"
        labelProps={{
          labelText: 'Bold label text',
          labelTextProps: { weight: 'strong' }
        }}
      />
      <RadioInput
        name="custom1"
        value="brand"
        labelProps={{
          labelText: 'Brand colored text',
          labelTextProps: { theme: 'brand' }
        }}
        defaultChecked
      />
      <RadioInput
        name="custom1"
        value="small"
        labelProps={{
          labelText: 'Small text variant',
          labelTextProps: { variant: 'subtext' }
        }}
      />

      <h4 className="text-sm font-medium">Container Styling</h4>
      <RadioInput
        name="custom2"
        value="highlighted"
        labelProps={{
          labelText: 'Highlighted option',
          className: 'bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800'
        }}
        defaultChecked
      />
      <RadioInput
        name="custom2"
        value="success"
        labelProps={{
          labelText: 'Success themed option',
          className: 'bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800',
          labelTextProps: { theme: 'success' }
        }}
      />

      <h4 className="text-sm font-medium">Card-Style Options</h4>
      <RadioInput
        name="custom3"
        value="card1"
        labelProps={{
          labelText: 'Premium Plan - $29/month',
          className: 'bg-white dark:bg-gray-800 border-2 border-gray-200 dark:border-gray-700 shadow-sm hover:shadow-md transition-shadow'
        }}
      />
      <RadioInput
        name="custom3"
        value="card2"
        labelProps={{
          labelText: 'Enterprise Plan - $99/month',
          className: 'bg-gradient-to-r from-purple-50 to-blue-50 dark:from-purple-900/20 dark:to-blue-900/20 border-2 border-purple-200 dark:border-purple-800',
          labelTextProps: { weight: 'strong', theme: 'brand' }
        }}
        defaultChecked
      />
    </div>

    <div className="grid grid-cols-1 gap-3 text-sm mt-6">
      <div>
        <strong>Text Customization:</strong> Use labelTextProps for text styling
      </div>
      <div>
        <strong>Container Styling:</strong> Apply className to the label wrapper
      </div>
      <div>
        <strong>Visual Hierarchy:</strong> Use colors and styling to show importance
      </div>
      <div>
        <strong>Card Patterns:</strong> Create card-like selection interfaces
      </div>
    </div>
  </div>
)

export const InForms = () => (
  <div className="space-y-6 max-w-md">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Radio Inputs in Forms</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Common form patterns using radio inputs for single-choice selections
        including preferences, configurations, and user choices with proper
        form structure and validation.
      </p>
    </div>

    <form className="space-y-6">
      <fieldset className="space-y-3">
        <legend className="text-sm font-medium text-gray-900 dark:text-gray-100">
          Notification Preferences
        </legend>
        <div className="space-y-2 pl-2 border-l-2 border-gray-200 dark:border-gray-700">
          <RadioInput
            name="notifications"
            value="all"
            labelProps={{ labelText: 'All notifications' }}
            defaultChecked
          />
          <RadioInput
            name="notifications"
            value="important"
            labelProps={{ labelText: 'Important only' }}
          />
          <RadioInput
            name="notifications"
            value="none"
            labelProps={{ labelText: 'No notifications' }}
          />
        </div>
      </fieldset>

      <fieldset className="space-y-3">
        <legend className="text-sm font-medium text-gray-900 dark:text-gray-100">
          Account Type *
        </legend>
        <div className="space-y-2 pl-2 border-l-2 border-gray-200 dark:border-gray-700">
          <RadioInput
            name="account_type"
            value="personal"
            labelProps={{ labelText: 'Personal Account' }}
            required
          />
          <RadioInput
            name="account_type"
            value="business"
            labelProps={{ labelText: 'Business Account' }}
            required
          />
          <RadioInput
            name="account_type"
            value="enterprise"
            labelProps={{ labelText: 'Enterprise Account' }}
            required
          />
        </div>
      </fieldset>

      <fieldset className="space-y-3">
        <legend className="text-sm font-medium text-gray-900 dark:text-gray-100">
          Data Processing Location
        </legend>
        <div className="space-y-2 pl-2 border-l-2 border-gray-200 dark:border-gray-700">
          <RadioInput
            name="location"
            value="us-east"
            labelProps={{ labelText: 'US East (Virginia)' }}
            defaultChecked
          />
          <RadioInput
            name="location"
            value="us-west"
            labelProps={{ labelText: 'US West (Oregon)' }}
          />
          <RadioInput
            name="location"
            value="eu-west"
            labelProps={{ labelText: 'EU West (Ireland)' }}
          />
          <RadioInput
            name="location"
            value="ap-southeast"
            labelProps={{ labelText: 'Asia Pacific (Sydney)' }}
          />
        </div>
      </fieldset>
    </form>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Form Structure Best Practices:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Use fieldset and legend for grouping related radio inputs</li>
        <li>Mark required groups with asterisks in the legend</li>
        <li>Provide clear, descriptive labels for each option</li>
        <li>Set sensible defaults to reduce user friction</li>
        <li>Use consistent naming conventions across the form</li>
        <li>Test keyboard navigation and screen reader compatibility</li>
      </ul>
    </div>
  </div>
)