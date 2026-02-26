/* eslint-disable react/no-unescaped-entities */
import { CheckboxInput, Checkbox, CheckboxInputWithButton } from './CheckboxInput'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'

export default {
  title: 'Forms/CheckboxInput',
  component: CheckboxInput,
}

export const BasicUsage = () => (
  <div className="space-y-6 max-w-md">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Basic Checkbox Usage</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Checkbox components provide multiple variants for different use cases.
        The basic checkbox provides just the input element, while CheckboxInput
        includes integrated labels and hover states.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Basic Checkbox</h4>
      <div className="flex items-center gap-3">
        <Checkbox id="basic-1" />
        <label htmlFor="basic-1" className="text-sm">
          Basic checkbox with external label
        </label>
      </div>
      
      <h4 className="text-sm font-medium">CheckboxInput with Integrated Label</h4>
      <CheckboxInput
        labelProps={{ labelText: 'Agree to terms and conditions' }}
      />
      <CheckboxInput
        labelProps={{ labelText: 'Subscribe to newsletter' }}
        defaultChecked
      />
      <CheckboxInput
        labelProps={{ labelText: 'Enable notifications' }}
        disabled
      />
    </div>
  </div>
)

export const States = () => (
  <div className="space-y-6 max-w-md">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Checkbox States</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Checkboxes support various states including checked, unchecked, disabled,
        and indeterminate states. The component automatically handles focus
        states and provides visual feedback.
      </p>
    </div>

    <div className="space-y-4">
      <CheckboxInput
        labelProps={{ labelText: 'Unchecked state (default)' }}
      />
      <CheckboxInput
        labelProps={{ labelText: 'Checked state' }}
        defaultChecked
      />
      <CheckboxInput
        labelProps={{ labelText: 'Disabled unchecked' }}
        disabled
      />
      <CheckboxInput
        labelProps={{ labelText: 'Disabled checked' }}
        disabled
        defaultChecked
      />
      <CheckboxInput
        labelProps={{ labelText: 'Required field' }}
        required
      />
    </div>

    <div className="grid grid-cols-1 gap-3 text-sm mt-6">
      <div>
        <strong>Unchecked:</strong> Default state ready for user interaction
      </div>
      <div>
        <strong>Checked:</strong> Selected state with primary color accent
      </div>
      <div>
        <strong>Disabled:</strong> Non-interactive state with muted styling
      </div>
      <div>
        <strong>Required:</strong> Semantic HTML required attribute for forms
      </div>
    </div>
  </div>
)

export const WithCustomStyling = () => (
  <div className="space-y-6 max-w-md">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Custom Styling</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        CheckboxInput components support custom styling through labelTextProps
        and className overrides. You can customize the label text appearance
        and container styling.
      </p>
    </div>

    <div className="space-y-4">
      <CheckboxInput
        labelProps={{
          labelText: 'Bold label text',
          labelTextProps: { weight: 'strong' }
        }}
      />
      <CheckboxInput
        labelProps={{
          labelText: 'Custom colored text',
          labelTextProps: { theme: 'brand' }
        }}
      />
      <CheckboxInput
        labelProps={{
          labelText: 'Small text variant',
          labelTextProps: { variant: 'subtext' }
        }}
      />
      <CheckboxInput
        labelProps={{
          labelText: 'Custom container styling',
          className: 'bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800'
        }}
      />
    </div>

    <div className="grid grid-cols-1 gap-3 text-sm mt-6">
      <div>
        <strong>Label Styling:</strong> Use labelTextProps to customize text appearance
      </div>
      <div>
        <strong>Container Styling:</strong> Apply className to the label container
      </div>
      <div>
        <strong>Text Variants:</strong> Support all Text component variants and themes
      </div>
    </div>
  </div>
)

export const InForms = () => (
  <div className="space-y-6 max-w-md">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Checkboxes in Forms</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Common patterns for using checkboxes in forms including grouped options,
        settings toggles, and consent checkboxes with proper accessibility.
      </p>
    </div>

    <div className="space-y-6">
      <div className="space-y-3">
        <h4 className="text-sm font-medium">Preferences</h4>
        <div className="space-y-2 pl-2 border-l-2 border-gray-200 dark:border-gray-700">
          <CheckboxInput
            name="preferences"
            value="email"
            labelProps={{ labelText: 'Email notifications' }}
          />
          <CheckboxInput
            name="preferences"
            value="sms"
            labelProps={{ labelText: 'SMS notifications' }}
          />
          <CheckboxInput
            name="preferences"
            value="push"
            labelProps={{ labelText: 'Push notifications' }}
            defaultChecked
          />
        </div>
      </div>

      <div className="space-y-3">
        <h4 className="text-sm font-medium">Account Settings</h4>
        <div className="space-y-2 pl-2 border-l-2 border-gray-200 dark:border-gray-700">
          <CheckboxInput
            name="settings"
            value="public_profile"
            labelProps={{ labelText: 'Make profile public' }}
          />
          <CheckboxInput
            name="settings"
            value="two_factor"
            labelProps={{ 
              labelText: 'Enable two-factor authentication',
              labelTextProps: { weight: 'strong', theme: 'success' }
            }}
          />
          <CheckboxInput
            name="settings"
            value="data_sharing"
            labelProps={{ labelText: 'Allow usage data sharing' }}
          />
        </div>
      </div>

      <div className="space-y-3">
        <h4 className="text-sm font-medium">Legal Consent</h4>
        <div className="space-y-2">
          <CheckboxInput
            required
            name="consent"
            value="terms"
            labelProps={{ 
              labelText: 'I agree to the Terms of Service *',
              labelTextProps: { variant: 'body' }
            }}
          />
          <CheckboxInput
            name="consent"
            value="marketing"
            labelProps={{ labelText: 'I consent to receiving marketing emails' }}
          />
        </div>
      </div>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Form Best Practices:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Use consistent naming for grouped checkboxes</li>
        <li>Mark required checkboxes with asterisks and required attribute</li>
        <li>Group related options with visual hierarchy</li>
        <li>Provide clear, descriptive labels</li>
        <li>Use appropriate default states based on user expectations</li>
      </ul>
    </div>
  </div>
)

export const WithButtons = () => (
  <div className="space-y-6 max-w-md">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Checkbox with Button</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        CheckboxInputWithButton combines a checkbox with a button for more
        complex interactions. Useful for selection lists where each item
        has additional actions.
      </p>
    </div>

    <div className="space-y-4">
      <CheckboxInputWithButton
        buttonProps={{
          children: (
            <div className="flex items-center justify-between w-full">
              <span>Project Alpha</span>
              <Icon variant="Gear" size="16" />
            </div>
          )
        }}
      />
      <CheckboxInputWithButton
        defaultChecked
        buttonProps={{
          children: (
            <div className="flex items-center justify-between w-full">
              <span>Project Beta</span>
              <Icon variant="Gear" size="16" />
            </div>
          )
        }}
      />
      <CheckboxInputWithButton
        disabled
        buttonProps={{
          children: (
            <div className="flex items-center justify-between w-full">
              <span>Project Gamma (Archived)</span>
              <Icon variant="Archive" size="16" />
            </div>
          ),
          disabled: true
        }}
      />
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Button Checkbox Usage:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Ideal for selection lists with item-specific actions</li>
        <li>Button receives ghost variant styling by default</li>
        <li>Checkbox and button states can be controlled independently</li>
        <li>Supports all Button component props through buttonProps</li>
      </ul>
    </div>
  </div>
)