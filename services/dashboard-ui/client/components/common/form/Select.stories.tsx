import { Select } from './Select'
import { AWS_REGIONS, AZURE_REGIONS } from '@/configs/cloud-regions'
import { getFlagEmoji } from '@/utils/string-utils'

export default {
  title: 'Forms/Select',
  component: Select,
}

const sampleOptions = [
  { value: 'option1', label: 'Option 1' },
  { value: 'option2', label: 'Option 2' },
  { value: 'option3', label: 'Option 3' },
  { value: 'disabled', label: 'Disabled Option', disabled: true },
]

const awsRegionOptions = AWS_REGIONS.map((region) => ({
  value: region.value,
  label: region?.iconVariant
    ? `${getFlagEmoji(region.iconVariant.substring(5))} ${region.text} [${region.value}]`
    : region.text,
}))

const azureRegionOptions = AZURE_REGIONS.map((region) => ({
  value: region.value,
  label: region?.iconVariant
    ? `${getFlagEmoji(region.iconVariant.substring(5))} ${region.text}`
    : region.text,
}))

export const BasicUsage = () => (
  <div className="space-y-6 max-w-md">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Basic Select Usage</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Basic select dropdown with different configurations. The select component
        provides native HTML select functionality with custom styling and
        accessibility support.
      </p>
    </div>

    <div className="space-y-4">
      <Select options={sampleOptions} placeholder="Choose an option..." />
      <Select
        options={sampleOptions}
        placeholder="Select with label"
        labelProps={{ labelText: 'Choose Option' }}
      />
      <Select
        options={sampleOptions}
        placeholder="Select with helper text"
        helperText="This is some helpful information"
        labelProps={{ labelText: 'Configuration' }}
      />
    </div>
  </div>
)

export const Sizes = () => (
  <div className="space-y-6 max-w-md">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Select Sizes</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        The{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          size
        </code>{' '}
        prop controls the dimensions and padding of the select. Default size is{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          md
        </code>{' '}
        if no size prop is provided.
      </p>
    </div>

    <div className="space-y-4">
      <Select
        size="sm"
        options={sampleOptions}
        placeholder="Small select"
        labelProps={{ labelText: 'Small' }}
      />
      <Select
        size="md"
        options={sampleOptions}
        placeholder="Medium select (default)"
        labelProps={{ labelText: 'Medium' }}
      />
      <Select
        size="lg"
        options={sampleOptions}
        placeholder="Large select"
        labelProps={{ labelText: 'Large' }}
      />
    </div>

    <div className="grid grid-cols-1 gap-3 text-sm mt-4">
      <div>
        <strong>sm:</strong> 32px height, 4px vertical padding
      </div>
      <div>
        <strong>md:</strong> 40px height, 8px vertical padding (default)
      </div>
      <div>
        <strong>lg:</strong> 48px height, 12px vertical padding
      </div>
    </div>
  </div>
)

export const States = () => (
  <div className="space-y-6 max-w-md">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Select States</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Select components support various states including error states, disabled
        states, and helper text for better user experience and validation feedback.
      </p>
    </div>

    <div className="space-y-4">
      <Select
        options={sampleOptions}
        placeholder="Normal select"
        labelProps={{ labelText: 'Normal State' }}
        helperText="This is a normal select field"
      />
      
      <Select
        options={sampleOptions}
        placeholder="Error select"
        labelProps={{ labelText: 'Error State' }}
        error
        errorMessage="Please select a valid option"
      />
      
      <Select
        options={sampleOptions}
        placeholder="Disabled select"
        labelProps={{ labelText: 'Disabled State' }}
        disabled
        helperText="This field is currently disabled"
      />
      
      <Select
        options={sampleOptions}
        placeholder="Required field"
        labelProps={{ labelText: 'Required Field' }}
        required
        helperText="This field is required"
      />
    </div>

    <div className="grid grid-cols-1 gap-3 text-sm mt-6">
      <div>
        <strong>Normal:</strong> Default styling with focus states
      </div>
      <div>
        <strong>Error:</strong> Red border and error message styling
      </div>
      <div>
        <strong>Disabled:</strong> Muted colors and no interaction
      </div>
      <div>
        <strong>Required:</strong> Semantic HTML required attribute
      </div>
    </div>
  </div>
)

export const WithLabelsAndText = () => (
  <div className="space-y-6 max-w-md">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Labels and Helper Text</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Selects support labels and helper text to provide context and guidance
        to users. Labels are automatically associated with selects for accessibility.
      </p>
    </div>

    <div className="space-y-4">
      <Select
        options={sampleOptions}
        placeholder="Choose your preference"
        labelProps={{ 
          labelText: 'Preference',
          labelTextProps: { weight: 'strong' }
        }}
        helperText="Select your preferred option"
      />
      
      <Select
        options={awsRegionOptions}
        placeholder="Select region"
        labelProps={{ labelText: 'AWS Region' }}
        helperText="Choose the region closest to your users"
      />
      
      <Select
        options={sampleOptions}
        placeholder="Optional selection"
        labelProps={{ labelText: 'Category (Optional)' }}
      />
    </div>

    <div className="grid grid-cols-1 gap-3 text-sm mt-6">
      <div>
        <strong>Labels:</strong> Automatically linked to selects with htmlFor attribute
      </div>
      <div>
        <strong>Helper Text:</strong> Provides additional context below the select
      </div>
      <div>
        <strong>Styling:</strong> Labels support all Text component props for customization
      </div>
      <div>
        <strong>Accessibility:</strong> Proper ARIA attributes and associations
      </div>
    </div>
  </div>
)

export const ErrorHandling = () => (
  <div className="space-y-6 max-w-md">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Error Handling</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Error states provide visual feedback and error messages to guide users in
        making correct selections. Error messages replace helper text when present.
      </p>
    </div>

    <div className="space-y-4">
      <Select
        options={sampleOptions}
        placeholder="Select an option"
        labelProps={{ labelText: 'Required Selection' }}
        error
        errorMessage="This field is required"
      />
      
      <Select
        options={awsRegionOptions}
        placeholder="Select region"
        labelProps={{ labelText: 'AWS Region' }}
        error
        errorMessage="Please select a valid region"
      />
      
      <Select
        options={sampleOptions}
        placeholder="Choose category"
        labelProps={{ labelText: 'Category' }}
        error
        errorMessage="Invalid selection, please choose again"
      />
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Error State Features:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Red border and focus ring for visual indication</li>
        <li>Error messages displayed below the select</li>
        <li>Proper ARIA attributes for screen readers</li>
        <li>Error messages override helper text when both are present</li>
        <li>Custom styling support for error messages</li>
        <li>Native HTML validation integration</li>
      </ul>
    </div>
  </div>
)

export const RealWorldExamples = () => (
  <div className="space-y-6 max-w-md">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Real World Examples</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Examples of select components as they might appear in real forms and applications.
      </p>
    </div>

    <div className="space-y-4">
      <Select
        options={awsRegionOptions}
        placeholder="Select AWS region"
        labelProps={{ labelText: 'AWS Region' }}
        helperText="Choose the region for your deployment"
        required
      />
      
      <Select
        options={[
          { value: 'development', label: 'Development' },
          { value: 'staging', label: 'Staging' },
          { value: 'production', label: 'Production' },
        ]}
        placeholder="Select environment"
        labelProps={{ labelText: 'Environment' }}
        helperText="Target environment for this install"
        defaultValue="development"
      />
      
      <Select
        options={[
          { value: 'small', label: 'Small (1-2 CPUs, 2-4GB RAM)' },
          { value: 'medium', label: 'Medium (2-4 CPUs, 4-8GB RAM)' },
          { value: 'large', label: 'Large (4-8 CPUs, 8-16GB RAM)' },
          { value: 'xlarge', label: 'Extra Large (8+ CPUs, 16+ GB RAM)' },
        ]}
        placeholder="Select instance size"
        labelProps={{ labelText: 'Instance Size' }}
        helperText="Choose the appropriate size for your workload"
      />
    </div>
  </div>
)

export function AWSRegions() {
  return (
    <div className="space-y-6 max-w-md">
      <div className="space-y-3">
        <h3 className="text-lg font-semibold">AWS Regions</h3>
        <p className="text-sm text-gray-600 dark:text-gray-400">
          AWS region selector with flag emojis and region codes. This example uses
          the actual AWS regions from the configuration.
        </p>
      </div>

      <div className="space-y-4">
        <Select
          options={awsRegionOptions}
          labelProps={{
            labelText: 'AWS Region *',
            labelTextProps: { variant: 'body' }
          }}
          placeholder="Choose AWS region"
          helperText="Select the AWS region for your deployment"
          required
        />
      </div>
    </div>
  )
}

export function AzureRegions() {
  return (
    <div className="space-y-6 max-w-md">
      <div className="space-y-3">
        <h3 className="text-lg font-semibold">Azure Regions</h3>
        <p className="text-sm text-gray-600 dark:text-gray-400">
          Azure location selector with flag emojis. This example uses
          the actual Azure regions from the configuration.
        </p>
      </div>

      <div className="space-y-4">
        <Select
          options={azureRegionOptions}
          labelProps={{
            labelText: 'Azure Location *',
            labelTextProps: { variant: 'body' }
          }}
          placeholder="Choose Azure location"
          helperText="Select the Azure location for your deployment"
          required
        />
      </div>
    </div>
  )
}
